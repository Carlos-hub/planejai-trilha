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
