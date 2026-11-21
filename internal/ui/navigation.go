package ui

import (
	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/ui/component"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/models"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

// This file is the coordinator's view plumbing: building the views,
// switching between them, sizing them, and fanning messages out to them. It
// is all coordinator-goroutine state (c.views, c.currentView, c.prevView,
// c.width, c.height): nothing here may be reached from a tea.Cmd closure.

func (c *Coordinator) initViews() {
	c.views[ViewDashboard] = views.NewDashboard(c.dashboardCallbacks())
	cardEditor := views.NewCardEditor(c.cardEditorCallbacks())
	cardEditor.SetSettings(editorSettings(c.settings.ShowPersonasInCardEditor, c.settings.ShowMCPsInCardEditor))
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
	c.views[ViewMCPManager] = views.NewMCPManager()
	c.views[ViewMCPForm] = views.NewMCPForm(c.mcpCallbacks())
	c.views[ViewSessionLauncher] = views.NewSessionLauncher(c.sessionLauncherCallbacks())
	c.views[ViewAgentSwitch] = views.NewAgentSwitchView(c.agentSwitchCallbacks())
	c.views[ViewConversationChoice] = views.NewConversationChoiceView(c.conversationChoiceCallbacks())
}

func (c *Coordinator) cancelConversationRead() {
	c.conversationEpoch++
	c.lifecycleRequest++
	c.conversationReading = false
	c.lifecycleExecuting = false
	if c.conversationCancel != nil {
		c.conversationCancel()
		c.conversationCancel = nil
	}
}

func (c *Coordinator) invalidateSessionRequestsOnNavigation(vt ViewType) {
	if c.currentView == ViewDashboard && vt != ViewDashboard {
		c.renameRequest++
		c.workspaceRequest++
		c.mutationRequest++
		if vt != ViewSessionLauncher {
			c.launchRequest++
		}
	}
	switch {
	case c.currentView == ViewConversationChoice && vt != ViewConversationChoice:
		c.cancelConversationRead()
	case c.currentView == ViewDashboard && vt != ViewDashboard && vt != ViewConversationChoice:
		c.cancelConversationRead()
	case c.currentView == ViewAgentSwitch && vt != ViewAgentSwitch && vt != ViewConversationChoice:
		c.cancelConversationRead()
	}
	if c.currentView == ViewSessionLauncher && vt != ViewSessionLauncher && vt != ViewSourceManagerWorkspaces {
		c.renameRequest++
		c.launchRequest++
	}
}

// navigateTo switches the current view and returns the new view's Init command.
//
// Never call it (directly or through a callback) from inside a view's
// Update. Coordinator.Update writes the view returned by updateCurrentView
// back into c.views[c.currentView] *after* the view's Update returns; if
// navigateTo ran during that Update, c.currentView has already changed and
// the old view is stored under the new view's key. Instead, return a Cmd
// that produces a NavigateMsg (or SessionAttachRequestedMsg) and navigate
// from a handler in handleSpecialMessages / handleSessionMessages, which
// return early and never write c.views.
//
// Leaving the dashboard also releases the selected-session center. navigateTo does
// not touch the error banner: that is cleared by the next keypress (see
// Update), so an ErrorMsg batched beside a NavigateMsg survives either order.
func (c *Coordinator) navigateTo(vt ViewType) tea.Cmd {
	c.invalidateDeletionVisit(vt)
	if c.currentView != vt && c.attachmentOwner != nil {
		c.attachmentOwner.clear()
	}
	c.invalidateSessionRequestsOnNavigation(vt)

	return c.navigateToView(vt)
}

func (c *Coordinator) invalidateDeletionVisit(vt ViewType) {
	if c.currentView != vt {
		c.deletionVisit++
		c.dashboardDeleting = false
		c.cardDeleting = false
		if dashboard, ok := c.views[ViewDashboard].(*views.Dashboard); ok && c.currentView == ViewDashboard {
			dashboard.CancelCardDeletionImpact()
		}
		if list, ok := c.views[ViewDashboardList].(*views.DashboardList); ok && c.currentView == ViewDashboardList {
			list.CancelDeletionInteraction()
		}
	}
}

func (c *Coordinator) navigateToView(vt ViewType) tea.Cmd {
	if c.currentView == ViewMCPManager && vt != ViewMCPManager {
		if manager, ok := c.views[ViewMCPManager].(*views.MCPManager); ok {
			manager.Leave()
		}
	}
	if vt == ViewMCPManager {
		c.mcpManagerVisit++
		if manager, ok := c.views[ViewMCPManager].(*views.MCPManager); ok {
			manager.SetScope(messages.MCPManagerScope{
				DashboardID: c.activeDashboard, Generation: c.viewGeneration, Visit: c.mcpManagerVisit,
			})
		}
	}
	if vt == ViewSettings {
		c.settingsVisit++
	}
	if vt == ViewMCPForm {
		c.mcpSaving = false
		c.mcpReturn = c.currentView
		c.mcpRequest++
	}
	if c.currentView == ViewDashboard && vt != ViewDashboard && vt != ViewCardEditor {
		c.editorEpoch++
	}
	if c.currentView == ViewDashboard && vt != ViewDashboard {
		c.suspendCenter()
	}
	c.prevView = c.currentView
	c.currentView = vt
	if view, ok := c.views[vt]; ok {
		return tea.Batch(view.Init(), c.reconcileCenter())
	}
	return nil
}

func (c *Coordinator) resizeAll() {
	for viewType, view := range c.views {
		height := c.height - views.CoordinatorRows
		if viewType == ViewDashboardList {
			height = c.height
		}
		view.SetSize(c.width, height)
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

func (c *Coordinator) broadcastData(msg messages.DataLoadedMsg) {
	for _, view := range c.views {
		if dataReceiver, ok := view.(messages.DataReceiver); ok {
			dataReceiver.SetData(msg)
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
	case messages.ViewMCPManager:
		return ViewMCPManager
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
	case messages.ViewDashboardList:
		return ViewDashboardList
	default:
		return ViewDashboard
	}
}

func editorSettings(personas, mcps bool) models.EditorSettings {
	return models.EditorSettings{ShowPersonas: personas, ShowMCPs: mcps}
}
