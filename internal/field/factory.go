package field

import (
	"sort"

	"github.com/graphql-go/graphql"
)

type Config struct {
	GQLObjectNamingFn func(key string) string
	Resolver          Resolver
}

// DefaultFieldFactory is the default implementation.
type DefaultFieldFactory struct {
	unionInfo     unionMap
	gqlTypesCache gqlTypesCache
	resolver      Resolver
	objectNameFn  func(key string) string
}

// NewDefaultFieldFactory creates a new DefaultFieldFactory.
func NewDefaultFieldFactory(config Config) (*DefaultFieldFactory, error) {
	if config.Resolver == nil {
		return nil, ErrResolverNotProvided
	}

	if config.GQLObjectNamingFn == nil {
		config.GQLObjectNamingFn = defaultObjectNamingFunciton
	}

	return &DefaultFieldFactory{
		unionInfo:     make(unionMap),
		gqlTypesCache: make(gqlTypesCache),
		objectNameFn:  config.GQLObjectNamingFn,
		resolver:      config.Resolver,
	}, nil
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
	f.gqlTypesCache.reset()
	f.unionInfo.reset()
}

// allMaps returns true if every element in the array is a map.
// Also converts the array to a slice of maps.
func allMaps(arr []interface{}) (bool, []map[string]interface{}) {
	var res []map[string]interface{}

	for _, e := range arr {
		v, ok := e.(map[string]interface{})
		if !ok {
			return false, nil
		}

		res = append(res, v)
	}
	return true, res
}
