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
