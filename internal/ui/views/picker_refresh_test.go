package views_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/BlackGiack/kankuro/internal/ui/models"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

func TestWorkspacePickerKeepsActiveIdentityAcrossRefresh(t *testing.T) {
	m := views.NewWorkspacePickerModal()
	m.SetSize(100, 30)
	a, b := models.WorkspaceOption{ID: "a", Name: "A"}, models.WorkspaceOption{ID: "b", Name: "B"}
	m.SetWorkspaces([]models.WorkspaceOption{a, b})
	m.SetSelectedID("a")
	m.Open()
	for _, options := range [][]models.WorkspaceOption{{b, a}, {b}, {}, {a, b}} {
		m.SetWorkspaces(options)
		if got := m.SelectedID(); got != "a" {
			t.Fatalf("refresh changed selection to %q", got)
		}
	}
	m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if got := m.SelectedID(); got != "b" {
		t.Fatalf("navigation selected %q, want b", got)
	}
	for _, options := range [][]models.WorkspaceOption{{a}, {}, {b, a}} {
		m.SetWorkspaces(options)
		if got := m.SelectedID(); got != "b" {
			t.Fatalf("refresh lost explicit selection b: %q", got)
		}
	}
	m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if got := m.SelectedID(); got != "b" {
		t.Fatalf("confirmation selected %q", got)
	}
}

func TestPuppetPickerKeepsActiveIdentityAcrossRefresh(t *testing.T) {
	m := views.NewPuppetSelectorModal()
	a, b := models.PuppetOption{Name: "a"}, models.PuppetOption{Name: "b"}
	m.SetPuppets([]models.PuppetOption{a, b})
	m.SetSelected("a")
	m.Open()
	for _, options := range [][]models.PuppetOption{{b, a}, {b}, {}, {a, b}} {
		m.SetPuppets(options)
		if got := m.Selected(); got != "a" {
			t.Fatalf("refresh changed selection to %q", got)
		}
	}
	m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	if got := m.Selected(); got != "b" {
		t.Fatalf("navigation selected %q, want b", got)
	}
	for _, options := range [][]models.PuppetOption{{a}, {}, {b, a}} {
		m.SetPuppets(options)
		if got := m.Selected(); got != "b" {
			t.Fatalf("refresh lost explicit selection b: %q", got)
		}
	}
	m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if got := m.Selected(); got != "b" {
		t.Fatalf("confirmation selected %q", got)
	}
}

func TestWorkspacePickerSearchSelectionSurvivesRefresh(t *testing.T) {
	m := views.NewWorkspacePickerModal()
	m.SetSize(100, 30)
	options := []models.WorkspaceOption{{ID: "a", Name: "Alpha"}, {ID: "b", Name: "Beta"}}
	m.SetWorkspaces(options)
	m.SetSelectedID("a")
	m.Open()
	m.Update(tea.KeyPressMsg{Text: "Beta"})
	if got := m.SelectedID(); got != "b" {
		t.Fatalf("search selected %q, want b", got)
	}
	m.SetWorkspaces(options)
	if got := m.SelectedID(); got != "b" {
		t.Fatalf("refresh lost highlighted search result: %q", got)
	}
	m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if got := m.SelectedID(); got != "b" {
		t.Fatalf("confirmation selected %q, want b", got)
	}
}

func TestWorkspacePickerRefreshPreservesUnresolvedSelectionWhenSearchHasNoResults(t *testing.T) {
	m := views.NewWorkspacePickerModal()
	m.SetSize(100, 30)
	options := []models.WorkspaceOption{{ID: "a", Name: "Alpha"}}
	m.SetWorkspaces(options)
	m.SetSelectedID("missing")
	m.Open()
	m.Update(tea.KeyPressMsg{Text: "not-found"})
	m.SetWorkspaces(options)
	if got := m.SelectedID(); got != "missing" {
		t.Fatalf("refresh lost unresolved selection: %q", got)
	}
}

func TestWorkspacePickerSearchChoosesVisibleOptionFromUnresolvedSelection(t *testing.T) {
	m := views.NewWorkspacePickerModal()
	m.SetSize(100, 30)
	options := []models.WorkspaceOption{{ID: "a", Name: "Alpha"}, {ID: "b", Name: "Beta"}}
	m.SetWorkspaces(options)
	m.SetSelectedID("missing")
	m.Open()
	m.Update(tea.KeyPressMsg{Text: "Beta"})
	if got := m.SelectedID(); got != "b" {
		t.Fatalf("search should choose visible option, got %q", got)
	}
	m.SetWorkspaces(options)
	m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if got := m.SelectedID(); got != "b" {
		t.Fatalf("confirmation selected %q, want b", got)
	}
}
