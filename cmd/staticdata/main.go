package main

import (
	"log"
	"net/http"
)

func main() {
	schemaBuilder := NewGraphQLSchemaBuilder()
	httpHandler := NewGraphQLHandler(schemaBuilder)
	if err := httpHandler.RefreshSchema(); err != nil {
		log.Fatalf("failed to refresh schema, error: %v", err)
	}

	httpHandler.StartAutoRefresh()

	http.Handle("/graphql", httpHandler)
	log.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
