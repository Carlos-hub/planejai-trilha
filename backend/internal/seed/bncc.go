package seed

import (
	"context"
	"encoding/json"
	"os"

	"github.com/Carlos-hub/planejai/backend/internal/store"
)

type skill struct {
	Code       string  `json:"code"`
	Etapa      string  `json:"etapa"`
	Disciplina string  `json:"disciplina"`
	Anos       []int32 `json:"anos"`
	Assunto    string  `json:"assunto"`
	Descricao  string  `json:"descricao"`
}

// BNCC loads the static BNCC catalog. InsertBnccSkill upserts by code, so
// this is safe to run on every boot and keeps an existing database in sync
// with the seed file as the catalog grows.
func BNCC(ctx context.Context, q *store.Queries, path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var items []skill
	if err := json.Unmarshal(b, &items); err != nil {
		return err
	}
	for _, it := range items {
		if err := q.InsertBnccSkill(ctx, store.InsertBnccSkillParams{
			Code:       it.Code,
			Etapa:      it.Etapa,
			Disciplina: it.Disciplina,
			Anos:       it.Anos,
			Assunto:    it.Assunto,
			Descricao:  it.Descricao,
		}); err != nil {
			return err
		}
	}
	return nil
}
