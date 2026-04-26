package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"kankuro/internal/config"
	"kankuro/internal/session"
	"kankuro/internal/ui/commands"
	"kankuro/internal/ui/component"
	"kankuro/internal/ui/messages"
	"kankuro/internal/ui/views"
)

type ViewType int

const (
	ViewDashboard ViewType = iota
	ViewCardEditor
	ViewSourceManagerSkills
	ViewSourceManagerPersonas
	ViewSourceManagerWorkspaces
	ViewIDEConfig
	ViewMCPForm
	ViewSessionLauncher
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

	ideConfig config.IDEConfig

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

	// Load IDE config
	ideCfg, err := repo.GetIDEConfig()
	if err != nil {
		ideCfg = config.IDEConfig{}
	}
	c.ideConfig = ideCfg

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
	c.views[ViewCardEditor] = views.NewCardEditor(c.cardEditorCallbacks(), c.cmds)

	smSkills := views.NewSourceManager(views.SourceTypeSkills, c.sourceManagerCallbacks(views.SourceTypeSkills), c.cmds)
	c.views[ViewSourceManagerSkills] = smSkills

	smPersonas := views.NewSourceManager(views.SourceTypePersonas, c.sourceManagerCallbacks(views.SourceTypePersonas), c.cmds)
	c.views[ViewSourceManagerPersonas] = smPersonas

	smWorkspaces := views.NewSourceManager(views.SourceTypeWorkspaces, c.sourceManagerCallbacks(views.SourceTypeWorkspaces), c.cmds)
	c.views[ViewSourceManagerWorkspaces] = smWorkspaces

	ideView := views.NewIDEConfig(c.ideCallbacks(), c.cmds)
	ideView.SetConfig(c.ideConfig)
	c.views[ViewIDEConfig] = ideView
	c.views[ViewMCPForm] = views.NewMCPForm(c.mcpCallbacks(), c.cmds)
	c.views[ViewSessionLauncher] = views.NewSessionLauncher(c.sessionLauncherCallbacks(), c.cmds)
}

func (c *Coordinator) Init() tea.Cmd {
	return tea.Batch(
		c.refreshData(),
		tick(),
	)
}

func (c *Coordinator) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		c.width = m.Width
		c.height = m.Height
		c.resizeAll()
		return c, nil

	case tickMsg:
		return c, tea.Batch(c.refreshData(), tick())

	case messages.DataLoadedMsg:
		c.broadcastData(m)
		// If the dashboard auto-selected a session (e.g. on first launch), subscribe immediately
		if dashboard, ok := c.views[ViewDashboard].(*views.Dashboard); ok {
			id := dashboard.PreviewSessionID()
			if id != "" && id != c.previewSessionID {
				return c, c.handlePreviewRequest(id)
			}
		}
		return c, nil

	case messages.SessionPreviewRequestedMsg:
		return c, c.handlePreviewRequest(m.SessionID)

	case messages.SessionOutputMsg:
		// Only forward if it's the currently previewed session
		if m.SessionID == c.previewSessionID {
			c.broadcastData(messages.DataLoadedMsg{
				PreviewSessionID:  m.SessionID,
				PreviewScrollback: m.Output,
				IsIncremental:     true,
			})
			return c, c.waitForOutput(m.SessionID, c.previewSubChan)
		}
		return c, nil

	case messages.NavigateMsg:
		return c, c.navigateTo(convertViewType(m.View))

	case messages.NavigateBackMsg:
		return c, c.navigateTo(c.prevView)

	case messages.ErrorMsg:
		c.err = m.Err
		return c, nil

	case messages.SessionLaunchCompletedMsg:
		launcher, ok := c.views[ViewSessionLauncher].(*views.SessionLauncher)
		if ok {
			launcher.SetLaunchResult(m.SessionID, m.Err)
		}
		return c, c.refreshData()

	case messages.SessionAttachRequestedMsg:
		return c, c.attachSession(m.SessionID)

	case messages.AttachFinishedMsg:
		if m.Err != nil {
			c.err = m.Err
		}
		return c, c.refreshData()

	case messages.MCPDeleteRequestedMsg:
		err := c.repo.DeleteMCP(m.MCPID)
		if err != nil {
			c.err = err
			return c, nil
		}
		// Emit message to notify card editor
		return c, func() tea.Msg {
			return messages.MCPDeletedMsg(m)
		}

	case messages.MCPUpdatedMsg:
		// Refresh data to update MCP list in card editor
		return c, c.refreshData()

	case messages.ClearErrorMsg:
		c.err = nil
		return c, nil
	}

	updated, cmd := c.updateCurrentView(msg)
	if updated != nil {
		c.views[c.currentView] = updated
	}
	return c, cmd
}

func (c *Coordinator) View() string {
	if c.err != nil {
		return views.ErrorStyle.Render(c.err.Error()) + "\n\n" + c.currentViewView()
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

func (c *Coordinator) refreshData() tea.Cmd {
	return func() tea.Msg {
		cards, _ := c.repo.ListIDCards()
		sessions := c.session.List()
		skills, _ := c.repo.ListSkills()
		personas, _ := c.repo.ListPersonas()
		mcps, _ := c.repo.ListMCPs()
		workspaces, _ := c.repo.ListWorkspaces()

		var previewID string
		var scrollback []byte
		if dashboard, ok := c.views[ViewDashboard].(*views.Dashboard); ok {
			previewID = dashboard.PreviewSessionID()
			if previewID != "" {
				scrollback = c.session.GetScrollback(previewID)
			}
		}

		return messages.DataLoadedMsg{
			Cards:             cards,
			Sessions:          sessions,
			Skills:            skills,
			Personas:          personas,
			MCPs:              mcps,
			Workspaces:        workspaces,
			PreviewSessionID:  previewID,
			PreviewScrollback: scrollback,
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

func (c *Coordinator) handlePreviewRequest(sessionID string) tea.Cmd {
	// Unsubscribe from old session
	if c.previewSubChan != nil {
		c.session.UnsubscribeOutput(c.previewSessionID, c.previewSubChan)
		c.previewSubChan = nil
	}

	c.previewSessionID = sessionID

	// Immediate scrollback fetch
	scrollback := c.session.GetScrollback(sessionID)
	c.broadcastData(messages.DataLoadedMsg{
		PreviewSessionID:  sessionID,
		PreviewScrollback: scrollback,
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
	case messages.ViewMCPForm:
		return ViewMCPForm
	case messages.ViewIDEConfig:
		return ViewIDEConfig
	case messages.ViewSessionLauncher:
		return ViewSessionLauncher
	default:
		return ViewDashboard
	}
}
