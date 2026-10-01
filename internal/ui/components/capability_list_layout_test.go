package components_test

import (
	"fmt"
	"strings"
	"testing"

	bubbletea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"kankuro/internal/ui/components"
)

// Contract (bonus_impl_plan.md, part B): a CapabilityList that was given a
// height renders EXACTLY that many lines in every state, and the row each
// element sits on depends on the size and on the unfiltered item count only,
// never on the cursor or the search:
//
//	row 0     the header
//	rows 1..  the body, Height-1 rows. When the list fits, items start on
//	          row 1. When it overflows (and the body has 3+ rows) the first and
//	          last body rows are reserved for "  ..." (blank when nothing is cut
//	          off on that side) and items sit between them.
//
// Seen to fail: before the fix a list scrolled to the middle rendered Height+1
// rows, and a search matching 1 item or none rendered 2.

type capItem struct{ id string }

func (i capItem) GetID() string          { return i.id }
func (i capItem) GetLabel() string       { return i.id }
func (i capItem) GetDescription() string { return "" }

func capItems(n int) []components.CapabilityItem {
	items := make([]components.CapabilityItem, n)
	for i := range items {
		items[i] = capItem{id: fmt.Sprintf("item-%02d", i)}
	}
	return items
}

func newSizedCapList(n, height int) *components.CapabilityList {
	l := components.NewCapabilityList(components.CapabilityListConfig{Title: "Skills", SourceHint: "hint", MaxVisible: 5})
	l.SetItems(capItems(n))
	l.SetSize(40, height)
	l.SetFocused(true)
	return l
}

func capLines(l *components.CapabilityList) []string {
	return strings.Split(ansi.Strip(l.View()), "\n")
}

// capQueries are a search that matches nothing special (none typed), one item,
// several, and nothing.
const ellipsisRow = "  ..."

var capQueries = []string{"", "item-07", "item-1", "zzz"} //nolint:gochecknoglobals // shared test fixture

// forEveryCapState calls fn for each cursor position of each query.
func forEveryCapState(l *components.CapabilityList, fn func(label string)) {
	for _, q := range capQueries {
		l.SetSearchQuery(q)
		n := len(l.FilteredItems())
		for pos := range max(n, 1) {
			l.SetCursor(pos)
			fn(fmt.Sprintf("query=%q cursor=%d", q, pos))
		}
	}
}

func TestCapabilityListRendersExactlyItsHeight(t *testing.T) {
	for height := 2; height <= 12; height++ {
		slots := height - 1
		for _, n := range []int{0, 1, slots, slots + 1, 30} {
			l := newSizedCapList(n, height)
			forEveryCapState(l, func(state string) {
				if got := len(capLines(l)); got != height {
					t.Errorf("height=%d items=%d %s: rendered %d lines, want %d:\n%s",
						height, n, state, got, height, ansi.Strip(l.View()))
				}
			})
		}
	}
}

func isCapItemRow(s string) bool {
	return strings.HasPrefix(strings.TrimLeft(s, "> "), "[") || strings.Contains(s, "(no matching items)")
}

// assertRowsInPlace checks one rendered state of a list of the given height:
// the header, the first item row, the indicator rows and the padding.
func assertRowsInPlace(t *testing.T, lines []string, height, firstItem int, overflow bool, label string) {
	t.Helper()
	if len(lines) != height {
		t.Fatalf("%s: %d lines, want %d", label, len(lines), height)
	}
	if !strings.HasPrefix(lines[0], "Skills (hint)") {
		t.Errorf("%s: row 0 is %q, want the header", label, lines[0])
	}
	if !isCapItemRow(lines[firstItem]) {
		t.Errorf("%s: row %d is %q, want the first item row:\n%s",
			label, firstItem, lines[firstItem], strings.Join(lines, "\n"))
	}
	if overflow {
		for _, row := range []int{1, height - 1} {
			if lines[row] != "" && lines[row] != ellipsisRow {
				t.Errorf("%s: row %d is %q, want blank or %q", label, row, lines[row], ellipsisRow)
			}
		}
	}
	for _, line := range lines[firstItem:] {
		if strings.TrimSpace(line) == "" && line != "" {
			t.Errorf("%s: padding line %q is not empty", label, line)
		}
	}
}

// assertIndicators checks that the "..." rows say which side has more.
func assertIndicators(t *testing.T, l *components.CapabilityList, n, height int) {
	t.Helper()
	l.SetSearchQuery("")
	for _, c := range []struct {
		name       string
		cursor     int
		top, final string
	}{
		{"top", 0, "", ellipsisRow},
		{"middle", n / 2, ellipsisRow, ellipsisRow},
		{"bottom", n - 1, ellipsisRow, ""},
	} {
		l.SetCursor(c.cursor)
		if lines := capLines(l); lines[1] != c.top || lines[height-1] != c.final {
			t.Errorf("height=%d items=%d %s: rows 1 and last are %q and %q, want %q and %q",
				height, n, c.name, lines[1], lines[height-1], c.top, c.final)
		}
	}
}

func TestCapabilityListItemsKeepTheirRows(t *testing.T) {
	for height := 4; height <= 12; height++ {
		slots := height - 1
		for _, n := range []int{2, slots, slots + 1, 30} {
			l := newSizedCapList(n, height)
			overflow := n > slots && slots >= 3
			firstItem := 1
			if overflow {
				firstItem = 2
			}
			forEveryCapState(l, func(state string) {
				assertRowsInPlace(t, capLines(l), height, firstItem, overflow, fmt.Sprintf("height=%d items=%d %s", height, n, state))
			})
			if overflow {
				assertIndicators(t, l, n, height)
			}
		}
	}
}

// TestCapabilityListScrollsThroughKeys drives the list the way the editor does
// and checks that the focused row is always rendered.
func TestCapabilityListScrollsThroughKeys(t *testing.T) {
	l := newSizedCapList(20, 7)
	for range 20 {
		plain := ansi.Strip(l.View())
		if !strings.Contains(plain, "> [") {
			t.Fatalf("cursor %d is not on screen:\n%s", l.Cursor(), plain)
		}
		l.Update(bubbletea.KeyPressMsg{Code: bubbletea.KeyDown})
	}
}

// TestCapabilityListUnsizedStillRenders: a list built before SetSize runs keeps
// the old, content-driven layout.
func TestCapabilityListUnsizedStillRenders(t *testing.T) {
	l := components.NewCapabilityList(components.CapabilityListConfig{Title: "Skills", MaxVisible: 3})
	l.SetItems(capItems(10))
	plain := ansi.Strip(l.View())
	if got := strings.Count(plain, "[ ]"); got != 3 {
		t.Errorf("unsized list shows %d items, want its MaxVisible (3):\n%s", got, plain)
	}
}
