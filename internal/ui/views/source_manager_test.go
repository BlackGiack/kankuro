package views_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/ui/components"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

func TestSourceManager_New_Skills(t *testing.T) {
	callbacks := views.SourceManagerCallbacks{}
	sm := views.NewSourceManager(views.SourceTypeSkills, callbacks)

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
	sm := views.NewSourceManager(views.SourceTypePersonas, callbacks)

	if sm.SourceType() != views.SourceTypePersonas {
		t.Errorf("Expected SourceTypePersonas, got %s", sm.SourceType())
	}

	if sm.TypeConfig().FormFields != 1 {
		t.Errorf("Personas should have 1 form field, got %d", sm.TypeConfig().FormFields)
	}
}

func TestSourceManager_New_Workspaces(t *testing.T) {
	callbacks := views.SourceManagerCallbacks{}
	sm := views.NewSourceManager(views.SourceTypeWorkspaces, callbacks)

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

func TestSourceManager_Reload(t *testing.T) {
	loaded := false
	sm := views.NewSourceManager(
		views.SourceTypeSkills,
		views.SourceManagerCallbacks{Load: func(contracts.SourceKind) ([]contracts.SourceRow, error) {
			loaded = true
			return []contracts.SourceRow{{Path: "/path/to/source", Identity: "id"}}, nil
		}},
	)
	cmd := sm.Init()
	if loaded || cmd == nil {
		t.Fatal("load must run asynchronously")
	}
	sm.Update(cmd())
	if !loaded || len(sm.SourceList().Items) != 1 {
		t.Fatal("load result not applied")
	}
}

func TestSourceManager_Escape_NavigatesBack(t *testing.T) {
	sm := views.NewSourceManager(views.SourceTypeSkills, views.SourceManagerCallbacks{})
	sm.Init()

	_, cmd := sm.Update(tea.KeyPressMsg{Code: tea.KeyEsc})

	if cmd == nil {
		t.Error("Should return a command on escape")
	}

	msg := cmd()
	if _, ok := msg.(messages.NavigateBackMsg); !ok {
		t.Errorf("Should emit NavigateBackMsg, got %T", msg)
	}
}

func TestSourceManager_View(t *testing.T) {
	sm := views.NewSourceManager(
		views.SourceTypeSkills,
		views.SourceManagerCallbacks{Load: func(contracts.SourceKind) ([]contracts.SourceRow, error) {
			return []contracts.SourceRow{{Path: "/path1", Identity: "first"}}, nil
		}},
	)
	sm.Update(sm.Init()())
	if !strings.Contains(sm.View(), "/path1") {
		t.Fatal("missing source")
	}
}

func TestSourceManager_Delete(t *testing.T) {
	var got contracts.SourceMutation
	sm := views.NewSourceManager(views.SourceTypeSkills, views.SourceManagerCallbacks{
		Load: func(contracts.SourceKind) ([]contracts.SourceRow, error) {
			return []contracts.SourceRow{{Path: "/one", Identity: "first"}, {Path: "/two", Identity: "second"}}, nil
		},
		Mutate: func(_ context.Context, input contracts.SourceMutation) contracts.SourceMutationResult {
			got = input
			return contracts.SourceMutationResult{Err: errTest}
		},
	})
	sm.Update(sm.Init()())
	sm.SourceList().Cursor = 1
	_, cmd := sm.Update(tea.KeyPressMsg{Text: "d"})
	if cmd == nil {
		t.Fatal("missing command")
	}
	_, reply := sm.Update(cmd())
	if !got.Delete || got.Identity != "second" {
		t.Fatalf("mutation: %+v", got)
	}
	if reply == nil {
		t.Fatal("missing error")
	}
	result, ok := reply().(messages.ErrorMsg)
	if !ok || !errors.Is(result.Err, errTest) {
		t.Fatalf("error: %#v", result)
	}
	if len(sm.SourceList().Items) != 2 {
		t.Fatal("failed deletion lost items")
	}
}

func TestSourceManager_Escape_InFormMode(t *testing.T) {
	sm := views.NewSourceManager(views.SourceTypeSkills, views.SourceManagerCallbacks{})
	sm.Init()

	sm.SourceList().OpenForm(-1)

	model, _ := sm.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	sm = model.(*views.SourceManager)

	if sm.SourceList().Mode != components.SourceListModeList {
		t.Error("Should return to list mode on escape in form")
	}
}

func TestSourceManager_SetSize(t *testing.T) {
	sm := views.NewSourceManager(views.SourceTypeSkills, views.SourceManagerCallbacks{})
	sm.SetSize(80, 24)

	if sm.Width != 80 {
		t.Errorf("Expected width 80, got %d", sm.Width)
	}

	if sm.Height != 24 {
		t.Errorf("Expected height 24, got %d", sm.Height)
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
		{views.SourceTypePuppets, "Manage Puppet Sources", 1},
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

func TestSourceManagerRetainsDraftOnStaleEdit(t *testing.T) {
	var got contracts.SourceMutation
	manager := views.NewSourceManager(views.SourceTypeSkills, views.SourceManagerCallbacks{
		Load: func(contracts.SourceKind) ([]contracts.SourceRow, error) {
			return []contracts.SourceRow{{Path: "/one", Identity: "original"}}, nil
		},
		Mutate: func(_ context.Context, input contracts.SourceMutation) contracts.SourceMutationResult {
			got = input
			return contracts.SourceMutationResult{Err: errTest, Reload: true}
		},
	})
	manager.Update(manager.Init()())
	manager.SourceList().OpenForm(0)
	manager.SourceList().Inputs[0].SetValue("/draft")
	manager.SourceList().SetItems([]components.SourceItem{components.PathSourceItem{Path: "/replacement", Identity: "replacement"}})
	_, command := manager.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if command == nil {
		t.Fatal("missing save")
	}
	_, reply := manager.Update(command())
	if got.Identity != "original" || got.Path != "/draft" {
		t.Fatalf("mutation retargeted: %+v", got)
	}
	if manager.SourceList().Mode != components.SourceListModeForm || manager.SourceList().Inputs[0].Value() != "/draft" || reply == nil {
		t.Fatal("conflict lost draft or feedback")
	}
}

func TestSourceManagerIgnoresOlderLoadAndRetainsFailedData(t *testing.T) {
	call := 0
	manager := views.NewSourceManager(
		views.SourceTypeSkills,
		views.SourceManagerCallbacks{Load: func(contracts.SourceKind) ([]contracts.SourceRow, error) {
			call++
			if call == 3 {
				return nil, errTest
			}
			return []contracts.SourceRow{{Path: fmt.Sprintf("/%d", call), Identity: "id"}}, nil
		}},
	)
	first := manager.Init()
	second := manager.Init()
	manager.Update(second())
	manager.Update(first())
	if len(manager.SourceList().Items) != 1 || manager.SourceList().Items[0].Display() != "/1" {
		t.Fatal("old visit replaced newer result")
	}
	_, feedback := manager.Update(manager.Reload()())
	if feedback == nil || manager.SourceList().Items[0].Display() != "/1" {
		t.Fatal("failed read lost data")
	}
}
