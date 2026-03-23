package ui

import "github.com/charmbracelet/lipgloss"

//nolint:gochecknoglobals // These are style definitions.
var (
	// Theme Colors.
	deepCharcoal = lipgloss.Color("#1A1A1A")
	richPurple   = lipgloss.Color("#8E44AD")
	crimsonRed   = lipgloss.Color("#C0392B")
	tealCyan     = lipgloss.Color("#16A085")
	white        = lipgloss.Color("#FFFFFF")
	gray         = lipgloss.Color("240")

	// Semantic Colors.
	primaryColor     = richPurple
	secondaryColor   = crimsonRed
	highlightColor   = tealCyan
	personaPillColor = crimsonRed

	// Text Styles.
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor)

	errorStyle = lipgloss.NewStyle().
			Foreground(secondaryColor)

	// Borders.
	focusedBorderColor = primaryColor
	normalBorderColor  = gray

	// Components.
	faceBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(primaryColor).
			Background(deepCharcoal).
			Align(lipgloss.Center, lipgloss.Center)
)
