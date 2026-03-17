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

// Update handles key events, timers, and view-specific state transitions.
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
		a.activeSessionID = msg.sessionID
		a.mode = viewDashboard
		a.refreshSessionList()
		if msg.err != nil {
			a.err = msg.err
		}
	}

	switch a.mode {
	case viewCardForm:
		return a.updateCardForm(msg)
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
	case viewDashboard:
		return a.updateDashboard(msg)
	default:
		return a, nil
	}
}
