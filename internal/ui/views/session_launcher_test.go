package views_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

func TestSessionLauncher_LaunchFlow(t *testing.T) {
	card := contracts.CardSummary{ID: "c1", Name: "Test Card"}

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
	l.SetLaunchResult(contracts.LifecycleResult{Effect: contracts.UsableActivation, SessionID: "s1", ProcessActivationID: "p"})
	if l.LaunchDone() != true {
		t.Error("Should be marked as launchDone")
	}
	if l.SessionID() != "s1" {
		t.Errorf("Expected session ID s1, got %s", l.SessionID())
	}

	// Simulate launch failure
	l.SetLaunchResult(
		contracts.LifecycleResult{Effect: contracts.NoEffect, Diagnostic: "launch failed", DiagnosticCode: contracts.DiagnosticFailure},
	)
	if l.LaunchErr() == nil || l.LaunchErr().Error() != "launch failed" {
		t.Errorf("Expected launch error 'launch failed', got %v", l.LaunchErr())
	}
}

// A session.ErrPersistFailed result means the agent process is
// running and the session is fully usable, so the view must not render the
// hard "Launch failed" dead end (Esc-only), and Enter must still submit the
// name and attach - unlike a genuine launch failure, which must not.
func TestSessionLauncher_PersistFailureStillOffersAttach(t *testing.T) {
	card := contracts.CardSummary{ID: "c1", Name: "Test Card"}
	persistErr := errors.New("session state could not be saved: disk full")

	var setNameCalled bool
	l := views.NewSessionLauncher(views.SessionLauncherCallbacks{
		OnSetNameAsync: func(_, _, _ string) tea.Cmd {
			setNameCalled = true
			return func() tea.Msg { return nil }
		},
	})
	l.SetCard(card)
	l.SetLaunchResult(
		contracts.LifecycleResult{
			Effect:              contracts.UsableActivation,
			SessionID:           "s1",
			ProcessActivationID: "p",
			Diagnostic:          persistErr.Error(),
			DiagnosticCode:      contracts.DiagnosticPersistence,
		},
	)

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
}

// TestSessionLauncher_FatalFailureBlocksAttach pins the other half: when no
// process ever started, the view must show the hard failure and Enter must
// not attempt to attach or submit a name.
func TestSessionLauncher_FatalFailureBlocksAttach(t *testing.T) {
	card := contracts.CardSummary{ID: "c1", Name: "Test Card"}

	var setNameCalled bool
	l := views.NewSessionLauncher(views.SessionLauncherCallbacks{
		OnSetNameAsync: func(_, _, _ string) tea.Cmd {
			setNameCalled = true
			return func() tea.Msg { return nil }
		},
	})
	l.SetCard(card)
	l.SetLaunchResult(
		contracts.LifecycleResult{
			Effect:         contracts.NoEffect,
			Diagnostic:     "spawn failed: no such driver",
			DiagnosticCode: contracts.DiagnosticFailure,
		},
	)

	view := l.View()
	if !strings.Contains(view, "Launch failed") {
		t.Errorf("expected the hard failure message, got view: %q", view)
	}
	if strings.Contains(view, "Enter: Confirm name and attach") {
		t.Errorf("expected the attach flow to be withheld, got view: %q", view)
	}

	enterMsg := tea.KeyPressMsg{Code: tea.KeyEnter}
	l.Update(enterMsg)
	if setNameCalled {
		t.Errorf("expected Enter to be a no-op on a fatal launch failure")
	}
}

func TestSessionLauncherWaitsForNameResultAndAllowsRetry(t *testing.T) {
	calls := 0
	launcher := views.NewSessionLauncher(views.SessionLauncherCallbacks{
		OnSetNameAsync: func(_, _, _ string) tea.Cmd {
			calls++
			return func() tea.Msg { return nil }
		},
	})
	launcher.SetLaunchResult(contracts.LifecycleResult{Effect: contracts.UsableActivation, SessionID: "session", ProcessActivationID: "p"})
	launcher.SetName("name")
	_, first := launcher.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if first == nil || calls != 1 {
		t.Fatal("name submission did not start")
	}
	if _, duplicate := launcher.Update(tea.KeyPressMsg{Code: tea.KeyEnter}); duplicate != nil || calls != 1 {
		t.Fatal("duplicate naming started while the first request was pending")
	}
	launcher.SetNameResult(errors.New("rename failed"))
	if !strings.Contains(launcher.View(), "rename failed") {
		t.Fatal("rename error was not shown")
	}
	_, retry := launcher.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if retry == nil || calls != 2 {
		t.Fatal("rename could not be retried")
	}
}

func TestSessionLauncherRejectsInvalidAndUncertainOutcomes(t *testing.T) {
	for _, result := range []contracts.LifecycleResult{
		{Effect: contracts.UsableActivation, SessionID: "s"},
		{Effect: "unknown", Diagnostic: "unknown"},
		{Effect: contracts.UncertainEffect, SessionID: "s", Diagnostic: "teardown uncertain", DiagnosticCode: contracts.DiagnosticUncertain},
	} {
		called := false
		launcher := views.NewSessionLauncher(
			views.SessionLauncherCallbacks{OnSetNameAsync: func(_, _, _ string) tea.Cmd { called = true; return nil }},
		)
		launcher.SetLaunchResult(result)
		launcher.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		if called || !strings.Contains(launcher.View(), "uncertain") {
			t.Fatalf("unsafe result enabled naming: %+v", result)
		}
	}
}

func TestSessionLauncherWorkspaceDraft(t *testing.T) {
	var calls int
	var selection []string
	launcher := views.NewSessionLauncher(
		views.SessionLauncherCallbacks{OnLaunch: func(ids []string) tea.Cmd { calls++; selection = ids; return nil }},
	)
	launcher.SetSize(80, 21)
	draft := contracts.LaunchPreparation{
		Card:            contracts.CardSummary{ID: "c", Name: "Card"},
		WorkspaceID:     "main",
		SubWorkspaceIDs: []string{"sub"},
		Workspaces: contracts.ReadResult[[]contracts.WorkspaceOption]{
			Attempted: true,
			Value:     []contracts.WorkspaceOption{{ID: "main", Name: "Main"}, {ID: "sub", Name: "Sub"}},
		},
	}
	launcher.SetCard(draft.Card)
	launcher.SetWorkspaceDraft(draft)
	if view := launcher.View(); !strings.Contains(view, "Workspace: Main") || !strings.Contains(view, "Sub Workspaces: Sub") {
		t.Fatalf("missing defaults: %s", view)
	}
	launcher.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	launcher.Update(tea.KeyPressMsg{Code: tea.KeySpace})
	launcher.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if calls != 0 {
		t.Fatal("picker confirmation launched session")
	}
	launcher.Update(tea.KeyPressMsg{Code: 'x', Mod: tea.ModCtrl})
	launcher.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if calls != 1 || len(selection) != 1 || selection[0] != "main" {
		t.Fatalf("launch selection: %v calls %d", selection, calls)
	}
	launcher.Reset()
	launcher.SetCard(draft.Card)
	launcher.SetWorkspaceDraft(draft)
	launcher.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if calls != 2 || !reflect.DeepEqual(selection, []string{"main", "sub"}) {
		t.Fatalf("untouched draft must submit displayed workspace IDs: %v", selection)
	}
}

func TestSessionLauncherPromotesSubWorkspace(t *testing.T) {
	var selected []string
	launcher := views.NewSessionLauncher(
		views.SessionLauncherCallbacks{OnLaunch: func(ids []string) tea.Cmd { selected = ids; return nil }},
	)
	launcher.SetSize(80, 21)
	launcher.SetWorkspaceDraft(contracts.LaunchPreparation{
		WorkspaceID:     "a",
		SubWorkspaceIDs: []string{"b"},
		Workspaces: contracts.ReadResult[[]contracts.WorkspaceOption]{
			Attempted: true,
			Value:     []contracts.WorkspaceOption{{ID: "a", Name: "A"}, {ID: "b", Name: "B"}},
		},
	})
	launcher.Update(tea.KeyPressMsg{Code: tea.KeySpace})
	launcher.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	launcher.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	launcher.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if len(selected) != 1 || selected[0] != "b" {
		t.Fatalf("promoted selection = %v", selected)
	}
}

func TestWorkspacePickerCtrlCMatchesEditor(t *testing.T) {
	choices := contracts.ReadResult[[]contracts.WorkspaceOption]{Attempted: true, Value: []contracts.WorkspaceOption{{ID: "a", Name: "A"}}}
	launcher := views.NewSessionLauncher(views.SessionLauncherCallbacks{})
	launcher.SetSize(80, 21)
	launcher.SetWorkspaceDraft(contracts.LaunchPreparation{WorkspaceID: "a", Workspaces: choices})
	editor := views.NewCardEditor(views.CardEditorCallbacks{})
	editor.SetSize(80, 21)
	editor.SetChoices(contracts.CardEditorChoices{Workspaces: choices})
	editor.LoadEditor(contracts.CardEditorData{CanEdit: true, Values: contracts.CardValues{WorkspaceID: "a"}})
	for editor.FocusIndex() != views.FocusWorkspace {
		editor.CycleFocus(1)
	}
	launcher.Update(tea.KeyPressMsg{Code: tea.KeySpace})
	editor.Update(tea.KeyPressMsg{Code: tea.KeySpace})
	if !strings.Contains(launcher.View(), "Select Workspace") || !strings.Contains(editor.View(), "Select Workspace") {
		t.Fatal("workspace pickers did not open")
	}
	key := tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}
	_, launcherCmd := launcher.Update(key)
	_, editorCmd := editor.Update(key)
	var launcherMsg, editorMsg tea.Msg
	if launcherCmd != nil {
		launcherMsg = launcherCmd()
	}
	if editorCmd != nil {
		editorMsg = editorCmd()
	}
	if reflect.TypeOf(launcherMsg) != reflect.TypeOf(editorMsg) {
		t.Fatalf("picker Ctrl+C: launcher %T, editor %T", launcherMsg, editorMsg)
	}
}
