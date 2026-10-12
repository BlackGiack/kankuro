package components_test

import (
	"strings"
	"testing"

	"github.com/BlackGiack/kankuro/internal/ui/components"
)

func TestModal_OpenClose(t *testing.T) {
	m := components.NewModal("Test Modal")

	if m.IsOpen() {
		t.Error("Modal should start closed")
	}

	m.Open()
	if !m.IsOpen() {
		t.Error("Modal should be open after Open()")
	}

	m.Close()
	if m.IsOpen() {
		t.Error("Modal should be closed after Close()")
	}
}

func TestModal_ViewHidden(t *testing.T) {
	m := components.NewModal("Hidden Modal")

	view := m.View()
	if view != "" {
		t.Errorf("Hidden modal should render empty string, got: %q", view)
	}
}

func TestModal_ViewVisible(t *testing.T) {
	m := components.NewModal("Visible Modal")
	m.SetContent("Are you sure?")
	m.Open()

	view := m.View()
	if view == "" {
		t.Error("Open modal should render content")
	}
	if !strings.Contains(view, "Visible Modal") {
		t.Error("Should contain title")
	}
	if !strings.Contains(view, "Are you sure?") {
		t.Error("Should contain content")
	}
}

func TestModalStateLifecycle(t *testing.T) {
	tests := []struct {
		name          string
		steps         func(s *components.ModalState)
		wantOpen      bool
		wantConfirmed bool
		wantCancelled bool
	}{
		{"zero value", func(*components.ModalState) {}, false, false, false},
		{"open", func(s *components.ModalState) { s.Open() }, true, false, false},
		{"confirm sets only confirmed", func(s *components.ModalState) { s.Open(); s.Confirm() }, true, true, false},
		{"cancel sets only cancelled", func(s *components.ModalState) { s.Open(); s.Cancel() }, true, false, true},
		{"confirm does not close", func(s *components.ModalState) { s.Open(); s.Confirm(); s.Cancel() }, true, true, true},
		{"open clears previous flags", func(s *components.ModalState) { s.Open(); s.Confirm(); s.Cancel(); s.Open() }, true, false, false},
		{"close clears everything", func(s *components.ModalState) { s.Open(); s.Confirm(); s.Cancel(); s.Close() }, false, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s components.ModalState
			tt.steps(&s)
			if s.IsOpen() != tt.wantOpen || s.Confirmed() != tt.wantConfirmed || s.Cancelled() != tt.wantCancelled {
				t.Errorf("open/confirmed/cancelled = %v/%v/%v, want %v/%v/%v",
					s.IsOpen(), s.Confirmed(), s.Cancelled(), tt.wantOpen, tt.wantConfirmed, tt.wantCancelled)
			}
		})
	}
}
