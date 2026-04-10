package ui

import "github.com/charmbracelet/lipgloss"

func (a *App) viewCardFormRender() string {
	title := "New ID Card"
	if a.cardForm.EditingID != "" {
		title = "Edit ID Card"
	}
	workspaceLine := a.renderWorkspaceLine()
	puppetLine := a.renderPuppetLine()
	return titleStyle.Render(title) + "\n\n" +
		"Name: " + a.cardForm.Inputs[0].View() + "\n" +
		"Driver: " + a.cardForm.Inputs[1].View() + "\n" +
		"Model: " + a.cardForm.Inputs[2].View() + "\n" +
		workspaceLine + "\n" +
		"Face Style: " + a.cardForm.Inputs[3].View() + "\n" +
		puppetLine + "\n" +
		"Use Tmp Workspace: " + a.cardForm.Inputs[4].View() + "\n\n" +
		"AGENTS.md:\n" + a.renderAgentsMDRow() + "\n\n" +
		a.renderSkillList() + "\n\n" +
		a.renderPersonaList() + "\n\n" +
		a.renderMCPList() + "\n\n" +
		"Enter=save Tab=next Esc=cancel"
}

func (a *App) renderWorkspaceLine() string {
	line := "Workspace: " + a.workspaceSummary()
	if a.cardForm.FocusIndex == cardWorkspaceIndex {
		line = "> " + line
		line = lipgloss.NewStyle().Foreground(highlightColor).Bold(true).Render(line)
	}
	return line
}

func (a *App) renderPuppetLine() string {
	line := "Puppet: " + a.cardForm.Puppet
	if a.isPuppetFocused() {
		line = "> " + line
		line = lipgloss.NewStyle().Foreground(highlightColor).Bold(true).Render(line)
	}
	return line
}
