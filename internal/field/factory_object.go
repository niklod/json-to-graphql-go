package field

import (
	"github.com/graphql-go/graphql"
)

// createObjectField is responsible for dynamically constructing a GraphQL object field from a given JSON structure.
// What it does:
//   - Checks if the GraphQL type is already cached (to avoid redundant type creation).
//   - Collects all possible keys for the object (including inconsistencies).
//   - Iterates over all discovered keys to create corresponding GraphQL fields.
//   - Stores the generated type in the cache for future use.
//   - Defines a Resolve function that dynamically fetches JSON data at runtime.
func (f *DefaultFieldFactory) createObjectField(key string, m map[string]interface{}, depth int) *graphql.Field {
	typeName := f.objectNameFn(key)
	if cached, ok := f.gqlTypesCache.get(typeName); ok {
		return &graphql.Field{Type: cached}
	}

	// Build fields separately
	fields := f.buildGraphQLFields(key, m, depth)

	objType := graphql.NewObject(graphql.ObjectConfig{
		Name:   typeName,
		Fields: fields,
	})

	f.gqlTypesCache.set(typeName, objType)

	return &graphql.Field{
		Type:    objType,
		Resolve: f.resolver.ResolveObjectValue,
	}
}

// buildGraphQLFields generates GraphQL fields for a given JSON object.
// It processes each key in the object, determines its type, and creates the corresponding GraphQL field.
//
// This function is responsible for handling inconsistencies in JSON structures by:
// - Merging all possible keys from the given object (`inputObj`) and `unionInfo`.
// - Ensuring that missing fields are either set to `nil` (to prevent schema mismatches) or processed correctly.
// - Recursively creating fields for nested structures.
func (f *DefaultFieldFactory) buildGraphQLFields(key string, inputObj map[string]interface{}, depth int) graphql.Fields {
	keys := f.mergeKeys(key, inputObj)
	fields := graphql.Fields{}

	for _, k := range keys {
		var subVal interface{}

		// Assign value if key exists in the current object
		if val, ok := inputObj[k]; ok {
			subVal = val
		}

		// If key is in unionInfo but missing in inputObj, set subVal to nil
		if subVal == nil && f.unionInfo.subkeyExists(key, k) {
			subVal = nil
		}

		fields[k] = f.CreateField(k, subVal, depth+1)
	}

	return fields
}
