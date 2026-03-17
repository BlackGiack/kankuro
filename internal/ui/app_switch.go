package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

func (a *App) updateSwitchAgent(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "q", keyCtrlC:
			return a, tea.Quit
		case keyEsc:
			a.switchSessionID = ""
			a.mode = viewDashboard
			return a, nil
		case keyEnter:
			card, okSwitch := a.selectedSwitchCard()
			if !okSwitch {
				return a, nil
			}
			skills := a.skillsForCard(card)
			if err := a.session.SwitchAgent(a.switchSessionID, card, skills); err != nil {
				a.err = fmt.Errorf("switch agent failed: %w", err)
				return a, nil
			}
			a.sessions = a.session.List()
			a.refreshSessionList()
			id := a.switchSessionID
			a.switchSessionID = ""
			a.mode = viewDashboard
			a.activeSessionID = id
			return a, a.attachSession(id)
		}
	}
	var cmd tea.Cmd
	a.switchList, cmd = a.switchList.Update(msg)
	return a, cmd
}

func (a *App) updateDeleteConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "q", keyCtrlC:
			return a, tea.Quit
		case "y", "Y":
			if err := a.session.Remove(a.pendingDeleteID); err != nil {
				a.err = err
			}
			a.pendingDeleteID = ""
			a.mode = viewDashboard
			a.refreshSessionList()
			return a, nil
		case "n", "N", keyEsc:
			a.pendingDeleteID = ""
			a.mode = viewDashboard
			return a, nil
		}
	}
	return a, nil
}
