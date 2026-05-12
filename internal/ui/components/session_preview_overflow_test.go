package components_test

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"kankuro/internal/config"
	"kankuro/internal/session"
	"kankuro/internal/ui/components"
)

func TestSessionPreview_WidthClamp_WideRunes(t *testing.T) {
	p := components.NewSessionPreview()
	snap := &session.Snapshot{ID: "s1", Name: "Test"}
	p.SetSession(snap, nil)

	// Wide runes: CJK, Emoji, Box Drawing
	input := strings.Join([]string{
		"漢字漢字漢字漢字漢字漢字漢字漢字漢字漢字", // 10 * 2 = 20 width
		"🚀🚀🚀🚀🚀🚀🚀🚀🚀🚀",           // 10 * 2 = 20 width
		"─━│┃┄┅┆┇┈┉┊┋┌┍┎┏┐┑┒┓└┕┖┗┘┙┚┛├┝┞┟┠┡┢┣┤┥┦┧┨┩┪┫┬┭┮┯┰┱┲┳┴┵┶┷┸┹┺┻", // many width 1
	}, "\n")

	p.SetScrollback([]byte(input))

	sizes := []struct {
		w, h int
	}{
		{40, 15},
		{80, 30},
		{120, 40},
	}

	for _, size := range sizes {
		p.SetSize(size.w, size.h)
		view := p.View()

		h := lipgloss.Height(view)
		if h > size.h {
			t.Errorf("Size %dx%d: Height overflow: %d > %d", size.w, size.h, h, size.h)
		}

		lines := strings.Split(view, "\n")
		for i, line := range lines {
			w := lipgloss.Width(line)
			if w > size.w {
				t.Errorf("Size %dx%d: Line %d exceeds width: %d > %d. Content: %q", size.w, size.h, i, w, size.w, line)
			}
		}
	}
}

func TestSessionPreview_HeightClamp_LongStream(t *testing.T) {
	p := components.NewSessionPreview()
	snap := &session.Snapshot{ID: "s1", Name: "Test"}
	p.SetSession(snap, nil)

	// Append 200 short lines
	for range 200 {
		p.AppendScrollback([]byte("line\n"))
	}

	h := 20
	p.SetSize(60, h)
	view := p.View()
	height := lipgloss.Height(view)
	if height > h {
		t.Errorf("View height %d exceeds set height %d", height, h)
	}
}

func TestSessionPreview_HeightClamp_TallPuppet(t *testing.T) {
	p := components.NewSessionPreview()

	// Create a tall puppet
	tallPuppetLines := make([]string, 30)
	for i := range 30 {
		tallPuppetLines[i] = "|"
	}

	puppets := []config.Puppet{
		{ID: "tall-id", Name: "tall", Lines: tallPuppetLines, Builtin: false},
	}
	p.SetPuppets(puppets)

	snap := &session.Snapshot{ID: "s1", Name: "Test"}
	card := &config.IDCard{Puppet: "tall"}
	p.SetSession(snap, card)
	p.SetScrollback([]byte("some output"))

	h := 20
	p.SetSize(60, h)
	view := p.View()
	height := lipgloss.Height(view)
	if height > h {
		t.Errorf("View height %d exceeds set height %d with tall puppet", height, h)
	}
}

// TestSessionPreview_HardClamp_PathologicalOutput exercises the worst case:
// extremely wide lines, wide runes, many ANSI styles, and many lines. The
// clampToSize safety net must keep View() within bounds even when the
// proactive truncation and the lipgloss Max* clamps are bypassed.
func TestSessionPreview_HardClamp_PathologicalOutput(t *testing.T) {
	p := components.NewSessionPreview()
	snap := &session.Snapshot{ID: "s1", Name: "Test"}
	p.SetSession(snap, nil)

	// Inject styled wide content much larger than any reasonable pane.
	var buf strings.Builder
	for range 500 {
		buf.WriteString("\x1b[31m")
		buf.WriteString(strings.Repeat("漢", 200))
		buf.WriteString("\x1b[0m")
		buf.WriteString("\n")
	}
	p.AppendScrollback([]byte(buf.String()))

	for _, size := range []struct{ w, h int }{{40, 15}, {80, 30}, {200, 50}} {
		p.SetSize(size.w, size.h)
		view := p.View()

		if h := lipgloss.Height(view); h > size.h {
			t.Errorf("size %dx%d: height %d exceeds %d", size.w, size.h, h, size.h)
		}
		for i, line := range strings.Split(view, "\n") {
			if w := lipgloss.Width(line); w > size.w {
				t.Errorf("size %dx%d: line %d width %d exceeds %d", size.w, size.h, i, w, size.w)
			}
		}
	}
}

func TestSessionPreview_OpencodeMixedStream(t *testing.T) {
	p := components.NewSessionPreview()
	snap := &session.Snapshot{ID: "s1", Name: "Test"}
	p.SetSession(snap, nil)

	// Combine wide-rune lines, ANSI-like redraws
	p.AppendScrollback([]byte("Progress: 🚀🚀🚀\r"))
	p.AppendScrollback([]byte("Progress: 🚀🚀🚀🚀🚀\n"))
	p.AppendScrollback([]byte("漢字 message with a very long tail " + strings.Repeat("!", 100) + "\n"))

	w, h := 40, 15
	p.SetSize(w, h)
	view := p.View()

	if lipgloss.Width(view) > w {
		t.Errorf("View width %d exceeds %d", lipgloss.Width(view), w)
	}
	if lipgloss.Height(view) > h {
		t.Errorf("View height %d exceeds %d", lipgloss.Height(view), h)
	}
}
