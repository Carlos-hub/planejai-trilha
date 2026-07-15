-- name: CreateAttempt :one
INSERT INTO student_attempts (study_trail_id, nome_aluno, student_id) VALUES ($1,$2,$3) RETURNING *;
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
