-- name: CountBnccSkills :one
SELECT count(*) FROM bncc_skills;
-- name: InsertBnccSkill :exec
INSERT INTO bncc_skills (code, disciplina, ano, assunto, descricao)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (code) DO UPDATE SET
  disciplina = EXCLUDED.disciplina,
  ano = EXCLUDED.ano,
  assunto = EXCLUDED.assunto,
  descricao = EXCLUDED.descricao;
-- name: ListBnccSkills :many
SELECT * FROM bncc_skills
WHERE ($1::text = '' OR disciplina = $1)
  AND ($2::text = '' OR ano = $2)
  AND ($3::text = '' OR assunto = $3)
ORDER BY disciplina, ano, code;
-- name: GetBnccSkill :one
SELECT * FROM bncc_skills WHERE id = $1;
