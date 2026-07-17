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
