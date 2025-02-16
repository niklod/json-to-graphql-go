package field

// mergeMaps returns the union of a slice of maps recursively.
func mergeMaps(maps []map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})

	for _, m := range maps {
		for k, v := range m {
			existing, exists := merged[k]

			if !exists {
				merged[k] = v

				continue
			}

			exMap, ok1 := existing.(map[string]interface{})
			newMap, ok2 := v.(map[string]interface{})

			if ok1 && ok2 {
				merged[k] = mergeMaps([]map[string]interface{}{exMap, newMap})

				continue
			}

			merged[k] = v
		}
	}

	return merged
}
