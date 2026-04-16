package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type tickMsg time.Time

type attachFinishedMsg struct {
	sessionID string
	err       error
}

type launchFinishedMsg struct {
	sessionID string
	err       error
}

// Update handles key events, timers, and view-specific state transitions.
//
//nolint:funlen // Central event dispatcher for Bubble Tea view modes.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.resize()
	case tickMsg:
		a.sessions = a.session.List()
		a.refreshSessionList()
		return a, tick()
	case attachFinishedMsg:
		a.sessionView.ActiveID = msg.sessionID
		a.mode = viewDashboard
		a.refreshSessionList()
		if msg.err != nil {
			a.err = msg.err
		}
	case launchFinishedMsg:
		a.naming.LaunchDone = true
		a.naming.LaunchErr = msg.err
		a.naming.SessionID = msg.sessionID
	}

	switch a.mode {
	case viewCardForm:
		return a.updateCardForm(msg)
	case viewPuppetSelector:
		return a.updatePuppetSelector(msg)
	case viewMCPForm:
		return a.updateMCPForm(msg)
	case viewDeleteConfirm:
		return a.updateDeleteConfirm(msg)
	case viewSwitchAgent:
		return a.updateSwitchAgent(msg)
	case viewAgentsEditor:
		return a.updateAgentsEditor(msg)
	case viewAgentsPathInput:
		return a.updateAgentsPathInput(msg)
	case viewSkillSources:
		return a.updateSkillSources(msg)
	case viewSkillSourceForm:
		return a.updateSkillSourceForm(msg)
	case viewPersonaSources:
		return a.updatePersonaSources(msg)
	case viewPersonaSourceForm:
		return a.updatePersonaSourceForm(msg)
	case viewWorkspaces:
		return a.updateWorkspaces(msg)
	case viewWorkspaceSourceForm:
		return a.updateWorkspaceSourceForm(msg)
	case viewWorkspacePicker:
		return a.updateWorkspacePicker(msg)
	case viewDashboard:
		return a.updateDashboard(msg)
	case viewNaming:
		return a.updateNaming(msg)
	case viewIDE:
		return a.updateIDE(msg)
	case viewIDEForm:
		return a.updateIDEForm(msg)
	default:
		return a, nil
	}
}

// refreshPreviewScrollback loads the latest scrollback for the preview session.
func (a *App) refreshPreviewScrollback() {
	if a.sessionView.PreviewSessionID == "" {
		a.sessionView.PreviewScrollback = nil
		return
	}
	a.sessionView.PreviewScrollback = a.session.GetScrollback(a.sessionView.PreviewSessionID)
}
