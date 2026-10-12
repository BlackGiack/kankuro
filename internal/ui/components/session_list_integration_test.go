package components_test

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/BlackGiack/kankuro/internal/ui/components"
)

func TestSessionListIntegration(t *testing.T) {
	// Simulate App workflow

	// 1. Create component with callbacks
	var selectedID string
	var previews []string

	sl := components.NewSessionList(components.SessionListCallbacks{
		OnSelect: func(s session.Snapshot) tea.Cmd {
			selectedID = s.ID
			return nil
		},
		OnPreview: func(sessionID string) tea.Cmd {
			previews = append(previews, sessionID)
			return nil
		},
	})

	// 2. Set data
	sessions := []session.Snapshot{
		{ID: "s1", Name: "Alpha", Status: session.StatusRunning, StartedAt: time.Now()},
		{ID: "s2", Name: "Beta", Status: session.StatusWaiting, StartedAt: time.Now().Add(-time.Hour)},
	}
	cards := []config.IDCard{
		{ID: "c1", Name: "Card1", FaceStyle: "happy"},
	}

	sl.SetSessions(sessions, cards)

	// 3. Initialize
	sl.Init()
	sl.SetSize(30, 10)

	// 4. Test selection
	_, ok := sl.Selected()
	if !ok {
		t.Log("No initial selection (list empty or no items)")
	}

	// 5. Test SetSelectedByID
	sl.SetSelectedByID("s2")
	sel, ok := sl.Selected()
	if !ok {
		t.Error("Should have selection after SetSelectedByID")
	}
	if sel.ID != "s2" {
		t.Errorf("Expected session ID 's2', got '%s'", sel.ID)
	}

	// 6. Test search filter
	sl.SetSearchQuery("alpha")
	// Verify filter works via exported method
	filtered := sl.FilterSessions(sessions)
	if len(filtered) != 1 {
		t.Errorf("Expected 1 session after filter, got %d", len(filtered))
	}

	t.Logf("SessionList integration test completed successfully")
	t.Logf("Previews called: %d times", len(previews))
	_ = selectedID
}
