package main

import (
	"context"
	"log"
	"net/http"
	"os"

	apihttp "github.com/Carlos-hub/planejai/backend/internal/http"
	"github.com/Carlos-hub/planejai/backend/internal/lesson"
	"github.com/Carlos-hub/planejai/backend/internal/seed"
	"github.com/Carlos-hub/planejai/backend/internal/store"
)

func main() {
	ctx := context.Background()
	pool, err := store.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	deps := apihttp.Deps{Store: store.New(pool), Pool: pool}
	gen, err := lesson.NewLangChainGenerator()
	if err != nil {
		log.Printf("lesson generator unavailable: %v", err)
	}
	deps.Gen = gen
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
