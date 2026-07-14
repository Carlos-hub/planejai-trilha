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
