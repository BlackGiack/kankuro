package app_test

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/BlackGiack/kankuro/internal/app"
	"github.com/BlackGiack/kankuro/internal/ui/uitest"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

// Contract (impl_plan_5.md B3): below views.MinTerminalWidth x MinTerminalHeight
// the coordinator's View() is a "Terminal too small" notice that is EXACT: every
// line is the terminal's width and there are exactly as many lines as the
// terminal has rows, at every size from 1x1 up to just under the minimum. Every
// keypress except quit is dropped, so a blind keypress cannot launch or delete
// anything, and the notice gives way to the view as soon as the terminal grows.
//
// Seen to fail: the shape sweep against the unguarded coordinator (the dashboard
// renders a different size and its preview width goes negative under ~62
// columns); the key test with the guard's early return removed (n opens the card
// editor).

func newSizedCoordinator(t *testing.T, w, h int) *app.Coordinator {
	t.Helper()
	coord := newBannerCoordinator(t)
	coord.Update(tea.WindowSizeMsg{Width: w, Height: h})
	return coord
}

func TestCoordinatorTooSmallScreenIsExact(t *testing.T) {
	coord := newBannerCoordinator(t)
	for w := 1; w <= views.MinTerminalWidth+5; w++ {
		for h := 1; h <= views.MinTerminalHeight+3; h++ {
			coord.Update(tea.WindowSizeMsg{Width: w, Height: h})
			if w >= views.MinTerminalWidth && h >= views.MinTerminalHeight {
				if strings.Contains(coord.View(), "Terminal too small") {
					t.Fatalf("%dx%d is large enough but shows the notice", w, h)
				}
				continue
			}
			out := coord.View()
			label := fmt.Sprintf("%dx%d", w, h)
			uitest.AssertExactRect(t, out, w, h, label)
			if h >= 2 && w >= 21 && !strings.Contains(ansi.Strip(out), "have") {
				t.Errorf("%s: the notice does not say the terminal's size:\n%s", label, ansi.Strip(out))
			}
		}
	}
}

func TestCoordinatorTooSmallScreenText(t *testing.T) {
	coord := newSizedCoordinator(t, 60, 20)
	if got := ansi.Strip(coord.View()); !strings.Contains(got, "Terminal too small: need 80×24, have 60×20") {
		t.Errorf("notice text wrong:\n%s", got)
	}
}

func TestCoordinatorDropsKeysWhenTooSmall(t *testing.T) {
	for _, size := range [][2]int{{79, 24}, {80, 23}, {40, 10}} {
		coord := newSizedCoordinator(t, size[0], size[1])
		before := coord.CurrentView()
		for _, key := range []tea.KeyPressMsg{
			{Code: 0, Text: "n"}, {Code: tea.KeyEnter}, {Code: tea.KeyTab}, {Code: tea.KeyBackspace}, {Code: 0, Text: "c"},
		} {
			if _, cmd := coord.Update(key); cmd != nil {
				t.Errorf("%v: key %v returned a command while too small", size, key)
			}
			if coord.CurrentView() != before {
				t.Fatalf("%v: key %v changed the view to %v", size, key, coord.CurrentView())
			}
		}
	}
}

func TestCoordinatorTooSmallStillQuits(t *testing.T) {
	for _, key := range []tea.KeyPressMsg{{Code: 0, Text: "q"}, {Code: 'c', Mod: tea.ModCtrl}} {
		coord := newSizedCoordinator(t, 40, 10)
		_, cmd := coord.Update(key)
		if cmd == nil {
			t.Fatalf("key %v returned no command", key)
		}
		if _, ok := cmd().(tea.QuitMsg); !ok {
			t.Errorf("key %v did not quit", key)
		}
	}
}

// TestCoordinatorRecoversWhenTerminalGrows: at the minimum size the notice is
// gone and the dashboard accepts keys again ("n" asks for a new card).
func TestCoordinatorRecoversWhenTerminalGrows(t *testing.T) {
	coord := newSizedCoordinator(t, 60, 20)
	if _, cmd := coord.Update(tea.KeyPressMsg{Code: 0, Text: "n"}); cmd != nil {
		t.Fatal("a key was accepted while too small")
	}
	coord.Update(tea.WindowSizeMsg{Width: views.MinTerminalWidth, Height: views.MinTerminalHeight})
	if strings.Contains(coord.View(), "Terminal too small") {
		t.Fatal("still showing the notice at the minimum size")
	}
	if _, cmd := coord.Update(tea.KeyPressMsg{Code: 0, Text: "n"}); cmd == nil {
		t.Error("the dashboard dropped a key at the minimum size")
	}
}

// TestCoordinatorDashboardFillsTheTerminal: through the coordinator, at and
// above the minimum size, the dashboard's render is exactly the terminal's rows
// - the coordinator's reserve (views.CoordinatorRows) and the dashboard's use of
// it must agree. Seen to fail with one row added to the dashboard's pane height.
func TestCoordinatorDashboardFillsTheTerminal(t *testing.T) {
	for _, size := range [][2]int{{80, 24}, {81, 25}, {120, 40}, {200, 60}} {
		coord := newSizedCoordinator(t, size[0], size[1])
		uitest.AssertWithin(t, coord.View(), size[0], size[1], fmt.Sprintf("%dx%d", size[0], size[1]))
		if got := strings.Count(coord.View(), "\n") + 1; got != size[1] {
			t.Errorf("%dx%d: dashboard is %d rows, want %d", size[0], size[1], got, size[1])
		}
	}
}
