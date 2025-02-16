package field

import (
	"encoding/json"
	"fmt"

	"github.com/graphql-go/graphql"
	"github.com/tidwall/gjson"
)

// createListField function is responsible for creating a GraphQL field that represents an array (graphql.List).
// Its main task is to correctly determine the type of array elements, even if the JSON contains different data structures within the same list.
func (f *DefaultFieldFactory) createListField(key string, arr []interface{}, depth int) *graphql.Field {
	if len(arr) == 0 {
		return &graphql.Field{Type: graphql.NewList(graphql.String)}
	}

	// If every element in the array is a map (array of json objects), we run a special case.
	// In this case we merge all maps and create a field for the merged map.
	if ok, arrMaps := allMaps(arr); ok {
		return f.createListFieldForMaps(key, arrMaps, depth)
	}

	// Create a field for the first element in the array.
	// For code simplicity we assume that all elements in the array have the same type.
	elementField := f.CreateField(key, arr[0], depth+1)

	return &graphql.Field{
		Type: graphql.NewList(elementField.Type),
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			var key string
			for i, k := range p.Info.Path.AsArray() {
				if i > 0 {
					key += "."
				}
				key += fmt.Sprintf("%v", k)
			}

			data := gjson.Get(string(f.jsonData), key)

			result := []interface{}{}
			json.Unmarshal([]byte(data.Raw), &result)

			// fmt.Println("list resolved", key, result)

			return result, nil
		},
	}
}

// The createListFieldForMaps function is responsible for processing lists of objects (map[string]interface{}) in JSON and generating a GraphQL List field with the correct schema.
// In simple terms:
//   - It analyzes all objects within the array.
//   - Identifies all possible fields that may appear.
//   - Creates a unified GraphQL type that includes the merged schema of all possible fields.
//   - Returns a GraphQL list of objects (graphql.List).
func (f *DefaultFieldFactory) createListFieldForMaps(key string, arrObjects []map[string]interface{}, depth int) *graphql.Field {
	mergedDefaults := mergeMaps(arrObjects)
	mergedField := f.CreateField(key, mergedDefaults, depth+1)
	listType := graphql.NewList(mergedField.Type)

	return &graphql.Field{
		Type: listType,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			var key string
			for i, k := range p.Info.Path.AsArray() {
				if i > 0 {
					key += "."
				}
				key += fmt.Sprintf("%v", k)
			}

			data := gjson.Get(string(f.jsonData), key)

			result := []interface{}{}
			json.Unmarshal([]byte(data.Raw), &result)

			return result, nil
		},
	}
}
