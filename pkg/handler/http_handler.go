package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/graphql-go/graphql"
)

type SchemaBuilder interface {
	BuildSchema(jsonData map[string]interface{}) (*graphql.Schema, error)
}

type GraphQLHandler struct {
	schemaBuilder SchemaBuilder
	schema        *graphql.Schema
}

func NewGraphQLHandler(
	schemaBuilder SchemaBuilder,
) *GraphQLHandler {
	return &GraphQLHandler{
		schemaBuilder: schemaBuilder,
	}
}

func (h *GraphQLHandler) UpdateSchema(schema *graphql.Schema) {
	h.schema = schema
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
