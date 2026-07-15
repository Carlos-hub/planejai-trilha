-- name: CountBnccSkills :one
SELECT count(*) FROM bncc_skills;
-- name: InsertBnccSkill :exec
INSERT INTO bncc_skills (code, etapa, disciplina, anos, assunto, descricao)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (code) DO UPDATE SET
  etapa = EXCLUDED.etapa,
  disciplina = EXCLUDED.disciplina,
  anos = EXCLUDED.anos,
  assunto = EXCLUDED.assunto,
  descricao = EXCLUDED.descricao;
-- name: ListBnccSkills :many
SELECT * FROM bncc_skills
WHERE (sqlc.arg(etapa)::text = '' OR etapa = sqlc.arg(etapa))
  AND (sqlc.arg(disciplina)::text = '' OR disciplina = sqlc.arg(disciplina))
  AND (
    sqlc.arg(q)::text = ''
    OR code ILIKE '%' || sqlc.arg(q) || '%'
    OR disciplina ILIKE '%' || sqlc.arg(q) || '%'
    OR assunto ILIKE '%' || sqlc.arg(q) || '%'
    OR descricao ILIKE '%' || sqlc.arg(q) || '%'
  )
ORDER BY etapa, disciplina, code;
-- name: GetBnccSkill :one
SELECT * FROM bncc_skills WHERE id = $1;
