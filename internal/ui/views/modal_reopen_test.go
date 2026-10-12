package views_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/ui/component"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

type reopenModal interface {
	component.Model
	Open()
	Close()
	IsOpen() bool
	Confirmed() bool
	Cancelled() bool
}

// TestModalsForgetConfirmationOnReopen: a modal that was confirmed (or
// cancelled) and closed must come back clean. A modal that overrides Open and
// forgets to call ModalState.Open would keep the stale flag, and the card
// editor would act on it the moment the modal reopens.
func TestModalsForgetConfirmationOnReopen(t *testing.T) {
	workspacePicker := func() reopenModal {
		m := views.NewWorkspacePickerModal()
		m.SetWorkspaces([]config.Workspace{{ID: "w1", Name: "one", Path: "/one"}})
		return m
	}

	tests := []struct {
		name string
		new  func() reopenModal
		// confirm and cancel drive the modal through its own Update.
		confirm tea.KeyPressMsg
		cancel  tea.KeyPressMsg
	}{
		{"workspace picker", workspacePicker, tea.KeyPressMsg{Code: tea.KeyEnter}, tea.KeyPressMsg{Code: tea.KeyEsc}},
		{"puppet selector", func() reopenModal { return views.NewPuppetSelectorModal() },
			tea.KeyPressMsg{Code: tea.KeyEnter}, tea.KeyPressMsg{Code: tea.KeyEsc}},
		{"agents.md editor", func() reopenModal { return views.NewAgentsMDEditorModal() },
			tea.KeyPressMsg{Code: tea.KeyEnter, Mod: tea.ModCtrl}, tea.KeyPressMsg{Code: tea.KeyEsc}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.new()

			m.Open()
			m.Update(tt.confirm)
			if !m.Confirmed() {
				t.Fatal("modal did not report Confirmed after the confirm key")
			}
			m.Close()
			m.Open()
			if m.Confirmed() {
				t.Error("Confirmed() still true after Close + Open")
			}

			m.Update(tt.cancel)
			if !m.Cancelled() {
				t.Fatal("modal did not report Cancelled after the cancel key")
			}
			m.Close()
			m.Open()
			if m.Cancelled() {
				t.Error("Cancelled() still true after Close + Open")
			}
			if !m.IsOpen() {
				t.Error("modal should be open after Open")
			}
		})
	}
}

// TestWorkspacePickerReopenWithoutClose covers Open called twice with no
// Close between, which is how a re-entrant card editor may reopen it.
func TestWorkspacePickerReopenWithoutClose(t *testing.T) {
	m := views.NewWorkspacePickerModal()
	m.Open()
	m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m.Open()
	if m.Confirmed() {
		t.Error("Confirmed() still true after a second Open")
	}
}
