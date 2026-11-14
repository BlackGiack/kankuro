package app

import (
	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/models"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

// This file is the coordinator-goroutine half of data refresh and the
// preview: the methods that snapshot state for refresher / previewSub and
// apply what comes back to the views. They touch c.views, c.err and
// c.preview, so they must only run from Update/Init, never inside a tea.Cmd
// closure. The pure halves live in refresh.go and preview.go, which have no
// Coordinator pointer.

// The four wrappers below are the seam where invariant 1 matters: they run on
// the coordinator goroutine, snapshot the state a command needs
// (currentPreviewSessionID, configFingerprint) *now*, and hand plain values
// to the refresher, whose closures have no way to reach the coordinator.

// refreshDataSeeded is the full refresh every mutation path uses; it also
// seeds the fingerprint baseline (see refresher.fullSeeded).
func (c *Coordinator) refreshDataSeeded() tea.Cmd {
	refresh := c.refresh
	refresh.sessionEpoch = c.sessionEpoch
	requestID := c.nextRefreshRequest()
	inventoryRequestID := c.nextInventoryRequest()
	return refresh.scopedFull(
		c.reads, c.activeDashboard, c.viewGeneration, requestID, inventoryRequestID, c.currentPreviewSessionID(),
	)
}

// refreshSessions is the idle tick's sessions-only refresh.
func (c *Coordinator) refreshSessions() tea.Cmd {
	refresh := c.refresh
	refresh.sessionEpoch = c.sessionEpoch
	return refresh.scopedSessions(
		c.reads, c.activeDashboard, c.viewGeneration, c.nextInventoryRequest(), c.currentPreviewSessionID(),
	)
}

// maybeRefreshConfig is the idle tick's config half; it hands the current
// baseline to the refresher by value.
func (c *Coordinator) maybeRefreshConfig() tea.Cmd {
	refresh := c.refresh
	refresh.sessionEpoch = c.sessionEpoch
	requestID := c.nextRefreshRequest()
	inventoryRequestID := c.nextInventoryRequest()
	return refresh.scopedConfigIfChanged(
		c.reads, c.activeDashboard, c.viewGeneration, c.configFingerprint, requestID, inventoryRequestID,
		c.currentPreviewSessionID(),
	)
}

func (c *Coordinator) currentPreviewSessionID() string {
	if dashboard, ok := c.views[ViewDashboard].(*views.Dashboard); ok {
		return dashboard.PreviewSessionID()
	}
	return c.preview.current()
}

func (c *Coordinator) nextRefreshRequest() uint64 {
	c.refreshRequestID++
	return c.refreshRequestID
}

func (c *Coordinator) nextInventoryRequest() uint64 {
	c.inventoryRequestID++
	return c.inventoryRequestID
}

func (c *Coordinator) hasStaleSessionSnapshot(msg messages.DataLoadedMsg) bool {
	return msg.Inventory.Attempted && msg.SessionEpoch < c.sessionEpoch
}

func (c *Coordinator) acceptInventorySnapshot(msg *messages.DataLoadedMsg) bool {
	if c.hasStaleSessionSnapshot(*msg) {
		return false
	}
	if !msg.Inventory.Attempted {
		return true
	}
	if msg.InventoryRequestID < c.lastAppliedInventoryRequest {
		msg.Inventory = contracts.ReadResult[contracts.DashboardInventory]{}
		msg.PreviewSessionID = ""
		msg.PreviewScrollback = nil
		msg.Err = msg.RegistryErr
		return true
	}
	c.lastAppliedInventoryRequest = msg.InventoryRequestID
	return true
}

// handleConfigFingerprint adopts only current, scope-valid results. Unchanged
// polls carry no data or token and cannot overwrite a forced refresh.
func (c *Coordinator) handleConfigFingerprint(m configFingerprintMsg) tea.Cmd {
	if m.requestID < c.lastAppliedRefreshRequest ||
		m.dashboardID != c.activeDashboard || m.generation != c.viewGeneration ||
		m.sessionEpoch < c.sessionEpoch {
		return nil
	}
	c.lastAppliedRefreshRequest = m.requestID
	if m.data == nil {
		return nil
	}
	if m.data.DashboardID != m.dashboardID || m.data.Generation != m.generation ||
		m.data.SessionEpoch != m.sessionEpoch || c.hasStaleSessionSnapshot(*m.data) {
		return nil
	}
	c.configFingerprint = m.token
	return c.handleDataLoaded(*m.data)
}

//nolint:gocognit // Applies one scoped refresh to its independent consumers.
func (c *Coordinator) handleDataLoaded(m messages.DataLoadedMsg) tea.Cmd {
	if m.DashboardID != c.activeDashboard || m.Generation != c.viewGeneration {
		return nil
	}
	if !c.acceptInventorySnapshot(&m) {
		return nil
	}
	if m.Err != nil {
		c.err = m.Err
		// A failed registry read is not a valid fingerprint baseline: retry
		// even if the underlying file's metadata does not change.
		c.configFingerprint = ""
	}
	c.applyDisplayChoices(m.Choices)
	if editor, ok := c.views[ViewCardEditor].(*views.CardEditor); ok {
		editor.SetChoices(m.EditorChoices)
	}
	if c.activeDashboard == "" {
		for viewType, view := range c.views {
			if viewType == ViewDashboard || viewType == ViewAgentSwitch || viewType == ViewConversationChoice ||
				viewType == ViewSessionLauncher {
				continue
			}
			if receiver, ok := view.(messages.DataReceiver); ok {
				receiver.SetData(m)
			}
		}
		return nil
	}
	c.broadcastData(m)
	if dashboard, ok := c.views[ViewDashboard].(*views.Dashboard); ok {
		dashboard.SetInventory(m.Inventory, m.Choices)
	}
	if switchView, ok := c.views[ViewAgentSwitch].(*views.AgentSwitchView); ok && m.Inventory.Attempted && m.Inventory.Err == nil {
		switchView.SetRows(models.CardRows(m.Inventory.Value.Cards, c.displayChoices))
	}
	if c.currentView == ViewConversationChoice {
		c.refreshConversationChoiceContext()
	}
	if dashboard, ok := c.views[ViewDashboard].(*views.Dashboard); ok && m.PreviewSessionID != "" &&
		dashboard.PreviewSessionID() == m.PreviewSessionID {
		c.broadcastPreview(messages.PreviewDataMsg{
			SessionID:  m.PreviewSessionID,
			Scrollback: m.PreviewScrollback,
		})
	}
	return c.maybeSubscribeAutoSelectedPreview()
}

// handleSessionsRefreshed applies scoped cards, sessions and preview data to their
// consumers without reloading unrelated configuration views.
func (c *Coordinator) handleSessionsRefreshed(m messages.SessionsRefreshedMsg) tea.Cmd {
	if m.DashboardID != c.activeDashboard || m.Generation != c.viewGeneration {
		return nil
	}
	if m.SessionEpoch < c.sessionEpoch {
		return nil
	}
	if !m.Inventory.Attempted || m.InventoryRequestID < c.lastAppliedInventoryRequest {
		return nil
	}
	c.lastAppliedInventoryRequest = m.InventoryRequestID
	if c.activeDashboard == "" {
		return nil
	}
	if m.Err != nil {
		c.err = m.Err
	}
	if dashboard, ok := c.views[ViewDashboard].(*views.Dashboard); ok {
		dashboard.SetInventory(m.Inventory, m.Choices)
	}
	if switchView, ok := c.views[ViewAgentSwitch].(*views.AgentSwitchView); ok && m.Inventory.Attempted && m.Inventory.Err == nil {
		switchView.SetRows(models.CardRows(m.Inventory.Value.Cards, c.displayChoices))
	}

	if c.currentView == ViewConversationChoice {
		c.refreshConversationChoiceContext()
	}
	if dashboard, ok := c.views[ViewDashboard].(*views.Dashboard); ok && m.PreviewSessionID != "" &&
		dashboard.PreviewSessionID() == m.PreviewSessionID {
		c.broadcastPreview(messages.PreviewDataMsg{
			SessionID:  m.PreviewSessionID,
			Scrollback: m.PreviewScrollback,
		})
	}
	return c.maybeSubscribeAutoSelectedPreview()
}

func (c *Coordinator) applyDisplayChoices(results contracts.DisplayChoicesResult) {
	if results.Skills.Attempted && results.Skills.Err == nil {
		c.displayChoices.Skills = append([]contracts.NamedChoice(nil), results.Skills.Value...)
	}
	if results.Personas.Attempted && results.Personas.Err == nil {
		c.displayChoices.Personas = append([]contracts.NamedChoice(nil), results.Personas.Value...)
	}
	if results.Puppets.Attempted && results.Puppets.Err == nil {
		c.displayChoices.Puppets = append([]contracts.PuppetChoice(nil), results.Puppets.Value...)
		for i := range c.displayChoices.Puppets {
			c.displayChoices.Puppets[i].Lines = append([]string(nil), results.Puppets.Value[i].Lines...)
		}
	}
}

// maybeSubscribeAutoSelectedPreview subscribes to a session the dashboard
// auto-selected on its own (e.g. the first session on a fresh launch),
// shared by handleDataLoaded and handleSessionsRefreshed.
func (c *Coordinator) maybeSubscribeAutoSelectedPreview() tea.Cmd {
	if c.activeDashboard == "" || c.currentView != ViewDashboard {
		return nil
	}
	if dashboard, ok := c.views[ViewDashboard].(*views.Dashboard); ok {
		id := dashboard.PreviewSessionID()
		if id != "" && id != c.preview.current() {
			return c.handlePreviewRequest(id)
		}
	}
	return nil
}

// handlePreviewRequest switches the preview to sessionID: previewSub does the
// subscription bookkeeping and hands back the scrollback, which is broadcast
// to the views that render a preview.
func (c *Coordinator) handlePreviewRequest(sessionID string) tea.Cmd {
	item, exists := c.session.GetSession(sessionID)
	if !exists || item.DashboardID != c.activeDashboard || c.currentView != ViewDashboard {
		return nil
	}

	scrollback, cmd := c.preview.subscribe(sessionID)
	c.broadcastPreview(messages.PreviewDataMsg{
		SessionID:  sessionID,
		Scrollback: scrollback,
	})
	return cmd
}

// handleSessionOutput forwards a chunk of live output to the preview, but
// only if it is for the session we are still subscribed to - navigateTo
// unsubscribes (clearing both), so a SessionOutputMsg already queued for a
// session we've since navigated away from cannot re-arm next on a nil
// channel.
func (c *Coordinator) handleSessionOutput(m messages.SessionOutputMsg) tea.Cmd {
	if !c.preview.accepts(m) {
		return nil
	}
	c.broadcastPreview(messages.PreviewDataMsg{
		SessionID:     m.SessionID,
		Scrollback:    m.Output,
		IsIncremental: true,
	})
	return c.preview.next()
}
