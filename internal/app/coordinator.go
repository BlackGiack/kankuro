package app

import (
	"errors"

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
)

type Coordinator struct {
	// --- coordinator-goroutine state: never read or written from a tea.Cmd closure ---
	currentView ViewType
	prevView    ViewType
	views       map[ViewType]component.Model
	width       int
	height      int
	err         error
	settings    config.AppSettings
	preview     previewSub

	// configFingerprint is the mtime+size marker of every registry file as
	// of the last config load, used by the idle tick to skip re-reading
	// unchanged YAML - see maybeRefreshConfig and debt_impl_plan_5.md step
	// 1. Seeded by refreshDataSeeded (Init and every mutation path), so it
	// is nil only until that first result arrives, or if Fingerprint itself
	// failed. maybeRefreshConfig hands it to the refresher by value.
	configFingerprint map[string]config.FileStamp

	// --- safe from any goroutine: concurrency-safe handles, and helpers that
	// hold no coordinator pointer ---
	repo    config.Repository
	session *session.Manager
	refresh refresher
	actions sessionActions
}

func NewCoordinator(repo config.Repository, sm *session.Manager) *Coordinator {
	c := &Coordinator{
		repo:        repo,
		session:     sm,
		preview:     previewSub{sessions: sm},
		refresh:     refresher{repo: repo, sessions: sm},
		actions:     sessionActions{repo: repo, sessions: sm, service: api.New(repo, sm)},
		views:       make(map[ViewType]component.Model),
		currentView: ViewDashboard,
	}

	settings, err := repo.GetAppSettings()
	if err != nil {
		settings = config.DefaultAppSettings()
	}
	c.settings = settings

	c.initViews()
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

// NavigateTo changes the current view.
func (c *Coordinator) NavigateTo(vt ViewType) tea.Cmd {
	return c.navigateTo(vt)
}

// Width returns the coordinator width.
func (c *Coordinator) Width() int {
	return c.width
}

// Height returns the coordinator height.
func (c *Coordinator) Height() int {
	return c.height
}

// ResizeAll triggers a resize of all views.
func (c *Coordinator) ResizeAll() {
	c.resizeAll()
}

// DashboardCallbacks returns the dashboard callbacks.
func (c *Coordinator) DashboardCallbacks() views.DashboardCallbacks {
	return c.dashboardCallbacks()
}

// SourceManagerCallbacks returns the source manager callbacks for a given type.
func (c *Coordinator) SourceManagerCallbacks(st views.SourceType) views.SourceManagerCallbacks {
	return c.sourceManagerCallbacks(st)
}

// AgentSwitchCallbacks returns the agent-switch view callbacks.
func (c *Coordinator) AgentSwitchCallbacks() views.AgentSwitchCallbacks {
	return c.agentSwitchCallbacks()
}

// RefreshData triggers a data refresh. It is a test helper: real callers
// inside the coordinator use refreshDataSeeded, which additionally seeds the
// idle tick's fingerprint baseline (see debt_impl_plan_5.md review finding
// 2 in review_dp_5.md). This wraps the unseeded refreshData so tests get one
// deterministic DataLoadedMsg back without also driving a configFingerprintMsg.
func (c *Coordinator) RefreshData() tea.Cmd {
	return c.refreshData()
}

// Err returns the current error, if any.
func (c *Coordinator) Err() error {
	return c.err
}

func (c *Coordinator) Init() tea.Cmd {
	// Also run the initial view's own Init (the dashboard's, at startup):
	// navigating between views already does this via navigateTo, but
	// nothing previously ran it for the view Coordinator starts on, so
	// Dashboard.Init's transition-only work (disabling the Kitty keyboard
	// protocol left on by a previous run that died mid-attach; see
	// review_dp_4.md finding 2) never fired at cold start. This also
	// belatedly runs cardList/sessionList/preview.Init(), which had the same
	// gap.
	var initialViewCmd tea.Cmd
	if view, ok := c.views[c.currentView]; ok {
		initialViewCmd = view.Init()
	}
	return tea.Batch(
		initialViewCmd,
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

func (c *Coordinator) handleSpecialMessages(msg tea.Msg) (bool, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		c.width = m.Width
		c.height = m.Height
		c.resizeAll()
		return true, nil

	case tickMsg:
		// Sessions (in-memory) refresh every tick; config (YAML on disk)
		// only refreshes when maybeRefreshConfig's mtime check says a
		// registry file actually changed - see debt_impl_plan_5.md step 1.
		return true, tea.Batch(c.refreshSessions(), c.maybeRefreshConfig(), tick())

	case configFingerprintMsg:
		return true, c.handleConfigFingerprint(m)

	case messages.DataRefreshRequestedMsg:
		return true, c.refreshDataSeeded()

	case messages.PuppetsSyncedMsg:
		return true, c.refreshDataSeeded()

	case messages.NavigateMsg:
		return true, c.navigateTo(convertViewType(m.View))

	case messages.NavigateBackMsg:
		return true, c.navigateTo(c.prevView)

	case messages.CardSavedMsg:
		if m.Err != nil {
			c.err = m.Err
			return true, nil
		}
		return true, tea.Batch(c.refreshDataSeeded(), c.navigateTo(ViewDashboard))

	case messages.MCPSavedMsg:
		if m.Err != nil {
			c.err = m.Err
			return true, nil
		}
		// Refresh data to update MCP list in card editor
		return true, tea.Batch(c.refreshDataSeeded(), c.navigateTo(c.prevView))

	case messages.SettingsSavedMsg:
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
		return true, nil

	case messages.MCPDeleteRequestedMsg:
		// Disk I/O deferred to a tea.Cmd, consistent with every other
		// mutation - this used to run synchronously inside Update.
		return true, c.actions.deleteMCP(m)

	case messages.ClearErrorMsg:
		c.err = nil
		return true, nil

	case messages.IDELaunchedMsg:
		return c.handleWorkspaceCommandResult(m.Err)
	case messages.TerminalLaunchedMsg:
		return c.handleWorkspaceCommandResult(m.Err)
	}

	return c.handleSessionMessages(msg)
}

func (c *Coordinator) handleWorkspaceCommandResult(err error) (bool, tea.Cmd) {
	if err != nil {
		c.err = err
	}
	return true, nil
}

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
		launcher, ok := c.views[ViewSessionLauncher].(*views.SessionLauncher)
		if ok {
			launcher.SetLaunchResult(m.SessionID, m.Err)
		}
		return true, c.refreshDataSeeded()

	case messages.AgentSwitchRequestedMsg:
		switchView, ok := c.views[ViewAgentSwitch].(*views.AgentSwitchView)
		if ok {
			switchView.SetSession(m.SessionID, m.SessionName)
		}
		return true, c.navigateTo(ViewAgentSwitch)

	case messages.AgentSwitchedMsg:
		return true, c.handleAgentSwitched(m)

	case messages.SessionAttachRequestedMsg:
		// Navigate here, not in the callback that produced this message
		// (invariant 2, see navigateTo's doc comment).
		var initCmd tea.Cmd
		if c.currentView != ViewDashboard {
			initCmd = c.navigateTo(ViewDashboard)
		}
		if m.Warning != nil {
			c.err = m.Warning
		}
		return true, tea.Batch(initCmd, c.attachSession(m.SessionID))

	case messages.AttachFinishedMsg:
		if m.Err != nil {
			c.err = m.Err
		}
		return true, c.refreshDataSeeded()
	}

	return false, nil
}

func (c *Coordinator) handleAgentSwitched(m messages.AgentSwitchedMsg) tea.Cmd {
	if m.Err != nil && !errors.Is(m.Err, session.ErrPersistFailed) {
		c.err = m.Err
		return nil
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
			return messages.SessionAttachRequestedMsg{SessionID: m.SessionID, Warning: m.Err}
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
