package app

import (
	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/ui/messages"
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
	return c.refresh.scopedFull(c.actions.service, c.activeDashboard, c.viewGeneration, c.currentPreviewSessionID())
}

// refreshSessions is the idle tick's sessions-only refresh.
func (c *Coordinator) refreshSessions() tea.Cmd {
	return c.refresh.scopedSessions(c.actions.service, c.activeDashboard, c.viewGeneration, c.currentPreviewSessionID())
}

// maybeRefreshConfig is the idle tick's config half; it hands the current
// baseline to the refresher by value.
func (c *Coordinator) maybeRefreshConfig() tea.Cmd {
	return c.refresh.scopedConfigIfChanged(c.actions.service, c.activeDashboard, c.viewGeneration, c.configFingerprint)
}

// handleConfigFingerprint adopts non-nil stamps and data. An unchanged result
// must not overwrite a newer baseline from an overlapping mutation refresh.
func (c *Coordinator) handleConfigFingerprint(m configFingerprintMsg) tea.Cmd {
	if m.data != nil &&
		(m.data.DashboardID != c.activeDashboard || m.data.Generation != c.viewGeneration) {
		return nil
	}
	if m.stamps != nil {
		c.configFingerprint = m.stamps
	}
	if m.data == nil {
		return nil
	}
	return c.handleDataLoaded(*m.data)
}

func (c *Coordinator) handleDataLoaded(m messages.DataLoadedMsg) tea.Cmd {
	if m.DashboardID != c.activeDashboard || m.Generation != c.viewGeneration {
		return nil
	}
	if m.Err != nil {
		c.err = m.Err
		// A failed registry read is not a valid fingerprint baseline: retry
		// even if the underlying file's metadata does not change.
		c.configFingerprint = nil
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
	if c.currentView == ViewConversationChoice {
		c.refreshConversationChoiceContext()
	}
	if m.PreviewSessionID != "" {
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
	if c.activeDashboard == "" {
		return nil
	}
	if m.Err != nil {
		c.err = m.Err
	}
	partial := messages.DataLoadedMsg{Cards: m.Cards, Sessions: m.Sessions}
	for _, viewType := range []ViewType{ViewDashboard, ViewAgentSwitch} {
		if receiver, ok := c.views[viewType].(messages.DataReceiver); ok {
			receiver.SetData(partial)
		}
	}

	if c.currentView == ViewConversationChoice {
		c.refreshConversationChoiceContext()
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
