package app_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/app"
)

func TestDashboardListRequiresExplicitQuit(t *testing.T) {
	coord := newBannerCoordinator(t)
	_, cmd := coord.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	for _, msg := range collectMsgs(t, cmd) {
		coord.Update(msg)
	}
	if coord.CurrentView() != app.ViewDashboardList {
		t.Fatal("Escape did not return to the dashboard list")
	}
	for _, size := range [][2]int{{120, 40}, {40, 10}, {120, 40}} {
		coord.Update(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
		for _, key := range []tea.KeyPressMsg{
			{Code: tea.KeyEsc}, {Code: tea.KeyEsc}, {Code: 'c', Mod: tea.ModCtrl},
		} {
			_, cmd = coord.Update(key)
			for _, msg := range collectMsgs(t, cmd) {
				if _, quit := msg.(tea.QuitMsg); quit {
					t.Fatalf("key %v quit the dashboard list", key)
				}
				coord.Update(msg)
			}
			if coord.CurrentView() != app.ViewDashboardList {
				t.Fatalf("key %v left the dashboard list", key)
			}
		}
	}
	_, cmd = coord.Update(tea.KeyPressMsg{Text: "q", Code: 'q'})
	if cmd == nil {
		t.Fatal("q returned no quit command")
	}
	if _, quit := cmd().(tea.QuitMsg); !quit {
		t.Fatal("q did not quit the dashboard list")
	}
}
