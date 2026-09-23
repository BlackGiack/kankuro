package component

import tea "charm.land/bubbletea/v2"

// Model is the base interface for all UI components.
type Model interface {
	// Init initializes the component (called once on mount).
	Init() tea.Cmd

	// Update handles messages and returns updated model + commands.
	Update(tea.Msg) (Model, tea.Cmd)

	// View renders the component as a string.
	View() string

	// SetSize updates component dimensions.
	SetSize(width, height int)
}

// Base provides common fields embedded in components.
// It implements SetSize and provides Width/Height accessors.
type Base struct {
	Width   int
	Height  int
	Focused bool
}

func (b *Base) SetSize(width, height int) {
	b.Width = width
	b.Height = height
}
