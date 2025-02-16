package builder

import (
	"github.com/graphql-go/graphql"
)

// GraphQLSchemaBuilder implements SchemaBuilder.
type GraphQLSchemaBuilder struct {
	fieldFactory FieldFactory
}

// NewGraphQLSchemaBuilder creates a new GraphQLSchemaBuilder.
func NewGraphQLSchemaBuilder(factory FieldFactory) *GraphQLSchemaBuilder {
	return &GraphQLSchemaBuilder{
		fieldFactory: factory,
	}
}

// BuildSchema parses JSON, gathers union info, and builds the GraphQL schema.
func (b *GraphQLSchemaBuilder) BuildSchema(jsonData map[string]interface{}) (*graphql.Schema, error) {
	// Reset internal factory caches
	b.fieldFactory.ResetCache()

	b.fieldFactory.GatherUnionInfo(jsonData)

	fields := graphql.Fields{}
	for key, value := range jsonData {
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

	return &schema, err
}
