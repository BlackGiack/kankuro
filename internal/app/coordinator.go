package app

import (
	"errors"
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"

	"kankuro/internal/config"
	"kankuro/internal/session"
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

	width  int
	height int
	err    error

	settings config.AppSettings

	previewSubChan   chan []byte
	previewSessionID string

	// configFingerprint is the mtime+size marker of every registry file as
	// of the last config load, used by the idle tick to skip re-reading
	// unchanged YAML - see maybeRefreshConfig and debt_impl_plan_5.md step
	// 1. Seeded by refreshDataSeeded (Init and every mutation path), so it
	// is nil only until that first result arrives, or if Fingerprint itself
	// failed.
	configFingerprint map[string]config.FileStamp
}

func NewCoordinator(repo config.Repository, sm *session.Manager) *Coordinator {
	c := &Coordinator{
		repo:        repo,
		session:     sm,
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

func (c *Coordinator) initViews() {
	c.views[ViewDashboard] = views.NewDashboard(c.dashboardCallbacks())
	cardEditor := views.NewCardEditor(c.cardEditorCallbacks())
	cardEditor.SetSettings(c.settings)
	c.views[ViewCardEditor] = cardEditor

	smSkills := views.NewSourceManager(views.SourceTypeSkills, c.sourceManagerCallbacks(views.SourceTypeSkills))
	c.views[ViewSourceManagerSkills] = smSkills

	smPersonas := views.NewSourceManager(views.SourceTypePersonas, c.sourceManagerCallbacks(views.SourceTypePersonas))
	c.views[ViewSourceManagerPersonas] = smPersonas

	smWorkspaces := views.NewSourceManager(views.SourceTypeWorkspaces, c.sourceManagerCallbacks(views.SourceTypeWorkspaces))
	c.views[ViewSourceManagerWorkspaces] = smWorkspaces

	smPuppets := views.NewSourceManager(views.SourceTypePuppets, c.sourceManagerCallbacks(views.SourceTypePuppets))
	c.views[ViewSourceManagerPuppets] = smPuppets

	settingsView := views.NewSettingsView(c.settingsCallbacks())
	settingsView.SetSettings(c.settings)
	c.views[ViewSettings] = settingsView
	c.views[ViewMCPForm] = views.NewMCPForm(c.mcpCallbacks())
	c.views[ViewSessionLauncher] = views.NewSessionLauncher(c.sessionLauncherCallbacks())
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
		c.refreshDataSeeded(),
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

	case messages.SessionsRefreshedMsg:
		return true, c.handleSessionsRefreshed(m)

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
		return true, c.refreshDataSeeded()
	}

	return false, nil
}

// handleConfigFingerprint applies a config-fingerprint check's result. It
// only adopts m.stamps when non-nil: maybeRefreshConfig's "unchanged" branch
// sends a nil stamps map specifically so it does NOT touch the baseline (see
// debt_impl_plan_5.md review finding 1 in review_dp_5_2.md). Without this, an
// in-flight "unchanged" tick result racing a refreshDataSeeded from a
// mutation could stomp the newer baseline the mutation just set, costing one
// redundant full reload on the following tick - not a correctness bug (data
// stays right), just the exact double-parse finding 2 (in review_dp_5.md)
// removed. A nil m.data means either "nothing changed" or "Fingerprint
// itself failed", so it never feeds a data reload either way.
func (c *Coordinator) handleConfigFingerprint(m configFingerprintMsg) tea.Cmd {
	if m.stamps != nil {
		c.configFingerprint = m.stamps
	}
	if m.data == nil {
		return nil
	}
	return c.handleDataLoaded(*m.data)
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
	return c.maybeSubscribeAutoSelectedPreview()
}

// handleSessionsRefreshed applies the idle tick's sessions-only refresh: it
// updates only the dashboard (the sole consumer of Sessions), not every
// DataReceiver - see refreshSessions and debt_impl_plan_5.md review finding
// 1.
func (c *Coordinator) handleSessionsRefreshed(m messages.SessionsRefreshedMsg) tea.Cmd {
	if dashboard, ok := c.views[ViewDashboard].(*views.Dashboard); ok {
		dashboard.SetSessions(m.Sessions)
	}
	if m.PreviewSessionID != "" {
		c.broadcastPreview(messages.PreviewDataMsg{
			SessionID:  m.PreviewSessionID,
			Scrollback: m.PreviewScrollback,
		})
	}
	return c.maybeSubscribeAutoSelectedPreview()
}

// maybeSubscribeAutoSelectedPreview subscribes to a session the dashboard
// auto-selected on its own (e.g. the first session on a fresh launch),
// shared by handleDataLoaded and handleSessionsRefreshed.
func (c *Coordinator) maybeSubscribeAutoSelectedPreview() tea.Cmd {
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
		c.refreshDataSeeded(),
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

// refreshData performs a full, unconditional refresh of every registry plus
// the session list and preview scrollback in one DataLoadedMsg, without
// touching c.configFingerprint. It exists solely for RefreshData(), the
// exported wrapper integration tests drive to get one deterministic
// DataLoadedMsg back - see refreshDataSeeded for the internal equivalent
// every real call site uses, which also seeds the fingerprint baseline.
func (c *Coordinator) refreshData() tea.Cmd {
	// Snapshot coordinator-goroutine state (c.views) now, on the caller's
	// goroutine. The returned closure runs on its own goroutine (it's a
	// tea.Cmd) and must never touch c.views/c.previewSessionID/c.currentView
	// directly - see debt_impl_plan_4.md step 1. c.session and c.repo are
	// safe to call from either goroutine.
	previewID := c.currentPreviewSessionID()

	return func() tea.Msg {
		return c.loadFullData(previewID)
	}
}

// refreshDataSeeded is the combined refresh every mutation-triggered path
// (Init, DataRefreshRequestedMsg, CardSavedMsg, ...) actually uses. Unlike
// refreshData, it also stats the registries (same as maybeRefreshConfig)
// and wraps the result in configFingerprintMsg so the coordinator adopts the
// new baseline on the main goroutine. Without this, a full refresh left
// c.configFingerprint stale, so the very next idle tick always saw a
// "changed" stamp and reloaded everything a second time - every mutation was
// parsed twice - see debt_impl_plan_5.md review finding 2.
func (c *Coordinator) refreshDataSeeded() tea.Cmd {
	previewID := c.currentPreviewSessionID()

	return func() tea.Msg {
		stamps, err := c.repo.Fingerprint()
		data := c.loadFullData(previewID)
		if err != nil {
			// Fingerprint itself failed - don't block the refresh on it,
			// just skip seeding the baseline; the next tick falls back to
			// one extra reload, same as before this fix.
			return data
		}
		return configFingerprintMsg{stamps: stamps, data: &data}
	}
}

// loadFullData reads every config registry plus the session list and
// preview scrollback into one DataLoadedMsg. It performs blocking disk I/O
// and must only be called from inside a tea.Cmd closure.
func (c *Coordinator) loadFullData(previewID string) messages.DataLoadedMsg {
	configMsg := c.loadConfigData()

	sessions := c.session.List()
	var scrollback []byte
	if previewID != "" {
		scrollback = c.session.GetScrollback(previewID)
	}

	configMsg.Sessions = sessions
	configMsg.PreviewSessionID = previewID
	configMsg.PreviewScrollback = scrollback
	return configMsg
}

// refreshSessions reads only the (in-memory) session list and the current
// preview's scrollback - no disk I/O - and returns it as a
// SessionsRefreshedMsg, not a DataLoadedMsg: the latter would broadcast to
// every DataReceiver, including config-registry views like SourceManager,
// whose SetData reloads source files from disk regardless of what changed
// (see debt_impl_plan_5.md review finding 1). This is what the idle tick
// calls every 2s.
func (c *Coordinator) refreshSessions() tea.Cmd {
	previewID := c.currentPreviewSessionID()

	return func() tea.Msg {
		sessions := c.session.List()
		var scrollback []byte
		if previewID != "" {
			scrollback = c.session.GetScrollback(previewID)
		}
		return messages.SessionsRefreshedMsg{
			Sessions:          sessions,
			PreviewSessionID:  previewID,
			PreviewScrollback: scrollback,
		}
	}
}

// configFingerprintMsg reports the result of a fingerprint check. stamps is
// only non-nil when the baseline should actually change - the "unchanged"
// result from maybeRefreshConfig deliberately leaves it nil so it never
// clobbers a newer baseline a concurrent refreshDataSeeded may have just set
// (see handleConfigFingerprint). data is only non-nil when a registry
// actually changed and was re-read.
type configFingerprintMsg struct {
	stamps map[string]config.FileStamp
	data   *messages.DataLoadedMsg
}

// maybeRefreshConfig is the idle tick's config half: it stats every registry
// file (cheap - no parsing) and only re-reads/re-parses/re-sorts them if at
// least one changed since the last check. See debt_impl_plan_5.md step 1 and
// config.FileRepository.Fingerprint for the mtime-precision caveat this
// accepts.
func (c *Coordinator) maybeRefreshConfig() tea.Cmd {
	prev := c.configFingerprint

	return func() tea.Msg {
		stamps, err := c.repo.Fingerprint()
		if err == nil && fingerprintsEqual(prev, stamps) {
			// Nothing changed - leave stamps nil so this result cannot
			// stomp a newer baseline a concurrent mutation-triggered
			// refreshDataSeeded may have just set (review_dp_5_2.md
			// finding 1).
			return configFingerprintMsg{}
		}
		data := c.loadConfigData()
		return configFingerprintMsg{stamps: stamps, data: &data}
	}
}

func fingerprintsEqual(a, b map[string]config.FileStamp) bool {
	if a == nil || b == nil {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	for name, stampA := range a {
		stampB, ok := b[name]
		if !ok || !stampA.ModTime.Equal(stampB.ModTime) || stampA.Size != stampB.Size {
			return false
		}
	}
	return true
}

// loadConfigData reads every config registry once and returns a
// DataLoadedMsg carrying just those fields (Sessions/PreviewSessionID/
// PreviewScrollback left zero) - the shared core of loadFullData (used by
// refreshData and refreshDataSeeded) and maybeRefreshConfig. It performs
// blocking disk I/O and must only be called from inside a tea.Cmd closure,
// never on the main coordinator goroutine.
func (c *Coordinator) loadConfigData() messages.DataLoadedMsg {
	var errs []error

	cards, err := c.repo.ListIDCards()
	if err != nil {
		errs = append(errs, fmt.Errorf("list cards: %w", err))
		cards = nil
	}
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

	return messages.DataLoadedMsg{
		Cards:      cards,
		Skills:     skills,
		Personas:   personas,
		MCPs:       mcps,
		Workspaces: workspaces,
		Puppets:    puppets,
		Err:        errors.Join(errs...),
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
