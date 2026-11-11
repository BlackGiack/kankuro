package app

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

func TestInlineRenameRejectsOlderRefreshSnapshots(t *testing.T) {
	for _, kind := range []string{"full", "sessions", "fingerprint"} {
		t.Run(kind, func(t *testing.T) {
			coordinator, dashboard, id := newInlineRenameCoordinator(t)
			before, ok := coordinator.session.GetSession(id)
			if !ok || before.PID <= 0 || !before.ProcessStarted {
				t.Fatalf("missing live session: %+v", before)
			}
			// Execute the reads before the rename, but withhold their results
			// until the rename's own refresh has installed the new name.
			full, ok := coordinator.refreshDataSeeded()().(configFingerprintMsg)
			if !ok || full.data == nil || full.data.Err != nil {
				t.Fatalf("full refresh failed: %+v", full)
			}
			var stale tea.Msg
			switch kind {
			case "full":
				stale = *full.data
			case "sessions":
				stale = coordinator.refreshSessions()()
			case "fingerprint":
				stale = full
			}

			coordinator.Update(tea.KeyPressMsg{Text: "e"})
			if !strings.Contains(dashboard.View(), "Rename Session") {
				t.Fatal("session edit did not open the inline rename dialog")
			}
			coordinator.Update(tea.KeyPressMsg{Code: 'a', Mod: tea.ModCtrl})
			coordinator.Update(tea.KeyPressMsg{Code: 'k', Mod: tea.ModCtrl})
			const renamed = "Renamed 日本語 session"
			coordinator.Update(tea.KeyPressMsg{Text: renamed})
			_, submit := coordinator.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
			if submit == nil {
				t.Fatal("inline rename returned no save command")
			}
			result, ok := submit().(messages.DashboardSessionRenamedMsg)
			if !ok || result.Err != nil || result.SessionID != id {
				t.Fatalf("inline rename failed: %+v", result)
			}
			_, refresh := coordinator.Update(result)
			if refresh == nil {
				t.Fatal("successful inline rename returned no refresh")
			}
			current, ok := refresh().(configFingerprintMsg)
			if !ok || current.data == nil || current.data.Err != nil {
				t.Fatalf("rename must return a full refresh, not attachment: %+v", current)
			}
			coordinator.Update(current)
			assertInlineRenamedSession(t, coordinator, dashboard, id, renamed)

			coordinator.Update(stale)
			assertInlineRenamedSession(t, coordinator, dashboard, id, renamed)
			after, ok := coordinator.session.GetSession(id)
			if !ok || after.PID != before.PID || after.ActivationID != before.ActivationID || !after.ProcessStarted {
				t.Fatalf("rename changed activation: before=%+v after=%+v", before, after)
			}
		})
	}
}

func newInlineRenameCoordinator(t *testing.T) (*Coordinator, *views.Dashboard, string) {
	t.Helper()
	coordinator, dashboard := newLaunchFocusCoordinator(t)
	if err := coordinator.repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: t.TempDir()}}); err != nil {
		t.Fatal(err)
	}
	workspaces, err := coordinator.repo.ListWorkspaces()
	if err != nil || len(workspaces) != 1 {
		t.Fatalf("workspace discovery: %v, %v", workspaces, err)
	}
	card, err := coordinator.repo.CreateIDCard(config.IDCard{
		DashboardID: "default", Name: "Original session", Driver: "cat", WorkspaceID: workspaces[0].ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	id, err := coordinator.actions.service.LaunchInDashboard(t.Context(), "default", card.ID)
	if err != nil {
		t.Fatal(err)
	}
	coordinator.Update(coordinator.refreshDataSeeded()())
	coordinator.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	dashboard.SessionList().SetSelectedByID(id)
	dashboard.SyncPreviewSelection()
	return coordinator, dashboard, id
}

func assertInlineRenamedSession(t *testing.T, coordinator *Coordinator, dashboard *views.Dashboard, id, name string) {
	t.Helper()
	if coordinator.currentView != ViewDashboard || coordinator.err != nil || strings.Contains(dashboard.View(), "Rename Session") {
		t.Fatalf("inline rename did not leave a usable dashboard: view=%v err=%v", coordinator.currentView, coordinator.err)
	}
	assertFocusedSession(t, dashboard, id)
	selected, ok := dashboard.SessionList().Selected()
	if !ok || selected.Name != name || !strings.Contains(dashboard.Preview().View(), name) {
		t.Fatalf("refresh displaced renamed selection or preview: selected=%+v preview=%q", selected, dashboard.Preview().View())
	}
}
