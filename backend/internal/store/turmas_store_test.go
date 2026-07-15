package store

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func storeTestPool(t *testing.T) *Queries {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return New(pool)
}

func TestCreateTurmaAndStudent(t *testing.T) {
	q := storeTestPool(t)
	ctx := context.Background()
	u, err := q.CreateUser(ctx, CreateUserParams{Email: "prof-turma@test.com", SenhaHash: "x", Nome: "Prof"})
	if err != nil {
		t.Fatal(err)
	}
	turma, err := q.CreateTurma(ctx, CreateTurmaParams{UserID: u.ID, Nome: "6A", Etapa: "EF", Anos: []int32{6}})
	if err != nil {
		t.Fatal(err)
	}
	if turma.Nome != "6A" {
		t.Fatalf("nome = %q", turma.Nome)
	}
	s, err := q.CreateStudent(ctx, CreateStudentParams{TurmaID: turma.ID, Nome: "Ana", Usuario: "ana.test.9z", SenhaHash: "h"})
	if err != nil {
		t.Fatal(err)
	}
	if s.TurmaID != turma.ID {
		t.Fatalf("turma_id = %d", s.TurmaID)
	}
}
