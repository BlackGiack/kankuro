package ui_test

import (
	"strings"
	"testing"

	"kankuro/internal/ui"
)

type testItem struct {
	Name        string
	Description string
}

func TestFilterByQuery(t *testing.T) {
	items := []testItem{
		{Name: "Go", Description: "Programming language"},
		{Name: "Rust", Description: "Systems programming"},
		{Name: "Python", Description: "Scripting language"},
	}

	tests := []struct {
		name     string
		query    string
		wantLen  int
		wantName string
	}{
		{"empty query returns all", "", 3, "Go"},
		{"case insensitive match", "GO", 1, "Go"},
		{"matches description", "script", 1, "Python"},
		{"no match", "java", 0, ""},
		{"partial match", "rust", 1, "Rust"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ui.FilterByQuery(items, tt.query, func(i testItem, q string) bool {
				return contains(i.Name, q) || contains(i.Description, q)
			})
			if len(got) != tt.wantLen {
				t.Errorf("len = %d, want %d", len(got), tt.wantLen)
			}
			if tt.wantLen > 0 && got[0].Name != tt.wantName {
				t.Errorf("got[0].Name = %s, want %s", got[0].Name, tt.wantName)
			}
		})
	}
}

func TestCalculateWindow(t *testing.T) {
	tests := []struct {
		name       string
		cursor     int
		total      int
		viewHeight int
		wantStart  int
		wantEnd    int
	}{
		{"empty list", 0, 0, 10, 0, 0},
		{"list shorter than view", 0, 5, 10, 0, 5},
		{"cursor at top", 2, 20, 6, 0, 6},
		{"cursor centered", 10, 20, 6, 7, 13},
		{"cursor at bottom", 18, 20, 6, 14, 20},
		{"exact fit", 0, 6, 6, 0, 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end := ui.CalculateWindow(tt.cursor, tt.total, tt.viewHeight)
			if start != tt.wantStart {
				t.Errorf("start = %d, want %d", start, tt.wantStart)
			}
			if end != tt.wantEnd {
				t.Errorf("end = %d, want %d", end, tt.wantEnd)
			}
		})
	}
}

func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
