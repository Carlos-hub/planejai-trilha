package main

import (
	"log"
	"net/http"
	"os"

	apihttp "github.com/Carlos-hub/planejai/backend/internal/http"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("listening on :%s", port)
	if err := http.ListenAndServe(":"+port, apihttp.NewRouter(apihttp.Deps{})); err != nil {
		log.Fatal(err)
	}
}
