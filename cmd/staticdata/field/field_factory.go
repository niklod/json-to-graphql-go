package field

import (
	"sort"

	"github.com/graphql-go/graphql"
)

// DefaultFieldFactory is the default implementation.
type DefaultFieldFactory struct {
	unionInfo map[string]map[string]bool // bare key -> subfield -> isObject
	typeCache map[string]*graphql.Object // type name -> GraphQL object
	jsonData  []byte
}

// NewDefaultFieldFactory creates a new DefaultFieldFactory.
func NewDefaultFieldFactory(data []byte) *DefaultFieldFactory {
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
