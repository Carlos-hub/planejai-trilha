package store

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func aiTokPool(t *testing.T) *Queries {
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

func TestUpsertAIToken(t *testing.T) {
	q := aiTokPool(t)
	ctx := context.Background()
	u, err := q.CreateUser(ctx, CreateUserParams{Email: "ai-tok-store@t.com", SenhaHash: "x", Nome: "P"})
	if err != nil {
		t.Fatal(err)
	}
	first, err := q.UpsertAIToken(ctx, UpsertAITokenParams{UserID: u.ID, Provider: "anthropic", TokenCiphertext: []byte{1, 2, 3}, TokenNonce: []byte{9}})
	if err != nil {
		t.Fatal(err)
	}
	if first.Provider != "anthropic" {
		t.Fatalf("provider = %q", first.Provider)
	}
	// second upsert replaces provider + ciphertext, still one row
	second, err := q.UpsertAIToken(ctx, UpsertAITokenParams{UserID: u.ID, Provider: "openai", TokenCiphertext: []byte{4, 5}, TokenNonce: []byte{8}})
	if err != nil {
		t.Fatal(err)
	}
	if second.ID != first.ID || second.Provider != "openai" {
		t.Fatalf("upsert did not update in place: %+v", second)
	}
	got, err := q.GetAIToken(ctx, u.ID)
	if err != nil || got.Provider != "openai" {
		t.Fatalf("get = %+v err=%v", got, err)
	}
	if err := q.DeleteAIToken(ctx, u.ID); err != nil {
		t.Fatal(err)
	}
}
