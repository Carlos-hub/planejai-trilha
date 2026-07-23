-- name: AttachTurmaLesson :one
INSERT INTO turma_lessons (turma_id, lesson_plan_id, ordem)
VALUES ($1, $2, COALESCE((SELECT max(ordem) FROM turma_lessons WHERE turma_id=$1), 0) + 1)
RETURNING *;

-- name: DetachTurmaLesson :exec
DELETE FROM turma_lessons WHERE turma_id=$1 AND lesson_plan_id=$2;

-- name: RenumberTurmaLessons :exec
WITH ordered AS (
  SELECT id, row_number() OVER (ORDER BY ordem) AS rn
  FROM turma_lessons tl2 WHERE tl2.turma_id=$1
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
       (sa.concluido_em IS NOT NULL)::boolean AS concluido
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
