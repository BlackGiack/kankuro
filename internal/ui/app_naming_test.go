//nolint:testpackage // These tests exercise unexported UI state transitions directly.
package ui

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"kankuro/internal/config"
	"kankuro/internal/session"
)

func TestNamingFlowSubmitsBeforeLaunchCompletes(t *testing.T) {
	app, _ := newNamingTestApp(t)

	_, launchCmd := app.handleDashboardEnter()
	if launchCmd == nil {
		t.Fatal("expected launch command")
	}
	if app.mode != viewNaming {
		t.Fatalf("expected naming view, got %v", app.mode)
	}

	app.namingInput.SetValue("feature run")

	model, submitCmd := app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	app = model.(*App)
	if submitCmd != nil {
		t.Fatal("expected no attach command before launch completes")
	}
	if !app.namingSubmitted {
		t.Fatal("expected naming submission to be recorded")
	}

	msg := launchCmd()
	launchFinished, ok := msg.(launchFinishedMsg)
	if !ok {
		t.Fatalf("expected launchFinishedMsg, got %T", msg)
	}
	defer killSessionIfNeeded(t, app.session, launchFinished.sessionID)

	model, attachCmd := app.Update(launchFinished)
	app = model.(*App)
	if attachCmd == nil {
		t.Fatal("expected attach command after launch completion")
	}
	if app.mode != viewDashboard {
		t.Fatalf("expected dashboard view, got %v", app.mode)
	}
	if app.activeSessionID != launchFinished.sessionID {
		t.Fatalf("expected active session %q, got %q", launchFinished.sessionID, app.activeSessionID)
	}

	assertSessionName(t, app.session, launchFinished.sessionID, "feature run")
}

func TestNamingFlowUsesSessionIDWhenBlank(t *testing.T) {
	app, _ := newNamingTestApp(t)

	_, launchCmd := app.handleDashboardEnter()
	if launchCmd == nil {
		t.Fatal("expected launch command")
	}

	msg := launchCmd()
	launchFinished, ok := msg.(launchFinishedMsg)
	if !ok {
		t.Fatalf("expected launchFinishedMsg, got %T", msg)
	}
	defer killSessionIfNeeded(t, app.session, launchFinished.sessionID)

	model, cmd := app.Update(launchFinished)
	app = model.(*App)
	if cmd != nil {
		t.Fatal("expected no attach command until name is confirmed")
	}
	if app.mode != viewNaming {
		t.Fatalf("expected naming view, got %v", app.mode)
	}

	app.namingInput.SetValue("   ")
	model, attachCmd := app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	app = model.(*App)
	if attachCmd == nil {
		t.Fatal("expected attach command after confirming blank name")
	}

	assertSessionName(t, app.session, launchFinished.sessionID, launchFinished.sessionID)
}

func TestNamingFlowReturnsToDashboardOnLaunchError(t *testing.T) {
	app, _ := newNamingTestApp(t)

	_, launchCmd := app.handleDashboardEnter()
	if launchCmd == nil {
		t.Fatal("expected launch command")
	}

	launchErr := errors.New("launch failed")
	model, cmd := app.Update(launchFinishedMsg{err: launchErr})
	app = model.(*App)
	if cmd != nil {
		t.Fatal("expected no command on launch failure")
	}
	if app.mode != viewDashboard {
		t.Fatalf("expected dashboard view, got %v", app.mode)
	}
	if !errors.Is(app.err, launchErr) {
		t.Fatalf("expected error %v, got %v", launchErr, app.err)
	}
}

func TestRefreshSessionListShowsCardNameAboveFace(t *testing.T) {
	app, card := newNamingTestApp(t)

	sessionID, err := app.session.Launch(card, nil, nil, nil)
	if err != nil {
		t.Fatalf("launch session: %v", err)
	}
	defer killSessionIfNeeded(t, app.session, sessionID)

	if setNameErr := app.session.SetName(sessionID, "release-check"); setNameErr != nil {
		t.Fatalf("set session name: %v", setNameErr)
	}

	app.refreshSessionList()

	item, ok := app.sessionsList.Items()[0].(sessionListItem)
	if !ok {
		t.Fatalf("expected sessionListItem, got %T", app.sessionsList.Items()[0])
	}
	if item.Title() != "release-check" {
		t.Fatalf("expected title to be session name, got %q", item.Title())
	}
	desc := item.Description()
	if !strings.Contains(desc, "builder ") {
		t.Fatalf("expected description to contain card name, got %q", desc)
	}
	if strings.Contains(desc, "release-check ") {
		t.Fatalf("expected description not to repeat session name, got %q", desc)
	}
}

func newNamingTestApp(t *testing.T) (*App, config.IDCard) {
	t.Helper()

	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	wd := t.TempDir()
	if saveErr := repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}}); saveErr != nil {
		t.Fatalf("save workspace source: %v", saveErr)
	}
	workspaces, err := repo.ListWorkspaces()
	if err != nil {
		t.Fatalf("list workspaces: %v", err)
	}
	if len(workspaces) == 0 {
		t.Fatal("expected at least one workspace")
	}

	card, err := repo.CreateIDCard(config.IDCard{
		Name:        "builder",
		Driver:      "cat",
		WorkspaceID: workspaces[0].ID,
		FaceStyle:   "happy",
	})
	if err != nil {
		t.Fatalf("create card: %v", err)
	}

	sm := session.NewManager(t.TempDir(), repo)
	app, err := New(repo, sm)
	if err != nil {
		t.Fatalf("new app: %v", err)
	}

	return app, card
}

func assertSessionName(t *testing.T, sm *session.Manager, id, want string) {
	t.Helper()

	s, ok := sm.GetSession(id)
	if !ok {
		t.Fatalf("session %q not found", id)
	}
	if s.Name != want {
		t.Fatalf("expected session name %q, got %q", want, s.Name)
	}
}

func killSessionIfNeeded(t *testing.T, sm *session.Manager, id string) {
	t.Helper()
	if id == "" {
		return
	}
	if err := sm.Kill(id); err != nil && !strings.Contains(err.Error(), "session not found") {
		t.Fatalf("kill session %q: %v", id, err)
	}
}
