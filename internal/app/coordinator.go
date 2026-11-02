package app

import (
	"errors"
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/api"
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
	ViewSessionLauncher
	ViewAgentSwitch
	ViewConversationChoice
	ViewDashboardList
)

type Coordinator struct {
	// --- coordinator-goroutine state: never read or written from a tea.Cmd closure ---
	currentView          ViewType
	prevView             ViewType
	views                map[ViewType]component.Model
	width                int
	height               int
	err                  error
	settings             config.AppSettings
	preview              previewSub
	activeDashboard      string
	viewGeneration       uint64
	dashboardListRequest uint64
	dashboardViews       map[string]*views.Dashboard
	editingCard          config.IDCard

	// configFingerprint tracks the last loaded registry stamps. Seeded refreshes
	// update the baseline; idle ticks skip configuration reads when it is unchanged.
	configFingerprint map[string]config.FileStamp

	// --- safe from any goroutine: concurrency-safe handles, and helpers that
	// hold no coordinator pointer ---
	repo    config.Repository
	session *session.Manager
	refresh refresher
	actions sessionActions
}

func NewCoordinator(repo config.Repository, sm *session.Manager, service *api.Service) *Coordinator {
	c := &Coordinator{
		repo:           repo,
		session:        sm,
		preview:        previewSub{sessions: sm},
		refresh:        refresher{repo: repo, sessions: sm},
		actions:        sessionActions{repo: repo, sessions: sm, service: service},
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
		return true, confirmCardDeletionCommand(c.actions.service, m.DashboardID, m.CardID, m.Token, m.Generation, false)

	case messages.DashboardOpenRequestedMsg:
		if c.currentView != ViewDashboardList || m.Generation != c.viewGeneration || m.Dashboard.ID == "" ||
			m.Dashboard.State != config.DashboardActive {
			return true, nil
		}
		c.activeDashboard = m.Dashboard.ID
		c.viewGeneration++
		dashboard := c.dashboardViews[m.Dashboard.ID]
		if dashboard == nil {
			dashboard = views.NewDashboard(c.dashboardCallbacks())
			c.dashboardViews[m.Dashboard.ID] = dashboard
		}
		dashboard.SetDashboardName(m.Dashboard.Name)
		dashboard.SetDashboardScope(m.Dashboard.ID, c.viewGeneration)
		dashboard.SetSize(c.width, c.height-views.CoordinatorRows)
		c.views[ViewDashboard] = dashboard
		return true, tea.Batch(c.navigateTo(ViewDashboard), c.refreshDataSeeded())

	case messages.DashboardListRequestedMsg:

		if m.DashboardID != c.activeDashboard || m.Generation != c.viewGeneration {
			return true, nil
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

	case messages.ScopedNavigateMsg:
		if m.DashboardID != c.activeDashboard || m.Generation != c.viewGeneration {
			return true, nil
		}
		return true, c.navigateTo(convertViewType(m.View))

	case messages.NavigateBackMsg:
		return true, c.navigateTo(c.prevView)

	case messages.CardSavedMsg:
		if m.DashboardID != c.activeDashboard || m.Generation != c.viewGeneration {
			return true, nil
		}
		if m.Err != nil {
			c.err = m.Err
			return true, nil
		}
		return true, tea.Batch(c.refreshDataSeeded(), c.navigateTo(ViewDashboard))

	case messages.MCPSavedMsg:
		if m.DashboardID != c.activeDashboard || m.Generation != c.viewGeneration {
			return true, c.refreshDataSeeded()
		}
		if m.Err != nil {
			c.err = m.Err
			return true, nil
		}
		// Refresh data to update MCP list in card editor
		return true, tea.Batch(c.refreshDataSeeded(), c.navigateTo(c.prevView))

	case messages.SettingsSavedMsg:
		if m.DashboardID != c.activeDashboard || m.Generation != c.viewGeneration {
			if m.Err == nil {
				c.settings = m.Settings
				if editor, ok := c.views[ViewCardEditor].(*views.CardEditor); ok {
					editor.SetSettings(m.Settings)
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
			editor.SetSettings(m.Settings)
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

	case messages.MCPDeleteRequestedMsg:
		// Disk I/O deferred to a tea.Cmd, consistent with every other
		// mutation - this used to run synchronously inside Update.
		return true, c.actions.deleteMCP(m)

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
		launcher, ok := c.views[ViewSessionLauncher].(*views.SessionLauncher)
		if ok {
			launcher.SetLaunchResult(m.SessionID, m.Err)
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

	case messages.ConversationChoiceRequestedMsg:
		if m.DashboardID != c.activeDashboard || m.Generation != c.viewGeneration {
			return true, nil
		}
		return true, c.handleConversationChoiceRequested(m)

	case messages.AgentSwitchedMsg:
		if m.DashboardID != c.activeDashboard || m.Generation != c.viewGeneration {
			return true, nil
		}
		return true, c.handleAgentSwitched(m)

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
		if m.DashboardID != c.activeDashboard || m.Generation != c.viewGeneration {
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
		if m.DashboardID != c.activeDashboard || m.Generation != c.viewGeneration ||
			c.currentView != ViewSessionLauncher {
			return true, nil
		}
		if launcher, ok := c.views[ViewSessionLauncher].(*views.SessionLauncher); ok {
			launcher.SetNameResult(m.Err)
		}
		if m.Err != nil {
			return true, nil
		}
		return true, tea.Batch(c.navigateTo(ViewDashboard), c.requestAttach(m.SessionID))

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

func (c *Coordinator) handleConversationChoiceRequested(m messages.ConversationChoiceRequestedMsg) tea.Cmd {
	snapshot, found := c.session.GetSession(m.SessionID)
	if !found {
		c.err = fmt.Errorf("session not found: %s", m.SessionID)
		return nil
	}
	if snapshot.DashboardID != c.activeDashboard {
		c.err = errors.New("session is outside dashboard scope")
		return nil
	}
	if snapshot.Status != session.StatusTerminated {
		c.err = errors.New("session is no longer terminated; refresh and retry the action")
		return c.refreshDataSeeded()
	}
	cardID := m.CardID
	if cardID == "" {
		cardID = snapshot.CardID
	}
	bindings := snapshot.Conversations[cardID]
	if len(bindings) == 0 {
		c.err = errors.New("saved conversations are no longer available; refresh and retry the action")
		return c.refreshDataSeeded()
	}
	choiceView, ok := c.views[ViewConversationChoice].(*views.ConversationChoiceView)
	if !ok {
		c.err = errors.New("conversation selection view is unavailable")
		return nil
	}
	choiceView.SetContext(m.SessionID, cardID, m.Driver, m.ReturnView, bindings)
	return c.navigateTo(ViewConversationChoice)
}

func (c *Coordinator) refreshConversationChoiceContext() {
	choiceView, ok := c.views[ViewConversationChoice].(*views.ConversationChoiceView)
	if !ok {
		return
	}
	sessionID, cardID, _ := choiceView.Target()
	snapshot, found := c.session.GetSession(sessionID)
	if !found {
		c.setConversationChoiceError(choiceView, "The selected session was removed. Cancel and start again.")
		return
	}
	if snapshot.Status != session.StatusTerminated {
		c.setConversationChoiceError(choiceView, "The session is no longer terminated. Cancel and choose another action.")
		return
	}
	dashboard, ok := c.views[ViewDashboard].(*views.Dashboard)
	if !ok {
		c.setConversationChoiceError(choiceView, "Session data is unavailable. Cancel and try again.")
		return
	}
	var destination *config.IDCard
	for i := range dashboard.Cards() {
		card := dashboard.Cards()[i]
		if card.ID == cardID {
			destination = &card
			break
		}
	}
	if destination == nil {
		c.setConversationChoiceError(choiceView, "The destination card was removed. Cancel and choose another card.")
		return
	}
	bindings := snapshot.Conversations[cardID]
	if len(bindings) == 0 {
		c.setConversationChoiceError(choiceView, "Saved conversations are no longer available. Cancel and restart the action.")
		return
	}
	if selected, hasSelected := choiceView.SelectedBinding(); hasSelected {
		foundSelected := false
		for _, binding := range bindings {
			if binding.ID == selected.ID && binding.Driver == selected.Driver {
				foundSelected = true
				break
			}
		}
		if !foundSelected {
			c.setConversationChoiceError(
				choiceView,
				"The selected conversation is no longer available for this driver. Cancel and restart the action.",
			)
			return
		}
	}
	choiceView.RefreshContext(sessionID, cardID, destination.Driver, bindings)
}

func (c *Coordinator) setConversationChoiceError(view *views.ConversationChoiceView, message string) {
	err := errors.New(message)
	view.SetContextError(message)
	c.err = err
}

func (c *Coordinator) handleAgentSwitched(m messages.AgentSwitchedMsg) tea.Cmd {
	if m.Err != nil && !errors.Is(m.Err, session.ErrPersistFailed) {
		c.err = m.Err
		return c.refreshDataSeeded()
	}
	// The agent did switch (its process restarted); a persist failure only
	// means its on-disk record didn't save, so still attach, but warn. Set
	// c.err directly here (not only via SessionAttachRequestedMsg.Warning
	// below) so the warning is visible even to a caller that processes this
	// message without driving the follow-on Cmd - the same guarantee the
	// "relaunch" path gets for free by returning a single message instead of
	// a batch.
	if m.Err != nil {
		c.err = m.Err
	}
	// Route the warning through SessionAttachRequestedMsg.Warning too - the
	// same path relaunch uses. This handler runs outside updateCurrentView's
	// write-back, so unlike the OnSwitch callback that produced this message
	// it is a safe place to hand off to the handler that navigates.
	return tea.Batch(
		c.refreshDataSeeded(),
		func() tea.Msg {
			return messages.SessionAttachRequestedMsg{
				SessionID:   m.SessionID,
				Warning:     m.Err,
				DashboardID: m.DashboardID,
				Generation:  m.Generation,
			}
		},
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
