package ui

import "strings"

func FilterByQuery[T any](items []T, query string, matcher func(T, string) bool) []T {
	if query == "" {
		return items
	}
	var filtered []T
	lowerQuery := strings.ToLower(query)
	for _, item := range items {
		if matcher(item, lowerQuery) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func CalculateWindow(cursor, total, viewHeight int) (int, int) {
	if total == 0 {
		return 0, 0
	}
	if viewHeight >= total {
		return 0, total
	}
	start := max(0, cursor-viewHeight/2)
	end := start + viewHeight
	if end > total {
		end = total
		start = max(0, end-viewHeight)
	}
	return start, end
}
