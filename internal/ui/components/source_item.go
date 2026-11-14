package components

import (
	"fmt"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
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
	Source contracts.SourceRow
}

func (w WorkspaceSourceItem) Display() string {
	marker := " "
	if w.Source.IsRoot {
		marker = "R"
	}
	return fmt.Sprintf("[%s] %s", marker, w.Source.Path)
}

func (w WorkspaceSourceItem) ID() string {
	return w.Source.Identity
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
func ToWorkspaceItems(sources []contracts.SourceRow) []SourceItem {
	items := make([]SourceItem, len(sources))
	for i, s := range sources {
		items[i] = WorkspaceSourceItem{Source: s}
	}
	return items
}

// PathSourceItem carries the identity of a path-only source.
type PathSourceItem struct{ Path, Identity string }

func (p PathSourceItem) Display() string { return p.Path }
func (p PathSourceItem) ID() string      { return p.Identity }
