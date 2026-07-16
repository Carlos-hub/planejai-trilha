package http

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"os"
	"testing"

	"github.com/Carlos-hub/planejai/backend/internal/lesson"
	"github.com/Carlos-hub/planejai/backend/internal/secret"
	"github.com/Carlos-hub/planejai/backend/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

func testPool(t *testing.T) *pgxpool.Pool {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func testDeps(t *testing.T) Deps {
	pool := testPool(t)
	return Deps{Store: store.New(pool), Pool: pool, SessionSecret: "test-secret"}
}

// withAI configures d with a secret box + a mock NewGen and seeds an encrypted
// token for userID, so generatorForUser resolves to a MockGenerator. It returns
// a pointer that captures the provider requested from NewGen.
func withAI(t *testing.T, d *Deps, userID int64, provider string) *string {
	t.Helper()
	k := make([]byte, 32)
	rand.Read(k)
	box, err := secret.NewBox(base64.StdEncoding.EncodeToString(k))
	if err != nil {
		t.Fatal(err)
	}
	d.Secret = box
	captured := new(string)
	d.NewGen = func(ctx context.Context, prov, key string) (lesson.Generator, error) {
		*captured = prov
		return &lesson.MockGenerator{}, nil
	}
	ct, nonce, err := box.Seal([]byte("fake-token"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.Store.UpsertAIToken(context.Background(), store.UpsertAITokenParams{
		UserID: userID, Provider: provider, TokenCiphertext: ct, TokenNonce: nonce,
	}); err != nil {
		t.Fatal(err)
	}
	return captured
}
