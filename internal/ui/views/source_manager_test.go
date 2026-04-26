package views_test

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"kankuro/internal/config"
	"kankuro/internal/ui/commands"
	"kankuro/internal/ui/components"
	"kankuro/internal/ui/messages"
	"kankuro/internal/ui/views"
)

func TestSourceManager_New_Skills(t *testing.T) {
	callbacks := views.SourceManagerCallbacks{}
	sm := views.NewSourceManager(views.SourceTypeSkills, callbacks, &commands.Commands{})

	if sm == nil {
		t.Fatal("NewSourceManager returned nil")
	}

	if sm.SourceType() != views.SourceTypeSkills {
		t.Errorf("Expected SourceTypeSkills, got %s", sm.SourceType())
	}

	if sm.TypeConfig().Title != "Manage Skill Sources" {
		t.Errorf("Wrong title: %s", sm.TypeConfig().Title)
	}

	if sm.TypeConfig().FormFields != 1 {
		t.Errorf("Skills should have 1 form field, got %d", sm.TypeConfig().FormFields)
	}
}

func TestSourceManager_New_Personas(t *testing.T) {
	callbacks := views.SourceManagerCallbacks{}
	sm := views.NewSourceManager(views.SourceTypePersonas, callbacks, &commands.Commands{})

	if sm.SourceType() != views.SourceTypePersonas {
		t.Errorf("Expected SourceTypePersonas, got %s", sm.SourceType())
	}

	if sm.TypeConfig().FormFields != 1 {
		t.Errorf("Personas should have 1 form field, got %d", sm.TypeConfig().FormFields)
	}
}

func TestSourceManager_New_Workspaces(t *testing.T) {
	callbacks := views.SourceManagerCallbacks{}
	sm := views.NewSourceManager(views.SourceTypeWorkspaces, callbacks, &commands.Commands{})

	if sm.SourceType() != views.SourceTypeWorkspaces {
		t.Errorf("Expected SourceTypeWorkspaces, got %s", sm.SourceType())
	}

	if sm.TypeConfig().FormFields != 2 {
		t.Errorf("Workspaces should have 2 form fields, got %d", sm.TypeConfig().FormFields)
	}

	if sm.TypeConfig().Legend == "" {
		t.Error("Workspaces should have a legend")
	}
}

func TestSourceManager_SetData(t *testing.T) {
	loaded := false

	callbacks := views.SourceManagerCallbacks{
		LoadSources: func() []components.SourceItem {
			loaded = true
			return []components.SourceItem{
				components.SimpleSourceItem("/path/to/source"),
			}
		},
	}

	sm := views.NewSourceManager(views.SourceTypeSkills, callbacks, &commands.Commands{})
	sm.SetData()

	if !loaded {
		t.Error("LoadSources should have been called")
	}
}

func TestSourceManager_Escape_NavigatesBack(t *testing.T) {
	sm := views.NewSourceManager(views.SourceTypeSkills, views.SourceManagerCallbacks{}, &commands.Commands{})
	sm.Init()

	_, cmd := sm.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if cmd == nil {
		t.Error("Should return a command on escape")
	}

	msg := cmd()
	if _, ok := msg.(messages.NavigateBackMsg); !ok {
		t.Errorf("Should emit NavigateBackMsg, got %T", msg)
	}
}

func TestSourceManager_View(t *testing.T) {
	callbacks := views.SourceManagerCallbacks{
		LoadSources: func() []components.SourceItem {
			return []components.SourceItem{
				components.SimpleSourceItem("/path1"),
				components.SimpleSourceItem("/path2"),
			}
		},
	}

	sm := views.NewSourceManager(views.SourceTypeSkills, callbacks, &commands.Commands{})
	sm.Init()

	view := sm.View()

	if view == "" {
		t.Error("View should not be empty")
	}

	if !strings.Contains(view, "Manage Skill Sources") {
		t.Error("View should contain title")
	}

	if !strings.Contains(view, "/path1") {
		t.Error("View should contain sources")
	}
}

func TestSourceManager_Save_Simple(t *testing.T) {
	var savedPath string
	var savedIndex int

	callbacks := views.SourceManagerCallbacks{
		SaveSimpleSource: func(path string, idx int) (tea.Cmd, error) {
			savedPath = path
			savedIndex = idx
			return func() tea.Msg { return nil }, nil
		},
		Refresh: func() tea.Cmd { return nil },
	}

	item := components.SimpleSourceItem("/new/source")

	_, err := views.HandleSave(views.SourceTypeSkills, callbacks, item, -1)

	if err != nil {
		t.Errorf("Save should not error: %v", err)
	}

	if savedPath != "/new/source" {
		t.Errorf("Expected path '/new/source', got '%s'", savedPath)
	}

	if savedIndex != -1 {
		t.Errorf("Expected index -1 for new item, got %d", savedIndex)
	}
}

func TestSourceManager_Save_Workspace(t *testing.T) {
	var savedSource config.WorkspaceSource

	callbacks := views.SourceManagerCallbacks{
		SaveWorkspaceSource: func(source config.WorkspaceSource, _ int) (tea.Cmd, error) {
			savedSource = source
			return func() tea.Msg { return nil }, nil
		},
		Refresh: func() tea.Cmd { return nil },
	}

	item := components.WorkspaceSourceItem{
		Source: config.WorkspaceSource{
			ID:     "test-id",
			Path:   "/workspace/path",
			IsRoot: true,
		},
	}

	_, err := views.HandleSave(views.SourceTypeWorkspaces, callbacks, item, 0)

	if err != nil {
		t.Errorf("Save should not error: %v", err)
	}

	if savedSource.Path != "/workspace/path" {
		t.Errorf("Expected path '/workspace/path', got '%s'", savedSource.Path)
	}

	if !savedSource.IsRoot {
		t.Error("IsRoot should be true")
	}
}

func TestSourceManager_Delete(_ *testing.T) {
	callbacks := views.SourceManagerCallbacks{
		DeleteSource: func(_ int) (tea.Cmd, error) {
			return func() tea.Msg { return nil }, errTest
		},
		Refresh: func() tea.Cmd { return nil },
	}

	sm := views.NewSourceManager(views.SourceTypeSkills, callbacks, &commands.Commands{})

	sm.SetData()

	sm.SourceList().SetItems([]components.SourceItem{
		components.SimpleSourceItem("/path1"),
		components.SimpleSourceItem("/path2"),
	})

	_, _ = sm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
}

func TestSourceManager_EscapeInListMode_NavigatesBack(t *testing.T) {
	sm := views.NewSourceManager(views.SourceTypeSkills, views.SourceManagerCallbacks{}, &commands.Commands{})
	sm.Init()

	_, cmd := sm.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if cmd == nil {
		t.Error("Should return a command on escape")
	}

	msg := cmd()
	if _, ok := msg.(messages.NavigateBackMsg); !ok {
		t.Errorf("Should emit NavigateBackMsg, got %T", msg)
	}
}

func TestSourceManager_Escape_InFormMode(t *testing.T) {
	sm := views.NewSourceManager(views.SourceTypeSkills, views.SourceManagerCallbacks{}, &commands.Commands{})
	sm.Init()

	sm.SourceList().OpenForm(-1)

	model, _ := sm.Update(tea.KeyMsg{Type: tea.KeyEsc})
	sm = model.(*views.SourceManager)

	if sm.SourceList().Mode != components.SourceListModeList {
		t.Error("Should return to list mode on escape in form")
	}
}

func TestSourceManager_SetSize(t *testing.T) {
	sm := views.NewSourceManager(views.SourceTypeSkills, views.SourceManagerCallbacks{}, &commands.Commands{})
	sm.SetSize(80, 24)

	if sm.Width != 80 {
		t.Errorf("Expected width 80, got %d", sm.Width)
	}

	if sm.Height != 24 {
		t.Errorf("Expected height 24, got %d", sm.Height)
	}
}

func TestSourceManager_Error(t *testing.T) {
	callbacks := views.SourceManagerCallbacks{
		SaveSimpleSource: func(_ string, _ int) (tea.Cmd, error) {
			return nil, errors.New("save failed")
		},
	}

	sm := views.NewSourceManager(views.SourceTypeSkills, callbacks, &commands.Commands{})
	sm.SetError(errors.New("test error"))

	view := sm.View()

	if !strings.Contains(view, "Error:") {
		t.Error("Should display error in view")
	}
}

func TestGetSourceTypeConfig(t *testing.T) {
	tests := []struct {
		sourceType     views.SourceType
		expectedTitle  string
		expectedFields int
	}{
		{views.SourceTypeSkills, "Manage Skill Sources", 1},
		{views.SourceTypePersonas, "Manage Persona Sources", 1},
		{views.SourceTypeWorkspaces, "Manage Workspace Sources", 2},
	}

	for _, tt := range tests {
		t.Run(string(tt.sourceType), func(t *testing.T) {
			config := views.GetSourceTypeConfig(tt.sourceType)

			if config.Title != tt.expectedTitle {
				t.Errorf("Expected title '%s', got '%s'", tt.expectedTitle, config.Title)
			}

			if config.FormFields != tt.expectedFields {
				t.Errorf("Expected %d fields, got %d", tt.expectedFields, config.FormFields)
			}
		})
	}
}

func TestGenerateID(t *testing.T) {
	id1 := views.GenerateID()
	id2 := views.GenerateID()

	if id1 == id2 {
		t.Error("IDs should be unique")
	}

	if len(id1) < 5 {
		t.Errorf("ID too short: %s", id1)
	}
}
