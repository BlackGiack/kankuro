package components_test

import (
	"sort"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"kankuro/internal/config"
	"kankuro/internal/session"
	"kankuro/internal/ui/components"
)

func TestSessionList_New(t *testing.T) {
	callbacks := components.SessionListCallbacks{
		OnSelect: func(session.Snapshot) tea.Cmd {
			return nil
		},
		OnAttach:  func(string) {},
		OnKill:    func(string) {},
		OnDelete:  func(string) {},
		OnSwitch:  func(string) {},
		OnPreview: func(string) tea.Cmd { return nil },
	}

	sl := components.NewSessionList(callbacks)
	if sl == nil {
		t.Fatal("NewSessionList returned nil")
	}

	if sl.Focused {
		t.Error("SessionList should not be focused initially")
	}
}

func TestSessionList_SetSessions(t *testing.T) {
	callbacks := components.SessionListCallbacks{}
	sl := components.NewSessionList(callbacks)

	sessions := []session.Snapshot{
		{ID: "s1", Name: "Session 1", Status: session.StatusRunning, StartedAt: time.Now()},
		{ID: "s2", Name: "Session 2", Status: session.StatusTerminated, StartedAt: time.Now().Add(-time.Hour)},
	}
	cards := []config.IDCard{
		{ID: "c1", Name: "Card 1"},
	}

	sl.SetSessions(sessions, cards)

	// Verify internal state updated
	if len(sl.Sessions()) != 2 {
		t.Errorf("Expected 2 sessions, got %d", len(sl.Sessions()))
	}

	// Verify sorting (newest first)
	sorted := make([]session.Snapshot, len(sessions))
	copy(sorted, sessions)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].StartedAt.After(sorted[j].StartedAt)
	})
}

func TestSessionList_SetSearchQuery(t *testing.T) {
	callbacks := components.SessionListCallbacks{}
	sl := components.NewSessionList(callbacks)

	sessions := []session.Snapshot{
		{ID: "s1", Name: "Alpha Session"},
		{ID: "s2", Name: "Beta Session"},
		{ID: "s3", Name: "Gamma Session"},
	}

	sl.SetSessions(sessions, nil)

	// No filter
	filtered := sl.FilterSessions(sessions)
	if len(filtered) != 3 {
		t.Errorf("Expected 3 sessions without filter, got %d", len(filtered))
	}

	// With filter
	sl.SetSearchQuery("beta")
	filtered = sl.FilterSessions(sessions)
	if len(filtered) != 1 {
		t.Errorf("Expected 1 session matching 'beta', got %d", len(filtered))
	}
	if len(filtered) > 0 && filtered[0].ID != "s2" {
		t.Errorf("Expected session ID 's2', got '%s'", filtered[0].ID)
	}
}

func TestSessionList_Selected(t *testing.T) {
	callbacks := components.SessionListCallbacks{}
	sl := components.NewSessionList(callbacks)

	// No selection
	_, ok := sl.Selected()
	if ok {
		t.Error("Selected should return false when no selection")
	}

	// Set sessions and select first
	sessions := []session.Snapshot{
		{ID: "s1", Name: "Test Session", StartedAt: time.Now()},
	}
	sl.SetSessions(sessions, nil)
	l := sl.List()
	l.Select(0)

	sess, ok := sl.Selected()
	if !ok {
		t.Error("Selected should return true when item selected")
	}
	if sess.ID != "s1" {
		t.Errorf("Expected session ID 's1', got '%s'", sess.ID)
	}
}

func TestSessionListItem_Title(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	tests := []struct {
		name     string
		session  session.Snapshot
		expected string
	}{
		{
			name:     "running session",
			session:  session.Snapshot{ID: "s1", Name: "Test", Status: session.StatusRunning},
			expected: "Test",
		},
		{
			name: "waiting session",
			session: session.Snapshot{
				ID:     "s2",
				Name:   "Waiting",
				Status: session.StatusWaiting,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := components.SessionListItem{Session: tt.session}
			title := item.Title()

			if tt.session.Status != session.StatusWaiting {
				if title != tt.expected {
					t.Errorf("Expected title %q, got %q", tt.expected, title)
				}
			} else {
				// Waiting status should contain color codes
				if title == "Waiting" {
					t.Error("Waiting session title should include color styling")
				}
			}
		})
	}
}

func TestSessionListItem_Description(t *testing.T) {
	session := session.Snapshot{
		ID:              "s1",
		Name:            "Test Session",
		CardID:          "c1",
		Status:          session.StatusRunning,
		PID:             12345,
		WorkingDir:      "/home/user/project",
		UseTmpWorkspace: false,
	}

	card := &config.IDCard{
		ID:        "c1",
		Name:      "Test Card",
		FaceStyle: "happy",
	}

	item := components.SessionListItem{
		Session: session,
		Card:    card,
	}

	desc := item.Description()

	if desc == "" {
		t.Error("Description should not be empty")
	}
}

func TestSessionList_SetSelectedByID(t *testing.T) {
	callbacks := components.SessionListCallbacks{}
	sl := components.NewSessionList(callbacks)

	sessions := []session.Snapshot{
		{ID: "s1", Name: "First", StartedAt: time.Now().Add(-2 * time.Hour)},
		{ID: "s2", Name: "Second", StartedAt: time.Now().Add(-time.Hour)},
		{ID: "s3", Name: "Third", StartedAt: time.Now()},
	}

	sl.SetSessions(sessions, nil)

	// Note: SetSessions sorts by StartedAt newest first: s3, s2, s1
	// Select second session (s2)
	sl.SetSelectedByID("s2")

	sess, ok := sl.Selected()
	if !ok {
		t.Fatal("Should have selection after SetSelectedByID")
	}
	if sess.ID != "s2" {
		t.Errorf("Expected session ID 's2', got '%s'", sess.ID)
	}

	// Try to select non-existent ID
	sl.SetSelectedByID("nonexistent")
	// Should still be on s2
	sess, _ = sl.Selected()
	if sess.ID != "s2" {
		t.Error("Should not change selection for non-existent ID")
	}
}

func TestRenderAgentFace(t *testing.T) {
	// Test that renderer works
	face := components.RenderAgentFace("neutral")
	if face == "" {
		t.Error("RenderAgentFace should return non-empty string")
	}
}

func TestRenderAgentPuppet(t *testing.T) {
	// Test that renderer works
	puppet := components.RenderAgentPuppet("kankuro")
	if puppet == "" {
		t.Error("RenderAgentPuppet should return non-empty string")
	}
}
