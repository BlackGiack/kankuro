package ui

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

//nolint:gocognit // Exercise each outcome through both coordinator views and failed refreshes.
func TestDeletionOutcomesReachCardAndDashboardViews(t *testing.T) {
	cases := []struct {
		outcome contracts.DeletionOutcome
		want    string
	}{
		{contracts.DeletionNoIntent, "No deletion intent was saved"},
		{contracts.DeletionPendingOutcome, "Deletion is pending"},
		{contracts.DeletionCompleted, "Deletion completed"},
		{contracts.DeletionUncertain, "Deletion outcome is uncertain"},
	}
	for _, tc := range cases {
		t.Run(string(tc.outcome), func(t *testing.T) {
			for _, kind := range []string{"card", "dashboard"} {
				t.Run(kind, func(t *testing.T) {
					c, _ := newLaunchFocusCoordinator(t)
					p := contracts.DeletionProgress{
						DashboardID: c.activeDashboard,
						Outcome:     tc.outcome,
						Diagnostic:  "cleanup error",
						Sessions:    []contracts.DeletionSession{{ID: "checkpoint", Stopped: true}},
					}
					if kind == "card" {
						p.CardID = "card"
						c.cardDeleteAction = 1
						c.cardDeleteTarget = "card"
						c.Update(
							messages.ScopedActionFinishedMsg{
								DashboardID: c.activeDashboard,
								Generation:  c.viewGeneration,
								Request:     1,
								Visit:       c.deletionVisit,
								Progress:    p,
								Err:         errors.New("cleanup error"),
							},
						)
					} else {
						c.navigateTo(ViewDashboardList)
						c.dashboardDeleteAction = 1
						c.dashboardDeleteTarget = p.DashboardID
						c.Update(
							messages.DashboardMutationFinishedMsg{
								Generation: c.viewGeneration,
								Request:    1,
								Visit:      c.deletionVisit,
								Progress:   p,
								Err:        errors.New("cleanup error"),
							},
						)
					}
					check := func() {
						t.Helper()
						rendered := c.View()
						for _, want := range []string{tc.want, "cleanup error", "checkpoint"} {
							if !strings.Contains(rendered, want) {
								t.Fatalf("missing %q in view: %s", want, rendered)
							}
						}
					}
					check()
					if kind == "card" {
						readErr := errors.New("inventory failed")
						c.Update(messages.SessionsRefreshedMsg{
							DashboardID: c.activeDashboard, Generation: c.viewGeneration, SessionEpoch: c.sessionEpoch,
							InventoryRequestID: c.lastAppliedInventoryRequest + 1, Err: readErr,
							Inventory: contracts.ReadResult[contracts.DashboardInventory]{Attempted: true, Err: readErr},
						})
					} else {
						c.Update(messages.DashboardListLoadedMsg{RequestID: c.dashboardListRequest, Err: errors.New("inventory failed")})
					}
					check()
					if kind == "dashboard" {
						list := c.views[ViewDashboardList].(*views.DashboardList)
						if tc.outcome != contracts.DeletionPendingOutcome && strings.Contains(list.View(), "Delete: retry") {
							t.Fatal("unsafe retry offered for outcome")
						}
					}
				})
			}
		})
	}
}

func TestDashboardExecutionProgressSurvivesStatusReadFailure(t *testing.T) {
	c, _ := newLaunchFocusCoordinator(t)
	c.navigateTo(ViewDashboardList)
	list := c.views[ViewDashboardList].(*views.DashboardList)
	list.SetDashboards(
		[]contracts.DashboardSummary{
			{ID: "d", Name: "Dashboard", State: contracts.DashboardDeleting, Inspect: contracts.Availability{Enabled: true}},
		},
	)
	c.dashboardDeleteAction = 1
	c.dashboardDeleteTarget = "d"
	c.Update(
		messages.DashboardMutationFinishedMsg{
			Generation: c.viewGeneration,
			Request:    1,
			Visit:      c.deletionVisit,
			Progress: contracts.DeletionProgress{
				DashboardID: "d",
				Outcome:     contracts.DeletionPendingOutcome,
				Sessions:    []contracts.DeletionSession{{ID: "checkpoint", Removed: true}},
			},
			Err: errors.New("cleanup"),
		},
	)
	c.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	c.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	c.Update(
		messages.DashboardDeletionStatusMsg{
			DashboardID: "d",
			Request:     c.dashboardDeleteRead,
			Visit:       c.deletionVisit,
			Err:         errors.New("status unavailable"),
		},
	)
	rendered := c.View()
	for _, want := range []string{"Deletion is pending", "checkpoint", "status unavailable"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("missing %q: %s", want, rendered)
		}
	}
}

func TestDeletionResultsFromOldVisitsAreIgnored(t *testing.T) {
	c, _ := newLaunchFocusCoordinator(t)
	c.cardDeleteAction = 2
	c.cardDeleteTarget = "new"
	c.Update(
		messages.ScopedActionFinishedMsg{
			DashboardID: c.activeDashboard,
			Generation:  c.viewGeneration,
			Request:     1,
			Visit:       c.deletionVisit,
			Progress:    contracts.DeletionProgress{CardID: "old", Outcome: contracts.DeletionPendingOutcome},
		},
	)
	if strings.Contains(c.View(), "Deletion is pending") {
		t.Fatal("stale card result displayed")
	}
	c.navigateTo(ViewDashboardList)
	c.dashboardDeleteAction = 2
	c.dashboardDeleteTarget = "new"
	c.Update(
		messages.DashboardMutationFinishedMsg{
			Generation: c.viewGeneration,
			Request:    1,
			Visit:      c.deletionVisit,
			Progress:   contracts.DeletionProgress{DashboardID: "old", Outcome: contracts.DeletionPendingOutcome},
		},
	)
	if strings.Contains(c.View(), "Deletion is pending") {
		t.Fatal("stale dashboard result displayed")
	}
}
