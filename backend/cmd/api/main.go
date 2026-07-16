package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"

	apihttp "github.com/Carlos-hub/planejai/backend/internal/http"
	"github.com/Carlos-hub/planejai/backend/internal/lesson"
	"github.com/Carlos-hub/planejai/backend/internal/secret"
	"github.com/Carlos-hub/planejai/backend/internal/seed"
	"github.com/Carlos-hub/planejai/backend/internal/store"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func main() {
	ctx := context.Background()

	dbURL := os.Getenv("DATABASE_URL")
	if err := migrate(dbURL); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	pool, err := store.Connect(ctx, dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	deps := apihttp.Deps{Store: store.New(pool), Pool: pool}
	if box, err := secret.NewBox(os.Getenv("TOKEN_ENC_KEY")); err != nil {
		log.Printf("token encryption unavailable: %v", err)
	} else {
		deps.Secret = box
	}
	deps.NewGen = lesson.NewGeneratorForProvider
	if err := seed.BNCC(ctx, deps.Store, "seed/bncc.json"); err != nil {
		log.Printf("seed: %v", err)
	}
	if err := seed.DemoTeacher(ctx, deps.Store); err != nil {
		log.Printf("seed: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("listening on :%s", port)
	if err := http.ListenAndServe(":"+port, apihttp.NewRouter(deps)); err != nil {
		log.Fatal(err)
	}
}

func migrate(dbURL string) error {
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return err
	}
	defer db.Close()

	dir := os.Getenv("MIGRATIONS_DIR")
	if dir == "" {
		dir = "/migrations"
	}

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return goose.Up(db, dir)
}
