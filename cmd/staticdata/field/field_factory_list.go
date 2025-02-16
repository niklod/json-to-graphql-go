package field

import (
	"encoding/json"
	"fmt"

	"github.com/graphql-go/graphql"
	"github.com/tidwall/gjson"
)

// createListField creates a list field for a JSON array.
// For arrays of objects, it merges union defaults and uses a custom resolver.
func (f *DefaultFieldFactory) createListField(key string, arr []interface{}, depth int) *graphql.Field {
	if len(arr) == 0 {
		return &graphql.Field{Type: graphql.NewList(graphql.String)}
	}

	if ok, arrMaps := allMaps(arr); ok {
		return f.createListFieldForMaps(key, arrMaps, depth)
	}

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

			fmt.Println("list resolved", key, result)

			return result, nil
		},
	}
}

// createListFieldForMaps creates a list field for an array of maps.
func (f *DefaultFieldFactory) createListFieldForMaps(key string, arrMaps []interface{}, depth int) *graphql.Field {
	var mapsSlice []map[string]interface{}
	for _, elem := range arrMaps {
		mapsSlice = append(mapsSlice, elem.(map[string]interface{}))
	}

	mergedDefaults := mergeMaps(mapsSlice)
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
