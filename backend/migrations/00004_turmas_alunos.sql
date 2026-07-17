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
