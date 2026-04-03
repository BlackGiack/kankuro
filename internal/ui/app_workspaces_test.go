package ui //nolint:testpackage // Workspace flow tests need unexported UI helpers.

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"kankuro/internal/config"
)

type workspaceRepoStub struct {
	listWorkspaceSourcesFn  func() ([]config.WorkspaceSource, error)
	saveWorkspaceSourcesFn  func([]config.WorkspaceSource) error
	syncWorkspacesFn        func() error
	listWorkspacesFn        func() ([]config.Workspace, error)
	deleteWorkspaceSourceFn func(string) error
}

func (s workspaceRepoStub) ListIDCards() ([]config.IDCard, error)                  { return nil, nil }
func (s workspaceRepoStub) CreateIDCard(card config.IDCard) (config.IDCard, error) { return card, nil }
func (s workspaceRepoStub) UpdateIDCard(card config.IDCard) (config.IDCard, error) { return card, nil }
func (s workspaceRepoStub) DeleteIDCard(string) error                              { return nil }
func (s workspaceRepoStub) ListWorkspaceSources() ([]config.WorkspaceSource, error) {
	return s.listWorkspaceSourcesFn()
}
func (s workspaceRepoStub) SaveWorkspaceSources(src []config.WorkspaceSource) error {
	return s.saveWorkspaceSourcesFn(src)
}
func (s workspaceRepoStub) SyncWorkspacesFromSources() error            { return s.syncWorkspacesFn() }
func (s workspaceRepoStub) ListWorkspaces() ([]config.Workspace, error) { return s.listWorkspacesFn() }
func (s workspaceRepoStub) GetWorkspaceByPath(string) (config.Workspace, bool) {
	return config.Workspace{}, false
}
func (s workspaceRepoStub) DeleteWorkspaceSource(id string) error {
	return s.deleteWorkspaceSourceFn(id)
}
func (s workspaceRepoStub) ListSkillSources() ([]string, error)                  { return nil, nil }
func (s workspaceRepoStub) SaveSkillSources([]string) error                      { return nil }
func (s workspaceRepoStub) SyncSkillsFromSources() error                         { return nil }
func (s workspaceRepoStub) ListSkills() ([]config.Skill, error)                  { return nil, nil }
func (s workspaceRepoStub) UpsertSkill(skill config.Skill) (config.Skill, error) { return skill, nil }
func (s workspaceRepoStub) DeleteSkill(string) error                             { return nil }
func (s workspaceRepoStub) ListPersonaSources() ([]string, error)                { return nil, nil }
func (s workspaceRepoStub) SavePersonaSources([]string) error                    { return nil }
func (s workspaceRepoStub) SyncPersonasFromSources() error                       { return nil }
func (s workspaceRepoStub) ListPersonas() ([]config.Persona, error)              { return nil, nil }
func (s workspaceRepoStub) UpsertPersona(persona config.Persona) (config.Persona, error) {
	return persona, nil
}
func (s workspaceRepoStub) DeletePersona(string) error                               { return nil }
func (s workspaceRepoStub) ListMCPs() ([]config.MCPServer, error)                    { return nil, nil }
func (s workspaceRepoStub) UpsertMCP(mcp config.MCPServer) (config.MCPServer, error) { return mcp, nil }
func (s workspaceRepoStub) DeleteMCP(string) error                                   { return nil }
func (s workspaceRepoStub) Dir() string                                              { return "" }

func TestUpdateWorkspaceSourceFormSaveErrorStaysInForm(t *testing.T) {
	expectedErr := errors.New("save failed")
	app := &App{
		repo: workspaceRepoStub{
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

func TestDeleteSelectedWorkspaceSourceSetsError(t *testing.T) {
	expectedErr := errors.New("delete failed")
	app := &App{
		repo: workspaceRepoStub{
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
		repo: workspaceRepoStub{
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
