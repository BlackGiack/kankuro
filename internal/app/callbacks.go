package app

import (
	"errors"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/BlackGiack/kankuro/internal/ui/components"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/views"
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

		OnRelaunch: c.requestRelaunch,

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
				err := c.actions.service.RemoveSession(sessionID)
				if err != nil {
					return messages.ErrorMsg{Err: err}
				}
				return messages.DataRefreshRequestedMsg{}
			}, nil
		},

		OnDeleteCard: func(cardID string) (tea.Cmd, error) {
			return c.actions.deleteCard(cardID), nil
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

		OnLaunchIDE:      c.launchIDE,
		OnLaunchTerminal: c.launchTerminal,
	}
}

func (c *Coordinator) launchCard(card config.IDCard) tea.Cmd {
	launcher, ok := c.views[ViewSessionLauncher].(*views.SessionLauncher)
	if !ok {
		return nil
	}
	launcher.Reset()
	launcher.SetCard(card)

	// The view reset above is coordinator-goroutine work and has already
	// happened; the blocking launch is sessionActions' closure.
	return tea.Batch(
		c.actions.launch(card),
		func() tea.Msg {
			return messages.NavigateMsg{View: messages.ViewSessionLauncher}
		},
	)
}

// launchIDE snapshots the IDE command here, on the coordinator goroutine:
// c.settings is coordinator-goroutine state (SettingsSavedMsg writes it), so
// the closure sessionActions builds must receive it by value.
func (c *Coordinator) launchIDE(sessionID string) tea.Cmd {
	return c.actions.launchIDE(sessionID, c.settings.IDECommand)
}

func (c *Coordinator) launchTerminal(sessionID string) tea.Cmd {
	return c.actions.launchTerminal(sessionID, c.settings.TerminalCommand)
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
	if _, ok := errors.AsType[*config.SyncFailedError](err); ok {
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
		// OnSwitch must not switch synchronously: this closure runs from
		// inside AgentSwitchView.Update, on the main Bubble Tea goroutine,
		// and the switch syncs driver files to disk and spawns a PTY
		// process - exactly the kind of blocking work that must run inside
		// the returned tea.Cmd, off that goroutine. sessionActions.switchAgent
		// builds that Cmd (it has no Coordinator pointer, so it cannot touch
		// coordinator state). Completion is reported via AgentSwitchedMsg
		// and handled in handleSessionMessages, which does the refresh +
		// attach + warning dance (moved there because it needs
		// c.currentView, which this closure cannot safely touch either - see
		// that handler's comment for the navigation hazard).
		OnSwitch: c.requestSwitch,
		OnCancel: func() tea.Cmd {
			return func() tea.Msg {
				return messages.NavigateMsg{View: messages.ViewDashboard}
			}
		},
	}
}

func (c *Coordinator) requestRelaunch(sessionID string) tea.Cmd {
	snapshot, found := c.session.GetSession(sessionID)
	if found && len(snapshot.Conversations[snapshot.CardID]) > 0 {
		driverName := ""
		if dashboard, ok := c.views[ViewDashboard].(*views.Dashboard); ok {
			for _, card := range dashboard.Cards() {
				if card.ID == snapshot.CardID {
					driverName = card.Driver
					break
				}
			}
		}
		return func() tea.Msg {
			return messages.ConversationChoiceRequestedMsg{
				SessionID:  sessionID,
				CardID:     snapshot.CardID,
				Driver:     driverName,
				ReturnView: messages.ViewDashboard,
			}
		}
	}
	return c.actions.relaunch(sessionID)
}

func (c *Coordinator) requestSwitch(sessionID, cardID, driverName string) tea.Cmd {
	snapshot, found := c.session.GetSession(sessionID)
	if found && len(snapshot.Conversations[cardID]) > 0 {
		return func() tea.Msg {
			return messages.ConversationChoiceRequestedMsg{
				SessionID:  sessionID,
				CardID:     cardID,
				Driver:     driverName,
				ReturnView: messages.ViewAgentSwitch,
			}
		}
	}
	return c.actions.switchAgent(sessionID, cardID)
}

func (c *Coordinator) conversationChoiceCallbacks() views.ConversationChoiceCallbacks {
	return views.ConversationChoiceCallbacks{
		OnSelect: func(sessionID, cardID string, choice session.ConversationChoice) tea.Cmd {
			if cardID == "" {
				return c.actions.relaunchWithChoice(sessionID, choice)
			}
			return c.actions.switchAgentWithChoice(sessionID, cardID, choice)
		},
		OnCancel: func(returnView messages.ViewType) tea.Cmd {
			return func() tea.Msg { return messages.NavigateMsg{View: returnView} }
		},
	}
}
