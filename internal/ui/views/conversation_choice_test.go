package views_test

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/uitest"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

func TestConversationChoiceRequiresExplicitCompatibleChoice(t *testing.T) {
	var selected contracts.ConversationChoice
	var calls int
	v := views.NewConversationChoiceView(views.ConversationChoiceCallbacks{
		OnSelect: func(_, _ string, _ bool, choice contracts.ConversationChoice) tea.Cmd {
			calls++
			selected = choice
			return nil
		},
	})
	v.SetSize(100, 30)
	setTestConversationContext(v, "session", "card", "claude", messages.ViewAgentSwitch, []testBinding{
		{ID: "wrong-driver", Name: "Wrong", Driver: "codex"},
		{ID: "right-driver", Name: "Right", Driver: "claude"},
	})
	v.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if calls != 1 || selected.SelectionID != "claude\x00right-driver" || selected.Kind != contracts.ConversationResume {
		t.Fatalf("initial selection = %#v, calls=%d", selected, calls)
	}

	// Return to the view and select the incompatible entry; it must remain
	// visible but must never be submitted to the activation service.
	setTestConversationContext(v, "session", "card", "claude", messages.ViewAgentSwitch, []testBinding{
		{ID: "wrong-driver", Name: "Wrong", Driver: "codex"},
	})
	v.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	v.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if calls != 1 {
		t.Fatalf("incompatible conversation was submitted; calls=%d", calls)
	}
}

func TestConversationChoiceViewFitsMinimumTerminal(t *testing.T) {
	v := views.NewConversationChoiceView(views.ConversationChoiceCallbacks{})
	bindings := make([]testBinding, 20)
	for i := range bindings {
		bindings[i] = testBinding{
			ID:     fmt.Sprintf("native-%03d-%s", i, strings.Repeat("x", 500)),
			Name:   uitest.AdversarialNames[i%len(uitest.AdversarialNames)],
			Driver: "codex",
		}
	}
	setTestConversationContext(v, "session", "card", "codex", messages.ViewAgentSwitch, bindings)
	v.SetSize(80, 20)
	uitest.AssertWithin(t, v.View(), 80, 20, "conversation choice minimum terminal")
}

func TestConversationChoiceFitsAfterSelectionAndAdversarialFiltering(t *testing.T) {
	bindings := []testBinding{
		{ID: "short-id", Name: "A short conversation", Driver: "codex"},
		{ID: strings.Repeat("native-id-", 52), Name: "B long conversation", Driver: "codex"},
	}
	v := views.NewConversationChoiceView(views.ConversationChoiceCallbacks{})
	setTestConversationContext(v, "session", "card", "codex", messages.ViewAgentSwitch, bindings)
	v.SetSize(80, 20)
	uitest.AssertWithin(t, v.View(), 80, 20, "short selected ID")
	v.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	uitest.AssertWithin(t, v.View(), 80, 20, "long selected ID after navigation")

	for _, query := range []string{
		strings.Repeat("q", 500),
		strings.Repeat("界", 120),
		"filter\twith\ncontrols\x1b[31m",
	} {
		filterView := views.NewConversationChoiceView(views.ConversationChoiceCallbacks{})
		setTestConversationContext(filterView, "session", "card", "codex", messages.ViewAgentSwitch, bindings)
		filterView.SetSize(80, 20)
		filterView.Update(tea.KeyPressMsg{Code: 0, Text: query})
		uitest.AssertWithin(t, filterView.View(), 80, 20, "adversarial search query")
		if strings.Contains(filterView.View(), "\x1b[31m") {
			t.Fatal("search display retained an ANSI control sequence")
		}
	}
}

//nolint:gocognit // Table exercises all maximum-detail selection and scroll paths at the terminal minimum.
func TestConversationChoiceFitsMaximumUnbrokenIDAndScrollableDriverDetails(t *testing.T) {
	longID := strings.Repeat("x", 512)
	for _, test := range []struct {
		name         string
		firstID      string
		secondID     string
		secondDriver string
		selectSecond bool
	}{
		{name: "maximum ID initially selected", firstID: longID, secondID: "short", secondDriver: "cat"},
		{name: "maximum ID selected by navigation", firstID: "short", secondID: longID, secondDriver: "cat", selectSecond: true},
		{
			name:    "maximum ID and long driver path scroll",
			firstID: "short", secondID: longID,
			secondDriver: "/long-driver/" + strings.Repeat("driver-segment/", 30) + "DRIVER_END",
			selectSecond: true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			v := views.NewConversationChoiceView(views.ConversationChoiceCallbacks{})
			setTestConversationContext(v, "session", "card", "cat", messages.ViewAgentSwitch, []testBinding{
				{ID: test.firstID, Name: "A", Driver: "cat"},
				{ID: test.secondID, Name: "B", Driver: test.secondDriver},
			})
			v.SetSize(80, 20)
			if test.selectSecond {
				v.Update(tea.KeyPressMsg{Code: tea.KeyDown})
			}
			uitest.AssertWithin(t, v.View(), 80, 20, "conversation picker maximum details")
			if test.firstID == longID || test.selectSecond {
				if strings.Count(v.View(), "x") < len(longID) {
					t.Fatal("full unbroken conversation ID is not visible in the detail viewport")
				}
			}
			if !strings.Contains(v.View(), "Enter=choose") {
				t.Fatal("picker controls/footer are not visible")
			}
			if strings.Contains(test.secondDriver, "DRIVER_END") {
				v.Update(tea.KeyPressMsg{Code: tea.KeyPgDown})
				uitest.AssertWithin(t, v.View(), 80, 20, "scrolled long driver detail")
				if !strings.Contains(v.View(), "DRIVER_END") {
					t.Fatal("scrolling did not expose the end of the full driver identity")
				}
			}
		})
	}
}

func TestConversationChoiceKeepsNewAvailableDuringFiltering(t *testing.T) {
	var selected contracts.ConversationChoice
	v := views.NewConversationChoiceView(views.ConversationChoiceCallbacks{
		OnSelect: func(_, _ string, _ bool, choice contracts.ConversationChoice) tea.Cmd { selected = choice; return nil },
	})
	v.SetSize(100, 30)
	setTestConversationContext(v, "session", "card", "codex", messages.ViewAgentSwitch, []testBinding{
		{ID: "conversation-id", Name: "Saved work", Driver: "codex"},
	})
	v.Update(tea.KeyPressMsg{Code: 0, Text: "no-match"})
	v.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if selected.Kind != contracts.ConversationNew {
		t.Fatalf("filtering hid new conversation choice: %#v", selected)
	}
}

func TestConversationChoiceCancelDoesNotActivate(t *testing.T) {
	var selected, cancelled int
	v := views.NewConversationChoiceView(views.ConversationChoiceCallbacks{
		OnSelect: func(_, _ string, _ bool, _ contracts.ConversationChoice) tea.Cmd { selected++; return nil },
		OnCancel: func(returnView messages.ViewType) tea.Cmd {
			cancelled++
			return func() tea.Msg { return messages.ScopedNavigateMsg{View: returnView} }
		},
	})
	v.SetSize(80, 20)
	setTestConversationContext(v, "session", "card", "codex", messages.ViewAgentSwitch, []testBinding{
		{ID: "saved-id", Name: "Saved", Driver: "codex"},
	})
	_, cmd := v.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("cancel should return a navigation command")
	}
	if _, ok := cmd().(messages.ScopedNavigateMsg); !ok {
		t.Fatal("cancel command did not return a navigation message")
	}
	if selected != 0 || cancelled != 1 {
		t.Fatalf("cancel activated=%d times; cancellation=%d", selected, cancelled)
	}
}

func TestConversationChoiceRefreshPreservesSelectedBindingIdentity(t *testing.T) {
	v := views.NewConversationChoiceView(views.ConversationChoiceCallbacks{})
	v.SetSize(80, 20)
	setTestConversationContext(v, "session", "card", "codex", messages.ViewAgentSwitch, []testBinding{
		{ID: "first-id", Name: "A", Driver: "codex"},
		{ID: "selected-id", Name: "B", Driver: "codex"},
	})
	v.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if selected, ok := v.SelectedOption(); !ok || selected.NativeID != "selected-id" {
		t.Fatalf("selected binding before refresh = %#v, found=%v", selected, ok)
	}
	refreshTestConversationContext(v, "session", "card", "codex", []testBinding{
		{ID: "first-id", Name: "Z", Driver: "codex"},
		{ID: "selected-id", Name: "A", Driver: "codex"},
	})
	if selected, ok := v.SelectedOption(); !ok || selected.NativeID != "selected-id" {
		t.Fatalf("refresh did not preserve binding identity: %#v, found=%v", selected, ok)
	}
}

type testBinding struct{ ID, Name, Driver string }

func testConversationOptions(_, _, driverName string, bindings []testBinding) []contracts.ConversationOption {
	options := make([]contracts.ConversationOption, 0, len(bindings))
	for _, binding := range bindings {
		option := contracts.ConversationOption{
			SelectionID: binding.Driver + "\x00" + binding.ID,
			NativeID:    binding.ID,
			Name:        binding.Name,
			Driver:      binding.Driver,
			Compatible:  binding.Driver == driverName,
		}
		if !option.Compatible {
			option.Reason = "incompatible driver"
		}
		options = append(options, option)
	}
	return options
}

//nolint:unparam // The helper mirrors the picker boundary for readable fixtures.
func setTestConversationContext(
	v *views.ConversationChoiceView,
	sessionID, cardID, driverName string,
	returnView messages.ViewType,
	bindings []testBinding,
) {
	v.SetContext(sessionID, cardID, returnView, testConversationOptions(sessionID, cardID, driverName, bindings))
}
func refreshTestConversationContext(v *views.ConversationChoiceView, sessionID, cardID, driverName string, bindings []testBinding) {
	v.RefreshContext(sessionID, cardID, testConversationOptions(sessionID, cardID, driverName, bindings))
}

func TestConversationOptionsAreDetachedFromCaller(t *testing.T) {
	options := []contracts.ConversationOption{
		{SelectionID: "token", NativeID: "native", Name: "Original", Driver: "driver", Compatible: true},
	}
	view := views.NewConversationChoiceView(views.ConversationChoiceCallbacks{})
	view.SetContext("session", "card", messages.ViewDashboard, options)
	options[0].Name = "mutated"
	selected, ok := view.SelectedOption()
	if !ok || selected.Name != "Original" {
		t.Fatalf("view retained caller slice: %+v", selected)
	}
}

//nolint:gocognit // The table exercises repeated invalidation and explicit selection for both fallback types.
func TestConversationInvalidSelectionSurvivesRepeatedRefreshes(t *testing.T) {
	original := contracts.ConversationOption{SelectionID: "original", NativeID: "one", Name: "A", Driver: "driver", Compatible: true}
	replacement := contracts.ConversationOption{SelectionID: "replacement", NativeID: "two", Name: "B", Driver: "driver", Compatible: true}
	for _, tc := range []struct {
		name    string
		options []contracts.ConversationOption
		want    contracts.ConversationChoiceKind
	}{
		{name: "empty options", want: contracts.ConversationNew},
		{name: "different compatible option", options: []contracts.ConversationOption{replacement}, want: contracts.ConversationResume},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var choices []contracts.ConversationChoice
			v := views.NewConversationChoiceView(
				views.ConversationChoiceCallbacks{OnSelect: func(_, _ string, _ bool, choice contracts.ConversationChoice) tea.Cmd {
					choices = append(choices, choice)
					return nil
				}},
			)
			v.SetSize(100, 30)
			v.SetContext("session", "card", messages.ViewDashboard, []contracts.ConversationOption{original})
			for range 3 {
				v.SetContextError("Refreshing conversations")
				v.RefreshContext("session", "card", tc.options)
				v.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
				if len(choices) != 0 {
					t.Fatalf("refresh silently chose %+v", choices)
				}
			}
			if !strings.Contains(v.View(), "no longer available") {
				t.Fatal("invalidation reason is missing")
			}
			// An arrow acknowledges the highlighted choice even when New is the only item.
			v.Update(tea.KeyPressMsg{Code: tea.KeyDown})
			if tc.want == contracts.ConversationResume {
				v.Update(tea.KeyPressMsg{Code: tea.KeyUp})
			}
			v.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
			if len(choices) != 1 || choices[0].Kind != tc.want {
				t.Fatalf("explicit choice: %+v", choices)
			}
			if tc.want == contracts.ConversationResume && choices[0].SelectionID != replacement.SelectionID {
				t.Fatalf("wrong saved choice: %+v", choices[0])
			}
		})
	}
}

func TestConversationInvalidSelectionRecoversOnlyOriginalChoice(t *testing.T) {
	original := contracts.ConversationOption{SelectionID: "original", NativeID: "one", Name: "A", Driver: "driver", Compatible: true}
	calls := 0
	v := views.NewConversationChoiceView(
		views.ConversationChoiceCallbacks{OnSelect: func(_, _ string, _ bool, choice contracts.ConversationChoice) tea.Cmd {
			if choice.Kind != contracts.ConversationResume || choice.SelectionID != original.SelectionID {
				t.Fatalf("unexpected choice: %+v", choice)
			}
			calls++
			return nil
		}},
	)
	v.SetSize(100, 30)
	v.SetContext("session", "card", messages.ViewDashboard, []contracts.ConversationOption{original})
	v.RefreshContext("session", "card", nil)
	v.SetContextError("read failed")
	v.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if calls != 0 {
		t.Fatal("failed read enabled selection")
	}
	v.SetContextError("Refreshing conversations")
	v.RefreshContext("session", "card", []contracts.ConversationOption{original})
	v.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if calls != 1 {
		t.Fatal("recovered original choice remained blocked")
	}
}

func TestConversationReadErrorStillBlocksAnExplicitChoice(t *testing.T) {
	original := contracts.ConversationOption{SelectionID: "original", NativeID: "one", Name: "A", Driver: "driver", Compatible: true}
	calls := 0
	v := views.NewConversationChoiceView(
		views.ConversationChoiceCallbacks{OnSelect: func(_, _ string, _ bool, choice contracts.ConversationChoice) tea.Cmd {
			if choice.Kind != contracts.ConversationNew {
				t.Fatalf("unexpected choice: %+v", choice)
			}
			calls++
			return nil
		}},
	)
	v.SetSize(100, 30)
	v.SetContext("session", "card", messages.ViewDashboard, []contracts.ConversationOption{original})
	v.RefreshContext("session", "card", nil)
	v.SetContextError("read failed")
	v.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	v.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if calls != 0 {
		t.Fatal("a failed read permitted submission")
	}
	v.RefreshContext("session", "card", nil)
	v.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if calls != 1 {
		t.Fatal("confirmed read did not allow the explicitly chosen New option")
	}
}

func TestConversationIncompatibleSelectionRemainsBlockedUntilRecovered(t *testing.T) {
	option := contracts.ConversationOption{SelectionID: "original", NativeID: "one", Name: "Saved", Driver: "driver", Compatible: true}
	calls := 0
	v := views.NewConversationChoiceView(
		views.ConversationChoiceCallbacks{OnSelect: func(_, _ string, _ bool, choice contracts.ConversationChoice) tea.Cmd {
			if choice.SelectionID != option.SelectionID {
				t.Fatalf("unexpected choice: %+v", choice)
			}
			calls++
			return nil
		}},
	)
	v.SetSize(100, 30)
	v.SetContext("session", "card", messages.ViewDashboard, []contracts.ConversationOption{option})
	incompatible := option
	incompatible.Compatible = false
	incompatible.Reason = "driver changed"
	for range 2 {
		v.SetContextError("Refreshing conversations")
		v.RefreshContext("session", "card", []contracts.ConversationOption{incompatible})
		v.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		if calls != 0 {
			t.Fatal("incompatible choice was silently replaced")
		}
	}
	v.RefreshContext("session", "card", []contracts.ConversationOption{option})
	v.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if calls != 1 {
		t.Fatal("original compatible choice was not restored")
	}
}
