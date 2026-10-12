package app_test

import (
	"testing"

	"github.com/BlackGiack/kankuro/internal/app"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

func TestFullSessionLaunchFlow(t *testing.T) {
	repo, _ := config.NewFileRepository(t.TempDir())
	sm := newTestManager(t, repo.Dir(), repo)

	wd := t.TempDir()
	repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})

	_, _ = repo.CreateIDCard(config.IDCard{
		Name:   "Test Card",
		Driver: "echo",
		Model:  "test",
	})

	coord := app.NewCoordinator(repo, sm)
	msg := coord.RefreshData()()
	coord.Update(msg)

	if coord.CurrentView() != app.ViewDashboard {
		t.Errorf("Expected ViewDashboard, got %v", coord.CurrentView())
	}

	if len(coord.Views()) != 10 {
		t.Errorf("Expected 10 views, got %d", len(coord.Views()))
	}
}

func TestCardCRUDFlow(t *testing.T) {
	repo, _ := config.NewFileRepository(t.TempDir())
	sm := newTestManager(t, repo.Dir(), repo)

	wd := t.TempDir()
	repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	coord := app.NewCoordinator(repo, sm)
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

	saved, err := repo.CreateIDCard(testCard)
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
	repo, _ := config.NewFileRepository(t.TempDir())
	sm := newTestManager(t, repo.Dir(), repo)

	coord := app.NewCoordinator(repo, sm)
	coord.Init()

	coord.NavigateTo(app.ViewSourceManagerSkills)
	sourceManager := coord.Views()[app.ViewSourceManagerSkills].(*views.SourceManager)

	// In the original test it was sourceManager.SetData() but it requires an argument.
	// We'll use the coordinator's DataLoadedMsg propagation.
	msg := coord.RefreshData()()
	coord.Update(msg)

	// Check if sourceManager received data (indirectly)
	if sourceManager == nil {
		t.Fatal("SourceManager view not found")
	}

	sources, _ := repo.ListSkillSources()
	if len(sources) != 0 {
		t.Errorf("Expected 0 skill sources, got %d", len(sources))
	}
}

func TestSessionManagementFlow(t *testing.T) {
	repo, _ := config.NewFileRepository(t.TempDir())
	sm := newTestManager(t, repo.Dir(), repo)

	wd := t.TempDir()
	repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	card, _ := repo.CreateIDCard(config.IDCard{
		Name:        "Test",
		Driver:      "echo",
		Model:       "test",
		WorkspaceID: workspaces[0].ID,
	})

	coord := app.NewCoordinator(repo, sm)
	coord.Init()

	sessionID, err := sm.Launch(card, nil, nil, nil)
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
	repo, _ := config.NewFileRepository(t.TempDir())
	sm := newTestManager(t, repo.Dir(), repo)

	coord := app.NewCoordinator(repo, sm)
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
	repo, _ := config.NewFileRepository(t.TempDir())
	sm := newTestManager(t, repo.Dir(), repo)

	coord := app.NewCoordinator(repo, sm)
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
	repo, _ := config.NewFileRepository(t.TempDir())
	sm := newTestManager(t, repo.Dir(), repo)

	wd := t.TempDir()
	repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	card, _ := repo.CreateIDCard(config.IDCard{
		Name:        "Test Card",
		Driver:      "echo",
		Model:       "test",
		WorkspaceID: workspaces[0].ID,
	})

	coord := app.NewCoordinator(repo, sm)
	coord.Init()

	// Launch a session
	sessionID, err := sm.Launch(card, nil, nil, nil)
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
	newCard, err := repo.CreateIDCard(config.IDCard{
		Name:        "New Agent",
		Driver:      "echo",
		Model:       "test",
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
	switchCmd := coord.AgentSwitchCallbacks().OnSwitch(sessionID, newCard.ID)
	if switchCmd == nil {
		t.Fatal("Expected cmd for switch")
	}
	coord.Update(switchCmd())

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
	repo, _ := config.NewFileRepository(t.TempDir())
	sm := newTestManager(t, repo.Dir(), repo)

	coord := app.NewCoordinator(repo, sm)
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
	repo, _ := config.NewFileRepository(t.TempDir())
	sm := newTestManager(t, repo.Dir(), repo)

	coord := app.NewCoordinator(repo, sm)
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
	repo, _ := config.NewFileRepository(t.TempDir())
	sm := newTestManager(t, repo.Dir(), repo)

	wd := t.TempDir()
	repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	_, _ = repo.CreateIDCard(config.IDCard{
		Name:        "Test",
		Driver:      "claude",
		Model:       "opus",
		WorkspaceID: workspaces[0].ID,
	})

	coord := app.NewCoordinator(repo, sm)
	coord.Init()

	msg := coord.RefreshData()()
	coord.Update(msg)

	cards, _ := repo.ListIDCards()
	if len(cards) != 1 {
		t.Errorf("Expected 1 card in repo, got %d", len(cards))
	}
}

func TestSessionLauncherBehavior(t *testing.T) {
	repo, _ := config.NewFileRepository(t.TempDir())
	sm := newTestManager(t, repo.Dir(), repo)

	wd := t.TempDir()
	repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	card, _ := repo.CreateIDCard(config.IDCard{
		Name:        "Test Card",
		Driver:      "echo",
		Model:       "test",
		WorkspaceID: workspaces[0].ID,
	})

	coord := app.NewCoordinator(repo, sm)
	coord.Init()

	// Navigate to session launcher
	launcher := coord.Views()[app.ViewSessionLauncher].(*views.SessionLauncher)
	launcher.SetCard(card)
	launcher.SetName("Test Session")

	// Simulate async launch completion
	sessionID, err := sm.Launch(card, nil, nil, nil)
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
	err = sm.SetName(sessionID, "Test Session")
	if err != nil {
		t.Fatalf("SetName failed: %v", err)
	}

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
