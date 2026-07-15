package views_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"kankuro/internal/config"
	"kankuro/internal/session"
	"kankuro/internal/ui/views"
)

func TestSessionLauncher_LaunchFlow(t *testing.T) {
	card := config.IDCard{ID: "c1", Name: "Test Card"}

	l := views.NewSessionLauncher(views.SessionLauncherCallbacks{})
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

// TestSessionLauncher_PersistFailureStillOffersAttach pins the Finding-1
// fix: a session.ErrPersistFailed result means the agent process is
// running and the session is fully usable, so the view must not render the
// hard "Launch failed" dead end (Esc-only), and Enter must still submit the
// name and attach - unlike a genuine launch failure, which must not.
func TestSessionLauncher_PersistFailureStillOffersAttach(t *testing.T) {
	card := config.IDCard{ID: "c1", Name: "Test Card"}
	persistErr := fmt.Errorf("%w: disk full", session.ErrPersistFailed)

	var attachCalled bool
	var setNameCalled bool
	l := views.NewSessionLauncher(views.SessionLauncherCallbacks{
		OnSetName: func(_, _ string) error {
			setNameCalled = true
			return nil
		},
		OnAttach: func(_ string) tea.Cmd {
			attachCalled = true
			return nil
		},
	})
	l.SetCard(card)
	l.SetLaunchResult("s1", persistErr)

	view := l.View()
	if strings.Contains(view, "Launch failed") {
		t.Errorf("expected no hard failure message for ErrPersistFailed, got view: %q", view)
	}
	if !strings.Contains(view, "Warning") {
		t.Errorf("expected a warning to be rendered for ErrPersistFailed, got view: %q", view)
	}
	if !strings.Contains(view, "Enter: Confirm name and attach") {
		t.Errorf("expected the normal attach flow to remain available, got view: %q", view)
	}

	// Enter must still submit the name and attach, exactly like the
	// no-error path.
	enterMsg := tea.KeyPressMsg{Code: tea.KeyEnter}
	if _, cmd := l.Update(enterMsg); cmd == nil {
		t.Fatalf("expected Update to return a command on Enter")
	} else {
		cmd() // execute the batched commands (NavigateBackMsg etc.)
	}
	if !setNameCalled {
		t.Errorf("expected OnSetName to be called")
	}
	if !attachCalled {
		t.Errorf("expected OnAttach to be called")
	}
}

// TestSessionLauncher_FatalFailureBlocksAttach pins the other half: when no
// process ever started, the view must show the hard failure and Enter must
// not attempt to attach or submit a name.
func TestSessionLauncher_FatalFailureBlocksAttach(t *testing.T) {
	card := config.IDCard{ID: "c1", Name: "Test Card"}

	var attachCalled bool
	var setNameCalled bool
	l := views.NewSessionLauncher(views.SessionLauncherCallbacks{
		OnSetName: func(_, _ string) error {
			setNameCalled = true
			return nil
		},
		OnAttach: func(_ string) tea.Cmd {
			attachCalled = true
			return nil
		},
	})
	l.SetCard(card)
	l.SetLaunchResult("", errors.New("spawn failed: no such driver"))

	view := l.View()
	if !strings.Contains(view, "Launch failed") {
		t.Errorf("expected the hard failure message, got view: %q", view)
	}
	if strings.Contains(view, "Enter: Confirm name and attach") {
		t.Errorf("expected the attach flow to be withheld, got view: %q", view)
	}

	enterMsg := tea.KeyPressMsg{Code: tea.KeyEnter}
	l.Update(enterMsg)
	if setNameCalled || attachCalled {
		t.Errorf("expected Enter to be a no-op on a fatal launch failure")
	}
}
