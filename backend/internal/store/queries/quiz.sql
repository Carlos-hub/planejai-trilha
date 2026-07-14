-- name: CreateQuiz :one
INSERT INTO quizzes (study_trail_id) VALUES ($1)
ON CONFLICT (study_trail_id) DO UPDATE SET study_trail_id=EXCLUDED.study_trail_id
RETURNING *;
-- name: GetQuizByTrail :one
SELECT * FROM quizzes WHERE study_trail_id=$1;
-- name: DeleteQuestionsByQuiz :exec
DELETE FROM quiz_questions WHERE quiz_id=$1;
-- name: InsertQuestion :exec
INSERT INTO quiz_questions (quiz_id, ordem, enunciado, opcoes, correta) VALUES ($1,$2,$3,$4,$5);
-- name: ListQuestions :many
SELECT * FROM quiz_questions WHERE quiz_id=$1 ORDER BY ordem;
-- name: ListQuestionsPublic :many
SELECT id, ordem, enunciado, opcoes FROM quiz_questions WHERE quiz_id=$1 ORDER BY ordem;
