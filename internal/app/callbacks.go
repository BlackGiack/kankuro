package app

import (
	"errors"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

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
			return func() tea.Msg {
				err := c.session.Kill(sessionID)
				if err != nil {
					return messages.ErrorMsg{Err: err}
				}
				return messages.DataRefreshRequestedMsg{}
			}, nil
		},

		OnRemove: func(sessionID string) (tea.Cmd, error) {
			return func() tea.Msg {
				err := c.session.Remove(sessionID)
				if err != nil {
					return messages.ErrorMsg{Err: err}
				}
				return messages.DataRefreshRequestedMsg{}
			}, nil
		},

		OnDeleteCard: func(cardID string) (tea.Cmd, error) {
			return func() tea.Msg {
				return c.deleteCard(cardID)
			}, nil
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

		OnOpenSettings: func() tea.Cmd {
			settingsView, ok := c.views[ViewType(messages.ViewSettings)].(*views.SettingsView)
			if ok {
				settingsView.SetSettings(c.settings)
			}
			return func() tea.Msg {
				return messages.NavigateMsg{View: messages.ViewSettings}
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

// deleteCard removes an ID card, refusing when a persisted session still
// references it. DeleteIDCard itself has no reference guard (internal/config
// cannot see sessions), so the check lives here where the session manager is
// in hand: a card a session still points at cannot be deleted, or that
// session's relaunch/switch would break with a bare "card not found".
func (c *Coordinator) deleteCard(cardID string) tea.Msg {
	var referencedBy []string
	for _, s := range c.session.List() {
		if s.CardID == cardID {
			referencedBy = append(referencedBy, s.Name)
		}
	}
	if len(referencedBy) > 0 {
		return messages.ErrorMsg{Err: fmt.Errorf(
			"cannot delete card: referenced by session(s) %s",
			strings.Join(referencedBy, ", "),
		)}
	}
	if err := c.repo.DeleteIDCard(cardID); err != nil {
		return messages.ErrorMsg{Err: err}
	}
	return messages.DataRefreshRequestedMsg{}
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
			if !errors.Is(relaunchErr, session.ErrPersistFailed) {
				return messages.ErrorMsg{Err: relaunchErr}
			}
			// The process did restart; only its on-disk record failed to
			// save. Still attach, but let the user know.
			return messages.SessionAttachRequestedMsg{SessionID: sessionID, Warning: relaunchErr}
		}
		return messages.SessionAttachRequestedMsg{SessionID: sessionID}
	}
}

// switchAgentAsync performs the actual (blocking) agent switch: it syncs
// driver files to disk and spawns a new PTY process. Callers must run it
// inside a tea.Cmd closure, never synchronously inside Update - see
// agentSwitchCallbacks' OnSwitch, the only caller. A synchronous wrapper
// used to exist here for DashboardCallbacks.OnSwitch, which was never
// invoked by the real Dashboard view (Dashboard.handleSwitch emits
// AgentSwitchRequestedMsg instead); both were deleted together (review_dp_4_2.md
// finding 4) rather than left as a dead, reachable-by-one-wiring-change copy
// of the blocking-work-inside-Update bug step 3 removed.
func (c *Coordinator) switchAgentAsync(sessionID, newCardID string) error {
	card, err := c.getCardByID(newCardID)
	if err != nil {
		return err
	}
	skills := c.skillsForCard(card)
	personas := c.personasForCard(card)
	mcps := c.mcpsForCard(card)
	return c.session.SwitchAgent(sessionID, card, skills, personas, mcps)
}

// launchIDE spawns the configured editor for a session's workspace. It
// returns a tea.Cmd so the (potentially slow) process spawn runs off the UI
// goroutine; completion is reported via messages.IDELaunchedMsg rather than a
// synchronous error return.
func (c *Coordinator) launchIDE(sessionID string) tea.Cmd {
	return func() tea.Msg {
		sessions := c.session.List()
		for _, s := range sessions {
			if s.ID != sessionID {
				continue
			}
			workspacePath := s.WorkingDir
			if workspacePath == "" {
				return messages.IDELaunchedMsg{Err: errors.New("session has no workspace path")}
			}
			if c.settings.IDECommand == "" {
				return messages.IDELaunchedMsg{Err: session.ErrIDENotConfigured}
			}
			launcher := session.NewIDELauncher(c.settings.IDECommand)
			return messages.IDELaunchedMsg{Err: launcher.Launch(workspacePath)}
		}
		return messages.IDELaunchedMsg{Err: fmt.Errorf("session not found: %s", sessionID)}
	}
}

func (c *Coordinator) cardEditorCallbacks() views.CardEditorCallbacks {
	return views.CardEditorCallbacks{
		OnSave: func(card config.IDCard) tea.Cmd {
			return func() tea.Msg {
				var saved config.IDCard
				var err error
				if card.ID == "" {
					saved, err = c.repo.CreateIDCard(card)
				} else {
					saved, err = c.repo.UpdateIDCard(card)
				}
				return messages.CardSavedMsg{Card: saved, Err: err}
			}
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

		OnOpenPuppetSources: func() tea.Cmd {
			return func() tea.Msg {
				return messages.NavigateMsg{View: messages.ViewSourceManagerPuppets}
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
			return func() tea.Msg {
				err := c.repo.UpdateWorkspaceSources(func(sources []config.WorkspaceSource) ([]config.WorkspaceSource, error) {
					if editingIndex >= 0 && editingIndex < len(sources) {
						sources[editingIndex] = source
					} else {
						// ID is assigned inside UpdateWorkspaceSources so
						// there is one scheme for workspace-source
						// identifiers.
						sources = append(sources, source)
					}
					return sources, nil
				})
				return c.sourceUpdateMsg(err)
			}, nil
		},

		DeleteSource: func(index int) (tea.Cmd, error) {
			return c.deleteSource(sourceType, index)
		},

		Refresh: func() tea.Cmd {
			return func() tea.Msg {
				var err error
				switch sourceType {
				case views.SourceTypeWorkspaces:
					err = c.repo.SyncWorkspacesFromSources()
				case views.SourceTypeSkills, views.SourceTypePersonas:
					// No specific refresh needed for skills and personas
				case views.SourceTypePuppets:
					err = c.repo.SyncPuppetsFromSources()
				}
				if err != nil {
					return tea.Batch(refreshRequestedCmd(), errorCmd(err))()
				}
				return messages.DataRefreshRequestedMsg{}
			}
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
	case views.SourceTypePuppets:
		sources, _ := c.repo.ListPuppetSources()
		return components.ToSimpleItems(sources)
	case views.SourceTypeWorkspaces:
		sources, _ := c.repo.ListWorkspaceSources()
		return components.ToWorkspaceItems(sources)
	}
	return nil
}

func (c *Coordinator) saveSimpleSource(sourceType views.SourceType, path string, editingIndex int) (tea.Cmd, error) {
	// The error return is always nil here (an edit/append can't fail), but
	// the signature must match UpdateXxxSources' func([]string) ([]string, error).
	mutate := func(sources []string) ([]string, error) { //nolint:unparam // must match UpdateXxxSources signature
		if editingIndex >= 0 && editingIndex < len(sources) {
			sources[editingIndex] = path
		} else {
			sources = append(sources, path)
		}
		return sources, nil
	}

	var updateFn func(func([]string) ([]string, error)) error
	switch sourceType {
	case views.SourceTypeSkills:
		updateFn = c.repo.UpdateSkillSources
	case views.SourceTypePersonas:
		updateFn = c.repo.UpdatePersonaSources
	case views.SourceTypePuppets:
		updateFn = c.repo.UpdatePuppetSources
	case views.SourceTypeWorkspaces:
		return func() tea.Msg {
			return messages.ErrorMsg{Err: errors.New("unsupported source type for simple save")}
		}, nil
	default:
		return func() tea.Msg {
			return messages.ErrorMsg{Err: errors.New("unknown source type")}
		}, nil
	}

	return func() tea.Msg {
		return c.sourceUpdateMsg(updateFn(mutate))
	}, nil
}

func (c *Coordinator) deleteSource(sourceType views.SourceType, index int) (tea.Cmd, error) {
	deleteAt := func(sources []string) ([]string, error) {
		if index < 0 || index >= len(sources) {
			return nil, errors.New("invalid index")
		}
		return append(sources[:index], sources[index+1:]...), nil
	}

	var updateFn func(func([]string) ([]string, error)) error
	switch sourceType {
	case views.SourceTypeSkills:
		updateFn = c.repo.UpdateSkillSources
	case views.SourceTypePersonas:
		updateFn = c.repo.UpdatePersonaSources
	case views.SourceTypePuppets:
		updateFn = c.repo.UpdatePuppetSources
	case views.SourceTypeWorkspaces:
		return func() tea.Msg {
			err := c.repo.UpdateWorkspaceSources(func(sources []config.WorkspaceSource) ([]config.WorkspaceSource, error) {
				if index < 0 || index >= len(sources) {
					return nil, errors.New("invalid index")
				}
				return append(sources[:index], sources[index+1:]...), nil
			})
			return c.sourceUpdateMsg(err)
		}, nil
	default:
		return nil, errors.New("unknown source type")
	}

	return func() tea.Msg {
		return c.sourceUpdateMsg(updateFn(deleteAt))
	}, nil
}

// sourceUpdateMsg turns the result of a repository UpdateXxxSources call into
// the message the UI should react to: nil is a plain data refresh, a
// *SyncFailedError still refreshes (the source list itself did save) but
// also surfaces a warning that the derived registry is now stale, and any
// other error is reported without refreshing (nothing was persisted).
func (c *Coordinator) sourceUpdateMsg(err error) tea.Msg {
	if err == nil {
		return messages.DataRefreshRequestedMsg{}
	}
	var syncErr *config.SyncFailedError
	if errors.As(err, &syncErr) {
		return tea.Batch(refreshRequestedCmd(), errorCmd(err))()
	}
	return messages.ErrorMsg{Err: err}
}

// refreshRequestedCmd asks the coordinator to refresh on its own goroutine.
// It exists so callers already inside a tea.Cmd closure can batch a refresh
// alongside another message without building refreshData() there - see
// messages.DataRefreshRequestedMsg.
func refreshRequestedCmd() tea.Cmd {
	return func() tea.Msg { return messages.DataRefreshRequestedMsg{} }
}

// errorCmd wraps err in a tea.Cmd that reports it via messages.ErrorMsg.
func errorCmd(err error) tea.Cmd {
	return func() tea.Msg { return messages.ErrorMsg{Err: err} }
}

func (c *Coordinator) settingsCallbacks() views.SettingsCallbacks {
	return views.SettingsCallbacks{
		OnSave: func(settings config.AppSettings) tea.Cmd {
			return func() tea.Msg {
				err := c.repo.SaveAppSettings(settings)
				return messages.SettingsSavedMsg{Settings: settings, Err: err}
			}
		},
		OnOpenPuppetSources: func() tea.Cmd {
			return func() tea.Msg {
				return messages.NavigateMsg{View: messages.ViewSourceManagerPuppets}
			}
		},
	}
}

func (c *Coordinator) mcpCallbacks() views.MCPFormCallbacks {
	return views.MCPFormCallbacks{
		OnSave: func(mcp config.MCPServer) tea.Cmd {
			return func() tea.Msg {
				saved, err := c.repo.UpsertMCP(mcp)
				return messages.MCPSavedMsg{MCP: saved, Err: err}
			}
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
		// OnSwitch must not call switchAgentAsync synchronously: this
		// closure runs from inside AgentSwitchView.Update, on the main
		// Bubble Tea goroutine, and switchAgentAsync syncs driver files to
		// disk and spawns a PTY process - exactly the kind of blocking work
		// that must run inside the returned tea.Cmd, off that goroutine.
		// Completion is reported via AgentSwitchedMsg and handled in
		// handleSessionMessages, which does the refresh + attach + warning
		// dance below (moved there because it needs c.currentView, which
		// this closure cannot safely touch either - see that handler's
		// comment for the navigation hazard).
		OnSwitch: func(sessionID, cardID string) tea.Cmd {
			return func() tea.Msg {
				err := c.switchAgentAsync(sessionID, cardID)
				return messages.AgentSwitchedMsg{SessionID: sessionID, Err: err}
			}
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
