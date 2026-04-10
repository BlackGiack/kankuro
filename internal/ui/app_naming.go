package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (a *App) updateNaming(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, isKey := msg.(tea.KeyMsg)
	if isKey {
		switch keyMsg.String() {
		case "q", keyCtrlC:
			return a, tea.Quit
		case keyEnter:
			return a.handleNamingEnter()
		case keyEsc:
		}
	}

	if a.naming.LaunchDone && a.naming.LaunchErr != nil {
		a.err = a.naming.LaunchErr
		a.mode = viewDashboard
		return a, nil
	}

	if a.naming.LaunchDone && a.naming.Submitted {
		return a.doAttach()
	}

	var cmd tea.Cmd
	a.naming.Input, cmd = a.naming.Input.Update(msg)
	return a, cmd
}

func (a *App) handleNamingEnter() (tea.Model, tea.Cmd) {
	if !a.naming.LaunchDone {
		a.naming.Submitted = true
		return a, nil
	}
	return a.doAttach()
}

func (a *App) doAttach() (tea.Model, tea.Cmd) {
	name := strings.TrimSpace(a.naming.Input.Value())
	if name == "" {
		name = a.naming.SessionID
	}
	if err := a.session.SetName(a.naming.SessionID, name); err != nil {
		a.err = err
	}
	a.sessionView.ActiveID = a.naming.SessionID
	a.mode = viewDashboard
	return a, a.attachSession(a.naming.SessionID)
}
