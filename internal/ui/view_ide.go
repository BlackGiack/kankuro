package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

const (
	ideCurrentBoxPaddingInner = 2
	ideCurrentBoxWidthOffset  = 4
	ideExplanationWidthOffset = 4
	ideInputBoxPaddingInner   = 2
)

func (a *App) viewIDERender() string {
	title := titleStyle.Padding(0, 1).Render("IDE Configuration")

	currentCmd := a.ideConfig.Command
	if currentCmd == "" {
		currentCmd = "(not configured)"
	}

	currentBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(gray).
		Padding(1, ideCurrentBoxPaddingInner).
		Width(min(a.width-ideCurrentBoxWidthOffset, inputWidthDefault+ideCurrentBoxWidthOffset)).
		Render(fmt.Sprintf("Command: %s", currentCmd))

	helpText := lipgloss.NewStyle().
		Foreground(gray).
		Padding(1, 0).
		Render("Press e or Enter to edit \u2022 Esc to go back")

	pathHelp := lipgloss.NewStyle().
		Foreground(secondaryColor).
		Render(fmt.Sprintf("Use %s in your command to reference the session workspace path", idePathPlaceholder))

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		currentBox,
		"",
		pathHelp,
		"",
		helpText,
	)
}

func (a *App) viewIDEFormRender() string {
	title := titleStyle.Padding(0, 1).Render("Edit IDE Command")

	explanation := lipgloss.NewStyle().
		Foreground(secondaryColor).
		Width(min(a.width-ideExplanationWidthOffset, inputWidthDefault)).
		Render(fmt.Sprintf(
			"Enter the command to launch your IDE.\n"+
				"Use %s as a placeholder for the session workspace path.\n"+
				"Example: code %s",
			idePathPlaceholder, idePathPlaceholder,
		))

	inputBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(focusedBorderColor).
		Padding(1, ideInputBoxPaddingInner).
		Render(a.ideInput.View())

	footer := lipgloss.NewStyle().
		Foreground(gray).
		Render("Enter to save \u2022 Esc to cancel")

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		explanation,
		"",
		inputBox,
		"",
		footer,
	)
}
