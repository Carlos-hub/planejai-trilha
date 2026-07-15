-- +goose Up
-- Move to the official BNCC catalog shape: a skill carries an etapa (EF/EM) and
-- spans one or more school years (EF67 = anos 6 and 7), so ano (single text)
-- becomes anos (integer array).
ALTER TABLE bncc_skills ADD COLUMN etapa TEXT NOT NULL DEFAULT '';
ALTER TABLE bncc_skills ADD COLUMN anos INTEGER[] NOT NULL DEFAULT '{}';
DROP INDEX IF EXISTS idx_bncc_skills_ano;
ALTER TABLE bncc_skills DROP COLUMN ano;
CREATE INDEX idx_bncc_skills_etapa ON bncc_skills (etapa);
CREATE INDEX idx_bncc_skills_anos ON bncc_skills USING GIN (anos);

-- +goose Down
DROP INDEX IF EXISTS idx_bncc_skills_anos;
DROP INDEX IF EXISTS idx_bncc_skills_etapa;
ALTER TABLE bncc_skills ADD COLUMN ano TEXT NOT NULL DEFAULT '';
ALTER TABLE bncc_skills DROP COLUMN anos;
ALTER TABLE bncc_skills DROP COLUMN etapa;
CREATE INDEX idx_bncc_skills_ano ON bncc_skills (ano);
