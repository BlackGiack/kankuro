package components

import "strings"

// Filter returns the items whose label (as produced by label) contains query
// as a case-insensitive substring. An empty query returns items unchanged.
//
// This is the shared implementation behind card/workspace/agent-switch
// search boxes, which previously each hand-rolled the same
// lowercase-and-contains loop.
func Filter[T any](items []T, query string, label func(T) string) []T {
	if query == "" {
		return items
	}

	q := strings.ToLower(query)
	var filtered []T
	for _, item := range items {
		if strings.Contains(strings.ToLower(label(item)), q) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}
