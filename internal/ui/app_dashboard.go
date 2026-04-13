package ui

import (
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"kankuro/internal/config"
	"kankuro/internal/session"
)

//nolint:funlen // Central event dispatcher for dashboard view.
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
			// open mcp form
			a.cardForm.MCPFormPrevMode = viewDashboard
			for i := range a.mcpInputs {
				a.mcpInputs[i].SetValue("")
			}
			a.mcpInputs[0].Focus()
			a.mode = viewMCPForm
			return a, nil
		case "w":
			a.refreshWorkspaces()
			a.workspaceForm.Cursor = 0
			a.mode = viewWorkspaces
			return a, nil
		case "i":
			return a.openIDEConfig()
		case "o":
			if a.sessionView.FocusedPane != paneSessions {
				return a, nil
			}
			a.launchIDE()
			a.refreshSessionList()
			return a, nil
		case keyEnter:
			return a.handleDashboardEnter()
		case "a":
			return a, a.activateSelectedSession()
		case "x":
			if a.sessionView.FocusedPane != paneSessions {
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

	if a.sessionView.FocusedPane == paneSessions {
		var cmd tea.Cmd
		a.sessionsList, cmd = a.sessionsList.Update(msg)
		return a, cmd
	}
	var cmd tea.Cmd
	a.cardsList, cmd = a.cardsList.Update(msg)
	return a, cmd
}

func (a *App) handleDashboardEdit() (tea.Model, tea.Cmd) {
	if a.sessionView.FocusedPane != paneCards {
		return a, nil
	}
	selected, ok := a.selectedCard()
	if !ok {
		return a, nil
	}
	a.cardForm.Inputs[0].SetValue(selected.Name)
	a.cardForm.Inputs[1].SetValue(selected.Driver)
	a.cardForm.Inputs[2].SetValue(selected.Model)
	a.cardForm.SelectedWorkspaceID = selected.WorkspaceID
	a.cardForm.Inputs[3].SetValue(selected.FaceStyle)
	// puppet field
	if selected.Puppet != "" {
		a.cardForm.Puppet = selected.Puppet
	} else {
		a.cardForm.Puppet = defaultPuppet
	}
	tmpVal := "n"
	if selected.UseTmpWorkspace {
		tmpVal = "y"
	}
	a.cardForm.Inputs[4].SetValue(tmpVal)
	a.setSelectedSkills(selected.SkillIDs)
	a.setSelectedPersonas(selected.PersonaIDs)
	a.setSelectedMCPs(selected.MCPIDs)
	a.cardForm.EditingID = selected.ID
	a.cardForm.Content = selected.AgentsMD
	a.cardForm.Path = selected.AgentsMDPath
	a.cardForm.PreferPath = selected.AgentsMDPreferPath
	a.cardForm.PathInput.SetValue(selected.AgentsMDPath)
	a.cardForm.FocusIndex = 0
	a.setCardFocus(a.cardForm.FocusIndex)
	a.mode = viewCardForm
	return a, nil
}

func (a *App) handleDashboardEnter() (tea.Model, tea.Cmd) {
	if a.sessionView.FocusedPane == paneSessions {
		return a, a.activateSelectedSession()
	}
	selected, ok := a.selectedCard()
	if !ok {
		return a, nil
	}
	skills := a.skillsForCard(selected)
	personas := a.personasForCard(selected)
	mcps := a.mcpsForCard(selected)

	a.naming.Card = selected
	a.naming.Input.SetValue("")
	a.naming.LaunchDone = false
	a.naming.LaunchErr = nil
	a.naming.SessionID = ""
	a.naming.Submitted = false
	a.naming.Input.Focus()
	a.mode = viewNaming

	return a, a.launchSessionAsync(selected, skills, personas, mcps)
}

func (a *App) launchSessionAsync(card config.IDCard, skills []config.Skill, personas []config.Persona, mcps []config.MCPServer) tea.Cmd {
	return func() tea.Msg {
		id, err := a.session.Launch(card, skills, personas, mcps)
		return launchFinishedMsg{sessionID: id, err: err}
	}
}

func (a *App) handleDashboardDelete() (tea.Model, tea.Cmd) {
	if a.sessionView.FocusedPane != paneSessions {
		return a, nil
	}
	sel := a.sessionsList.Index()
	if sel < 0 || sel >= len(a.sessions) {
		return a, nil
	}
	s := a.sessions[sel]
	if s.UseTmpWorkspace {
		a.sessionView.PendingDeleteID = s.ID
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
	if a.sessionView.FocusedPane != paneSessions {
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
	a.sessionView.SwitchID = s.ID
	a.switchList.SetItems(toCardItems(a.cards, a.skills, a.personas))
	a.switchList.Select(0)
	a.mode = viewSwitchAgent
	return a, nil
}

func (a *App) toggleDashboardPane() {
	if a.sessionView.FocusedPane == paneCards && len(a.sessions) > 0 {
		a.sessionView.FocusedPane = paneSessions
		return
	}
	a.sessionView.FocusedPane = paneCards
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
		personas := a.personasForCard(card)
		mcps := a.mcpsForCard(card)
		if err := a.session.Relaunch(s.ID, card, skills, personas, mcps); err != nil {
			a.err = fmt.Errorf("relaunch failed: %w", err)
			return nil
		}
		// Refresh locally to reflect status change immediately
		a.sessions = a.session.List()
		a.refreshSessionList()
	}

	a.sessionView.ActiveID = s.ID
	a.mode = viewDashboard
	return a.attachSession(a.sessionView.ActiveID)
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
