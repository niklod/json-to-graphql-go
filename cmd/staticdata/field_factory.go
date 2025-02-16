package main

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/graphql-go/graphql"
	"github.com/tidwall/gjson"
)

// DefaultFieldFactory is the default implementation.
type DefaultFieldFactory struct {
	unionInfo map[string]map[string]bool // bare key -> subfield -> isObject
	typeCache map[string]*graphql.Object // type name -> GraphQL object
	jsonData  []byte
}

// NewDefaultFieldFactory creates a new DefaultFieldFactory.
func NewDefaultFieldFactory(data []byte) FieldFactory {
	return &DefaultFieldFactory{
		unionInfo: make(map[string]map[string]bool),
		typeCache: make(map[string]*graphql.Object),
		jsonData:  data,
	}
}

// GatherUnionInfo recursively scans the JSON and updates union info using bare keys.
func (f *DefaultFieldFactory) GatherUnionInfo(data interface{}) {
	switch d := data.(type) {
	case map[string]interface{}:
		for key, value := range d {
			// If value is an object, record its subkeys under the bare key.
			if m, ok := value.(map[string]interface{}); ok {
				if f.unionInfo[key] == nil {
					f.unionInfo[key] = make(map[string]bool)
				}
				for subKey, subVal := range m {
					_, isObj := subVal.(map[string]interface{})
					if isObj || !f.unionInfo[key][subKey] {
						f.unionInfo[key][subKey] = isObj
					}
				}
			}
			// Recurse into the value.
			f.GatherUnionInfo(value)
		}
	case []interface{}:
		for _, item := range d {
			f.GatherUnionInfo(item)
		}
	}
}

// CreateField dispatches field creation based on the type of JSON value.
func (f *DefaultFieldFactory) CreateField(key string, value interface{}, depth int) *graphql.Field {
	if depth > 10 {
		return &graphql.Field{Type: graphql.String}
	}

	switch v := value.(type) {
	case string, float64, bool, nil:
		return f.createScalarField(v)
	case map[string]interface{}:
		return f.createObjectField(key, v, depth)
	case []interface{}:
		return f.createListField(key, v, depth)
	default:
		return &graphql.Field{Type: graphql.String}
	}
}

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
	for _, k := range keys {
		var subVal interface{}
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

// mergeKeys merges keys from the current object and union info.
func (f *DefaultFieldFactory) mergeKeys(key string, m map[string]interface{}) []string {
	keysSet := make(map[string]bool)
	for k := range m {
		keysSet[k] = true
	}
	if info, exists := f.unionInfo[key]; exists {
		for subKey := range info {
			keysSet[subKey] = true
		}
	}
	var keys []string
	for k := range keysSet {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

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

func (f *DefaultFieldFactory) ResetCache() {
	f.typeCache = make(map[string]*graphql.Object)
	f.unionInfo = make(map[string]map[string]bool)
}

// mergeOverride recursively merges two maps; keys in 'over' take precedence.
func mergeOverride(def, over map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	for k, v := range def {
		res[k] = v
	}
	for k, ov := range over {
		if d, exists := res[k]; exists {
			if dMap, ok1 := d.(map[string]interface{}); ok1 {
				if oMap, ok2 := ov.(map[string]interface{}); ok2 {
					res[k] = mergeOverride(dMap, oMap)
					continue
				}
			}
		}
		res[k] = ov
	}
	return res
}

// mergeMaps returns the union of a slice of maps recursively.
func mergeMaps(maps []map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	for _, m := range maps {
		for k, v := range m {
			if existing, ok := merged[k]; ok {
				if exMap, ok1 := existing.(map[string]interface{}); ok1 {
					if newMap, ok2 := v.(map[string]interface{}); ok2 {
						merged[k] = mergeMaps([]map[string]interface{}{exMap, newMap})
						continue
					}
				}
			} else {
				merged[k] = v
			}
		}
	}
	return merged
}

// allMaps returns true if every element in the array is a map.
func allMaps(arr []interface{}) (bool, []interface{}) {
	for _, e := range arr {
		if _, ok := e.(map[string]interface{}); !ok {
			return false, nil
		}
	}
	return true, arr
}
