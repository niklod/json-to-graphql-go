package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/graphql-go/graphql"
)

type GraphQLHandler struct {
	schemaBuilder SchemaBuilder
	schema        *graphql.Schema
	staticData    *map[string]interface{}
}

func NewGraphQLHandler(
	schemaBuilder SchemaBuilder,
) *GraphQLHandler {
	return &GraphQLHandler{
		schemaBuilder: schemaBuilder,
	}
}

func (h *GraphQLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Query string `json:"query"`
	}

	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, "failed to decode request body", http.StatusBadRequest)

		return
	}

	if h.schema == nil {
		http.Error(w, "schema is not configured", http.StatusInternalServerError)

		return
	}

	result := graphql.Do(graphql.Params{
		Schema:        *h.schema,
		RequestString: params.Query,
		RootObject:    *h.staticData,
	})

	if len(result.Errors) > 0 {
		log.Printf("failed to execute graphql operation, errors: %+v", result.Errors)
		http.Error(w, "failed to execute graphql operation", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *GraphQLHandler) RefreshSchema() error {
	data, err := os.ReadFile("./cmd/staticdata/data.json")
	if err != nil {
		return fmt.Errorf("failed to read data.json: %w", err)
	}

	schema, staticData, err := h.schemaBuilder.BuildSchema(data)
	if err != nil {
		return err
	}

	h.schema = schema
	h.staticData = &staticData
	log.Println("GraphQL schema successfully refreshed")

	return nil
}

func (h *GraphQLHandler) StartAutoRefresh() {
	go func() {
		for {
			if err := h.RefreshSchema(); err != nil {
				log.Printf("failed to refresh schema: %v", err)
			}
			time.Sleep(5 * time.Second)
		}
	}()
}
