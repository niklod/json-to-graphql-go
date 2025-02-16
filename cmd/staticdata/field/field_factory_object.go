package field

import (
	"encoding/json"
	"fmt"

	"github.com/graphql-go/graphql"
	"github.com/tidwall/gjson"
)

// createObjectField is responsible for dynamically constructing a GraphQL object field from a given JSON structure.
// What it does:
//   - Checks if the GraphQL type is already cached (to avoid redundant type creation).
//   - Collects all possible keys for the object (including inconsistencies).
//   - Iterates over all discovered keys to create corresponding GraphQL fields.
//   - Stores the generated type in the cache for future use.
//   - Defines a Resolve function that dynamically fetches JSON data at runtime.
func (f *DefaultFieldFactory) createObjectField(key string, inputObj map[string]interface{}, depth int) *graphql.Field {
	typeName := key + "Object"
	if cached, ok := f.gqlTypesCache.get(typeName); ok {
		return &graphql.Field{Type: cached}
	}

	// collecting all possible field names for the object, including:
	//  - Keys from inputObj.
	//  - Keys from f.unionInfo, which contains inconsistencies observed in previous objects.
	mergedKeySet := f.mergeKeys(key, inputObj)

	fields := graphql.Fields{}
	for _, key := range mergedKeySet {
		var subVal interface{}

		// If the key is present in the current object, use it.
		if val, ok := inputObj[key]; ok {
			subVal = val
		} else if info, exists := f.unionInfo.get(key); exists && info[key] {
			// Force missing subfields to be objects.
			subVal = map[string]interface{}{}
		} else {
			subVal = nil
		}

		fields[key] = f.CreateField(key, subVal, depth+1)
	}

	objType := graphql.NewObject(graphql.ObjectConfig{
		Name:   typeName,
		Fields: fields,
	},
	)

	f.gqlTypesCache.set(typeName, objType)

	return &graphql.Field{
		Type: objType,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			var key string
			for i, k := range p.Info.Path.AsArray() {
				if i > 0 {
					key += "."
				}
				key += fmt.Sprintf("%v", k)
			}

			data := gjson.Get(string(f.jsonData), key)

			result := map[string]interface{}{}
			json.Unmarshal([]byte(data.Raw), &result)

			// fmt.Println("object resolved", key, result)

			return result, nil
		},
	}
}
