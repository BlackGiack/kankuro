package views_test

import (
	"errors"
	"testing"

	"kankuro/internal/config"
	"kankuro/internal/ui/commands"
	"kankuro/internal/ui/views"
)

func TestSessionLauncher_LaunchFlow(t *testing.T) {
	card := config.IDCard{ID: "c1", Name: "Test Card"}

	l := views.NewSessionLauncher(views.SessionLauncherCallbacks{}, &commands.Commands{})
	l.SetCard(card)

	if l.Card().ID != "c1" {
		t.Errorf("Expected card ID c1, got %s", l.Card().ID)
	}

	// Set name
	l.SetName("My Session")
	if l.Input().Value() != "My Session" {
		t.Errorf("Expected name 'My Session', got '%s'", l.Input().Value())
	}

	// Simulate launch success
	l.SetLaunchResult("s1", nil)
	if l.LaunchDone() != true {
		t.Error("Should be marked as launchDone")
	}
	if l.SessionID() != "s1" {
		t.Errorf("Expected session ID s1, got %s", l.SessionID())
	}

	// Simulate launch failure
	l.SetLaunchResult("", errors.New("launch failed"))
	if l.LaunchErr() == nil || l.LaunchErr().Error() != "launch failed" {
		t.Errorf("Expected launch error 'launch failed', got %v", l.LaunchErr())
	}
}
