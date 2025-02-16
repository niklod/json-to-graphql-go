package builder

import (
	"encoding/json"
	"testing"

	"github.com/graphql-go/graphql"
	"github.com/niklod/json-to-graphql-go/internal/field"
	"github.com/niklod/json-to-graphql-go/pkg/resolver"
	"github.com/stretchr/testify/assert"
)

// TestTierStatsUnion checks that tierStats union info includes lifesteal and ammo.
func TestTierStatsUnion(t *testing.T) {
	jsonData := `{
        "items": [
            {
                "tierStats": [
                    {
                        "tier": "Silver",
                        "descriptions": ["Shield 5.", "Gain +1 Income."],
                        "cooldown": 3,
                        "multicast": 1,
                        "lifesteal": 1
                    },
                    {
                        "tier": "Gold",
                        "descriptions": ["Shield 10.", "Gain +2 Income."],
                        "cooldown": 3,
                        "multicast": 1,
                        "ammo": 12
                    }
                ]
            }
        ]
    }`

	factory, err := field.NewDefaultFieldFactory(field.Config{Resolver: resolver.NewJSONResolver([]byte(jsonData))})
	builder := NewGraphQLSchemaBuilder(factory)

	var j map[string]interface{}
	err = json.Unmarshal([]byte(jsonData), &j)
	assert.NoError(t, err, "JSON unmarshalling should not error")

	schema, err := builder.BuildSchema(j)
	assert.NoError(t, err, "Schema creation should not error")

	query := `{ 
        items {
            tierStats {
                tier
                descriptions
                cooldown
                multicast
                lifesteal
                ammo
            }
        }
    }`
	params := graphql.Params{
		Schema:        *schema,
		RequestString: query,
	}

	result := graphql.Do(params)
	assert.Empty(t, result.Errors, "GraphQL execution should not error")

	data, ok := result.Data.(map[string]interface{})
	assert.True(t, ok, "Result data should be a map")

	items, ok := data["items"].([]interface{})
	assert.True(t, ok, "Items should be a list")
	assert.Equal(t, 1, len(items), "There should be 1 item")

	item := items[0].(map[string]interface{})

	tierStats, ok := item["tierStats"].([]interface{})
	assert.True(t, ok, "tierStats should be a list")

	// Check that union fields are present.
	for _, ts := range tierStats {
		stat := ts.(map[string]interface{})
		// If lifesteal is missing, it should return null.
		// Likewise for ammo.
		_, _ = stat["lifesteal"], stat["ammo"]
	}
}

// TestEmptySchema verifies that an empty JSON object produces a valid schema.
func TestEmptySchema(t *testing.T) {
	jsonData := `{}`

	factory, err := field.NewDefaultFieldFactory(field.Config{Resolver: resolver.NewJSONResolver([]byte(jsonData))})
	builder := NewGraphQLSchemaBuilder(factory)

	var j map[string]interface{}
	err = json.Unmarshal([]byte(jsonData), &j)
	assert.NoError(t, err, "JSON unmarshalling should not error")

	schema, err := builder.BuildSchema(j)
	assert.NoError(t, err, "Schema creation should not error for empty JSON")

	query := `{ anyField }`
	params := graphql.Params{
		Schema:        *schema,
		RequestString: query,
	}
	result := graphql.Do(params)
	assert.Empty(t, result.Errors, "GraphQL execution should not error")
	data, ok := result.Data.(map[string]interface{})
	assert.True(t, ok, "Result data should be a map")
	assert.Nil(t, data["anyField"], "anyField should be nil")
}

// TestNestedObjects verifies that nested objects are handled correctly.
func TestNestedObjects(t *testing.T) {
	jsonData := `{
        "user": {
            "name": "John",
            "address": {
                "city": "New York",
                "zip": 10001
            }
        }
    }`

	factory, err := field.NewDefaultFieldFactory(field.Config{Resolver: resolver.NewJSONResolver([]byte(jsonData))})
	builder := NewGraphQLSchemaBuilder(factory)

	var j map[string]interface{}
	err = json.Unmarshal([]byte(jsonData), &j)
	assert.NoError(t, err, "JSON unmarshalling should not error")

	schema, err := builder.BuildSchema(j)
	assert.NoError(t, err, "Schema creation should not error")

	query := `{ user { name address { city zip } } }`
	params := graphql.Params{
		Schema:        *schema,
		RequestString: query,
	}

	result := graphql.Do(params)
	assert.Empty(t, result.Errors, "GraphQL execution should not error")

	data, ok := result.Data.(map[string]interface{})
	assert.True(t, ok, "Result data should be a map")

	user, ok := data["user"].(map[string]interface{})
	assert.True(t, ok, "User should be a map")
	assert.Equal(t, "John", user["name"], "Name should be John")

	address, ok := user["address"].(map[string]interface{})
	assert.True(t, ok, "Address should be a map")
	assert.Equal(t, "New York", address["city"], "City should be New York")
	assert.Equal(t, 10001, int(address["zip"].(float64)), "Zip should be 10001")
}

// TestMixedTypesInArray verifies that arrays with mixed types are handled correctly.
func TestMixedTypesInArray(t *testing.T) {
	jsonData := `{
        "values": ["string", 123, true, {"key": "value"}]
    }`

	factory, err := field.NewDefaultFieldFactory(field.Config{Resolver: resolver.NewJSONResolver([]byte(jsonData))})
	builder := NewGraphQLSchemaBuilder(factory)

	var j map[string]interface{}
	err = json.Unmarshal([]byte(jsonData), &j)
	assert.NoError(t, err, "JSON unmarshalling should not error")

	schema, err := builder.BuildSchema(j)
	assert.NoError(t, err, "Schema creation should not error")

	query := `{ values }`
	params := graphql.Params{
		Schema:        *schema,
		RequestString: query,
	}

	result := graphql.Do(params)
	assert.Empty(t, result.Errors, "GraphQL execution should not error")

	data, ok := result.Data.(map[string]interface{})
	assert.True(t, ok, "Result data should be a map")

	values, ok := data["values"].([]interface{})
	assert.True(t, ok, "Values should be a list")
	assert.Equal(t, 4, len(values), "There should be 4 values")
}

// TestEmptyArray verifies that empty arrays are handled correctly.
func TestEmptyArray(t *testing.T) {
	jsonData := `{
        "items": []
    }`

	factory, err := field.NewDefaultFieldFactory(field.Config{Resolver: resolver.NewJSONResolver([]byte(jsonData))})
	builder := NewGraphQLSchemaBuilder(factory)

	var j map[string]interface{}
	err = json.Unmarshal([]byte(jsonData), &j)
	assert.NoError(t, err, "JSON unmarshalling should not error")

	schema, err := builder.BuildSchema(j)
	assert.NoError(t, err, "Schema creation should not error")

	query := `{ items }`
	params := graphql.Params{
		Schema:        *schema,
		RequestString: query,
	}

	result := graphql.Do(params)
	assert.Empty(t, result.Errors, "GraphQL execution should not error")

	data, ok := result.Data.(map[string]interface{})
	assert.True(t, ok, "Result data should be a map")

	items, ok := data["items"]
	assert.True(t, ok, "Items should be presented")
	assert.Nil(t, items, "Items should be nil")
}

// TestComplexNestedStructures verifies that complex nested structures are handled correctly.
func TestComplexNestedStructures(t *testing.T) {
	jsonData := `{
        "user": {
            "name": "John",
            "address": {
                "city": "New York",
                "zip": 10001
            },
            "friends": [
                {"name": "Alice", "age": 25},
                {"name": "Bob", "age": 30}
            ]
        }
    }`

	factory, err := field.NewDefaultFieldFactory(field.Config{Resolver: resolver.NewJSONResolver([]byte(jsonData))})
	builder := NewGraphQLSchemaBuilder(factory)

	var j map[string]interface{}
	err = json.Unmarshal([]byte(jsonData), &j)
	assert.NoError(t, err, "JSON unmarshalling should not error")

	schema, err := builder.BuildSchema(j)
	assert.NoError(t, err, "Schema creation should not error")

	query := `{ user { name address { city zip } friends { name age } } }`
	params := graphql.Params{
		Schema:        *schema,
		RequestString: query,
	}

	result := graphql.Do(params)
	assert.Empty(t, result.Errors, "GraphQL execution should not error")

	data, ok := result.Data.(map[string]interface{})
	assert.True(t, ok, "Result data should be a map")

	user, ok := data["user"].(map[string]interface{})
	assert.True(t, ok, "User should be a map")
	assert.Equal(t, "John", user["name"], "Name should be John")

	address, ok := user["address"].(map[string]interface{})
	assert.True(t, ok, "Address should be a map")
	assert.Equal(t, "New York", address["city"], "City should be New York")
	assert.Equal(t, 10001, int(address["zip"].(float64)), "Zip should be 10001")

	friends, ok := user["friends"].([]interface{})
	assert.True(t, ok, "Friends should be a list")
	assert.Equal(t, 2, len(friends), "There should be 2 friends")

	friend1 := friends[0].(map[string]interface{})
	assert.Equal(t, "Alice", friend1["name"], "First friend's name should be Alice")
	assert.Equal(t, 25, int(friend1["age"].(float64)), "First friend's age should be 25")

	friend2 := friends[1].(map[string]interface{})
	assert.Equal(t, "Bob", friend2["name"], "Second friend's name should be Bob")
	assert.Equal(t, 30, int(friend2["age"].(float64)), "Second friend's age should be 30")
}

// TestItemStatsUnion verifies that item stats union info includes all possible attributes.
func TestItemStatsUnion(t *testing.T) {
	jsonData := `{
        "items": [
            {
                "name": "item1",
                "price": 100,
                "stat": {
                    "name": "fire",
                    "level": 5,
                    "value": 50
                }
            },
            {
                "name": "item2",
                "price": 200,
                "stat": {
                    "name": "ice",
                    "level": 3,
                    "resist": 20
                }
            },
            {
                "name": "item2",
                "price": 200,
                "stat": {
                    "name": "ice",
                    "level": 3,
                    "block": 200
                }
            }
        ]
    }`

	factory, err := field.NewDefaultFieldFactory(field.Config{Resolver: resolver.NewJSONResolver([]byte(jsonData))})
	builder := NewGraphQLSchemaBuilder(factory)

	var j map[string]interface{}
	err = json.Unmarshal([]byte(jsonData), &j)
	assert.NoError(t, err, "JSON unmarshalling should not error")

	schema, err := builder.BuildSchema(j)
	assert.NoError(t, err, "Schema creation should not error")

	query := `{ items { name price stat { name level value resist block } } }`
	params := graphql.Params{
		Schema:        *schema,
		RequestString: query,
	}

	result := graphql.Do(params)
	assert.Empty(t, result.Errors, "GraphQL execution should not error")

	data, ok := result.Data.(map[string]interface{})
	assert.True(t, ok, "Result data should be a map")

	items, ok := data["items"].([]interface{})
	assert.True(t, ok, "Items should be a list")
	assert.Equal(t, 3, len(items), "There should be 3 items")

	for _, item := range items {
		itemMap := item.(map[string]interface{})

		stat, ok := itemMap["stat"].(map[string]interface{})
		assert.True(t, ok, "Stat should be a map")

		// Check that all possible attributes are present.
		_ = stat["name"]
		_ = stat["level"]
		_ = stat["value"]
		_ = stat["resist"]
		_ = stat["block"]
	}
}
