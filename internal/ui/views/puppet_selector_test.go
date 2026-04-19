package views_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"kankuro/internal/ui/views"
)

func TestPuppetSelectorModal(t *testing.T) {
	modal := views.NewPuppetSelectorModal()

	// Test initial state
	if modal.IsOpen() {
		t.Error("Puppet selector should not be open initially")
	}

	// Test opening
	modal.Open()
	if !modal.IsOpen() {
		t.Error("Puppet selector should be open after Open()")
	}

	// Test view contains help text
	output := modal.View()
	if !strings.Contains(output, "Esc=back") {
		t.Error("Puppet selector view should contain 'Esc=back' help text")
	}
	if !strings.Contains(output, "← → to navigate") {
		t.Error("Puppet selector view should contain navigation help")
	}
	if !strings.Contains(output, "Enter to confirm") {
		t.Error("Puppet selector view should contain confirm help")
	}
}

func TestPuppetSelectorNavigation(t *testing.T) {
	modal := views.NewPuppetSelectorModal()
	modal.Open()

	initialCursor := modal.Cursor()

	// Test right navigation
	_, _ = modal.Update(tea.KeyMsg{Type: tea.KeyRight})
	if modal.Cursor() == initialCursor {
		t.Error("Right key should change cursor position")
	}

	// Test left navigation
	_, _ = modal.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if modal.Cursor() != initialCursor {
		t.Error("Left key should change cursor position back")
	}
}

func TestPuppetSelectorConfirmation(t *testing.T) {
	modal := views.NewPuppetSelectorModal()
	modal.Open()

	// Test enter confirmation
	_, _ = modal.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !modal.Confirmed() {
		t.Error("Enter key should set confirmed state")
	}

	// Reset and test esc cancellation
	modal = views.NewPuppetSelectorModal()
	modal.Open()
	_, _ = modal.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if !modal.Cancelled() {
		t.Error("Esc key should set cancelled state")
	}
}

func TestPuppetSelectorSelected(t *testing.T) {
	modal := views.NewPuppetSelectorModal()

	// Test setting selected puppet
	modal.SetSelected("cat")
	if modal.Selected() != "cat" {
		t.Errorf("Expected selected puppet 'cat', got '%s'", modal.Selected())
	}

	// Test with unknown puppet (should default to first)
	modal.SetSelected("unknown")
	if modal.Selected() == "" {
		t.Error("Unknown puppet should default to first available")
	}
}
