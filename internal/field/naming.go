package field

// e.g. "user" -> "userObject"
func defaultObjectNamingFunciton(key string) string {
	return key + "Object"
}
