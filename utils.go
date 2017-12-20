package jargo

// TODO: remove if unused
func containsValue(slice []string, value string) bool {
	for _, val := range slice {
		if val == value {
			return true
		}
	}

	return false
}