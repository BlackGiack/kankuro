package ui

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/ui/component"
	"github.com/BlackGiack/kankuro/internal/ui/models"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

type centerTestPort struct{}

func (centerTestPort) ReadSessionReports(context.Context, contracts.SessionReportsRequest) (contracts.SessionReportsResult, error) {
	return contracts.SessionReportsResult{}, nil
}
func (centerTestPort) SubscribeSessionReports(context.Context, contracts.ScopedSessionRequest) (contracts.ReportSubscription, error) {
	return nil, errors.New("test")
}
func (centerTestPort) ReadSessionWorkspace(context.Context, contracts.ScopedSessionRequest) (contracts.SessionWorkspaceResult, error) {
	return contracts.SessionWorkspaceResult{}, nil
}

type centerTestSubscription struct {
	ch       chan struct{}
	released bool
}

func (s *centerTestSubscription) Changes() <-chan struct{} { return s.ch }
func (s *centerTestSubscription) Err() error               { return nil }
func (s *centerTestSubscription) Release()                 { s.released = true }
func centerTestCoordinator(t *testing.T) *Coordinator {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	d := views.NewDashboard(views.DashboardCallbacks{})
	d.SetSize(120, 35)
	viewMap := make(map[ViewType]component.Model)
	viewMap[ViewDashboard] = d
	c := &Coordinator{currentView: ViewDashboard, activeDashboard: "d", views: viewMap}
	c.center = centerState{
		port:        centerTestPort{},
		sessionID:   "s",
		dashboardID: "d",
		generation:  1,
		ctx:         ctx,
		cancel:      cancel,
		limit:       8,
		data:        models.SessionCenterData{SessionID: "s"},
	}
	return c
}
func TestCenterReportTakesOverAndRejectsLateWorkspace(t *testing.T) {
	c := centerTestCoordinator(t)
	s := &c.center
	if c.centerRead() == nil {
		t.Fatal("no report read")
	}
	_, workspace := c.handleCenter(centerReportsMsg{generation: 1, request: s.request, result: contracts.SessionReportsResult{}})
	if workspace == nil || !s.working {
		t.Fatal("empty reports did not begin workspace read")
	}
	old := s.workspaceRequest
	sub := &centerTestSubscription{ch: make(chan struct{})}
	s.subscription = sub
	_, next := c.handleCenter(centerChangeMsg{generation: 1, subscription: sub})
	if next == nil || !s.working || !s.reading {
		t.Fatal("report event should read while retaining workspace work")
	}
	c.handleCenter(
		centerReportsMsg{
			generation: 1,
			request:    s.request,
			result:     contracts.SessionReportsResult{Reports: []contracts.SessionReport{{Text: "new"}}},
		},
	)
	if s.working {
		t.Fatal("first report did not cancel obsolete workspace work")
	}
	c.handleCenter(centerWorkspaceMsg{generation: 1, request: old, result: contracts.SessionWorkspaceResult{WorkingDir: "obsolete"}})
	if len(s.data.Reports) != 1 || s.data.Workspace != nil {
		t.Fatalf("late workspace overwrote reports: %+v", s.data)
	}
}
func TestCenterCoalescesAndReleasesStaleSetup(t *testing.T) {
	c := centerTestCoordinator(t)
	s := &c.center
	sub := &centerTestSubscription{ch: make(chan struct{})}
	s.subscription = sub
	c.centerRead()
	c.handleCenter(centerChangeMsg{generation: 1, subscription: sub})
	c.handleCenter(centerChangeMsg{generation: 1, subscription: sub})
	if !s.pending {
		t.Fatal("missing pending refresh")
	}
	before := s.request
	_, cmd := c.handleCenter(centerReportsMsg{generation: 1, request: before})
	if cmd == nil || s.request != before+1 || s.pending {
		t.Fatal("events did not coalesce into a follow-up read")
	}
	stale := &centerTestSubscription{ch: make(chan struct{})}
	c.clearCenter()
	c.handleCenter(centerSetupMsg{generation: 1, subscription: stale})
	if !stale.released || !sub.released {
		t.Fatal("selection change leaked subscription")
	}
}
func TestCenterScopeFailureClearsCachedReports(t *testing.T) {
	c := centerTestCoordinator(t)
	s := &c.center
	s.data.Reports = []contracts.SessionReport{{Text: "secret"}}
	s.reading = true
	s.request = 2
	c.handleCenter(centerReportsMsg{generation: 1, request: 2, err: contracts.ErrSessionScope})
	if c.center.sessionID != "" || len(c.center.data.Reports) != 0 {
		t.Fatal("scope rejection retained selection data")
	}
	dashboard := c.views[ViewDashboard].(*views.Dashboard)
	dashboard.SetInventory(
		contracts.ReadResult[contracts.DashboardInventory]{
			Attempted: true,
			Value:     contracts.DashboardInventory{Sessions: []contracts.SessionSummary{{ID: "s"}}},
		},
		contracts.DisplayChoicesResult{},
	)
	if cmd := c.reconcileCenter(); cmd != nil || c.center.sessionID != "" {
		t.Fatal("scope rejection caused immediate retry")
	}
}

type fakeCenterClock struct {
	delays chan time.Duration
	ticks  chan struct{}
}

func (f *fakeCenterClock) Wait(ctx context.Context, delay time.Duration) bool {
	f.delays <- delay
	select {
	case <-f.ticks:
		return true
	case <-ctx.Done():
		return false
	}
}
func TestCenterRetryClockIsNonOverlappingAndCanceled(t *testing.T) {
	c := centerTestCoordinator(t)
	clock := &fakeCenterClock{delays: make(chan time.Duration, 1), ticks: make(chan struct{})}
	c.center.clock = clock
	cmd := c.centerSchedule()
	if cmd == nil || c.centerSchedule() != nil {
		t.Fatal("expected exactly one timer")
	}
	done := make(chan tea.Msg, 1)
	go func() { done <- cmd() }()
	if delay := <-clock.delays; delay != 30*time.Second {
		t.Fatalf("delay %s", delay)
	}
	clock.ticks <- struct{}{}
	msg := <-done
	_, next := c.handleCenter(msg)
	if next == nil || !c.center.setup || c.centerSchedule() != nil {
		t.Fatal("retry must resubscribe before reading and cannot overlap")
	}
	c.clearCenter()
}

func TestCenterFailurePolicyRetainsOnlySameSessionData(t *testing.T) {
	c := centerTestCoordinator(t)
	st := &c.center
	st.data.Reports = []contracts.SessionReport{{Text: "saved"}}
	st.reading = true
	st.request = 1
	_, cmd := c.handleCenter(centerReportsMsg{generation: 1, request: 1, err: errors.New("history error")})
	if len(st.data.Reports) != 1 || !st.data.Stale || st.reportWarning == "" || cmd == nil {
		t.Fatalf("failed history did not retain stale reports: %+v", st.data)
	}
	st.reading = true
	st.request++
	_, cmd = c.handleCenter(centerReportsMsg{generation: 1, request: st.request, result: contracts.SessionReportsResult{}})
	if len(st.data.Reports) != 0 || st.data.Stale || cmd == nil || !st.working {
		t.Fatalf("confirmed empty did not clear reports: %+v", st.data)
	}
	st.data.Workspace = &contracts.SessionWorkspaceResult{WorkingDir: "saved"}
	c.handleCenter(
		centerWorkspaceMsg{
			generation: 1,
			request:    st.workspaceRequest,
			err:        errors.New("missing"),
			result:     contracts.SessionWorkspaceResult{Warnings: []string{"Git failed"}},
		},
	)
	if !st.timer || !st.data.Stale || st.data.Workspace == nil || st.workspaceWarning == "" {
		t.Fatalf("failed workspace did not retain stale data and warnings: %+v", st.data)
	}
	c.clearCenter()
	if len(c.center.data.Reports) != 0 || c.center.data.Workspace != nil {
		t.Fatal("selection clear retained old content")
	}
}

func TestCenterClosedSubscriptionUsesRetry(t *testing.T) {
	c := centerTestCoordinator(t)
	st := &c.center
	sub := &centerTestSubscription{ch: make(chan struct{})}
	st.subscription = sub
	st.data.Reports = []contracts.SessionReport{{Text: "saved"}}
	_, cmd := c.handleCenter(centerChangeMsg{generation: 1, subscription: sub, closed: true, err: errors.New("closed")})
	if cmd == nil || !st.timer || st.subscription != nil || st.subscriptionWarning == "" || len(st.data.Reports) != 1 {
		t.Fatal("closed stream did not preserve content and schedule recovery")
	}
}

func fireCenterClock(t *testing.T, clock *fakeCenterClock, cmd tea.Cmd) tea.Msg {
	t.Helper()
	if cmd == nil {
		t.Fatal("missing timer command")
	}
	done := make(chan tea.Msg, 1)
	go func() { done <- cmd() }()
	if delay := <-clock.delays; delay != centerRetryInterval {
		t.Fatalf("unexpected retry interval %v", delay)
	}
	clock.ticks <- struct{}{}
	return <-done
}

func TestCenterFailedReadCoalescesChangesUntilRetry(t *testing.T) {
	c := centerTestCoordinator(t)
	st := &c.center
	clock := &fakeCenterClock{delays: make(chan time.Duration, 1), ticks: make(chan struct{})}
	st.clock = clock
	sub := &centerTestSubscription{ch: make(chan struct{})}
	st.subscription = sub
	st.data.Reports = []contracts.SessionReport{{Text: "retained"}}
	c.centerRead()
	_, timer := c.handleCenter(centerReportsMsg{generation: 1, request: st.request, err: errors.New("history failed")})
	request := st.request
	for range 4 {
		c.handleCenter(centerChangeMsg{generation: 1, subscription: sub})
	}
	if st.reading || st.request != request || !st.pending || !st.timer {
		t.Fatalf("change bypassed recovery interval: %+v", st)
	}
	_, read := c.handleCenter(fireCenterClock(t, clock, timer))
	if read == nil || !st.reading || st.request != request+1 || st.pending {
		t.Fatal("timer did not start one coalesced retry")
	}
	for range 3 {
		c.handleCenter(centerChangeMsg{generation: 1, subscription: sub})
	}
	_, timer = c.handleCenter(centerReportsMsg{generation: 1, request: st.request, err: errors.New("history still failed")})
	if st.reading || !st.pending || !st.timer || timer == nil {
		t.Fatal("failed retry started another read instead of scheduling recovery")
	}
}

func TestCenterChangesDoNotRestartWorkspace(t *testing.T) {
	c := centerTestCoordinator(t)
	st := &c.center
	clock := &fakeCenterClock{delays: make(chan time.Duration, 1), ticks: make(chan struct{})}
	st.clock = clock
	sub := &centerTestSubscription{ch: make(chan struct{})}
	st.subscription = sub
	st.data.Workspace = &contracts.SessionWorkspaceResult{WorkingDir: "existing"}
	timer := c.centerSchedule()
	if timer == nil {
		t.Fatal("missing workspace timer")
	}
	for range 3 {
		c.handleCenter(centerChangeMsg{generation: 1, subscription: sub})
		if !st.reading {
			t.Fatal("healthy change did not read reports")
		}
		c.handleCenter(centerReportsMsg{generation: 1, request: st.request, result: contracts.SessionReportsResult{}})
		if st.working || !st.timer {
			t.Fatal("empty event read restarted workspace before timer")
		}
	}
	_, read := c.handleCenter(fireCenterClock(t, clock, timer))
	if read == nil {
		t.Fatal("timer did not start refresh")
	}
	_, workspace := c.handleCenter(centerReportsMsg{generation: 1, request: st.request})
	if workspace == nil || !st.working {
		t.Fatal("scheduled refresh did not start workspace inspection")
	}
	workspaceRequest := st.workspaceRequest
	for range 3 {
		c.handleCenter(centerChangeMsg{generation: 1, subscription: sub})
		if !st.working || st.workspaceRequest != workspaceRequest {
			t.Fatal("change canceled slow workspace inspection")
		}
		c.handleCenter(centerReportsMsg{generation: 1, request: st.request})
		if !st.working || st.workspaceRequest != workspaceRequest {
			t.Fatal("empty report read restarted slow workspace inspection")
		}
	}
}

func TestCenterFailedWorkspaceWithoutCacheWaitsForTimer(t *testing.T) {
	c := centerTestCoordinator(t)
	st := &c.center
	clock := &fakeCenterClock{delays: make(chan time.Duration, 1), ticks: make(chan struct{})}
	st.clock = clock
	sub := &centerTestSubscription{ch: make(chan struct{})}
	st.subscription = sub
	c.centerRead()
	_, workspace := c.handleCenter(centerReportsMsg{generation: 1, request: st.request})
	if workspace == nil || !st.working {
		t.Fatal("initial workspace load did not start")
	}
	_, timer := c.handleCenter(centerWorkspaceMsg{generation: 1, request: st.workspaceRequest, err: errors.New("inaccessible")})
	if timer == nil || !st.timer || st.data.Workspace != nil {
		t.Fatal("failure did not start recovery timer")
	}
	previous := st.workspaceRequest
	for range 4 {
		c.handleCenter(centerChangeMsg{generation: 1, subscription: sub})
		if !st.reading {
			t.Fatal("healthy report discovery stopped")
		}
		c.handleCenter(centerReportsMsg{generation: 1, request: st.request})
		if st.working || st.workspaceRequest != previous || !st.timer {
			t.Fatal("workspace was retried before timer")
		}
	}
	_, read := c.handleCenter(fireCenterClock(t, clock, timer))
	if read == nil || !st.reading {
		t.Fatal("timer did not retry reports")
	}
	_, workspace = c.handleCenter(centerReportsMsg{generation: 1, request: st.request})
	if workspace == nil || !st.working || st.workspaceRequest != previous+1 {
		t.Fatal("timer did not permit workspace retry")
	}
}

func TestCenterWorkspaceStaleUntilSuccessfulWorkspaceOrReport(t *testing.T) {
	c := centerTestCoordinator(t)
	st := &c.center
	clock := &fakeCenterClock{delays: make(chan time.Duration, 1), ticks: make(chan struct{})}
	st.clock = clock
	sub := &centerTestSubscription{ch: make(chan struct{})}
	st.subscription = sub
	st.data.Workspace = &contracts.SessionWorkspaceResult{WorkingDir: "cached"}
	c.centerWorkspace()
	_, timer := c.handleCenter(centerWorkspaceMsg{generation: 1, request: st.workspaceRequest, err: errors.New("unavailable")})
	if !st.data.Stale || timer == nil {
		t.Fatal("failed refresh did not mark cached workspace stale")
	}
	c.handleCenter(centerChangeMsg{generation: 1, subscription: sub})
	c.handleCenter(centerReportsMsg{generation: 1, request: st.request})
	if !st.data.Stale || st.working {
		t.Fatal("empty report read changed stale workspace")
	}
	c.handleCenter(centerChangeMsg{generation: 1, subscription: sub})
	c.handleCenter(centerReportsMsg{generation: 1, request: st.request, err: errors.New("report failure")})
	if !st.data.Stale || st.data.Workspace == nil {
		t.Fatal("failed report read changed stale workspace")
	}
	_, read := c.handleCenter(fireCenterClock(t, clock, timer))
	if read == nil {
		t.Fatal("missing recovery read")
	}
	_, workspace := c.handleCenter(centerReportsMsg{generation: 1, request: st.request})
	if workspace == nil || !st.working || !st.data.Stale {
		t.Fatal("expected stale workspace during retry")
	}
	c.handleCenter(
		centerWorkspaceMsg{generation: 1, request: st.workspaceRequest, result: contracts.SessionWorkspaceResult{WorkingDir: "fresh"}},
	)
	if st.data.Stale || st.workspaceStale || st.data.Workspace.WorkingDir != "fresh" {
		t.Fatal("successful workspace did not clear stale status")
	}
	// A new successful authored report also replaces stale workspace content.
	c.centerWorkspace()
	c.handleCenter(centerWorkspaceMsg{generation: 1, request: st.workspaceRequest, err: errors.New("again")})
	c.centerRead()
	c.handleCenter(
		centerReportsMsg{
			generation: 1,
			request:    st.request,
			result:     contracts.SessionReportsResult{Reports: []contracts.SessionReport{{Text: "report"}}},
		},
	)
	if st.data.Stale || st.data.Workspace != nil || len(st.data.Reports) != 1 {
		t.Fatal("report takeover retained workspace stale state")
	}
}

func TestCenterNavigationRestoresSameSessionAndClearsOnEmpty(t *testing.T) {
	c := centerTestCoordinator(t)
	d := c.views[ViewDashboard].(*views.Dashboard)
	d.SetInventory(
		contracts.ReadResult[contracts.DashboardInventory]{
			Attempted: true,
			Value:     contracts.DashboardInventory{Sessions: []contracts.SessionSummary{{ID: "s"}, {ID: "other"}}},
		},
		contracts.DisplayChoicesResult{},
	)
	d.SessionList().SetSelectedByID("s")
	d.SyncCenterSelection()
	st := &c.center
	st.data.Reports = []contracts.SessionReport{{Text: "saved report", Kind: "completed"}}
	st.subscription = &centerTestSubscription{ch: make(chan struct{})}
	c.publishCenter()
	c.navigateTo(ViewSettings)
	if st.saved == nil || len(st.saved.data.Reports) != 1 || st.ctx != nil {
		t.Fatal("navigation did not detach cached content and cancel resources")
	}
	c.navigateTo(ViewDashboard)
	if len(st.data.Reports) != 1 || st.data.Reports[0].Text != "saved report" {
		t.Fatal("same session cache was not restored")
	}
	sub := &centerTestSubscription{ch: make(chan struct{})}
	c.handleCenter(centerSetupMsg{generation: st.generation, subscription: sub})
	c.handleCenter(centerReportsMsg{generation: st.generation, request: st.request, err: errors.New("history unavailable")})
	if !st.data.Stale || len(st.data.Reports) != 1 {
		t.Fatal("failed reread did not retain stale reports")
	}
	if rendered := d.Center().View(); !strings.Contains(rendered, "saved report") || !strings.Contains(rendered, "[stale]") {
		t.Fatalf("panel did not show stale cached report: %q", rendered)
	}
	// A confirmed empty result clears old reports before another navigation.
	c.centerRead()
	c.handleCenter(centerReportsMsg{generation: st.generation, request: st.request})
	if len(st.data.Reports) != 0 {
		t.Fatal("successful empty read did not clear reports")
	}
	c.navigateTo(ViewSettings)
	c.navigateTo(ViewDashboard)
	if len(st.data.Reports) != 0 {
		t.Fatal("cleared reports returned after navigation")
	}
}

func TestCenterNavigationCacheNeverMovesToAnotherSelection(t *testing.T) {
	c := centerTestCoordinator(t)
	d := c.views[ViewDashboard].(*views.Dashboard)
	d.SetInventory(
		contracts.ReadResult[contracts.DashboardInventory]{
			Attempted: true,
			Value:     contracts.DashboardInventory{Sessions: []contracts.SessionSummary{{ID: "s"}, {ID: "other"}}},
		},
		contracts.DisplayChoicesResult{},
	)
	d.SessionList().SetSelectedByID("s")
	d.SyncCenterSelection()
	c.center.data.Workspace = &contracts.SessionWorkspaceResult{WorkingDir: "only-s"}
	c.publishCenter()
	c.navigateTo(ViewSettings)
	d.TogglePane()
	d.SessionList().SetSelectedByID("other")
	d.SyncCenterSelection()
	c.navigateTo(ViewDashboard)
	if c.center.sessionID != "other" || c.center.data.Workspace != nil || len(c.center.data.Reports) != 0 {
		t.Fatalf("other selection inherited cached content: %+v", c.center.data)
	}
}

func TestCenterNavigationRestoresWorkspaceAndClearsOnScopeLoss(t *testing.T) {
	c := centerTestCoordinator(t)
	d := c.views[ViewDashboard].(*views.Dashboard)
	d.SetInventory(
		contracts.ReadResult[contracts.DashboardInventory]{
			Attempted: true,
			Value:     contracts.DashboardInventory{Sessions: []contracts.SessionSummary{{ID: "s"}}},
		},
		contracts.DisplayChoicesResult{},
	)
	st := &c.center
	st.data.Workspace = &contracts.SessionWorkspaceResult{WorkingDir: "saved workspace"}
	st.workspaceAttempted = true
	c.publishCenter()
	c.navigateTo(ViewSettings)
	c.navigateTo(ViewDashboard)
	if st.data.Workspace == nil || st.data.Workspace.WorkingDir != "saved workspace" {
		t.Fatal("workspace cache was lost on navigation")
	}
	sub := &centerTestSubscription{ch: make(chan struct{})}
	c.handleCenter(centerSetupMsg{generation: st.generation, subscription: sub})
	c.handleCenter(centerReportsMsg{generation: st.generation, request: st.request, err: errors.New("history unavailable")})
	if st.data.Workspace == nil || st.reportWarning == "" {
		t.Fatal("report failure lost the cached workspace fallback")
	}
	c.centerRead()
	c.handleCenter(centerReportsMsg{generation: st.generation, request: st.request, err: contracts.ErrSessionScope})
	if st.data.Workspace != nil || st.saved != nil || st.sessionID != "" {
		t.Fatal("scope rejection retained cached content")
	}
}
