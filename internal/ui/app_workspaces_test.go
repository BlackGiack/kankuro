package ui //nolint:testpackage // Workspace flow tests need unexported UI helpers.

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"kankuro/internal/config"
)

// workspaceOnlyStub implements workspaceStore for focused workspace tests.
type workspaceOnlyStub struct {
	listWorkspaceSourcesFn  func() ([]config.WorkspaceSource, error)
	saveWorkspaceSourcesFn  func([]config.WorkspaceSource) error
	syncWorkspacesFn        func() error
	listWorkspacesFn        func() ([]config.Workspace, error)
	deleteWorkspaceSourceFn func(string) error
	DirFn                   func() string
}

func (s workspaceOnlyStub) ListWorkspaceSources() ([]config.WorkspaceSource, error) {
	return s.listWorkspaceSourcesFn()
}
func (s workspaceOnlyStub) SaveWorkspaceSources(src []config.WorkspaceSource) error {
	return s.saveWorkspaceSourcesFn(src)
}
func (s workspaceOnlyStub) SyncWorkspacesFromSources() error            { return s.syncWorkspacesFn() }
func (s workspaceOnlyStub) ListWorkspaces() ([]config.Workspace, error) { return s.listWorkspacesFn() }
func (s workspaceOnlyStub) GetWorkspaceByPath(string) (config.Workspace, bool) {
	return config.Workspace{}, false
}
func (s workspaceOnlyStub) DeleteWorkspaceSource(id string) error {
	return s.deleteWorkspaceSourceFn(id)
}

func (s workspaceOnlyStub) Dir() string {
	if s.DirFn != nil {
		return s.DirFn()
	}
	return ""
}

func (s workspaceOnlyStub) ListIDCards() ([]config.IDCard, error)                  { return nil, nil }
func (s workspaceOnlyStub) CreateIDCard(card config.IDCard) (config.IDCard, error) { return card, nil }
func (s workspaceOnlyStub) UpdateIDCard(card config.IDCard) (config.IDCard, error) { return card, nil }
func (s workspaceOnlyStub) DeleteIDCard(string) error                              { return nil }
func (s workspaceOnlyStub) ListSkillSources() ([]string, error)                    { return nil, nil }
func (s workspaceOnlyStub) SaveSkillSources([]string) error                        { return nil }
func (s workspaceOnlyStub) SyncSkillsFromSources() error                           { return nil }
func (s workspaceOnlyStub) ListSkills() ([]config.Skill, error)                    { return nil, nil }
func (s workspaceOnlyStub) UpsertSkill(skill config.Skill) (config.Skill, error)   { return skill, nil }
func (s workspaceOnlyStub) DeleteSkill(string) error                               { return nil }
func (s workspaceOnlyStub) ListPersonaSources() ([]string, error)                  { return nil, nil }
func (s workspaceOnlyStub) SavePersonaSources([]string) error                      { return nil }
func (s workspaceOnlyStub) SyncPersonasFromSources() error                         { return nil }
func (s workspaceOnlyStub) ListPersonas() ([]config.Persona, error)                { return nil, nil }
func (s workspaceOnlyStub) UpsertPersona(persona config.Persona) (config.Persona, error) {
	return persona, nil
}
func (s workspaceOnlyStub) DeletePersona(string) error                               { return nil }
func (s workspaceOnlyStub) ListMCPs() ([]config.MCPServer, error)                    { return nil, nil }
func (s workspaceOnlyStub) UpsertMCP(mcp config.MCPServer) (config.MCPServer, error) { return mcp, nil }
func (s workspaceOnlyStub) DeleteMCP(string) error                                   { return nil }
func (s workspaceOnlyStub) GetIDEConfig() (config.IDEConfig, error)                  { return config.IDEConfig{}, nil }
func (s workspaceOnlyStub) SaveIDEConfig(config.IDEConfig) error                     { return nil }

func TestUpdateWorkspaceSourceFormSaveErrorStaysInForm(t *testing.T) {
	expectedErr := errors.New("save failed")
	app := &App{
		repo: workspaceOnlyStub{
			listWorkspaceSourcesFn: func() ([]config.WorkspaceSource, error) {
				return nil, nil
			},
			saveWorkspaceSourcesFn: func([]config.WorkspaceSource) error {
				return expectedErr
			},
			listWorkspacesFn: func() ([]config.Workspace, error) {
				return nil, nil
			},
			syncWorkspacesFn:        func() error { return nil },
			deleteWorkspaceSourceFn: func(string) error { return nil },
		},
		mode:          viewWorkspaceSourceForm,
		cardForm:      NewCardFormState(),
		workspaceForm: NewWorkspaceFormState(),
	}
	app.initInputs()
	app.workspaceForm.Inputs[0].SetValue("/tmp/demo")
	app.workspaceForm.Inputs[1].SetValue("y")

	model, _ := app.updateWorkspaceSourceForm(tea.KeyMsg{Type: tea.KeyEnter})
	got := model.(*App)

	if !errors.Is(got.err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, got.err)
	}
	if got.mode != viewWorkspaceSourceForm {
		t.Fatalf("expected to stay in workspace source form, got mode %v", got.mode)
	}
}

func TestUpdateWorkspaceSourceFormSuccessfulSaveClearsStaleErrorAndReturnsToList(t *testing.T) {
	saved := false
	app := &App{
		repo: workspaceOnlyStub{
			listWorkspaceSourcesFn: func() ([]config.WorkspaceSource, error) {
				return nil, nil
			},
			saveWorkspaceSourcesFn: func([]config.WorkspaceSource) error {
				saved = true
				return nil
			},
			listWorkspacesFn: func() ([]config.Workspace, error) {
				return nil, nil
			},
			syncWorkspacesFn:        func() error { return nil },
			deleteWorkspaceSourceFn: func(string) error { return nil },
		},
		mode:          viewWorkspaceSourceForm,
		err:           errors.New("old error"),
		cardForm:      NewCardFormState(),
		workspaceForm: NewWorkspaceFormState(),
	}
	app.initInputs()
	app.workspaceForm.Inputs[0].SetValue("/tmp/demo")
	app.workspaceForm.Inputs[1].SetValue("y")

	model, _ := app.updateWorkspaceSourceForm(tea.KeyMsg{Type: tea.KeyEnter})
	got := model.(*App)

	if !saved {
		t.Fatalf("expected workspace source to be saved")
	}
	if got.err != nil {
		t.Fatalf("expected stale error to be cleared, got %v", got.err)
	}
	if got.mode != viewWorkspaces {
		t.Fatalf("expected to return to workspace list, got mode %v", got.mode)
	}
}

func TestDeleteSelectedWorkspaceSourceSetsError(t *testing.T) {
	expectedErr := errors.New("delete failed")
	app := &App{
		repo: workspaceOnlyStub{
			listWorkspaceSourcesFn:  func() ([]config.WorkspaceSource, error) { return nil, nil },
			saveWorkspaceSourcesFn:  func([]config.WorkspaceSource) error { return nil },
			listWorkspacesFn:        func() ([]config.Workspace, error) { return nil, nil },
			syncWorkspacesFn:        func() error { return nil },
			deleteWorkspaceSourceFn: func(string) error { return expectedErr },
		},
		workspaceForm: NewWorkspaceFormState(),
	}
	app.workspaceForm.Sources = []config.WorkspaceSource{{ID: "source-1", Path: "/tmp/demo"}}

	app.deleteSelectedWorkspaceSource()

	if !errors.Is(app.err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, app.err)
	}
	if len(app.workspaceForm.Sources) != 1 {
		t.Fatalf("expected sources to remain unchanged on delete error")
	}
}

func TestUpdateWorkspacesRefreshSetsError(t *testing.T) {
	expectedErr := errors.New("refresh failed")
	app := &App{
		repo: workspaceOnlyStub{
			listWorkspaceSourcesFn:  func() ([]config.WorkspaceSource, error) { return nil, nil },
			saveWorkspaceSourcesFn:  func([]config.WorkspaceSource) error { return nil },
			listWorkspacesFn:        func() ([]config.Workspace, error) { return nil, nil },
			syncWorkspacesFn:        func() error { return expectedErr },
			deleteWorkspaceSourceFn: func(string) error { return nil },
		},
		mode:          viewWorkspaces,
		workspaceForm: NewWorkspaceFormState(),
	}

	model, _ := app.updateWorkspaces(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	got := model.(*App)

	if !errors.Is(got.err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, got.err)
	}
	if got.mode != viewWorkspaces {
		t.Fatalf("expected to stay in workspace view, got mode %v", got.mode)
	}
}

func TestFilteredWorkspacesMatchesNameAndPath(t *testing.T) {
	app := &App{
		workspaces: []config.Workspace{
			{ID: "1", Name: "Project Alpha", Path: "/home/user/code/alpha"},
			{ID: "2", Name: "Beta Workspace", Path: "/home/user/beta"},
			{ID: "3", Name: "Gamma", Path: "/home/user/code/gamma"},
		},
	}
	app.cardForm.WorkspaceSearchQuery = "alpha"
	got := app.filteredWorkspaces()
	if len(got) != 1 || got[0].ID != "1" {
		t.Fatalf("expected workspace ID 1 by name, got %v", got)
	}

	app.cardForm.WorkspaceSearchQuery = "/home/user/code"
	got = app.filteredWorkspaces()
	if len(got) != 2 || got[0].ID != "1" || got[1].ID != "3" {
		t.Fatalf("expected workspaces 1 and 3 by path, got %v", got)
	}

	app.cardForm.WorkspaceSearchQuery = ""
	got = app.filteredWorkspaces()
	if len(got) != 3 {
		t.Fatalf("expected all 3 workspaces with empty query, got %v", got)
	}
}

func TestWorkspaceWindowTopMiddleBottom(t *testing.T) {
	app := &App{
		height: 38,
	}
	app.cardForm.WorkspaceCursor = 0
	app.workspaces = make([]config.Workspace, 20)

	start, end := app.workspaceWindow(20)
	if start != 0 || end != 6 {
		t.Fatalf("top: expected (0, 6), got (%d, %d)", start, end)
	}

	app.cardForm.WorkspaceCursor = 10
	start, end = app.workspaceWindow(20)
	if start != 7 || end != 13 {
		t.Fatalf("middle: expected (7, 13), got (%d, %d)", start, end)
	}

	app.cardForm.WorkspaceCursor = 18
	start, end = app.workspaceWindow(20)
	if start != 14 || end != 20 {
		t.Fatalf("bottom: expected (14, 20), got (%d, %d)", start, end)
	}
}
