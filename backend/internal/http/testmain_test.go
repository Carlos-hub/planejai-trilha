package http

import (
	"context"
	"os"
	"testing"

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
