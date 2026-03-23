package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"kankuro/internal/ascii"
	"kankuro/internal/ascii/characters"
)

// View renders the active screen for the Bubble Tea program.
// TODO: This function is quite long and could be refactored into smaller render functions for each view mode.
//
//nolint:funlen // This is the main view function that delegates to sub-renderers based on the current mode.
func (a *App) View() string {
	errLine := ""
	if a.err != nil {
		errLine = errorStyle.Render(a.err.Error()) + "\n"
	}

	switch a.mode {
	case viewDeleteConfirm:
		return errLine + titleStyle.Render("Delete Session?") + "\n\n" +
			fmt.Sprintf("Are you sure you want to delete session %s?\n", a.pendingDeleteID) +
			"This will also delete its temporary workspace if it exists.\n\n" +
			"y=Confirm  n=Cancel"
	case viewCardForm:
		title := "New ID Card"
		if a.editingCardID != "" {
			title = "Edit ID Card"
		}
		// Render puppet row with focus indicator (no preview)
		puppetLine := "Puppet: " + a.cardPuppet
		if a.isPuppetFocused() {
			puppetLine = "> " + puppetLine
			puppetLine = lipgloss.NewStyle().Foreground(highlightColor).Bold(true).Render(puppetLine)
		}
		return errLine + titleStyle.Render(title) + "\n\n" +
			"Name: " + a.cardInputs[0].View() + "\n" +
			"Driver: " + a.cardInputs[1].View() + "\n" +
			"Model: " + a.cardInputs[2].View() + "\n" +
			"Working Dir: " + a.cardInputs[3].View() + "\n" +
			"Face Style: " + a.cardInputs[4].View() + "\n" +
			puppetLine + "\n" +
			"Use Tmp Workspace: " + a.cardInputs[5].View() + "\n\n" +
			"AGENTS.md:\n" + a.renderAgentsMDRow() + "\n\n" +
			a.renderSkillList() + "\n\n" +
			a.renderPersonaList() + "\n\n" +
			"Enter=save Tab=next Esc=cancel"
	case viewSkillSources:
		return errLine + titleStyle.Render("Manage Skill Sources") + "\n\n" +
			"Defined sources:\n" + a.renderSkillSourcesList() + "\n\n" +
			"Enter/Space=edit n=new d=delete Esc=back"
	case viewSkillSourceForm:
		title := "New Skill Source"
		if a.editingSkillSourceIndex != -1 {
			title = "Edit Skill Source"
		}
		return errLine + titleStyle.Render(title) + "\n\n" +
			"Path: " + a.skillSourceInput.View() + "\n\n" +
			"Enter=save Esc=cancel"
	case viewPuppetSelector:
		// render puppet selector
		title := "Select Puppet"
		selected := puppetStyles()[a.puppetSelectorIndex]
		p := characters.NewPuppet(func(c *ascii.Config) { c.Set("style", selected) })
		w, h := p.Bounds()
		canvas := ascii.NewCanvas(w, h)
		p.Render(canvas, 0, 0)
		preview := "\n" + canvas.String() + "\n"
		lines := titleStyle.Render(title) + "\n\n" +
			"Use Left/Right to change puppet, Enter=confirm, Esc=cancel\n\n" +
			"Selected: " + selected + preview
		return errLine + lines
	case viewPersonaSources:
		return errLine + titleStyle.Render("Manage Persona Sources") + "\n\n" +
			"Defined sources:\n" + a.renderPersonaSourcesList() + "\n\n" +
			"Enter/Space=edit n=new d=delete Esc=back"
	case viewPersonaSourceForm:
		title := "New Persona Source"
		if a.editingPersonaSourceIndex != -1 {
			title = "Edit Persona Source"
		}
		return errLine + titleStyle.Render(title) + "\n\n" +
			"Path: " + a.personaSourceInput.View() + "\n\n" +
			"Enter=save Esc=cancel"
	case viewAgentsEditor:
		return errLine + titleStyle.Render("Editing AGENTS.md") + "\n\n" +
			a.agentsEditor.View() + "\n\n" +
			"Ctrl+s=save Esc=cancel"
	case viewAgentsPathInput:
		return errLine + titleStyle.Render("Enter AGENTS.md Path") + "\n\n" +
			a.agentsPathInput.View() + "\n\n" +
			"Enter=save Esc=cancel"
	case viewMCPForm:
		return errLine + titleStyle.Render("Add MCP") + "\n\n" +
			"Name: " + a.mcpInputs[0].View() + "\n" +
			"Command: " + a.mcpInputs[1].View() + "\n\n" +
			"Enter=save Tab=next Esc=cancel"
	case viewSwitchAgent:
		header := "Switch Agent"
		if a.switchSessionID != "" {
			header = fmt.Sprintf("Switch Agent (session %s)", a.switchSessionID)
		}
		return errLine + titleStyle.Render(header) + "\n\n" +
			a.switchList.View() + "\n" +
			"Enter=switch Esc=cancel"
	case viewDashboard:
		return errLine + a.renderDashboard()
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

	if a.focusedPane == paneCards {
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
		lipgloss.NewStyle().Foreground(gray).Render(" [Card] Enter=Launch n=New e=Edit m=MCP "),
		lipgloss.NewStyle().Foreground(gray).Render(" [Session] Enter=Attach x=Kill Del=Remove "),
		lipgloss.NewStyle().Foreground(gray).Render(" [Global] q=Quit"),
	)

	return header + "\n" + body + "\n" + footer
}
