package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

func TestInventoryPublishersShareOrdering(t *testing.T) {
	tests := []struct {
		name     string
		olderCmd func(*Coordinator) tea.Cmd
		newerCmd func(*Coordinator) tea.Cmd
	}{
		{"older session poll after newer full", (*Coordinator).refreshSessions, (*Coordinator).refreshDataSeeded},
		{"older full after newer session poll", (*Coordinator).refreshDataSeeded, (*Coordinator).refreshSessions},
		{"session polls complete in reverse", (*Coordinator).refreshSessions, (*Coordinator).refreshSessions},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _, _ := newCountingCoordinator(t)
			dashboard := c.views[ViewDashboard].(*views.Dashboard)
			reader := &orderedInventoryReader{calls: make(chan *orderedInventoryCall, 2)}
			c.reads = reader
			oldDone := runInventoryCommand(tt.olderCmd(c))
			oldRead := <-reader.calls
			newDone := runInventoryCommand(tt.newerCmd(c))
			newRead := <-reader.calls
			oldSnapshot := inventorySnapshot("old", "waiting").Value
			newSnapshot := inventorySnapshot("new", "running").Value
			newRead.complete(newSnapshot)
			c.Update(<-newDone)
			oldRead.complete(oldSnapshot)
			c.Update(<-oldDone)
			rows := dashboard.Sessions()
			if len(rows) != 1 || rows[0].Name != "new" || rows[0].Status != "running" || !rows[0].Stop.Enabled {
				t.Fatalf("older publication regressed row/action facts: %+v", rows)
			}
			if selected, ok := dashboard.SessionList().Selected(); !ok || selected.ID != "session" {
				t.Fatalf("selected identity changed: %+v, %v", selected, ok)
			}
			if dashboard.CenterSessionID() != "session" {
				t.Fatalf("preview selection changed: %q", dashboard.CenterSessionID())
			}
		})
	}
}

type orderedInventoryCall struct {
	result chan contracts.DashboardInventory
}

func (call *orderedInventoryCall) complete(inventory contracts.DashboardInventory) {
	call.result <- inventory
}

type orderedInventoryReader struct {
	calls chan *orderedInventoryCall
}

func (*orderedInventoryReader) ReadDashboardInventory() ([]contracts.DashboardSummary, error) {
	return nil, nil
}
func (reader *orderedInventoryReader) ReadDashboardOverview(string) (contracts.DashboardInventory, error) {
	call := &orderedInventoryCall{result: make(chan contracts.DashboardInventory, 1)}
	reader.calls <- call
	return <-call.result, nil
}
func (*orderedInventoryReader) ReadDisplayChoices() contracts.DisplayChoicesResult {
	return emptySuccessfulChoices()
}
func (*orderedInventoryReader) ReadRegistryChangeToken() (contracts.RegistryChangeToken, error) {
	return "stable", nil
}

func runInventoryCommand(cmd tea.Cmd) <-chan tea.Msg {
	done := make(chan tea.Msg, 1)
	go func() { done <- cmd() }()
	return done
}

func inventorySnapshot(name, status string) contracts.ReadResult[contracts.DashboardInventory] {
	session := contracts.SessionSummary{
		ID: "session", DashboardID: "default", CardID: "card", Name: name, Status: status,
		Stop: contracts.Availability{Enabled: status == "running"},
	}
	return contracts.ReadResult[contracts.DashboardInventory]{
		Attempted: true,
		Value:     contracts.DashboardInventory{Sessions: []contracts.SessionSummary{session}},
	}
}

func (*orderedInventoryReader) ReadLaunchPreparation(string, string) (contracts.LaunchPreparation, error) {
	return contracts.LaunchPreparation{}, nil
}
