package components_test

import (
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/BlackGiack/kankuro/internal/ui/components"
)

func TestSessionPreview_RedrawOverwrites(t *testing.T) {
	p := components.NewSessionPreview()
	p.SetSize(100, 20)
	snap := &session.Snapshot{ID: "s1", Name: "Test"}
	p.SetSession(snap, nil)

	// Simulate redraw: Progress updates
	chunks := [][]byte{
		[]byte("Progress: 10%\r"),
		[]byte("Progress: 20%\r"),
		[]byte("Progress: 100%\nDone\n"),
	}

	for _, chunk := range chunks {
		p.AppendScrollback(chunk)
	}

	view := p.View()

	// Should ONLY contain the final states, not the intermediary ones.
	if strings.Contains(view, "10%") || strings.Contains(view, "20%") {
		t.Errorf("View should NOT contain intermediary redraw states. View content: %q", view)
	}
	if !strings.Contains(view, "100%") || !strings.Contains(view, "Done") {
		t.Errorf("View should contain the final redraw state. View content: %q", view)
	}
}

func TestSessionPreview_StaleContentOnSwitch(t *testing.T) {
	p := components.NewSessionPreview()
	p.SetSize(100, 20)

	// 1. Session A with output
	snapA := &session.Snapshot{ID: "A", Name: "Session A"}
	p.SetSession(snapA, nil)
	p.SetScrollback([]byte("Content from A"))

	if !strings.Contains(p.View(), "Content from A") {
		t.Fatal("View should show content from A")
	}

	// 2. Switch to Session B with NO output
	snapB := &session.Snapshot{ID: "B", Name: "Session B"}
	p.SetSession(snapB, nil)
	// Important: We DON'T call SetScrollback yet, simulating a fresh switch to a session with no data yet

	view := p.View()
	if strings.Contains(view, "Content from A") {
		t.Error("View should NOT show stale content from Session A after switching to B")
	}
	if !strings.Contains(view, "(no output yet)") {
		t.Errorf("View should show empty state for Session B. Got: %q", view)
	}
}

func TestSessionPreview_GrowingLongLine(t *testing.T) {
	p := components.NewSessionPreview()
	p.SetSize(60, 20)
	snap := &session.Snapshot{ID: "s1", Name: "Test"}
	p.SetSession(snap, nil)

	// 1. Initial part
	p.AppendScrollback([]byte("Long line: "))

	// 2. Append enough to exceed width
	p.AppendScrollback([]byte(strings.Repeat("a", 100)))

	// 3. Append MORE and verify it's visible (showing the tail)
	p.AppendScrollback([]byte("TARGET_CONTENT"))
	view := p.View()

	if !strings.Contains(view, "TARGET_CONTENT") {
		t.Errorf("View is frozen on the start of the long line. Tail content missing. View: %q", view)
	}
}

func TestSessionPreview_FreezePrevention(t *testing.T) {
	p := components.NewSessionPreview()
	p.SetSize(60, 20)
	snap := &session.Snapshot{ID: "s1", Name: "Test"}
	p.SetSession(snap, nil)

	// 1. Initial output
	p.AppendScrollback([]byte("Step 1: Start\n"))
	if !strings.Contains(p.View(), "Step 1: Start") {
		t.Fatal("Initial output should be visible")
	}

	// 2. Send noise
	p.AppendScrollback([]byte("... [123] ...\n"))

	// 3. Send new meaningful output. It MUST appear.
	p.AppendScrollback([]byte("Step 3: Success\n"))
	view := p.View()
	if !strings.Contains(view, "Step 3: Success") {
		t.Error("View is frozen! New meaningful line did not appear after noise.")
	}
}

func TestSessionPreview_New(t *testing.T) {
	p := components.NewSessionPreview()
	if p == nil {
		t.Fatal("NewSessionPreview returned nil")
	}

	if p.SessionID() != "" {
		t.Error("New preview should have empty session ID")
	}
}

func TestSessionPreview_Clear(t *testing.T) {
	p := components.NewSessionPreview()

	// Set some data
	snap := &session.Snapshot{ID: "s1", Name: "Test"}
	p.SetSession(snap, nil)
	p.SetScrollback([]byte("output"))

	// Clear
	p.Clear()

	if p.SessionID() != "" {
		t.Error("Session ID should be empty after clear")
	}
	if p.Snapshot() != nil {
		t.Error("Snapshot should be nil after clear")
	}
	if p.Scrollback() != nil {
		t.Error("Scrollback should be nil after clear")
	}
}

func TestSessionPreview_SetSession(t *testing.T) {
	p := components.NewSessionPreview()
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

	if p.SessionID() != "s1" {
		t.Errorf("Expected session ID 's1', got '%s'", p.SessionID())
	}

	if p.Snapshot() == nil {
		t.Error("Snapshot should not be nil")
	}

	if p.Card() == nil {
		t.Error("Card should not be nil")
	}
}

func TestSessionPreview_View_Empty(t *testing.T) {
	p := components.NewSessionPreview()
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
	p := components.NewSessionPreview()
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
	p := components.NewSessionPreview()
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
			result := components.StripAllControlCodes(tt.input)
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
			result := components.TruncatePath(tt.path, tt.limit)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestTruncatePath_MultibyteRuneSafety(t *testing.T) {
	tests := []struct {
		name  string
		path  string
		limit int
	}{
		{"CJK path", "/家/项目/很长的路径/名字/文件夹/子目录", 12},
		{"combining marks", "/home/usér/proje\u0301ct/nested/deep/path", 15},
		{"emoji", "/home/👍🎉🚀/project/deep/nested/path", 12},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := components.TruncatePath(tt.path, tt.limit)
			if !utf8.ValidString(result) {
				t.Fatalf("result is not valid UTF-8: %q", result)
			}
			if w := runewidth.StringWidth(result); w > tt.limit {
				t.Errorf("result width %d exceeds limit %d: %q", w, tt.limit, result)
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
			badge := components.RenderStatusBadge(tt.status)
			if !strings.Contains(badge, tt.label) {
				t.Errorf("Badge should contain %q", tt.label)
			}
		})
	}
}

func TestSessionPreview_View_WaitingStatus(t *testing.T) {
	p := components.NewSessionPreview()
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

func TestSessionPreview_MeaningfulHeuristic(t *testing.T) {
	p := components.NewSessionPreview()
	p.SetSize(60, 20)
	snap := &session.Snapshot{ID: "s1", Name: "Test"}
	p.SetSession(snap, nil)

	// 1. Initial meaningful output
	p.SetScrollback([]byte("Hello, this is a real message from the agent."))
	view1 := p.View()
	if !strings.Contains(view1, "Hello, this is a real message") {
		t.Fatal("Initial view should contain meaningful output")
	}

	// 2. Control sequence residue (should NOT replace the sticky preview)
	// Example from review: 5;75HI[0 q[?25h[?2026l
	noise := "5;75HI[0 q[?25h[?2026l"
	p.SetScrollback([]byte(noise))
	view2 := p.View()

	if strings.Contains(view2, noise) {
		t.Errorf("View should NOT contain noise residue %q", noise)
	}
	if !strings.Contains(view2, "Hello, this is a real message") {
		t.Error("View should still contain the PREVIOUS meaningful output, not the noise")
	}
}

func TestSessionPreview_Sticky(t *testing.T) {
	p := components.NewSessionPreview()
	p.SetSize(60, 20)
	snap := &session.Snapshot{ID: "s1", Name: "Test"}
	p.SetSession(snap, nil)

	// 1. Initial meaningful output
	p.SetScrollback([]byte("First meaningful line"))
	view1 := p.View()
	if !strings.Contains(view1, "First meaningful line") {
		t.Error("Initial view should contain meaningful output")
	}

	// 2. Noisy output (should keep old meaningful output)
	p.SetScrollback([]byte("\x1b[123;456u\x1b[K")) // Just noise/control
	view2 := p.View()
	if !strings.Contains(view2, "First meaningful line") {
		t.Error("View should be sticky and keep old meaningful output when new output is noisy")
	}

	// 3. New meaningful output
	p.SetScrollback([]byte("Second meaningful line"))
	view3 := p.View()
	if !strings.Contains(view3, "Second meaningful line") {
		t.Error("View should update when new meaningful output arrives")
	}
	if strings.Contains(view3, "First meaningful line") {
		t.Error("View should not contain old meaningful output when new one arrives")
	}

	// 4. Session change (should clear cache)
	snap2 := &session.Snapshot{ID: "s2", Name: "Test 2"}
	p.SetSession(snap2, nil)
	p.SetScrollback([]byte("\x1b[K")) // Noisy
	view4 := p.View()
	if strings.Contains(view4, "Second meaningful line") {
		t.Error("Cache should be cleared when session changes")
	}
	if !strings.Contains(view4, "(no output yet)") {
		t.Error("Should show placeholder for new session with no meaningful output")
	}
}

func TestSessionPreview_CappedSequenceResidue(t *testing.T) {
	p := components.NewSessionPreview()
	p.SetSize(60, 20)
	snap := &session.Snapshot{ID: "s1", Name: "Test"}
	p.SetSession(snap, nil)

	// Part 1: Some initial meaningful text
	initialText := "Initial meaningful line\n"

	// Part 2: A long Kitty sequence that spans the 8KB boundary
	payload := strings.Repeat("ThisIsSomeBase64LikePayloadThatLooksMeaningful", 200) // ~9KB
	kittySequence := "\x1b_G" + payload + "\x1b\\"

	fullScrollback := initialText + kittySequence

	p.SetScrollback([]byte(fullScrollback))

	view := p.View()

	// If the fix is there (skipping the first partial line), it should NOT contain the payload residue.
	if strings.Contains(view, "ThisIsSomeBase64") {
		t.Errorf("View should NOT contain residue from cut Kitty payload. View content: %q", view)
	}
}

func TestSessionPreview_LargeSetScrollbackNoDiscard(t *testing.T) {
	p := components.NewSessionPreview()
	p.SetSize(100, 20)
	snap := &session.Snapshot{ID: "s1", Name: "Test"}
	p.SetSession(snap, nil)

	// Case 1: Large multiline buffer ( > 16KB but < 256KB)
	largeMultiline := strings.Repeat("Other lines\n", 1000) + "Final line: important"
	p.SetScrollback([]byte(largeMultiline))
	view1 := p.View()
	if !strings.Contains(view1, "Final line: important") {
		t.Errorf("Should NOT discard the last line of a multiline buffer. View content: %q", view1)
	}

	// Case 2: Large single line buffer ( > 16KB but < 256KB)
	// We check for the TAIL content since we now use tail truncation.
	largeSingleLine := strings.Repeat("a", 15000) + "EndOfLine"
	p.SetScrollback([]byte(largeSingleLine))
	view2 := p.View()
	if !strings.Contains(view2, "EndOfLine") {
		t.Errorf("Should NOT discard a large single line buffer and should show the tail. View content: %q", view2)
	}

	// Case 3: Exactly session.ScrollbackLimit buffer that is meaningful text (contains spaces)
	limitText := strings.Repeat("word ", session.ScrollbackLimit/5-10) + "\nAtLimit"
	limitData := []byte(limitText)
	if len(limitData) > session.ScrollbackLimit {
		limitData = limitData[len(limitData)-session.ScrollbackLimit:]
	}
	p.SetScrollback(limitData)
	view3 := p.View()
	if !strings.Contains(view3, "AtLimit") {
		t.Errorf("Should NOT discard meaningful text even if at session limit. View content: %q", view3)
	}

	// Case 4: Exactly session.ScrollbackLimit buffer starting with residue (no spaces, long)
	// Residue must now be > 128 chars to be discarded.
	residue := strings.Repeat("ThisIsSomeVeryLongResidueWithNoSpacesAndNoSymbols", 10) // ~500 chars
	limitResidue := residue + "\nValid text after residue"
	limitResidueData := make([]byte, session.ScrollbackLimit)
	copy(limitResidueData[session.ScrollbackLimit-len(limitResidue):], limitResidue)

	p.SetScrollback(limitResidueData)
	view4 := p.View()
	if strings.Contains(view4, residue) {
		t.Errorf("Should HAVE discarded residue at session limit. View content: %q", view4)
	}
	if !strings.Contains(view4, "Valid text after residue") {
		t.Error("Should have kept valid text after residue")
	}
}

func TestSessionPreview_LineBreakRecovery(t *testing.T) {
	p := components.NewSessionPreview()
	p.SetSize(100, 20)
	snap := &session.Snapshot{ID: "s1", Name: "Test"}
	p.SetSession(snap, nil)

	// Case 1: \r as redraw overwrite. "Line 1" is overwritten by "Line 2".
	p.SetScrollback([]byte("Line 1\rLine 2\rLine 3"))
	view1 := p.View()
	if strings.Contains(view1, "Line 1") || strings.Contains(view1, "Line 2") {
		t.Errorf("Should have overwritten previous content on the same line with \\r. View: %q", view1)
	}
	if !strings.Contains(view1, "Line 3") {
		t.Errorf("Should contain final line after redraw. View: %q", view1)
	}

	// Case 2: Mixed \r\n and redraws.
	// Redraws within a line should collapse to the final state of that line.
	// Lines separated by \n should remain separate.
	redrawData := "Step 1: In Progress\rStep 1: Done\nStep 2: Starting\rStep 2: Done"
	p.SetScrollback([]byte(redrawData))
	view2 := p.View()
	if !strings.Contains(view2, "Step 1: Done") || !strings.Contains(view2, "Step 2: Done") {
		t.Errorf("Should show final state of each line. View: %q", view2)
	}
	if strings.Contains(view2, "In Progress") || strings.Contains(view2, "Starting") {
		t.Errorf("Should NOT contain intermediary redraw states. View: %q", view2)
	}
}
