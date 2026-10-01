package components

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// ellipsis is what TruncateHead and TruncateTail put where text was cut.
const ellipsis = "..."

// TruncateHead trims s to at most w display columns, keeping the start and
// ending in "..." when something was cut. Measured by display width, never
// by byte, so wide runes (CJK, emoji) are not split. When w is too small for
// the ellipsis the text is cut without one.
func TruncateHead(s string, w int) string {
	if w <= 0 {
		return ""
	}
	if ansi.StringWidth(s) <= w {
		return s
	}
	tail := ellipsis
	if w <= len(ellipsis) {
		tail = ""
	}
	out := ansi.Truncate(s, w, tail)
	// A wide rune at the cut can make the grapheme-aware measure disagree with
	// the cell count Truncate worked to; never hand back something too wide.
	for ansi.StringWidth(out) > w && w > 0 {
		w--
		out = ansi.Truncate(s, w, tail)
	}
	return out
}

// TruncateTail trims s to at most w display columns, keeping the end and
// starting with "..." when something was cut. It is the right choice for
// paths (the last segments matter most) and for text being typed. See
// TruncateHead for the width rules.
func TruncateTail(s string, w int) string {
	if w <= 0 {
		return ""
	}
	total := ansi.StringWidth(s)
	if total <= w {
		return s
	}
	prefix := ellipsis
	if w <= len(ellipsis) {
		prefix = ""
	}
	keep := w - len(prefix)
	out := ansi.TruncateLeft(s, total-keep, prefix)
	for ansi.StringWidth(out) > w && keep > 0 {
		keep--
		out = ansi.TruncateLeft(s, total-keep, prefix)
	}
	return out
}

// TruncatePath truncates a path with ellipsis, keeping the tail (the most
// specific path segments) visible.
func TruncatePath(path string, limit int) string {
	return TruncateTail(path, limit)
}

// DisplayText makes user-provided text safe for a one-line slot: tabs and
// line breaks become a space, other control bytes and all ANSI escape
// sequences are dropped. A tab left in place is expanded by lipgloss to four
// columns *after* the text was fitted, which is what pushed rows past their
// box. The stored text is not touched; call this when rendering.
func DisplayText(s string) string {
	s = ansi.Strip(s)
	return strings.Map(func(r rune) rune {
		switch {
		case r == '\t' || r == '\n' || r == '\r':
			return ' '
		case r < ' ' || r == 0x7f:
			return -1
		default:
			return r
		}
	}, s)
}
