# Turmas + Contas de Aluno (Subsistema A) — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add teacher-managed turmas, student accounts (login/senha created via CSV import), and turma-gated activities that tie each attempt to a logged-in student — while keeping the existing anonymous short-code flow for standalone trails.

**Architecture:** New Postgres tables (`turmas`, `students`, `student_sessions`) plus two nullable columns on existing tables (`study_trails.turma_id`, `student_attempts.student_id`). Go backend (chi + sqlc + pgx) gets student auth mirroring the existing professor auth, turma CRUD + CSV import handlers, and gating in the public trail flow. Next.js frontend gets a student login page and teacher turma-management pages.

**Tech Stack:** Go 1.23, chi/v5, sqlc (pgx/v5), bcrypt (`internal/auth`), PostgreSQL 16, goose migrations, Next.js/TypeScript/React.

## Global Constraints

- Reuse `internal/auth` for all password hashing/session-id generation — no new crypto. (`HashPassword`, `CheckPassword`, `NewSessionID`, `SessionTTL`).
- Professor session cookie is `sid`; student session cookie MUST be a distinct name `student_sid` with the same attributes (`Path:"/"`, `HttpOnly`, `SameSite=Lax`).
- sqlc config: `sql_package: pgx/v5`, `emit_pointers_for_null_types: true` — nullable columns become `*T` in Go; queries live in `internal/store/queries/*.sql`, regenerated with `sqlc generate` (run from `backend/`).
- Ownership checks return `404` (never `403`) for a turma owned by another professor, to avoid leaking existence — matching the codebase's non-disclosure style.
- Backend tests skip when `TEST_DATABASE_URL` is unset (see `internal/http/testmain_test.go`), and follow the existing `*_handlers_test.go` structure.
- Student `usuario` is globally UNIQUE. Initial password is returned in plaintext exactly once (import response); only `senha_hash` is persisted.
- All `go` commands run from `backend/`. All `npm` commands run from `frontend/`.

---

### Task 1: DB schema + queries + sqlc generate

**Files:**
- Create: `backend/migrations/00004_turmas_alunos.sql`
- Create: `backend/internal/store/queries/turmas.sql`
- Create: `backend/internal/store/queries/students.sql`
- Create: `backend/internal/store/queries/student_sessions.sql`
- Modify: `backend/internal/store/queries/attempts.sql` (CreateAttempt gains student_id)
- Modify: `backend/internal/store/queries/trails.sql` (add SetTrailTurma + turma-aware reads implicit via `SELECT *`)
- Regenerated (do not hand-edit): `backend/internal/store/*.sql.go`, `models.go`
- Test: `backend/internal/store/turmas_store_test.go`

**Interfaces:**
- Produces (sqlc-generated, consumed by later tasks):
  - `store.Turma{ID int64; UserID int64; Nome string; Etapa string; Anos []int32; CreatedAt pgtype.Timestamptz}`
  - `store.Student{ID int64; TurmaID int64; Nome string; Usuario string; SenhaHash string; Matricula *string; CreatedAt pgtype.Timestamptz}`
  - `store.StudentSession{ID string; StudentID int64; ExpiresAt pgtype.Timestamptz; CreatedAt pgtype.Timestamptz}`
  - Query methods: `CreateTurma(ctx, CreateTurmaParams) (Turma, error)`, `ListTurmasByUser(ctx, userID) ([]Turma, error)`, `GetTurma(ctx, id) (Turma, error)`, `UpdateTurma(ctx, UpdateTurmaParams) (Turma, error)`, `DeleteTurma(ctx, id) error`, `CreateStudent(ctx, CreateStudentParams) (Student, error)`, `ListStudentsByTurma(ctx, turmaID) ([]Student, error)`, `GetStudentByUsuario(ctx, usuario) (Student, error)`, `GetStudent(ctx, id) (Student, error)`, `UpdateStudentPassword(ctx, UpdateStudentPasswordParams) error`, `CreateStudentSession(ctx, CreateStudentSessionParams) (StudentSession, error)`, `GetStudentSession(ctx, id) (StudentSession, error)`, `DeleteStudentSession(ctx, id) error`, `SetTrailTurma(ctx, SetTrailTurmaParams) error`.
  - `store.StudyTrail` gains `TurmaID *int64`; `store.StudentAttempt` gains `StudentID *int64`; `store.CreateAttemptParams` gains `StudentID *int64`.

- [ ] **Step 1: Write the migration**

Create `backend/migrations/00004_turmas_alunos.sql`:

```sql
-- +goose Up
CREATE TABLE turmas (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  nome TEXT NOT NULL,
  etapa TEXT NOT NULL DEFAULT '',
  anos INT[] NOT NULL DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE students (
  id BIGSERIAL PRIMARY KEY,
  turma_id BIGINT NOT NULL REFERENCES turmas(id) ON DELETE CASCADE,
  nome TEXT NOT NULL,
  usuario TEXT UNIQUE NOT NULL,
  senha_hash TEXT NOT NULL,
  matricula TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE student_sessions (
  id TEXT PRIMARY KEY,
  student_id BIGINT NOT NULL REFERENCES students(id) ON DELETE CASCADE,
  expires_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE study_trails
  ADD COLUMN turma_id BIGINT REFERENCES turmas(id) ON DELETE SET NULL;

ALTER TABLE student_attempts
  ADD COLUMN student_id BIGINT REFERENCES students(id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE student_attempts DROP COLUMN student_id;
ALTER TABLE study_trails DROP COLUMN turma_id;
DROP TABLE student_sessions, students, turmas;
```

- [ ] **Step 2: Write the query files**

Create `backend/internal/store/queries/turmas.sql`:

```sql
-- name: CreateTurma :one
INSERT INTO turmas (user_id, nome, etapa, anos) VALUES ($1,$2,$3,$4) RETURNING *;
-- name: ListTurmasByUser :many
SELECT * FROM turmas WHERE user_id=$1 ORDER BY created_at DESC;
-- name: GetTurma :one
SELECT * FROM turmas WHERE id=$1;
-- name: UpdateTurma :one
UPDATE turmas SET nome=$2, etapa=$3, anos=$4 WHERE id=$1 RETURNING *;
-- name: DeleteTurma :exec
DELETE FROM turmas WHERE id=$1;
```

Create `backend/internal/store/queries/students.sql`:

```sql
-- name: CreateStudent :one
INSERT INTO students (turma_id, nome, usuario, senha_hash, matricula)
VALUES ($1,$2,$3,$4,$5) RETURNING *;
-- name: ListStudentsByTurma :many
SELECT * FROM students WHERE turma_id=$1 ORDER BY nome;
-- name: GetStudentByUsuario :one
SELECT * FROM students WHERE usuario=$1;
-- name: GetStudent :one
SELECT * FROM students WHERE id=$1;
-- name: UpdateStudentPassword :exec
UPDATE students SET senha_hash=$2 WHERE id=$1;
```

Create `backend/internal/store/queries/student_sessions.sql`:

```sql
-- name: CreateStudentSession :one
INSERT INTO student_sessions (id, student_id, expires_at) VALUES ($1,$2,$3) RETURNING *;
-- name: GetStudentSession :one
SELECT * FROM student_sessions WHERE id=$1 AND expires_at > now();
-- name: DeleteStudentSession :exec
DELETE FROM student_sessions WHERE id=$1;
```

- [ ] **Step 3: Modify attempts + trails queries**

In `backend/internal/store/queries/attempts.sql`, replace the `CreateAttempt` line:

```sql
-- name: CreateAttempt :one
INSERT INTO student_attempts (study_trail_id, nome_aluno, student_id) VALUES ($1,$2,$3) RETURNING *;
```

Append to `backend/internal/store/queries/trails.sql`:

```sql
-- name: SetTrailTurma :exec
UPDATE study_trails SET turma_id=$2 WHERE id=$1;
```

- [ ] **Step 4: Regenerate sqlc**

Run: `cd backend && sqlc generate`
Expected: no errors; `git status` shows modified `internal/store/models.go`, `attempts.sql.go`, `trails.sql.go` and new `turmas.sql.go`, `students.sql.go`, `student_sessions.sql.go`.

- [ ] **Step 5: Fix the existing CreateAttempt call site**

`internal/http/public_handlers.go` `startAttempt` now must pass `StudentID`. For now pass `nil` (anonymous). Change:

```go
	attempt, err := d.Store.CreateAttempt(ctx, store.CreateAttemptParams{
		StudyTrailID: trail.ID,
		NomeAluno:    req.Nome,
		StudentID:    nil,
	})
```

- [ ] **Step 6: Write a store integration test**

Create `backend/internal/store/turmas_store_test.go`:

```go
package store

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func storeTestPool(t *testing.T) *Queries {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return New(pool)
}

func TestCreateTurmaAndStudent(t *testing.T) {
	q := storeTestPool(t)
	ctx := context.Background()
	u, err := q.CreateUser(ctx, CreateUserParams{Email: "prof-turma@test.com", SenhaHash: "x", Nome: "Prof"})
	if err != nil {
		t.Fatal(err)
	}
	turma, err := q.CreateTurma(ctx, CreateTurmaParams{UserID: u.ID, Nome: "6A", Etapa: "EF", Anos: []int32{6}})
	if err != nil {
		t.Fatal(err)
	}
	if turma.Nome != "6A" {
		t.Fatalf("nome = %q", turma.Nome)
	}
	s, err := q.CreateStudent(ctx, CreateStudentParams{TurmaID: turma.ID, Nome: "Ana", Usuario: "ana.test.9z", SenhaHash: "h"})
	if err != nil {
		t.Fatal(err)
	}
	if s.TurmaID != turma.ID {
		t.Fatalf("turma_id = %d", s.TurmaID)
	}
}
```

- [ ] **Step 7: Build + test**

Run: `cd backend && go build ./... && go test ./internal/store/ -run TestCreateTurmaAndStudent -v`
Expected: build passes; test PASS (or SKIP if `TEST_DATABASE_URL` unset). Also run `go build ./...` to confirm the `startAttempt` edit compiles.

- [ ] **Step 8: Commit**

```bash
git add backend/migrations backend/internal/store backend/internal/http/public_handlers.go
git commit -m "feat(turmas): schema + sqlc queries for turmas, students, sessions"
```

---

### Task 2: Student credential generators

**Files:**
- Create: `backend/internal/auth/student.go`
- Test: `backend/internal/auth/student_test.go`

**Interfaces:**
- Produces:
  - `func UsernameSlug(nome string) string` — lowercase, accents stripped, spaces→`.`, only `[a-z0-9.]`.
  - `func RandomSuffix() (string, error)` — 4 chars from an unambiguous alphabet.
  - `func GenerateInitialPassword() (string, error)` — 8 chars from an unambiguous alphabet.
  - Uniqueness retry is done by the caller (Task 6), combining `UsernameSlug(nome) + "." + RandomSuffix()`.

- [ ] **Step 1: Write the failing test**

Create `backend/internal/auth/student_test.go`:

```go
package auth

import (
	"regexp"
	"strings"
	"testing"
)

func TestUsernameSlug(t *testing.T) {
	cases := map[string]string{
		"Ana Clara":      "ana.clara",
		"João da Silva":  "joao.da.silva",
		"MARIA":          "maria",
		"  Zé  Ninguém ": "ze.ninguem",
	}
	for in, want := range cases {
		if got := UsernameSlug(in); got != want {
			t.Errorf("UsernameSlug(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRandomSuffixAndPassword(t *testing.T) {
	ok := regexp.MustCompile(`^[a-hj-np-z2-9]+$`)
	suf, err := RandomSuffix()
	if err != nil || len(suf) != 4 || !ok.MatchString(suf) {
		t.Fatalf("suffix=%q err=%v", suf, err)
	}
	pw, err := GenerateInitialPassword()
	if err != nil || len(pw) != 8 || !ok.MatchString(pw) {
		t.Fatalf("pw=%q err=%v", pw, err)
	}
	if strings.ContainsAny(pw, "il1o0") {
		t.Fatalf("ambiguous chars in %q", pw)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/auth/ -run 'TestUsernameSlug|TestRandomSuffixAndPassword' -v`
Expected: FAIL / build error `undefined: UsernameSlug`.

- [ ] **Step 3: Write the implementation**

Create `backend/internal/auth/student.go`:

```go
package auth

import (
	"crypto/rand"
	"math/big"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// alphabet excludes ambiguous chars (i, l, 1, o, 0) for readability.
const credAlphabet = "abcdefghjkmnpqrstuvwxyz23456789"

// UsernameSlug normalizes a name into the stable part of a login: lowercase,
// accents removed, runs of non-alphanumerics collapsed to a single dot.
func UsernameSlug(nome string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	normalized, _, err := transform.String(t, nome)
	if err != nil {
		normalized = nome
	}
	normalized = strings.ToLower(normalized)
	var b strings.Builder
	lastDot := true // avoid leading dot
	for _, r := range normalized {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastDot = false
		} else if !lastDot {
			b.WriteByte('.')
			lastDot = true
		}
	}
	return strings.Trim(b.String(), ".")
}

func randomFrom(alphabet string, n int) (string, error) {
	b := make([]byte, n)
	max := big.NewInt(int64(len(alphabet)))
	for i := range b {
		idx, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		b[i] = alphabet[idx.Int64()]
	}
	return string(b), nil
}

// RandomSuffix returns a 4-char unambiguous suffix for username uniqueness.
func RandomSuffix() (string, error) { return randomFrom(credAlphabet, 4) }

// GenerateInitialPassword returns an 8-char unambiguous initial password.
func GenerateInitialPassword() (string, error) { return randomFrom(credAlphabet, 8) }
```

- [ ] **Step 4: Ensure the text module is available**

Run: `cd backend && go get golang.org/x/text && go mod tidy`
Expected: `golang.org/x/text` present in `go.mod`.

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd backend && go test ./internal/auth/ -v`
Expected: PASS (all auth tests, including existing).

- [ ] **Step 6: Commit**

```bash
git add backend/internal/auth backend/go.mod backend/go.sum
git commit -m "feat(turmas): student username/password generators"
```

---

### Task 3: Student session middleware + login/logout

**Files:**
- Create: `backend/internal/http/student_handlers.go`
- Modify: `backend/internal/http/middleware.go` (add `RequireStudent` + `studentIDFromContext`)
- Modify: `backend/internal/http/router.go` (register student routes)
- Test: `backend/internal/http/student_handlers_test.go`

**Interfaces:**
- Consumes: `store.GetStudentByUsuario`, `store.CreateStudentSession`, `store.GetStudentSession`, `store.DeleteStudentSession`, `auth.CheckPassword`, `auth.NewSessionID`, `auth.SessionTTL`.
- Produces: `func (d Deps) RequireStudent(next http.Handler) http.Handler`; `func studentIDFromContext(r *http.Request) (int64, bool)`; handlers `studentLogin`, `studentLogout`; helper `setStudentSessionCookie(w, sid, exp)`; cookie name constant `studentCookie = "student_sid"`.

- [ ] **Step 1: Write the failing test**

Create `backend/internal/http/student_handlers_test.go`:

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

// seedStudent creates a professor, turma, and one student with a known password.
func seedStudent(t *testing.T, d Deps, usuario, senha string) store.Student {
	t.Helper()
	ctx := context.Background()
	u, err := d.Store.CreateUser(ctx, store.CreateUserParams{Email: usuario + "-prof@t.com", SenhaHash: "x", Nome: "P"})
	if err != nil {
		t.Fatal(err)
	}
	turma, err := d.Store.CreateTurma(ctx, store.CreateTurmaParams{UserID: u.ID, Nome: "T"})
	if err != nil {
		t.Fatal(err)
	}
	hash, _ := auth.HashPassword(senha)
	s, err := d.Store.CreateStudent(ctx, store.CreateStudentParams{TurmaID: turma.ID, Nome: "Aluno", Usuario: usuario, SenhaHash: hash})
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestStudentLogin(t *testing.T) {
	d := testDeps(t)
	seedStudent(t, d, "aluno.login.aa", "segredo123")
	r := NewRouter(d)

	body, _ := json.Marshal(map[string]string{"usuario": "aluno.login.aa", "senha": "segredo123"})
	req := httptest.NewRequest("POST", "/api/student/login", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("login status = %d body=%s", w.Code, w.Body)
	}
	if len(w.Result().Cookies()) == 0 || w.Result().Cookies()[0].Name != "student_sid" {
		t.Fatalf("expected student_sid cookie")
	}

	bad, _ := json.Marshal(map[string]string{"usuario": "aluno.login.aa", "senha": "errada"})
	req2 := httptest.NewRequest("POST", "/api/student/login", bytes.NewReader(bad))
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusUnauthorized {
		t.Fatalf("bad login status = %d", w2.Code)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/http/ -run TestStudentLogin -v`
Expected: FAIL (route 404 / undefined handler).

- [ ] **Step 3: Add the middleware**

Append to `backend/internal/http/middleware.go`:

```go
const studentIDKey ctxKey = "studentID"

func (d Deps) RequireStudent(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("student_sid")
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "não autenticado"})
			return
		}
		sess, err := d.Store.GetStudentSession(r.Context(), c.Value)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "sessão inválida"})
			return
		}
		ctx := context.WithValue(r.Context(), studentIDKey, sess.StudentID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func studentIDFromContext(r *http.Request) (int64, bool) {
	v, ok := r.Context().Value(studentIDKey).(int64)
	return v, ok
}
```

- [ ] **Step 4: Add the handlers**

Create `backend/internal/http/student_handlers.go`:

```go
package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Carlos-hub/planejai/backend/internal/auth"
	"github.com/Carlos-hub/planejai/backend/internal/store"
)

const studentCookie = "student_sid"

func setStudentSessionCookie(w http.ResponseWriter, sid string, exp time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name: studentCookie, Value: sid, Path: "/", HttpOnly: true,
		SameSite: http.SameSiteLaxMode, Expires: exp,
	})
}

func (d Deps) studentLogin(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Usuario string `json:"usuario"`
		Senha   string `json:"senha"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, 400, map[string]string{"error": "json inválido"})
		return
	}
	s, err := d.Store.GetStudentByUsuario(r.Context(), in.Usuario)
	if err != nil || !auth.CheckPassword(s.SenhaHash, in.Senha) {
		writeJSON(w, 401, map[string]string{"error": "credenciais inválidas"})
		return
	}
	sid, _ := auth.NewSessionID()
	exp := time.Now().Add(auth.SessionTTL)
	if _, err := d.Store.CreateStudentSession(r.Context(), store.CreateStudentSessionParams{
		ID: sid, StudentID: s.ID, ExpiresAt: pgTime(exp),
	}); err != nil {
		writeJSON(w, 500, map[string]string{"error": "erro"})
		return
	}
	setStudentSessionCookie(w, sid, exp)
	writeJSON(w, 200, map[string]any{"id": s.ID, "nome": s.Nome, "usuario": s.Usuario, "turma_id": s.TurmaID})
}

func (d Deps) studentLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(studentCookie); err == nil {
		_ = d.Store.DeleteStudentSession(r.Context(), c.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: studentCookie, Value: "", Path: "/", MaxAge: -1})
	w.WriteHeader(204)
}
```

- [ ] **Step 5: Register routes**

In `backend/internal/http/router.go`, inside `r.Route("/api", ...)`, after the `r.Post("/attempts/{id}/answers", ...)` line and before the professor `r.Group`, add:

```go
		r.Post("/student/login", d.studentLogin)
		r.Post("/student/logout", d.studentLogout)
```

- [ ] **Step 6: Run tests to verify they pass**

Run: `cd backend && go test ./internal/http/ -run TestStudentLogin -v`
Expected: PASS (or SKIP without `TEST_DATABASE_URL`). Also `go build ./...` passes.

- [ ] **Step 7: Commit**

```bash
git add backend/internal/http/student_handlers.go backend/internal/http/middleware.go backend/internal/http/router.go backend/internal/http/student_handlers_test.go
git commit -m "feat(turmas): student login/logout + RequireStudent middleware"
```

---

### Task 4: Student password change

**Files:**
- Modify: `backend/internal/http/student_handlers.go` (add `studentChangePassword`)
- Modify: `backend/internal/http/router.go` (guarded route)
- Test: `backend/internal/http/student_handlers_test.go` (add case)

**Interfaces:**
- Consumes: `RequireStudent`, `studentIDFromContext`, `store.GetStudent`, `store.UpdateStudentPassword`, `auth.CheckPassword`, `auth.HashPassword`.
- Produces: handler `studentChangePassword` at `POST /api/student/password` (student-guarded).

- [ ] **Step 1: Write the failing test**

Add to `backend/internal/http/student_handlers_test.go`:

```go
func TestStudentChangePassword(t *testing.T) {
	d := testDeps(t)
	seedStudent(t, d, "aluno.pw.bb", "antiga123")
	r := NewRouter(d)

	// login to get cookie
	body, _ := json.Marshal(map[string]string{"usuario": "aluno.pw.bb", "senha": "antiga123"})
	lreq := httptest.NewRequest("POST", "/api/student/login", bytes.NewReader(body))
	lw := httptest.NewRecorder()
	r.ServeHTTP(lw, lreq)
	cookie := lw.Result().Cookies()[0]

	// wrong current password → 401
	ch, _ := json.Marshal(map[string]string{"senha_atual": "errada", "senha_nova": "nova12345"})
	req := httptest.NewRequest("POST", "/api/student/password", bytes.NewReader(ch))
	req.AddCookie(cookie)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("wrong-current status = %d", w.Code)
	}

	// correct → 204, and new password logs in
	ch2, _ := json.Marshal(map[string]string{"senha_atual": "antiga123", "senha_nova": "nova12345"})
	req2 := httptest.NewRequest("POST", "/api/student/password", bytes.NewReader(ch2))
	req2.AddCookie(cookie)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusNoContent {
		t.Fatalf("change status = %d body=%s", w2.Code, w2.Body)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/http/ -run TestStudentChangePassword -v`
Expected: FAIL (route 404).

- [ ] **Step 3: Add the handler**

Append to `backend/internal/http/student_handlers.go`:

```go
func (d Deps) studentChangePassword(w http.ResponseWriter, r *http.Request) {
	sid, ok := studentIDFromContext(r)
	if !ok {
		writeJSON(w, 401, map[string]string{"error": "não autenticado"})
		return
	}
	var in struct {
		SenhaAtual string `json:"senha_atual"`
		SenhaNova  string `json:"senha_nova"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, 400, map[string]string{"error": "json inválido"})
		return
	}
	if len(in.SenhaNova) < 6 {
		writeJSON(w, 400, map[string]string{"error": "senha nova muito curta"})
		return
	}
	s, err := d.Store.GetStudent(r.Context(), sid)
	if err != nil || !auth.CheckPassword(s.SenhaHash, in.SenhaAtual) {
		writeJSON(w, 401, map[string]string{"error": "senha atual incorreta"})
		return
	}
	hash, err := auth.HashPassword(in.SenhaNova)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": "erro"})
		return
	}
	if err := d.Store.UpdateStudentPassword(r.Context(), store.UpdateStudentPasswordParams{ID: sid, SenhaHash: hash}); err != nil {
		writeJSON(w, 500, map[string]string{"error": "erro"})
		return
	}
	w.WriteHeader(204)
}
```

- [ ] **Step 4: Register the guarded route**

In `backend/internal/http/router.go`, add a student-guarded group after the public student routes:

```go
		r.Group(func(r chi.Router) {
			r.Use(d.RequireStudent)
			r.Post("/student/password", d.studentChangePassword)
		})
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd backend && go test ./internal/http/ -run TestStudentChangePassword -v`
Expected: PASS (or SKIP).

- [ ] **Step 6: Commit**

```bash
git add backend/internal/http/student_handlers.go backend/internal/http/router.go backend/internal/http/student_handlers_test.go
git commit -m "feat(turmas): student password change"
```

---

### Task 5: Turma CRUD handlers

**Files:**
- Create: `backend/internal/http/turma_handlers.go`
- Modify: `backend/internal/http/router.go` (register under professor group)
- Test: `backend/internal/http/turma_handlers_test.go`

**Interfaces:**
- Consumes: `RequireAuth`, `userIDFromContext`, `store.CreateTurma`, `store.ListTurmasByUser`, `store.GetTurma`, `store.UpdateTurma`, `store.DeleteTurma`, `store.ListStudentsByTurma`.
- Produces: handlers `createTurma`, `listTurmas`, `getTurma`, `patchTurma`, `deleteTurma`; helper `func (d Deps) loadOwnedTurma(w, r, userID) (store.Turma, bool)` that parses `{id}`, loads it, and returns `404` if missing or not owned.

- [ ] **Step 1: Write the failing test**

Create `backend/internal/http/turma_handlers_test.go`:

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

// loginProfessor creates a professor and returns an authenticated session cookie.
func loginProfessor(t *testing.T, d Deps, email string) *http.Cookie {
	t.Helper()
	hash, _ := auth.HashPassword("senha123")
	u, err := d.Store.CreateUser(context.Background(), store.CreateUserParams{Email: email, SenhaHash: hash, Nome: "Prof"})
	if err != nil {
		t.Fatal(err)
	}
	sid, _ := auth.NewSessionID()
	_, err = d.Store.CreateSession(context.Background(), store.CreateSessionParams{ID: sid, UserID: u.ID, ExpiresAt: pgTime(timePlus24h())})
	if err != nil {
		t.Fatal(err)
	}
	return &http.Cookie{Name: "sid", Value: sid}
}

func TestTurmaCRUDAndOwnership(t *testing.T) {
	d := testDeps(t)
	r := NewRouter(d)
	cookieA := loginProfessor(t, d, "profA-turma@t.com")
	cookieB := loginProfessor(t, d, "profB-turma@t.com")

	// create
	body, _ := json.Marshal(map[string]any{"nome": "6A", "etapa": "EF", "anos": []int{6}})
	req := httptest.NewRequest("POST", "/api/turmas", bytes.NewReader(body))
	req.AddCookie(cookieA)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("create status = %d body=%s", w.Code, w.Body)
	}
	var created struct {
		ID int64 `json:"id"`
	}
	json.Unmarshal(w.Body.Bytes(), &created)

	// professor B cannot GET professor A's turma → 404
	req2 := httptest.NewRequest("GET", "/api/turmas/"+itoa(created.ID), nil)
	req2.AddCookie(cookieB)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusNotFound {
		t.Fatalf("cross-owner GET = %d, want 404", w2.Code)
	}
}
```

Add helpers at the bottom of the same file (only if not already present elsewhere):

```go
import "strconv"

func itoa(n int64) string { return strconv.FormatInt(n, 10) }
```

(If `timePlus24h` does not already exist in the test package, add it too:)

```go
import "time"

func timePlus24h() time.Time { return time.Now().Add(24 * time.Hour) }
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/http/ -run TestTurmaCRUDAndOwnership -v`
Expected: FAIL (route 404). If a duplicate declaration error appears for `itoa`/`timePlus24h`, remove the local copy and reuse the existing one.

- [ ] **Step 3: Write the handlers**

Create `backend/internal/http/turma_handlers.go`:

```go
package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"github.com/Carlos-hub/planejai/backend/internal/store"
)

type turmaInput struct {
	Nome  string  `json:"nome"`
	Etapa string  `json:"etapa"`
	Anos  []int32 `json:"anos"`
}

func (d Deps) loadOwnedTurma(w http.ResponseWriter, r *http.Request, userID int64) (store.Turma, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
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

func (d Deps) createTurma(w http.ResponseWriter, r *http.Request) {
	userID, _ := userIDFromContext(r)
	var in turmaInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Nome == "" {
		writeJSON(w, 400, map[string]string{"error": "nome é obrigatório"})
		return
	}
	if in.Anos == nil {
		in.Anos = []int32{}
	}
	turma, err := d.Store.CreateTurma(r.Context(), store.CreateTurmaParams{
		UserID: userID, Nome: in.Nome, Etapa: in.Etapa, Anos: in.Anos,
	})
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": "erro ao criar turma"})
		return
	}
	writeJSON(w, 201, turma)
}

func (d Deps) listTurmas(w http.ResponseWriter, r *http.Request) {
	userID, _ := userIDFromContext(r)
	turmas, err := d.Store.ListTurmasByUser(r.Context(), userID)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": "erro ao listar turmas"})
		return
	}
	writeJSON(w, 200, turmas)
}

// studentPublic is a student without any secret fields.
type studentPublic struct {
	ID        int64   `json:"id"`
	Nome      string  `json:"nome"`
	Usuario   string  `json:"usuario"`
	Matricula *string `json:"matricula"`
}

func (d Deps) getTurma(w http.ResponseWriter, r *http.Request) {
	userID, _ := userIDFromContext(r)
	turma, ok := d.loadOwnedTurma(w, r, userID)
	if !ok {
		return
	}
	students, err := d.Store.ListStudentsByTurma(r.Context(), turma.ID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		writeJSON(w, 500, map[string]string{"error": "erro ao carregar alunos"})
		return
	}
	pub := make([]studentPublic, 0, len(students))
	for _, s := range students {
		pub = append(pub, studentPublic{ID: s.ID, Nome: s.Nome, Usuario: s.Usuario, Matricula: s.Matricula})
	}
	writeJSON(w, 200, map[string]any{"turma": turma, "alunos": pub})
}

func (d Deps) patchTurma(w http.ResponseWriter, r *http.Request) {
	userID, _ := userIDFromContext(r)
	turma, ok := d.loadOwnedTurma(w, r, userID)
	if !ok {
		return
	}
	var in turmaInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, 400, map[string]string{"error": "json inválido"})
		return
	}
	if in.Nome == "" {
		in.Nome = turma.Nome
	}
	if in.Anos == nil {
		in.Anos = turma.Anos
	}
	updated, err := d.Store.UpdateTurma(r.Context(), store.UpdateTurmaParams{
		ID: turma.ID, Nome: in.Nome, Etapa: in.Etapa, Anos: in.Anos,
	})
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": "erro ao atualizar turma"})
		return
	}
	writeJSON(w, 200, updated)
}

func (d Deps) deleteTurma(w http.ResponseWriter, r *http.Request) {
	userID, _ := userIDFromContext(r)
	turma, ok := d.loadOwnedTurma(w, r, userID)
	if !ok {
		return
	}
	if err := d.Store.DeleteTurma(r.Context(), turma.ID); err != nil {
		writeJSON(w, 500, map[string]string{"error": "erro ao remover turma"})
		return
	}
	w.WriteHeader(204)
}
```

- [ ] **Step 4: Register routes**

In `backend/internal/http/router.go`, inside the professor `r.Group(func(r chi.Router){ r.Use(d.RequireAuth) ... })`, add:

```go
			r.Post("/turmas", d.createTurma)
			r.Get("/turmas", d.listTurmas)
			r.Get("/turmas/{id}", d.getTurma)
			r.Patch("/turmas/{id}", d.patchTurma)
			r.Delete("/turmas/{id}", d.deleteTurma)
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd backend && go test ./internal/http/ -run TestTurmaCRUDAndOwnership -v`
Expected: PASS (or SKIP). `go build ./...` passes.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/http/turma_handlers.go backend/internal/http/router.go backend/internal/http/turma_handlers_test.go
git commit -m "feat(turmas): turma CRUD with ownership checks"
```

---

### Task 6: CSV student import

**Files:**
- Modify: `backend/internal/http/turma_handlers.go` (add `importStudents`)
- Modify: `backend/internal/http/router.go` (route)
- Test: `backend/internal/http/turma_handlers_test.go` (add case)

**Interfaces:**
- Consumes: `loadOwnedTurma`, `auth.UsernameSlug`, `auth.RandomSuffix`, `auth.GenerateInitialPassword`, `auth.HashPassword`, `store.CreateStudent`.
- Produces: handler `importStudents` at `POST /api/turmas/{id}/students/import`. Request body is CSV text (header row with `nome` required, `matricula` optional). Response `201`: `{"criados": [{"nome","matricula","usuario","senha"}]}` with plaintext `senha` (one-time). Invalid CSV (missing `nome` header, or any row with empty nome) → `400` and nothing is inserted (validate all rows before inserting any).

- [ ] **Step 1: Write the failing test**

Add to `backend/internal/http/turma_handlers_test.go`:

```go
func TestImportStudents(t *testing.T) {
	d := testDeps(t)
	r := NewRouter(d)
	cookie := loginProfessor(t, d, "profImport-turma@t.com")

	// create a turma
	body, _ := json.Marshal(map[string]any{"nome": "6B"})
	cr := httptest.NewRequest("POST", "/api/turmas", bytes.NewReader(body))
	cr.AddCookie(cookie)
	cw := httptest.NewRecorder()
	r.ServeHTTP(cw, cr)
	var created struct {
		ID int64 `json:"id"`
	}
	json.Unmarshal(cw.Body.Bytes(), &created)

	csv := "nome,matricula\nAna Clara,1001\nJoão da Silva,1002\n"
	req := httptest.NewRequest("POST", "/api/turmas/"+itoa(created.ID)+"/students/import", bytes.NewReader([]byte(csv)))
	req.AddCookie(cookie)
	req.Header.Set("Content-Type", "text/csv")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("import status = %d body=%s", w.Code, w.Body)
	}
	var resp struct {
		Criados []struct {
			Nome, Usuario, Senha string
		} `json:"criados"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.Criados) != 2 {
		t.Fatalf("criados = %d, want 2", len(resp.Criados))
	}
	if resp.Criados[0].Senha == "" || resp.Criados[0].Usuario == "" {
		t.Fatalf("missing generated credentials: %+v", resp.Criados[0])
	}
	if resp.Criados[0].Usuario == resp.Criados[1].Usuario {
		t.Fatalf("usuarios must be unique")
	}

	// students actually persisted
	gr := httptest.NewRequest("GET", "/api/turmas/"+itoa(created.ID), nil)
	gr.AddCookie(cookie)
	gw := httptest.NewRecorder()
	r.ServeHTTP(gw, gr)
	if !bytes.Contains(gw.Body.Bytes(), []byte("Ana Clara")) {
		t.Fatalf("student not persisted: %s", gw.Body)
	}
	// plaintext password must never be in a GET
	if bytes.Contains(gw.Body.Bytes(), []byte(resp.Criados[0].Senha)) {
		t.Fatalf("plaintext password leaked in GET")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/http/ -run TestImportStudents -v`
Expected: FAIL (route 404).

- [ ] **Step 3: Write the handler**

Append to `backend/internal/http/turma_handlers.go` (add `"encoding/csv"`, `"io"`, `"strings"` to the import block):

```go
type importedStudent struct {
	Nome      string  `json:"nome"`
	Matricula *string `json:"matricula"`
	Usuario   string  `json:"usuario"`
	Senha     string  `json:"senha"`
}

// createUniqueStudent generates a unique username and inserts one student,
// retrying the random suffix on a unique-violation (Postgres 23505).
func (d Deps) createUniqueStudent(r *http.Request, turmaID int64, nome string, matricula *string) (importedStudent, error) {
	slug := auth.UsernameSlug(nome)
	if slug == "" {
		slug = "aluno"
	}
	senha, err := auth.GenerateInitialPassword()
	if err != nil {
		return importedStudent{}, err
	}
	hash, err := auth.HashPassword(senha)
	if err != nil {
		return importedStudent{}, err
	}
	for attempt := 0; attempt < 8; attempt++ {
		suffix, err := auth.RandomSuffix()
		if err != nil {
			return importedStudent{}, err
		}
		usuario := slug + "." + suffix
		s, err := d.Store.CreateStudent(r.Context(), store.CreateStudentParams{
			TurmaID: turmaID, Nome: nome, Usuario: usuario, SenhaHash: hash, Matricula: matricula,
		})
		if err == nil {
			return importedStudent{Nome: s.Nome, Matricula: s.Matricula, Usuario: s.Usuario, Senha: senha}, nil
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			continue
		}
		return importedStudent{}, err
	}
	return importedStudent{}, errors.New("não foi possível gerar usuário único")
}

func (d Deps) importStudents(w http.ResponseWriter, r *http.Request) {
	userID, _ := userIDFromContext(r)
	turma, ok := d.loadOwnedTurma(w, r, userID)
	if !ok {
		return
	}

	reader := csv.NewReader(r.Body)
	reader.TrimLeadingSpace = true
	header, err := reader.Read()
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": "csv vazio ou inválido"})
		return
	}
	nomeIdx, matIdx := -1, -1
	for i, h := range header {
		switch strings.ToLower(strings.TrimSpace(h)) {
		case "nome":
			nomeIdx = i
		case "matricula", "matrícula":
			matIdx = i
		}
	}
	if nomeIdx == -1 {
		writeJSON(w, 400, map[string]string{"error": "coluna 'nome' obrigatória"})
		return
	}

	type parsedRow struct {
		nome      string
		matricula *string
	}
	var rows []parsedRow
	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			writeJSON(w, 400, map[string]string{"error": "csv inválido"})
			return
		}
		if nomeIdx >= len(rec) || strings.TrimSpace(rec[nomeIdx]) == "" {
			writeJSON(w, 400, map[string]string{"error": "todas as linhas precisam de nome"})
			return
		}
		var mat *string
		if matIdx != -1 && matIdx < len(rec) && strings.TrimSpace(rec[matIdx]) != "" {
			v := strings.TrimSpace(rec[matIdx])
			mat = &v
		}
		rows = append(rows, parsedRow{nome: strings.TrimSpace(rec[nomeIdx]), matricula: mat})
	}

	criados := make([]importedStudent, 0, len(rows))
	for _, row := range rows {
		imp, err := d.createUniqueStudent(r, turma.ID, row.nome, row.matricula)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "erro ao criar aluno: " + row.nome})
			return
		}
		criados = append(criados, imp)
	}
	writeJSON(w, 201, map[string]any{"criados": criados})
}
```

Add `"github.com/jackc/pgx/v5/pgconn"` and `"github.com/Carlos-hub/planejai/backend/internal/auth"` to the import block if not already present.

- [ ] **Step 4: Register the route**

In `backend/internal/http/router.go`, in the professor group, add:

```go
			r.Post("/turmas/{id}/students/import", d.importStudents)
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd backend && go test ./internal/http/ -run TestImportStudents -v`
Expected: PASS (or SKIP). `go build ./...` passes.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/http/turma_handlers.go backend/internal/http/router.go backend/internal/http/turma_handlers_test.go
git commit -m "feat(turmas): CSV student import with generated credentials"
```

---

### Task 7: Publish activity to a turma

**Files:**
- Modify: `backend/internal/http/lesson_handlers.go` (`publishTrail` reads optional `turma_id`)
- Test: `backend/internal/http/publish_test.go` (add case) — or `turma_handlers_test.go` if simpler

**Interfaces:**
- Consumes: `store.GetTurma`, `store.SetTrailTurma`.
- Produces: `publishTrail` now accepts an optional JSON body `{"turma_id": <int64>}`. When present and owned by the caller, it stores `study_trails.turma_id`; when absent, behavior is unchanged. Unknown/other-owner turma → `404`.

- [ ] **Step 1: Read the current publishTrail signature**

Run: `cd backend && sed -n '376,432p' internal/http/lesson_handlers.go`
Expected: confirms the function currently takes no request body.

- [ ] **Step 2: Write the failing test**

Add to `backend/internal/http/turma_handlers_test.go` a test that publishes with a `turma_id` and asserts a subsequent anonymous `GET /api/t/{code}` (from Task 8) is gated. Since gating lands in Task 8, this task's test asserts only that publishing with a valid `turma_id` returns `200` and an unknown `turma_id` returns `404`:

```go
func TestPublishWithTurma(t *testing.T) {
	d := testDeps(t)
	r := NewRouter(d)
	cookie := loginProfessor(t, d, "profPub-turma@t.com")

	// create turma
	tb, _ := json.Marshal(map[string]any{"nome": "9C"})
	tr := httptest.NewRequest("POST", "/api/turmas", bytes.NewReader(tb))
	tr.AddCookie(cookie)
	tw := httptest.NewRecorder()
	r.ServeHTTP(tw, tr)
	var turma struct {
		ID int64 `json:"id"`
	}
	json.Unmarshal(tw.Body.Bytes(), &turma)

	// publishing to a turma owned by another professor → 404
	other := loginProfessor(t, d, "profOther-turma@t.com")
	trail := seedPublishableTrail(t, d, other) // helper: creates lesson+trail ready to publish, returns lesson id
	pb, _ := json.Marshal(map[string]any{"turma_id": turma.ID})
	pr := httptest.NewRequest("POST", "/api/trails/"+itoa(trail)+"/publish", bytes.NewReader(pb))
	pr.AddCookie(other)
	pw := httptest.NewRecorder()
	r.ServeHTTP(pw, pr)
	if pw.Code != http.StatusNotFound {
		t.Fatalf("publish to foreign turma = %d, want 404", pw.Code)
	}
}
```

If a helper that builds a publishable trail already exists in `publish_test.go` (look for how `TestPublish...` sets up a lesson with status `pronto` and a trail), reuse it and delete `seedPublishableTrail`. Otherwise add a minimal helper mirroring that existing setup.

- [ ] **Step 3: Run test to verify it fails**

Run: `cd backend && go test ./internal/http/ -run TestPublishWithTurma -v`
Expected: FAIL (turma_id ignored → publish succeeds with 200 instead of 404).

- [ ] **Step 4: Modify publishTrail**

In `backend/internal/http/lesson_handlers.go`, at the start of `publishTrail` (after `lp, ok := d.loadOwnedLesson(...)` succeeds and before generating the code), parse the optional body and validate turma ownership:

```go
	var body struct {
		TurmaID *int64 `json:"turma_id"`
	}
	// body is optional; ignore decode errors on empty body
	_ = json.NewDecoder(r.Body).Decode(&body)
	if body.TurmaID != nil {
		turma, err := d.Store.GetTurma(ctx, *body.TurmaID)
		if err != nil || turma.UserID != userID {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "turma não encontrada"})
			return
		}
	}
```

Then, after a code is successfully assigned (`codigo != ""`), persist the turma link:

```go
	if body.TurmaID != nil {
		if err := d.Store.SetTrailTurma(ctx, store.SetTrailTurmaParams{ID: trail.ID, TurmaID: body.TurmaID}); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao vincular turma"})
			return
		}
	}
```

Ensure `ctx` is defined before first use (it is, via `ctx := r.Context()` already in the function — move the body parse below that line).

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd backend && go test ./internal/http/ -run 'TestPublishWithTurma|TestPublish' -v`
Expected: PASS (or SKIP). Existing publish tests (no body) still pass because the body is optional.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/http/lesson_handlers.go backend/internal/http/turma_handlers_test.go
git commit -m "feat(turmas): publish activity bound to a turma"
```

---

### Task 8: Gate the public trail flow by turma

**Files:**
- Modify: `backend/internal/http/public_handlers.go` (`publicTrail` + `startAttempt` gating)
- Test: `backend/internal/http/public_test.go` (add cases) or `turma_handlers_test.go`

**Interfaces:**
- Consumes: `store.GetStudentSession`, `store.GetStudent`, `studentCookie`.
- Produces: helper `func (d Deps) resolveStudentForTrail(r, trail) (*store.Student, int)` returning the logged-in student and an HTTP status: `0` = ok/anonymous-allowed, `401` = login required, `403` = wrong turma. When `trail.TurmaID == nil`, always returns `(nil, 0)`. `startAttempt` sets `StudentID` from the resolved student.

- [ ] **Step 1: Write the failing test**

Add to `backend/internal/http/turma_handlers_test.go`:

```go
func TestGatedTrailAccess(t *testing.T) {
	d := testDeps(t)
	r := NewRouter(d)
	cookie := loginProfessor(t, d, "profGate-turma@t.com")

	// turma + student
	tb, _ := json.Marshal(map[string]any{"nome": "7A"})
	tr := httptest.NewRequest("POST", "/api/turmas", bytes.NewReader(tb))
	tr.AddCookie(cookie)
	tw := httptest.NewRecorder()
	r.ServeHTTP(tw, tr)
	var turma struct{ ID int64 `json:"id"` }
	json.Unmarshal(tw.Body.Bytes(), &turma)

	// publish a trail bound to this turma; capture its code
	code := publishTrailToTurma(t, d, r, cookie, turma.ID) // helper returns the short code

	// anonymous access to a gated trail → 401
	ar := httptest.NewRequest("GET", "/api/t/"+code, nil)
	aw := httptest.NewRecorder()
	r.ServeHTTP(aw, ar)
	if aw.Code != http.StatusUnauthorized {
		t.Fatalf("anon gated GET = %d, want 401", aw.Code)
	}

	// enrolled student can access
	hash, _ := auth.HashPassword("aluno123")
	s, _ := d.Store.CreateStudent(context.Background(), store.CreateStudentParams{TurmaID: turma.ID, Nome: "Bia", Usuario: "bia.gate.cc", SenhaHash: hash})
	ssid, _ := auth.NewSessionID()
	d.Store.CreateStudentSession(context.Background(), store.CreateStudentSessionParams{ID: ssid, StudentID: s.ID, ExpiresAt: pgTime(timePlus24h())})
	sr := httptest.NewRequest("GET", "/api/t/"+code, nil)
	sr.AddCookie(&http.Cookie{Name: "student_sid", Value: ssid})
	sw := httptest.NewRecorder()
	r.ServeHTTP(sw, sr)
	if sw.Code != http.StatusOK {
		t.Fatalf("enrolled gated GET = %d, want 200", sw.Code)
	}
}
```

Add helper `publishTrailToTurma` (reuse the publishable-trail setup from Task 7's helper; publish with `{"turma_id":...}` and return `resp.Codigo`).

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/http/ -run TestGatedTrailAccess -v`
Expected: FAIL (anonymous gets 200 instead of 401).

- [ ] **Step 3: Add the gating helper**

Append to `backend/internal/http/public_handlers.go`:

```go
// resolveStudentForTrail enforces turma gating. For an ungated trail
// (TurmaID == nil) it returns (nil, 0). For a gated trail it requires a
// valid student session whose student belongs to the trail's turma:
// status 401 = not logged in / invalid session, 403 = wrong turma.
func (d Deps) resolveStudentForTrail(r *http.Request, trail store.StudyTrail) (*store.Student, int) {
	if trail.TurmaID == nil {
		return nil, 0
	}
	c, err := r.Cookie(studentCookie)
	if err != nil {
		return nil, http.StatusUnauthorized
	}
	sess, err := d.Store.GetStudentSession(r.Context(), c.Value)
	if err != nil {
		return nil, http.StatusUnauthorized
	}
	s, err := d.Store.GetStudent(r.Context(), sess.StudentID)
	if err != nil {
		return nil, http.StatusUnauthorized
	}
	if s.TurmaID != *trail.TurmaID {
		return nil, http.StatusForbidden
	}
	return &s, 0
}
```

- [ ] **Step 4: Enforce gating in publicTrail**

In `publicTrail`, immediately after the `if !trail.PublicadaEm.Valid { ... }` block, add:

```go
	if _, status := d.resolveStudentForTrail(r, trail); status != 0 {
		writeJSON(w, status, map[string]string{"error": "acesso restrito à turma"})
		return
	}
```

- [ ] **Step 5: Enforce gating + attach student in startAttempt**

In `startAttempt`, after loading `trail` and its `PublicadaEm` check, add:

```go
	student, status := d.resolveStudentForTrail(r, trail)
	if status != 0 {
		writeJSON(w, status, map[string]string{"error": "acesso restrito à turma"})
		return
	}
	var studentID *int64
	nome := req.Nome
	if student != nil {
		studentID = &student.ID
		nome = student.Nome
	}
```

Then change the `CreateAttempt` call to use them:

```go
	attempt, err := d.Store.CreateAttempt(ctx, store.CreateAttemptParams{
		StudyTrailID: trail.ID,
		NomeAluno:    nome,
		StudentID:    studentID,
	})
```

Note: for a gated trail the student is identified by session, so `req.Nome` may be empty; the earlier `req.Nome == ""` rejection must be relaxed to only apply when `student == nil`. Move the nome-required check to after `resolveStudentForTrail`:

```go
	if student == nil && strings.TrimSpace(req.Nome) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "nome é obrigatório"})
		return
	}
```

(Remove the original `if strings.TrimSpace(req.Nome) == ""` block near the top of the handler.)

- [ ] **Step 6: Run tests to verify they pass**

Run: `cd backend && go test ./internal/http/ -run 'TestGatedTrailAccess|TestPublic|TestAttempt' -v`
Expected: PASS (or SKIP). Ungated trails still allow anonymous attempts.

- [ ] **Step 7: Commit**

```bash
git add backend/internal/http/public_handlers.go backend/internal/http/turma_handlers_test.go
git commit -m "feat(turmas): gate trail access + attempts by turma enrollment"
```

---

### Task 9: Frontend — student login + gated trail

**Files:**
- Modify: `frontend/lib/api.ts` (add student auth helpers) or `frontend/lib/hooks.ts`
- Modify: `frontend/lib/types.ts` (student types)
- Create: `frontend/app/aluno/login/page.tsx`
- Modify: `frontend/app/t/[code]/page.tsx` (on 401, redirect to `/aluno/login?next=/t/<code>`)

**Interfaces:**
- Consumes: `apiFetch` (from `lib/api.ts`), backend `POST /api/student/login`, gated `GET /api/t/{code}`.
- Produces: `studentLogin(usuario, senha)`, `Student` type `{ id:number; nome:string; usuario:string; turma_id:number }`.

- [ ] **Step 1: Add types + api helpers**

In `frontend/lib/types.ts`, add:

```ts
export interface Student {
  id: number;
  nome: string;
  usuario: string;
  turma_id: number;
}
```

In `frontend/lib/api.ts`, add:

```ts
import type { Student } from "./types";

export function studentLogin(usuario: string, senha: string): Promise<Student> {
  return apiFetch<Student>("/api/student/login", {
    method: "POST",
    body: JSON.stringify({ usuario, senha }),
  });
}
```

- [ ] **Step 2: Create the student login page**

Create `frontend/app/aluno/login/page.tsx`:

```tsx
"use client";
import { useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { studentLogin } from "@/lib/api";

export default function AlunoLoginPage() {
  const router = useRouter();
  const params = useSearchParams();
  const next = params.get("next") || "/";
  const [usuario, setUsuario] = useState("");
  const [senha, setSenha] = useState("");
  const [erro, setErro] = useState("");

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setErro("");
    try {
      await studentLogin(usuario, senha);
      router.push(next);
    } catch {
      setErro("Usuário ou senha inválidos.");
    }
  }

  return (
    <main style={{ maxWidth: 360, margin: "4rem auto", padding: "0 1rem" }}>
      <h1>Entrar como aluno</h1>
      <form onSubmit={onSubmit}>
        <label>Usuário<input value={usuario} onChange={(e) => setUsuario(e.target.value)} autoFocus /></label>
        <label>Senha<input type="password" value={senha} onChange={(e) => setSenha(e.target.value)} /></label>
        {erro && <p style={{ color: "crimson" }}>{erro}</p>}
        <button type="submit">Entrar</button>
      </form>
    </main>
  );
}
```

(Match the styling/components used by the existing teacher login page `frontend/app/(teacher)/login/page.tsx` — read it first and mirror its form components rather than the inline styles above if it uses a UI kit.)

- [ ] **Step 3: Handle gated 401 on the trail page**

In `frontend/app/t/[code]/page.tsx`, where the trail is fetched, catch a `401` from `apiFetch` (error message starts with `API 401`) and redirect:

```tsx
// inside the fetch error handler for the trail load:
if (err instanceof Error && err.message.startsWith("API 401")) {
  router.push(`/aluno/login?next=/t/${code}`);
  return;
}
```

(Read the current data-loading code in that file and place this in its existing catch/error path; import `useRouter` from `next/navigation` if not already present.)

- [ ] **Step 4: Verify build**

Run: `cd frontend && npm run build`
Expected: build succeeds, no type errors. `/aluno/login` route is emitted.

- [ ] **Step 5: Manual smoke (with backend running)**

Bring up the stack (`docker compose up` or the project's usual dev command). Create a professor, a turma, import a student CSV, publish an activity to that turma. In a private window open `/t/<code>` → expect redirect to `/aluno/login`. Log in with the generated `usuario`/senha → expect the trail to render.

- [ ] **Step 6: Commit**

```bash
git add frontend/app/aluno frontend/lib/api.ts frontend/lib/types.ts frontend/app/t/[code]/page.tsx
git commit -m "feat(web): student login page + gated trail redirect"
```

---

### Task 10: Frontend — turma management pages

**Files:**
- Modify: `frontend/lib/api.ts` (turma helpers)
- Modify: `frontend/lib/types.ts` (Turma types)
- Create: `frontend/app/(teacher)/turmas/page.tsx` (list + create)
- Create: `frontend/app/(teacher)/turmas/[id]/page.tsx` (detail + CSV import + credentials export)
- Modify: teacher nav shell to add a "Turmas" link (find the nav component used by `app/(teacher)/layout.tsx`)

**Interfaces:**
- Consumes: `apiFetch`, backend `/api/turmas` CRUD + `/api/turmas/{id}/students/import`.
- Produces: `Turma` type `{ id:number; nome:string; etapa:string; anos:number[] }`; helpers `listTurmas()`, `createTurma(input)`, `getTurma(id)`, `importStudentsCSV(id, csvText)`.

- [ ] **Step 1: Add types + api helpers**

In `frontend/lib/types.ts`:

```ts
export interface Turma {
  id: number;
  nome: string;
  etapa: string;
  anos: number[];
}

export interface ImportedStudent {
  nome: string;
  matricula: string | null;
  usuario: string;
  senha: string;
}
```

In `frontend/lib/api.ts`:

```ts
import type { Turma, ImportedStudent } from "./types";

export const listTurmas = () => apiFetch<Turma[]>("/api/turmas");
export const getTurma = (id: number) =>
  apiFetch<{ turma: Turma; alunos: { id: number; nome: string; usuario: string; matricula: string | null }[] }>(`/api/turmas/${id}`);
export const createTurma = (input: { nome: string; etapa?: string; anos?: number[] }) =>
  apiFetch<Turma>("/api/turmas", { method: "POST", body: JSON.stringify(input) });

export async function importStudentsCSV(id: number, csvText: string): Promise<{ criados: ImportedStudent[] }> {
  return apiFetch<{ criados: ImportedStudent[] }>(`/api/turmas/${id}/students/import`, {
    method: "POST",
    headers: { "Content-Type": "text/csv" },
    body: csvText,
  });
}
```

Note: `apiFetch` sets `Content-Type: application/json` by default; the `headers` override above replaces it with `text/csv`. Verify the merge in `apiFetch` lets the override win (it spreads `init.headers` last — it does).

- [ ] **Step 2: Build the turmas list page**

Create `frontend/app/(teacher)/turmas/page.tsx` with a list of turmas (from `listTurmas`) and a create form (`createTurma`). Each row links to `/turmas/[id]`. Mirror the styling/components of the existing lessons list page (`frontend/app/(teacher)/lessons/page.tsx` if present — read it first and reuse its layout/components).

```tsx
"use client";
import { useEffect, useState } from "react";
import Link from "next/link";
import { listTurmas, createTurma } from "@/lib/api";
import type { Turma } from "@/lib/types";

export default function TurmasPage() {
  const [turmas, setTurmas] = useState<Turma[]>([]);
  const [nome, setNome] = useState("");

  async function reload() { setTurmas(await listTurmas()); }
  useEffect(() => { reload(); }, []);

  async function onCreate(e: React.FormEvent) {
    e.preventDefault();
    if (!nome.trim()) return;
    await createTurma({ nome });
    setNome("");
    reload();
  }

  return (
    <main>
      <h1>Turmas</h1>
      <form onSubmit={onCreate}>
        <input value={nome} onChange={(e) => setNome(e.target.value)} placeholder="Nome da turma" />
        <button type="submit">Criar turma</button>
      </form>
      <ul>
        {turmas.map((t) => (
          <li key={t.id}><Link href={`/turmas/${t.id}`}>{t.nome}</Link></li>
        ))}
      </ul>
    </main>
  );
}
```

- [ ] **Step 3: Build the turma detail page (CSV import + credentials)**

Create `frontend/app/(teacher)/turmas/[id]/page.tsx`: shows students, a CSV upload (file → text → `importStudentsCSV`), and renders the returned credentials table with a "Baixar credenciais (CSV)" button (build a CSV blob client-side) and a print button. The plaintext passwords are only in the import response — make clear in the UI they won't be shown again.

```tsx
"use client";
import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { getTurma, importStudentsCSV } from "@/lib/api";
import type { ImportedStudent } from "@/lib/types";

export default function TurmaDetailPage() {
  const { id } = useParams<{ id: string }>();
  const turmaId = Number(id);
  const [nome, setNome] = useState("");
  const [alunos, setAlunos] = useState<{ id: number; nome: string; usuario: string }[]>([]);
  const [criados, setCriados] = useState<ImportedStudent[]>([]);

  async function reload() {
    const data = await getTurma(turmaId);
    setNome(data.turma.nome);
    setAlunos(data.alunos);
  }
  useEffect(() => { reload(); }, [turmaId]);

  async function onFile(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;
    const text = await file.text();
    const res = await importStudentsCSV(turmaId, text);
    setCriados(res.criados);
    reload();
  }

  function downloadCredentials() {
    const rows = [["nome", "usuario", "senha"], ...criados.map((c) => [c.nome, c.usuario, c.senha])];
    const csv = rows.map((r) => r.join(",")).join("\n");
    const url = URL.createObjectURL(new Blob([csv], { type: "text/csv" }));
    const a = document.createElement("a");
    a.href = url; a.download = `credenciais-${nome}.csv`; a.click();
    URL.revokeObjectURL(url);
  }

  return (
    <main>
      <h1>{nome}</h1>
      <section>
        <h2>Importar alunos (CSV)</h2>
        <p>Colunas: <code>nome</code> (obrigatória), <code>matricula</code> (opcional).</p>
        <input type="file" accept=".csv,text/csv" onChange={onFile} />
      </section>

      {criados.length > 0 && (
        <section>
          <h2>Credenciais geradas</h2>
          <p><strong>Guarde agora</strong> — as senhas não serão exibidas novamente.</p>
          <button onClick={downloadCredentials}>Baixar credenciais (CSV)</button>
          <button onClick={() => window.print()}>Imprimir</button>
          <table>
            <thead><tr><th>Nome</th><th>Usuário</th><th>Senha</th></tr></thead>
            <tbody>
              {criados.map((c) => (
                <tr key={c.usuario}><td>{c.nome}</td><td>{c.usuario}</td><td>{c.senha}</td></tr>
              ))}
            </tbody>
          </table>
        </section>
      )}

      <section>
        <h2>Alunos ({alunos.length})</h2>
        <ul>{alunos.map((a) => <li key={a.id}>{a.nome} — <code>{a.usuario}</code></li>)}</ul>
      </section>
    </main>
  );
}
```

- [ ] **Step 4: Add nav link + (optional) turma selector on publish**

Add a "Turmas" link to the teacher nav (find the component rendered by `frontend/app/(teacher)/layout.tsx`). If the publish UI is easy to locate, add an optional turma `<select>` that passes `turma_id` to the publish call; otherwise leave publishing turma-binding for a follow-up and note it. (This keeps the task shippable even if the publish UI is intricate.)

- [ ] **Step 5: Verify build**

Run: `cd frontend && npm run build`
Expected: build succeeds, `/turmas` and `/turmas/[id]` routes emitted, no type errors.

- [ ] **Step 6: Manual smoke**

With the stack running and logged in as a professor: open `/turmas`, create a turma, open it, upload a CSV (`nome,matricula\nAna,1\nBruno,2`), confirm the credentials table appears, download the CSV, confirm the two students appear in the list.

- [ ] **Step 7: Commit**

```bash
git add frontend/app/(teacher)/turmas frontend/lib/api.ts frontend/lib/types.ts frontend/app/(teacher)/layout.tsx
git commit -m "feat(web): turma management + CSV import UI with credentials export"
```

---

### Task 11: Update README

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Update the flow + feature docs**

Edit `README.md` to reflect: turmas managed by the professor, student accounts created via CSV import (system-generated usuario+senha), and turma-gated activities requiring student login (while standalone trails keep the anonymous short-code flow). Update the "Fluxo do MVP" block to include the student-login path for turma activities.

- [ ] **Step 2: Verify + commit**

Run: `git diff README.md` and confirm the described flow matches what was built.

```bash
git add README.md
git commit -m "docs: document turmas + student accounts in README"
```

---

## Self-Review Notes

- **Spec coverage:** turmas table/CRUD (Task 1, 5), student accounts + CSV import with one-time credentials (Task 1, 2, 6), student auth + password change (Task 3, 4), turma-gated access with anonymous fallback preserved (Task 7, 8), attempt tied to `student_id` (Task 8), frontend student login + turma pages (Task 9, 10), README (Task 11). Encryption-at-rest for AI tokens is explicitly out of scope (Subsystem B) — not in this plan.
- **Ownership 404 rule:** applied in `loadOwnedTurma` (Task 5), publish (Task 7).
- **Type consistency:** `store.CreateAttemptParams.StudentID *int64` set to `nil` in Task 1 and to `&student.ID` in Task 8; `store.SetTrailTurmaParams{ID, TurmaID *int64}` produced in Task 1, consumed in Task 7; `studentCookie = "student_sid"` defined in Task 3, reused in Task 8; `importedStudent` JSON shape produced in Task 6, consumed by frontend `ImportedStudent` in Task 10.
- **Migration reversibility:** `00004` down-migration drops columns before tables in dependency order.
