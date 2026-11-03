package app_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/app"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

// Preview messages must not replace card-editor registries. Rearming the subscription
// after navigation exercises receiver isolation independently of unsubscribe.
func TestPartialPreviewBroadcastDoesNotWipeCardEditor(t *testing.T) {
	repo, _ := newCurrentRepository(t.TempDir())
	sm := newTestManager(t, repo.Dir(), repo)

	wd := t.TempDir()
	_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	card, errCreate := createCurrentCard(repo, config.IDCard{
		Name:        "With Skills",
		Driver:      "cat",
		WorkspaceID: workspaces[0].ID,
		SkillIDs:    []string{config.BuiltinSkillMCPCLI},
	})
	if errCreate != nil {
		t.Fatalf("create card: %v", errCreate)
	}

	sessionID, errLaunch := launchCurrentSession(repo, sm, card)
	if errLaunch != nil {
		t.Fatalf("launch: %v", errLaunch)
	}
	defer func() { _ = sm.Kill(sessionID) }()

	coord := newOpenedCoordinator(repo, sm)
	coord.Init()
	coord.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	coord.Update(coord.RefreshData()())

	editor, ok := coord.Views()[app.ViewCardEditor].(*views.CardEditor)
	if !ok {
		t.Fatal("expected *views.CardEditor for ViewCardEditor")
	}
	editor.SetCard(card)
	coord.Update(messages.ScopedNavigateMsg{View: messages.ViewCardEditor, DashboardID: "default", Generation: 1})

	if got := editor.Card().SkillIDs; len(got) != 1 || got[0] != config.BuiltinSkillMCPCLI {
		t.Fatalf("card editor lost its skills before the broadcast even ran: %v", got)
	}

	// Re-arm the preview subscription for the session while the card editor
	// is the active view (bypassing #12's reachability fix), then deliver
	// output for it - simulating a SessionOutputMsg that arrived for a
	// previously previewed session while the editor is open.
	coord.Update(messages.SessionPreviewRequestedMsg{SessionID: sessionID})
	coord.Update(messages.SessionOutputMsg{SessionID: sessionID, Output: []byte("hello")})

	got := editor.Card().SkillIDs
	if len(got) != 1 || got[0] != config.BuiltinSkillMCPCLI {
		t.Fatalf("card editor's skills were wiped by a partial preview broadcast: %v", got)
	}
}

// A failed skills read leaves the card editor's previous data and selection intact.
func TestTransientSkillsReadFailureDoesNotWipeCardEditor(t *testing.T) {
	repo, errRepo := newCurrentRepository(t.TempDir())
	if errRepo != nil {
		t.Fatalf("new repo: %v", errRepo)
	}
	sm := newTestManager(t, repo.Dir(), repo)

	wd := t.TempDir()
	_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	card, errCreate := createCurrentCard(repo, config.IDCard{
		Name:        "With Skills",
		Driver:      "cat",
		WorkspaceID: workspaces[0].ID,
		SkillIDs:    []string{config.BuiltinSkillMCPCLI},
	})
	if errCreate != nil {
		t.Fatalf("create card: %v", errCreate)
	}

	coord := newOpenedCoordinator(repo, sm)
	coord.Init()
	coord.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	coord.Update(coord.RefreshData()())

	editor, ok := coord.Views()[app.ViewCardEditor].(*views.CardEditor)
	if !ok {
		t.Fatal("expected *views.CardEditor for ViewCardEditor")
	}
	editor.SetCard(card)

	if got := editor.Card().SkillIDs; len(got) != 1 || got[0] != config.BuiltinSkillMCPCLI {
		t.Fatalf("card editor lost its skills before the read failure even happened: %v", got)
	}

	// Corrupt skills.yaml so the next ListSkills() call fails.
	skillsPath := filepath.Join(repo.Dir(), config.SkillsFilename)
	if err := os.WriteFile(skillsPath, []byte("not: valid: yaml: at: all: :::"), 0o600); err != nil {
		t.Fatalf("corrupt skills.yaml: %v", err)
	}

	msg, ok := coord.RefreshData()().(messages.DataLoadedMsg)
	if !ok {
		t.Fatal("expected refreshData to produce a DataLoadedMsg")
	}
	if msg.Err == nil {
		t.Fatal("expected refreshData to surface the corrupt skills.yaml read error")
	}
	if msg.Skills != nil {
		t.Fatalf("expected Skills to be nil on a read error (not an emptied slice), got: %v", msg.Skills)
	}

	coord.Update(msg)

	got := editor.Card().SkillIDs
	if len(got) != 1 || got[0] != config.BuiltinSkillMCPCLI {
		t.Fatalf("a transient skills.yaml read failure wiped the card editor's skills: %v", got)
	}
}

// TestPartialPreviewBroadcastDoesNotEmptyAgentSwitchCardList is the
// agent-switch analog of the card editor test above: AgentSwitchView.SetData
// must not be reachable from a partial preview update either.
func TestPartialPreviewBroadcastDoesNotEmptyAgentSwitchCardList(t *testing.T) {
	repo, _ := newCurrentRepository(t.TempDir())
	sm := newTestManager(t, repo.Dir(), repo)

	wd := t.TempDir()
	_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	card, errCreate := createCurrentCard(repo, config.IDCard{
		Name:        "Switch Target Card",
		Driver:      "cat",
		WorkspaceID: workspaces[0].ID,
	})
	if errCreate != nil {
		t.Fatalf("create card: %v", errCreate)
	}

	sessionID, errLaunch := launchCurrentSession(repo, sm, card)
	if errLaunch != nil {
		t.Fatalf("launch: %v", errLaunch)
	}
	defer func() { _ = sm.Kill(sessionID) }()

	coord := newOpenedCoordinator(repo, sm)
	coord.Init()
	coord.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	coord.Update(coord.RefreshData()())

	// SessionName is deliberately different from card.Name: the switch
	// view's header always echoes SessionName, so asserting on card.Name
	// (the string under test - it only appears if the card LIST is
	// populated) would otherwise pass trivially via the header regardless
	// of whether SetData actually populated the list.
	coord.Update(
		messages.AgentSwitchRequestedMsg{SessionID: sessionID, SessionName: "Switch Session", DashboardID: "default", Generation: 1},
	)
	if coord.CurrentView() != app.ViewAgentSwitch {
		t.Fatalf("expected ViewAgentSwitch, got %v", coord.CurrentView())
	}

	switchView, ok := coord.Views()[app.ViewAgentSwitch].(*views.AgentSwitchView)
	if !ok {
		t.Fatal("expected *views.AgentSwitchView for ViewAgentSwitch")
	}

	if view := switchView.View(); !strings.Contains(view, card.Name) {
		t.Fatalf("agent-switch view's card list was empty before the broadcast even ran: %q", view)
	}

	coord.Update(messages.SessionPreviewRequestedMsg{SessionID: sessionID})
	coord.Update(messages.SessionOutputMsg{SessionID: sessionID, Output: []byte("hello")})

	if view := switchView.View(); !strings.Contains(view, card.Name) {
		t.Fatalf("agent-switch view's card list was emptied by a partial preview broadcast: %q", view)
	}
}

// A failed cards read leaves the agent-switch list intact until a successful refresh.
func TestTransientCardsReadFailureDoesNotEmptyAgentSwitchCardList(t *testing.T) {
	repo, errRepo := newCurrentRepository(t.TempDir())
	if errRepo != nil {
		t.Fatalf("new repo: %v", errRepo)
	}
	sm := newTestManager(t, repo.Dir(), repo)

	wd := t.TempDir()
	_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	card, errCreate := createCurrentCard(repo, config.IDCard{
		Name:        "Switch Target Card",
		Driver:      "cat",
		WorkspaceID: workspaces[0].ID,
	})
	if errCreate != nil {
		t.Fatalf("create card: %v", errCreate)
	}

	sessionID, errLaunch := launchCurrentSession(repo, sm, card)
	if errLaunch != nil {
		t.Fatalf("launch: %v", errLaunch)
	}
	defer func() { _ = sm.Kill(sessionID) }()

	coord := newOpenedCoordinator(repo, sm)
	coord.Init()
	coord.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	coord.Update(coord.RefreshData()())

	// SessionName is deliberately different from card.Name - see the
	// comment on the sibling preview-trigger test above for why: the
	// header always echoes SessionName, which would make an assertion on
	// card.Name pass trivially regardless of the card list's real content.
	coord.Update(
		messages.AgentSwitchRequestedMsg{SessionID: sessionID, SessionName: "Switch Session", DashboardID: "default", Generation: 1},
	)
	if coord.CurrentView() != app.ViewAgentSwitch {
		t.Fatalf("expected ViewAgentSwitch, got %v", coord.CurrentView())
	}

	switchView, ok := coord.Views()[app.ViewAgentSwitch].(*views.AgentSwitchView)
	if !ok {
		t.Fatal("expected *views.AgentSwitchView for ViewAgentSwitch")
	}

	if view := switchView.View(); !strings.Contains(view, card.Name) {
		t.Fatalf("agent-switch view's card list was empty before the read failure even happened: %q", view)
	}

	// Corrupt id_cards.yaml so the next ListIDCards() call fails.
	idCardsPath := filepath.Join(repo.Dir(), config.IDCardsFilename)
	if err := os.WriteFile(idCardsPath, []byte("not: valid: yaml: at: all: :::"), 0o600); err != nil {
		t.Fatalf("corrupt id_cards.yaml: %v", err)
	}

	msg, ok := coord.RefreshData()().(messages.DataLoadedMsg)
	if !ok {
		t.Fatal("expected refreshData to produce a DataLoadedMsg")
	}
	if msg.Err == nil {
		t.Fatal("expected refreshData to surface the corrupt id_cards.yaml read error")
	}
	if msg.Cards != nil {
		t.Fatalf("expected Cards to be nil on a read error (not an emptied slice), got: %v", msg.Cards)
	}

	coord.Update(msg)

	if view := switchView.View(); !strings.Contains(view, card.Name) {
		t.Fatalf("a transient id_cards.yaml read failure emptied the agent-switch view's card list: %q", view)
	}
}

// Queued output after navigation must not rearm a nil subscription channel.
func TestWaitForOutputNilChannelIsSafe(t *testing.T) {
	repo, _ := newCurrentRepository(t.TempDir())
	sm := newTestManager(t, repo.Dir(), repo)

	wd := t.TempDir()
	_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	card, _ := createCurrentCard(repo, config.IDCard{
		Name:        "Preview Target",
		Driver:      "cat",
		WorkspaceID: workspaces[0].ID,
	})

	sessionID, errLaunch := launchCurrentSession(repo, sm, card)
	if errLaunch != nil {
		t.Fatalf("launch: %v", errLaunch)
	}
	defer func() { _ = sm.Kill(sessionID) }()

	coord := newOpenedCoordinator(repo, sm)
	coord.Init()
	coord.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	coord.Update(coord.RefreshData()())

	coord.Update(messages.SessionPreviewRequestedMsg{SessionID: sessionID})

	// Navigate away from the dashboard: this unsubscribes and clears
	// previewSessionID (the #12 fix).
	coord.NavigateTo(app.ViewCardEditor)

	// A SessionOutputMsg for the session we just left, already in flight,
	// must not re-arm a nil-channel wait.
	_, cmd := coord.Update(messages.SessionOutputMsg{SessionID: sessionID, Output: []byte("stale")})
	if cmd != nil {
		t.Fatal("expected nil cmd for a stale SessionOutputMsg after navigating away")
	}
}
