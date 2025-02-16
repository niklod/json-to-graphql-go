package field

import "github.com/graphql-go/graphql"

type unionMap map[string]map[string]bool // bare key -> subfield -> isObject

func (u unionMap) makeIfNotExists(key string) {
	if u[key] == nil {
		u[key] = make(map[string]bool)
	}
}

func (u unionMap) subkeyExists(key, subKey string) bool {
	if u[key] == nil {
		return false
	}

	return u[key][subKey]
}

func (u unionMap) setSubkey(key, subKey string, value bool) {
	u.makeIfNotExists(key)
	u[key][subKey] = value
}

func (u unionMap) reset() {
	u = make(unionMap)
}

func (u unionMap) get(key string) (map[string]bool, bool) {
	v, ok := u[key]
	return v, ok
}

func (u unionMap) getSubkey(key, subKey string) (bool, bool) {
	if u[key] == nil {
		return false, false
	}

	v, ok := u[key][subKey]
	return v, ok
}

func (u unionMap) printDebugState() {
	println("=== Union Info Debug ===")
	for key, subkeys := range u {
		for subkey, isObj := range subkeys {
			if isObj {
				println(key, ":", subkey, "isObject")
			} else {
				println(key, ":", subkey)
			}
		}
	}
	println("=========================")
}

type gqlTypesCache map[string]*graphql.Object // type name -> GraphQL object

func (g gqlTypesCache) reset() {
	g = make(gqlTypesCache)
}

func (g gqlTypesCache) set(key string, value *graphql.Object) {
	g[key] = value
}

func (g gqlTypesCache) get(key string) (*graphql.Object, bool) {
	v, ok := g[key]
	return v, ok
}

func valueIsObject(value interface{}) (map[string]interface{}, bool) {
	v, ok := value.(map[string]interface{})
	return v, ok
}

func valueIsArray(value interface{}) ([]interface{}, bool) {
	v, ok := value.([]interface{})
	return v, ok
}
