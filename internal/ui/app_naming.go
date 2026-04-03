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

	if a.namingLaunchDone && a.namingLaunchErr != nil {
		a.err = a.namingLaunchErr
		a.mode = viewDashboard
		return a, nil
	}

	if a.namingLaunchDone && a.namingSubmitted {
		return a.doAttach()
	}

	var cmd tea.Cmd
	a.namingInput, cmd = a.namingInput.Update(msg)
	return a, cmd
}

func (a *App) handleNamingEnter() (tea.Model, tea.Cmd) {
	if !a.namingLaunchDone {
		a.namingSubmitted = true
		return a, nil
	}
	return a.doAttach()
}

func (a *App) doAttach() (tea.Model, tea.Cmd) {
	name := strings.TrimSpace(a.namingInput.Value())
	if name == "" {
		name = a.namingSessionID
	}
	if err := a.session.SetName(a.namingSessionID, name); err != nil {
		a.err = err
	}
	a.activeSessionID = a.namingSessionID
	a.mode = viewDashboard
	return a, a.attachSession(a.namingSessionID)
}
