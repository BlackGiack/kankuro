package views_test

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

var errTest = errors.New("test")

func TestDashboard_New(t *testing.T) {
	callbacks := views.DashboardCallbacks{
		OnEditCard: func(_ config.IDCard) tea.Cmd {
			return func() tea.Msg { return nil }
		},
		OnLaunchCard: func(_ config.IDCard) tea.Cmd {
			return func() tea.Msg { return nil }
		},
	}

	d := views.NewDashboard(callbacks)

	if d == nil {
		t.Fatal("NewDashboard returned nil")
	}

	if d.FocusedPane() != views.PaneCards {
		t.Error("Dashboard should start with cards pane focused")
	}
}

func TestDashboard_SetData(t *testing.T) {
	d := views.NewDashboard(views.DashboardCallbacks{})

	cards := []config.IDCard{
		{ID: "c1", Name: "Card 1"},
	}
	sessions := []session.Snapshot{
		{ID: "s1", Name: "Session 1"},
	}

	d.SetData(messages.DataLoadedMsg{
		Cards:    cards,
		Sessions: sessions,
	})

	if len(d.Cards()) != 1 {
		t.Errorf("Expected 1 card, got %d", len(d.Cards()))
	}
	if len(d.Sessions()) != 1 {
		t.Errorf("Expected 1 session, got %d", len(d.Sessions()))
	}
}

func TestDashboard_FocusSessionWaitsForDataAndClearsOnlySessionSearch(t *testing.T) {
	d := views.NewDashboard(views.DashboardCallbacks{})
	d.SetData(messages.DataLoadedMsg{
		Sessions: []session.Snapshot{{ID: "existing", Name: "Existing"}},
	})
	d.FocusSession("new")
	if d.FocusedPane() != views.PaneCards {
		t.Fatal("missing target must not focus an unrelated session")
	}
	d.TogglePane()
	setDashboardSearchQuery(d, "Existing")
	if d.Search().SessionsQuery != "Existing" {
		t.Fatalf("session filter was not established: %+v", d.Search())
	}
	d.SetSessions([]session.Snapshot{
		{ID: "new", Name: "New"},
		{ID: "existing", Name: "Existing"},
	})
	selected, ok := d.SessionList().Selected()
	if !ok || selected.ID != "new" || d.FocusedPane() != views.PaneSessions || d.PreviewSessionID() != "new" {
		t.Fatalf("focus not applied to new session: selected=%+v ok=%v pane=%v preview=%q",
			selected, ok, d.FocusedPane(), d.PreviewSessionID())
	}
	if d.Search().SessionsQuery != "" {
		t.Fatalf("filter hiding target was not cleared: %+v", d.Search())
	}

	d.SessionList().SetSelectedByID("existing")
	d.SetSessions([]session.Snapshot{
		{ID: "existing", Name: "Existing"},
		{ID: "new", Name: "New"},
	})
	selected, ok = d.SessionList().Selected()
	if !ok || selected.ID != "existing" || d.PreviewSessionID() != "existing" {
		t.Fatalf("consumed focus request stole selection on later refresh: selected=%+v ok=%v preview=%q",
			selected, ok, d.PreviewSessionID())
	}
}

func TestDashboard_FocusSessionPreservesMatchingSearch(t *testing.T) {
	d := views.NewDashboard(views.DashboardCallbacks{})
	d.SetData(messages.DataLoadedMsg{Sessions: []session.Snapshot{{ID: "existing", Name: "Existing"}}})
	d.TogglePane()
	setDashboardSearchQuery(d, "New")
	d.FocusSession("new")
	d.SetSessions([]session.Snapshot{
		{ID: "existing", Name: "Existing"},
		{ID: "new", Name: "New session"},
	})
	selected, ok := d.SessionList().Selected()
	if !ok || selected.ID != "new" || d.Search().SessionsQuery != "New" || d.PreviewSessionID() != "new" {
		t.Fatalf("matching filter was not preserved with target selected: selected=%+v ok=%v search=%+v preview=%q",
			selected, ok, d.Search(), d.PreviewSessionID())
	}
}

func setDashboardSearchQuery(d *views.Dashboard, query string) {
	d.Update(tea.KeyPressMsg{Text: "/"})
	d.Update(tea.KeyPressMsg{Text: query})
	d.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
}

func TestDashboard_TogglePane(t *testing.T) {
	d := views.NewDashboard(views.DashboardCallbacks{})
	d.SetData(messages.DataLoadedMsg{
		Sessions: []session.Snapshot{{ID: "s1"}},
	})

	if d.FocusedPane() != views.PaneCards {
		t.Error("Should start on cards pane")
	}

	d.TogglePane()
	if d.FocusedPane() != views.PaneSessions {
		t.Error("Should switch to sessions pane")
	}

	d.TogglePane()
	if d.FocusedPane() != views.PaneCards {
		t.Error("Should switch back to cards pane")
	}
}

func TestDashboard_Search(t *testing.T) {
	d := views.NewDashboard(views.DashboardCallbacks{})

	// Initialize the dashboard
	d.Init()

	// Activate search
	keyMsg := tea.KeyPressMsg{Code: 0, Text: "/"}
	d.Update(keyMsg)

	if !d.Search().IsActive {
		t.Error("Search should be active after pressing '/'")
	}

	// Type search query
	testChars := []rune{'t', 'e', 's', 't'}
	for _, r := range testChars {
		charMsg := tea.KeyPressMsg{Code: 0, Text: string(r)}
		d.Update(charMsg)
	}

	if d.Search().CardsQuery != "test" {
		t.Errorf("Expected query 'test', got '%s'", d.Search().CardsQuery)
	}

	// Exit search
	keyMsg = tea.KeyPressMsg{Code: tea.KeyEsc}
	d.Update(keyMsg)

	if d.Search().IsActive {
		t.Error("Search should be inactive after escape")
	}
}

func TestDashboard_Search_NonASCII(t *testing.T) {
	d := views.NewDashboard(views.DashboardCallbacks{})
	d.Init()

	d.Update(tea.KeyPressMsg{Code: 0, Text: "/"})
	d.Update(tea.KeyPressMsg{Code: 0, Text: "é"})
	d.Update(tea.KeyPressMsg{Code: 0, Text: "日"})

	if d.Search().CardsQuery != "é日" {
		t.Errorf("Expected non-ASCII runes to reach the search query, got %q", d.Search().CardsQuery)
	}
}

func TestDashboard_RenameSession(t *testing.T) {
	var gotID, gotName string
	var submissions int
	d := views.NewDashboard(views.DashboardCallbacks{OnRenameSession: func(id, name string) tea.Cmd {
		gotID, gotName = id, name
		submissions++
		return func() tea.Msg { return nil }
	}})
	d.SetSize(80, 24-views.CoordinatorRows)
	d.SetData(messages.DataLoadedMsg{Sessions: []session.Snapshot{{ID: "session-1", Name: "Seed"}}})
	d.TogglePane()
	d.SessionList().SetSelectedByID("session-1")
	d.Update(tea.KeyPressMsg{Text: "e"})
	if !strings.Contains(d.View(), "Seed") {
		t.Fatal("rename dialog did not show the selected session name")
	}
	d.Update(tea.KeyPressMsg{Text: "é   "})
	d.SetData(messages.DataLoadedMsg{Sessions: []session.Snapshot{{ID: "session-1", Name: "Refreshed"}}})
	if !strings.Contains(d.View(), "Seedé") {
		t.Fatal("periodic data refresh discarded the rename draft")
	}
	_, cmd := d.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil || submissions != 1 {
		t.Fatalf("Enter submissions = %d, command nil = %t", submissions, cmd == nil)
	}
	if gotID != "session-1" || gotName != "Seedé" {
		t.Fatalf("rename submitted (%q, %q), want captured target and trimmed Unicode draft", gotID, gotName)
	}
	d.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if submissions != 1 {
		t.Fatalf("pending rename submitted %d times", submissions)
	}
	if !d.CompleteRename("session-1", errTest) {
		t.Fatal("matching rename result was ignored")
	}
	if !strings.Contains(d.View(), "Save failed") {
		t.Fatal("failed rename error was not shown in the dialog")
	}
	d.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if submissions != 2 {
		t.Fatalf("failed rename could not be retried: submissions=%d", submissions)
	}
	d.CompleteRename("session-1", nil)
	d.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if strings.Contains(d.View(), "Save failed") {
		t.Fatal("Esc did not cancel the retry dialog")
	}
}

func TestDashboard_RenameSessionEmptyUsesIDAndNoSelectionIsNoop(t *testing.T) {
	submissions := 0
	var gotID, gotName string
	d := views.NewDashboard(views.DashboardCallbacks{OnRenameSession: func(id, name string) tea.Cmd {
		submissions++
		gotID, gotName = id, name
		return func() tea.Msg { return nil }
	}})
	d.SetSize(80, 24-views.CoordinatorRows)
	d.Update(tea.KeyPressMsg{Text: "e"})
	if strings.Contains(d.View(), "Rename Session") {
		t.Fatal("rename dialog opened with no selected session")
	}
	d.SetData(messages.DataLoadedMsg{Sessions: []session.Snapshot{{ID: "session-id", Name: "Name"}}})
	d.TogglePane()
	d.SessionList().SetSelectedByID("session-id")
	d.Update(tea.KeyPressMsg{Text: "e"})
	for range 8 {
		d.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	}
	d.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if submissions != 1 || gotID != "session-id" || gotName != "session-id" {
		t.Fatalf("blank name rename = (%q, %q), submissions %d", gotID, gotName, submissions)
	}
}

func TestDashboard_RenameSessionDiscardsDraftWithoutSaving(t *testing.T) {
	for _, action := range []string{"escape", "dashboard change", "generation change"} {
		t.Run(action, func(t *testing.T) { checkRenameDraftDiscard(t, action) })
	}
}

func checkRenameDraftDiscard(t *testing.T, action string) {
	t.Helper()
	var savedNames []string
	d := views.NewDashboard(views.DashboardCallbacks{OnRenameSession: func(_ string, name string) tea.Cmd {
		savedNames = append(savedNames, name)
		return func() tea.Msg { return nil }
	}})
	d.SetSize(80, 24-views.CoordinatorRows)
	d.SetDashboardScope("dashboard", 1)
	d.SetData(messages.DataLoadedMsg{Sessions: []session.Snapshot{{ID: "session-id", Name: "Original"}}})
	d.TogglePane()
	d.SessionList().SetSelectedByID("session-id")
	d.Update(tea.KeyPressMsg{Text: "e"})
	d.Update(tea.KeyPressMsg{Text: " discarded draft"})
	if !strings.Contains(d.View(), "discarded draft") {
		t.Fatal("rename draft was not edited")
	}

	switch action {
	case "escape":
		_, cmd := d.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
		if cmd != nil {
			t.Fatal("cancelling rename returned a command")
		}
	case "dashboard change":
		d.SetDashboardScope("other-dashboard", 1)
	case "generation change":
		d.SetDashboardScope("dashboard", 2)
	}
	if strings.Contains(d.View(), "Rename Session") {
		t.Fatal("discarded rename dialog is still open")
	}
	if len(savedNames) != 0 {
		t.Fatalf("discarding draft saved names: %q", savedNames)
	}
	if got := d.Sessions()[0].Name; got != "Original" {
		t.Fatalf("discarding draft changed session name to %q", got)
	}

	d.Update(tea.KeyPressMsg{Text: "e"})
	if strings.Contains(d.View(), "discarded draft") {
		t.Fatal("reopening rename retained the discarded draft")
	}
	d.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if len(savedNames) != 1 || savedNames[0] != "Original" {
		t.Fatalf("reopened rename saved %q, want original name once", savedNames)
	}
}

func TestDashboard_LaunchTerminalRouting(t *testing.T) {
	for _, status := range []session.Status{session.StatusRunning, session.StatusWaiting, session.StatusTerminated} {
		t.Run(string(status), func(t *testing.T) {
			var selected string
			d := views.NewDashboard(views.DashboardCallbacks{OnLaunchTerminal: func(id string) tea.Cmd {
				selected = id
				return func() tea.Msg { return nil }
			}})
			d.SetData(messages.DataLoadedMsg{Sessions: []session.Snapshot{{ID: "target", Status: status}}})
			d.TogglePane()
			d.SessionList().SetSelectedByID("target")
			_, cmd := d.Update(tea.KeyPressMsg{Code: 0, Text: "t"})
			if selected != "target" || cmd == nil {
				t.Fatalf("terminal action selected %q, cmd %v", selected, cmd)
			}
		})
	}
	d := views.NewDashboard(views.DashboardCallbacks{
		OnLaunchTerminal: func(string) tea.Cmd {
			t.Fatal("launched with no selection")
			return nil
		},
	})
	d.TogglePane()
	d.Update(tea.KeyPressMsg{Code: 0, Text: "t"})
	d.SetData(messages.DataLoadedMsg{Sessions: []session.Snapshot{{ID: "target"}}})
	_, _ = d.Update(tea.KeyPressMsg{Code: 0, Text: "/"})
	_, _ = d.Update(tea.KeyPressMsg{Code: 0, Text: "t"})
	if d.Search().CardsQuery != "t" {
		t.Fatalf("search typing was consumed as action: %+v", d.Search())
	}
	noCallback := views.NewDashboard(views.DashboardCallbacks{})
	noCallback.SetData(messages.DataLoadedMsg{Sessions: []session.Snapshot{{ID: "target"}}})
	noCallback.TogglePane()
	noCallback.SessionList().SetSelectedByID("target")
	_, cmd := noCallback.Update(tea.KeyPressMsg{Code: 0, Text: "t"})
	if cmd != nil {
		t.Fatal("nil terminal callback returned a command")
	}
	var ideID string
	ide := views.NewDashboard(views.DashboardCallbacks{OnLaunchIDE: func(id string) tea.Cmd { ideID = id; return nil }})
	ide.SetData(messages.DataLoadedMsg{Sessions: []session.Snapshot{{ID: "ide-target"}}})
	ide.TogglePane()
	ide.SessionList().SetSelectedByID("ide-target")
	ide.Update(tea.KeyPressMsg{Code: 0, Text: "o"})
	if ideID != "ide-target" {
		t.Fatalf("IDE shortcut regression: selected %q", ideID)
	}
}

func TestDashboard_DeleteModal(t *testing.T) {
	var deletedSessionID string

	callbacks := views.DashboardCallbacks{
		OnRemove: func(sessionID string) (tea.Cmd, error) {
			deletedSessionID = sessionID
			return func() tea.Msg { return nil }, errTest
		},
	}

	d := views.NewDashboard(callbacks)
	d.SetData(messages.DataLoadedMsg{
		Sessions: []session.Snapshot{
			{ID: "s1", Name: "Test Session", UseTmpWorkspace: true, Status: session.StatusTerminated},
		},
	})
	// We need to be able to set focused pane for test
	// I'll add a setter or just use toggle if it starts on cards
	if d.FocusedPane() == views.PaneCards {
		d.TogglePane()
	}

	// Initialize and select the first session
	d.Init()
	d.SessionList().SetSelectedByID("s1")

	// Trigger delete
	keyMsg := tea.KeyPressMsg{Code: tea.KeyBackspace}
	d.Update(keyMsg)

	if !d.DeleteModal().IsOpen() {
		t.Error("Delete modal should be open")
	}

	// Confirm delete
	keyMsg = tea.KeyPressMsg{Code: 0, Text: "y"}
	d.Update(keyMsg)

	if d.DeleteModal().IsOpen() {
		t.Error("Delete modal should be closed after confirm")
	}
	if deletedSessionID != "s1" {
		t.Errorf("Expected deleted session 's1', got '%s'", deletedSessionID)
	}
}

func TestDashboard_DeleteModal_CancelWithEsc(t *testing.T) {
	var deletedSessionID string

	callbacks := views.DashboardCallbacks{
		OnRemove: func(sessionID string) (tea.Cmd, error) {
			deletedSessionID = sessionID
			return func() tea.Msg { return nil }, errTest
		},
	}

	d := views.NewDashboard(callbacks)
	d.SetData(messages.DataLoadedMsg{
		Sessions: []session.Snapshot{
			{ID: "s1", Name: "Test Session", UseTmpWorkspace: true, Status: session.StatusTerminated},
		},
	})
	if d.FocusedPane() == views.PaneCards {
		d.TogglePane()
	}

	d.Init()
	d.SessionList().SetSelectedByID("s1")

	// Trigger delete
	keyMsg := tea.KeyPressMsg{Code: tea.KeyBackspace}
	d.Update(keyMsg)

	if !d.DeleteModal().IsOpen() {
		t.Error("Delete modal should be open")
	}

	// Cancel with Esc
	keyMsg = tea.KeyPressMsg{Code: tea.KeyEsc}
	d.Update(keyMsg)

	if d.DeleteModal().IsOpen() {
		t.Error("Delete modal should be closed after Esc")
	}
	if deletedSessionID != "" {
		t.Errorf("Expected no session deleted, got '%s'", deletedSessionID)
	}
}

func TestDashboard_SetSize_Flexible(t *testing.T) {
	d := views.NewDashboard(views.DashboardCallbacks{})

	// Test wide terminal
	// width=200, dashboardSessionsWidthRatio=4, dashboardPreviewWidthRatio=2
	// sessionsWidth = 200 / 4 = 50
	// previewWidth = 50 * 2 = 100
	// cardsWidth = 200 - 50 - 100 = 50
	// cardsWidth (50) >= 46, so no diff adjustment.
	d.SetSize(200, 40)
	if d.CardsWidth() != 50 {
		t.Errorf("Expected cards width 50, got %d", d.CardsWidth())
	}

	// Test narrow terminal
	// width=80
	// sessionsWidth = 80 / 4 = 20
	// previewWidth = 20 * 2 = 40
	// cardsWidth = 80 - 20 - 40 = 20
	// cardsWidth (20) < 46, so:
	// diff = 46 - 20 = 26
	// cardsWidth = 20 + 26 = 46
	// previewWidth = 40 - 26 = 14
	d.SetSize(80, 40)
	if d.CardsWidth() != 46 {
		t.Errorf("Expected cards width 46, got %d", d.CardsWidth())
	}
	if d.PreviewWidth() != 14 {
		t.Errorf("Expected preview width 14, got %d", d.PreviewWidth())
	}
}

func TestDashboard_PreviewPersistence(t *testing.T) {
	d := views.NewDashboard(views.DashboardCallbacks{})
	d.Init()

	s1 := session.Snapshot{ID: "s1", Name: "Session 1"}
	s2 := session.Snapshot{ID: "s2", Name: "Session 2"}
	cards := []config.IDCard{{ID: "c1", Name: "Card 1"}}

	// Load sessions
	d.SetData(messages.DataLoadedMsg{
		Sessions: []session.Snapshot{s1, s2},
		Cards:    cards,
	})

	// Focus sessions and select s2
	if d.FocusedPane() == views.PaneCards {
		d.TogglePane()
	}
	d.SessionList().SetSelectedByID("s2")
	d.SyncPreviewSelection()

	if d.Preview().SessionID() != "s2" {
		t.Errorf("Expected previewing s2, got %s", d.Preview().SessionID())
	}

	// Switch back to cards pane
	d.TogglePane()
	d.SyncPreviewSelection()

	// Should still preview s2 (stable preview)
	if d.Preview().SessionID() != "s2" {
		t.Errorf("Expected preview to stay on s2 when switching panes, got %s", d.Preview().SessionID())
	}

	// Remove s2
	d.SetData(messages.DataLoadedMsg{
		Sessions: []session.Snapshot{s1},
		Cards:    cards,
	})

	// Preview should recover to s1
	if d.Preview().SessionID() != "s1" {
		t.Errorf("Expected preview to recover to s1 after s2 removed, got %s", d.Preview().SessionID())
	}

	// Remove all sessions
	d.SetData(messages.DataLoadedMsg{
		Sessions: []session.Snapshot{},
		Cards:    cards,
	})

	// Preview should be cleared
	if d.Preview().SessionID() != "" {
		t.Error("Expected preview to be cleared when no sessions remain")
	}
}

func TestDashboard_MCPActions(t *testing.T) {
	var mcpOpened bool
	callbacks := views.DashboardCallbacks{
		OnOpenMCP: func() tea.Cmd {
			mcpOpened = true
			return nil
		},
	}
	d := views.NewDashboard(callbacks)
	d.Init()

	// Press 'm' to open MCP
	keyMsg := tea.KeyPressMsg{Code: 0, Text: "m"}
	d.Update(keyMsg)

	if !mcpOpened {
		t.Error("Expected MCP form to open on 'm'")
	}
}

func TestDashboard_View(t *testing.T) {
	d := views.NewDashboard(views.DashboardCallbacks{})
	d.SetData(messages.DataLoadedMsg{
		Cards:    []config.IDCard{{ID: "c1", Name: "Test Card", Driver: "claude", Model: "claude-3"}},
		Sessions: []session.Snapshot{{ID: "s1", Name: "Test Session"}},
	})
	d.SetSize(120, 40)

	view := d.View()

	if view == "" {
		t.Error("View should not be empty")
	}
	if !strings.Contains(view, "K A N K U R O") {
		t.Error("View should contain header")
	}
	if !strings.Contains(view, "ID Cards") {
		t.Error("View should contain cards pane title")
	}
	if !strings.Contains(view, "Active Sessions") {
		t.Error("View should contain sessions pane title")
	}
}
