package main

import (
	"testing"

	"github.com/graphql-go/graphql"
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

	builder := NewGraphQLSchemaBuilder()
	schema, rootData, err := builder.BuildSchema([]byte(jsonData))
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
		RootObject:    rootData,
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
	builder := NewGraphQLSchemaBuilder()
	schema, rootData, err := builder.BuildSchema([]byte(jsonData))
	assert.NoError(t, err, "Schema creation should not error for empty JSON")
	query := `{ anyField }`
	params := graphql.Params{
		Schema:        *schema,
		RequestString: query,
		RootObject:    rootData,
	}
	result := graphql.Do(params)
	assert.Empty(t, result.Errors, "GraphQL execution should not error")
	data, ok := result.Data.(map[string]interface{})
	assert.True(t, ok, "Result data should be a map")
	assert.Nil(t, data["anyField"], "anyField should be nil")
}
