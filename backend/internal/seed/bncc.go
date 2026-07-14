package seed

import (
	"context"
	"encoding/json"
	"os"

	"github.com/Carlos-hub/planejai/backend/internal/store"
)

type skill struct {
	Code       string `json:"code"`
	Disciplina string `json:"disciplina"`
	Ano        string `json:"ano"`
	Descricao  string `json:"descricao"`
}

func BNCC(ctx context.Context, q *store.Queries, path string) error {
	n, err := q.CountBnccSkills(ctx)
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
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
			Disciplina: it.Disciplina,
			Ano:        it.Ano,
			Descricao:  it.Descricao,
		}); err != nil {
			return err
		}
	}
	return nil
}
