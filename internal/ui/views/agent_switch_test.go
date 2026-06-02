package views_test

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"kankuro/internal/config"
	"kankuro/internal/ui/messages"
	"kankuro/internal/ui/views"
)

func TestAgentSwitchView_Rendering(t *testing.T) {
	v := views.NewAgentSwitchView(views.AgentSwitchCallbacks{})
	v.SetSession("s1", "Session 1")
	v.SetSize(100, 40)

	view := v.View()
	if !strings.Contains(view, "Switch Agent (session: Session 1)") {
		t.Error("View should contain session name in header")
	}
	if !strings.Contains(view, "Type to search...") {
		t.Error("View should show search placeholder")
	}
}

func TestAgentSwitchView_SearchAndFilter(t *testing.T) {
	v := views.NewAgentSwitchView(views.AgentSwitchCallbacks{})
	v.SetSize(100, 40)
	v.SetData(messages.DataLoadedMsg{
		Cards: []config.IDCard{
			{ID: "c1", Name: "Alpha"},
			{ID: "c2", Name: "Beta"},
		},
	})

	// Type 'a'
	v.Update(tea.KeyPressMsg{Code: 0, Text: "a"})
	view := v.View()
	if !strings.Contains(view, "Search: a") {
		t.Error("View should show current search")
	}
	if !strings.Contains(view, "Alpha") {
		t.Errorf("Alpha should be visible, got: %s", view)
	}
}

func TestAgentSwitchView_EscapeBehavior(t *testing.T) {
	var cancelCmdCalled bool
	v := views.NewAgentSwitchView(views.AgentSwitchCallbacks{
		OnCancel: func() tea.Cmd {
			return func() tea.Msg {
				cancelCmdCalled = true
				return nil
			}
		},
	})

	// 1. With search active
	v.Update(tea.KeyPressMsg{Code: 0, Text: "a"})
	_, cmd := v.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if cancelCmdCalled {
		t.Error("Esc should not cancel while search is active")
	}
	if strings.Contains(v.View(), "Search: a") {
		t.Error("Search should be cleared")
	}
	if cmd != nil {
		t.Error("Esc with active search should not emit a navigation/cancel command")
	}

	// 2. Without search active
	_, cmd = v.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("Esc with empty search should emit cancel command")
	}
	_ = cmd()
	if !cancelCmdCalled {
		t.Error("Esc should cancel while search is empty")
	}
}

func TestAgentSwitchView_Switch(t *testing.T) {
	var switchedSessionID, switchedCardID string
	v := views.NewAgentSwitchView(views.AgentSwitchCallbacks{
		OnSwitch: func(sid, cid string) tea.Cmd {
			switchedSessionID = sid
			switchedCardID = cid
			return nil
		},
	})
	v.SetSession("s1", "Session 1")
	v.SetData(messages.DataLoadedMsg{
		Cards: []config.IDCard{{ID: "c1", Name: "Alpha"}},
	})

	// Press Enter
	v.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if switchedSessionID != "s1" || switchedCardID != "c1" {
		t.Errorf("Expected switch(s1, c1), got (%s, %s)", switchedSessionID, switchedCardID)
	}
}
