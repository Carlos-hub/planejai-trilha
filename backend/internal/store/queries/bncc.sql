-- name: CountBnccSkills :one
SELECT count(*) FROM bncc_skills;
-- name: InsertBnccSkill :exec
INSERT INTO bncc_skills (code, disciplina, ano, descricao) VALUES ($1,$2,$3,$4)
ON CONFLICT (code) DO NOTHING;
-- name: ListBnccSkills :many
SELECT * FROM bncc_skills
WHERE ($1::text = '' OR disciplina = $1) AND ($2::text = '' OR ano = $2)
ORDER BY code;
-- name: GetBnccSkill :one
SELECT * FROM bncc_skills WHERE id = $1;
