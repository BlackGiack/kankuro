package components_test

import (
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"

	"kankuro/internal/config"
	"kankuro/internal/ui/components"
)

var errTest = errors.New("test")

func TestSourceListIntegration_Simple(t *testing.T) {
	// Simulate skill/persona source workflow

	var savedSources []string

	conf := components.SourceListConfig{
		Title:      "Skill Sources",
		FormPrompt: "Path:",
		FormFields: 1,
	}

	callbacks := components.SourceListCallbacks{
		OnSave: func(item components.SourceItem, idx int) (tea.Cmd, error) {
			if idx == -1 {
				savedSources = append(savedSources, item.Display())
			} else {
				savedSources[idx] = item.Display()
			}
			return nil, errTest
		},
		OnRefresh: func() tea.Cmd { return nil },
		OnClose:   func() {},
	}

	sl := components.NewSourceList(conf, callbacks)

	// 1. Initial empty state
	sl.SetItems([]components.SourceItem{})

	view := sl.View()
	if view == "" {
		t.Error("View should render empty state")
	}

	// 2. Add sources
	sources := []string{"/home/user/skills", "/home/user/more-skills"}
	sl.SetItems(components.ToSimpleItems(sources))

	view = sl.View()
	if view == "" {
		t.Error("View should render with items")
	}

	// 3. Simulate "n" key (new item)
	sl.Cursor = len(sources) // Select "New" option
	sl.OpenForm(len(sources))

	if sl.Mode != components.SourceListModeForm {
		t.Error("Should be in form mode")
	}

	t.Log("Simple source list integration completed successfully")
}

func TestSourceListIntegration_Workspace(t *testing.T) {
	// Simulate workspace source workflow

	var savedItems []components.SourceItem

	conf := components.SourceListConfig{
		Title:       "Workspace Sources",
		Legend:      "(R=Root folder)",
		FormFields:  2,
		Field2Label: "Is Root:",
	}

	callbacks := components.SourceListCallbacks{
		OnSave: func(item components.SourceItem, idx int) (tea.Cmd, error) {
			if idx == -1 {
				savedItems = append(savedItems, item)
			} else {
				savedItems[idx] = item
			}
			return nil, errTest
		},
		OnRefresh: func() tea.Cmd { return nil },
		OnClose:   func() {},
	}

	sl := components.NewSourceList(conf, callbacks)

	// 1. Set workspace sources
	wsSources := []config.WorkspaceSource{
		{ID: "ws1", Path: "/home/user/project", IsRoot: false},
		{ID: "ws2", Path: "/home/user/projects", IsRoot: true},
	}
	sl.SetItems(components.ToWorkspaceItems(wsSources))

	view := sl.View()
	if view == "" {
		t.Error("View should render")
	}

	// 2. Verify display format
	// Item 0: [ ] /home/user/project
	// Item 1: [R] /home/user/projects

	t.Log("Workspace source list integration completed successfully")
}

func TestSourceListIntegration_CRUD(t *testing.T) {
	// Test full CRUD cycle

	var sources []string

	conf := components.SourceListConfig{Title: "Test", FormFields: 1}

	callbacks := components.SourceListCallbacks{
		OnSave: func(item components.SourceItem, idx int) (tea.Cmd, error) {
			if idx == -1 {
				sources = append(sources, item.Display())
			} else if idx < len(sources) {
				sources[idx] = item.Display()
			}
			return nil, errTest
		},
		OnDelete: func(idx int) (tea.Cmd, error) {
			if idx >= 0 && idx < len(sources) {
				sources = append(sources[:idx], sources[idx+1:]...)
			}
			return nil, errTest
		},
		OnRefresh: func() tea.Cmd { return nil },
		OnClose:   func() {},
	}

	sl := components.NewSourceList(conf, callbacks)

	// CREATE
	sl.SetItems([]components.SourceItem{})
	sl.OpenForm(-1)
	sl.Inputs[0].SetValue("/new/source")
	sl.HandleSave()

	if len(sources) != 1 {
		t.Errorf("Expected 1 source after create, got %d", len(sources))
	}

	// READ (via SetItems/View)
	sl.SetItems(components.ToSimpleItems(sources))

	// UPDATE
	sl.SetItems(components.ToSimpleItems(sources))
	sl.OpenForm(0)
	sl.Inputs[0].SetValue("/updated/source")
	sl.HandleSave()

	if sources[0] != "/updated/source" {
		t.Errorf("Expected '/updated/source', got '%s'", sources[0])
	}

	// DELETE
	sl.SetItems(components.ToSimpleItems(sources))
	sl.Cursor = 0
	sl.HandleDelete()

	if len(sources) != 0 {
		t.Errorf("Expected 0 sources after delete, got %d", len(sources))
	}

	t.Log("CRUD cycle completed successfully")
}
