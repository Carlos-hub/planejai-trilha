# Aulas sequenciais por turma — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Uma turma tem várias aulas ordenadas com progressão travada; o aluno só libera a próxima ao concluir (enviar o quiz) a anterior. Aulas reutilizáveis entre turmas.

**Architecture:** Nova tabela de ligação `turma_lessons(turma_id, lesson_plan_id, ordem)`. Conclusão/unlock derivados de `student_attempts.concluido_em` (sem tabela de progresso). Endpoints do professor gerenciam a sequência na página da turma; endpoint do aluno devolve a sequência com estado de trava. Frontend: seção "Aulas da turma" na página da turma + nova home `/aluno`.

**Tech Stack:** Go (chi, pgx v5, sqlc, goose), Postgres 16, Next.js 16 (App Router, TS, Tailwind).

## Global Constraints

- Migrations: goose, em `backend/migrations/`, prefixo sequencial `0000N_`, blocos `-- +goose Up` / `-- +goose Down`.
- Queries: sqlc v2, arquivos em `backend/internal/store/queries/*.sql`, geração via `cd backend && sqlc generate` (gera `internal/store/*.sql.go`). Nunca editar `*.sql.go` à mão.
- Nullable → sqlc emite ponteiros (`emit_pointers_for_null_types: true`). `codigo` é `*string`.
- Testes Go: HTTP via `NewRouter(d)` + `httptest`; `loginProfessor(t, d, email)` semeia professor+sessão e devolve cookie `sid`. Emails únicos por teste (evita colisão de fixture). Rodar: `cd backend && TEST_DATABASE_URL=postgres://planejai:planejai@localhost:5433/planejai?sslmode=disable go test ./internal/http/ -run <Name> -v`.
- DB de dev/teste: porta `5433`, user/db/senha `planejai`.
- `Deps` (router.go): `Store *store.Queries`, `Pool *pgxpool.Pool`. Transações via `d.Pool.Begin(ctx)` + `d.Store.WithTx(tx)`.
- Auth: `userIDFromContext(r) (int64, bool)` (professor); `studentIDFromContext(r) (int64, bool)` (aluno, grupo `RequireStudent`).
- Frontend: chamadas via `apiFetch<T>(path, init)` em `frontend/lib/api.ts`; tipos em `frontend/lib/types.ts`. Verificação: `cd frontend && npx tsc --noEmit`.

---

### Task 1: Migração `turma_lessons` + backfill

**Files:**
- Create: `backend/migrations/00006_turma_lessons.sql`

**Interfaces:**
- Produces: tabela `turma_lessons(id, turma_id, lesson_plan_id, ordem, created_at)`, `UNIQUE(turma_id, lesson_plan_id)`, índice `(turma_id, ordem)`.

- [ ] **Step 1: Escrever a migração**

Create `backend/migrations/00006_turma_lessons.sql`:

```sql
-- +goose Up
CREATE TABLE turma_lessons (
  id BIGSERIAL PRIMARY KEY,
  turma_id BIGINT NOT NULL REFERENCES turmas(id) ON DELETE CASCADE,
  lesson_plan_id BIGINT NOT NULL REFERENCES lesson_plans(id) ON DELETE CASCADE,
  ordem INT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (turma_id, lesson_plan_id)
);

CREATE INDEX idx_turma_lessons_turma_ordem ON turma_lessons (turma_id, ordem);

-- Backfill: cada trilha já vinculada a uma turma vira uma aula da turma,
-- ordenada pela data de publicação (fallback: id).
INSERT INTO turma_lessons (turma_id, lesson_plan_id, ordem)
SELECT st.turma_id,
       st.lesson_plan_id,
       row_number() OVER (PARTITION BY st.turma_id
                          ORDER BY st.publicada_em NULLS LAST, st.id)
FROM study_trails st
WHERE st.turma_id IS NOT NULL;

-- +goose Down
DROP TABLE turma_lessons;
```

- [ ] **Step 2: Aplicar a migração**

Run: `cd backend && goose -dir migrations postgres "postgres://planejai:planejai@localhost:5433/planejai?sslmode=disable" up`
Expected: `OK   00006_turma_lessons.sql`

- [ ] **Step 3: Verificar tabela + índice**

Run: `docker compose exec -T db sh -c 'psql -U $POSTGRES_USER -d $POSTGRES_DB -c "\d turma_lessons"'`
Expected: colunas `id, turma_id, lesson_plan_id, ordem, created_at`; UNIQUE `(turma_id, lesson_plan_id)`; índice `idx_turma_lessons_turma_ordem`.

- [ ] **Step 4: Commit**

```bash
git add backend/migrations/00006_turma_lessons.sql
git commit -m "feat(db): turma_lessons join table for sequential lessons"
```

---

### Task 2: Queries sqlc para `turma_lessons`

**Files:**
- Create: `backend/internal/store/queries/turma_lessons.sql`
- Generated (do not edit): `backend/internal/store/turma_lessons.sql.go`, `backend/internal/store/models.go` (novo tipo `TurmaLesson`)

**Interfaces:**
- Produces (Go, package `store`):
  - `AttachTurmaLesson(ctx, AttachTurmaLessonParams{TurmaID, LessonPlanID int64}) (TurmaLesson, error)`
  - `DetachTurmaLesson(ctx, DetachTurmaLessonParams{TurmaID, LessonPlanID int64}) error`
  - `RenumberTurmaLessons(ctx, turmaID int64) error`
  - `SetTurmaLessonOrder(ctx, SetTurmaLessonOrderParams{TurmaID int64, LessonPlanIds []int64}) error`
  - `ListTurmaLessons(ctx, turmaID int64) ([]ListTurmaLessonsRow, error)` — row: `LessonPlanID int64, Ordem int32, Status string, Codigo *string, Label string`
  - `ListStudentTurmaLessons(ctx, ListStudentTurmaLessonsParams{TurmaID, StudentID int64}) ([]ListStudentTurmaLessonsRow, error)` — row: `LessonPlanID int64, Ordem int32, Codigo *string, Label string, Pontos int32, Concluido bool`

- [ ] **Step 1: Escrever as queries**

Create `backend/internal/store/queries/turma_lessons.sql`:

```sql
-- name: AttachTurmaLesson :one
INSERT INTO turma_lessons (turma_id, lesson_plan_id, ordem)
VALUES ($1, $2, COALESCE((SELECT max(ordem) FROM turma_lessons WHERE turma_id=$1), 0) + 1)
RETURNING *;

-- name: DetachTurmaLesson :exec
DELETE FROM turma_lessons WHERE turma_id=$1 AND lesson_plan_id=$2;

-- name: RenumberTurmaLessons :exec
WITH ordered AS (
  SELECT id, row_number() OVER (ORDER BY ordem) AS rn
  FROM turma_lessons WHERE turma_id=$1
)
UPDATE turma_lessons tl SET ordem = ordered.rn
FROM ordered WHERE tl.id = ordered.id;

-- name: SetTurmaLessonOrder :exec
UPDATE turma_lessons tl
SET ordem = data.ord
FROM (
  SELECT unnest(@lesson_plan_ids::bigint[]) AS lesson_plan_id,
         generate_subscripts(@lesson_plan_ids::bigint[], 1) AS ord
) data
WHERE tl.turma_id = @turma_id AND tl.lesson_plan_id = data.lesson_plan_id;

-- name: ListTurmaLessons :many
SELECT tl.lesson_plan_id, tl.ordem, lp.status, st.codigo,
       COALESCE(bs.assunto, 'Aula #' || lp.id::text) AS label
FROM turma_lessons tl
JOIN lesson_plans lp ON lp.id = tl.lesson_plan_id
LEFT JOIN study_trails st ON st.lesson_plan_id = lp.id
LEFT JOIN bncc_skills bs ON bs.id = lp.bncc_skill_id
WHERE tl.turma_id = $1
ORDER BY tl.ordem;

-- name: ListStudentTurmaLessons :many
SELECT tl.lesson_plan_id, tl.ordem, st.codigo,
       COALESCE(bs.assunto, 'Aula #' || lp.id::text) AS label,
       COALESCE(sa.pontos, 0)::int AS pontos,
       (sa.concluido_em IS NOT NULL) AS concluido
FROM turma_lessons tl
JOIN lesson_plans lp ON lp.id = tl.lesson_plan_id
LEFT JOIN study_trails st ON st.lesson_plan_id = lp.id
LEFT JOIN bncc_skills bs ON bs.id = lp.bncc_skill_id
LEFT JOIN LATERAL (
  SELECT sa.pontos, sa.concluido_em
  FROM student_attempts sa
  WHERE sa.study_trail_id = st.id
    AND sa.student_id = @student_id
    AND sa.concluido_em IS NOT NULL
  ORDER BY sa.concluido_em DESC
  LIMIT 1
) sa ON true
WHERE tl.turma_id = @turma_id
ORDER BY tl.ordem;
```

- [ ] **Step 2: Gerar código**

Run: `cd backend && sqlc generate`
Expected: sem erros; cria `internal/store/turma_lessons.sql.go` e adiciona `TurmaLesson` em `internal/store/models.go`.

- [ ] **Step 3: Verificar compilação**

Run: `cd backend && go build ./...`
Expected: sem saída (sucesso).

- [ ] **Step 4: Commit**

```bash
git add backend/internal/store/queries/turma_lessons.sql backend/internal/store/turma_lessons.sql.go backend/internal/store/models.go
git commit -m "feat(store): sqlc queries for turma_lessons"
```

---

### Task 3: Helper `ensureTrailCodigo` + endpoint anexar aula

**Files:**
- Create: `backend/internal/http/turma_lessons_handlers.go`
- Modify: `backend/internal/http/lesson_handlers.go` (extrair loop de código de `publishTrail`)
- Modify: `backend/internal/http/router.go:61` (registrar rota)
- Test: `backend/internal/http/turma_lessons_handlers_test.go`

**Interfaces:**
- Consumes: `store.AttachTurmaLesson`, `store.GetTurma`, `store.GetLessonPlan`, `store.GetTrailByLesson`, `store.PublishTrail`, `domain.NewTrailCode`.
- Produces:
  - `func (d Deps) ensureTrailCodigo(ctx context.Context, trail store.StudyTrail) (string, error)`
  - `func (d Deps) attachTurmaLesson(w http.ResponseWriter, r *http.Request)` → rota `POST /api/turmas/{id}/lessons`, body `{"lesson_plan_id": <int>}`, resposta `200 {"lesson_plan_id","ordem"}`.

- [ ] **Step 1: Escrever o teste (falha)**

Create `backend/internal/http/turma_lessons_handlers_test.go`:

```go
package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Carlos-hub/planejai/backend/internal/store"
)

// seedReadyLesson creates a lesson_plan (status pronto) with a trail for userID.
func seedReadyLesson(t *testing.T, d Deps, userID int64) int64 {
	t.Helper()
	ctx := context.Background()
	lp, err := d.Store.CreateLessonPlan(ctx, store.CreateLessonPlanParams{
		UserID: userID, DuracaoMin: 50, Origem: "ia", Status: "pronto",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.Store.CreateTrail(ctx, lp.ID); err != nil {
		t.Fatal(err)
	}
	return lp.ID
}

func seedTurma(t *testing.T, d Deps, userID int64, nome string) int64 {
	t.Helper()
	turma, err := d.Store.CreateTurma(context.Background(), store.CreateTurmaParams{
		UserID: userID, Nome: nome,
	})
	if err != nil {
		t.Fatal(err)
	}
	return turma.ID
}

func TestAttachTurmaLesson(t *testing.T) {
	d := testDeps(t)
	r := NewRouter(d)
	cookie := loginProfessor(t, d, "attach@t.com")
	uid := userIDFromCookie(t, d, cookie)
	turmaID := seedTurma(t, d, uid, "6A-attach")
	lessonID := seedReadyLesson(t, d, uid)

	body, _ := json.Marshal(map[string]any{"lesson_plan_id": lessonID})
	req := httptest.NewRequest("POST", "/api/turmas/"+itoa(turmaID)+"/lessons", bytes.NewReader(body))
	req.AddCookie(cookie)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("attach status=%d body=%s", w.Code, w.Body)
	}
	var got struct {
		LessonPlanID int64 `json:"lesson_plan_id"`
		Ordem        int   `json:"ordem"`
	}
	json.Unmarshal(w.Body.Bytes(), &got)
	if got.LessonPlanID != lessonID || got.Ordem != 1 {
		t.Fatalf("got %+v, want lesson=%d ordem=1", got, lessonID)
	}

	// duplicata → 409
	req2 := httptest.NewRequest("POST", "/api/turmas/"+itoa(turmaID)+"/lessons", bytes.NewReader(body))
	req2.AddCookie(cookie)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusConflict {
		t.Fatalf("duplicate status=%d, want 409", w2.Code)
	}
}
```

Add helper `userIDFromCookie` to the same file:

```go
func userIDFromCookie(t *testing.T, d Deps, c *http.Cookie) int64 {
	t.Helper()
	s, err := d.Store.GetSession(context.Background(), c.Value)
	if err != nil {
		t.Fatal(err)
	}
	return s.UserID
}
```

- [ ] **Step 2: Rodar o teste (deve falhar a compilação)**

Run: `cd backend && TEST_DATABASE_URL=postgres://planejai:planejai@localhost:5433/planejai?sslmode=disable go test ./internal/http/ -run TestAttachTurmaLesson -v`
Expected: FAIL — `undefined: attachTurmaLesson` / rota inexistente (404).

- [ ] **Step 3: Extrair `ensureTrailCodigo` de `publishTrail`**

Em `backend/internal/http/lesson_handlers.go`, dentro de `publishTrail`, substituir o bloco do loop `for attempt := 0; attempt < maxPublishAttempts; attempt++ { ... }` (que produz `codigo`) por:

```go
	codigo, err := d.ensureTrailCodigo(ctx, trail)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao publicar trilha"})
		return
	}
```

(Remover também a checagem `if codigo == ""` que ficava logo após, pois `ensureTrailCodigo` já trata isso retornando erro.)

- [ ] **Step 4: Escrever o handler + helper**

Create `backend/internal/http/turma_lessons_handlers.go`:

```go
package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/Carlos-hub/planejai/backend/internal/domain"
	"github.com/Carlos-hub/planejai/backend/internal/store"
)

// ensureTrailCodigo returns the trail's public code, generating one (with
// retry on unique-collision) if it has not been published yet.
func (d Deps) ensureTrailCodigo(ctx context.Context, trail store.StudyTrail) (string, error) {
	if trail.Codigo != nil && *trail.Codigo != "" {
		return *trail.Codigo, nil
	}
	for attempt := 0; attempt < maxPublishAttempts; attempt++ {
		candidate := domain.NewTrailCode()
		published, err := d.Store.PublishTrail(ctx, store.PublishTrailParams{ID: trail.ID, Codigo: &candidate})
		if err == nil {
			return *published.Codigo, nil
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			continue
		}
		return "", err
	}
	return "", errors.New("não foi possível gerar um código único")
}

// ownedTurma loads a turma and confirms it belongs to userID. Writes 404 and
// returns ok=false otherwise.
func (d Deps) ownedTurma(w http.ResponseWriter, r *http.Request, userID int64) (store.Turma, bool) {
	id, err := strconvParseID(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "turma não encontrada"})
		return store.Turma{}, false
	}
	turma, err := d.Store.GetTurma(r.Context(), id)
	if err != nil || turma.UserID != userID {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "turma não encontrada"})
		return store.Turma{}, false
	}
	return turma, true
}

func (d Deps) attachTurmaLesson(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "não autenticado"})
		return
	}
	turma, ok := d.ownedTurma(w, r, userID)
	if !ok {
		return
	}
	var in struct {
		LessonPlanID int64 `json:"lesson_plan_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "json inválido"})
		return
	}
	ctx := r.Context()

	lp, err := d.Store.GetLessonPlan(ctx, in.LessonPlanID)
	if err != nil || lp.UserID != userID {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "aula não encontrada"})
		return
	}
	if lp.Status != "pronto" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "aula ainda não está pronta"})
		return
	}

	trail, err := d.Store.GetTrailByLesson(ctx, lp.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "aula sem trilha"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao carregar trilha"})
		return
	}
	if _, err := d.ensureTrailCodigo(ctx, trail); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao publicar trilha"})
		return
	}

	row, err := d.Store.AttachTurmaLesson(ctx, store.AttachTurmaLessonParams{
		TurmaID: turma.ID, LessonPlanID: lp.ID,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "aula já está nesta turma"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao anexar aula"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"lesson_plan_id": row.LessonPlanID, "ordem": row.Ordem})
}

// strconvParseID parses a base-10 int64 path parameter.
func strconvParseID(s string) (int64, error) {
	return parseInt64(s)
}

var _ = context.Background
```

If a `parseInt64` helper does not already exist in the package, add to the same file:

```go
import "strconv"

func parseInt64(s string) (int64, error) { return strconv.ParseInt(s, 10, 64) }
```

(Check first: `grep -rn "func parseInt64\|strconv.ParseInt" backend/internal/http/*.go`. If an equivalent exists, call it instead and drop the wrapper.)

- [ ] **Step 5: Registrar a rota**

Em `backend/internal/http/router.go`, após a linha `r.Post("/turmas/{id}/students/import", d.importStudents)` adicionar:

```go
			r.Post("/turmas/{id}/lessons", d.attachTurmaLesson)
```

- [ ] **Step 6: Rodar o teste (deve passar)**

Run: `cd backend && TEST_DATABASE_URL=postgres://planejai:planejai@localhost:5433/planejai?sslmode=disable go test ./internal/http/ -run TestAttachTurmaLesson -v`
Expected: PASS.

- [ ] **Step 7: Garantir que os testes de publish continuam passando**

Run: `cd backend && TEST_DATABASE_URL=postgres://planejai:planejai@localhost:5433/planejai?sslmode=disable go test ./internal/http/ -run TestPublish -v`
Expected: PASS (refactor de `ensureTrailCodigo` não quebrou nada).

- [ ] **Step 8: Commit**

```bash
git add backend/internal/http/turma_lessons_handlers.go backend/internal/http/turma_lessons_handlers_test.go backend/internal/http/lesson_handlers.go backend/internal/http/router.go
git commit -m "feat(turma): attach lesson endpoint + ensureTrailCodigo helper"
```

---

### Task 4: Aulas na resposta de `GET /api/turmas/{id}`

**Files:**
- Modify: `backend/internal/http/turma_handlers.go:81-96` (`getTurma`)
- Test: `backend/internal/http/turma_lessons_handlers_test.go` (adicionar `TestGetTurmaIncludesAulas`)

**Interfaces:**
- Consumes: `store.ListTurmaLessons`.
- Produces: resposta de `getTurma` ganha campo `"aulas": [{lesson_plan_id, ordem, status, codigo, label}]`.

- [ ] **Step 1: Escrever o teste (falha)**

Adicionar em `backend/internal/http/turma_lessons_handlers_test.go`:

```go
func TestGetTurmaIncludesAulas(t *testing.T) {
	d := testDeps(t)
	r := NewRouter(d)
	cookie := loginProfessor(t, d, "getaulas@t.com")
	uid := userIDFromCookie(t, d, cookie)
	turmaID := seedTurma(t, d, uid, "6A-getaulas")
	lessonID := seedReadyLesson(t, d, uid)

	body, _ := json.Marshal(map[string]any{"lesson_plan_id": lessonID})
	att := httptest.NewRequest("POST", "/api/turmas/"+itoa(turmaID)+"/lessons", bytes.NewReader(body))
	att.AddCookie(cookie)
	r.ServeHTTP(httptest.NewRecorder(), att)

	req := httptest.NewRequest("GET", "/api/turmas/"+itoa(turmaID), nil)
	req.AddCookie(cookie)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("get turma status=%d body=%s", w.Code, w.Body)
	}
	var resp struct {
		Aulas []struct {
			LessonPlanID int64  `json:"lesson_plan_id"`
			Ordem        int    `json:"ordem"`
			Status       string `json:"status"`
			Label        string `json:"label"`
		} `json:"aulas"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.Aulas) != 1 || resp.Aulas[0].LessonPlanID != lessonID {
		t.Fatalf("aulas=%+v, want 1 with lesson=%d", resp.Aulas, lessonID)
	}
}
```

- [ ] **Step 2: Rodar (deve falhar)**

Run: `cd backend && TEST_DATABASE_URL=postgres://planejai:planejai@localhost:5433/planejai?sslmode=disable go test ./internal/http/ -run TestGetTurmaIncludesAulas -v`
Expected: FAIL — `resp.Aulas` vazio (campo ainda não existe).

- [ ] **Step 3: Estender `getTurma`**

Em `backend/internal/http/turma_handlers.go`, no fim de `getTurma`, antes do `writeJSON` final, buscar as aulas e incluí-las. Substituir a linha `writeJSON(w, 200, map[string]any{"turma": turma, "alunos": pub})` por:

```go
	aulas, err := d.Store.ListTurmaLessons(r.Context(), turma.ID)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": "erro ao carregar aulas"})
		return
	}
	if aulas == nil {
		aulas = []store.ListTurmaLessonsRow{}
	}
	writeJSON(w, 200, map[string]any{"turma": turma, "alunos": pub, "aulas": aulas})
```

(O tipo `ListTurmaLessonsRow` já tem tags JSON `lesson_plan_id, ordem, status, codigo, label` geradas pelo sqlc.)

- [ ] **Step 4: Rodar (deve passar)**

Run: `cd backend && TEST_DATABASE_URL=postgres://planejai:planejai@localhost:5433/planejai?sslmode=disable go test ./internal/http/ -run TestGetTurmaIncludesAulas -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/http/turma_handlers.go backend/internal/http/turma_lessons_handlers_test.go
git commit -m "feat(turma): include ordered aulas in GET /api/turmas/:id"
```

---

### Task 5: Endpoint remover aula (detach + renumber)

**Files:**
- Modify: `backend/internal/http/turma_lessons_handlers.go` (novo handler `detachTurmaLesson`)
- Modify: `backend/internal/http/router.go` (rota)
- Test: `backend/internal/http/turma_lessons_handlers_test.go` (`TestDetachTurmaLessonRenumbers`)

**Interfaces:**
- Consumes: `store.DetachTurmaLesson`, `store.RenumberTurmaLessons` (numa transação).
- Produces: `func (d Deps) detachTurmaLesson(...)` → `DELETE /api/turmas/{id}/lessons/{lessonId}` → `204`.

- [ ] **Step 1: Escrever o teste (falha)**

Adicionar em `turma_lessons_handlers_test.go`:

```go
func TestDetachTurmaLessonRenumbers(t *testing.T) {
	d := testDeps(t)
	r := NewRouter(d)
	cookie := loginProfessor(t, d, "detach@t.com")
	uid := userIDFromCookie(t, d, cookie)
	turmaID := seedTurma(t, d, uid, "6A-detach")
	l1 := seedReadyLesson(t, d, uid)
	l2 := seedReadyLesson(t, d, uid)
	l3 := seedReadyLesson(t, d, uid)
	for _, id := range []int64{l1, l2, l3} {
		b, _ := json.Marshal(map[string]any{"lesson_plan_id": id})
		req := httptest.NewRequest("POST", "/api/turmas/"+itoa(turmaID)+"/lessons", bytes.NewReader(b))
		req.AddCookie(cookie)
		r.ServeHTTP(httptest.NewRecorder(), req)
	}

	// remove a do meio (l2)
	del := httptest.NewRequest("DELETE", "/api/turmas/"+itoa(turmaID)+"/lessons/"+itoa(l2), nil)
	del.AddCookie(cookie)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, del)
	if w.Code != http.StatusNoContent {
		t.Fatalf("detach status=%d", w.Code)
	}

	rows, err := d.Store.ListTurmaLessons(context.Background(), turmaID)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[0].Ordem != 1 || rows[1].Ordem != 2 {
		t.Fatalf("after detach rows=%+v, want ordens 1,2 contíguas", rows)
	}
	if rows[0].LessonPlanID != l1 || rows[1].LessonPlanID != l3 {
		t.Fatalf("wrong remaining lessons: %+v", rows)
	}
}
```

- [ ] **Step 2: Rodar (deve falhar)**

Run: `cd backend && TEST_DATABASE_URL=postgres://planejai:planejai@localhost:5433/planejai?sslmode=disable go test ./internal/http/ -run TestDetachTurmaLessonRenumbers -v`
Expected: FAIL — rota 404 / `detachTurmaLesson` indefinido.

- [ ] **Step 3: Escrever o handler**

Adicionar em `turma_lessons_handlers.go`:

```go
func (d Deps) detachTurmaLesson(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "não autenticado"})
		return
	}
	turma, ok := d.ownedTurma(w, r, userID)
	if !ok {
		return
	}
	lessonID, err := parseInt64(chi.URLParam(r, "lessonId"))
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "aula não encontrada"})
		return
	}
	ctx := r.Context()
	tx, err := d.Pool.Begin(ctx)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro"})
		return
	}
	defer tx.Rollback(ctx)
	q := d.Store.WithTx(tx)
	if err := q.DetachTurmaLesson(ctx, store.DetachTurmaLessonParams{TurmaID: turma.ID, LessonPlanID: lessonID}); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao remover aula"})
		return
	}
	if err := q.RenumberTurmaLessons(ctx, turma.ID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao reordenar"})
		return
	}
	if err := tx.Commit(ctx); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao salvar"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
```

- [ ] **Step 4: Registrar a rota**

Em `router.go`, após `r.Post("/turmas/{id}/lessons", d.attachTurmaLesson)`:

```go
			r.Delete("/turmas/{id}/lessons/{lessonId}", d.detachTurmaLesson)
```

- [ ] **Step 5: Rodar (deve passar)**

Run: `cd backend && TEST_DATABASE_URL=postgres://planejai:planejai@localhost:5433/planejai?sslmode=disable go test ./internal/http/ -run TestDetachTurmaLessonRenumbers -v`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/http/turma_lessons_handlers.go backend/internal/http/router.go backend/internal/http/turma_lessons_handlers_test.go
git commit -m "feat(turma): detach lesson endpoint with renumber"
```

---

### Task 6: Endpoint reordenar (`PATCH .../lessons`)

**Files:**
- Modify: `backend/internal/http/turma_lessons_handlers.go` (handler `reorderTurmaLessons`)
- Modify: `backend/internal/http/router.go` (rota)
- Test: `backend/internal/http/turma_lessons_handlers_test.go` (`TestReorderTurmaLessons`)

**Interfaces:**
- Consumes: `store.ListTurmaLessons`, `store.SetTurmaLessonOrder`.
- Produces: `func (d Deps) reorderTurmaLessons(...)` → `PATCH /api/turmas/{id}/lessons`, body `{"ordered_ids":[<lesson_plan_id>,...]}` → `204`. Rejeita (400) se o conjunto não bate com as aulas atuais.

- [ ] **Step 1: Escrever o teste (falha)**

Adicionar:

```go
func TestReorderTurmaLessons(t *testing.T) {
	d := testDeps(t)
	r := NewRouter(d)
	cookie := loginProfessor(t, d, "reorder@t.com")
	uid := userIDFromCookie(t, d, cookie)
	turmaID := seedTurma(t, d, uid, "6A-reorder")
	l1 := seedReadyLesson(t, d, uid)
	l2 := seedReadyLesson(t, d, uid)
	for _, id := range []int64{l1, l2} {
		b, _ := json.Marshal(map[string]any{"lesson_plan_id": id})
		req := httptest.NewRequest("POST", "/api/turmas/"+itoa(turmaID)+"/lessons", bytes.NewReader(b))
		req.AddCookie(cookie)
		r.ServeHTTP(httptest.NewRecorder(), req)
	}

	// inverter ordem
	body, _ := json.Marshal(map[string]any{"ordered_ids": []int64{l2, l1}})
	req := httptest.NewRequest("PATCH", "/api/turmas/"+itoa(turmaID)+"/lessons", bytes.NewReader(body))
	req.AddCookie(cookie)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("reorder status=%d body=%s", w.Code, w.Body)
	}
	rows, _ := d.Store.ListTurmaLessons(context.Background(), turmaID)
	if rows[0].LessonPlanID != l2 || rows[1].LessonPlanID != l1 {
		t.Fatalf("order not applied: %+v", rows)
	}

	// conjunto que não bate → 400
	bad, _ := json.Marshal(map[string]any{"ordered_ids": []int64{l1}})
	reqBad := httptest.NewRequest("PATCH", "/api/turmas/"+itoa(turmaID)+"/lessons", bytes.NewReader(bad))
	reqBad.AddCookie(cookie)
	wBad := httptest.NewRecorder()
	r.ServeHTTP(wBad, reqBad)
	if wBad.Code != http.StatusBadRequest {
		t.Fatalf("mismatched set status=%d, want 400", wBad.Code)
	}
}
```

- [ ] **Step 2: Rodar (deve falhar)**

Run: `cd backend && TEST_DATABASE_URL=postgres://planejai:planejai@localhost:5433/planejai?sslmode=disable go test ./internal/http/ -run TestReorderTurmaLessons -v`
Expected: FAIL — rota inexistente.

- [ ] **Step 3: Escrever o handler**

Adicionar em `turma_lessons_handlers.go`:

```go
func (d Deps) reorderTurmaLessons(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "não autenticado"})
		return
	}
	turma, ok := d.ownedTurma(w, r, userID)
	if !ok {
		return
	}
	var in struct {
		OrderedIDs []int64 `json:"ordered_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "json inválido"})
		return
	}
	ctx := r.Context()
	current, err := d.Store.ListTurmaLessons(ctx, turma.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro"})
		return
	}
	// O conjunto enviado deve ser exatamente o conjunto atual (mesmos ids).
	want := map[int64]bool{}
	for _, row := range current {
		want[row.LessonPlanID] = true
	}
	if len(in.OrderedIDs) != len(current) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "lista de aulas não confere"})
		return
	}
	seen := map[int64]bool{}
	for _, id := range in.OrderedIDs {
		if !want[id] || seen[id] {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "lista de aulas não confere"})
			return
		}
		seen[id] = true
	}
	if err := d.Store.SetTurmaLessonOrder(ctx, store.SetTurmaLessonOrderParams{
		TurmaID: turma.ID, LessonPlanIds: in.OrderedIDs,
	}); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao reordenar"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
```

(Confirmar o nome do campo gerado pelo sqlc para o array: `LessonPlanIds`. Se o sqlc gerar `LessonPlanIds` diferente, ajustar a struct `SetTurmaLessonOrderParams` conforme `turma_lessons.sql.go`.)

- [ ] **Step 4: Registrar a rota**

Em `router.go`, após a rota de delete:

```go
			r.Patch("/turmas/{id}/lessons", d.reorderTurmaLessons)
```

- [ ] **Step 5: Rodar (deve passar)**

Run: `cd backend && TEST_DATABASE_URL=postgres://planejai:planejai@localhost:5433/planejai?sslmode=disable go test ./internal/http/ -run TestReorderTurmaLessons -v`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/http/turma_lessons_handlers.go backend/internal/http/router.go backend/internal/http/turma_lessons_handlers_test.go
git commit -m "feat(turma): reorder lessons endpoint"
```

---

### Task 7: Endpoint do aluno `GET /api/student/lessons`

**Files:**
- Create: `backend/internal/http/student_lessons_handlers.go`
- Modify: `backend/internal/http/router.go:47-49` (grupo `RequireStudent`)
- Test: `backend/internal/http/student_lessons_handlers_test.go`

**Interfaces:**
- Consumes: `store.GetStudent` (→ `TurmaID`), `store.ListStudentTurmaLessons`.
- Produces: `func (d Deps) studentLessons(...)` → `GET /api/student/lessons` → `200 {"aulas":[{ordem,label,codigo,concluido,pontos,unlocked}]}`. `unlocked` calculado em Go: ordem 1 sempre; ordem i liberada se a anterior `concluido`.

- [ ] **Step 1: Escrever o teste (falha)**

Create `backend/internal/http/student_lessons_handlers_test.go`:

```go
package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Carlos-hub/planejai/backend/internal/auth"
	"github.com/Carlos-hub/planejai/backend/internal/store"
)

// loginStudent creates a student in turmaID and returns their session cookie.
func loginStudent(t *testing.T, d Deps, turmaID int64, usuario string) (*http.Cookie, int64) {
	t.Helper()
	hash, _ := auth.HashPassword("senha123")
	s, err := d.Store.CreateStudent(context.Background(), store.CreateStudentParams{
		TurmaID: turmaID, Nome: "Aluno", Usuario: usuario, SenhaHash: hash,
	})
	if err != nil {
		t.Fatal(err)
	}
	sid, _ := auth.NewSessionID()
	_, err = d.Store.CreateStudentSession(context.Background(), store.CreateStudentSessionParams{
		ID: sid, StudentID: s.ID, ExpiresAt: pgTime(timePlus24h()),
	})
	if err != nil {
		t.Fatal(err)
	}
	return &http.Cookie{Name: "student_sid", Value: sid}, s.ID
}

func TestStudentLessonsGating(t *testing.T) {
	d := testDeps(t)
	r := NewRouter(d)
	cookie := loginProfessor(t, d, "stlessons@t.com")
	uid := userIDFromCookie(t, d, cookie)
	turmaID := seedTurma(t, d, uid, "6A-stlessons")
	l1 := seedReadyLesson(t, d, uid)
	l2 := seedReadyLesson(t, d, uid)
	for _, id := range []int64{l1, l2} {
		b, _ := json.Marshal(map[string]any{"lesson_plan_id": id})
		req := httptest.NewRequest("POST", "/api/turmas/"+itoa(turmaID)+"/lessons", bytes.NewReader(b))
		req.AddCookie(cookie)
		r.ServeHTTP(httptest.NewRecorder(), req)
	}

	stCookie, studentID := loginStudent(t, d, turmaID, "aluno-stlessons")

	// antes de concluir nada: aula 1 unlocked, aula 2 locked
	aulas := fetchStudentAulas(t, r, stCookie)
	if len(aulas) != 2 || !aulas[0].Unlocked || aulas[1].Unlocked {
		t.Fatalf("initial gating wrong: %+v", aulas)
	}

	// concluir a aula 1: attempt com concluido_em na trilha de l1
	trail1, _ := d.Store.GetTrailByLesson(context.Background(), l1)
	att, _ := d.Store.CreateAttempt(context.Background(), store.CreateAttemptParams{
		StudyTrailID: trail1.ID, NomeAluno: "Aluno", StudentID: &studentID,
	})
	if _, err := d.Store.CompleteAttempt(context.Background(), store.CompleteAttemptParams{ID: att.ID, Pontos: 3}); err != nil {
		t.Fatal(err)
	}

	// agora aula 2 libera
	aulas = fetchStudentAulas(t, r, stCookie)
	if !aulas[1].Unlocked || !aulas[0].Concluido {
		t.Fatalf("after completion gating wrong: %+v", aulas)
	}
}

type studentAula struct {
	Ordem     int    `json:"ordem"`
	Unlocked  bool   `json:"unlocked"`
	Concluido bool   `json:"concluido"`
	Codigo    string `json:"codigo"`
}

func fetchStudentAulas(t *testing.T, r http.Handler, c *http.Cookie) []studentAula {
	t.Helper()
	req := httptest.NewRequest("GET", "/api/student/lessons", nil)
	req.AddCookie(c)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("student lessons status=%d body=%s", w.Code, w.Body)
	}
	var resp struct {
		Aulas []studentAula `json:"aulas"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	return resp.Aulas
}
```

Note (confirmado): `pgTime`, `timePlus24h`, `itoa`, `GetSession`, `CreateStudentSession` já existem; cookie do aluno é `student_sid`. `parseInt64` NÃO existe no pacote — usar `strconv.ParseInt(s, 10, 64)` inline (como o restante do código) ou adicionar o wrapper indicado na Task 3.

- [ ] **Step 2: Rodar (deve falhar)**

Run: `cd backend && TEST_DATABASE_URL=postgres://planejai:planejai@localhost:5433/planejai?sslmode=disable go test ./internal/http/ -run TestStudentLessonsGating -v`
Expected: FAIL — rota `/api/student/lessons` inexistente (404) ou `studentLessons` indefinido.

- [ ] **Step 3: Escrever o handler**

Create `backend/internal/http/student_lessons_handlers.go`:

```go
package http

import (
	"net/http"

	"github.com/Carlos-hub/planejai/backend/internal/store"
)

type studentAulaResponse struct {
	Ordem     int32  `json:"ordem"`
	Label     string `json:"label"`
	Codigo    string `json:"codigo"`
	Concluido bool   `json:"concluido"`
	Pontos    int32  `json:"pontos"`
	Unlocked  bool   `json:"unlocked"`
}

func (d Deps) studentLessons(w http.ResponseWriter, r *http.Request) {
	studentID, ok := studentIDFromContext(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "não autenticado"})
		return
	}
	ctx := r.Context()
	student, err := d.Store.GetStudent(ctx, studentID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro"})
		return
	}
	rows, err := d.Store.ListStudentTurmaLessons(ctx, store.ListStudentTurmaLessonsParams{
		TurmaID: student.TurmaID, StudentID: studentID,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao carregar aulas"})
		return
	}
	aulas := make([]studentAulaResponse, 0, len(rows))
	prevConcluido := true // a primeira aula está sempre liberada
	for _, row := range rows {
		codigo := ""
		if row.Codigo != nil {
			codigo = *row.Codigo
		}
		aulas = append(aulas, studentAulaResponse{
			Ordem:     row.Ordem,
			Label:     row.Label,
			Codigo:    codigo,
			Concluido: row.Concluido,
			Pontos:    row.Pontos,
			Unlocked:  prevConcluido,
		})
		prevConcluido = row.Concluido
	}
	writeJSON(w, http.StatusOK, map[string]any{"aulas": aulas})
}
```

- [ ] **Step 4: Registrar a rota**

Em `router.go`, dentro do grupo com `r.Use(d.RequireStudent)` (junto de `/student/password`):

```go
			r.Get("/student/lessons", d.studentLessons)
```

- [ ] **Step 5: Rodar (deve passar)**

Run: `cd backend && TEST_DATABASE_URL=postgres://planejai:planejai@localhost:5433/planejai?sslmode=disable go test ./internal/http/ -run TestStudentLessonsGating -v`
Expected: PASS.

- [ ] **Step 6: Rodar toda a suíte http**

Run: `cd backend && TEST_DATABASE_URL=postgres://planejai:planejai@localhost:5433/planejai?sslmode=disable go test ./internal/http/ -v`
Expected: PASS (nenhuma regressão).

- [ ] **Step 7: Commit**

```bash
git add backend/internal/http/student_lessons_handlers.go backend/internal/http/router.go backend/internal/http/student_lessons_handlers_test.go
git commit -m "feat(student): GET /api/student/lessons with gated unlock"
```

---

### Task 8: Frontend — tipos + funções de API

**Files:**
- Modify: `frontend/lib/types.ts`
- Modify: `frontend/lib/api.ts`

**Interfaces:**
- Produces (TS):
  - `TurmaAula { lesson_plan_id: number; ordem: number; status: string; codigo: string | null; label: string }`
  - `StudentAula { ordem: number; label: string; codigo: string; concluido: boolean; pontos: number; unlocked: boolean }`
  - `getTurma` passa a retornar também `aulas: TurmaAula[]`.
  - `attachTurmaLesson(turmaId, lessonPlanId)`, `detachTurmaLesson(turmaId, lessonPlanId)`, `reorderTurmaLessons(turmaId, orderedIds)`, `listLessons()`, `getStudentLessons()`.

- [ ] **Step 1: Adicionar tipos**

Em `frontend/lib/types.ts`, adicionar:

```ts
export interface TurmaAula {
  lesson_plan_id: number;
  ordem: number;
  status: string;
  codigo: string | null;
  label: string;
}

export interface StudentAula {
  ordem: number;
  label: string;
  codigo: string;
  concluido: boolean;
  pontos: number;
  unlocked: boolean;
}
```

- [ ] **Step 2: Estender `getTurma` + novas funções**

Em `frontend/lib/api.ts`, substituir o `getTurma` existente por:

```ts
export const getTurma = (id: number) =>
  apiFetch<{
    turma: Turma;
    alunos: { id: number; nome: string; usuario: string; matricula: string | null }[];
    aulas: TurmaAula[];
  }>(`/api/turmas/${id}`);
```

E adicionar ao fim do arquivo:

```ts
export const listLessons = () => apiFetch<LessonSummary[]>("/api/lessons");

export const attachTurmaLesson = (turmaId: number, lessonPlanId: number) =>
  apiFetch<{ lesson_plan_id: number; ordem: number }>(
    `/api/turmas/${turmaId}/lessons`,
    { method: "POST", body: JSON.stringify({ lesson_plan_id: lessonPlanId }) }
  );

export const detachTurmaLesson = (turmaId: number, lessonPlanId: number) =>
  apiFetch<void>(`/api/turmas/${turmaId}/lessons/${lessonPlanId}`, {
    method: "DELETE",
  });

export const reorderTurmaLessons = (turmaId: number, orderedIds: number[]) =>
  apiFetch<void>(`/api/turmas/${turmaId}/lessons`, {
    method: "PATCH",
    body: JSON.stringify({ ordered_ids: orderedIds }),
  });

export const getStudentLessons = () =>
  apiFetch<{ aulas: StudentAula[] }>("/api/student/lessons");
```

Garantir os imports de tipo no topo de `api.ts` (`TurmaAula`, `StudentAula`, `LessonSummary`, `Turma` — adicionar os que faltarem à linha de import de `@/lib/types`).

- [ ] **Step 3: Typecheck**

Run: `cd frontend && npx tsc --noEmit`
Expected: sem erros.

- [ ] **Step 4: Commit**

```bash
git add frontend/lib/types.ts frontend/lib/api.ts
git commit -m "feat(web): api client + types for turma aulas and student lessons"
```

---

### Task 9: Frontend — seção "Aulas da turma" (professor)

**Files:**
- Create: `frontend/components/turma-aulas.tsx`
- Modify: `frontend/app/(teacher)/turmas/[id]/page.tsx` (renderizar a seção + carregar `aulas`)

**Interfaces:**
- Consumes: `getTurma`, `listLessons`, `attachTurmaLesson`, `detachTurmaLesson`, `reorderTurmaLessons`, tipos `TurmaAula`, `LessonSummary`.
- Produces: componente `<TurmaAulas turmaId aulas onChanged />` que lista, reordena (▲▼), remove e adiciona aulas.

- [ ] **Step 1: Criar o componente**

Create `frontend/components/turma-aulas.tsx`:

```tsx
"use client";

import { useState } from "react";
import { ChevronUp, ChevronDown, Trash2, Plus, Lock } from "lucide-react";
import type { TurmaAula, LessonSummary } from "@/lib/types";
import {
  listLessons,
  attachTurmaLesson,
  detachTurmaLesson,
  reorderTurmaLessons,
} from "@/lib/api";
import { Button } from "@/components/ui/button";

export function TurmaAulas({
  turmaId,
  aulas,
  onChanged,
}: {
  turmaId: number;
  aulas: TurmaAula[];
  onChanged: () => void;
}) {
  const [picking, setPicking] = useState(false);
  const [options, setOptions] = useState<LessonSummary[] | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function openPicker() {
    setError(null);
    setPicking(true);
    try {
      const all = await listLessons();
      const inTurma = new Set(aulas.map((a) => a.lesson_plan_id));
      setOptions(all.filter((l) => l.status === "pronto" && !inTurma.has(l.id)));
    } catch {
      setError("Não foi possível carregar suas aulas.");
    }
  }

  async function add(lessonId: number) {
    setBusy(true);
    setError(null);
    try {
      await attachTurmaLesson(turmaId, lessonId);
      setPicking(false);
      setOptions(null);
      onChanged();
    } catch {
      setError("Não foi possível adicionar a aula.");
    } finally {
      setBusy(false);
    }
  }

  async function remove(lessonId: number) {
    setBusy(true);
    try {
      await detachTurmaLesson(turmaId, lessonId);
      onChanged();
    } finally {
      setBusy(false);
    }
  }

  async function move(index: number, dir: -1 | 1) {
    const next = index + dir;
    if (next < 0 || next >= aulas.length) return;
    const ids = aulas.map((a) => a.lesson_plan_id);
    [ids[index], ids[next]] = [ids[next], ids[index]];
    setBusy(true);
    try {
      await reorderTurmaLessons(turmaId, ids);
      onChanged();
    } finally {
      setBusy(false);
    }
  }

  return (
    <section className="flex flex-col gap-4">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-lg font-semibold">Aulas da turma</h2>
          <p className="text-sm text-muted-foreground">
            O aluno percorre as aulas em ordem — a próxima libera ao concluir a anterior.
          </p>
        </div>
        <Button type="button" size="sm" onClick={openPicker} disabled={busy}>
          <Plus className="size-4" />
          Adicionar aula
        </Button>
      </div>

      {error && (
        <p className="text-sm text-destructive" role="alert">
          {error}
        </p>
      )}

      {aulas.length === 0 ? (
        <p className="rounded-lg border border-dashed bg-muted/30 px-4 py-6 text-center text-sm text-muted-foreground">
          Nenhuma aula ainda. Adicione a primeira aula da sequência.
        </p>
      ) : (
        <ol className="flex flex-col gap-2">
          {aulas.map((a, i) => (
            <li
              key={a.lesson_plan_id}
              className="flex items-center gap-3 rounded-xl border bg-card p-3"
            >
              <span className="flex size-7 shrink-0 items-center justify-center rounded-full bg-primary/10 text-sm font-semibold text-primary">
                {a.ordem}
              </span>
              <span className="flex-1 truncate text-sm font-medium">{a.label}</span>
              <div className="flex items-center gap-1">
                <button
                  type="button"
                  aria-label="Mover para cima"
                  disabled={busy || i === 0}
                  onClick={() => move(i, -1)}
                  className="flex size-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted disabled:opacity-30"
                >
                  <ChevronUp className="size-4" />
                </button>
                <button
                  type="button"
                  aria-label="Mover para baixo"
                  disabled={busy || i === aulas.length - 1}
                  onClick={() => move(i, 1)}
                  className="flex size-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted disabled:opacity-30"
                >
                  <ChevronDown className="size-4" />
                </button>
                <button
                  type="button"
                  aria-label="Remover aula"
                  disabled={busy}
                  onClick={() => remove(a.lesson_plan_id)}
                  className="flex size-8 items-center justify-center rounded-md text-muted-foreground hover:bg-destructive/10 hover:text-destructive disabled:opacity-30"
                >
                  <Trash2 className="size-4" />
                </button>
              </div>
            </li>
          ))}
        </ol>
      )}

      {picking && (
        <div className="rounded-xl border bg-card p-3">
          <div className="mb-2 flex items-center justify-between">
            <span className="text-sm font-medium">Escolha uma aula pronta</span>
            <Button type="button" variant="ghost" size="sm" onClick={() => setPicking(false)}>
              Fechar
            </Button>
          </div>
          {options === null ? (
            <p className="text-sm text-muted-foreground">Carregando…</p>
          ) : options.length === 0 ? (
            <p className="flex items-center gap-2 text-sm text-muted-foreground">
              <Lock className="size-4" />
              Nenhuma aula pronta disponível. Crie e finalize uma aula primeiro.
            </p>
          ) : (
            <ul className="flex flex-col gap-1">
              {options.map((l) => (
                <li key={l.id}>
                  <button
                    type="button"
                    disabled={busy}
                    onClick={() => add(l.id)}
                    className="flex w-full items-center justify-between rounded-lg px-3 py-2 text-left text-sm hover:bg-muted disabled:opacity-50"
                  >
                    <span>Aula #{l.id}</span>
                    <span className="text-xs text-muted-foreground">
                      {l.duracao} min · {l.origem}
                    </span>
                  </button>
                </li>
              ))}
            </ul>
          )}
        </div>
      )}
    </section>
  );
}
```

- [ ] **Step 2: Renderizar na página da turma**

Em `frontend/app/(teacher)/turmas/[id]/page.tsx`:
1. Importar: `import { TurmaAulas } from "@/components/turma-aulas";` e `import type { TurmaAula } from "@/lib/types";`.
2. Adicionar estado `const [aulas, setAulas] = useState<TurmaAula[]>([]);`.
3. Onde a página consome o retorno de `getTurma(...)`, guardar as aulas: `setAulas(data.aulas ?? []);` (usar o mesmo callback de load; expor esse load como função reutilizável, ex. `reload`).
4. Renderizar o componente (ex.: acima ou abaixo da seção de alunos):

```tsx
<TurmaAulas turmaId={turmaId} aulas={aulas} onChanged={reload} />
```

onde `turmaId` é o id numérico da turma da rota e `reload` re-executa `getTurma` e atualiza `aulas` (e alunos).

- [ ] **Step 3: Typecheck**

Run: `cd frontend && npx tsc --noEmit`
Expected: sem erros.

- [ ] **Step 4: Verificação manual**

```bash
docker compose up -d --build web api
```
No app: abrir uma turma → seção "Aulas da turma" → "Adicionar aula" lista as aulas Prontas → adicionar → aparece com ordem 1 → adicionar outra → mover ▲▼ reordena → remover tira e renumera.

- [ ] **Step 5: Commit**

```bash
git add frontend/components/turma-aulas.tsx "frontend/app/(teacher)/turmas/[id]/page.tsx"
git commit -m "feat(web): turma aulas section (add, reorder, remove)"
```

---

### Task 10: Frontend — home do aluno `/aluno` + redirect no login

**Files:**
- Create: `frontend/app/aluno/page.tsx`
- Modify: `frontend/app/aluno/login/page.tsx` (redirect default → `/aluno`)

**Interfaces:**
- Consumes: `getStudentLessons`, tipo `StudentAula`.
- Produces: página `/aluno` listando as aulas da turma com estado cadeado/liberada/concluída; login do aluno passa a mandar para `/aluno`.

- [ ] **Step 1: Criar a página do aluno**

Create `frontend/app/aluno/page.tsx`:

```tsx
"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Lock, ArrowRight, CheckCircle2 } from "lucide-react";
import { getStudentLessons } from "@/lib/api";
import type { StudentAula } from "@/lib/types";

export default function AlunoHomePage() {
  const [aulas, setAulas] = useState<StudentAula[] | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    getStudentLessons()
      .then((data) => setAulas(data.aulas))
      .catch(() => setError("Não foi possível carregar suas aulas."));
  }, []);

  return (
    <div className="mx-auto flex max-w-2xl flex-col gap-6 px-4 py-8">
      <header>
        <h1 className="text-2xl font-semibold tracking-tight">Minhas aulas</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Conclua cada aula para liberar a próxima.
        </p>
      </header>

      {error && (
        <p className="text-sm text-destructive" role="alert">
          {error}
        </p>
      )}

      {aulas === null && !error && (
        <p className="text-sm text-muted-foreground">Carregando…</p>
      )}

      {aulas && aulas.length === 0 && (
        <p className="rounded-lg border border-dashed bg-muted/30 px-4 py-6 text-center text-sm text-muted-foreground">
          Sua turma ainda não tem aulas. Volte mais tarde.
        </p>
      )}

      {aulas && aulas.length > 0 && (
        <ol className="flex flex-col gap-2">
          {aulas.map((a) => {
            const state = a.concluido
              ? "done"
              : a.unlocked
                ? "open"
                : "locked";
            const inner = (
              <div
                className={`flex items-center gap-3 rounded-xl border p-4 ${
                  state === "locked"
                    ? "bg-muted/40 opacity-70"
                    : "bg-card hover:border-primary/40"
                }`}
              >
                <span className="flex size-8 shrink-0 items-center justify-center rounded-full bg-primary/10 text-sm font-semibold text-primary">
                  {a.ordem}
                </span>
                <div className="flex-1">
                  <p className="text-sm font-medium">{a.label}</p>
                  {state === "done" && (
                    <p className="text-xs text-muted-foreground">
                      Concluída · {a.pontos} pontos
                    </p>
                  )}
                  {state === "locked" && (
                    <p className="text-xs text-muted-foreground">
                      Conclua a aula anterior para liberar.
                    </p>
                  )}
                </div>
                {state === "done" && (
                  <CheckCircle2 className="size-5 text-emerald-500" />
                )}
                {state === "open" && <ArrowRight className="size-5 text-primary" />}
                {state === "locked" && (
                  <Lock className="size-5 text-muted-foreground" />
                )}
              </div>
            );
            return (
              <li key={a.ordem}>
                {state === "locked" ? (
                  inner
                ) : (
                  <Link href={`/t/${a.codigo}`}>{inner}</Link>
                )}
              </li>
            );
          })}
        </ol>
      )}
    </div>
  );
}
```

- [ ] **Step 2: Redirecionar login do aluno para `/aluno`**

Em `frontend/app/aluno/login/page.tsx`, onde hoje o default do redirect é `/` (variável `next`), trocar o fallback para `/aluno`. Localizar a lógica que define `next` (comentário: “redirect off-site) — anything else defaults to '/'”) e ajustar o default de `"/"` para `"/aluno"`. Exemplo do trecho final:

```tsx
      const next = safeNext ?? "/aluno";
      router.push(next);
```

(Se a variável tiver outro nome, manter a lógica de sanitização e só trocar o fallback para `"/aluno"`.)

- [ ] **Step 3: Typecheck**

Run: `cd frontend && npx tsc --noEmit`
Expected: sem erros.

- [ ] **Step 4: Verificação manual (fluxo ponta a ponta)**

```bash
docker compose up -d --build web api
```
1. Professor: criar turma, adicionar 2 aulas prontas, importar/criar um aluno.
2. Deslogar / abrir `/aluno/login`, entrar como o aluno → cai em `/aluno`.
3. Aula 1 liberada (seta), aula 2 cadeado.
4. Abrir aula 1, responder e enviar o quiz em `/t/{codigo}`.
5. Voltar a `/aluno` → aula 1 com ✓ e pontos, aula 2 liberada.

- [ ] **Step 5: Commit**

```bash
git add frontend/app/aluno/page.tsx frontend/app/aluno/login/page.tsx
git commit -m "feat(web): student home /aluno with gated lesson sequence"
```

---

## Self-Review (preenchido)

**Cobertura do spec:**
- Tabela `turma_lessons` + backfill → Task 1 ✅
- Conclusão/unlock derivado de `student_attempts` → Task 7 (query LATERAL + cálculo em Go) ✅
- `GET /api/turmas/{id}` com aulas → Task 4 ✅
- Anexar (owned + pronto + garante codigo + UNIQUE) → Task 3 ✅
- Remover + renumerar → Task 5 ✅
- Reordenar atômico + validação de conjunto → Task 6 ✅
- `GET /api/student/lessons` → Task 7 ✅
- Seletor via `GET /api/lessons` → Task 9 ✅
- Seção professor (add/reorder/remove) → Task 9 ✅
- Home `/aluno` + redirect → Task 10 ✅
- Label = assunto BNCC / fallback → Task 2 (queries) ✅
- Testes Go (attach/detach/reorder/gating/ownership) → Tasks 3–7 ✅

**Consistência de tipos:** nomes de params/rows sqlc (`AttachTurmaLessonParams`, `ListTurmaLessonsRow`, `ListStudentTurmaLessonsParams`, `ListStudentTurmaLessonsRow`, `SetTurmaLessonOrderParams.LessonPlanIds`) são referenciados igualmente no backend; Tasks marcam “confirmar nome gerado” onde o sqlc pode variar (arrays, `LessonPlanIds`).

**Pontos a confirmar durante execução (anotados nos passos):** existência de `parseInt64`, `pgTime`, `timePlus24h`, `jsonMarshal`/`bytesReader`, `CreateStudentSession`, `GetSession` no pacote — usar o helper existente ou o fallback indicado.

**Riscos:** cookie do aluno tem nome real a confirmar (`student_sid` no teste é suposição — checar `middleware.go`/`studentLogin` no Step de teste da Task 7 e ajustar). Marcado como confirmação obrigatória antes de rodar o teste do aluno.
