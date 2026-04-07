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
		mode: viewWorkspaceSourceForm,
	}
	app.initInputs()
	app.workspaceInputs[0].SetValue("/tmp/demo")
	app.workspaceInputs[1].SetValue("y")

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
		mode: viewWorkspaceSourceForm,
		err:  errors.New("old error"),
	}
	app.initInputs()
	app.workspaceInputs[0].SetValue("/tmp/demo")
	app.workspaceInputs[1].SetValue("y")

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
		workspaceSources: []config.WorkspaceSource{{ID: "source-1", Path: "/tmp/demo"}},
	}

	app.deleteSelectedWorkspaceSource()

	if !errors.Is(app.err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, app.err)
	}
	if len(app.workspaceSources) != 1 {
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
		mode: viewWorkspaces,
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
