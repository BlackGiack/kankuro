package app

import (
	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/ui/component"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

// This file is the coordinator's view plumbing: building the views,
// switching between them, sizing them, and fanning messages out to them. It
// is all coordinator-goroutine state (c.views, c.currentView, c.prevView,
// c.width, c.height): nothing here may be reached from a tea.Cmd closure.

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
	c.views[ViewConversationChoice] = views.NewConversationChoiceView(c.conversationChoiceCallbacks())
}

// navigateTo switches the current view and returns the new view's Init command.
//
// Never call it (directly or through a callback) from inside a view's
// Update. Coordinator.Update writes the view returned by updateCurrentView
// back into c.views[c.currentView] *after* the view's Update returns; if
// navigateTo ran during that Update, c.currentView has already changed and
// the old view is stored under the new view's key. Act 1 did exactly this
// once: the agent-switch view's OnSwitch callback navigated to the dashboard
// and the agent-switch view replaced the dashboard. Instead, return a Cmd
// that produces a NavigateMsg (or SessionAttachRequestedMsg) and navigate
// from a handler in handleSpecialMessages / handleSessionMessages, which
// return early and never write c.views.
//
// Leaving the dashboard also drops the preview subscription. navigateTo does
// not touch the error banner: that is cleared by the next keypress (see
// Update), so an ErrorMsg batched beside a NavigateMsg survives either order.
func (c *Coordinator) navigateTo(vt ViewType) tea.Cmd {
	if c.currentView == ViewDashboard && vt != ViewDashboard {
		c.preview.unsubscribe()
	}
	c.prevView = c.currentView
	c.currentView = vt
	if view, ok := c.views[vt]; ok {
		return view.Init()
	}
	return nil
}

func (c *Coordinator) resizeAll() {
	for _, view := range c.views {
		view.SetSize(c.width, c.height-views.CoordinatorRows)
	}
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

// currentPreviewSessionID reads the dashboard's active preview session. It
// must only be called from the main coordinator goroutine (Update/Init), not
// from inside a tea.Cmd closure - see refreshData.
func (c *Coordinator) currentPreviewSessionID() string {
	if dashboard, ok := c.views[ViewDashboard].(*views.Dashboard); ok {
		return dashboard.PreviewSessionID()
	}
	return ""
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
	case messages.ViewConversationChoice:
		return ViewConversationChoice
	default:
		return ViewDashboard
	}
}
