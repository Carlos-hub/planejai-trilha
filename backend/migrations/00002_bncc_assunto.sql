-- +goose Up
ALTER TABLE bncc_skills ADD COLUMN assunto TEXT NOT NULL DEFAULT '';
CREATE INDEX idx_bncc_skills_ano ON bncc_skills (ano);
CREATE INDEX idx_bncc_skills_disciplina ON bncc_skills (disciplina);

-- +goose Down
DROP INDEX IF EXISTS idx_bncc_skills_disciplina;
DROP INDEX IF EXISTS idx_bncc_skills_ano;
ALTER TABLE bncc_skills DROP COLUMN assunto;
