package main

import (
	"encoding/json"

	"github.com/graphql-go/graphql"
)

// FieldFactory creates GraphQL fields from JSON values.
type FieldFactory interface {
	// CreateField returns a GraphQL field for the given key and JSON value.
	CreateField(key string, value interface{}, depth int) *graphql.Field
	// GatherUnionInfo scans JSON data and records union metadata.
	GatherUnionInfo(data interface{})
	ResetCache()
}

// GraphQLSchemaBuilder implements SchemaBuilder.
type GraphQLSchemaBuilder struct {
	fieldFactory FieldFactory
}

// NewGraphQLSchemaBuilder creates a new GraphQLSchemaBuilder.
func NewGraphQLSchemaBuilder() SchemaBuilder {
	return &GraphQLSchemaBuilder{}
}

// BuildSchema parses JSON, gathers union info, and builds the GraphQL schema.
func (b *GraphQLSchemaBuilder) BuildSchema(jsonData []byte) (*graphql.Schema, map[string]interface{}, error) {
	// Update factory's with new json
	b.fieldFactory = NewDefaultFieldFactory(jsonData)

	// Reset internal caches for each update
	b.fieldFactory.ResetCache()

	// Extract data from JSON
	var data map[string]interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, nil, err
	}

	b.fieldFactory.GatherUnionInfo(data)

	fields := graphql.Fields{}
	for key, value := range data {
		fields[key] = b.fieldFactory.CreateField(key, value, 0)
	}

	// Ensure at least one field exists.
	if len(fields) == 0 {
		fields["anyField"] = &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return nil, nil
			},
		}
	}

	// Create root query object
	rootQuery := graphql.NewObject(graphql.ObjectConfig{
		Name:   "RootQuery",
		Fields: fields,
	})

	// Create GraphQL schema
	schema, err := graphql.NewSchema(graphql.SchemaConfig{Query: rootQuery})
	return &schema, data, err
}
