package strings

// ContainsString is helper functions to check string from a slice of strings.
func ContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}

	return false
}

// RemoveString is helper functions to remove string from a slice of strings.
func RemoveString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}

		result = append(result, item)
	}

	return
}
