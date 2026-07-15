package seed

import (
	"context"
	"encoding/json"
	"os"

	"github.com/Carlos-hub/planejai/backend/internal/auth"
	"github.com/Carlos-hub/planejai/backend/internal/store"
)

type skill struct {
	Code       string `json:"code"`
	Disciplina string `json:"disciplina"`
	Ano        string `json:"ano"`
	Assunto    string `json:"assunto"`
	Descricao  string `json:"descricao"`
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
			Disciplina: it.Disciplina,
			Ano:        it.Ano,
			Assunto:    it.Assunto,
			Descricao:  it.Descricao,
		}); err != nil {
			return err
		}
	}
	return nil
}

func DemoTeacher(ctx context.Context, q *store.Queries) error {
	_, err := q.GetUserByEmail(ctx, "prof@demo.com")
	if err == nil {
		return nil
	}
	senhaHash, err := auth.HashPassword("demo1234")
	if err != nil {
		return err
	}
	_, err = q.CreateUser(ctx, store.CreateUserParams{
		Email:     "prof@demo.com",
		SenhaHash: senhaHash,
		Nome:      "Professor Demo",
	})
	return err
}
