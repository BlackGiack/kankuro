package app

import (
	"context"
	"errors"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/ui/components"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/models"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

func (c *Coordinator) navigationCommand(view messages.ViewType) tea.Cmd {
	dashboardID, generation := c.activeDashboard, c.viewGeneration
	return func() tea.Msg {
		return messages.ScopedNavigateMsg{View: view, DashboardID: dashboardID, Generation: generation}
	}
}

func (c *Coordinator) dashboardCallbacks() views.DashboardCallbacks {
	return views.DashboardCallbacks{
		OnBack: func() tea.Cmd {
			dashboardID, generation := c.activeDashboard, c.viewGeneration
			return func() tea.Msg {
				return messages.DashboardListRequestedMsg{DashboardID: dashboardID, Generation: generation}
			}
		},
		OnEditCard: func(cardID string) tea.Cmd {
			return c.requestCardEditor(cardID)
		},
		OnRenameSession: c.renameDashboardSession,

		OnLaunchCard: func(cardID string) tea.Cmd {
			c.launchRequest++
			c.launchCardID = cardID
			return dashboardCardCommand(c.reads, c.activeDashboard, cardID, c.viewGeneration, c.launchRequest, messages.DashboardCardLaunch)
		},

		OnAttach: func(sessionID string) tea.Cmd {
			return c.requestAttach(sessionID)
		},

		OnRelaunch: c.requestRelaunch,

		OnKill: func(sessionID string) (tea.Cmd, error) {
			c.mutationRequest++
			dashboardID, generation, request, service := c.activeDashboard, c.viewGeneration, c.mutationRequest, c.actions.port
			return func() tea.Msg {
				return messages.SessionStoppedMsg{
					DashboardID: dashboardID,
					Generation:  generation,
					SessionID:   sessionID,
					Request:     request,
					Result: service.StopDashboardSession(
						context.Background(),
						contracts.ScopedSessionRequest{DashboardID: dashboardID, SessionID: sessionID},
					),
				}
			}, nil
		},
		OnRemove: func(sessionID string) (tea.Cmd, error) {
			c.mutationRequest++
			dashboardID, generation, request, service := c.activeDashboard, c.viewGeneration, c.mutationRequest, c.actions.port
			return func() tea.Msg {
				return messages.SessionRemovedMsg{
					DashboardID: dashboardID,
					Generation:  generation,
					SessionID:   sessionID,
					Request:     request,
					Result: service.RemoveDashboardSession(
						context.Background(),
						contracts.ScopedSessionRequest{DashboardID: dashboardID, SessionID: sessionID},
					),
				}
			}, nil
		},

		OnDeleteCard: func(cardID string) (tea.Cmd, error) {
			return cardDeletionPreflightCommand(c.legacy.service, c.activeDashboard, cardID, c.viewGeneration), nil
		},
		OnConfirmCardDeletion: func(cardID, token string, foreign bool) tea.Cmd {
			return confirmCardDeletionCommand(c.legacy.service, c.activeDashboard, cardID, token, c.viewGeneration, foreign)
		},
		OnRetryCardDeletion: func(cardID string) tea.Cmd {
			return retryCardDeletionCommand(c.legacy.service, c.activeDashboard, cardID, c.viewGeneration)
		},

		OnOpenMCP: func() tea.Cmd {
			return c.navigationCommand(messages.ViewMCPManager)
		},

		OnOpenWorkspaces: func() tea.Cmd {
			return c.navigationCommand(messages.ViewSourceManagerWorkspaces)
		},

		OnOpenSettings: func() tea.Cmd {
			settingsView, ok := c.views[ViewType(messages.ViewSettings)].(*views.SettingsView)
			if ok {
				settingsView.SetSettings(c.settings)
			}
			return c.navigationCommand(messages.ViewSettings)
		},

		OnOpenCardEditor: func() tea.Cmd { return c.requestCardEditor("") },

		OnLaunchIDE:      c.launchIDE,
		OnLaunchTerminal: c.launchTerminal,
	}
}

func (c *Coordinator) renameDashboardSession(sessionID, name string) tea.Cmd {
	c.renameRequest++
	dashboardID, generation, request, service := c.activeDashboard, c.viewGeneration, c.renameRequest, c.actions.port
	return func() tea.Msg {
		result := service.RenameDashboardSession(
			context.Background(),
			contracts.RenameSessionRequest{DashboardID: dashboardID, SessionID: sessionID, Name: name},
		)
		return messages.DashboardSessionRenamedMsg{
			Request:     request,
			DashboardID: dashboardID,
			Generation:  generation,
			SessionID:   sessionID,
			Result:      result,
		}
	}
}

func (c *Coordinator) launchCard(card contracts.CardSummary, request uint64) tea.Cmd {
	launcher, ok := c.views[ViewSessionLauncher].(*views.SessionLauncher)
	if !ok {
		return nil
	}
	launcher.Reset()
	launcher.SetCard(card)

	// The view reset above is coordinator-goroutine work and has already
	// happened; the blocking launch is sessionActions' closure.
	launch := c.actions.launch(c.activeDashboard, c.viewGeneration, card.ID, request)

	// This runs in the checked card-loaded handler, outside a view Update.
	// Enter the launcher now so a queued generic navigation cannot reopen it
	// after the user leaves or starts another launch.
	return tea.Batch(c.navigateTo(ViewSessionLauncher), launch)
}

func (c *Coordinator) requestCardEditor(cardID string) tea.Cmd {
	c.editorEpoch++
	c.editorRequest++
	c.editorCardID, c.editorRevision, c.editorSaving, c.editorCompleted = cardID, "", false, false
	return cardEditorLoadCommand(c.editor, c.activeDashboard, cardID, c.viewGeneration, c.editorEpoch, c.editorRequest)
}

func cardEditorLoadCommand(service cardEditorPort, dashboardID, cardID string, generation, epoch, request uint64) tea.Cmd {
	return func() tea.Msg {
		var data contracts.CardEditorData
		var err error
		if cardID == "" {
			data, err = service.NewCardEditor(dashboardID)
		} else {
			data, err = service.LoadCardEditor(dashboardID, cardID)
		}
		return messages.CardEditorLoadedMsg{
			DashboardID: dashboardID,
			CardID:      cardID,
			Generation:  generation,
			Epoch:       epoch,
			Request:     request,
			Data:        data,
			Choices:     service.ReadCardEditorChoices(),
			Err:         err,
		}
	}
}

// Workspace commands resolve their settings inside the application operation.
func (c *Coordinator) launchIDE(sessionID string) tea.Cmd {
	c.workspaceRequest++
	return c.actions.launchWorkspace(c.activeDashboard, c.viewGeneration, c.workspaceRequest, sessionID, contracts.WorkspaceIDE)
}

func (c *Coordinator) launchTerminal(sessionID string) tea.Cmd {
	c.workspaceRequest++
	return c.actions.launchWorkspace(c.activeDashboard, c.viewGeneration, c.workspaceRequest, sessionID, contracts.WorkspaceTerminal)
}

func (c *Coordinator) editorNavigationCommand(view messages.ViewType) tea.Cmd {
	dashboardID, generation, epoch := c.activeDashboard, c.viewGeneration, c.editorEpoch
	return func() tea.Msg {
		return messages.EditorNavigateMsg{View: view, DashboardID: dashboardID, Generation: generation, Epoch: epoch}
	}
}

func (c *Coordinator) cardEditorCallbacks() views.CardEditorCallbacks {
	return views.CardEditorCallbacks{
		OnSave: func(draft models.CardEditorDraft) tea.Cmd {
			if c.editorSaving || c.editorCompleted {
				return nil
			}
			draft = models.CloneCardEditorDraft(draft)
			c.editorSaving = true
			if editor, ok := c.views[ViewCardEditor].(*views.CardEditor); ok {
				editor.SetSaving(true)
			}
			c.editorRequest++
			dashboardID, cardID, revision, generation, epoch, request, service := c.activeDashboard, c.editorCardID, c.editorRevision, c.viewGeneration, c.editorEpoch, c.editorRequest, c.editor
			return func() tea.Msg {
				var saved contracts.CardSaveResult
				var err error
				if cardID == "" {
					saved, err = service.CreateEditorCard(contracts.CreateCardInput{DashboardID: dashboardID, Values: draft.Values})
				} else {
					saved, err = service.UpdateEditorCard(
						contracts.UpdateCardInput{
							DashboardID:      dashboardID,
							CardID:           cardID,
							ExpectedRevision: revision,
							Changes:          draft.Changes,
						},
					)
				}
				return messages.CardSavedMsg{
					Result:      saved,
					Err:         err,
					DashboardID: dashboardID,
					CardID:      cardID,
					Generation:  generation,
					Epoch:       epoch,
					Request:     request,
				}
			}
		},
		OnReload: func() tea.Cmd {
			if c.editorSaving {
				return nil
			}
			c.editorEpoch++
			c.editorRequest++
			return cardEditorLoadCommand(c.editor, c.activeDashboard, c.editorCardID, c.viewGeneration, c.editorEpoch, c.editorRequest)
		},
		OnCancel: func() tea.Cmd {
			if c.editorSaving {
				return nil
			}
			dashboardID, generation, epoch, request := c.activeDashboard, c.viewGeneration, c.editorEpoch, c.editorRequest
			return func() tea.Msg {
				return messages.EditorCancelMsg{DashboardID: dashboardID, Generation: generation, Epoch: epoch, Request: request}
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

			return c.editorNavigationCommand(vt)
		},

		OnOpenMCPForm: func() tea.Cmd {
			return c.editorNavigationCommand(messages.ViewMCPForm)
		},

		OnOpenWorkspaces: func() tea.Cmd {
			return c.editorNavigationCommand(messages.ViewSourceManagerWorkspaces)
		},

		OnOpenPuppetSources: func() tea.Cmd {
			return c.editorNavigationCommand(messages.ViewSourceManagerPuppets)
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
		OnBack: func() tea.Cmd { return c.navigationCommand(messages.ViewDashboard) },
		OnSave: func(settings config.AppSettings) tea.Cmd {
			dashboardID, generation, repo := c.activeDashboard, c.viewGeneration, c.repo
			return func() tea.Msg {
				err := repo.SaveAppSettings(settings)
				return messages.SettingsSavedMsg{Settings: settings, Err: err, DashboardID: dashboardID, Generation: generation}
			}
		},
		OnOpenPuppetSources: func() tea.Cmd {
			return c.navigationCommand(messages.ViewSourceManagerPuppets)
		},
	}
}

func (c *Coordinator) mcpCallbacks() views.MCPFormCallbacks {
	return views.MCPFormCallbacks{
		OnSave: func(mcp config.MCPServer) tea.Cmd {
			dashboardID, generation, request, repo := c.activeDashboard, c.viewGeneration, c.mcpRequest, c.repo
			return func() tea.Msg {
				saved, err := repo.UpsertMCP(mcp)
				return messages.MCPSavedMsg{MCP: saved, Err: err, DashboardID: dashboardID, Generation: generation, Request: request}
			}
		},
	}
}

func (c *Coordinator) sessionLauncherCallbacks() views.SessionLauncherCallbacks {
	return views.SessionLauncherCallbacks{
		OnSetNameAsync: func(sessionID, name, processID string) tea.Cmd {
			c.renameRequest++
			dashboardID, generation, request, service := c.activeDashboard, c.viewGeneration, c.renameRequest, c.actions.port
			return func() tea.Msg {
				return messages.ScopedSessionNamedMsg{
					Request:     request,
					DashboardID: dashboardID,
					Generation:  generation,
					Name:        name,
					SessionID:   sessionID,
					Result: service.RenameDashboardSession(
						context.Background(),
						contracts.RenameSessionRequest{
							DashboardID:       dashboardID,
							SessionID:         sessionID,
							Name:              name,
							ExpectedProcessID: processID,
						},
					),
				}
			}
		},
	}
}

func (c *Coordinator) agentSwitchCallbacks() views.AgentSwitchCallbacks {
	return views.AgentSwitchCallbacks{
		OnSwitch: c.requestSwitch,
		OnCancel: func() tea.Cmd {
			return c.navigationCommand(messages.ViewDashboard)
		},
	}
}

func (c *Coordinator) requestRelaunch(sessionID string) tea.Cmd {
	return c.prepareConversation(sessionID, "", messages.ViewDashboard)
}

func (c *Coordinator) requestSwitch(sessionID, cardID string) tea.Cmd {
	return c.prepareConversation(sessionID, cardID, messages.ViewAgentSwitch)
}

func (c *Coordinator) prepareConversation(sessionID, cardID string, returnView messages.ViewType) tea.Cmd {
	if c.conversationCancel != nil {
		c.conversationCancel()
	}
	c.conversationEpoch++
	c.conversationReadRequest++
	c.conversationReading = true
	c.lifecycleRequest++
	c.lifecycleExecuting = false
	ctx, cancel := context.WithCancel(context.Background())
	c.conversationCancel = cancel
	return conversationReadCommand(
		ctx,
		c.actions.port,
		c.activeDashboard,
		c.viewGeneration,
		c.conversationEpoch,
		c.conversationReadRequest,
		sessionID,
		cardID,
		returnView,
		true,
	)
}

type conversationReader interface {
	ReadConversationPreparationResult(context.Context, contracts.ConversationPreparationRequest) contracts.ConversationPreparation
}

func conversationReadCommand(
	ctx context.Context,
	service conversationReader,
	dashboardID string,
	generation, epoch, request uint64,
	sessionID, cardID string,
	returnView messages.ViewType,
	initial bool,
) tea.Cmd {
	return func() tea.Msg {
		result := service.ReadConversationPreparationResult(
			ctx,
			contracts.ConversationPreparationRequest{DashboardID: dashboardID, SessionID: sessionID, CardID: cardID},
		)
		return messages.ConversationPreparedMsg{
			DashboardID: dashboardID,
			SessionID:   sessionID, CardID: cardID,
			Generation:  generation,
			Epoch:       epoch,
			Request:     request,
			ReturnView:  returnView,
			Initial:     initial,
			Preparation: result,
		}
	}
}

func (c *Coordinator) conversationChoiceCallbacks() views.ConversationChoiceCallbacks {
	return views.ConversationChoiceCallbacks{
		OnSelect: func(sessionID, cardID string, relaunch bool, choice contracts.ConversationChoice) tea.Cmd {
			if c.lifecycleExecuting {
				return nil
			}
			c.lifecycleExecuting = true
			c.lifecycleRequest++
			return c.actions.activate(c.activeDashboard, c.viewGeneration, c.lifecycleRequest, sessionID, cardID, relaunch, choice)
		},
		OnCancel: func(returnView messages.ViewType) tea.Cmd { return c.navigationCommand(returnView) },
	}
}
