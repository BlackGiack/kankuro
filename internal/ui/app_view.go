package ui

import (
	"fmt"

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
		return errLine + titleStyle.Render("Delete Session?") + "\n\n" +
			fmt.Sprintf("Are you sure you want to delete session %s?\n", a.pendingDeleteID) +
			"This will also delete its temporary workspace if it exists.\n\n" +
			"y=Confirm  n=Cancel"
	case viewCardForm:
		title := "New ID Card"
		if a.editingCardID != "" {
			title = "Edit ID Card"
		}
		return errLine + titleStyle.Render(title) + "\n\n" +
			"Name: " + a.cardInputs[0].View() + "\n" +
			"Driver: " + a.cardInputs[1].View() + "\n" +
			"Model: " + a.cardInputs[2].View() + "\n" +
			"Working Dir: " + a.cardInputs[3].View() + "\n" +
			"Face Style: " + a.cardInputs[4].View() + "\n" +
			"Use Tmp Workspace: " + a.cardInputs[5].View() + "\n\n" +
			"AGENTS.md:\n" + a.renderAgentsMDRow() + "\n\n" +
			"Skills (type=search, space=toggle, ctrl+e=sources):\n" + a.renderSkillList() + "\n\n" +
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
		right := a.sessionsList.View()
		body := lipgloss.JoinHorizontal(lipgloss.Top,
			leftStyle.Render(left),
			rightStyle.Render(right),
		)
		return errLine + titleStyle.Render("KANKURO Dashboard") + "\n" + body + "\n" +
			"Tab=switch Enter=launch/attach x=kill Del=remove terminated s=switch n=new e=edit m=mcp q=quit"
	default:
		return "Unknown view"
	}
}
