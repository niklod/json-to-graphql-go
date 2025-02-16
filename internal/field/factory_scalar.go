package field

import (
	"github.com/graphql-go/graphql"
)

// createScalarField returns a field for scalar values.
func (f *DefaultFieldFactory) createScalarField(value interface{}) *graphql.Field {
	var t graphql.Output
	switch value.(type) {
	case string:
		t = graphql.String
	case float64:
		t = graphql.Float
	case bool:
		t = graphql.Boolean
	default:
		t = graphql.String
	}

	return &graphql.Field{
		Type:    t,
		Resolve: f.resolver.ResolveScalarValue,
	}
}
