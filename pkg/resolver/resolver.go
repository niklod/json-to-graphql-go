package resolver

import (
	"fmt"
	"strings"

	"github.com/graphql-go/graphql"
	"github.com/tidwall/gjson"
)

type JSONResolver struct {
	jsonData []byte
}

func NewJSONResolver(jsonData []byte) *JSONResolver {
	return &JSONResolver{jsonData: jsonData}
}

func (r *JSONResolver) UpdateJsonData(jsonData []byte) {
	r.jsonData = jsonData
}

func (r *JSONResolver) ResolveScalarValue(p graphql.ResolveParams) (interface{}, error) {
	var key string
	for i, k := range p.Info.Path.AsArray() {
		if i > 0 {
			key += "."
		}
		key += fmt.Sprintf("%v", k)
	}

	data := gjson.Get(string(r.jsonData), key)

	result := data.Value()

	// fmt.Println("scalar resolved", key, result)

	return result, nil
}

func (r *JSONResolver) ResolveObjectValue(p graphql.ResolveParams) (interface{}, error) {
	var keyBuilder strings.Builder
	for i, k := range p.Info.Path.AsArray() {
		if i > 0 {
			keyBuilder.WriteString(".")
		}
		keyBuilder.WriteString(fmt.Sprintf("%v", k))
	}
	key := keyBuilder.String()

	data := gjson.Get(string(r.jsonData), key)

	return data.Map(), nil
}

func (r *JSONResolver) ResolveArrayValue(p graphql.ResolveParams) (interface{}, error) {
	var key string
	for i, k := range p.Info.Path.AsArray() {
		if i > 0 {
			key += "."
		}
		key += fmt.Sprintf("%v", k)
	}

	data := gjson.Get(string(r.jsonData), key)

	// fmt.Println("list resolved", key, data.Array())

	return data.Array(), nil
}
