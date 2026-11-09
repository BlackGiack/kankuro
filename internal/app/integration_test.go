package app_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/app"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

func TestFullSessionLaunchFlow(t *testing.T) {
	repo, err := newCurrentRepository(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	sm := newTestManager(t, repo.Dir(), repo)
	workspace := t.TempDir()
	if err = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: workspace}}); err != nil {
		t.Fatal(err)
	}
	workspaces, err := repo.ListWorkspaces()
	if err != nil || len(workspaces) != 1 {
		t.Fatalf("workspaces: %v %v", workspaces, err)
	}
	card, err := createCurrentCard(repo, config.IDCard{Name: "Launch test", Driver: "cat", WorkspaceID: workspaces[0].ID})
	if err != nil {
		t.Fatal(err)
	}
	coord := newOpenedCoordinator(repo, sm)
	coord.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	coord.Update(coord.RefreshData()())
	_, cmd := coord.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	var launch messages.SessionLaunchCompletedMsg
	for _, msg := range collectMsgs(t, cmd) {
		if result, ok := msg.(messages.SessionLaunchCompletedMsg); ok {
			launch = result
		}
		coord.Update(msg)
	}
	if launch.Err != nil || launch.SessionID == "" {
		t.Fatalf("launch did not succeed: %+v", launch)
	}
	snap, ok := sm.GetSession(launch.SessionID)
	if !ok || snap.CardID != card.ID || snap.WorkingDir != workspace || !snap.ProcessStarted || snap.Status != session.StatusRunning ||
		snap.PID <= 0 {
		t.Fatalf("missing live activation: %+v", snap)
	}
	bytes, err := os.ReadFile(filepath.Join(repo.Dir(), "sessions", launch.SessionID+".json"))
	if err != nil {
		t.Fatal(err)
	}
	var persisted session.Data
	if err = json.Unmarshal(bytes, &persisted); err != nil {
		t.Fatal(err)
	}
	if persisted.ID != launch.SessionID || persisted.CardID != card.ID || persisted.DashboardID != "default" ||
		persisted.WorkingDir != workspace {
		t.Fatalf("persisted wrong session: %+v", persisted)
	}
}

func TestCardCRUDFlow(t *testing.T) {
	repo, _ := newCurrentRepository(t.TempDir())
	sm := newTestManager(t, repo.Dir(), repo)

	wd := t.TempDir()
	repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	coord := newOpenedCoordinator(repo, sm)
	coord.Init()

	editor := coord.Views()[app.ViewCardEditor]

	testCard := config.IDCard{
		Name:        "New Card",
		Driver:      "claude",
		Model:       "opus",
		WorkspaceID: workspaces[0].ID,
	}

	editor.(*views.CardEditor).SetCard(testCard)

	card := editor.(*views.CardEditor).Card()
	if card.Name != testCard.Name {
		t.Errorf("Expected card name %s, got %s", testCard.Name, card.Name)
	}

	saved, err := createCurrentCard(repo, testCard)
	if err != nil {
		t.Fatalf("CreateIDCard failed: %v", err)
	}

	if saved.ID == "" {
		t.Error("Saved card should have an ID")
	}

	cards, _ := repo.ListIDCards()
	if len(cards) != 1 {
		t.Errorf("Expected 1 card in repo, got %d", len(cards))
	}
}

func TestSourceManagementFlow(t *testing.T) {
	repo, err := newCurrentRepository(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	sm := newTestManager(t, repo.Dir(), repo)
	coord := newOpenedCoordinator(repo, sm)
	coord.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	drive(t, coord, coord.NavigateTo(app.ViewSourceManagerSkills))
	manager := coord.Views()[app.ViewSourceManagerSkills].(*views.SourceManager)
	if len(manager.SourceList().Items) != 0 {
		t.Fatal("expected empty initial list")
	}
	if err = repo.SaveSkillSources([]string{"/first/source", "/second/source"}); err != nil {
		t.Fatal(err)
	}
	coord.Update(coord.RefreshData()())
	items := manager.SourceList().Items
	if len(items) != 2 || items[0].Display() != "/first/source" || items[1].Display() != "/second/source" {
		t.Fatalf("refresh was not propagated to source view: %v", items)
	}
	if err = repo.SaveSkillSources(nil); err != nil {
		t.Fatal(err)
	}
	coord.Update(coord.RefreshData()())
	if len(manager.SourceList().Items) != 0 {
		t.Fatal("refresh did not clear removed sources")
	}
}

func TestSessionManagementFlow(t *testing.T) {
	repo, _ := newCurrentRepository(t.TempDir())
	sm := newTestManager(t, repo.Dir(), repo)

	wd := t.TempDir()
	repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	card, _ := createCurrentCard(repo, config.IDCard{
		Name:        "Test",
		Driver:      "echo",
		Model:       "test",
		WorkspaceID: workspaces[0].ID,
	})

	coord := newOpenedCoordinator(repo, sm)
	coord.Init()

	sessionID, err := launchCurrentSession(repo, sm, card)
	if err != nil {
		t.Fatalf("Launch failed: %v", err)
	}

	sessions := sm.List()
	if len(sessions) != 1 {
		t.Errorf("Expected 1 session, got %d", len(sessions))
	}

	if sessions[0].ID != sessionID {
		t.Error("Session ID mismatch")
	}
}

func TestNavigationFlow(t *testing.T) {
	repo, _ := newCurrentRepository(t.TempDir())
	sm := newTestManager(t, repo.Dir(), repo)

	coord := newOpenedCoordinator(repo, sm)
	coord.Init()

	viewsList := []app.ViewType{
		app.ViewDashboard,
		app.ViewCardEditor,
		app.ViewSourceManagerSkills,
		app.ViewSourceManagerPersonas,
		app.ViewSourceManagerPuppets,
		app.ViewSourceManagerWorkspaces,
		app.ViewSettings,
		app.ViewMCPForm,
	}

	for _, v := range viewsList {
		coord.NavigateTo(v)
		if coord.CurrentView() != v {
			t.Errorf("Expected %v, got %v", v, coord.CurrentView())
		}
	}
}

func TestPreviousViewRestoration(t *testing.T) {
	repo, _ := newCurrentRepository(t.TempDir())
	sm := newTestManager(t, repo.Dir(), repo)

	coord := newOpenedCoordinator(repo, sm)
	coord.Init()

	// 1. Dashboard -> Source Manager -> Back to Dashboard
	coord.NavigateTo(app.ViewDashboard)
	coord.NavigateTo(app.ViewSourceManagerSkills)
	if coord.CurrentView() != app.ViewSourceManagerSkills {
		t.Errorf("Expected ViewSourceManagerSkills, got %v", coord.CurrentView())
	}
	coord.Update(messages.NavigateBackMsg{})
	if coord.CurrentView() != app.ViewDashboard {
		t.Errorf("Expected ViewDashboard after back from source manager, got %v", coord.CurrentView())
	}

	// 2. Card Editor -> Source Manager -> Back to Card Editor
	coord.NavigateTo(app.ViewCardEditor)
	coord.NavigateTo(app.ViewSourceManagerPersonas)
	if coord.CurrentView() != app.ViewSourceManagerPersonas {
		t.Errorf("Expected ViewSourceManagerPersonas, got %v", coord.CurrentView())
	}
	coord.Update(messages.NavigateBackMsg{})
	if coord.CurrentView() != app.ViewCardEditor {
		t.Errorf("Expected ViewCardEditor after back from source manager, got %v", coord.CurrentView())
	}

	// 3. Card Editor -> MCP Form -> Back to Card Editor
	coord.NavigateTo(app.ViewCardEditor)
	coord.NavigateTo(app.ViewMCPForm)
	if coord.CurrentView() != app.ViewMCPForm {
		t.Errorf("Expected ViewMCPForm, got %v", coord.CurrentView())
	}
	coord.Update(messages.NavigateBackMsg{})
	if coord.CurrentView() != app.ViewCardEditor {
		t.Errorf("Expected ViewCardEditor after back from MCP form, got %v", coord.CurrentView())
	}
}

func TestSessionActionsFlow(t *testing.T) {
	repo, _ := newCurrentRepository(t.TempDir())
	sm := newTestManager(t, repo.Dir(), repo)

	wd := t.TempDir()
	repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	card, _ := createCurrentCard(repo, config.IDCard{
		Name:        "Test Card",
		Driver:      "cat",
		WorkspaceID: workspaces[0].ID,
	})

	coord := newOpenedCoordinator(repo, sm)
	coord.Init()

	// Launch a session
	sessionID, err := launchCurrentSession(repo, sm, card)
	if err != nil {
		t.Fatalf("Launch failed: %v", err)
	}

	msg := coord.RefreshData()()
	coord.Update(msg)

	// Test Attach on running session
	// We need to simulate the Enter key on the session list
	// but we can also test the callback directly via coordinator
	callbacks := coord.DashboardCallbacks()

	// Running session -> Attach
	cmd := callbacks.OnAttach(sessionID)
	if cmd == nil {
		t.Error("Expected cmd for attach")
	}

	// Terminate the session
	_ = sm.Kill(sessionID)
	msg = coord.RefreshData()()
	coord.Update(msg)

	// Terminated session -> Relaunch
	cmd = callbacks.OnRelaunch(sessionID)
	if cmd == nil {
		t.Error("Expected cmd for relaunch")
	}

	// Terminated session -> Switch Agent
	newCard, err := createCurrentCard(repo, config.IDCard{
		Name:        "New Agent",
		Driver:      "cat",
		WorkspaceID: workspaces[0].ID,
	})
	if err != nil {
		t.Fatalf("CreateIDCard failed: %v", err)
	}
	// Agent-switch is only reachable through AgentSwitchCallbacks (the
	// AgentSwitchView's callback), which - unlike DashboardCallbacks -
	// actually gets invoked by real input. It's async (the blocking
	// SwitchAgent call runs inside the returned Cmd, not synchronously
	// here), so drive it through Update the way a real Bubble Tea runtime
	// would: execute the Cmd, feed the resulting message back in.
	switchCmd := coord.AgentSwitchCallbacks().OnSwitch(sessionID, newCard.ID, newCard.Driver)
	if switchCmd == nil {
		t.Fatal("Expected cmd for switch")
	}
	result := switchCmd()
	if switched, ok := result.(messages.AgentSwitchedMsg); !ok || switched.Err != nil {
		t.Fatalf("switch result: %#v", result)
	}
	coord.Update(result)

	// Verify session now uses new card
	sess, ok := sm.GetSession(sessionID)
	if !ok {
		t.Fatalf("Session %s not found", sessionID)
	}
	if sess.CardID != newCard.ID {
		t.Errorf("Expected card ID %s, got %s", newCard.ID, sess.CardID)
	}

	// Terminated session -> Remove
	// We must kill it again because SwitchAgent started it
	_ = sm.Kill(sessionID)
	cmd, err = callbacks.OnRemove(sessionID)
	if err != nil {
		t.Errorf("Remove failed: %v", err)
	}
	if cmd != nil {
		cmd()
	}
	msg = coord.RefreshData()()
	coord.Update(msg)
	if len(sm.List()) != 0 {
		t.Error("Expected 0 sessions after removal")
	}
}

func TestSourceSyncSideEffects(t *testing.T) {
	repo, _ := newCurrentRepository(t.TempDir())
	sm := newTestManager(t, repo.Dir(), repo)

	coord := newOpenedCoordinator(repo, sm)
	coord.Init()

	callbacks := coord.SourceManagerCallbacks(views.SourceTypeSkills)

	// Add a skill source (dummy path)
	cmd, err := callbacks.SaveSimpleSource("/tmp/skills", -1)
	if err != nil {
		t.Fatalf("SaveSimpleSource failed: %v", err)
	}
	if cmd != nil {
		cmd()
	}

	sources, _ := repo.ListSkillSources()
	if len(sources) != 1 || sources[0] != "/tmp/skills" {
		t.Errorf("Expected 1 skill source, got %v", sources)
	}

	// Syncing happens in the callback. We can't easily verify the side effect of SyncSkillsFromSources
	// without real files, but we verified the repo call was made.
}

func TestCoordinatorErrorHandling(t *testing.T) {
	repo, _ := newCurrentRepository(t.TempDir())
	sm := newTestManager(t, repo.Dir(), repo)

	coord := newOpenedCoordinator(repo, sm)
	coord.Init()

	coord.Update(messages.ErrorMsg{Err: errTest})

	if coord.Err() == nil {
		t.Error("Expected error to be set")
	}

	coord.Update(messages.ClearErrorMsg{})

	if coord.Err() != nil {
		t.Error("Expected error to be cleared")
	}
}

func TestCoordinatorDataPropagation(t *testing.T) {
	repo, err := newCurrentRepository(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	sm := newTestManager(t, repo.Dir(), repo)
	coord := newOpenedCoordinator(repo, sm)
	coord.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	dashboard := coord.Views()[app.ViewDashboard].(*views.Dashboard)
	coord.Update(coord.RefreshData()())
	if len(dashboard.Cards()) != 0 {
		t.Fatal("expected initially empty dashboard")
	}
	if err = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: t.TempDir()}}); err != nil {
		t.Fatal(err)
	}
	workspaces, err := repo.ListWorkspaces()
	if err != nil || len(workspaces) != 1 {
		t.Fatalf("workspaces: %v %v", workspaces, err)
	}
	card, err := createCurrentCard(repo, config.IDCard{Name: "Propagated", Driver: "cat", WorkspaceID: workspaces[0].ID})
	if err != nil {
		t.Fatal(err)
	}
	coord.Update(coord.RefreshData()())
	cards := dashboard.Cards()
	if len(cards) != 1 || cards[0].ID != card.ID || cards[0].Name != "Propagated" {
		t.Fatalf("dashboard cards: %+v", cards)
	}
	card.Name = "Updated"
	if _, err = repo.UpdateIDCard(card); err != nil {
		t.Fatal(err)
	}
	coord.Update(coord.RefreshData()())
	if got := dashboard.Cards(); len(got) != 1 || got[0].Name != "Updated" {
		t.Fatalf("stale dashboard cards: %+v", got)
	}
}

func TestSessionLauncherBehavior(t *testing.T) {
	repo, _ := newCurrentRepository(t.TempDir())
	sm := newTestManager(t, repo.Dir(), repo)

	wd := t.TempDir()
	repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	card, _ := createCurrentCard(repo, config.IDCard{
		Name:        "Test Card",
		Driver:      "echo",
		Model:       "test",
		WorkspaceID: workspaces[0].ID,
	})

	coord := newOpenedCoordinator(repo, sm)
	coord.Init()

	// Navigate to session launcher
	launcher := coord.Views()[app.ViewSessionLauncher].(*views.SessionLauncher)
	launcher.SetCard(card)
	launcher.SetName("Test Session")

	// Simulate async launch completion
	sessionID, err := launchCurrentSession(repo, sm, card)
	if err != nil {
		t.Fatalf("Launch failed: %v", err)
	}

	// Simulate the launch completion message
	launcher.SetLaunchResult(sessionID, nil)

	if !launcher.LaunchDone() {
		t.Error("Expected launch to be marked as done")
	}

	if launcher.LaunchErr() != nil {
		t.Errorf("Expected no launch error, got %v", launcher.LaunchErr())
	}

	if launcher.SessionID() != sessionID {
		t.Errorf("Expected session ID %s, got %s", sessionID, launcher.SessionID())
	}

	// Simulate name submission
	_, nameCmd := launcher.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if nameCmd == nil {
		t.Fatal("name submission returned no command")
	}
	nameResult, ok := nameCmd().(messages.ScopedSessionNamedMsg)
	if !ok || nameResult.Err != nil {
		t.Fatalf("name submission: %#v", nameResult)
	}
	coord.Update(nameResult)

	// Verify session was created with correct name
	sessions := sm.List()
	if len(sessions) != 1 {
		t.Fatalf("Expected 1 session, got %d", len(sessions))
	}

	if sessions[0].Name != "Test Session" {
		t.Errorf("Expected session name 'Test Session', got '%s'", sessions[0].Name)
	}

	if sessions[0].ID != sessionID {
		t.Error("Session ID mismatch")
	}
}

var errTest = &testError{}

type testError struct{}

func (e *testError) Error() string {
	return "test error"
}
