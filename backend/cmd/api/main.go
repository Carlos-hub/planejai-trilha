package main

import (
	"context"
	"log"
	"net/http"
	"os"

	apihttp "github.com/Carlos-hub/planejai/backend/internal/http"
	"github.com/Carlos-hub/planejai/backend/internal/store"
)

func main() {
	ctx := context.Background()
	pool, err := store.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	deps := apihttp.Deps{Store: store.New(pool), Pool: pool}
	log.Printf("listening on :%s", port)
	if err := http.ListenAndServe(":"+port, apihttp.NewRouter(deps)); err != nil {
		log.Fatal(err)
	}
}
