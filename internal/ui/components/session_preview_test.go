package components

import (
	"strings"
	"testing"
	"time"

	"kankuro/internal/config"
	"kankuro/internal/session"
)

func TestSessionPreview_New(t *testing.T) {
	p := NewSessionPreview()
	if p == nil {
		t.Fatal("NewSessionPreview returned nil")
	}

	if p.sessionID != "" {
		t.Error("New preview should have empty session ID")
	}
}

func TestSessionPreview_Clear(t *testing.T) {
	p := NewSessionPreview()

	// Set some data
	snap := &session.Snapshot{ID: "s1", Name: "Test"}
	p.SetSession(snap, nil)
	p.SetScrollback([]byte("output"))

	// Clear
	p.Clear()

	if p.sessionID != "" {
		t.Error("Session ID should be empty after clear")
	}
	if p.snapshot != nil {
		t.Error("Snapshot should be nil after clear")
	}
	if p.scrollback != nil {
		t.Error("Scrollback should be nil after clear")
	}
}

func TestSessionPreview_SetSession(t *testing.T) {
	p := NewSessionPreview()
	p.SetSize(50, 20)

	snap := &session.Snapshot{
		ID:         "s1",
		Name:       "Test Session",
		Status:     session.StatusRunning,
		PID:        12345,
		WorkingDir: "/home/user/project",
		StartedAt:  time.Now(),
	}

	card := &config.IDCard{
		ID:     "c1",
		Name:   "Test Card",
		Puppet: "kankuro",
	}

	p.SetSession(snap, card)

	if p.sessionID != "s1" {
		t.Errorf("Expected session ID 's1', got '%s'", p.sessionID)
	}

	if p.snapshot == nil {
		t.Error("Snapshot should not be nil")
	}

	if p.card == nil {
		t.Error("Card should not be nil")
	}
}

func TestSessionPreview_View_Empty(t *testing.T) {
	p := NewSessionPreview()
	p.SetSize(40, 15)

	view := p.View()

	if view == "" {
		t.Error("View should not be empty")
	}

	// Should show placeholder
	if !strings.Contains(view, "No session") {
		t.Error("Should show placeholder when no session")
	}
}

func TestSessionPreview_View_WithSession(t *testing.T) {
	p := NewSessionPreview()
	p.SetSize(50, 20)

	snap := &session.Snapshot{
		ID:         "s1",
		Name:       "Test Session",
		Status:     session.StatusRunning,
		PID:        12345,
		WorkingDir: "/home/user/project",
		StartedAt:  time.Now(),
	}

	p.SetSession(snap, nil)

	view := p.View()

	if view == "" {
		t.Error("View should not be empty")
	}

	// Should contain session name
	if !strings.Contains(view, "Test Session") {
		t.Error("Should contain session name")
	}

	// Should contain status badge
	if !strings.Contains(view, "Running") {
		t.Error("Should contain status badge")
	}
}

func TestSessionPreview_View_WithScrollback(t *testing.T) {
	p := NewSessionPreview()
	p.SetSize(60, 25)

	snap := &session.Snapshot{
		ID:        "s1",
		Name:      "Test",
		Status:    session.StatusRunning,
		StartedAt: time.Now(),
	}

	p.SetSession(snap, nil)
	p.SetScrollback([]byte("Line 1\nLine 2\nLine 3"))

	view := p.View()

	// Should contain scrollback content
	if !strings.Contains(view, "Line 1") {
		t.Error("Should contain scrollback content")
	}
}

func TestSessionPreview_SanitizeScrollback(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
		omit     string
	}{
		{
			name:     "plain text",
			input:    "Hello World",
			contains: "Hello World",
			omit:     "",
		},
		{
			name:     "ANSI codes removed",
			input:    "\x1b[32mGreen\x1b[0m text",
			contains: "Green text",
			omit:     "\x1b[",
		},
		{
			name:     "control chars removed",
			input:    "Line1\x00\x01Line2",
			contains: "Line1Line2",
			omit:     "\x00",
		},
		{
			name:     "carriage return removed",
			input:    "Line1\r\nLine2",
			contains: "Line1\nLine2",
			omit:     "\r",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeScrollback([]byte(tt.input))

			if tt.contains != "" && !strings.Contains(result, tt.contains) {
				t.Errorf("Expected result to contain %q, got %q", tt.contains, result)
			}

			if tt.omit != "" && strings.Contains(result, tt.omit) {
				t.Errorf("Result should not contain %q, got %q", tt.omit, result)
			}
		})
	}
}

func TestStripAllControlCodes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "preserves printable",
			input:    "Hello World!",
			expected: "Hello World!",
		},
		{
			name:     "preserves newlines",
			input:    "Line1\nLine2",
			expected: "Line1\nLine2",
		},
		{
			name:     "preserves tabs",
			input:    "Col1\tCol2",
			expected: "Col1\tCol2",
		},
		{
			name:     "removes control chars",
			input:    "Hello\x00\x01\x02World",
			expected: "HelloWorld",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StripAllControlCodes(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestTruncatePath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		limit    int
		expected string
	}{
		{
			name:     "short path unchanged",
			path:     "/home/user/project",
			limit:    25,
			expected: "/home/user/project",
		},
		{
			name:     "long path truncated",
			path:     "/home/user/very/long/path/to/some/project",
			limit:    20,
			expected: "...h/to/some/project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncatePath(tt.path, tt.limit)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestRenderStatusBadge(t *testing.T) {
	tests := []struct {
		name   string
		status session.Status
		label  string
	}{
		{"running", session.StatusRunning, "Running"},
		{"waiting", session.StatusWaiting, "Waiting"},
		{"terminated", session.StatusTerminated, "Terminated"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			badge := renderStatusBadge(tt.status)
			if !strings.Contains(badge, tt.label) {
				t.Errorf("Badge should contain %q", tt.label)
			}
		})
	}
}

func TestSessionPreview_View_WaitingStatus(t *testing.T) {
	p := NewSessionPreview()
	p.SetSize(50, 20)

	snap := &session.Snapshot{
		ID:        "s1",
		Name:      "Waiting Session",
		Status:    session.StatusWaiting,
		StartedAt: time.Now(),
	}

	p.SetSession(snap, nil)

	view := p.View()

	// Waiting status should show hot pink badge
	if !strings.Contains(view, "Waiting") {
		t.Error("Should contain 'Waiting' status")
	}
}

func TestSessionPreview_SessionID(t *testing.T) {
	p := NewSessionPreview()

	if p.SessionID() != "" {
		t.Error("Initial session ID should be empty")
	}

	p.SetSessionByID("test-id")
	if p.SessionID() != "test-id" {
		t.Errorf("Expected 'test-id', got '%s'", p.SessionID())
	}
}
