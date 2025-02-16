package field

import "github.com/graphql-go/graphql"

type Resolver interface {
	UpdateJsonData(jsonData []byte)
	ResolveScalarValue(p graphql.ResolveParams) (interface{}, error)
	ResolveObjectValue(p graphql.ResolveParams) (interface{}, error)
	ResolveArrayValue(p graphql.ResolveParams) (interface{}, error)
}
