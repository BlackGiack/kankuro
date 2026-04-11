package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// View renders the active screen for the Bubble Tea program.
func (a *App) View() string {
	errLine := ""
	if a.err != nil {
		errLine = errorStyle.Render(a.err.Error()) + "\n"
	}

	switch a.mode {
	case viewDeleteConfirm:
		return errLine + a.viewDeleteConfirmRender()
	case viewCardForm:
		return errLine + a.viewCardFormRender()
	case viewPuppetSelector:
		return errLine + a.viewPuppetSelectorRender()
	case viewSkillSources:
		return errLine + a.viewSkillSourcesRender()
	case viewSkillSourceForm:
		return errLine + a.viewSkillSourceFormRender()
	case viewPersonaSources:
		return errLine + a.viewPersonaSourcesRender()
	case viewPersonaSourceForm:
		return errLine + a.viewPersonaSourceFormRender()
	case viewWorkspaces:
		return errLine + a.viewWorkspacesRender()
	case viewWorkspaceSourceForm:
		return errLine + a.viewWorkspaceSourceFormRender()
	case viewWorkspacePicker:
		return errLine + a.viewWorkspacePickerRender()
	case viewAgentsEditor:
		return errLine + a.viewAgentsEditorRender()
	case viewAgentsPathInput:
		return errLine + a.viewAgentsPathInputRender()
	case viewMCPForm:
		return errLine + a.viewMCPFormRender()
	case viewSwitchAgent:
		return errLine + a.viewSwitchAgentRender()
	case viewNaming:
		return errLine + a.viewNamingRender()
	case viewDashboard:
		return errLine + a.renderDashboard()
	case viewIDE:
		return errLine + a.viewIDERender()
	case viewIDEForm:
		return errLine + a.viewIDEFormRender()
	default:
		return "Unknown view"
	}
}

func (a *App) renderDashboard() string {
	const borderWidth = 2

	leftStyle := lipgloss.NewStyle().Width(max(switchListMinWidth, a.width-cardsListPadding)).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(normalBorderColor)
	rightStyle := lipgloss.NewStyle().Width(sessionListWidth).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(normalBorderColor)

	if a.sessionView.FocusedPane == paneCards {
		leftStyle = leftStyle.BorderForeground(focusedBorderColor)
	} else {
		rightStyle = rightStyle.BorderForeground(focusedBorderColor)
	}

	left := a.cardsList.View()
	var right string
	if len(a.sessions) == 0 {
		crowASCII := `
      ,   ,
     / \_/ \
    ( o . o )
     \  ^  /
      |||||

   The puppets
   are resting
`
		right = lipgloss.NewStyle().
			Width(sessionListWidth-borderWidth).
			Height(a.sessionsList.Height()).
			Align(lipgloss.Center, lipgloss.Center).
			Foreground(gray).
			Render(crowASCII)
	} else {
		right = a.sessionsList.View()
	}

	body := lipgloss.JoinHorizontal(lipgloss.Top,
		leftStyle.Render(left),
		rightStyle.Render(right),
	)

	header := lipgloss.NewStyle().
		Foreground(white).
		Background(primaryColor).
		Padding(0, headerPadding).
		MarginBottom(1).
		Bold(true).
		Render("K A N K U R O")

	footer := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Foreground(gray).Render("[Nav] Tab "),
		lipgloss.NewStyle().Foreground(gray).Render(" [Card] Enter=Launch n=New e=Edit "),
		lipgloss.NewStyle().Foreground(gray).Render(" [Session] Enter=Attach x=Kill Del=Remove o=Open "),
		lipgloss.NewStyle().Foreground(gray).Render(" [Global] m=MCP w=Workspaces i=IDE q=Quit"),
	)

	return header + "\n" + body + "\n" + footer
}
