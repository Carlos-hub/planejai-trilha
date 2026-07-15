-- name: CreateStudentSession :one
INSERT INTO student_sessions (id, student_id, expires_at) VALUES ($1,$2,$3) RETURNING *;
-- name: GetStudentSession :one
SELECT * FROM student_sessions WHERE id=$1 AND expires_at > now();
-- name: DeleteStudentSession :exec
DELETE FROM student_sessions WHERE id=$1;
