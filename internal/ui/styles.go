package ui

import "github.com/charmbracelet/lipgloss"

var (
	purple     = lipgloss.Color("99")
	hotPink    = lipgloss.Color("201")
	darkPurple = lipgloss.Color("57")
	white      = lipgloss.Color("255")
	black      = lipgloss.Color("0")
	gray       = lipgloss.Color("240")
	red        = lipgloss.Color("196")

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(purple)

	errorStyle = lipgloss.NewStyle().
			Foreground(red)

	focusedBorderColor = purple
	normalBorderColor  = gray
)
