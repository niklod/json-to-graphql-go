package field

import (
	"fmt"

	"github.com/graphql-go/graphql"
	"github.com/tidwall/gjson"
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
		Type: t,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			var key string
			for i, k := range p.Info.Path.AsArray() {
				if i > 0 {
					key += "."
				}
				key += fmt.Sprintf("%v", k)
			}

			data := gjson.Get(string(f.jsonData), key)

			result := data.Value()

			fmt.Println("scalar resolved", key, result)

			return result, nil
		},
	}
}
