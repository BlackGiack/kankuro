package views

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"kankuro/internal/config"
	"kankuro/internal/ui/commands"
	"kankuro/internal/ui/components"
	"kankuro/internal/ui/messages"
)

func TestSourceManager_New_Skills(t *testing.T) {
	callbacks := SourceManagerCallbacks{}
	sm := NewSourceManager(SourceTypeSkills, callbacks, &commands.Commands{})

	if sm == nil {
		t.Fatal("NewSourceManager returned nil")
	}

	if sm.sourceType != SourceTypeSkills {
		t.Errorf("Expected SourceTypeSkills, got %s", sm.sourceType)
	}

	if sm.typeConfig.Title != "Manage Skill Sources" {
		t.Errorf("Wrong title: %s", sm.typeConfig.Title)
	}

	if sm.typeConfig.FormFields != 1 {
		t.Errorf("Skills should have 1 form field, got %d", sm.typeConfig.FormFields)
	}
}

func TestSourceManager_New_Personas(t *testing.T) {
	callbacks := SourceManagerCallbacks{}
	sm := NewSourceManager(SourceTypePersonas, callbacks, &commands.Commands{})

	if sm.sourceType != SourceTypePersonas {
		t.Errorf("Expected SourceTypePersonas, got %s", sm.sourceType)
	}

	if sm.typeConfig.FormFields != 1 {
		t.Errorf("Personas should have 1 form field, got %d", sm.typeConfig.FormFields)
	}
}

func TestSourceManager_New_Workspaces(t *testing.T) {
	callbacks := SourceManagerCallbacks{}
	sm := NewSourceManager(SourceTypeWorkspaces, callbacks, &commands.Commands{})

	if sm.sourceType != SourceTypeWorkspaces {
		t.Errorf("Expected SourceTypeWorkspaces, got %s", sm.sourceType)
	}

	if sm.typeConfig.FormFields != 2 {
		t.Errorf("Workspaces should have 2 form fields, got %d", sm.typeConfig.FormFields)
	}

	if sm.typeConfig.Legend == "" {
		t.Error("Workspaces should have a legend")
	}
}

func TestSourceManager_SetData(t *testing.T) {
	loaded := false

	callbacks := SourceManagerCallbacks{
		LoadSources: func() []components.SourceItem {
			loaded = true
			return []components.SourceItem{
				components.SimpleSourceItem("/path/to/source"),
			}
		},
	}

	sm := NewSourceManager(SourceTypeSkills, callbacks, &commands.Commands{})
	sm.SetData()

	if !loaded {
		t.Error("LoadSources should have been called")
	}
}

func TestSourceManager_SetPrevView(t *testing.T) {
	sm := NewSourceManager(SourceTypeSkills, SourceManagerCallbacks{}, &commands.Commands{})

	sm.SetPrevView(messages.ViewCardEditor)

	if sm.prevView != messages.ViewCardEditor {
		t.Errorf("Expected ViewCardEditor, got %d", sm.prevView)
	}
}

func TestSourceManager_View(t *testing.T) {
	callbacks := SourceManagerCallbacks{
		LoadSources: func() []components.SourceItem {
			return []components.SourceItem{
				components.SimpleSourceItem("/path1"),
				components.SimpleSourceItem("/path2"),
			}
		},
	}

	sm := NewSourceManager(SourceTypeSkills, callbacks, &commands.Commands{})
	sm.Init()

	view := sm.View()

	if view == "" {
		t.Error("View should not be empty")
	}

	if !containsString(view, "Manage Skill Sources") {
		t.Error("View should contain title")
	}

	if !containsString(view, "/path1") {
		t.Error("View should contain sources")
	}
}

func TestSourceManager_Save_Simple(t *testing.T) {
	var savedPath string
	var savedIndex int

	callbacks := SourceManagerCallbacks{
		SaveSimpleSource: func(path string, idx int) error {
			savedPath = path
			savedIndex = idx
			return nil
		},
		Refresh: func() {},
	}

	item := components.SimpleSourceItem("/new/source")

	err := handleSave(SourceTypeSkills, callbacks, item, -1)

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

	callbacks := SourceManagerCallbacks{
		SaveWorkspaceSource: func(source config.WorkspaceSource, _ int) error {
			savedSource = source
			return nil
		},
		Refresh: func() {},
	}

	item := components.WorkspaceSourceItem{
		Source: config.WorkspaceSource{
			ID:     "test-id",
			Path:   "/workspace/path",
			IsRoot: true,
		},
	}

	err := handleSave(SourceTypeWorkspaces, callbacks, item, 0)

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
	callbacks := SourceManagerCallbacks{
		DeleteSource: func(_ int) error {
			return nil
		},
		Refresh: func() {},
	}

	sm := NewSourceManager(SourceTypeSkills, callbacks, &commands.Commands{})

	sm.SetData()

	sm.sourceList.SetItems([]components.SourceItem{
		components.SimpleSourceItem("/path1"),
		components.SimpleSourceItem("/path2"),
	})

	_, _ = sm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
}

func TestSourceManager_Escape_NavigatesBack(t *testing.T) {
	sm := NewSourceManager(SourceTypeSkills, SourceManagerCallbacks{}, &commands.Commands{})
	sm.SetPrevView(messages.ViewCardEditor)
	sm.Init()

	_, cmd := sm.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if cmd == nil {
		t.Error("Should return a command on escape")
	}

	msg := cmd()
	if navMsg, ok := msg.(messages.NavigateMsg); ok {
		if navMsg.View != messages.ViewCardEditor {
			t.Errorf("Should navigate to ViewCardEditor, got %d", navMsg.View)
		}
	} else {
		t.Error("Should emit NavigateMsg")
	}
}

func TestSourceManager_Escape_InFormMode(t *testing.T) {
	sm := NewSourceManager(SourceTypeSkills, SourceManagerCallbacks{}, &commands.Commands{})
	sm.Init()

	sm.sourceList.OpenForm(-1)

	model, _ := sm.Update(tea.KeyMsg{Type: tea.KeyEsc})
	sm = model.(*SourceManager)

	if sm.sourceList.Mode != components.SourceListModeList {
		t.Error("Should return to list mode on escape in form")
	}
}

func TestSourceManager_SetSize(t *testing.T) {
	sm := NewSourceManager(SourceTypeSkills, SourceManagerCallbacks{}, &commands.Commands{})
	sm.SetSize(80, 24)

	if sm.Width != 80 {
		t.Errorf("Expected width 80, got %d", sm.Width)
	}

	if sm.Height != 24 {
		t.Errorf("Expected height 24, got %d", sm.Height)
	}
}

func TestSourceManager_Error(t *testing.T) {
	callbacks := SourceManagerCallbacks{
		SaveSimpleSource: func(_ string, _ int) error {
			return errors.New("save failed")
		},
	}

	sm := NewSourceManager(SourceTypeSkills, callbacks, &commands.Commands{})
	sm.err = errors.New("test error")

	view := sm.View()

	if !containsString(view, "Error:") {
		t.Error("Should display error in view")
	}
}

func TestGetSourceTypeConfig(t *testing.T) {
	tests := []struct {
		sourceType     SourceType
		expectedTitle  string
		expectedFields int
	}{
		{SourceTypeSkills, "Manage Skill Sources", 1},
		{SourceTypePersonas, "Manage Persona Sources", 1},
		{SourceTypeWorkspaces, "Manage Workspace Sources", 2},
	}

	for _, tt := range tests {
		t.Run(string(tt.sourceType), func(t *testing.T) {
			config := GetSourceTypeConfig(tt.sourceType)

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
	id1 := GenerateID()
	id2 := GenerateID()

	if id1 == id2 {
		t.Error("IDs should be unique")
	}

	if len(id1) < 5 {
		t.Errorf("ID too short: %s", id1)
	}
}
