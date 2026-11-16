package app

import (
	"context"
	"errors"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/api"
	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/BlackGiack/kankuro/internal/ui/component"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/styles"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

type ViewType int

const (
	ViewDashboard ViewType = iota
	ViewCardEditor
	ViewSourceManagerSkills
	ViewSourceManagerPersonas
	ViewSourceManagerWorkspaces
	ViewSourceManagerPuppets
	ViewSettings
	ViewMCPForm
	ViewMCPManager
	ViewSessionLauncher
	ViewAgentSwitch
	ViewConversationChoice
	ViewDashboardList
)

type Coordinator struct {
	// --- coordinator-goroutine state: never read or written from a tea.Cmd closure ---
	currentView                                                  ViewType
	prevView                                                     ViewType
	views                                                        map[ViewType]component.Model
	width                                                        int
	height                                                       int
	err                                                          error
	settings                                                     config.AppSettings
	preview                                                      previewSub
	activeDashboard                                              string
	viewGeneration                                               uint64
	sessionEpoch                                                 uint64
	launchRequest                                                uint64
	launchCardID                                                 string
	mutationRequest                                              uint64
	renameRequest, workspaceRequest                              uint64
	conversationEpoch, conversationReadRequest, lifecycleRequest uint64
	conversationReading                                          bool
	lifecycleExecuting                                           bool
	conversationCancel                                           context.CancelFunc
	dashboardListRequest                                         uint64
	dashboardViews                                               map[string]*views.Dashboard
	editorEpoch                                                  uint64
	editorRequest                                                uint64
	editorCardID                                                 string
	editorRevision                                               string
	editorSaving                                                 bool
	editorCompleted                                              bool
	mcpReturn                                                    ViewType
	mcpRequest                                                   uint64
	mcpManagerVisit                                              uint64
	mcpDeleteRequest                                             uint64

	configFingerprint           contracts.RegistryChangeToken
	displayChoices              contracts.DisplayChoices
	refreshRequestID            uint64
	lastAppliedRefreshRequest   uint64
	inventoryRequestID          uint64
	lastAppliedInventoryRequest uint64

	// --- safe from any goroutine: concurrency-safe handles, and helpers that
	// hold no coordinator pointer ---
	repo    config.Repository
	session *session.Manager
	refresh refresher
	reads   dashboardInventoryReader
	editor  cardEditorPort
	actions sessionActions
	legacy  legacyActions
}

func NewCoordinator(repo config.Repository, sm *session.Manager, service *api.Service) *Coordinator {
	c := &Coordinator{
		repo:           repo,
		session:        sm,
		preview:        previewSub{sessions: sm},
		refresh:        refresher{repo: repo, sessions: sm, editor: service},
		reads:          service,
		editor:         service,
		actions:        sessionActions{port: service},
		legacy:         legacyActions{repo: repo, service: service},
		views:          make(map[ViewType]component.Model),
		currentView:    ViewDashboardList,
		dashboardViews: make(map[string]*views.Dashboard),
	}

	settings, err := repo.GetAppSettings()
	if err != nil {
		settings = config.DefaultAppSettings()
	}
	c.settings = settings

	c.initViews()
	c.views[ViewDashboardList] = views.NewDashboardList(c.dashboardListCallbacks())
	return c
}

// CurrentView returns the currently active view.
func (c *Coordinator) CurrentView() ViewType {
	return c.currentView
}

// PrevView returns the previous view.
func (c *Coordinator) PrevView() ViewType {
	return c.prevView
}

// Views returns the map of all views.
func (c *Coordinator) Views() map[ViewType]component.Model {
	return c.views
}

// Width returns the coordinator width.
func (c *Coordinator) Width() int {
	return c.width
}

// Height returns the coordinator height.
func (c *Coordinator) Height() int {
	return c.height
}

// Err returns the current error, if any.
func (c *Coordinator) Err() error {
	return c.err
}

func (c *Coordinator) Init() tea.Cmd {
	// Initialize the entry view so terminal setup runs before the first frame.
	var initialViewCmd tea.Cmd
	if view, ok := c.views[c.currentView]; ok {
		initialViewCmd = view.Init()
	}
	var dashboardListCmd = c.refreshDashboardList()

	return tea.Batch(
		initialViewCmd,
		dashboardListCmd,
		c.refreshDataSeeded(),
		tick(),
		c.refresh.syncPuppets(),
		c.refresh.sessionLoadErrors(),
	)
}

func (c *Coordinator) Update(msg tea.Msg) (*Coordinator, tea.Cmd) {
	// Below the minimum terminal size the user cannot see the view, so input
	// must not reach it (see size_guard.go).
	if c.tooSmall() {
		if handled, cmd := c.guardKey(msg); handled {
			return c, cmd
		}
	}

	// The error banner clears when the user acts, not as a side effect of
	// navigation. Every error reaches the coordinator as a later message
	// (ErrorMsg, CardSavedMsg{Err}, ...), never inside the Update of the
	// keypress that caused it, so clearing here and setting on the result
	// shows the new error rather than the old one.
	if _, ok := msg.(tea.KeyPressMsg); ok {
		c.err = nil
	}

	if handled, cmd := c.handleSpecialMessages(msg); handled {
		return c, cmd
	}

	updated, cmd := c.updateCurrentView(msg)
	if updated != nil {
		c.views[c.currentView] = updated
	}
	return c, cmd
}

//nolint:funlen,gocognit,gocyclo,cyclop // Central message router keeps ordering-sensitive app transitions together.
func (c *Coordinator) handleSpecialMessages(msg tea.Msg) (bool, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		c.width = m.Width
		c.height = m.Height
		c.resizeAll()
		return true, nil

	case tickMsg:
		// Refresh scoped sessions every tick; reread configuration only when its
		// fingerprint changes.
		var listCmd tea.Cmd
		if c.currentView == ViewDashboardList {
			listCmd = c.refreshDashboardList()
		}
		return true, tea.Batch(c.refreshSessions(), c.maybeRefreshConfig(), listCmd, tick())

	case messages.DashboardListLoadedMsg:

		if m.RequestID != c.dashboardListRequest {
			return true, nil
		}
		if m.Err != nil {
			c.err = m.Err
			return true, nil
		}
		if view, ok := c.views[ViewDashboardList].(*views.DashboardList); ok {
			view.SetDashboards(m.Dashboards)
		}
		return true, nil

	case messages.DashboardMutationFinishedMsg:
		if m.Generation != c.viewGeneration {
			return true, nil
		}
		if m.Err != nil {
			c.err = m.Err
		}
		return true, c.refreshDashboardList()

	case messages.CardDeletionPreparedMsg:
		if m.DashboardID != c.activeDashboard || m.Generation != c.viewGeneration || c.currentView != ViewDashboard {
			return true, nil
		}
		if m.Err != nil {
			c.err = m.Err
			return true, nil
		}
		if m.RequiresForeignConfirmation {
			if dashboard, ok := c.views[ViewDashboard].(*views.Dashboard); ok {
				dashboard.ShowCardDeletionImpact(m.CardID, m.Token, m.Summary)
			}
			return true, nil
		}
		return true, confirmCardDeletionCommand(c.legacy.service, m.DashboardID, m.CardID, m.Token, m.Generation, false)

	case messages.DashboardOpenRequestedMsg:
		if c.currentView != ViewDashboardList || m.Generation != c.viewGeneration || m.DashboardID == "" ||
			m.DashboardState != contracts.DashboardActive {
			return true, nil
		}
		c.activeDashboard = m.DashboardID
		c.viewGeneration++
		dashboard := c.dashboardViews[m.DashboardID]
		if dashboard == nil {
			dashboard = views.NewDashboard(c.dashboardCallbacks())
			c.dashboardViews[m.DashboardID] = dashboard
		}
		dashboard.SetDashboardName(m.DashboardName)
		dashboard.SetDashboardScope(m.DashboardID, c.viewGeneration)
		dashboard.SetSize(c.width, c.height-views.CoordinatorRows)
		c.views[ViewDashboard] = dashboard
		return true, tea.Batch(c.navigateTo(ViewDashboard), c.refreshDataSeeded())

	case messages.DashboardCardLoadedMsg:
		if m.DashboardID != c.activeDashboard || m.Generation != c.viewGeneration || m.Request != c.launchRequest ||
			c.currentView != ViewDashboard {
			return true, nil
		}
		if m.Err != nil {
			c.err = m.Err
			return true, nil
		}
		if m.Card.ID != m.CardID {
			c.err = errors.New("card read target mismatch")
			return true, nil
		}
		switch m.Action {
		case messages.DashboardCardLaunch:
			return true, c.launchCard(m.Card, m.Request)
		default:
			return true, nil
		}

	case messages.DashboardListRequestedMsg:

		if m.DashboardID != c.activeDashboard || m.Generation != c.viewGeneration {
			return true, nil
		}
		if dashboard, ok := c.views[ViewDashboard].(*views.Dashboard); ok {
			dashboard.SetDashboardScope("", c.viewGeneration+1)
		}
		c.activeDashboard = ""
		c.viewGeneration++
		return true, tea.Batch(c.navigateTo(ViewDashboardList), c.refreshDashboardList())

	case configFingerprintMsg:
		return true, c.handleConfigFingerprint(m)

	case messages.DataRefreshRequestedMsg:
		return true, c.refreshDataSeeded()

	case messages.PuppetsSyncedMsg:
		return true, c.refreshDataSeeded()

	case messages.EditorCancelMsg:
		if m.DashboardID != c.activeDashboard || m.Generation != c.viewGeneration || m.Epoch != c.editorEpoch ||
			m.Request != c.editorRequest ||
			c.currentView != ViewCardEditor ||
			c.editorSaving {
			return true, nil
		}
		c.editorEpoch++
		return true, c.navigateTo(ViewDashboard)

	case messages.EditorNavigateMsg:
		if m.DashboardID != c.activeDashboard || m.Generation != c.viewGeneration || m.Epoch != c.editorEpoch ||
			c.currentView != ViewCardEditor ||
			c.editorSaving ||
			c.editorCompleted {
			return true, nil
		}
		return true, c.navigateTo(convertViewType(m.View))

	case messages.ScopedNavigateMsg:
		if c.currentView == ViewCardEditor && c.editorSaving {
			return true, nil
		}
		if m.DashboardID != c.activeDashboard || m.Generation != c.viewGeneration {
			return true, nil
		}
		return true, c.navigateTo(convertViewType(m.View))

	case messages.NavigateBackMsg:
		if c.currentView == ViewMCPForm {
			c.mcpRequest++
			return true, c.navigateTo(c.mcpReturn)
		}
		if c.currentView == ViewMCPManager {
			return true, c.navigateTo(ViewDashboard)
		}
		return true, c.navigateTo(c.prevView)

	case messages.MCPManagerEditMsg:
		if !c.matchesMCPManagerScope(m.Scope) {
			return true, nil
		}
		cmd := c.navigateTo(ViewMCPForm)
		if form, ok := c.views[ViewMCPForm].(*views.MCPForm); ok {
			form.SetMCP(m.MCP)
		}
		return true, cmd

	case messages.MCPManagerDeleteMsg:
		manager, ok := c.views[ViewMCPManager].(*views.MCPManager)
		if !ok {
			return true, nil
		}
		if !c.matchesMCPManagerScope(m.Scope) {
			manager.DiscardDelete(m.Scope.Visit)
			return true, nil
		}
		if !manager.AdmitDelete(m.Scope.Visit) {
			return true, nil
		}
		c.mcpDeleteRequest++
		return true, c.legacy.deleteMCPFromManager(m.MCPID, c.activeDashboard, c.viewGeneration, c.mcpDeleteRequest)

	case messages.MCPManagerDeletedMsg:
		if m.Request == c.mcpDeleteRequest {
			if manager, ok := c.views[ViewMCPManager].(*views.MCPManager); ok {
				manager.DeleteFinished()
			}
			if m.Err != nil && m.DashboardID == c.activeDashboard && m.Generation == c.viewGeneration && c.currentView == ViewMCPManager {
				c.err = m.Err
			}
		}
		if m.Err == nil {
			return true, c.refreshDataSeeded()
		}
		return true, nil

	case messages.CardEditorLoadedMsg:
		if m.DashboardID != c.activeDashboard || m.Generation != c.viewGeneration || m.Epoch != c.editorEpoch ||
			m.Request != c.editorRequest ||
			m.CardID != c.editorCardID ||
			(c.currentView != ViewDashboard && c.currentView != ViewCardEditor) {
			return true, nil
		}
		if m.Err != nil {
			c.err = m.Err
			return true, nil
		}
		editor, ok := c.views[ViewCardEditor].(*views.CardEditor)
		if !ok {
			return true, nil
		}
		editor.SetChoices(m.Choices)
		if err := editorChoicesError(m.Choices); err != nil {
			c.err = err
		}
		editor.LoadEditor(m.Data)
		c.editorCompleted = false
		c.editorRevision = m.Data.Revision
		return true, c.navigateTo(ViewCardEditor)

	case messages.CardSavedMsg:
		matching := m.DashboardID == c.activeDashboard && m.Generation == c.viewGeneration && m.Epoch == c.editorEpoch &&
			m.Request == c.editorRequest &&
			m.CardID == c.editorCardID
		if matching {
			c.editorSaving = false
			if editor, ok := c.views[ViewCardEditor].(*views.CardEditor); ok {
				editor.SetSaving(false)
			}
		}
		if !matching {
			if m.Err == nil {
				return true, c.refreshDataSeeded()
			}
			return true, nil
		}
		if c.currentView != ViewCardEditor {
			if m.Err != nil {
				c.err = m.Err
				return true, nil
			}
			c.editorCompleted = true
			c.editorCardID, c.editorRevision = m.Result.CardID, m.Result.Revision
			if editor, ok := c.views[ViewCardEditor].(*views.CardEditor); ok {
				editor.MarkSaveCompleted()
			}
			return true, c.refreshDataSeeded()
		}
		if m.Err != nil {
			c.err = m.Err
			return true, nil
		}
		c.editorEpoch++
		return true, tea.Batch(c.refreshDataSeeded(), c.navigateTo(ViewDashboard))

	case messages.MCPSavedMsg:
		if m.DashboardID != c.activeDashboard || m.Generation != c.viewGeneration ||
			m.Request != c.mcpRequest || c.currentView != ViewMCPForm {
			if m.Err == nil {
				return true, c.refreshDataSeeded()
			}
			return true, nil
		}
		if m.Err != nil {
			c.err = m.Err
			return true, nil
		}
		return true, tea.Batch(c.refreshDataSeeded(), c.navigateTo(c.mcpReturn))

	case messages.SettingsSavedMsg:
		if m.DashboardID != c.activeDashboard || m.Generation != c.viewGeneration {
			if m.Err == nil {
				c.settings = m.Settings
				if editor, ok := c.views[ViewCardEditor].(*views.CardEditor); ok {
					editor.SetSettings(editorSettings(m.Settings.ShowPersonasInCardEditor, m.Settings.ShowMCPsInCardEditor))
				}
			}
			return true, c.refreshDataSeeded()
		}
		if m.Err != nil {
			c.err = m.Err
			return true, nil
		}
		c.settings = m.Settings
		if editor, ok := c.views[ViewCardEditor].(*views.CardEditor); ok {
			editor.SetSettings(editorSettings(m.Settings.ShowPersonasInCardEditor, m.Settings.ShowMCPsInCardEditor))
		}
		return true, tea.Batch(c.refreshDataSeeded(), c.navigateTo(ViewDashboard))

	case messages.ErrorMsg:
		c.err = m.Err
		if c.currentView == ViewConversationChoice {
			return true, c.refreshDataSeeded()
		}
		return true, nil

	case messages.MCPServerFailedMsg:
		c.err = m.Err
		return true, tea.Quit

	case messages.MCPDeletedMsg:
		if editor, ok := c.views[ViewCardEditor].(*views.CardEditor); ok {
			editor.Update(m)
		}
		return true, c.refreshDataSeeded()

	case messages.MCPDeleteRequestedMsg:
		// Disk I/O deferred to a tea.Cmd, consistent with every other
		// mutation - this used to run synchronously inside Update.
		return true, c.legacy.deleteMCP(m)

	case messages.ClearErrorMsg:
		c.err = nil
		return true, nil
	}

	return c.handleSessionMessages(msg)
}

//nolint:gocognit,gocyclo,cyclop,funlen // Session message routing preserves scoped transitions.
func (c *Coordinator) handleSessionMessages(msg tea.Msg) (bool, tea.Cmd) {
	switch m := msg.(type) {
	case messages.DataLoadedMsg:
		return true, c.handleDataLoaded(m)

	case messages.SessionsRefreshedMsg:
		return true, c.handleSessionsRefreshed(m)

	case messages.SessionPreviewRequestedMsg:
		return true, c.handlePreviewRequest(m.SessionID)

	case messages.SessionOutputMsg:
		return true, c.handleSessionOutput(m)

	case messages.SessionLaunchCompletedMsg:
		if m.DashboardID != c.activeDashboard || m.Generation != c.viewGeneration {
			return true, nil
		}
		if m.Request != c.launchRequest || m.CardID == "" || m.CardID != c.launchCardID {
			if m.Result.Validate() != nil || m.Result.Effect != contracts.NoEffect {
				c.sessionEpoch++
			}
			return true, c.refreshDataSeeded()
		}
		launcher, ok := c.views[ViewSessionLauncher].(*views.SessionLauncher)
		if ok {
			if launcher.Card().ID != m.CardID {
				return true, c.refreshDataSeeded()
			}
			launcher.SetLaunchResult(m.Result)
		}
		validationErr := m.Result.Validate()
		if validationErr != nil || m.Result.Effect != contracts.NoEffect {
			c.sessionEpoch++
		}
		if validationErr == nil && m.Result.Effect == contracts.UsableActivation {
			if dashboard, found := c.views[ViewDashboard].(*views.Dashboard); found {
				dashboard.FocusSession(m.Result.SessionID)
			}
		}
		return true, c.refreshDataSeeded()

	case messages.AgentSwitchRequestedMsg:
		if m.DashboardID != c.activeDashboard || m.Generation != c.viewGeneration {
			return true, nil
		}
		switchView, ok := c.views[ViewAgentSwitch].(*views.AgentSwitchView)
		if ok {
			switchView.SetSession(m.SessionID, m.SessionName)
		}
		return true, c.navigateTo(ViewAgentSwitch)

	case messages.ConversationPreparedMsg:
		return true, c.handleConversationPrepared(m)
	case messages.SessionActivationCompletedMsg:
		if m.DashboardID != c.activeDashboard || m.Generation != c.viewGeneration {
			return true, nil
		}
		if m.Request != c.lifecycleRequest {
			if m.Result.Validate() != nil || m.Result.Effect != contracts.NoEffect {
				c.sessionEpoch++
			}
			return true, c.refreshDataSeeded()
		}
		c.lifecycleExecuting = false
		result := m.Result
		if err := result.Validate(); err != nil {
			c.sessionEpoch++
			c.err = err
			return true, c.refreshDataSeeded()
		}
		if result.Effect != contracts.UsableActivation || result.SessionID != m.SessionID {
			if result.Effect != contracts.NoEffect {
				c.sessionEpoch++
			}
			if result.Diagnostic != "" {
				c.err = errors.New(result.Diagnostic)
			} else {
				c.err = errors.New("activation outcome is invalid")
			}
			return true, c.refreshDataSeeded()
		}
		if result.Diagnostic != "" {
			c.err = errors.New(result.Diagnostic)
		}
		c.sessionEpoch++
		init := c.navigateTo(ViewDashboard)
		return true, tea.Batch(
			init,
			c.refreshDataSeeded(),
			c.requestAttachExpected(result.SessionID, result.ProcessActivationID, messages.AttachRequestLifecycle),
		)

	case messages.SessionAttachRequestedMsg:
		if m.DashboardID != c.activeDashboard || m.Generation != c.viewGeneration {
			return true, nil
		}
		// Navigate here, not in the callback that produced this message
		// (invariant 2, see navigateTo's doc comment).
		var initCmd tea.Cmd
		if c.currentView != ViewDashboard {
			initCmd = c.navigateTo(ViewDashboard)
		}
		if m.Warning != nil {
			c.err = m.Warning
		}
		return true, tea.Batch(initCmd, c.requestAttach(m.SessionID))

	case messages.ScopedAttachReadyMsg:
		staleRequest := (m.RequestKind == messages.AttachRequestLaunch && m.Request != c.launchRequest) ||
			(m.RequestKind == messages.AttachRequestLifecycle && m.Request != c.lifecycleRequest) ||
			(m.RequestKind != "" && m.RequestKind != messages.AttachRequestLaunch && m.RequestKind != messages.AttachRequestLifecycle)
		if m.DashboardID != c.activeDashboard || m.Generation != c.viewGeneration || staleRequest {
			if m.Handle != nil {
				m.Handle.Close()
			}
			return true, nil
		}
		if m.Err != nil {
			c.err = m.Err
			return true, nil
		}
		if m.Handle == nil {
			c.err = errors.New("attachment unavailable")
			return true, nil
		}
		return true, scopedAttachCommand(m)

	case messages.ScopedSessionNamedMsg:
		if m.DashboardID != c.activeDashboard || m.Generation != c.viewGeneration {
			return true, nil
		}
		if m.Request != c.renameRequest || c.currentView != ViewSessionLauncher {
			if m.Result.Applied {
				c.sessionEpoch++
			}
			return true, c.refreshDataSeeded()
		}
		launcher, ok := c.views[ViewSessionLauncher].(*views.SessionLauncher)
		if !ok || launcher.SessionID() != m.SessionID || launcher.ProcessActivationID() == "" {
			return true, nil
		}
		var renameErr error
		if m.Result.Diagnostic != "" {
			renameErr = errors.New(m.Result.Diagnostic)
		}
		if !m.Result.Applied {
			if renameErr == nil {
				renameErr = errors.New("rename did not apply")
			}
			launcher.SetNameResult(renameErr)
			return true, c.refreshDataSeeded()
		}
		launcher.SetNameResult(nil)
		if renameErr != nil {
			c.err = renameErr
		}

		c.sessionEpoch++
		if dashboard, found := c.views[ViewDashboard].(*views.Dashboard); found {
			dashboard.FocusSessionAfterRename(m.SessionID, m.Name)
		}
		return true, tea.Batch(
			c.navigateTo(ViewDashboard),
			c.requestAttachExpected(m.SessionID, launcher.ProcessActivationID(), messages.AttachRequestLaunch),
		)

	case messages.DashboardSessionRenamedMsg:
		if m.DashboardID != c.activeDashboard || m.Generation != c.viewGeneration {
			return true, nil
		}
		if m.Request != c.renameRequest || c.currentView != ViewDashboard {
			if m.Result.Applied {
				c.sessionEpoch++
			}
			return true, c.refreshDataSeeded()
		}
		dashboard, ok := c.views[ViewDashboard].(*views.Dashboard)
		if !ok {
			return true, nil
		}
		var renameErr error
		if m.Result.Diagnostic != "" {
			renameErr = errors.New(m.Result.Diagnostic)
		}
		var formErr error
		if !m.Result.Applied {
			if renameErr == nil {
				renameErr = errors.New("rename did not apply")
			}
			formErr = renameErr
		}
		if !dashboard.CompleteRename(m.SessionID, formErr) {
			return true, nil
		}
		if m.Result.Applied {
			c.sessionEpoch++
			if renameErr != nil {
				c.err = renameErr
			}
		}
		return true, c.refreshDataSeeded()

	case messages.WorkspaceOpenedMsg:
		if m.DashboardID != c.activeDashboard || m.Generation != c.viewGeneration || m.Request != c.workspaceRequest {
			return true, nil
		}
		if m.Result.Diagnostic != "" {
			c.err = errors.New(m.Result.Diagnostic)
		} else if !m.Result.Started {
			c.err = errors.New("workspace command did not start")
		}
		return true, nil

	case messages.SessionStoppedMsg:
		if m.DashboardID != c.activeDashboard || m.Generation != c.viewGeneration {
			return true, nil
		}
		if m.Request != c.mutationRequest {
			if m.Result.ProcessTerminated {
				c.sessionEpoch++
			}
			return true, c.refreshDataSeeded()
		}
		if m.Result.Diagnostic != "" {
			c.err = errors.New(m.Result.Diagnostic)
		}
		if m.Result.ProcessTerminated {
			c.sessionEpoch++
		}
		return true, c.refreshDataSeeded()
	case messages.SessionRemovedMsg:
		if m.DashboardID != c.activeDashboard || m.Generation != c.viewGeneration {
			return true, nil
		}
		if m.Request != c.mutationRequest {
			if m.Result.InMemoryRemoved || m.Result.RecordDeleted || m.Result.WorkspaceRemoved {
				c.sessionEpoch++
			}
			return true, c.refreshDataSeeded()
		}
		if m.Result.Diagnostic != "" {
			c.err = errors.New(m.Result.Diagnostic)
		}
		if m.Result.InMemoryRemoved || m.Result.RecordDeleted || m.Result.WorkspaceRemoved {
			c.sessionEpoch++
		}
		return true, c.refreshDataSeeded()

	case messages.ScopedActionFinishedMsg:
		if m.DashboardID != c.activeDashboard || m.Generation != c.viewGeneration {
			return true, nil
		}
		if m.Err != nil {
			c.err = m.Err
		}
		return true, c.refreshDataSeeded()

	case messages.AttachFinishedMsg:
		if m.DashboardID != c.activeDashboard || m.Generation != c.viewGeneration {
			return true, nil
		}
		if m.Err != nil {
			c.err = m.Err
		}
		return true, c.refreshDataSeeded()
	}

	return false, nil
}

func (c *Coordinator) handleConversationPrepared(m messages.ConversationPreparedMsg) tea.Cmd {
	if m.DashboardID != c.activeDashboard || m.Generation != c.viewGeneration || m.Epoch != c.conversationEpoch ||
		m.Request != c.conversationReadRequest {
		return nil
	}
	c.conversationReading = false
	if c.conversationCancel != nil {
		c.conversationCancel()
		c.conversationCancel = nil
	}
	choiceView, ok := c.views[ViewConversationChoice].(*views.ConversationChoiceView)
	if !ok {
		c.err = errors.New("conversation selection view is unavailable")
		return nil
	}
	if err := validateConversationPrepared(m); err != nil {
		c.err = err
		if !m.Initial {
			choiceView.SetContextError(err.Error())
		}
		return nil
	}

	if m.Initial &&
		((m.ReturnView == messages.ViewDashboard && c.currentView != ViewDashboard) || (m.ReturnView == messages.ViewAgentSwitch && c.currentView != ViewAgentSwitch)) {
		return nil
	}
	if m.Initial && !m.Preparation.HasSaved {
		c.lifecycleExecuting = true
		return c.actions.activate(
			c.activeDashboard,
			c.viewGeneration,
			c.lifecycleRequest,
			m.Preparation.SessionID,
			m.Preparation.CardID,
			m.ReturnView == messages.ViewDashboard,
			contracts.ConversationChoice{Kind: contracts.ConversationNew},
		)
	}
	if m.Initial {
		choiceView.SetContext(m.Preparation.SessionID, m.Preparation.CardID, m.ReturnView, m.Preparation.Options)
		return c.navigateTo(ViewConversationChoice)
	}
	choiceView.RefreshContext(m.Preparation.SessionID, m.Preparation.CardID, m.Preparation.Options)
	return nil
}

func validateConversationPrepared(m messages.ConversationPreparedMsg) error {
	if err := m.Preparation.Validate(); err != nil {
		return err
	}
	if m.Preparation.Diagnostic != "" {
		return errors.New(m.Preparation.Diagnostic)
	}
	if m.Preparation.DashboardID != m.DashboardID || m.Preparation.SessionID != m.SessionID ||
		(m.CardID != "" && m.Preparation.CardID != m.CardID) || m.Preparation.CardID == "" {
		return errors.New("conversation preparation target mismatch")
	}
	if !m.Preparation.Availability.Enabled {
		return errors.New(m.Preparation.Availability.Reason)
	}
	return nil
}

func (c *Coordinator) refreshConversationChoiceContext() tea.Cmd {
	view, ok := c.views[ViewConversationChoice].(*views.ConversationChoiceView)
	if !ok || c.conversationReading {
		return nil
	}
	sessionID, cardID, returnView := view.Target()
	c.conversationReadRequest++
	c.conversationReading = true
	ctx, cancel := context.WithCancel(context.Background())
	c.conversationCancel = cancel
	view.SetContextError("Refreshing conversations")
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
		false,
	)
}

func (c *Coordinator) View() string {
	if c.tooSmall() {
		return c.tooSmallView()
	}
	if c.err != nil {
		return styles.ErrorStyle.Render(c.err.Error()) + "\n\n" + c.currentViewView()
	}
	return c.currentViewView()
}

func (c *Coordinator) matchesMCPManagerScope(scope messages.MCPManagerScope) bool {
	return c.currentView == ViewMCPManager && scope.DashboardID == c.activeDashboard &&
		scope.Generation == c.viewGeneration && scope.Visit == c.mcpManagerVisit
}
