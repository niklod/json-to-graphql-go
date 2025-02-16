package builder

import "github.com/graphql-go/graphql"

// FieldFactory creates GraphQL fields from JSON values.
type FieldFactory interface {
	// CreateField returns a GraphQL field for the given key and JSON value.
	CreateField(key string, value interface{}, depth int) *graphql.Field
	// GatherUnionInfo scans JSON data and records union metadata.
	GatherUnionInfo(data interface{})
	ResetCache()
}

type JsonProvider interface {
	GetJsonData() (map[string]interface{}, error)
}
