package styles

import "charm.land/lipgloss/v2"

// Theme colors shared across views.
var (
	PrimaryColor   = lipgloss.Color("#8E44AD")                                            //nolint:gochecknoglobals // UI theme colors
	HighlightColor = lipgloss.Color("#16A085")                                            //nolint:gochecknoglobals // UI theme colors
	Gray           = lipgloss.Color("240")                                                //nolint:gochecknoglobals // UI theme colors
	SecondaryColor = lipgloss.Color("#C0392B")                                            //nolint:gochecknoglobals // UI theme colors
	White          = lipgloss.Color("#FFFFFF")                                            //nolint:gochecknoglobals // UI theme colors
	HotPink        = lipgloss.Color("#FF69B4")                                            //nolint:gochecknoglobals // UI theme colors
	DeepCharcoal   = lipgloss.Color("#1A1A1A")                                            //nolint:gochecknoglobals // UI theme colors
	TitleStyle     = lipgloss.NewStyle().Bold(true).Foreground(PrimaryColor)              //nolint:gochecknoglobals // UI theme styles
	SuccessStyle   = lipgloss.NewStyle().Foreground(HighlightColor).Bold(true)            //nolint:gochecknoglobals // UI theme styles
	LabelStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true) //nolint:gochecknoglobals // UI theme styles
	HelpStyle      = lipgloss.NewStyle().Foreground(Gray)                                 //nolint:gochecknoglobals // UI theme styles
	ErrorStyle     = lipgloss.NewStyle().Foreground(SecondaryColor).Bold(true)            //nolint:gochecknoglobals // UI theme styles
)
