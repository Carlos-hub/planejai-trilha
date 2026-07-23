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
