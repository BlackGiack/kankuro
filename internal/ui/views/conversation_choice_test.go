package views_test

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/uitest"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

func TestConversationChoiceRequiresExplicitCompatibleChoice(t *testing.T) {
	var selected session.ConversationChoice
	var calls int
	v := views.NewConversationChoiceView(views.ConversationChoiceCallbacks{
		OnSelect: func(_, _ string, choice session.ConversationChoice) tea.Cmd {
			calls++
			selected = choice
			return nil
		},
	})
	v.SetSize(100, 30)
	v.SetContext("session", "card", "claude", messages.ViewAgentSwitch, []session.ConversationBinding{
		{ID: "wrong-driver", Name: "Wrong", Driver: "codex"},
		{ID: "right-driver", Name: "Right", Driver: "claude"},
	})
	v.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if calls != 1 || selected.ID != "right-driver" || selected.New {
		t.Fatalf("initial selection = %#v, calls=%d", selected, calls)
	}

	// Return to the view and select the incompatible entry; it must remain
	// visible but must never be submitted to the activation service.
	v.SetContext("session", "card", "claude", messages.ViewAgentSwitch, []session.ConversationBinding{
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
	bindings := make([]session.ConversationBinding, 20)
	for i := range bindings {
		bindings[i] = session.ConversationBinding{
			ID:     fmt.Sprintf("native-%03d-%s", i, strings.Repeat("x", 500)),
			Name:   uitest.AdversarialNames[i%len(uitest.AdversarialNames)],
			Driver: "codex",
		}
	}
	v.SetContext("session", "card", "codex", messages.ViewAgentSwitch, bindings)
	v.SetSize(80, 20)
	uitest.AssertWithin(t, v.View(), 80, 20, "conversation choice minimum terminal")
}

func TestConversationChoiceFitsAfterSelectionAndAdversarialFiltering(t *testing.T) {
	bindings := []session.ConversationBinding{
		{ID: "short-id", Name: "A short conversation", Driver: "codex"},
		{ID: strings.Repeat("native-id-", 52), Name: "B long conversation", Driver: "codex"},
	}
	v := views.NewConversationChoiceView(views.ConversationChoiceCallbacks{})
	v.SetContext("session", "card", "codex", messages.ViewAgentSwitch, bindings)
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
		filterView.SetContext("session", "card", "codex", messages.ViewAgentSwitch, bindings)
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
			v.SetContext("session", "card", "cat", messages.ViewAgentSwitch, []session.ConversationBinding{
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
	var selected session.ConversationChoice
	v := views.NewConversationChoiceView(views.ConversationChoiceCallbacks{
		OnSelect: func(_, _ string, choice session.ConversationChoice) tea.Cmd { selected = choice; return nil },
	})
	v.SetSize(100, 30)
	v.SetContext("session", "card", "codex", messages.ViewAgentSwitch, []session.ConversationBinding{
		{ID: "conversation-id", Name: "Saved work", Driver: "codex"},
	})
	v.Update(tea.KeyPressMsg{Code: 0, Text: "no-match"})
	v.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if !selected.New {
		t.Fatalf("filtering hid new conversation choice: %#v", selected)
	}
}

func TestConversationChoiceCancelDoesNotActivate(t *testing.T) {
	var selected, cancelled int
	v := views.NewConversationChoiceView(views.ConversationChoiceCallbacks{
		OnSelect: func(_, _ string, _ session.ConversationChoice) tea.Cmd { selected++; return nil },
		OnCancel: func(returnView messages.ViewType) tea.Cmd {
			cancelled++
			return func() tea.Msg { return messages.NavigateMsg{View: returnView} }
		},
	})
	v.SetSize(80, 20)
	v.SetContext("session", "card", "codex", messages.ViewAgentSwitch, []session.ConversationBinding{
		{ID: "saved-id", Name: "Saved", Driver: "codex"},
	})
	_, cmd := v.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("cancel should return a navigation command")
	}
	if _, ok := cmd().(messages.NavigateMsg); !ok {
		t.Fatal("cancel command did not return a navigation message")
	}
	if selected != 0 || cancelled != 1 {
		t.Fatalf("cancel activated=%d times; cancellation=%d", selected, cancelled)
	}
}

func TestConversationChoiceRefreshPreservesSelectedBindingIdentity(t *testing.T) {
	v := views.NewConversationChoiceView(views.ConversationChoiceCallbacks{})
	v.SetSize(80, 20)
	v.SetContext("session", "card", "codex", messages.ViewAgentSwitch, []session.ConversationBinding{
		{ID: "first-id", Name: "A", Driver: "codex"},
		{ID: "selected-id", Name: "B", Driver: "codex"},
	})
	v.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if selected, ok := v.SelectedBinding(); !ok || selected.ID != "selected-id" {
		t.Fatalf("selected binding before refresh = %#v, found=%v", selected, ok)
	}
	v.RefreshContext("session", "card", "codex", []session.ConversationBinding{
		{ID: "first-id", Name: "Z", Driver: "codex"},
		{ID: "selected-id", Name: "A", Driver: "codex"},
	})
	if selected, ok := v.SelectedBinding(); !ok || selected.ID != "selected-id" {
		t.Fatalf("refresh did not preserve binding identity: %#v, found=%v", selected, ok)
	}
}
