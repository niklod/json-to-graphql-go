package field

// GatherUnionInfo function is designed to collect information about potential fields in inconsistent JSON objects.
// It helps determine which fields may be present in the same object at different levels of nesting.
// This data is then used in createObjectField to correctly generate a GraphQL schema that includes all possible fields.
//
// There are three "address" objects in the example JSON data below:
//
//	{
//	    "id": 1,
//	    "name": "Alice",
//	    "parents": [
//	        {
//	            "name": "John",
//	            "age": 50,
//	            "address": { <-- first one: contains unique "phone" field
//	                "city": "New York",
//	                "zip": 10001,
//	                "phone": "123-456-7890"
//	            }
//	        },
//	        {
//	            "name": "Jane",
//	            "age": 45,
//	            "address": { <-- second one: contains unique "street" field
//	                "city": "New York",
//	                "zip": 10001,
//	                "street": "123 Main St"
//	            }
//	        }
//	    ],
//	    "address": { <-- third one: contains both "city" and "zip" fields
//	        "city": "New York",
//	        "zip": 10001
//	    }
//	}
//
// As a result, the unionInfo map will contain the following data:
// =========================
// === Union Info Debug ===
// address : zip
// address : phone
// address : street
// address : city
// =========================
func (f *DefaultFieldFactory) GatherUnionInfo(data interface{}) {
	switch dataType := data.(type) {

	// An object
	case map[string]interface{}:
		f.GatherUnionObjectInfo(dataType)

	// An array
	case []interface{}:
		f.GatherUnionListInfo(dataType)
	}
}

func (f *DefaultFieldFactory) GatherUnionObjectInfo(data map[string]interface{}) {
	// For each attribute in object
	for key, value := range data {
		// If value is an object, record its subkeys under the bare key.
		if subObj, ok := valueIsObject(value); ok {
			f.unionInfo.makeIfNotExists(key)

			for subKey, subVal := range subObj {
				_, isObj := valueIsObject(subVal)
				if isObj || !f.unionInfo.subkeyExists(key, subKey) {
					f.unionInfo.setSubkey(key, subKey, isObj)
					// f.unionInfo.printDebugState()
				}
			}
		}

		// Recurse into the value.
		f.GatherUnionInfo(value)
	}
}

func (f *DefaultFieldFactory) GatherUnionListInfo(data []interface{}) {
	for _, item := range data {
		f.GatherUnionInfo(item)
	}
}
