package app_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/BlackGiack/kankuro/internal/app"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

func TestMCPManagerStaleRequests(t *testing.T) {
	repo, _ := newCurrentRepository(t.TempDir())
	server, err := repo.UpsertMCP(config.MCPServer{Name: "Alpha", Command: "cmd"})
	if err != nil {
		t.Fatal(err)
	}
	sm := newTestManager(t, repo.Dir(), repo)
	coord := newScopedCoordinator(repo, sm)
	coord.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	openDefaultDashboard(coord)
	servers, err := repo.ListMCPs()
	if err != nil {
		t.Fatal(err)
	}
	coord.Update(messages.DataLoadedMsg{DashboardID: "default", Generation: 1, MCPs: servers})
	open := func() { t.Helper(); _, cmd := coord.Update(tea.KeyPressMsg{Text: "m"}); coord.Update(cmd()) }
	leave := func() { t.Helper(); _, cmd := coord.Update(tea.KeyPressMsg{Code: tea.KeyEsc}); coord.Update(cmd()) }
	open()
	_, heldEdit := coord.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	leave()
	open()
	coord.Update(heldEdit())
	if coord.CurrentView() != app.ViewMCPManager {
		t.Fatal("stale edit opened form")
	}
	coord.Update(tea.KeyPressMsg{Code: tea.KeyDelete})
	_, heldDelete := coord.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if heldDelete == nil {
		t.Fatal("confirmation did not produce request")
	}
	leave()
	open()
	coord.Update(heldDelete())
	coord.Update(tea.KeyPressMsg{Code: tea.KeyDelete})
	_, deleteRequest := coord.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if deleteRequest == nil {
		t.Fatal("discarded request stranded deletion")
	}
	_, deleteCmd := coord.Update(deleteRequest())
	if deleteCmd == nil {
		t.Fatal("fresh deletion was not admitted")
	}
	result := deleteCmd()
	leave()
	open()
	coord.Update(result)
	if coord.CurrentView() != app.ViewMCPManager {
		t.Fatal("delayed result changed the current view")
	}
	servers, err = repo.ListMCPs()
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range servers {
		if item.ID == server.ID {
			t.Fatal("fresh deletion did not remove server")
		}
	}
}

func TestMCPManagerEditPreservesMetadataAndNewIsBlank(t *testing.T) {
	repo, _ := newCurrentRepository(t.TempDir())
	original, err := repo.UpsertMCP(config.MCPServer{Name: "Alpha", Command: "cmd", Description: "keep this"})
	if err != nil {
		t.Fatal(err)
	}
	sm := newTestManager(t, repo.Dir(), repo)
	coord := newScopedCoordinator(repo, sm)
	coord.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	openDefaultDashboard(coord)
	servers, err := repo.ListMCPs()
	if err != nil {
		t.Fatal(err)
	}
	coord.Update(messages.DataLoadedMsg{DashboardID: "default", Generation: 1, MCPs: servers})
	_, cmd := coord.Update(tea.KeyPressMsg{Text: "m"})
	coord.Update(cmd())
	_, cmd = coord.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	coord.Update(cmd())
	form := coord.Views()[app.ViewMCPForm].(*views.MCPForm)
	form.Inputs()[0].SetValue("Updated")
	_, cmd = coord.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	saved := cmd().(messages.MCPSavedMsg)
	if saved.Err != nil || saved.MCP.ID != original.ID || saved.MCP.Description != original.Description {
		t.Fatalf("metadata lost: %#v", saved)
	}
	coord.Update(saved)
	servers, err = repo.ListMCPs()
	if err != nil {
		t.Fatal(err)
	}
	coord.Update(messages.DataLoadedMsg{DashboardID: "default", Generation: 1, MCPs: servers})
	// Move from the selected custom server through the built-in row to New.
	coord.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	coord.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	_, cmd = coord.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	coord.Update(cmd())
	form = coord.Views()[app.ViewMCPForm].(*views.MCPForm)
	for _, input := range form.Inputs() {
		if input.Value() != "" {
			t.Fatal("new form retained old value")
		}
	}
	form.Inputs()[0].SetValue("New")
	form.Inputs()[1].SetValue("cmd")
	_, cmd = coord.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	created := cmd().(messages.MCPSavedMsg)
	if created.Err != nil || created.MCP.ID == original.ID || created.MCP.Description != "" {
		t.Fatalf("new inherited metadata: %#v", created)
	}
}

func TestMCPManagerReferencedDeleteReportsError(t *testing.T) {
	repo, _ := newCurrentRepository(t.TempDir())
	server, err := repo.UpsertMCP(config.MCPServer{Name: "Alpha", Command: "cmd"})
	if err != nil {
		t.Fatal(err)
	}
	if err = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: t.TempDir()}}); err != nil {
		t.Fatal(err)
	}
	workspaces, err := repo.ListWorkspaces()
	if err != nil || len(workspaces) == 0 {
		t.Fatalf("workspaces: %v", err)
	}
	_, err = repo.CreateIDCard(config.IDCard{
		DashboardID: "default", Name: "Card", Driver: "cat", WorkspaceID: workspaces[0].ID, MCPIDs: []string{server.ID},
	})
	if err != nil {
		t.Fatal(err)
	}
	sm := newTestManager(t, repo.Dir(), repo)
	coord := newScopedCoordinator(repo, sm)
	coord.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	openDefaultDashboard(coord)
	servers, err := repo.ListMCPs()
	if err != nil {
		t.Fatal(err)
	}
	coord.Update(messages.DataLoadedMsg{DashboardID: "default", Generation: 1, MCPs: servers})
	_, cmd := coord.Update(tea.KeyPressMsg{Text: "m"})
	coord.Update(cmd())
	coord.Update(tea.KeyPressMsg{Code: tea.KeyDelete})
	_, cmd = coord.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	_, cmd = coord.Update(cmd())
	result := cmd().(messages.MCPManagerDeletedMsg)
	if result.Err == nil {
		t.Fatal("referenced server deletion succeeded")
	}
	coord.Update(result)
	if coord.Err() == nil || coord.CurrentView() != app.ViewMCPManager {
		t.Fatal("deletion failure not shown in manager")
	}
	servers, err = repo.ListMCPs()
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, item := range servers {
		found = found || item.ID == server.ID
	}
	if !found {
		t.Fatal("referenced server was removed")
	}
}
