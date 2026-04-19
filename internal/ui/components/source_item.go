package components

import (
	"fmt"

	"kankuro/internal/config"
)

// SourceItem defines the interface for items in a source list.
type SourceItem interface {
	// Display returns the string representation for the list.
	Display() string

	// ID returns a unique identifier (may be empty for simple items).
	ID() string
}

// SimpleSourceItem is a source item that's just a path string.
// Used for skills and personas sources.
type SimpleSourceItem string

func (s SimpleSourceItem) Display() string { return string(s) }
func (s SimpleSourceItem) ID() string      { return "" } // No separate ID

// WorkspaceSourceItem wraps a workspace source configuration.
type WorkspaceSourceItem struct {
	Source config.WorkspaceSource
}

func (w WorkspaceSourceItem) Display() string {
	marker := " "
	if w.Source.IsRoot {
		marker = "R"
	}
	return fmt.Sprintf("[%s] %s", marker, w.Source.Path)
}

func (w WorkspaceSourceItem) ID() string {
	return w.Source.ID
}

// ToSimpleItems converts string slice to SourceItem slice.
func ToSimpleItems(sources []string) []SourceItem {
	items := make([]SourceItem, len(sources))
	for i, s := range sources {
		items[i] = SimpleSourceItem(s)
	}
	return items
}

// ToWorkspaceItems converts workspace sources to SourceItem slice.
func ToWorkspaceItems(sources []config.WorkspaceSource) []SourceItem {
	items := make([]SourceItem, len(sources))
	for i, s := range sources {
		items[i] = WorkspaceSourceItem{Source: s}
	}
	return items
}

// FromSimpleItems extracts strings from SimpleSourceItem slice.
func FromSimpleItems(items []SourceItem) []string {
	sources := make([]string, 0, len(items))
	for _, item := range items {
		if s, ok := item.(SimpleSourceItem); ok {
			sources = append(sources, string(s))
		}
	}
	return sources
}
