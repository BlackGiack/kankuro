package app

import (
	"errors"
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"

	"kankuro/internal/config"
	"kankuro/internal/session"
	"kankuro/internal/ui/commands"
	"kankuro/internal/ui/component"
	"kankuro/internal/ui/messages"
	"kankuro/internal/ui/styles"
	"kankuro/internal/ui/views"
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
	currentView ViewType
	prevView    ViewType

	views map[ViewType]component.Model

	repo    config.Repository
	session *session.Manager
	cmds    *commands.Commands

	width  int
	height int
	err    error

	settings config.AppSettings

	previewSubChan   chan []byte
	previewSessionID string
}

func NewCoordinator(repo config.Repository, sm *session.Manager) *Coordinator {
	cmds := WireCommands(repo, sm)

	c := &Coordinator{
		repo:        repo,
		session:     sm,
		cmds:        cmds,
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

// RefreshData triggers a data refresh.
func (c *Coordinator) RefreshData() tea.Cmd {
	return c.refreshData()
}

// Err returns the current error, if any.
func (c *Coordinator) Err() error {
	return c.err
}

func (c *Coordinator) initViews() {
	c.views[ViewDashboard] = views.NewDashboard(c.dashboardCallbacks(), c.cmds)
	cardEditor := views.NewCardEditor(c.cardEditorCallbacks(), c.cmds)
	cardEditor.SetSettings(c.settings)
	c.views[ViewCardEditor] = cardEditor

	smSkills := views.NewSourceManager(views.SourceTypeSkills, c.sourceManagerCallbacks(views.SourceTypeSkills), c.cmds)
	c.views[ViewSourceManagerSkills] = smSkills

	smPersonas := views.NewSourceManager(views.SourceTypePersonas, c.sourceManagerCallbacks(views.SourceTypePersonas), c.cmds)
	c.views[ViewSourceManagerPersonas] = smPersonas

	smWorkspaces := views.NewSourceManager(views.SourceTypeWorkspaces, c.sourceManagerCallbacks(views.SourceTypeWorkspaces), c.cmds)
	c.views[ViewSourceManagerWorkspaces] = smWorkspaces

	smPuppets := views.NewSourceManager(views.SourceTypePuppets, c.sourceManagerCallbacks(views.SourceTypePuppets), c.cmds)
	c.views[ViewSourceManagerPuppets] = smPuppets

	settingsView := views.NewSettingsView(c.settingsCallbacks(), c.cmds)
	settingsView.SetSettings(c.settings)
	c.views[ViewSettings] = settingsView
	c.views[ViewMCPForm] = views.NewMCPForm(c.mcpCallbacks(), c.cmds)
	c.views[ViewSessionLauncher] = views.NewSessionLauncher(c.sessionLauncherCallbacks(), c.cmds)
	c.views[ViewAgentSwitch] = views.NewAgentSwitchView(c.agentSwitchCallbacks())
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
		c.refreshData(),
		tick(),
		c.syncPuppetsAsync(),
		c.reportSessionLoadErrors(),
	)
}

// reportSessionLoadErrors surfaces any error hit while hydrating persisted
// sessions at startup (corrupt/unreadable session files are skipped, not
// fatal, but the user should know they lost a session record).
func (c *Coordinator) reportSessionLoadErrors() tea.Cmd {
	return func() tea.Msg {
		if err := c.session.LoadErrors(); err != nil {
			return messages.ErrorMsg{Err: err}
		}
		return nil
	}
}

func (c *Coordinator) syncPuppetsAsync() tea.Cmd {
	return func() tea.Msg {
		_ = c.repo.SyncPuppetsFromSources()
		return messages.PuppetsSyncedMsg{}
	}
}

func (c *Coordinator) Update(msg tea.Msg) (*Coordinator, tea.Cmd) {
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
		return true, tea.Batch(c.refreshData(), tick())

	case messages.DataRefreshRequestedMsg:
		return true, c.refreshData()

	case messages.PuppetsSyncedMsg:
		return true, c.refreshData()

	case messages.NavigateMsg:
		return true, c.navigateTo(convertViewType(m.View))

	case messages.NavigateBackMsg:
		return true, c.navigateTo(c.prevView)

	case messages.CardSavedMsg:
		if m.Err != nil {
			c.err = m.Err
			return true, nil
		}
		return true, tea.Batch(c.refreshData(), c.navigateTo(ViewDashboard))

	case messages.MCPSavedMsg:
		if m.Err != nil {
			c.err = m.Err
			return true, nil
		}
		// Refresh data to update MCP list in card editor
		return true, tea.Batch(c.refreshData(), c.navigateTo(c.prevView))

	case messages.SettingsSavedMsg:
		if m.Err != nil {
			c.err = m.Err
			return true, nil
		}
		c.settings = m.Settings
		if editor, ok := c.views[ViewCardEditor].(*views.CardEditor); ok {
			editor.SetSettings(m.Settings)
		}
		return true, tea.Batch(c.refreshData(), c.navigateTo(ViewDashboard))

	case messages.ErrorMsg:
		c.err = m.Err
		return true, nil

	case messages.MCPDeleteRequestedMsg:
		// Disk I/O deferred to a tea.Cmd, consistent with every other
		// mutation - this used to run synchronously inside Update.
		return true, func() tea.Msg {
			if err := c.repo.DeleteMCP(m.MCPID); err != nil {
				return messages.ErrorMsg{Err: err}
			}
			return messages.MCPDeletedMsg(m)
		}

	case messages.ClearErrorMsg:
		c.err = nil
		return true, nil

	case messages.IDELaunchedMsg:
		if m.Err != nil {
			c.err = m.Err
		}
		return true, nil
	}

	return c.handleSessionMessages(msg)
}

func (c *Coordinator) handleSessionMessages(msg tea.Msg) (bool, tea.Cmd) {
	switch m := msg.(type) {
	case messages.DataLoadedMsg:
		return true, c.handleDataLoaded(m)

	case messages.SessionPreviewRequestedMsg:
		return true, c.handlePreviewRequest(m.SessionID)

	case messages.SessionOutputMsg:
		// Only forward if it's the currently previewed session and we are
		// still subscribed - navigateTo clears both together, so a
		// SessionOutputMsg already queued for a session we've since
		// navigated away from cannot re-arm waitForOutput on a nil channel.
		if m.SessionID == c.previewSessionID && c.previewSubChan != nil {
			c.broadcastPreview(messages.PreviewDataMsg{
				SessionID:     m.SessionID,
				Scrollback:    m.Output,
				IsIncremental: true,
			})
			return true, c.waitForOutput(m.SessionID, c.previewSubChan)
		}
		return true, nil

	case messages.SessionLaunchCompletedMsg:
		launcher, ok := c.views[ViewSessionLauncher].(*views.SessionLauncher)
		if ok {
			launcher.SetLaunchResult(m.SessionID, m.Err)
		}
		return true, c.refreshData()

	case messages.AgentSwitchRequestedMsg:
		switchView, ok := c.views[ViewAgentSwitch].(*views.AgentSwitchView)
		if ok {
			switchView.SetSession(m.SessionID, m.SessionName)
		}
		return true, c.navigateTo(ViewAgentSwitch)

	case messages.AgentSwitchedMsg:
		return true, c.handleAgentSwitched(m)

	case messages.SessionAttachRequestedMsg:
		// Navigate here, not in the callback that produced this message:
		// this handler runs outside updateCurrentView's write-back
		// (coordinator.go's Update), so flipping c.currentView cannot
		// clobber c.views[c.currentView] with whatever view originated the
		// message (e.g. the agent-switch view calling this to attach after
		// a successful switch). See callbacks.go's agentSwitchCallbacks for
		// the regression this guards against.
		var initCmd tea.Cmd
		if c.currentView != ViewDashboard {
			initCmd = c.navigateTo(ViewDashboard)
		}
		// Set after navigating: navigateTo unconditionally clears c.err.
		if m.Warning != nil {
			c.err = m.Warning
		}
		return true, tea.Batch(initCmd, c.attachSession(m.SessionID))

	case messages.AttachFinishedMsg:
		if m.Err != nil {
			c.err = m.Err
		}
		return true, c.refreshData()
	}

	return false, nil
}

func (c *Coordinator) handleDataLoaded(m messages.DataLoadedMsg) tea.Cmd {
	if m.Err != nil {
		c.err = m.Err
	}
	c.broadcastData(m)
	if m.PreviewSessionID != "" {
		c.broadcastPreview(messages.PreviewDataMsg{
			SessionID:  m.PreviewSessionID,
			Scrollback: m.PreviewScrollback,
		})
	}
	// If the dashboard auto-selected a session (e.g. on first launch), subscribe immediately
	if dashboard, ok := c.views[ViewDashboard].(*views.Dashboard); ok {
		id := dashboard.PreviewSessionID()
		if id != "" && id != c.previewSessionID {
			return c.handlePreviewRequest(id)
		}
	}
	return nil
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
	// same path relaunch uses - instead of batching a separate ErrorMsg
	// beside a NavigateMsg: tea.Batch makes no ordering guarantee, and
	// navigateTo unconditionally clears c.err, so a batched warning was lost
	// whenever the navigate message won the race.
	//
	// This handler runs from handleSessionMessages, outside
	// updateCurrentView's write-back, so it - unlike the OnSwitch callback
	// that produced this message - is a safe place to hand off to
	// SessionAttachRequestedMsg, which does the actual navigate.
	return tea.Batch(
		c.refreshData(),
		func() tea.Msg {
			return messages.SessionAttachRequestedMsg{SessionID: m.SessionID, Warning: m.Err}
		},
	)
}

func (c *Coordinator) View() string {
	if c.err != nil {
		return styles.ErrorStyle.Render(c.err.Error()) + "\n\n" + c.currentViewView()
	}
	return c.currentViewView()
}

func (c *Coordinator) currentViewView() string {
	if view, ok := c.views[c.currentView]; ok {
		return view.View()
	}
	return "Unknown view"
}

func (c *Coordinator) updateCurrentView(msg tea.Msg) (component.Model, tea.Cmd) {
	view, ok := c.views[c.currentView]
	if !ok {
		return nil, nil
	}

	return view.Update(msg)
}

func (c *Coordinator) navigateTo(vt ViewType) tea.Cmd {
	if c.currentView == ViewDashboard && vt != ViewDashboard {
		if c.previewSubChan != nil {
			c.session.UnsubscribeOutput(c.previewSessionID, c.previewSubChan)
			c.previewSubChan = nil
		}
		// Clear alongside the unsubscribe so a SessionOutputMsg already
		// queued for the session we just left cannot still match
		// m.SessionID == c.previewSessionID after we've navigated away (see
		// debt_impl_plan_4.md #12/#17).
		c.previewSessionID = ""
	}
	c.prevView = c.currentView
	c.currentView = vt
	c.err = nil
	if view, ok := c.views[vt]; ok {
		return view.Init()
	}
	return nil
}

const heightAdjustment = 4

func (c *Coordinator) resizeAll() {
	for _, view := range c.views {
		view.SetSize(c.width, c.height-heightAdjustment)
	}
}

// currentPreviewSessionID reads the dashboard's active preview session. It
// must only be called from the main coordinator goroutine (Update/Init), not
// from inside a tea.Cmd closure - see refreshData.
func (c *Coordinator) currentPreviewSessionID() string {
	if dashboard, ok := c.views[ViewDashboard].(*views.Dashboard); ok {
		return dashboard.PreviewSessionID()
	}
	return ""
}

func (c *Coordinator) refreshData() tea.Cmd {
	// Snapshot coordinator-goroutine state (c.views) now, on the caller's
	// goroutine. The returned closure runs on its own goroutine (it's a
	// tea.Cmd) and must never touch c.views/c.previewSessionID/c.currentView
	// directly - see debt_impl_plan_4.md step 1. c.session and c.repo are
	// safe to call from either goroutine.
	previewID := c.currentPreviewSessionID()

	return func() tea.Msg {
		var errs []error

		cards, err := c.repo.ListIDCards()
		if err != nil {
			errs = append(errs, fmt.Errorf("list cards: %w", err))
			cards = nil
		}
		sessions := c.session.List()
		skills, err := c.repo.ListSkills()
		if err != nil {
			errs = append(errs, fmt.Errorf("list skills: %w", err))
			skills = nil
		}
		personas, err := c.repo.ListPersonas()
		if err != nil {
			errs = append(errs, fmt.Errorf("list personas: %w", err))
			personas = nil
		}
		mcps, err := c.repo.ListMCPs()
		if err != nil {
			errs = append(errs, fmt.Errorf("list mcps: %w", err))
			mcps = nil
		}
		workspaces, err := c.repo.ListWorkspaces()
		if err != nil {
			errs = append(errs, fmt.Errorf("list workspaces: %w", err))
			workspaces = nil
		}
		puppets, err := c.repo.ListPuppets()
		if err != nil {
			errs = append(errs, fmt.Errorf("list puppets: %w", err))
			puppets = nil
		}

		var scrollback []byte
		if previewID != "" {
			scrollback = c.session.GetScrollback(previewID)
		}

		return messages.DataLoadedMsg{
			Cards:             cards,
			Sessions:          sessions,
			Skills:            skills,
			Personas:          personas,
			MCPs:              mcps,
			Workspaces:        workspaces,
			Puppets:           puppets,
			PreviewSessionID:  previewID,
			PreviewScrollback: scrollback,
			Err:               errors.Join(errs...),
		}
	}
}

func (c *Coordinator) broadcastData(msg messages.DataLoadedMsg) {
	for _, view := range c.views {
		if dataReceiver, ok := view.(messages.DataReceiver); ok {
			dataReceiver.SetData(msg)
		}
	}
}

// broadcastPreview delivers a partial, preview-only update to the views that
// render a live session preview. Unlike broadcastData, this never reaches
// DataReceiver.SetData - see messages.PreviewDataMsg for why the two must
// stay separate.
func (c *Coordinator) broadcastPreview(msg messages.PreviewDataMsg) {
	for _, view := range c.views {
		if previewReceiver, ok := view.(messages.PreviewReceiver); ok {
			previewReceiver.SetPreviewData(msg)
		}
	}
}

func (c *Coordinator) handlePreviewRequest(sessionID string) tea.Cmd {
	// Unsubscribe from old session
	if c.previewSubChan != nil {
		c.session.UnsubscribeOutput(c.previewSessionID, c.previewSubChan)
		c.previewSubChan = nil
	}

	c.previewSessionID = sessionID

	// Immediate scrollback fetch
	scrollback := c.session.GetScrollback(sessionID)
	c.broadcastPreview(messages.PreviewDataMsg{
		SessionID:  sessionID,
		Scrollback: scrollback,
	})

	// Subscribe to new session
	ch, err := c.session.SubscribeOutput(sessionID)
	if err == nil {
		c.previewSubChan = ch
		return c.waitForOutput(sessionID, ch)
	}

	return nil
}

func (c *Coordinator) waitForOutput(id string, ch chan []byte) tea.Cmd {
	if ch == nil {
		return nil
	}
	return func() tea.Msg {
		output, ok := <-ch
		if !ok {
			return nil
		}
		return messages.SessionOutputMsg{
			SessionID: id,
			Output:    output,
		}
	}
}

func tick() tea.Cmd {
	const tickInterval = time.Second * 2
	return tea.Tick(tickInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type tickMsg time.Time

func convertViewType(vt messages.ViewType) ViewType {
	switch vt {
	case messages.ViewDashboard:
		return ViewDashboard
	case messages.ViewCardEditor:
		return ViewCardEditor
	case messages.ViewSourceManagerSkills:
		return ViewSourceManagerSkills
	case messages.ViewSourceManagerPersonas:
		return ViewSourceManagerPersonas
	case messages.ViewSourceManagerWorkspaces:
		return ViewSourceManagerWorkspaces
	case messages.ViewSourceManagerPuppets:
		return ViewSourceManagerPuppets
	case messages.ViewMCPForm:
		return ViewMCPForm
	case messages.ViewSettings:
		return ViewSettings
	case messages.ViewSessionLauncher:
		return ViewSessionLauncher
	case messages.ViewAgentSwitch:
		return ViewAgentSwitch
	default:
		return ViewDashboard
	}
}
