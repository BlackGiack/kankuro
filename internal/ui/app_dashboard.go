package ui

import (
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"kankuro/internal/config"
	"kankuro/internal/session"
)

func (a *App) updateDashboard(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "q", "ctrl+c":
			return a, tea.Quit
		case keyEsc:
			return a, nil
		case keyTab, "shift+tab", "left", "right":
			a.toggleDashboardPane()
			return a, nil
		case "n":
			a.resetCardForm()
			a.mode = viewCardForm
			return a, nil
		case "e":
			return a.handleDashboardEdit()
		case "m":
			a.mcpInputs[0].SetValue("")
			a.mcpInputs[1].SetValue("")
			a.mcpInputs[0].Focus()
			a.mode = viewMCPForm
			return a, nil
		case keyEnter:
			return a.handleDashboardEnter()
		case "a":
			return a, a.activateSelectedSession()
		case "x":
			if a.focusedPane != paneSessions {
				return a, nil
			}
			if err := a.killSelectedSession(); err != nil {
				a.err = err
			}
			a.refreshSessionList()
			return a, nil
		case "backspace", "delete":
			return a.handleDashboardDelete()
		case "s":
			return a.handleDashboardSwitch()
		}
	}

	if a.focusedPane == paneSessions {
		var cmd tea.Cmd
		a.sessionsList, cmd = a.sessionsList.Update(msg)
		return a, cmd
	}
	var cmd tea.Cmd
	a.cardsList, cmd = a.cardsList.Update(msg)
	a.updateCardSelection()
	return a, cmd
}

func (a *App) handleDashboardEdit() (tea.Model, tea.Cmd) {
	if a.focusedPane != paneCards {
		return a, nil
	}
	selected, ok := a.selectedCard()
	if !ok {
		return a, nil
	}
	a.cardInputs[0].SetValue(selected.Name)
	a.cardInputs[1].SetValue(selected.Driver)
	a.cardInputs[2].SetValue(selected.Model)
	a.cardInputs[3].SetValue(selected.WorkingDir)
	a.cardInputs[4].SetValue(selected.FaceStyle)
	// puppet field
	if selected.Puppet != "" {
		a.cardPuppet = selected.Puppet
	} else {
		a.cardPuppet = defaultPuppet
	}
	tmpVal := "n"
	if selected.UseTmpWorkspace {
		tmpVal = "y"
	}
	a.cardInputs[5].SetValue(tmpVal)
	a.setSelectedSkills(selected.SkillIDs)
	a.editingCardID = selected.ID
	a.agentsMD = selected.AgentsMD
	a.agentsMDPath = selected.AgentsMDPath
	a.agentsMDPreferPath = selected.AgentsMDPreferPath
	a.agentsPathInput.SetValue(selected.AgentsMDPath)
	a.cardFocusIndex = 0
	a.setCardFocus(a.cardFocusIndex)
	a.mode = viewCardForm
	return a, nil
}

func (a *App) handleDashboardEnter() (tea.Model, tea.Cmd) {
	if a.focusedPane == paneSessions {
		return a, a.activateSelectedSession()
	}
	selected, ok := a.selectedCard()
	if !ok {
		return a, nil
	}
	skills := a.skillsForCard(selected)
	id, err := a.session.Launch(selected, skills)
	if err != nil {
		a.err = err
		return a, nil
	}
	a.activeSessionID = id
	a.mode = viewDashboard
	return a, a.attachSession(id)
}

func (a *App) handleDashboardDelete() (tea.Model, tea.Cmd) {
	if a.focusedPane != paneSessions {
		return a, nil
	}
	sel := a.sessionsList.Index()
	if sel < 0 || sel >= len(a.sessions) {
		return a, nil
	}
	s := a.sessions[sel]
	if s.UseTmpWorkspace {
		a.pendingDeleteID = s.ID
		a.mode = viewDeleteConfirm
		return a, nil
	}
	if err := a.session.Remove(s.ID); err != nil {
		a.err = err
	}
	a.refreshSessionList()
	return a, nil
}

func (a *App) handleDashboardSwitch() (tea.Model, tea.Cmd) {
	if a.focusedPane != paneSessions {
		return a, nil
	}
	sel := a.sessionsList.Index()
	if sel < 0 || sel >= len(a.sessions) {
		return a, nil
	}
	s := a.sessions[sel]
	if s.Status != session.StatusTerminated {
		a.err = errors.New("switch requires a terminated session")
		return a, nil
	}
	if len(a.cards) == 0 {
		a.err = errors.New("no ID cards available")
		return a, nil
	}
	a.switchSessionID = s.ID
	a.switchList.SetItems(toCardItems(a.cards, a.skills))
	a.switchList.Select(0)
	// mark selected for visual feedback
	a.updateSwitchSelection()
	a.mode = viewSwitchAgent
	return a, nil
}

func (a *App) toggleDashboardPane() {
	if a.focusedPane == paneCards && len(a.sessions) > 0 {
		a.focusedPane = paneSessions
		return
	}
	a.focusedPane = paneCards
}

func (a *App) activateSelectedSession() tea.Cmd {
	if len(a.sessions) == 0 {
		return nil
	}
	sel := a.sessionsList.Index()
	if sel < 0 || sel >= len(a.sessions) {
		return nil
	}
	s := a.sessions[sel]

	if s.Status == session.StatusTerminated {
		var card config.IDCard
		found := false
		for _, c := range a.cards {
			if c.ID == s.CardID {
				card = c
				found = true
				break
			}
		}
		if !found {
			a.err = errors.New("cannot relaunch: original ID card not found")
			return nil
		}
		skills := a.skillsForCard(card)
		if err := a.session.Relaunch(s.ID, card, skills); err != nil {
			a.err = fmt.Errorf("relaunch failed: %w", err)
			return nil
		}
		// Refresh locally to reflect status change immediately
		a.sessions = a.session.List()
		a.refreshSessionList()
	}

	a.activeSessionID = s.ID
	a.mode = viewDashboard
	return a.attachSession(a.activeSessionID)
}

func (a *App) killSelectedSession() error {
	if len(a.sessions) == 0 {
		return nil
	}
	sel := a.sessionsList.Index()
	if sel < 0 || sel >= len(a.sessions) {
		return nil
	}
	return a.session.Kill(a.sessions[sel].ID)
}
