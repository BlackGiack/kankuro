package app

import (
	"errors"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"kankuro/internal/config"
	"kankuro/internal/session"
	"kankuro/internal/ui/components"
	"kankuro/internal/ui/messages"
	"kankuro/internal/ui/views"
)

func (c *Coordinator) dashboardCallbacks() views.DashboardCallbacks {
	return views.DashboardCallbacks{
		OnEditCard: func(card config.IDCard) tea.Cmd {
			editor, ok := c.views[ViewCardEditor].(*views.CardEditor)
			if !ok {
				return nil
			}
			editor.SetCard(card)
			return func() tea.Msg {
				return messages.NavigateMsg{View: messages.ViewCardEditor}
			}
		},

		OnLaunchCard: c.launchCard,

		OnAttach: func(sessionID string) tea.Cmd {
			return c.attachSession(sessionID)
		},

		OnRelaunch: c.relaunchSession,

		OnKill: func(sessionID string) (tea.Cmd, error) {
			err := c.session.Kill(sessionID)
			return c.refreshData(), err
		},

		OnRemove: func(sessionID string) (tea.Cmd, error) {
			err := c.session.Remove(sessionID)
			return c.refreshData(), err
		},

		OnSwitch: c.switchAgent,

		OnDeleteCard: func(cardID string) (tea.Cmd, error) {
			err := c.repo.DeleteIDCard(cardID)
			return c.refreshData(), err
		},

		OnOpenMCP: func() tea.Cmd {
			return func() tea.Msg {
				return messages.NavigateMsg{View: messages.ViewMCPForm}
			}
		},

		OnOpenWorkspaces: func() tea.Cmd {
			return func() tea.Msg {
				return messages.NavigateMsg{View: messages.ViewSourceManagerWorkspaces}
			}
		},

		OnOpenIDE: func() tea.Cmd {
			// Set the IDE config on the view
			ideConfig, ok := c.views[ViewType(messages.ViewIDEConfig)].(*views.IDEConfig)
			if ok {
				ideConfig.SetConfig(c.ideConfig)
			}
			return func() tea.Msg {
				return messages.NavigateMsg{View: messages.ViewIDEConfig}
			}
		},

		OnOpenCardEditor: func() tea.Cmd {
			editor, ok := c.views[ViewCardEditor].(*views.CardEditor)
			if !ok {
				return nil
			}
			editor.Reset()
			return func() tea.Msg {
				return messages.NavigateMsg{View: messages.ViewCardEditor}
			}
		},

		OnLaunchIDE: c.launchIDE,
	}
}

func (c *Coordinator) launchCard(card config.IDCard) tea.Cmd {
	launcher, ok := c.views[ViewSessionLauncher].(*views.SessionLauncher)
	if !ok {
		return nil
	}
	launcher.Reset()
	launcher.SetCard(card)

	// Launch session async while navigating to the naming view
	skills := c.skillsForCard(card)
	personas := c.personasForCard(card)
	mcps := c.mcpsForCard(card)

	return tea.Batch(
		func() tea.Msg {
			id, err := c.session.Launch(card, skills, personas, mcps)
			return messages.SessionLaunchCompletedMsg{SessionID: id, Err: err}
		},
		func() tea.Msg {
			return messages.NavigateMsg{View: messages.ViewSessionLauncher}
		},
	)
}

func (c *Coordinator) relaunchSession(sessionID string) tea.Cmd {
	return func() tea.Msg {
		sessions := c.session.List()
		var sess *session.Snapshot
		for _, s := range sessions {
			if s.ID == sessionID {
				sess = &s
				break
			}
		}
		if sess == nil {
			return messages.ErrorMsg{Err: fmt.Errorf("session not found: %s", sessionID)}
		}
		card, err := c.getCardByID(sess.CardID)
		if err != nil {
			return messages.ErrorMsg{Err: err}
		}
		skills := c.skillsForCard(card)
		personas := c.personasForCard(card)
		mcps := c.mcpsForCard(card)
		relaunchErr := c.session.Relaunch(sessionID, card, skills, personas, mcps)
		if relaunchErr != nil {
			return messages.ErrorMsg{Err: relaunchErr}
		}
		return messages.SessionAttachRequestedMsg{SessionID: sessionID}
	}
}

func (c *Coordinator) switchAgent(sessionID, newCardID string) (tea.Cmd, error) {
	card, err := c.getCardByID(newCardID)
	if err != nil {
		return nil, err
	}
	skills := c.skillsForCard(card)
	personas := c.personasForCard(card)
	mcps := c.mcpsForCard(card)
	err = c.session.SwitchAgent(sessionID, card, skills, personas, mcps)
	return c.refreshData(), err
}

func (c *Coordinator) launchIDE(sessionID string) error {
	// Get the session to find its workspace path
	sessions := c.session.List()
	for _, s := range sessions {
		if s.ID == sessionID {
			// Get the workspace path from the session
			workspacePath := s.WorkingDir
			if workspacePath == "" {
				return errors.New("session has no workspace path")
			}

			// Use the stored IDE config
			if c.ideConfig.Command == "" {
				return session.ErrIDENotConfigured
			}
			launcher := session.NewIDELauncher(c.ideConfig.Command)
			return launcher.Launch(workspacePath)
		}
	}
	return fmt.Errorf("session not found: %s", sessionID)
}

func (c *Coordinator) cardEditorCallbacks() views.CardEditorCallbacks {
	return views.CardEditorCallbacks{
		OnSave: func(card config.IDCard) (config.IDCard, tea.Cmd, error) {
			var saved config.IDCard
			var err error
			if card.ID == "" {
				saved, err = c.repo.CreateIDCard(card)
			} else {
				saved, err = c.repo.UpdateIDCard(card)
			}
			if err != nil {
				return saved, nil, err
			}
			return saved, func() tea.Msg {
				return messages.NavigateMsg{View: messages.ViewDashboard}
			}, nil
		},

		OnCancel: func() tea.Cmd {
			return func() tea.Msg {
				return messages.NavigateMsg{View: messages.ViewDashboard}
			}
		},

		OnOpenSources: func(sourceType string) tea.Cmd {
			var vt messages.ViewType
			switch sourceType {
			case "skills":
				vt = messages.ViewSourceManagerSkills
			case "personas":
				vt = messages.ViewSourceManagerPersonas
			default:
				return nil
			}

			return func() tea.Msg {
				return messages.NavigateMsg{View: vt}
			}
		},

		OnOpenMCPForm: func() tea.Cmd {
			return func() tea.Msg {
				return messages.NavigateMsg{View: messages.ViewMCPForm}
			}
		},

		OnOpenWorkspaces: func() tea.Cmd {
			return func() tea.Msg {
				return messages.NavigateMsg{View: messages.ViewSourceManagerWorkspaces}
			}
		},
	}
}

func (c *Coordinator) sourceManagerCallbacks(sourceType views.SourceType) views.SourceManagerCallbacks {
	return views.SourceManagerCallbacks{
		LoadSources: func() []components.SourceItem {
			return c.loadSources(sourceType)
		},

		SaveSimpleSource: func(path string, editingIndex int) (tea.Cmd, error) {
			return c.saveSimpleSource(sourceType, path, editingIndex)
		},

		SaveWorkspaceSource: func(source config.WorkspaceSource, editingIndex int) (tea.Cmd, error) {
			sources, _ := c.repo.ListWorkspaceSources()
			if editingIndex >= 0 && editingIndex < len(sources) {
				sources[editingIndex] = source
			} else {
				source.ID = fmt.Sprintf("workspace-%d", time.Now().UnixNano())
				sources = append(sources, source)
			}
			err := c.repo.SaveWorkspaceSources(sources)
			return c.refreshData(), err
		},

		DeleteSource: func(index int) (tea.Cmd, error) {
			return c.deleteSource(sourceType, index)
		},

		Refresh: func() tea.Cmd {
			switch sourceType {
			case views.SourceTypeWorkspaces:
				_ = c.repo.SyncWorkspacesFromSources()
			case views.SourceTypeSkills, views.SourceTypePersonas:
				// No specific refresh needed for skills and personas
			}
			return c.refreshData()
		},
	}
}

func (c *Coordinator) loadSources(sourceType views.SourceType) []components.SourceItem {
	switch sourceType {
	case views.SourceTypeSkills:
		sources, _ := c.repo.ListSkillSources()
		return components.ToSimpleItems(sources)
	case views.SourceTypePersonas:
		sources, _ := c.repo.ListPersonaSources()
		return components.ToSimpleItems(sources)
	case views.SourceTypeWorkspaces:
		sources, _ := c.repo.ListWorkspaceSources()
		return components.ToWorkspaceItems(sources)
	}
	return nil
}

func (c *Coordinator) saveSimpleSource(sourceType views.SourceType, path string, editingIndex int) (tea.Cmd, error) {
	switch sourceType {
	case views.SourceTypeSkills:
		sources, _ := c.repo.ListSkillSources()
		if editingIndex >= 0 && editingIndex < len(sources) {
			sources[editingIndex] = path
		} else {
			sources = append(sources, path)
		}
		err := c.repo.SaveSkillSources(sources)
		if err == nil {
			_ = c.repo.SyncSkillsFromSources()
		}
		return c.refreshData(), err
	case views.SourceTypePersonas:
		sources, _ := c.repo.ListPersonaSources()
		if editingIndex >= 0 && editingIndex < len(sources) {
			sources[editingIndex] = path
		} else {
			sources = append(sources, path)
		}
		err := c.repo.SavePersonaSources(sources)
		if err == nil {
			_ = c.repo.SyncPersonasFromSources()
		}
		return c.refreshData(), err
	case views.SourceTypeWorkspaces:
		return nil, errors.New("unsupported source type for simple save")
	}
	return nil, errors.New("unknown source type")
}

func (c *Coordinator) deleteSource(sourceType views.SourceType, index int) (tea.Cmd, error) {
	switch sourceType {
	case views.SourceTypeSkills:
		sources, _ := c.repo.ListSkillSources()
		if index >= 0 && index < len(sources) {
			sources = append(sources[:index], sources[index+1:]...)
			err := c.repo.SaveSkillSources(sources)
			if err == nil {
				_ = c.repo.SyncSkillsFromSources()
			}
			return c.refreshData(), err
		}
	case views.SourceTypePersonas:
		sources, _ := c.repo.ListPersonaSources()
		if index >= 0 && index < len(sources) {
			sources = append(sources[:index], sources[index+1:]...)
			err := c.repo.SavePersonaSources(sources)
			if err == nil {
				_ = c.repo.SyncPersonasFromSources()
			}
			return c.refreshData(), err
		}
	case views.SourceTypeWorkspaces:
		sources, _ := c.repo.ListWorkspaceSources()
		if index >= 0 && index < len(sources) {
			sources = append(sources[:index], sources[index+1:]...)
			err := c.repo.SaveWorkspaceSources(sources)
			return c.refreshData(), err
		}
	}
	return nil, errors.New("unknown source type or invalid index")
}

func (c *Coordinator) ideCallbacks() views.IDEConfigCallbacks {
	return views.IDEConfigCallbacks{
		OnSave: func(command string) (tea.Cmd, error) {
			cfg := config.IDEConfig{Command: command}
			err := c.repo.SaveIDEConfig(cfg)
			if err == nil {
				c.ideConfig = cfg
			}
			return c.refreshData(), err
		},

		OnLaunch: func(sessionPath string) error {
			if c.ideConfig.Command == "" {
				return session.ErrIDENotConfigured
			}
			launcher := session.NewIDELauncher(c.ideConfig.Command)
			return launcher.Launch(sessionPath)
		},
	}
}

func (c *Coordinator) mcpCallbacks() views.MCPFormCallbacks {
	return views.MCPFormCallbacks{
		OnSave: func(mcp config.MCPServer) (config.MCPServer, tea.Cmd, error) {
			saved, err := c.repo.UpsertMCP(mcp)
			if err != nil {
				return saved, nil, err
			}
			return saved, func() tea.Msg {
				return messages.MCPUpdatedMsg{MCP: saved}
			}, nil
		},

		OnDelete: func(mcpID string) error {
			return c.repo.DeleteMCP(mcpID)
		},
	}
}

func (c *Coordinator) sessionLauncherCallbacks() views.SessionLauncherCallbacks {
	return views.SessionLauncherCallbacks{
		OnSetName: func(sessionID, name string) error {
			return c.session.SetName(sessionID, name)
		},

		OnAttach: func(sessionID string) tea.Cmd {
			return c.attachSession(sessionID)
		},
	}
}

func (c *Coordinator) agentSwitchCallbacks() views.AgentSwitchCallbacks {
	return views.AgentSwitchCallbacks{
		OnSwitch: func(sessionID, cardID string) tea.Cmd {
			refreshCmd, err := c.switchAgent(sessionID, cardID)
			if err != nil {
				return func() tea.Msg { return messages.ErrorMsg{Err: err} }
			}
			// After switch, attach to the session (it was restarted by switchAgent)
			return tea.Batch(
				refreshCmd,
				c.attachSession(sessionID),
				func() tea.Msg {
					return messages.NavigateMsg{View: messages.ViewDashboard}
				},
			)
		},
		OnCancel: func() tea.Cmd {
			return func() tea.Msg {
				return messages.NavigateMsg{View: messages.ViewDashboard}
			}
		},
	}
}

func (c *Coordinator) getCardByID(cardID string) (config.IDCard, error) {
	cards, _ := c.repo.ListIDCards()
	for _, card := range cards {
		if card.ID == cardID {
			return card, nil
		}
	}
	return config.IDCard{}, fmt.Errorf("card not found: %s", cardID)
}

func (c *Coordinator) skillsForCard(card config.IDCard) []config.Skill {
	var result []config.Skill
	allSkills, _ := c.repo.ListSkills()
	for _, skill := range allSkills {
		for _, id := range card.SkillIDs {
			if skill.ID == id {
				result = append(result, skill)
			}
		}
	}
	return result
}

func (c *Coordinator) personasForCard(card config.IDCard) []config.Persona {
	var result []config.Persona
	allPersonas, _ := c.repo.ListPersonas()
	for _, persona := range allPersonas {
		for _, id := range card.PersonaIDs {
			if persona.ID == id {
				result = append(result, persona)
			}
		}
	}
	return result
}

func (c *Coordinator) mcpsForCard(card config.IDCard) []config.MCPServer {
	var result []config.MCPServer
	allMCPs, _ := c.repo.ListMCPs()
	for _, mcp := range allMCPs {
		for _, id := range card.MCPIDs {
			if mcp.ID == id {
				result = append(result, mcp)
			}
		}
	}
	return result
}
