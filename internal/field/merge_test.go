package field

import (
	"reflect"
	"testing"
)

func TestMergeMaps(t *testing.T) {
	tests := []struct {
		name     string
		input    []map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "Distinct keys",
			input: []map[string]interface{}{
				{"a": 1},
				{"b": 2},
			},
			expected: map[string]interface{}{"a": 1, "b": 2},
		},
		{
			// In this case we'll keep the last value.
			name: "Overlapping keys with different types",
			input: []map[string]interface{}{
				{"a": 1},
				{"a": "string"},
			},
			expected: map[string]interface{}{"a": "string"},
		},
		{
			name: "Nested maps",
			input: []map[string]interface{}{
				{"a": map[string]interface{}{"x": 1}},
				{"a": map[string]interface{}{"y": 2}},
			},
			expected: map[string]interface{}{
				"a": map[string]interface{}{
					"x": 1,
					"y": 2,
				},
			},
		},
		{
			// In this case we'll keep the last value.
			name: "Nested map conflict",
			input: []map[string]interface{}{
				{"a": map[string]interface{}{"x": 1}},
				{"a": 42},
			},
			expected: map[string]interface{}{"a": 42},
		},
		{
			name:     "Empty slice",
			input:    []map[string]interface{}{},
			expected: map[string]interface{}{},
		},
		{
			name: "One empty map",
			input: []map[string]interface{}{
				{"a": 1},
				{},
			},
			expected: map[string]interface{}{"a": 1},
		},
		{
			name: "Deeply nested merge",
			input: []map[string]interface{}{
				{"a": map[string]interface{}{"b": map[string]interface{}{"c": 1}}},
				{"a": map[string]interface{}{"b": map[string]interface{}{"d": 2}}},
			},
			expected: map[string]interface{}{
				"a": map[string]interface{}{
					"b": map[string]interface{}{
						"c": 1,
						"d": 2,
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := mergeMaps(test.input)
			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("Test %s failed: expected %v, got %v", test.name, test.expected, result)
			}
		})
	}
}
