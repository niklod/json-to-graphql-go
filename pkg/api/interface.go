package api

import "github.com/graphql-go/graphql"

type Resolver interface {
	UpdateJsonData(jsonData []byte)
	ResolveScalarValue(p graphql.ResolveParams) (interface{}, error)
	ResolveObjectValue(p graphql.ResolveParams) (interface{}, error)
	ResolveArrayValue(p graphql.ResolveParams) (interface{}, error)
}

type JsonProvider interface {
	GetJsonData() (map[string]interface{}, error)
	GetRawJson() ([]byte, error)
}

type SchemaBuilder interface {
	BuildSchema(jsonData map[string]interface{}) (*graphql.Schema, error)
}
