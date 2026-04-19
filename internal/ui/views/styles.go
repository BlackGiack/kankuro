package views

import "github.com/charmbracelet/lipgloss"

// Theme colors shared across views
var (
	primaryColor   = lipgloss.Color("#8E44AD")                               //nolint:gochecknoglobals // UI theme colors
	highlightColor = lipgloss.Color("#16A085")                               //nolint:gochecknoglobals // UI theme colors
	gray           = lipgloss.Color("240")                                   //nolint:gochecknoglobals // UI theme colors
	secondaryColor = lipgloss.Color("#C0392B")                               //nolint:gochecknoglobals // UI theme colors
	titleStyle     = lipgloss.NewStyle().Bold(true).Foreground(primaryColor) //nolint:gochecknoglobals // UI theme styles
)
