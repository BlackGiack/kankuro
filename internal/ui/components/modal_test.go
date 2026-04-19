package components_test

import (
	"strings"
	"testing"

	"kankuro/internal/ui/components"
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
