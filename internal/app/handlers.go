package app

import (
	tea "charm.land/bubbletea/v2"

	"kankuro/internal/ui/messages"
	"kankuro/internal/ui/views"
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

// refreshData is the unseeded full refresh behind the exported RefreshData.
func (c *Coordinator) refreshData() tea.Cmd {
	return c.refresh.full(c.currentPreviewSessionID())
}

// refreshDataSeeded is the full refresh every mutation path uses; it also
// seeds the fingerprint baseline (see refresher.fullSeeded).
func (c *Coordinator) refreshDataSeeded() tea.Cmd {
	return c.refresh.fullSeeded(c.currentPreviewSessionID())
}

// refreshSessions is the idle tick's sessions-only refresh.
func (c *Coordinator) refreshSessions() tea.Cmd {
	return c.refresh.sessionsOnly(c.currentPreviewSessionID())
}

// maybeRefreshConfig is the idle tick's config half; it hands the current
// baseline to the refresher by value.
func (c *Coordinator) maybeRefreshConfig() tea.Cmd {
	return c.refresh.configIfChanged(c.configFingerprint)
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
