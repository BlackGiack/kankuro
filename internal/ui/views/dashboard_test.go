package views_test

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

var errTest = errors.New("test")

type dashboardFixture struct {
	Cards    []config.IDCard
	Sessions []session.Snapshot
	Skills   []config.Skill
	Personas []config.Persona
	Puppets  []config.Puppet
}

//nolint:gocognit // Test-only adapter converts existing backend fixtures to focused display contracts.
func setDashboardData(dashboard *views.Dashboard, msg dashboardFixture) {
	currentCards := dashboard.Cards()
	cards := make([]contracts.CardSummary, 0, len(currentCards))
	for _, card := range currentCards {
		cards = append(cards, contracts.CardSummary{
			ID: card.ID, OwnerDashboardID: card.OwnerDashboardID, Shared: card.Shared,
			Name: card.Name, Driver: card.Driver, Model: card.Model, SkillIDs: card.SkillIDs,
			PersonaIDs: card.PersonaIDs, Puppet: card.PuppetName, FaceStyle: card.FaceStyle,
			DeletionState: card.DeletionState, Launch: card.Launch, Edit: card.Edit, Delete: card.Delete,
			Retry: card.Retry, Inspect: card.Inspect,
		})
	}
	currentSessions := dashboard.SessionList().Sessions()
	sessions := make([]contracts.SessionSummary, 0, len(currentSessions))
	for _, row := range currentSessions {
		sessions = append(sessions, contracts.SessionSummary{
			ID: row.ID, DashboardID: row.DashboardID, CardID: row.CardID, Name: row.Name,
			Status: row.Status, StartedAt: row.StartedAt, PID: row.PID, WorkingDir: row.WorkingDir,
			OriginalDir: row.OriginalDir, TemporaryWorkspace: row.TemporaryWorkspace,
			CardDeleted: row.CardDeleted, CardUnavailable: row.CardUnavailable,
			Registration: row.Registration, RegistrationPersistence: row.RegistrationPersistence,
			RegistrationError: row.RegistrationError, ActiveConversationID: row.ActiveConversationID,
			ActiveConversationName: row.ActiveConversationName, FaceStyle: row.FaceStyle, Puppet: row.PuppetName,
			Attach: row.Attach, Relaunch: row.Relaunch, Stop: row.Stop, Remove: row.Remove,
			Switch: row.Switch, Rename: row.Rename, OpenIDE: row.OpenIDE, OpenTerminal: row.OpenTerminal,
		})
	}
	read := contracts.ReadResult[contracts.DashboardInventory]{}
	if msg.Cards != nil || msg.Sessions != nil {
		read.Attempted = true
		if msg.Cards != nil {
			cards = make([]contracts.CardSummary, 0, len(msg.Cards))
			for _, card := range msg.Cards {
				cards = append(cards, contracts.CardSummary{
					ID: card.ID, OwnerDashboardID: card.DashboardID, Shared: card.Shared,
					Name: card.Name, Driver: card.Driver, Model: card.Model, SkillIDs: card.SkillIDs,
					PersonaIDs: card.PersonaIDs, Puppet: card.Puppet, FaceStyle: card.FaceStyle,
					DeletionState: card.DeletionState, Launch: contracts.Availability{Enabled: card.DeletionState == ""},
					Inspect: contracts.Availability{Enabled: true}, Edit: contracts.Availability{Enabled: card.DeletionState == ""},
					Delete: contracts.Availability{Enabled: card.DeletionState == ""},
				})
			}
		}
		if msg.Sessions != nil {
			sessions = make([]contracts.SessionSummary, 0, len(msg.Sessions))
			for _, item := range msg.Sessions {
				status := item.Status
				if status == "" {
					status = session.StatusTerminated
				}
				row := contracts.SessionSummary{
					ID: item.ID, DashboardID: item.DashboardID, CardID: item.CardID, Name: item.Name,
					Status: string(status), StartedAt: item.StartedAt, PID: item.PID,
					WorkingDir: item.WorkingDir, OriginalDir: item.OriginalDir,
					TemporaryWorkspace: item.UseTmpWorkspace, CardDeleted: item.CardDeleted,
					Registration: string(item.Registration), RegistrationPersistence: string(item.RegistrationPersistence),
					RegistrationError: item.RegistrationErr, ActiveConversationID: item.ActiveConversationID,
				}
				enabled := contracts.Availability{Enabled: true}
				unavailable := contracts.Availability{Reason: "unavailable in fixture"}
				switch status {
				case session.StatusRunning, session.StatusWaiting:
					row.Attach, row.Stop, row.Rename, row.OpenIDE, row.OpenTerminal = enabled, enabled, enabled, enabled, enabled
					row.Remove, row.Relaunch, row.Switch = unavailable, unavailable, unavailable
				case session.StatusTerminated:
					row.Remove, row.Relaunch, row.Switch, row.Rename = enabled, enabled, enabled, enabled
					row.OpenIDE, row.OpenTerminal = enabled, enabled
					row.Attach, row.Stop = unavailable, unavailable
				}
				sessions = append(sessions, row)
			}
		}
		read.Value = contracts.DashboardInventory{Cards: cards, Sessions: sessions}
	}
	choices := contracts.DisplayChoicesResult{}
	if msg.Skills != nil {
		choices.Skills = contracts.ReadResult[[]contracts.NamedChoice]{
			Attempted: true,
			Value:     make([]contracts.NamedChoice, 0, len(msg.Skills)),
		}
		for _, item := range msg.Skills {
			choices.Skills.Value = append(choices.Skills.Value, contracts.NamedChoice{ID: item.ID, Name: item.Name})
		}
	}
	if msg.Personas != nil {
		choices.Personas = contracts.ReadResult[[]contracts.NamedChoice]{
			Attempted: true,
			Value:     make([]contracts.NamedChoice, 0, len(msg.Personas)),
		}
		for _, item := range msg.Personas {
			choices.Personas.Value = append(choices.Personas.Value, contracts.NamedChoice{ID: item.ID, Name: item.Name})
		}
	}
	if msg.Puppets != nil {
		choices.Puppets = contracts.ReadResult[[]contracts.PuppetChoice]{
			Attempted: true,
			Value:     make([]contracts.PuppetChoice, 0, len(msg.Puppets)),
		}
		for _, item := range msg.Puppets {
			choices.Puppets.Value = append(choices.Puppets.Value, contracts.PuppetChoice{
				ID: item.ID, Name: item.Name, Lines: item.Lines, Builtin: item.Builtin,
			})
		}
	}
	dashboard.SetInventory(read, choices)
}

func setDashboardSessions(dashboard *views.Dashboard, sessions []session.Snapshot) {
	setDashboardData(dashboard, dashboardFixture{Sessions: sessions})
}

func TestDashboardSetInventoryOwnsAndRetainsFocusedReads(t *testing.T) {
	dashboard := views.NewDashboard(views.DashboardCallbacks{})
	inventory := contracts.DashboardInventory{
		Cards:    []contracts.CardSummary{{ID: "card", Name: "Card", SkillIDs: []string{"skill"}, Puppet: "puppet"}},
		Sessions: []contracts.SessionSummary{{ID: "session", CardID: "card", Name: "Session", Status: "running", PID: 42}},
	}
	choices := contracts.DisplayChoicesResult{
		Skills: contracts.ReadResult[[]contracts.NamedChoice]{
			Attempted: true,
			Value:     []contracts.NamedChoice{{ID: "skill", Name: "Skill name"}},
		},
		Puppets: contracts.ReadResult[[]contracts.PuppetChoice]{
			Attempted: true,
			Value:     []contracts.PuppetChoice{{ID: "puppet", Name: "puppet", Lines: []string{"art"}}},
		},
	}
	inventoryResult := contracts.ReadResult[contracts.DashboardInventory]{Attempted: true, Value: inventory}
	dashboard.SetInventory(inventoryResult, choices)
	inventory.Cards[0].SkillIDs[0] = "changed"
	choices.Skills.Value[0].Name = "Changed name"
	choices.Puppets.Value[0].Lines[0] = "changed art"

	card := dashboard.Cards()[0]
	if len(card.SkillNames) != 1 || card.SkillNames[0] != "Skill name" || card.PuppetLines[0] != "art" {
		t.Fatalf("focused row aliases source values: %+v", card)
	}

	updated := contracts.DashboardInventory{
		Cards:    []contracts.CardSummary{{ID: "card", Name: "Card", SkillIDs: []string{"skill"}}},
		Sessions: []contracts.SessionSummary{{ID: "new-session", CardID: "card", Name: "New", Status: "waiting"}},
	}
	updatedRead := contracts.ReadResult[contracts.DashboardInventory]{Attempted: true, Value: updated}
	dashboard.SetInventory(updatedRead, contracts.DisplayChoicesResult{})
	if got := dashboard.Cards()[0].SkillNames[0]; got != "Skill name" {
		t.Fatalf("failed choice read discarded cached display name: %q", got)
	}
	if got := dashboard.SessionList().Sessions()[0].ID; got != "new-session" {
		t.Fatalf("successful inventory read did not replace sessions: %q", got)
	}
	emptyChoices := contracts.DisplayChoicesResult{
		Skills:   contracts.ReadResult[[]contracts.NamedChoice]{Attempted: true, Value: []contracts.NamedChoice{}},
		Personas: contracts.ReadResult[[]contracts.NamedChoice]{Attempted: true, Value: []contracts.NamedChoice{}},
		Puppets:  contracts.ReadResult[[]contracts.PuppetChoice]{Attempted: true, Value: []contracts.PuppetChoice{}},
	}
	dashboard.SetInventory(contracts.ReadResult[contracts.DashboardInventory]{}, emptyChoices)
	if got := dashboard.Cards()[0].SkillNames; len(got) != 0 {
		t.Fatalf("successful empty choices did not clear resolved names: %v", got)
	}
}

func TestDashboard_New(t *testing.T) {
	callbacks := views.DashboardCallbacks{
		OnEditCard: func(_ string) tea.Cmd {
			return func() tea.Msg { return nil }
		},
		OnLaunchCard: func(_ string) tea.Cmd {
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

	setDashboardData(d, dashboardFixture{
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
	setDashboardData(d, dashboardFixture{
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
	setDashboardSessions(d, []session.Snapshot{
		{ID: "new", Name: "New"},
		{ID: "existing", Name: "Existing"},
	})
	selected, ok := d.SessionList().Selected()
	if !ok || selected.ID != "new" || d.FocusedPane() != views.PaneSessions || d.CenterSessionID() != "new" {
		t.Fatalf("focus not applied to new session: selected=%+v ok=%v pane=%v preview=%q",
			selected, ok, d.FocusedPane(), d.CenterSessionID())
	}
	if d.Search().SessionsQuery != "" {
		t.Fatalf("filter hiding target was not cleared: %+v", d.Search())
	}

	d.SessionList().SetSelectedByID("existing")
	setDashboardSessions(d, []session.Snapshot{
		{ID: "existing", Name: "Existing"},
		{ID: "new", Name: "New"},
	})
	selected, ok = d.SessionList().Selected()
	if !ok || selected.ID != "existing" || d.CenterSessionID() != "existing" {
		t.Fatalf("consumed focus request stole selection on later refresh: selected=%+v ok=%v preview=%q",
			selected, ok, d.CenterSessionID())
	}
}

func TestDashboard_FocusSessionPreservesMatchingSearch(t *testing.T) {
	d := views.NewDashboard(views.DashboardCallbacks{})
	setDashboardData(d, dashboardFixture{Sessions: []session.Snapshot{{ID: "existing", Name: "Existing"}}})
	d.TogglePane()
	setDashboardSearchQuery(d, "New")
	d.FocusSession("new")
	setDashboardSessions(d, []session.Snapshot{
		{ID: "existing", Name: "Existing"},
		{ID: "new", Name: "New session"},
	})
	selected, ok := d.SessionList().Selected()
	if !ok || selected.ID != "new" || d.Search().SessionsQuery != "New" || d.CenterSessionID() != "new" {
		t.Fatalf("matching filter was not preserved with target selected: selected=%+v ok=%v search=%+v preview=%q",
			selected, ok, d.Search(), d.CenterSessionID())
	}
}

func setDashboardSearchQuery(d *views.Dashboard, query string) {
	d.Update(tea.KeyPressMsg{Text: "/"})
	d.Update(tea.KeyPressMsg{Text: query})
	d.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
}

func TestDashboard_TogglePane(t *testing.T) {
	d := views.NewDashboard(views.DashboardCallbacks{})
	setDashboardData(d, dashboardFixture{
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
	setDashboardData(d, dashboardFixture{Sessions: []session.Snapshot{{ID: "session-1", Name: "Seed"}}})
	d.TogglePane()
	d.SessionList().SetSelectedByID("session-1")
	d.Update(tea.KeyPressMsg{Text: "e"})
	if !strings.Contains(d.View(), "Seed") {
		t.Fatal("rename dialog did not show the selected session name")
	}
	d.Update(tea.KeyPressMsg{Text: "é   "})
	setDashboardData(d, dashboardFixture{Sessions: []session.Snapshot{{ID: "session-1", Name: "Refreshed"}}})
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
	setDashboardData(d, dashboardFixture{Sessions: []session.Snapshot{{ID: "session-id", Name: "Name"}}})
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
	setDashboardData(d, dashboardFixture{Sessions: []session.Snapshot{{ID: "session-id", Name: "Original"}}})
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
			setDashboardData(d, dashboardFixture{Sessions: []session.Snapshot{{ID: "target", Status: status}}})
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
	setDashboardData(d, dashboardFixture{Sessions: []session.Snapshot{{ID: "target"}}})
	_, _ = d.Update(tea.KeyPressMsg{Code: 0, Text: "/"})
	_, _ = d.Update(tea.KeyPressMsg{Code: 0, Text: "t"})
	if d.Search().CardsQuery != "t" {
		t.Fatalf("search typing was consumed as action: %+v", d.Search())
	}
	noCallback := views.NewDashboard(views.DashboardCallbacks{})
	setDashboardData(noCallback, dashboardFixture{Sessions: []session.Snapshot{{ID: "target"}}})
	noCallback.TogglePane()
	noCallback.SessionList().SetSelectedByID("target")
	_, cmd := noCallback.Update(tea.KeyPressMsg{Code: 0, Text: "t"})
	if cmd != nil {
		t.Fatal("nil terminal callback returned a command")
	}
	var ideID string
	ide := views.NewDashboard(views.DashboardCallbacks{OnLaunchIDE: func(id string) tea.Cmd { ideID = id; return nil }})
	setDashboardData(ide, dashboardFixture{Sessions: []session.Snapshot{{ID: "ide-target"}}})
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
	setDashboardData(d, dashboardFixture{
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
	setDashboardData(d, dashboardFixture{
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
	// width=200, dashboardSessionsWidthRatio=4, dashboardCenterWidthRatio=2
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
	if d.CenterWidth() != 14 {
		t.Errorf("Expected preview width 14, got %d", d.CenterWidth())
	}
}

func TestDashboard_PreviewPersistence(t *testing.T) {
	d := views.NewDashboard(views.DashboardCallbacks{})
	d.Init()

	s1 := session.Snapshot{ID: "s1", Name: "Session 1"}
	s2 := session.Snapshot{ID: "s2", Name: "Session 2"}
	cards := []config.IDCard{{ID: "c1", Name: "Card 1"}}

	// Load sessions
	setDashboardData(d, dashboardFixture{
		Sessions: []session.Snapshot{s1, s2},
		Cards:    cards,
	})

	// Focus sessions and select s2
	if d.FocusedPane() == views.PaneCards {
		d.TogglePane()
	}
	d.SessionList().SetSelectedByID("s2")
	d.SyncCenterSelection()

	if d.Center().SessionID() != "s2" {
		t.Errorf("Expected previewing s2, got %s", d.Center().SessionID())
	}

	// Switch back to cards pane
	d.TogglePane()
	d.SyncCenterSelection()

	// Should still preview s2 (stable preview)
	if d.Center().SessionID() != "s2" {
		t.Errorf("Expected preview to stay on s2 when switching panes, got %s", d.Center().SessionID())
	}

	// Remove s2
	setDashboardData(d, dashboardFixture{
		Sessions: []session.Snapshot{s1},
		Cards:    cards,
	})

	// Preview should recover to s1
	if d.Center().SessionID() != "s1" {
		t.Errorf("Expected preview to recover to s1 after s2 removed, got %s", d.Center().SessionID())
	}

	// Remove all sessions
	setDashboardData(d, dashboardFixture{
		Sessions: []session.Snapshot{},
		Cards:    cards,
	})

	// Preview should be cleared
	if d.Center().SessionID() != "" {
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
	setDashboardData(d, dashboardFixture{
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
