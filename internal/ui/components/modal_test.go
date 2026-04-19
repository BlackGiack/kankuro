package components

import (
	"testing"
)

func TestModal_OpenClose(t *testing.T) {
	m := NewModal("Test Modal")

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
	m := NewModal("Hidden Modal")

	view := m.View()
	if view != "" {
		t.Errorf("Hidden modal should render empty string, got: %q", view)
	}
}

func TestModal_ViewVisible(t *testing.T) {
	m := NewModal("Visible Modal")
	m.SetContent("Are you sure?")
	m.Open()

	view := m.View()
	if view == "" {
		t.Error("Open modal should render content")
	}
	if !hasSubstring(view, "Visible Modal") {
		t.Error("Should contain title")
	}
	if !hasSubstring(view, "Are you sure?") {
		t.Error("Should contain content")
	}
}

func hasSubstring(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || len(s) > 0 && containsRecursive(s, substr))
}

func containsRecursive(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
