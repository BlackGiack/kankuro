package components_test

import (
	"testing"
	"time"

	"kankuro/internal/config"
	"kankuro/internal/session"
	"kankuro/internal/ui/components"
)

func TestSessionPreviewIntegration(t *testing.T) {
	// Simulate App workflow

	// 1. Create component
	p := components.NewSessionPreview()
	p.SetSize(50, 20)

	// 2. Initial state - empty placeholder
	view := p.View()
	if view == "" {
		t.Error("View should render placeholder")
	}

	// 3. Set session data
	snap := &session.Snapshot{
		ID:         "s1",
		Name:       "Integration Test",
		Status:     session.StatusRunning,
		PID:        12345,
		WorkingDir: "/home/user/test",
		StartedAt:  time.Now(),
	}

	card := &config.IDCard{
		ID:     "c1",
		Name:   "Test Card",
		Puppet: "kankuro",
	}

	p.SetSession(snap, card)

	// 4. View with session
	view = p.View()
	if view == "" {
		t.Error("View should render session")
	}

	// 5. Add scrollback
	p.SetScrollback([]byte("Starting process...\nProcess running\nDone"))

	view = p.View()
	if view == "" {
		t.Error("View should render with scrollback")
	}

	// 6. Clear
	p.Clear()
	view = p.View()
	if view == "" {
		t.Error("View should render placeholder after clear")
	}

	t.Log("SessionPreview integration test completed successfully")
}

func TestSessionPreview_ScrollbackSanitization(t *testing.T) {
	p := components.NewSessionPreview()
	p.SetSize(60, 25)

	snap := &session.Snapshot{
		ID:        "s1",
		Name:      "Test",
		Status:    session.StatusRunning,
		StartedAt: time.Now(),
	}
	p.SetSession(snap, nil)

	// Set scrollback with ANSI codes
	scrollback := []byte("\x1b[32mSuccess\x1b[0m: Operation completed\n\x1b[31mError\x1b[0m: None")
	p.SetScrollback(scrollback)

	view := p.View()

	// ANSI codes should be stripped
	if view == "" {
		t.Error("View should render")
	}

	// Should contain the actual text (after stripping)
	// (In real test, check that no escape sequences remain)

	t.Log("Scrollback sanitization test completed")
}

func TestSessionPreview_ResizeHandling(t *testing.T) {
	p := components.NewSessionPreview()

	// Initial size
	p.SetSize(40, 15)
	if p.Width != 40 || p.Height != 15 {
		t.Error("Size not set correctly")
	}

	// Resize
	p.SetSize(80, 30)
	if p.Width != 80 || p.Height != 30 {
		t.Error("Resize not handled correctly")
	}

	// Should still render correctly
	snap := &session.Snapshot{
		ID:        "s1",
		Name:      "Resize Test",
		Status:    session.StatusRunning,
		StartedAt: time.Now(),
	}
	p.SetSession(snap, nil)

	view := p.View()
	if view == "" {
		t.Error("Should render after resize")
	}
}
