// Package uitest holds shape assertions and adversarial fixtures shared by the
// rendering-layout tests of internal/ui/components and internal/ui/views.
// It is imported only from _test files.
package uitest

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"kankuro/internal/config"
)

const (
	longNameLen = 200
	borderCols  = 2 // the two border runes of UserPuppet's rows
)

// AdversarialNames are display names chosen to break fixed-width layouts:
// empty, tiny, huge, wide runes, emoji, control characters and raw escapes.
var AdversarialNames = []string{ //nolint:gochecknoglobals // shared test fixtures
	"", "a", strings.Repeat("x", longNameLen), "漢字のセッション名", "🚀 rocket 🚀",
	"name\twith\ttabs", "trailing space   ", "\x1b[31mred\x1b[0m",
}

// UserPuppet builds a user puppet whose art is w columns wide and h lines
// tall, for testing puppets larger than the box they are drawn into.
func UserPuppet(w, h int) config.Puppet {
	lines := make([]string, h)
	for i := range lines {
		switch {
		case w <= 0:
		case i == 0 || i == h-1:
			lines[i] = "<" + strings.Repeat("=", max(0, w-borderCols)) + ">"
		default:
			lines[i] = "|" + strings.Repeat("#", max(0, w-borderCols)) + "|"
		}
	}
	return config.Puppet{ID: "user-puppet", Name: "user-puppet", Lines: lines}
}

// AssertExactRect fails unless out is exactly h lines of exactly w display
// columns. This is the contract for anything drawn with a border or placed
// into a fixed box.
func AssertExactRect(t testing.TB, out string, w, h int, label string) {
	t.Helper()
	lines := strings.Split(out, "\n")
	if len(lines) != h {
		t.Errorf("%s: %d lines, want %d:\n%s", label, len(lines), h, ansi.Strip(out))
		return
	}
	for i, l := range lines {
		if got := ansi.StringWidth(l); got != w {
			t.Errorf("%s: line %d is %d wide, want %d:\n%s", label, i, got, w, ansi.Strip(out))
			return
		}
	}
}

// AssertWithin fails if any line of out is wider than w or there are more
// than h lines. This is the contract for unpadded components whose parent
// pads them.
func AssertWithin(t testing.TB, out string, w, h int, label string) {
	t.Helper()
	lines := strings.Split(out, "\n")
	if len(lines) > h {
		t.Errorf("%s: %d lines, at most %d allowed:\n%s", label, len(lines), h, ansi.Strip(out))
		return
	}
	for i, l := range lines {
		if got := ansi.StringWidth(l); got > w {
			t.Errorf("%s: line %d is %d wide, at most %d allowed:\n%s", label, i, got, w, ansi.Strip(out))
			return
		}
	}
}

// Region returns the rectangle [x, x+w) x [y, y+h) of the ANSI-stripped
// render, for region checks ("this cell must be blank", "this row must not
// contain X"). x and w are display columns, not bytes or runes. Cells outside
// the render come back as spaces, so every returned row is w columns wide
// (barring a wide rune straddling the right edge).
func Region(out string, x, y, w, h int) []string {
	rows := strings.Split(ansi.Strip(out), "\n")
	region := make([]string, 0, h)
	for r := y; r < y+h; r++ {
		line := ""
		if r >= 0 && r < len(rows) {
			line = rows[r]
		}
		if x > 0 {
			line = ansi.TruncateLeft(line, x, "")
		}
		line = ansi.Truncate(line, w, "")
		if pad := w - ansi.StringWidth(line); pad > 0 {
			line += strings.Repeat(" ", pad)
		}
		region = append(region, line)
	}
	return region
}
