package app_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"kankuro/internal/app"
	"kankuro/internal/config"
	"kankuro/internal/session"
	"kankuro/internal/ui/messages"
	"kankuro/internal/ui/views"
)

// TestPartialPreviewBroadcastDoesNotWipeCardEditor pins finding #17 from
// tech_debt_founding.md: a preview-only update (previously a DataLoadedMsg
// with only the preview fields set) must never reach CardEditor.SetData and
// wipe its skills/personas/MCPs. Since debt_impl_plan_4.md step 2b, preview
// updates are a distinct message type (PreviewDataMsg) with a distinct
// receiver interface (PreviewReceiver) that CardEditor does not implement -
// so this is a structural guarantee, not a per-view nil-guard.
//
// The test deliberately re-arms the preview subscription for the previewed
// session *after* navigating to the card editor, to exercise the fix
// directly rather than relying on step 2's #12 fix (clearing
// previewSessionID on navigate-away) to make the scenario unreachable. Both
// fixes matter independently - see debt_impl_plan_4.md step 2b, point 4.
func TestPartialPreviewBroadcastDoesNotWipeCardEditor(t *testing.T) {
	repo, _ := config.NewFileRepository(t.TempDir())
	sm := session.NewManager(repo.Dir(), repo)

	wd := t.TempDir()
	_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	card, errCreate := repo.CreateIDCard(config.IDCard{
		Name:        "With Skills",
		Driver:      "cat",
		WorkspaceID: workspaces[0].ID,
		SkillIDs:    []string{config.BuiltinSkillMCPCLI},
	})
	if errCreate != nil {
		t.Fatalf("create card: %v", errCreate)
	}

	sessionID, errLaunch := sm.Launch(card, nil, nil, nil)
	if errLaunch != nil {
		t.Fatalf("launch: %v", errLaunch)
	}
	defer func() { _ = sm.Kill(sessionID) }()

	coord := app.NewCoordinator(repo, sm)
	coord.Init()
	coord.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	coord.Update(coord.RefreshData()())

	editor, ok := coord.Views()[app.ViewCardEditor].(*views.CardEditor)
	if !ok {
		t.Fatal("expected *views.CardEditor for ViewCardEditor")
	}
	editor.SetCard(card)
	coord.Update(messages.NavigateMsg{View: messages.ViewCardEditor})

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

// TestTransientSkillsReadFailureDoesNotWipeCardEditor pins finding 1 from
// review_dp_4.md: the other trigger for the same #17 data-loss class.
// refreshData already nils DataLoadedMsg.Skills on a read error rather than
// broadcasting an empty slice (debt_impl_plan_4.md step 5), but
// CardEditor.SetData used to assign every field unconditionally, so a
// transient skills.yaml read failure still wiped the editor's skill
// selection - and a save immediately after would have persisted the card
// with none. CardEditor.SetData now guards each field with `if != nil`, the
// same contract Dashboard.SetData already honored.
func TestTransientSkillsReadFailureDoesNotWipeCardEditor(t *testing.T) {
	repo, errRepo := config.NewFileRepository(t.TempDir())
	if errRepo != nil {
		t.Fatalf("new repo: %v", errRepo)
	}
	sm := session.NewManager(repo.Dir(), repo)

	wd := t.TempDir()
	_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	card, errCreate := repo.CreateIDCard(config.IDCard{
		Name:        "With Skills",
		Driver:      "cat",
		WorkspaceID: workspaces[0].ID,
		SkillIDs:    []string{config.BuiltinSkillMCPCLI},
	})
	if errCreate != nil {
		t.Fatalf("create card: %v", errCreate)
	}

	coord := app.NewCoordinator(repo, sm)
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
	repo, _ := config.NewFileRepository(t.TempDir())
	sm := session.NewManager(repo.Dir(), repo)

	wd := t.TempDir()
	_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	card, errCreate := repo.CreateIDCard(config.IDCard{
		Name:        "Switch Target Card",
		Driver:      "cat",
		WorkspaceID: workspaces[0].ID,
	})
	if errCreate != nil {
		t.Fatalf("create card: %v", errCreate)
	}

	sessionID, errLaunch := sm.Launch(card, nil, nil, nil)
	if errLaunch != nil {
		t.Fatalf("launch: %v", errLaunch)
	}
	defer func() { _ = sm.Kill(sessionID) }()

	coord := app.NewCoordinator(repo, sm)
	coord.Init()
	coord.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	coord.Update(coord.RefreshData()())

	// SessionName is deliberately different from card.Name: the switch
	// view's header always echoes SessionName, so asserting on card.Name
	// (the string under test - it only appears if the card LIST is
	// populated) would otherwise pass trivially via the header regardless
	// of whether SetData actually populated the list.
	coord.Update(messages.AgentSwitchRequestedMsg{SessionID: sessionID, SessionName: "Switch Session"})
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

// TestTransientCardsReadFailureDoesNotEmptyAgentSwitchCardList pins finding 3
// from review_dp_4_2.md: AgentSwitchView.SetData's `if msg.Cards != nil` guard
// (the read-failure half of the same #17 class as
// TestTransientSkillsReadFailureDoesNotWipeCardEditor above) had no test of
// its own - reverting it to an unconditional assignment left the whole suite
// green. refreshData already nils DataLoadedMsg.Cards on a read error rather
// than broadcasting an empty slice (debt_impl_plan_4.md step 5); this pins
// that the agent-switch view keeps its card list across that failure instead
// of emptying it until the next successful tick.
func TestTransientCardsReadFailureDoesNotEmptyAgentSwitchCardList(t *testing.T) {
	repo, errRepo := config.NewFileRepository(t.TempDir())
	if errRepo != nil {
		t.Fatalf("new repo: %v", errRepo)
	}
	sm := session.NewManager(repo.Dir(), repo)

	wd := t.TempDir()
	_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	card, errCreate := repo.CreateIDCard(config.IDCard{
		Name:        "Switch Target Card",
		Driver:      "cat",
		WorkspaceID: workspaces[0].ID,
	})
	if errCreate != nil {
		t.Fatalf("create card: %v", errCreate)
	}

	sessionID, errLaunch := sm.Launch(card, nil, nil, nil)
	if errLaunch != nil {
		t.Fatalf("launch: %v", errLaunch)
	}
	defer func() { _ = sm.Kill(sessionID) }()

	coord := app.NewCoordinator(repo, sm)
	coord.Init()
	coord.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	coord.Update(coord.RefreshData()())

	// SessionName is deliberately different from card.Name - see the
	// comment on the sibling preview-trigger test above for why: the
	// header always echoes SessionName, which would make an assertion on
	// card.Name pass trivially regardless of the card list's real content.
	coord.Update(messages.AgentSwitchRequestedMsg{SessionID: sessionID, SessionName: "Switch Session"})
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

// TestWaitForOutputNilChannelIsSafe pins finding #12 from
// tech_debt_founding.md: navigating away clears previewSubChan and
// previewSessionID together, so a SessionOutputMsg already queued for the
// session left behind cannot re-arm waitForOutput on a nil channel (which
// would block forever on <-ch). Both belt-and-braces guards from
// debt_impl_plan_4.md step 2 are exercised here: the nil check in the
// producer (the SessionOutputMsg branch) and clearing previewSessionID in
// navigateTo.
func TestWaitForOutputNilChannelIsSafe(t *testing.T) {
	repo, _ := config.NewFileRepository(t.TempDir())
	sm := session.NewManager(repo.Dir(), repo)

	wd := t.TempDir()
	_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	card, _ := repo.CreateIDCard(config.IDCard{
		Name:        "Preview Target",
		Driver:      "cat",
		WorkspaceID: workspaces[0].ID,
	})

	sessionID, errLaunch := sm.Launch(card, nil, nil, nil)
	if errLaunch != nil {
		t.Fatalf("launch: %v", errLaunch)
	}
	defer func() { _ = sm.Kill(sessionID) }()

	coord := app.NewCoordinator(repo, sm)
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
