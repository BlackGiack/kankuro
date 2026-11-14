package ui

import (
	"context"
	"errors"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/ui/models"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

const centerRetryInterval = 30 * time.Second

type centerState struct {
	port                                                  centerPort
	clock                                                 centerClock
	saved                                                 *centerCache
	sessionID, dashboardID                                string
	generation                                            uint64
	ctx                                                   context.Context
	cancel                                                context.CancelFunc
	subscription                                          contracts.ReportSubscription
	setup, reading, pending, working, timer, workspaceDue bool
	workspaceAttempted, workspaceStale                    bool
	request, workspaceRequest, timerRequest               uint64
	workspaceCancel, timerCancel                          context.CancelFunc
	reportWarning, subscriptionWarning, workspaceWarning  string
	data                                                  models.SessionCenterData
	limit                                                 int
}

// centerCache holds detached content only across a dashboard navigation.
type centerCache struct {
	dashboardID, sessionID                           string
	data                                             models.SessionCenterData
	workspaceAttempted, workspaceStale, workspaceDue bool
	reportWarning, workspaceWarning                  string
}
type centerSetupMsg struct {
	generation   uint64
	subscription contracts.ReportSubscription
	err          error
}
type centerChangeMsg struct {
	generation   uint64
	subscription contracts.ReportSubscription
	closed       bool
	err          error
}
type centerReportsMsg struct {
	generation, request uint64
	result              contracts.SessionReportsResult
	err                 error
}
type centerWorkspaceMsg struct {
	generation, request uint64
	result              contracts.SessionWorkspaceResult
	err                 error
}
type centerTimerMsg struct{ generation, request uint64 }

func (c *Coordinator) clearCenter() {
	st := &c.center
	st.generation++
	if st.cancel != nil {
		st.cancel()
	}
	if st.subscription != nil {
		st.subscription.Release()
	}
	if st.workspaceCancel != nil {
		st.workspaceCancel()
	}
	if st.timerCancel != nil {
		st.timerCancel()
	}
	port, clock, saved, generation := st.port, st.clock, st.saved, st.generation
	*st = centerState{port: port, clock: clock, saved: saved, generation: generation}
}
func (c *Coordinator) suspendCenter() {
	st := &c.center
	st.saved = nil
	if st.sessionID != "" && (len(st.data.Reports) > 0 || st.data.Workspace != nil) {
		st.saved = &centerCache{
			dashboardID:        st.dashboardID,
			sessionID:          st.sessionID,
			data:               models.CloneSessionCenterData(st.data),
			workspaceAttempted: st.workspaceAttempted,
			workspaceStale:     st.workspaceStale,
			workspaceDue:       st.workspaceDue,
			reportWarning:      st.reportWarning,
			workspaceWarning:   st.workspaceWarning,
		}
	}
	c.clearCenter()
}
func (c *Coordinator) reconcileCenter() tea.Cmd {
	if c.currentView != ViewDashboard || c.activeDashboard == "" {
		return nil
	}
	d, ok := c.views[ViewDashboard].(*views.Dashboard)
	if !ok {
		return nil
	}
	id := d.CenterSessionID()
	if c.centerSelectionBlocked(id) {
		return nil
	}
	if id == "" {
		c.center.saved = nil
		if c.center.sessionID != "" {
			c.clearCenter()
		}
		return nil
	}
	if c.center.sessionID != id || c.center.dashboardID != c.activeDashboard {
		if c.center.saved != nil && (c.center.saved.sessionID != id || c.center.saved.dashboardID != c.activeDashboard) {
			c.center.saved = nil
		}
		c.clearCenter()
		st := &c.center
		st.sessionID, st.dashboardID = id, c.activeDashboard
		st.ctx, st.cancel = context.WithCancel(context.Background())
		st.data = models.SessionCenterData{SessionID: id, Loading: true}
		if st.saved != nil {
			cached := st.saved
			st.data = models.CloneSessionCenterData(cached.data)
			st.workspaceAttempted, st.workspaceStale, st.workspaceDue = cached.workspaceAttempted, cached.workspaceStale, cached.workspaceDue
			st.reportWarning, st.workspaceWarning = cached.reportWarning, cached.workspaceWarning
			st.saved = nil
		}
		st.limit = d.CenterReportLimit()
		c.publishCenter()
		return c.centerSubscribe()
	}
	limit := d.CenterReportLimit()
	if limit > c.center.limit {
		c.center.limit = limit
		if c.center.reportWarning != "" {
			return nil
		}
		return c.centerRead()
	}
	return nil
}
func (c *Coordinator) centerSelectionBlocked(id string) bool {
	if c.centerBlockedSession == "" {
		return false
	}
	if id == c.centerBlockedSession && c.activeDashboard == c.centerBlockedDashboard {
		return true
	}
	c.centerBlockedSession, c.centerBlockedDashboard = "", ""
	return false
}
func (c *Coordinator) publishCenter() {
	st := &c.center
	st.data.Warnings = nil
	if st.reportWarning != "" {
		st.data.Warnings = append(st.data.Warnings, st.reportWarning)
	}
	if st.subscriptionWarning != "" {
		st.data.Warnings = append(st.data.Warnings, st.subscriptionWarning)
	}
	if st.workspaceWarning != "" {
		st.data.Warnings = append(st.data.Warnings, st.workspaceWarning)
	}
	if st.data.Workspace != nil {
		st.data.Warnings = append(st.data.Warnings, st.data.Workspace.Warnings...)
	}
	if d, ok := c.views[ViewDashboard].(*views.Dashboard); ok {
		d.SetCenterData(st.data)
	}
}
func (c *Coordinator) centerSubscribe() tea.Cmd {
	st := &c.center
	if st.setup || st.ctx == nil {
		return nil
	}
	st.setup = true
	ctx, port, generation := st.ctx, st.port, st.generation
	scope := contracts.ScopedSessionRequest{DashboardID: st.dashboardID, SessionID: st.sessionID}
	return func() tea.Msg {
		sub, err := port.SubscribeSessionReports(ctx, scope)
		if ctx.Err() != nil && sub != nil {
			sub.Release()
			sub = nil
		}
		return centerSetupMsg{generation: generation, subscription: sub, err: err}
	}
}
func centerWait(generation uint64, sub contracts.ReportSubscription) tea.Cmd {
	return func() tea.Msg {
		_, ok := <-sub.Changes()
		return centerChangeMsg{generation: generation, subscription: sub, closed: !ok, err: sub.Err()}
	}
}
func (c *Coordinator) centerRead() tea.Cmd {
	st := &c.center
	if st.ctx == nil || st.limit <= 0 {
		return nil
	}
	if st.reading {
		st.pending = true
		return nil
	}
	st.reading = true
	st.pending = false
	st.request++
	ctx, port, generation, request := st.ctx, st.port, st.generation, st.request
	req := contracts.SessionReportsRequest{Limit: st.limit}
	req.DashboardID, req.SessionID = st.dashboardID, st.sessionID
	return func() tea.Msg {
		result, err := port.ReadSessionReports(ctx, req)
		return centerReportsMsg{generation: generation, request: request, result: result, err: err}
	}
}
func (c *Coordinator) centerWorkspace() tea.Cmd {
	st := &c.center
	if st.ctx == nil || st.working {
		return nil
	}
	st.working = true
	st.workspaceAttempted = true
	st.workspaceDue = false
	st.workspaceRequest++
	ctx, cancel := context.WithCancel(st.ctx)
	st.workspaceCancel = cancel
	port, generation, request := st.port, st.generation, st.workspaceRequest
	scope := contracts.ScopedSessionRequest{DashboardID: st.dashboardID, SessionID: st.sessionID}
	return func() tea.Msg {
		result, err := port.ReadSessionWorkspace(ctx, scope)
		return centerWorkspaceMsg{generation: generation, request: request, result: result, err: err}
	}
}
func (c *Coordinator) cancelCenterWorkspace() {
	st := &c.center
	if st.workspaceCancel != nil {
		st.workspaceCancel()
		st.workspaceCancel = nil
	}
	st.workspaceRequest++
	st.working = false
}
func (c *Coordinator) centerSchedule() tea.Cmd {
	st := &c.center
	if st.timer || st.ctx == nil || st.reading || st.working || st.setup {
		return nil
	}
	if len(st.data.Reports) > 0 && st.reportWarning == "" && st.subscriptionWarning == "" {
		return nil
	}
	st.timer = true
	st.timerRequest++
	ctx, cancel := context.WithCancel(st.ctx)
	st.timerCancel = cancel
	generation, request := st.generation, st.timerRequest
	clock := st.clock
	if clock == nil {
		clock = systemCenterClock{}
	}
	return func() tea.Msg {
		if !clock.Wait(ctx, centerRetryInterval) {
			return nil
		}
		return centerTimerMsg{generation: generation, request: request}
	}
}

//nolint:gocognit,gocyclo,cyclop,funlen // The coordinator serializes correlated state transitions here.
func (c *Coordinator) handleCenter(msg tea.Msg) (bool, tea.Cmd) {
	st := &c.center
	switch m := msg.(type) {
	case centerSetupMsg:
		if m.generation != st.generation || st.ctx == nil {
			if m.subscription != nil {
				m.subscription.Release()
			}
			return true, nil
		}
		st.setup = false
		if m.err == nil && m.subscription == nil {
			m.err = errors.New("subscription unavailable")
		}
		if m.err != nil {
			if errors.Is(m.err, contracts.ErrSessionScope) {
				c.invalidateCenterScope()
				return true, nil
			}
			st.subscriptionWarning = "Report updates unavailable: " + m.err.Error()
		} else {
			st.subscription = m.subscription
			st.subscriptionWarning = ""
		}
		c.publishCenter()
		var wait tea.Cmd
		if st.subscription != nil {
			wait = centerWait(st.generation, st.subscription)
		}
		return true, tea.Batch(wait, c.centerRead())
	case centerChangeMsg:
		if m.generation != st.generation || st.subscription != m.subscription {
			return true, nil
		}
		if m.closed {
			st.subscription.Release()
			st.subscription = nil
			if errors.Is(m.err, contracts.ErrSessionScope) {
				c.invalidateCenterScope()
				return true, nil
			}
			st.subscriptionWarning = "Report updates disconnected"
			c.publishCenter()
			return true, c.centerSchedule()
		}
		wait := centerWait(st.generation, st.subscription)
		// A failed report read enters the timed recovery cycle. Global store
		// changes are coalesced until that cycle retries the read.
		if st.reportWarning != "" {
			st.pending = true
			return true, tea.Batch(wait, c.centerSchedule())
		}
		return true, tea.Batch(wait, c.centerRead())
	case centerReportsMsg:
		if m.generation != st.generation || m.request != st.request || st.ctx == nil {
			return true, nil
		}
		st.reading = false
		if errors.Is(m.err, contracts.ErrSessionScope) {
			c.invalidateCenterScope()
			return true, nil
		}
		if m.err != nil {
			st.reportWarning = "Reports unavailable: " + m.err.Error()
			st.data.Stale = len(st.data.Reports) > 0 || (st.data.Workspace != nil && st.workspaceStale)
		} else {
			st.reportWarning = ""
			st.data.Loading = false
			st.data.Reports = append([]contracts.SessionReport(nil), m.result.Reports...)
			st.data.Omitted = m.result.Omitted
			st.data.Stale = st.data.Workspace != nil && st.workspaceStale
			if len(st.data.Reports) > 0 {
				st.data.Stale = false
				st.workspaceAttempted, st.workspaceStale = false, false
				if st.reportWarning == "" && st.subscriptionWarning == "" && st.timerCancel != nil {
					st.timerCancel()
					st.timerCancel = nil
					st.timer = false
					st.timerRequest++
				}
				c.cancelCenterWorkspace()
				st.data.Workspace = nil
				st.workspaceWarning = ""
			}
		}
		c.publishCenter()
		if st.pending && m.err == nil {
			return true, c.centerRead()
		}
		if len(st.data.Reports) == 0 && !st.working && ((!st.workspaceAttempted && st.data.Workspace == nil) || st.workspaceDue) {
			return true, c.centerWorkspace()
		}
		return true, c.centerSchedule()
	case centerWorkspaceMsg:
		if m.generation != st.generation || m.request != st.workspaceRequest || st.ctx == nil {
			return true, nil
		}
		st.working = false
		if st.workspaceCancel != nil {
			st.workspaceCancel()
			st.workspaceCancel = nil
		}
		if errors.Is(m.err, contracts.ErrSessionScope) {
			c.invalidateCenterScope()
			return true, nil
		}
		st.data.Loading = false
		if m.err != nil {
			st.workspaceWarning = joinCenterWarning("Workspace unavailable: "+m.err.Error(), strings.Join(m.result.Warnings, "; "))
			st.workspaceStale = true
			st.data.Stale = st.data.Workspace != nil
		} else {
			st.workspaceStale = false
			st.data.Workspace = &m.result
			st.workspaceWarning = ""
			st.data.Stale = false
		}
		c.publishCenter()
		return true, c.centerSchedule()
	case centerTimerMsg:
		if m.generation != st.generation || m.request != st.timerRequest || st.ctx == nil {
			return true, nil
		}
		st.timer = false
		if len(st.data.Reports) == 0 {
			st.workspaceDue = true
		}
		if st.timerCancel != nil {
			st.timerCancel()
			st.timerCancel = nil
		}
		if st.subscription == nil {
			return true, c.centerSubscribe()
		}
		return true, c.centerRead()
	}
	return false, nil
}
func joinCenterWarning(a, b string) string {
	if a == "" {
		return b
	}
	if b == "" {
		return a
	}
	return a + "; " + b
}

func (c *Coordinator) invalidateCenterScope() {
	st := &c.center
	st.saved = nil
	c.centerBlockedSession, c.centerBlockedDashboard = st.sessionID, st.dashboardID
	st.data = models.SessionCenterData{SessionID: st.sessionID, Warnings: []string{"Session is no longer available"}}
	if d, ok := c.views[ViewDashboard].(*views.Dashboard); ok {
		d.SetCenterData(st.data)
	}
	c.clearCenter()
}

type centerClock interface {
	Wait(context.Context, time.Duration) bool
}
type systemCenterClock struct{}

func (systemCenterClock) Wait(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
