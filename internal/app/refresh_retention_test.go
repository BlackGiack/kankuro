package app_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/app"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

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
	editor.LoadEditor(
		contracts.CardEditorData{
			CardID:  card.ID,
			CanEdit: true,
			Values:  contracts.CardValues{Name: card.Name, Driver: card.Driver, WorkspaceID: card.WorkspaceID, SkillIDs: card.SkillIDs},
		},
	)

	if got := editor.Draft().Values.SkillIDs; len(got) != 1 || got[0] != config.BuiltinSkillMCPCLI {
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
	if !msg.EditorChoices.Skills.Attempted || msg.EditorChoices.Skills.Err == nil {
		t.Fatal("expected failed skills read")
	}

	coord.Update(msg)

	got := editor.Draft().Values.SkillIDs
	if len(got) != 1 || got[0] != config.BuiltinSkillMCPCLI {
		t.Fatalf("a transient skills.yaml read failure wiped the card editor's skills: %v", got)
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
	if !msg.Inventory.Attempted || msg.Inventory.Err == nil {
		t.Fatalf("expected an explicit failed inventory result, got: %+v", msg.Inventory)
	}

	coord.Update(msg)

	if view := switchView.View(); !strings.Contains(view, card.Name) {
		t.Fatalf("a transient id_cards.yaml read failure emptied the agent-switch view's card list: %q", view)
	}
}
