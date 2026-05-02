package utils

import "strings"

// SplitAndTrim splits a comma-separated string into a cleaned slice,
// trimming whitespace from each part and discarding empty entries.
func SplitAndTrim(s string) []string {
	if strings.TrimSpace(s) == "" {
		return []string{}
	}
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
