package views_test

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/ui/uitest"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

func TestDashboardCardForeignConfirmationCanBeCancelled(t *testing.T) {
	var confirmed bool
	dashboard := views.NewDashboard(
		views.DashboardCallbacks{OnConfirmCardDeletion: func(_, _ string, foreign bool) tea.Cmd { confirmed = foreign; return nil }},
	)
	dashboard.SetSize(80, 21)
	dashboard.ShowCardDeletionImpact("card", "token", []string{"Foreign dashboard: session"})
	uitest.AssertWithin(t, dashboard.View(), 80, 21, "card deletion impact")
	if !strings.Contains(dashboard.View(), "globally") {
		t.Fatal("global deletion not disclosed")
	}
	dashboard.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if confirmed {
		t.Fatal("cancel confirmed deletion")
	}
	dashboard.ShowCardDeletionImpact("card", "token", []string{"Foreign dashboard: session"})
	dashboard.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if !confirmed {
		t.Fatal("explicit confirmation not dispatched")
	}
}
