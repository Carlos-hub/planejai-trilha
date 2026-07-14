# PlanejAI + Trilha (Go + NextJS) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rebuild the PlanejAI+Trilha MVP as a Go REST API + NextJS frontend: teachers author BNCC-aligned lessons (manual / full-AI / AI-enhance), publish student trails accessed by short code, students take a self-graded quiz with gamification, teachers see class stats and export PDF/WhatsApp.

**Architecture:** Monorepo with `backend/` (Go: chi router, sqlc+pgx data layer, goose migrations, server-side sessions, a brand-agnostic `lesson.Generator` LLM interface implemented with langchaingo) and `frontend/` (NextJS App Router, TypeScript, Tailwind, shadcn/ui). Everything runs via `docker-compose` (postgres + go-api + next). The API is the single source of truth; the frontend is a typed client. Quiz answer keys never leave the server; scoring is server-side.

**Tech Stack:** Go 1.23, chi v5, pgx v5, sqlc, goose, langchaingo, bcrypt, maroto (PDF); NextJS 15 (App Router), TypeScript, Tailwind, shadcn/ui; PostgreSQL 16; Docker Compose.

## Global Constraints

- Domain field names in Portuguese (`objetivos`, `metodologia`, `recursos`, `avaliacao`, `atividade`, `topicos`, `quiz`, `questoes`, `enunciado`, `opcoes`, `correta`) — verbatim, matches spec.
- LLM provider/model from env only: `LLM_PROVIDER` (default `anthropic`), `LLM_MODEL` (default `claude-opus-4-8`). Domain depends only on the `lesson.Generator` interface — no provider imports outside `internal/lesson`.
- Quiz answer key (`correta`) MUST NOT be serialized in any student-facing response. Scoring is server-side only.
- `lesson_plans.origem` ∈ `{manual, ia, ia_aprimorado}`; `lesson_plans.status` ∈ `{rascunho, pronto, falha}`.
- No student accounts. Student access = trail short code + typed name.
- Teacher auth = server-side sessions (Postgres table), httpOnly cookie `sid`, bcrypt password hashing.
- Trail codes format `TR-XXXX` (4 chars, uppercase A–Z/2–9, no ambiguous chars).
- All money-free; all timestamps UTC `timestamptz`.
- Go module path: `github.com/Carlos-hub/planejai/backend`.
- Every DB access goes through sqlc-generated code in `internal/store`; no ad-hoc SQL in handlers.

---

## Phase 0 — Scaffolding & Infra

### Task 0.1: Monorepo skeleton + docker-compose

**Files:**
- Create: `backend/go.mod`, `backend/.gitignore`, `frontend/.gitignore`
- Create: `docker-compose.yml`
- Create: `.env.example`
- Create: `README.md` (rewrite for Go+NextJS)

**Interfaces:**
- Produces: docker services `db` (postgres:16, port 5432), `api` (go, port 8080), `web` (next, port 3000); env `DATABASE_URL`, `LLM_PROVIDER`, `LLM_MODEL`, `ANTHROPIC_API_KEY`, `SESSION_SECRET`, `NEXT_PUBLIC_API_URL`.

- [ ] **Step 1: Init Go module**

```bash
mkdir -p backend/cmd/api backend/internal
cd backend && go mod init github.com/Carlos-hub/planejai/backend && cd ..
```

- [ ] **Step 2: Write `.env.example`**

```
DATABASE_URL=postgres://planejai:planejai@db:5432/planejai?sslmode=disable
LLM_PROVIDER=anthropic
LLM_MODEL=claude-opus-4-8
ANTHROPIC_API_KEY=sk-ant-xxx
SESSION_SECRET=change-me-32-bytes-min
PORT=8080
NEXT_PUBLIC_API_URL=http://localhost:8080
```

- [ ] **Step 3: Write `docker-compose.yml`**

```yaml
services:
  db:
    image: postgres:16
    environment:
      POSTGRES_USER: planejai
      POSTGRES_PASSWORD: planejai
      POSTGRES_DB: planejai
    ports: ["5432:5432"]
    volumes: ["dbdata:/var/lib/postgresql/data"]
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U planejai"]
      interval: 5s
      timeout: 5s
      retries: 10
  api:
    build: ./backend
    env_file: [.env]
    environment:
      DATABASE_URL: postgres://planejai:planejai@db:5432/planejai?sslmode=disable
    ports: ["8080:8080"]
    depends_on:
      db: { condition: service_healthy }
  web:
    build: ./frontend
    environment:
      NEXT_PUBLIC_API_URL: http://localhost:8080
    ports: ["3000:3000"]
    depends_on: [api]
volumes:
  dbdata:
```

- [ ] **Step 4: Write `backend/.gitignore` and `frontend/.gitignore`**

`backend/.gitignore`:
```
/tmp
*.env
.env
```
`frontend/.gitignore`:
```
node_modules
.next
.env*.local
```

- [ ] **Step 5: Rewrite root `README.md`** with project overview, stack, and `cp .env.example .env && docker compose up` quickstart.

- [ ] **Step 6: Commit**

```bash
git add backend/go.mod backend/.gitignore frontend/.gitignore docker-compose.yml .env.example README.md
git commit -m "chore: monorepo scaffolding and docker-compose"
```

---

### Task 0.2: Backend Dockerfile + minimal health server

**Files:**
- Create: `backend/cmd/api/main.go`
- Create: `backend/internal/http/router.go`
- Create: `backend/Dockerfile`
- Test: `backend/internal/http/router_test.go`

**Interfaces:**
- Produces: `http.NewRouter(deps Deps) http.Handler` where `Deps` is a struct (grown in later tasks); `GET /healthz` → `200 {"status":"ok"}`.

- [ ] **Step 1: Write failing test**

```go
package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthz(t *testing.T) {
	srv := httptest.NewServer(NewRouter(Deps{}))
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("want 200 got %d", resp.StatusCode)
	}
}
```

- [ ] **Step 2: Run test — expect FAIL (undefined NewRouter/Deps)**

Run: `cd backend && go test ./internal/http/ -run TestHealthz`
Expected: build error / FAIL.

- [ ] **Step 3: Implement router**

`backend/internal/http/router.go`:
```go
package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Deps holds handler dependencies; extended in later tasks.
type Deps struct{}

func NewRouter(d Deps) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	return r
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
```

- [ ] **Step 4: Implement main**

`backend/cmd/api/main.go`:
```go
package main

import (
	"log"
	"net/http"
	"os"

	apihttp "github.com/Carlos-hub/planejai/backend/internal/http"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("listening on :%s", port)
	if err := http.ListenAndServe(":"+port, apihttp.NewRouter(apihttp.Deps{})); err != nil {
		log.Fatal(err)
	}
}
```

- [ ] **Step 5: `go mod tidy` and run test**

Run: `cd backend && go get github.com/go-chi/chi/v5 && go mod tidy && go test ./internal/http/ -run TestHealthz`
Expected: PASS.

- [ ] **Step 6: Write `backend/Dockerfile`**

```dockerfile
FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /bin/api ./cmd/api

FROM alpine:3.20
COPY --from=build /bin/api /bin/api
COPY --from=build /src/seed /seed
COPY --from=build /src/migrations /migrations
EXPOSE 8080
ENTRYPOINT ["/bin/api"]
```

- [ ] **Step 7: Commit**

```bash
git add backend/cmd backend/internal/http backend/Dockerfile backend/go.mod backend/go.sum
git commit -m "feat: health server with chi router"
```

---

### Task 0.3: NextJS app + Dockerfile + API client stub

**Files:**
- Create: `frontend/` (via create-next-app), `frontend/lib/api.ts`, `frontend/Dockerfile`

**Interfaces:**
- Produces: `apiFetch<T>(path: string, init?: RequestInit): Promise<T>` in `lib/api.ts`, sending cookies (`credentials: "include"`) to `process.env.NEXT_PUBLIC_API_URL`.

- [ ] **Step 1: Scaffold Next app** (TypeScript, Tailwind, App Router, no src dir)

```bash
npx create-next-app@latest frontend --ts --tailwind --app --eslint --no-src-dir --import-alias "@/*" --use-npm
```

- [ ] **Step 2: Init shadcn/ui**

```bash
cd frontend && npx shadcn@latest init -d && npx shadcn@latest add button input card form label textarea progress table && cd ..
```

- [ ] **Step 3: Write typed API client**

`frontend/lib/api.ts`:
```ts
const BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export async function apiFetch<T>(path: string, init: RequestInit = {}): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    ...init,
    credentials: "include",
    headers: { "Content-Type": "application/json", ...(init.headers ?? {}) },
  });
  if (!res.ok) {
    const body = await res.text();
    throw new Error(`API ${res.status}: ${body}`);
  }
  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}
```

- [ ] **Step 4: Write `frontend/Dockerfile`**

```dockerfile
FROM node:20-alpine AS deps
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
FROM node:20-alpine AS build
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
RUN npm run build
FROM node:20-alpine AS run
WORKDIR /app
ENV NODE_ENV=production
COPY --from=build /app/.next ./.next
COPY --from=build /app/public ./public
COPY --from=build /app/package.json ./package.json
COPY --from=build /app/node_modules ./node_modules
EXPOSE 3000
CMD ["npm", "start"]
```

- [ ] **Step 5: Verify build**

Run: `cd frontend && npm run build`
Expected: build succeeds.

- [ ] **Step 6: Commit**

```bash
git add frontend
git commit -m "feat: scaffold NextJS app with Tailwind, shadcn, api client"
```

---

## Phase 1 — Database, Migrations, sqlc, BNCC seed

### Task 1.1: Schema migrations (goose)

**Files:**
- Create: `backend/migrations/00001_init.sql`
- Create: `backend/sqlc.yaml`
- Create: `backend/Makefile`

**Interfaces:**
- Produces: tables `users, sessions, bncc_skills, lesson_plans, study_trails, trail_topics, quizzes, quiz_questions, student_attempts, attempt_answers`.

- [ ] **Step 1: Write migration**

`backend/migrations/00001_init.sql`:
```sql
-- +goose Up
CREATE TABLE users (
  id BIGSERIAL PRIMARY KEY,
  email TEXT UNIQUE NOT NULL,
  senha_hash TEXT NOT NULL,
  nome TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE sessions (
  id TEXT PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  expires_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE bncc_skills (
  id BIGSERIAL PRIMARY KEY,
  code TEXT UNIQUE NOT NULL,
  disciplina TEXT NOT NULL,
  ano TEXT NOT NULL,
  descricao TEXT NOT NULL
);

CREATE TABLE lesson_plans (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  bncc_skill_id BIGINT REFERENCES bncc_skills(id),
  duracao_min INT NOT NULL DEFAULT 50,
  origem TEXT NOT NULL CHECK (origem IN ('manual','ia','ia_aprimorado')),
  status TEXT NOT NULL DEFAULT 'rascunho' CHECK (status IN ('rascunho','pronto','falha')),
  objetivos TEXT NOT NULL DEFAULT '',
  metodologia TEXT NOT NULL DEFAULT '',
  recursos TEXT NOT NULL DEFAULT '',
  avaliacao TEXT NOT NULL DEFAULT '',
  atividade TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE study_trails (
  id BIGSERIAL PRIMARY KEY,
  lesson_plan_id BIGINT NOT NULL UNIQUE REFERENCES lesson_plans(id) ON DELETE CASCADE,
  codigo TEXT UNIQUE,
  publicada_em TIMESTAMPTZ
);

CREATE TABLE trail_topics (
  id BIGSERIAL PRIMARY KEY,
  study_trail_id BIGINT NOT NULL REFERENCES study_trails(id) ON DELETE CASCADE,
  ordem INT NOT NULL,
  titulo TEXT NOT NULL,
  resumo TEXT NOT NULL
);

CREATE TABLE quizzes (
  id BIGSERIAL PRIMARY KEY,
  study_trail_id BIGINT NOT NULL UNIQUE REFERENCES study_trails(id) ON DELETE CASCADE
);

CREATE TABLE quiz_questions (
  id BIGSERIAL PRIMARY KEY,
  quiz_id BIGINT NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
  ordem INT NOT NULL,
  enunciado TEXT NOT NULL,
  opcoes JSONB NOT NULL,
  correta INT NOT NULL
);

CREATE TABLE student_attempts (
  id BIGSERIAL PRIMARY KEY,
  study_trail_id BIGINT NOT NULL REFERENCES study_trails(id) ON DELETE CASCADE,
  nome_aluno TEXT NOT NULL,
  pontos INT NOT NULL DEFAULT 0,
  concluido_em TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE attempt_answers (
  id BIGSERIAL PRIMARY KEY,
  student_attempt_id BIGINT NOT NULL REFERENCES student_attempts(id) ON DELETE CASCADE,
  quiz_question_id BIGINT NOT NULL REFERENCES quiz_questions(id) ON DELETE CASCADE,
  escolhida INT NOT NULL,
  correta BOOLEAN NOT NULL
);

-- +goose Down
DROP TABLE attempt_answers, student_attempts, quiz_questions, quizzes,
  trail_topics, study_trails, lesson_plans, bncc_skills, sessions, users;
```

- [ ] **Step 2: Write `backend/Makefile`**

```makefile
DB_URL ?= postgres://planejai:planejai@localhost:5432/planejai?sslmode=disable

migrate-up:
	goose -dir migrations postgres "$(DB_URL)" up
migrate-down:
	goose -dir migrations postgres "$(DB_URL)" down
sqlc:
	sqlc generate
test:
	go test ./...
```

- [ ] **Step 3: Apply migration against a running postgres**

Run: `docker compose up -d db && cd backend && go run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations postgres "$DB_URL" up`
Expected: `OK 00001_init.sql`.

- [ ] **Step 4: Commit**

```bash
git add backend/migrations backend/Makefile
git commit -m "feat: initial database schema (goose)"
```

---

### Task 1.2: sqlc queries + generated store

**Files:**
- Create: `backend/sqlc.yaml`
- Create: `backend/internal/store/queries/*.sql`
- Generate: `backend/internal/store/*.go` (sqlc output)

**Interfaces:**
- Produces: `store.Queries` with methods used by later tasks:
  `CreateUser`, `GetUserByEmail`, `GetUserByID`,
  `CreateSession`, `GetSession`, `DeleteSession`, `DeleteExpiredSessions`,
  `ListBnccSkills`, `GetBnccSkill`, `CountBnccSkills`, `InsertBnccSkill`,
  `CreateLessonPlan`, `UpdateLessonPlan`, `GetLessonPlan`, `ListLessonPlansByUser`, `SetLessonStatus`,
  `CreateTrail`, `GetTrailByLesson`, `GetTrailByCode`, `PublishTrail`,
  `ReplaceTopics`(via delete+insert), `InsertTopic`, `ListTopics`,
  `CreateQuiz`, `GetQuizByTrail`, `InsertQuestion`, `DeleteQuestionsByQuiz`, `ListQuestions`, `ListQuestionsPublic`,
  `CreateAttempt`, `GetAttempt`, `CompleteAttempt`, `InsertAnswer`, `TrailStats`.

- [ ] **Step 1: Write `backend/sqlc.yaml`**

```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "internal/store/queries"
    schema: "migrations"
    gen:
      go:
        package: "store"
        out: "internal/store"
        sql_package: "pgx/v5"
        emit_json_tags: true
        emit_pointers_for_null_types: true
```

- [ ] **Step 2: Write query files**

`backend/internal/store/queries/users.sql`:
```sql
-- name: CreateUser :one
INSERT INTO users (email, senha_hash, nome) VALUES ($1,$2,$3) RETURNING *;
-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;
-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;
```

`backend/internal/store/queries/sessions.sql`:
```sql
-- name: CreateSession :one
INSERT INTO sessions (id, user_id, expires_at) VALUES ($1,$2,$3) RETURNING *;
-- name: GetSession :one
SELECT * FROM sessions WHERE id = $1 AND expires_at > now();
-- name: DeleteSession :exec
DELETE FROM sessions WHERE id = $1;
-- name: DeleteExpiredSessions :exec
DELETE FROM sessions WHERE expires_at <= now();
```

`backend/internal/store/queries/bncc.sql`:
```sql
-- name: CountBnccSkills :one
SELECT count(*) FROM bncc_skills;
-- name: InsertBnccSkill :exec
INSERT INTO bncc_skills (code, disciplina, ano, descricao) VALUES ($1,$2,$3,$4)
ON CONFLICT (code) DO NOTHING;
-- name: ListBnccSkills :many
SELECT * FROM bncc_skills
WHERE ($1::text = '' OR disciplina = $1) AND ($2::text = '' OR ano = $2)
ORDER BY code;
-- name: GetBnccSkill :one
SELECT * FROM bncc_skills WHERE id = $1;
```

`backend/internal/store/queries/lessons.sql`:
```sql
-- name: CreateLessonPlan :one
INSERT INTO lesson_plans
  (user_id, bncc_skill_id, duracao_min, origem, status, objetivos, metodologia, recursos, avaliacao, atividade)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING *;
-- name: UpdateLessonPlan :one
UPDATE lesson_plans SET
  bncc_skill_id=$2, duracao_min=$3, objetivos=$4, metodologia=$5,
  recursos=$6, avaliacao=$7, atividade=$8, origem=$9, updated_at=now()
WHERE id=$1 RETURNING *;
-- name: SetLessonStatus :exec
UPDATE lesson_plans SET status=$2, updated_at=now() WHERE id=$1;
-- name: GetLessonPlan :one
SELECT * FROM lesson_plans WHERE id=$1;
-- name: ListLessonPlansByUser :many
SELECT * FROM lesson_plans WHERE user_id=$1 ORDER BY created_at DESC;
```

`backend/internal/store/queries/trails.sql`:
```sql
-- name: CreateTrail :one
INSERT INTO study_trails (lesson_plan_id) VALUES ($1) RETURNING *;
-- name: GetTrailByLesson :one
SELECT * FROM study_trails WHERE lesson_plan_id=$1;
-- name: GetTrailByCode :one
SELECT * FROM study_trails WHERE codigo=$1;
-- name: PublishTrail :one
UPDATE study_trails SET codigo=$2, publicada_em=now() WHERE id=$1 RETURNING *;
-- name: DeleteTopics :exec
DELETE FROM trail_topics WHERE study_trail_id=$1;
-- name: InsertTopic :exec
INSERT INTO trail_topics (study_trail_id, ordem, titulo, resumo) VALUES ($1,$2,$3,$4);
-- name: ListTopics :many
SELECT * FROM trail_topics WHERE study_trail_id=$1 ORDER BY ordem;
```

`backend/internal/store/queries/quiz.sql`:
```sql
-- name: CreateQuiz :one
INSERT INTO quizzes (study_trail_id) VALUES ($1)
ON CONFLICT (study_trail_id) DO UPDATE SET study_trail_id=EXCLUDED.study_trail_id
RETURNING *;
-- name: GetQuizByTrail :one
SELECT * FROM quizzes WHERE study_trail_id=$1;
-- name: DeleteQuestionsByQuiz :exec
DELETE FROM quiz_questions WHERE quiz_id=$1;
-- name: InsertQuestion :exec
INSERT INTO quiz_questions (quiz_id, ordem, enunciado, opcoes, correta) VALUES ($1,$2,$3,$4,$5);
-- name: ListQuestions :many
SELECT * FROM quiz_questions WHERE quiz_id=$1 ORDER BY ordem;
-- name: ListQuestionsPublic :many
SELECT id, ordem, enunciado, opcoes FROM quiz_questions WHERE quiz_id=$1 ORDER BY ordem;
```

`backend/internal/store/queries/attempts.sql`:
```sql
-- name: CreateAttempt :one
INSERT INTO student_attempts (study_trail_id, nome_aluno) VALUES ($1,$2) RETURNING *;
-- name: GetAttempt :one
SELECT * FROM student_attempts WHERE id=$1;
-- name: CompleteAttempt :one
UPDATE student_attempts SET pontos=$2, concluido_em=now() WHERE id=$1 RETURNING *;
-- name: InsertAnswer :exec
INSERT INTO attempt_answers (student_attempt_id, quiz_question_id, escolhida, correta)
VALUES ($1,$2,$3,$4);
-- name: TrailStats :many
SELECT nome_aluno, pontos, concluido_em, created_at
FROM student_attempts WHERE study_trail_id=$1 ORDER BY created_at DESC;
```

- [ ] **Step 3: Generate**

Run: `cd backend && go run github.com/sqlc-dev/sqlc/cmd/sqlc@latest generate && go mod tidy && go build ./...`
Expected: `internal/store/*.go` generated, build passes.

- [ ] **Step 4: Commit**

```bash
git add backend/sqlc.yaml backend/internal/store
git commit -m "feat: sqlc queries and generated store"
```

---

### Task 1.3: DB connection pool + store wiring

**Files:**
- Create: `backend/internal/store/db.go`
- Modify: `backend/cmd/api/main.go`
- Modify: `backend/internal/http/router.go` (add `Store *store.Queries` to `Deps`)

**Interfaces:**
- Produces: `store.Connect(ctx, url) (*pgxpool.Pool, error)`; `Deps.Store *store.Queries`, `Deps.Pool *pgxpool.Pool`.

- [ ] **Step 1: Write `db.go`**

```go
package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, url string) (*pgxpool.Pool, error) {
	return pgxpool.New(ctx, url)
}
```

- [ ] **Step 2: Extend `Deps`** in `router.go`:

```go
type Deps struct {
	Store *store.Queries
	Pool  *pgxpool.Pool
}
```
(add imports for `store` and `pgxpool`.)

- [ ] **Step 3: Wire in `main.go`**

```go
pool, err := store.Connect(ctx, os.Getenv("DATABASE_URL"))
if err != nil { log.Fatal(err) }
defer pool.Close()
deps := apihttp.Deps{Store: store.New(pool), Pool: pool}
```
(use `context.Background()`, pass `deps` to `NewRouter`.)

- [ ] **Step 4: Build**

Run: `cd backend && go mod tidy && go build ./...`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/store/db.go backend/internal/http/router.go backend/cmd/api/main.go
git commit -m "feat: pgx pool and store wiring"
```

---

### Task 1.4: BNCC seed

**Files:**
- Create: `backend/seed/bncc.json`
- Create: `backend/internal/seed/bncc.go`
- Modify: `backend/cmd/api/main.go` (run seed on boot)
- Test: `backend/internal/seed/bncc_test.go`

**Interfaces:**
- Consumes: `store.Queries.CountBnccSkills`, `InsertBnccSkill`.
- Produces: `seed.BNCC(ctx, q *store.Queries, path string) error` — idempotent; inserts only when table empty.

- [ ] **Step 1: Write `seed/bncc.json`** (start with ~12 real BNCC skills across disciplines/anos)

```json
[
  {"code":"EF67LP28","disciplina":"Língua Portuguesa","ano":"7º ano","descricao":"Ler e compreender textos argumentativos identificando tese e argumentos."},
  {"code":"EF07MA09","disciplina":"Matemática","ano":"7º ano","descricao":"Resolver problemas com números racionais em operações de adição e subtração."}
]
```
(Fill to ~12 entries covering Português, Matemática, Ciências, História, Geografia across 6º–9º ano.)

- [ ] **Step 2: Write failing test**

```go
package seed

import (
	"encoding/json"
	"os"
	"testing"
)

func TestBnccJSONParses(t *testing.T) {
	b, err := os.ReadFile("../../seed/bncc.json")
	if err != nil { t.Fatal(err) }
	var items []struct {
		Code, Disciplina, Ano, Descricao string
	}
	if err := json.Unmarshal(b, &items); err != nil { t.Fatal(err) }
	if len(items) < 10 { t.Fatalf("want >=10 skills got %d", len(items)) }
	for _, it := range items {
		if it.Code == "" || it.Disciplina == "" || it.Ano == "" || it.Descricao == "" {
			t.Fatalf("empty field in %+v", it)
		}
	}
}
```

- [ ] **Step 3: Run test — expect PASS once JSON filled**

Run: `cd backend && go test ./internal/seed/ -run TestBnccJSONParses`
Expected: PASS.

- [ ] **Step 4: Write `seed/bncc.go`**

```go
package seed

import (
	"context"
	"encoding/json"
	"os"

	"github.com/Carlos-hub/planejai/backend/internal/store"
)

type skill struct {
	Code       string `json:"code"`
	Disciplina string `json:"disciplina"`
	Ano        string `json:"ano"`
	Descricao  string `json:"descricao"`
}

func BNCC(ctx context.Context, q *store.Queries, path string) error {
	n, err := q.CountBnccSkills(ctx)
	if err != nil { return err }
	if n > 0 { return nil }
	b, err := os.ReadFile(path)
	if err != nil { return err }
	var items []skill
	if err := json.Unmarshal(b, &items); err != nil { return err }
	for _, it := range items {
		if err := q.InsertBnccSkill(ctx, store.InsertBnccSkillParams{
			Code: it.Code, Disciplina: it.Disciplina, Ano: it.Ano, Descricao: it.Descricao,
		}); err != nil { return err }
	}
	return nil
}
```

- [ ] **Step 5: Call in `main.go`** after migrations/store init:

```go
if err := seed.BNCC(ctx, deps.Store, "seed/bncc.json"); err != nil { log.Printf("seed: %v", err) }
```

- [ ] **Step 6: Build + commit**

```bash
cd backend && go build ./... && cd ..
git add backend/seed backend/internal/seed backend/cmd/api/main.go
git commit -m "feat: BNCC static seed"
```

---

## Phase 2 — Auth (server-side sessions)

### Task 2.1: Password hashing + session helpers

**Files:**
- Create: `backend/internal/auth/auth.go`
- Test: `backend/internal/auth/auth_test.go`

**Interfaces:**
- Produces: `HashPassword(p string) (string, error)`, `CheckPassword(hash, p string) bool`, `NewSessionID() (string, error)` (32-byte base64url), `SessionTTL = 7*24h`.

- [ ] **Step 1: Write failing test**

```go
package auth

import "testing"

func TestHashAndCheck(t *testing.T) {
	h, err := HashPassword("segredo123")
	if err != nil { t.Fatal(err) }
	if !CheckPassword(h, "segredo123") { t.Fatal("should match") }
	if CheckPassword(h, "errado") { t.Fatal("should not match") }
}

func TestSessionIDUnique(t *testing.T) {
	a, _ := NewSessionID()
	b, _ := NewSessionID()
	if a == b || len(a) < 20 { t.Fatalf("bad session ids %q %q", a, b) }
}
```

- [ ] **Step 2: Run — expect FAIL**

Run: `cd backend && go test ./internal/auth/`
Expected: FAIL (undefined).

- [ ] **Step 3: Implement**

```go
package auth

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const SessionTTL = 7 * 24 * time.Hour

func HashPassword(p string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	return string(b), err
}

func CheckPassword(hash, p string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(p)) == nil
}

func NewSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil { return "", err }
	return base64.RawURLEncoding.EncodeToString(b), nil
}
```

- [ ] **Step 4: Run — expect PASS**

Run: `cd backend && go get golang.org/x/crypto/bcrypt && go mod tidy && go test ./internal/auth/`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/auth backend/go.mod backend/go.sum
git commit -m "feat: password hashing and session id helpers"
```

---

### Task 2.2: Auth middleware + login/logout/me handlers

**Files:**
- Create: `backend/internal/http/auth_handlers.go`
- Create: `backend/internal/http/middleware.go`
- Modify: `backend/internal/http/router.go` (routes + `SessionSecret`, cookie helpers)
- Test: `backend/internal/http/auth_handlers_test.go` (uses a real test DB)

**Interfaces:**
- Consumes: `store` session/user queries, `auth` helpers.
- Produces: `POST /api/auth/login {email,senha}` → 200 + `Set-Cookie: sid`; `POST /api/auth/logout` → 204; `GET /api/me` → `{id,email,nome}`; `RequireAuth` middleware injecting `userID` into context via `userIDFromContext(r) (int64, bool)`.

- [ ] **Step 1: Write test helper for DB** (`backend/internal/http/testmain_test.go`)

```go
package http

import (
	"context"
	"os"
	"testing"

	"github.com/Carlos-hub/planejai/backend/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

func testPool(t *testing.T) *pgxpool.Pool {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" { t.Skip("TEST_DATABASE_URL not set") }
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil { t.Fatal(err) }
	t.Cleanup(pool.Close)
	return pool
}

func testDeps(t *testing.T) Deps {
	pool := testPool(t)
	return Deps{Store: store.New(pool), Pool: pool, SessionSecret: "test-secret"}
}
```

- [ ] **Step 2: Write failing test**

```go
func TestLoginFlow(t *testing.T) {
	d := testDeps(t)
	ctx := context.Background()
	// seed a user
	h, _ := auth.HashPassword("segredo123")
	u, err := d.Store.CreateUser(ctx, store.CreateUserParams{
		Email: "prof@x.com", SenhaHash: h, Nome: "Prof",
	})
	if err != nil { t.Fatal(err) }
	t.Cleanup(func() { d.Pool.Exec(ctx, "DELETE FROM users WHERE id=$1", u.ID) })

	srv := httptest.NewServer(NewRouter(d))
	defer srv.Close()
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}

	body := `{"email":"prof@x.com","senha":"segredo123"}`
	resp, err := client.Post(srv.URL+"/api/auth/login", "application/json", strings.NewReader(body))
	if err != nil { t.Fatal(err) }
	if resp.StatusCode != 200 { t.Fatalf("login want 200 got %d", resp.StatusCode) }

	me, _ := client.Get(srv.URL + "/api/me")
	if me.StatusCode != 200 { t.Fatalf("me want 200 got %d", me.StatusCode) }
}
```
(Imports: `context, net/http, net/http/cookiejar, net/http/httptest, strings, testing`, plus `auth`, `store`.)

- [ ] **Step 3: Run — expect FAIL / skip if no DB**

Run: `TEST_DATABASE_URL=$DB_URL go test ./internal/http/ -run TestLoginFlow`
Expected: FAIL (routes not implemented).

- [ ] **Step 4: Implement middleware**

`backend/internal/http/middleware.go`:
```go
package http

import (
	"context"
	"net/http"
)

type ctxKey string

const userIDKey ctxKey = "userID"

func (d Deps) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("sid")
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "não autenticado"})
			return
		}
		sess, err := d.Store.GetSession(r.Context(), c.Value)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "sessão inválida"})
			return
		}
		ctx := context.WithValue(r.Context(), userIDKey, sess.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func userIDFromContext(r *http.Request) (int64, bool) {
	v, ok := r.Context().Value(userIDKey).(int64)
	return v, ok
}
```

- [ ] **Step 5: Implement handlers**

`backend/internal/http/auth_handlers.go`:
```go
package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Carlos-hub/planejai/backend/internal/auth"
	"github.com/Carlos-hub/planejai/backend/internal/store"
)

func (d Deps) login(w http.ResponseWriter, r *http.Request) {
	var in struct{ Email, Senha string }
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, 400, map[string]string{"error": "json inválido"}); return
	}
	u, err := d.Store.GetUserByEmail(r.Context(), in.Email)
	if err != nil || !auth.CheckPassword(u.SenhaHash, in.Senha) {
		writeJSON(w, 401, map[string]string{"error": "credenciais inválidas"}); return
	}
	sid, _ := auth.NewSessionID()
	exp := time.Now().Add(auth.SessionTTL)
	if _, err := d.Store.CreateSession(r.Context(), store.CreateSessionParams{
		ID: sid, UserID: u.ID, ExpiresAt: pgTime(exp),
	}); err != nil { writeJSON(w, 500, map[string]string{"error": "erro"}); return }
	http.SetCookie(w, &http.Cookie{
		Name: "sid", Value: sid, Path: "/", HttpOnly: true,
		SameSite: http.SameSiteLaxMode, Expires: exp,
	})
	writeJSON(w, 200, map[string]any{"id": u.ID, "email": u.Email, "nome": u.Nome})
}

func (d Deps) logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie("sid"); err == nil {
		_ = d.Store.DeleteSession(r.Context(), c.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: "sid", Value: "", Path: "/", MaxAge: -1})
	w.WriteHeader(204)
}

func (d Deps) me(w http.ResponseWriter, r *http.Request) {
	uid, _ := userIDFromContext(r)
	u, err := d.Store.GetUserByID(r.Context(), uid)
	if err != nil { writeJSON(w, 404, map[string]string{"error": "não encontrado"}); return }
	writeJSON(w, 200, map[string]any{"id": u.ID, "email": u.Email, "nome": u.Nome})
}
```

- [ ] **Step 6: Add `pgTime` helper + CORS + routes** in `router.go`

Add to `Deps`: `SessionSecret string`. Add CORS middleware allowing `NEXT_PUBLIC` origin with credentials. Add:
```go
r.Use(corsMiddleware) // allow localhost:3000, credentials true
r.Route("/api", func(r chi.Router) {
	r.Post("/auth/login", d.login)
	r.Post("/auth/logout", d.logout)
	r.Group(func(r chi.Router) {
		r.Use(d.RequireAuth)
		r.Get("/me", d.me)
	})
})
```
`pgTime`:
```go
func pgTime(t time.Time) pgtype.Timestamptz { return pgtype.Timestamptz{Time: t, Valid: true} }
```

- [ ] **Step 7: Run — expect PASS (with test DB)**

Run: `TEST_DATABASE_URL=$DB_URL go test ./internal/http/ -run TestLoginFlow`
Expected: PASS.

- [ ] **Step 8: Commit**

```bash
git add backend/internal/http
git commit -m "feat: session auth (login/logout/me) + RequireAuth middleware"
```

---

### Task 2.3: Register endpoint + CLI seed of a demo teacher

**Files:**
- Modify: `backend/internal/http/auth_handlers.go` (add `register`)
- Modify: `backend/internal/http/router.go` (route)
- Modify: `backend/internal/seed/bncc.go` → add `DemoTeacher`
- Modify: `backend/cmd/api/main.go`

**Interfaces:**
- Produces: `POST /api/auth/register {email,senha,nome}` → 201 + login cookie; `seed.DemoTeacher(ctx,q)` inserts `prof@demo.com / demo1234` if absent.

- [ ] **Step 1: Write `register` handler** (hash senha, CreateUser, set cookie — mirror `login`'s cookie code). Return 201.

- [ ] **Step 2: Add `seed.DemoTeacher`** using `GetUserByEmail`; if not found, `HashPassword` + `CreateUser`.

- [ ] **Step 3: Wire route + call seed in main.**

- [ ] **Step 4: Manual smoke test**

Run: `docker compose up -d db && cd backend && go run ./cmd/api &` then `curl -i -X POST localhost:8080/api/auth/login -d '{"email":"prof@demo.com","senha":"demo1234"}'`
Expected: 200 + `Set-Cookie: sid=...`.

- [ ] **Step 5: Commit**

```bash
git add backend/internal
git commit -m "feat: register endpoint and demo teacher seed"
```

---

## Phase 3 — LLM Generator (brand-agnostic)

### Task 3.1: LessonData domain type + JSON contract

**Files:**
- Create: `backend/internal/lesson/data.go`
- Test: `backend/internal/lesson/data_test.go`

**Interfaces:**
- Produces:
```go
type LessonData struct {
	Plano     Plano  `json:"plano"`
	Atividade string `json:"atividade"`
	Trilha    Trilha `json:"trilha"`
}
type Plano struct{ Objetivos, Metodologia, Recursos, Avaliacao string }
type Trilha struct {
	Topicos []Topico `json:"topicos"`
	Quiz    Quiz     `json:"quiz"`
}
type Topico struct{ Titulo, Resumo string }
type Quiz struct{ Questoes []Questao `json:"questoes"` }
type Questao struct {
	Enunciado string   `json:"enunciado"`
	Opcoes    []string `json:"opcoes"`
	Correta   int      `json:"correta"`
}
```
plus `ParseLessonData(raw []byte) (LessonData, error)` that validates: ≥1 topico, ≥1 questao, each questao has ≥2 opcoes and `0 <= correta < len(opcoes)`.

- [ ] **Step 1: Write failing test** using a fixture matching the spec JSON; assert parse + validation rejects `correta` out of range.

```go
func TestParseLessonData(t *testing.T) {
	raw := []byte(`{"plano":{"objetivos":"o","metodologia":"m","recursos":"r","avaliacao":"a"},"atividade":"at","trilha":{"topicos":[{"titulo":"t","resumo":"r"}],"quiz":{"questoes":[{"enunciado":"e","opcoes":["x","y"],"correta":1}]}}}`)
	d, err := ParseLessonData(raw)
	if err != nil { t.Fatal(err) }
	if len(d.Trilha.Topicos) != 1 || d.Trilha.Quiz.Questoes[0].Correta != 1 {
		t.Fatal("parse mismatch")
	}
}
func TestParseRejectsBadCorreta(t *testing.T) {
	raw := []byte(`{"plano":{},"atividade":"a","trilha":{"topicos":[{"titulo":"t","resumo":"r"}],"quiz":{"questoes":[{"enunciado":"e","opcoes":["x","y"],"correta":5}]}}}`)
	if _, err := ParseLessonData(raw); err == nil { t.Fatal("want validation error") }
}
```

- [ ] **Step 2: Run — FAIL. Step 3: Implement types + `ParseLessonData`. Step 4: Run — PASS.**

Run: `cd backend && go test ./internal/lesson/ -run TestParse`

- [ ] **Step 5: Commit**

```bash
git add backend/internal/lesson/data.go backend/internal/lesson/data_test.go
git commit -m "feat: LessonData domain type and validating parser"
```

---

### Task 3.2: Generator interface + mock

**Files:**
- Create: `backend/internal/lesson/generator.go`
- Create: `backend/internal/lesson/mock.go`
- Test: `backend/internal/lesson/mock_test.go`

**Interfaces:**
- Produces:
```go
type Generator interface {
	Generate(ctx context.Context, skill store.BnccSkill, duracaoMin int) (LessonData, error)
	Enhance(ctx context.Context, draft LessonData, skill store.BnccSkill) (LessonData, error)
}
```
and `MockGenerator` returning a canned `LessonData` (used by handler tests).

- [ ] **Step 1: Write mock test** asserting `MockGenerator.Generate` returns a valid `LessonData` (passes `ParseLessonData` re-serialization). Step 2: FAIL. Step 3: Implement interface + mock. Step 4: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/lesson/generator.go backend/internal/lesson/mock.go backend/internal/lesson/mock_test.go
git commit -m "feat: Generator interface and mock"
```

---

### Task 3.3: LangChainGo adapter

**Files:**
- Create: `backend/internal/lesson/langchain.go`
- Create: `backend/internal/lesson/prompts.go`
- Test: `backend/internal/lesson/langchain_test.go` (guarded by `ANTHROPIC_API_KEY`, else skip)

**Interfaces:**
- Consumes: `Generator`, `ParseLessonData`, env `LLM_PROVIDER`, `LLM_MODEL`.
- Produces: `NewLangChainGenerator() (Generator, error)` selecting provider by `LLM_PROVIDER` (`anthropic` supported now; switch statement leaves room for others).

- [ ] **Step 1: Write prompts** (`prompts.go`): a `systemPrompt` instructing strict JSON matching the schema (Portuguese fields), a `generateUserPrompt(skill, dur)` and `enhanceUserPrompt(skill, draftJSON)`.

- [ ] **Step 2: Implement adapter** using `github.com/tmc/langchaingo/llms/anthropic`; call `llms.GenerateFromSinglePrompt` with system+user, then `ParseLessonData([]byte(resp))`. On non-JSON, attempt to extract the first `{...}` block; if still invalid, return error (handler sets status=falha).

```go
func NewLangChainGenerator() (Generator, error) {
	provider := getenv("LLM_PROVIDER", "anthropic")
	model := getenv("LLM_MODEL", "claude-opus-4-8")
	switch provider {
	case "anthropic":
		m, err := anthropic.New(anthropic.WithModel(model))
		if err != nil { return nil, err }
		return &lcGen{llm: m}, nil
	default:
		return nil, fmt.Errorf("provider não suportado: %s", provider)
	}
}
```

- [ ] **Step 3: Test** (skipped without API key): assert `Generate` returns a valid LessonData for a sample skill.

Run: `cd backend && go get github.com/tmc/langchaingo && go mod tidy && go test ./internal/lesson/`
Expected: PASS (LLM test skipped if no key).

- [ ] **Step 4: Commit**

```bash
git add backend/internal/lesson backend/go.mod backend/go.sum
git commit -m "feat: langchaingo generator adapter (anthropic)"
```

---

## Phase 4 — Lesson endpoints (manual / generate / enhance / edit)

### Task 4.1: Persist LessonData helper (lesson + trail + topics + quiz in a tx)

**Files:**
- Create: `backend/internal/http/lesson_persist.go`
- Test: `backend/internal/http/lesson_persist_test.go`

**Interfaces:**
- Produces: `(d Deps) saveLessonContent(ctx, lessonID int64, data lesson.LessonData) error` — upserts trail for the lesson, replaces topics, upserts quiz, replaces questions; runs in a pgx transaction; sets lesson fields (objetivos…atividade) and `status='pronto'`.

- [ ] **Step 1: Write failing test** (test DB): create user+lesson, call `saveLessonContent` with mock data, assert topics+questions rows exist and lesson status='pronto'.

- [ ] **Step 2: FAIL. Step 3: Implement** using `d.Pool.Begin`, `store.New(tx).WithTx`... (use `q := d.Store.WithTx(tx)`), `GetTrailByLesson` or `CreateTrail`, `DeleteTopics`+`InsertTopic` loop, `CreateQuiz`, `DeleteQuestionsByQuiz`+`InsertQuestion` loop (marshal `opcoes` to JSON), `UpdateLessonPlan`, `SetLessonStatus`. Commit. **Step 4: PASS.**

- [ ] **Step 5: Commit**

```bash
git add backend/internal/http/lesson_persist.go backend/internal/http/lesson_persist_test.go
git commit -m "feat: transactional persistence of lesson content"
```

---

### Task 4.2: Manual create + edit + list + get

**Files:**
- Create: `backend/internal/http/lesson_handlers.go`
- Modify: `backend/internal/http/router.go` (routes under RequireAuth)
- Test: `backend/internal/http/lesson_handlers_test.go`

**Interfaces:**
- Produces (all under `/api`, RequireAuth, scoped to `userID`):
  `POST /api/lessons` (manual create, `origem=manual`, status per content), `GET /api/lessons`, `GET /api/lessons/:id`, `PATCH /api/lessons/:id`. Ownership enforced: 404 if lesson.user_id ≠ userID.

- [ ] **Step 1: Write failing test**: login (reuse helper), POST manual lesson, GET list contains it, PATCH updates `objetivos`, GET by id reflects change; a second user gets 404 on the first user's lesson.

- [ ] **Step 2: FAIL. Step 3: Implement handlers** (decode DTO with pt-BR fields incl. `topicos`/`quiz`, build `lesson.LessonData`, `CreateLessonPlan` then `saveLessonContent`). Enforce ownership via `GetLessonPlan` + compare userID. **Step 4: PASS.**

- [ ] **Step 5: Commit**

```bash
git add backend/internal/http
git commit -m "feat: manual lesson CRUD endpoints"
```

---

### Task 4.3: AI generate + enhance endpoints

**Files:**
- Modify: `backend/internal/http/lesson_handlers.go`
- Modify: `backend/internal/http/router.go` (add `Gen lesson.Generator` to `Deps`, routes)
- Modify: `backend/cmd/api/main.go` (construct generator)
- Test: `backend/internal/http/lesson_ai_test.go` (inject `lesson.MockGenerator`)

**Interfaces:**
- Consumes: `Deps.Gen lesson.Generator`, `saveLessonContent`.
- Produces: `POST /api/lessons/generate {bncc_skill_id,duracao}` → creates lesson `origem=ia`, calls `Gen.Generate`, persists, returns lesson; on generator error sets `status=falha` and returns 502. `POST /api/lessons/:id/enhance` → loads current lesson as `LessonData` draft, calls `Gen.Enhance`, persists with `origem=ia_aprimorado`.

- [ ] **Step 1: Write failing test** with `Deps.Gen = lesson.MockGenerator{}`: POST generate returns lesson with topics; enhance on an existing lesson updates origem to `ia_aprimorado`.

- [ ] **Step 2: FAIL. Step 3: Implement.** Add helper `lessonToData(plan, topics, questions) lesson.LessonData` for enhance. **Step 4: PASS.**

- [ ] **Step 5: Wire real generator in main:** `gen, err := lesson.NewLangChainGenerator()` (log + continue with nil-safe guard returning 503 if nil). Assign `deps.Gen = gen`.

- [ ] **Step 6: Commit**

```bash
git add backend/internal backend/cmd
git commit -m "feat: AI generate and enhance lesson endpoints"
```

---

## Phase 5 — Trail publish + student access + quiz scoring

### Task 5.1: Trail code generator

**Files:**
- Create: `backend/internal/domain/code.go`
- Test: `backend/internal/domain/code_test.go`

**Interfaces:**
- Produces: `NewTrailCode() string` → `TR-XXXX`, alphabet `ABCDEFGHJKLMNPQRSTUVWXYZ23456789` (no O/0/1/I).

- [ ] **Step 1: Write test** asserting prefix `TR-`, length 7, only allowed chars, and 1000 codes are unique. Step 2: FAIL. Step 3: Implement with `crypto/rand`. Step 4: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/domain/code.go backend/internal/domain/code_test.go
git commit -m "feat: trail short-code generator"
```

---

### Task 5.2: Publish trail endpoint

**Files:**
- Modify: `backend/internal/http/lesson_handlers.go` (add `publishTrail`)
- Modify: `backend/internal/http/router.go`
- Test: `backend/internal/http/publish_test.go`

**Interfaces:**
- Produces: `POST /api/trails/:id/publish` (`:id` = lesson id, RequireAuth, ownership) → ensures lesson.status=`pronto`, generates unique code (retry on unique violation up to 5x), sets `publicada_em`, returns `{codigo, publica_url}`.

- [ ] **Step 1: Write failing test**: manual lesson with content → publish → response has `TR-` code; GetTrailByCode finds it. Step 2: FAIL. Step 3: Implement (retry loop on `PublishTrail` conflict). Step 4: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/http
git commit -m "feat: publish trail with unique short code"
```

---

### Task 5.3: Public trail read (no answer key)

**Files:**
- Create: `backend/internal/http/public_handlers.go`
- Modify: `backend/internal/http/router.go` (public routes, NO RequireAuth)
- Test: `backend/internal/http/public_test.go`

**Interfaces:**
- Produces: `GET /api/t/:code` → `{titulo_aula, topicos:[{ordem,titulo,resumo}], quiz:{questoes:[{id,enunciado,opcoes}]}}` — **no `correta`**. 404 if code missing/unpublished.

- [ ] **Step 1: Write failing test** asserting the JSON body does NOT contain the substring `"correta"` and includes questoes with ids. Step 2: FAIL. Step 3: Implement using `GetTrailByCode`, `ListTopics`, `GetQuizByTrail`, `ListQuestionsPublic`. Step 4: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/http
git commit -m "feat: public trail read without answer key"
```

---

### Task 5.4: Attempt + server-side scoring

**Files:**
- Modify: `backend/internal/http/public_handlers.go`
- Modify: `backend/internal/http/router.go`
- Create: `backend/internal/domain/scoring.go`
- Test: `backend/internal/domain/scoring_test.go`, `backend/internal/http/attempt_test.go`

**Interfaces:**
- Produces: `domain.Score(answers map[int64]int, correct map[int64]int) (pontos, acertos, total int)` — 10 pts per correct; `POST /api/t/:code/attempt {nome}` → `{attempt_id}`; `POST /api/attempts/:id/answers {answers:[{quiz_question_id,escolhida}]}` → `{pontos, acertos, total}`, persists answers with server-computed `correta`, completes attempt.

- [ ] **Step 1: Write scoring unit test** (pure function) covering all-correct, none, partial. Step 2: FAIL. Step 3: Implement `Score`. Step 4: PASS.

- [ ] **Step 5: Write attempt integration test**: create attempt, submit answers, assert pontos = 10×correct and attempt.concluido_em set. Step 6: FAIL. Step 7: Implement handlers — load `ListQuestions` (with `correta`) server-side, compute via `domain.Score`, `InsertAnswer` per answer, `CompleteAttempt`. Step 8: PASS.

- [ ] **Step 9: Commit**

```bash
git add backend/internal/domain backend/internal/http
git commit -m "feat: student attempts and server-side quiz scoring"
```

---

## Phase 6 — Class dashboard stats

### Task 6.1: Trail stats endpoint

**Files:**
- Modify: `backend/internal/http/lesson_handlers.go` (add `trailStats`)
- Modify: `backend/internal/http/router.go`
- Test: `backend/internal/http/stats_test.go`

**Interfaces:**
- Produces: `GET /api/trails/:id/stats` (lesson id, RequireAuth, ownership) → `{total_alunos, concluidos, media_pontos, tentativas:[{nome_aluno,pontos,concluido_em}]}` via `TrailStats`.

- [ ] **Step 1: Write failing test**: publish trail, create 2 attempts+answers, assert stats counts + average. Step 2: FAIL. Step 3: Implement (aggregate in Go from `TrailStats` rows). Step 4: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/http
git commit -m "feat: class dashboard stats endpoint"
```

---

## Phase 7 — PDF export

### Task 7.1: Trail PDF export

**Files:**
- Create: `backend/internal/http/export_handlers.go`
- Modify: `backend/internal/http/router.go` (public route)
- Test: `backend/internal/http/export_test.go`

**Interfaces:**
- Produces: `GET /api/t/:code/export.pdf` → `application/pdf` body (light layout: título, tópicos, quiz sem gabarito), 200. Uses `github.com/johnfercher/maroto/v2`.

- [ ] **Step 1: Write failing test** asserting `Content-Type: application/pdf`, status 200, and body starts with `%PDF`. Step 2: FAIL. Step 3: Implement with maroto (build doc from topics+questions, no `correta`). Step 4: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/http backend/go.mod backend/go.sum
git commit -m "feat: trail PDF export"
```

---

## Phase 8 — Frontend (teacher)

### Task 8.1: Auth pages + session context

**Files:**
- Create: `frontend/app/(teacher)/login/page.tsx`
- Create: `frontend/app/(teacher)/layout.tsx`
- Create: `frontend/lib/types.ts`, `frontend/lib/hooks.ts`

**Interfaces:**
- Consumes: `apiFetch`, backend auth endpoints.
- Produces: login form (email/senha) posting to `/api/auth/login`; `(teacher)/layout` redirects to `/login` when `GET /api/me` 401s; `types.ts` with `Lesson`, `LessonData`, `Trail`, `Stats` TS interfaces mirroring backend JSON.

- [ ] **Step 1: Write `lib/types.ts`** with interfaces matching backend field names (pt-BR).

- [ ] **Step 2: Build login page** (shadcn `Card`+`Input`+`Button`), on submit `apiFetch("/api/auth/login",{method:"POST",body:...})` then `router.push("/")`.

- [ ] **Step 3: Build teacher layout guard** (server component: fetch `/api/me`, redirect on failure) or client hook `useMe()`.

- [ ] **Step 4: Verify build** — `cd frontend && npm run build`. Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/app frontend/lib
git commit -m "feat(web): teacher login and session guard"
```

---

### Task 8.2: Lesson list + dashboard

**Files:**
- Create: `frontend/app/(teacher)/page.tsx` (dashboard: list lessons, "Nova aula")
- Create: `frontend/components/lesson-card.tsx`

**Interfaces:**
- Consumes: `GET /api/lessons`.
- Produces: dashboard listing lessons with status badge + origem + "Publicar"/"Ver trilha" links.

- [ ] **Step 1: Build page** fetching lessons, mapping to `LessonCard`. Step 2: Build `LessonCard`. Step 3: `npm run build`. Step 4: Commit.

```bash
git add frontend/app frontend/components
git commit -m "feat(web): teacher dashboard lesson list"
```

---

### Task 8.3: Lesson editor with 3 modes

**Files:**
- Create: `frontend/app/(teacher)/lessons/new/page.tsx`
- Create: `frontend/app/(teacher)/lessons/[id]/page.tsx`
- Create: `frontend/components/lesson-editor.tsx`
- Create: `frontend/components/mode-picker.tsx`

**Interfaces:**
- Consumes: `GET /api/bncc-skills`, `POST /api/lessons`, `POST /api/lessons/generate`, `POST /api/lessons/:id/enhance`, `PATCH /api/lessons/:id`.
- Produces: `ModePicker` (Manual / IA completa / IA aprimora); `LessonEditor` with editable fields for plano (objetivos/metodologia/recursos/avaliacao), atividade, tópicos (add/remove), quiz questões (enunciado/opções/correta). Manual → POST; IA completa → pick skill+duração → generate; IA aprimora → edit draft then enhance. All fields editable after any AI call.

- [ ] **Step 1: Build `ModePicker`** (3 shadcn cards). Step 2: Build `LessonEditor` controlled form with topic/question array editing. Step 3: Wire `new/page.tsx` (mode → action). Step 4: Wire `[id]/page.tsx` (load lesson, edit, PATCH, enhance). Step 5: `npm run build`. Step 6: Commit.

```bash
git add frontend/app frontend/components
git commit -m "feat(web): lesson editor with manual/AI/enhance modes"
```

---

### Task 8.4: Publish + class stats views

**Files:**
- Modify: `frontend/app/(teacher)/lessons/[id]/page.tsx` (publish button + share)
- Create: `frontend/app/(teacher)/lessons/[id]/stats/page.tsx`
- Create: `frontend/components/share-links.tsx`

**Interfaces:**
- Consumes: `POST /api/trails/:id/publish`, `GET /api/trails/:id/stats`.
- Produces: publish button → shows code + copy link + `wa.me` share (`https://wa.me/?text=<encoded trail url>`) + link to `export.pdf`; stats page with shadcn `Table` (aluno, pontos, concluído) + summary cards.

- [ ] **Step 1: Build `ShareLinks`** (code, copy, WhatsApp, PDF). Step 2: Add publish flow to lesson page. Step 3: Build stats page. Step 4: `npm run build`. Step 5: Commit.

```bash
git add frontend/app frontend/components
git commit -m "feat(web): publish, share links, class stats"
```

---

## Phase 9 — Frontend (student)

### Task 9.1: Public trail + focus mode

**Files:**
- Create: `frontend/app/t/[code]/page.tsx`
- Create: `frontend/components/trail-view.tsx`
- Create: `frontend/components/name-gate.tsx`

**Interfaces:**
- Consumes: `GET /api/t/:code`, `POST /api/t/:code/attempt`.
- Produces: name gate (typed name → create attempt, store attempt_id in state/localStorage); `TrailView` showing one tópico at a time (focus mode) with next/prev + progress bar (shadcn `Progress`).

- [ ] **Step 1: Build `NameGate`. Step 2: Build `TrailView` (one topic at a time). Step 3: Wire page. Step 4: `npm run build`. Step 5: Commit.**

```bash
git add frontend/app frontend/components
git commit -m "feat(web): public student trail with focus mode"
```

---

### Task 9.2: Quiz + gamification

**Files:**
- Create: `frontend/components/quiz-runner.tsx`
- Modify: `frontend/app/t/[code]/page.tsx`

**Interfaces:**
- Consumes: `POST /api/attempts/:id/answers`.
- Produces: `QuizRunner` — one question at a time, collect `escolhida`, submit all, show `{pontos, acertos, total}` result screen with points + progress; PDF/WhatsApp share of the trail.

- [ ] **Step 1: Build `QuizRunner`** (select option per question, submit). Step 2: Result screen with pontos + barra. Step 3: Wire into trail page after topics. Step 4: `npm run build`. Step 5: Commit.

```bash
git add frontend/app frontend/components
git commit -m "feat(web): self-graded quiz with score and gamification"
```

---

## Phase 10 — End-to-end verification

### Task 10.1: Compose up + smoke script

**Files:**
- Create: `scripts/smoke.sh`
- Modify: `README.md` (add smoke steps)

**Interfaces:**
- Produces: `scripts/smoke.sh` that (against `docker compose up`) logs in as demo teacher, generates a lesson (mock or real), publishes, reads public trail, submits an attempt, asserts non-zero pontos, fetches stats, downloads PDF.

- [ ] **Step 1: Write `scripts/smoke.sh`** using `curl` + `jq`, `set -euo pipefail`, cookie jar for teacher calls.

- [ ] **Step 2: Run full stack**

Run: `cp .env.example .env` (set `ANTHROPIC_API_KEY`), `docker compose up --build -d`, wait for healthy, `bash scripts/smoke.sh`.
Expected: script prints trail code, scored attempt, stats, `PDF OK`, exits 0.

- [ ] **Step 3: Run all backend tests**

Run: `cd backend && TEST_DATABASE_URL=$DB_URL go test ./...`
Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add scripts/smoke.sh README.md
git commit -m "test: end-to-end smoke script"
```

---

## Self-Review Notes (coverage map)

- Spec "Stack" → Phase 0 (compose, dockerfiles), 1 (sqlc/goose), 2 (sessions), 3 (langchaingo).
- Spec "Modelo de dados" → Task 1.1 (all tables incl. sessions).
- Spec "Modos de autoria" (manual/ia/ia_aprimorado) → Tasks 4.2, 4.3; frontend 8.3.
- Spec "API" endpoints → Phases 2–7 (each endpoint has a task).
- Spec "Contrato de IA" (interface + Enhance, env-driven, schema) → Tasks 3.1–3.3, Global Constraints.
- Spec answer-key never client-side → Tasks 5.3, 5.4, 7.1 (public queries omit `correta`; explicit test).
- Spec gamificação (pontos + progresso) → Tasks 5.4, 9.2.
- Spec export PDF/WhatsApp → Tasks 7.1, 8.4, 9.2.
- Spec dashboard turma → Task 6.1, 8.4.
- Spec testing (mock Generator, sqlc integration, feature flow) → Tasks 3.2, 4.x, 5.4, 10.1.
