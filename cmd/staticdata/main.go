package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/niklod/json-to-graphql-go/pkg/api"
)

func main() {
	ctx := context.Background()
	app, err := api.New(api.Config{})
	if err != nil {
		log.Fatalf("failed to create app, error: %v", err)
	}

	app.StartBackgroundSchemaUpdate(ctx, time.Second*5)

	http.Handle("/graphql", app.Handler)
	log.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
