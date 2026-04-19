package ui

import (
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"kankuro/internal/config"
	"kankuro/internal/session"
	"kankuro/internal/ui/components"
)

//nolint:funlen // Central event dispatcher for dashboard view.
func (a *App) updateDashboard(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		// Handle search input if search is active or '/' is pressed
		if cmd, handled := a.handleDashboardSearch(keyMsg); handled {
			return a, cmd
		}

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
		a.syncPreviewSelection()
		return a, cmd
	}
	var cmd tea.Cmd
	a.cardsList, cmd = a.cardsList.Update(msg)
	return a, cmd
}

func (a *App) syncPreviewSelection() {
	// If there are no sessions at all, clear preview.
	if len(a.sessions) == 0 {
		a.sessionView.PreviewSessionID = ""
		a.sessionView.PreviewScrollback = nil
		return
	}

	// Check if current preview session still exists
	currentStillExists := false
	if a.sessionView.PreviewSessionID != "" {
		for _, s := range a.sessions {
			if s.ID == a.sessionView.PreviewSessionID {
				currentStillExists = true
				break
			}
		}
	}

	// If we are NOT in sessions pane AND current preview still exists, just refresh it.
	if a.sessionView.FocusedPane != paneSessions && currentStillExists {
		a.refreshPreviewScrollback()
		return
	}

	// Otherwise, try to follow selection.
	selectedItem := a.sessionsList.SelectedItem()
	if selectedItem == nil {
		a.sessionView.PreviewSessionID = ""
		a.sessionView.PreviewScrollback = nil
		return
	}

	item, ok := selectedItem.(sessionListItem)
	if !ok {
		a.sessionView.PreviewSessionID = ""
		a.sessionView.PreviewScrollback = nil
		return
	}

	if a.sessionView.PreviewSessionID != item.id {
		a.sessionView.PreviewSessionID = item.id
	}
	a.refreshPreviewScrollback()
}

// ===== Integration with new SessionPreview component =====

func (a *App) newSessionPreviewComponent() *components.SessionPreview {
	p := components.NewSessionPreview()
	p.SetSize(a.sessionView.PreviewWidth, a.sessionsList.Height())
	return p
}

// syncPreviewToComponent pushes preview state to the component.
// Called after syncPreviewSelection updates sessionView state.
func (a *App) syncPreviewToComponent(p *components.SessionPreview) {
	if a.sessionView.PreviewSessionID == "" {
		p.Clear()
		return
	}

	// Find session snapshot
	var snap *session.Snapshot
	for i := range a.sessions {
		if a.sessions[i].ID == a.sessionView.PreviewSessionID {
			snap = &a.sessions[i]
			break
		}
	}

	if snap == nil {
		p.Clear()
		return
	}

	// Find card
	var card *config.IDCard
	for i := range a.cards {
		if a.cards[i].ID == snap.CardID {
			card = &a.cards[i]
			break
		}
	}

	p.SetSession(snap, card)
	p.SetScrollback(a.sessionView.PreviewScrollback)
}

func (a *App) handleDashboardSearch(keyMsg tea.KeyMsg) (tea.Cmd, bool) {
	switch keyMsg.String() {
	case keyEsc:
		if a.sessionView.IsSearching {
			a.sessionView.CardsSearchQuery = ""
			a.sessionView.SessionsSearchQuery = ""
			a.sessionView.IsSearching = false
			a.refreshLists()
			return nil, true
		}
	case "/":
		if !a.sessionView.IsSearching {
			a.sessionView.IsSearching = true
			a.refreshLists()
			return nil, true
		}
	case keyEnter:
		if a.sessionView.IsSearching {
			a.sessionView.IsSearching = false
			a.refreshLists()
			return nil, true
		}
	case keyBackspace:
		if a.sessionView.IsSearching {
			return a.handleDashboardSearchBackspace()
		}
	}

	if len(keyMsg.Runes) > 0 && a.sessionView.IsSearching {
		return a.handleDashboardSearchRunes(keyMsg)
	}

	return nil, false
}

func (a *App) handleDashboardSearchBackspace() (tea.Cmd, bool) {
	isCards := a.sessionView.FocusedPane == paneCards
	isSessions := a.sessionView.FocusedPane == paneSessions

	if isCards && len(a.sessionView.CardsSearchQuery) > 0 {
		a.sessionView.CardsSearchQuery = a.sessionView.CardsSearchQuery[:len(a.sessionView.CardsSearchQuery)-1]
		a.refreshCardList()
		return nil, true
	}
	if isSessions && len(a.sessionView.SessionsSearchQuery) > 0 {
		a.sessionView.SessionsSearchQuery = a.sessionView.SessionsSearchQuery[:len(a.sessionView.SessionsSearchQuery)-1]
		a.refreshSessionList()
		return nil, true
	}
	return nil, false
}

func (a *App) handleDashboardSearchRunes(keyMsg tea.KeyMsg) (tea.Cmd, bool) {
	ch := keyMsg.Runes[0]
	if ch < printableCharMin || ch > printableCharMax {
		return nil, false
	}

	if a.sessionView.FocusedPane == paneCards {
		a.sessionView.CardsSearchQuery += string(ch)
		a.refreshCardList()
	} else {
		a.sessionView.SessionsSearchQuery += string(ch)
		a.refreshSessionList()
	}
	return nil, true
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
	if a.sessionView.FocusedPane == paneSessions {
		s, ok := a.selectedSession()
		if !ok {
			return a, nil
		}
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

	if a.sessionView.FocusedPane == paneCards {
		selected, ok := a.selectedCard()
		if !ok {
			return a, nil
		}
		a.sessionView.PendingDeleteCardID = selected.ID
		a.mode = viewDeleteConfirm
		return a, nil
	}

	return a, nil
}

func (a *App) handleDashboardSwitch() (tea.Model, tea.Cmd) {
	if a.sessionView.FocusedPane != paneSessions {
		return a, nil
	}
	s, ok := a.selectedSession()
	if !ok {
		return a, nil
	}
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
		a.syncPreviewSelection()
		return
	}
	a.sessionView.FocusedPane = paneCards
}

func (a *App) activateSelectedSession() tea.Cmd {
	s, ok := a.selectedSession()
	if !ok {
		return nil
	}

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
	s, ok := a.selectedSession()
	if !ok {
		return nil
	}
	return a.session.Kill(s.ID)
}
