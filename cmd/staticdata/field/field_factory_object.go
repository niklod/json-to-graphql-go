package field

import (
	"encoding/json"
	"fmt"

	"github.com/graphql-go/graphql"
	"github.com/tidwall/gjson"
)

// createObjectField creates an object field for a JSON object,
// merging keys from the current object with union info (using bare key).
func (f *DefaultFieldFactory) createObjectField(key string, m map[string]interface{}, depth int) *graphql.Field {
	typeName := key + "Object"
	if cached, ok := f.typeCache[typeName]; ok {
		return &graphql.Field{Type: cached}
	}

	// Build union of keys: from the current object and unionInfo.
	keys := f.mergeKeys(key, m)
	fields := graphql.Fields{}

	// Iterate over keys to create fields.
	for _, k := range keys {

		var subVal interface{}

		// If the key is present in the current object, use it.
		if val, ok := m[k]; ok {
			subVal = val
		} else if info, exists := f.unionInfo[key]; exists && info[k] {
			// Force missing subfields to be objects.
			subVal = map[string]interface{}{}
		} else {
			subVal = nil
		}

		fields[k] = f.CreateField(k, subVal, depth+1)
	}

	objType := graphql.NewObject(graphql.ObjectConfig{
		Name:   typeName,
		Fields: fields,
	},
	)

	f.typeCache[typeName] = objType

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

			fmt.Println("object resolved", key, result)

			return result, nil
		},
	}
}
