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
