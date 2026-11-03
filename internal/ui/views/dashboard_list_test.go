package views_test

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/uitest"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

func TestDashboardListShowsDeletionPreparationAndStatusErrors(t *testing.T) {
	tests := []struct {
		name    string
		state   string
		key     tea.KeyPressMsg
		failure tea.Msg
		success tea.Msg
	}{
		{"preflight", config.DashboardActive, tea.KeyPressMsg{Code: tea.KeyDelete},
			messages.DashboardDeletionPreparedMsg{DashboardID: "a", Err: errors.New("cannot prepare\nreason \x1b]0;unsafe\x07failed")},
			messages.DashboardDeletionPreparedMsg{DashboardID: "a"}},
		{"status", config.DashboardDeleting, tea.KeyPressMsg{Code: tea.KeyEnter},
			messages.DashboardDeletionStatusMsg{DashboardID: "a", Err: errors.New("cannot inspect\nreason \x1b]0;unsafe\x07failed")},
			messages.DashboardDeletionStatusMsg{DashboardID: "a"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			v := views.NewDashboardList(views.DashboardListCallbacks{
				PreflightDelete: func(string) tea.Cmd { return nil },
				InspectDelete:   func(string) tea.Cmd { return nil },
			})
			v.SetDashboards([]config.Dashboard{{ID: "a", Name: "A", State: tc.state}})
			for _, size := range []struct{ width, height int }{{80, 21}, {120, 37}, {160, 42}} {
				v.SetSize(size.width, size.height)
				v.Update(tc.key)
				v.Update(tc.failure)
				assertDashboardDeletionErrorVisible(t, v.View(), size.width, size.height)
				v.Update(tc.key)
				v.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
				if strings.Contains(v.View(), "reason failed") {
					t.Fatal("cancelled retry retained the old error")
				}
				v.Update(tc.key)
				v.Update(tc.success)
				v.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
				if strings.Contains(v.View(), "reason failed") {
					t.Fatal("successful retry retained the old error")
				}
			}
		})
	}
}

func assertDashboardDeletionErrorVisible(t *testing.T, view string, width, height int) {
	t.Helper()
	if !strings.Contains(view, "reason failed") || strings.Contains(view, "\x1b]0;unsafe") {
		t.Fatalf("error not visible or unsafe: %q", view)
	}
	if !strings.Contains(view, "q: quit") {
		t.Fatalf("footer missing after error: %q", view)
	}
	uitest.AssertWithin(t, view, width, height, "dashboard deletion error")
}

func TestDashboardListBoundsLongDeletionErrorWhenListBecomesEmpty(t *testing.T) {
	v := views.NewDashboardList(views.DashboardListCallbacks{PreflightDelete: func(string) tea.Cmd { return nil }})
	v.SetSize(80, 21)
	v.SetDashboards([]config.Dashboard{{ID: "a", Name: "A", State: config.DashboardActive}})
	v.Update(tea.KeyPressMsg{Code: tea.KeyDelete})
	v.Update(messages.DashboardDeletionPreparedMsg{DashboardID: "a", Err: errors.New("cleanup: " + strings.Repeat("界", 200))})
	v.SetDashboards(nil)
	view := v.View()
	if !strings.Contains(view, "cleanup:") || !strings.Contains(view, "No dashboards") {
		t.Fatalf("error or empty state missing: %q", view)
	}
	uitest.AssertWithin(t, view, 80, 21, "empty dashboard list with long error")
}

func TestDashboardListRequiresForeignDeletionConfirmation(t *testing.T) {
	var preflightID, confirmedID, confirmedToken string
	var confirmedForeign bool
	v := views.NewDashboardList(views.DashboardListCallbacks{
		PreflightDelete: func(id string) tea.Cmd { preflightID = id; return nil },
		ConfirmDelete: func(id, token string, foreign bool) tea.Cmd {
			confirmedID, confirmedToken, confirmedForeign = id, token, foreign
			return nil
		},
	})
	v.SetSize(80, 21)
	v.SetDashboards([]config.Dashboard{{ID: "a", Name: "Factory A", State: config.DashboardActive}})
	v.Update(tea.KeyPressMsg{Code: tea.KeyDelete})
	if preflightID != "a" {
		t.Fatalf("preflight ID = %q", preflightID)
	}
	prepared := messages.DashboardDeletionPreparedMsg{
		DashboardID:                 "a",
		Token:                       "token",
		Summary:                     []string{"Owned card: shared"},
		RequiresForeignConfirmation: true,
	}
	v.Update(prepared)
	v.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if confirmedID != "" {
		t.Fatal("foreign impact was confirmed without second prompt")
	}
	v.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if confirmedID != "" {
		t.Fatal("cancel performed deletion")
	}
	v.Update(tea.KeyPressMsg{Code: tea.KeyDelete})
	v.Update(prepared)
	v.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	v.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if confirmedID != "a" || confirmedToken != "token" || !confirmedForeign {
		t.Fatalf("confirmation = %q %q %v", confirmedID, confirmedToken, confirmedForeign)
	}
}

func TestDashboardListFitsAndSanitizesLongContent(t *testing.T) {
	v := views.NewDashboardList(views.DashboardListCallbacks{})
	for _, size := range []struct{ width, height int }{{80, 21}, {120, 37}, {160, 42}} {
		v.SetSize(size.width, size.height)
		v.SetDashboards(
			[]config.Dashboard{
				{
					ID:          strings.Repeat("id", 200),
					Name:        strings.Repeat("界", 120) + "\x1b[31m",
					Description: strings.Repeat("description ", 80),
					State:       config.DashboardActive,
				},
			},
		)
		uitest.AssertWithin(t, v.View(), size.width, size.height, "dashboard list")
		if strings.Contains(v.View(), "\x1b[31m") {
			t.Fatal("dashboard name retained unsafe escape")
		}
	}
}

func TestDashboardListOpensSelectedDashboard(t *testing.T) {
	var opened string
	v := views.NewDashboardList(views.DashboardListCallbacks{Open: func(d config.Dashboard) tea.Cmd { opened = d.ID; return nil }})
	v.SetSize(80, 21)
	v.SetDashboards(
		[]config.Dashboard{{ID: "a", Name: "A", State: config.DashboardActive}, {ID: "b", Name: "B", State: config.DashboardActive}},
	)
	v.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	v.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if opened != "b" {
		t.Fatalf("opened %q", opened)
	}
}

func TestDashboardListCreatesNamedDashboard(t *testing.T) {
	var savedName, savedDescription string
	v := views.NewDashboardList(views.DashboardListCallbacks{Save: func(_ string, name, description string) tea.Cmd {
		savedName, savedDescription = name, description
		return nil
	}})
	v.SetSize(80, 21)
	v.Update(tea.KeyPressMsg{Text: "n", Code: 'n'})
	v.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if savedName != "" || !strings.Contains(v.View(), "name is required") {
		t.Fatal("empty dashboard name was accepted")
	}
	v.Update(tea.KeyPressMsg{Text: "Factory", Code: 'F'})
	v.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	v.Update(tea.KeyPressMsg{Text: "Production", Code: 'P'})
	v.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if savedName != "Factory" || savedDescription != "Production" {
		t.Fatalf("saved %q, %q", savedName, savedDescription)
	}
}

func TestDashboardListShowsRetryableDeletionStatus(t *testing.T) {
	for _, key := range []rune{tea.KeyDelete, tea.KeyBackspace} {
		var inspected, retried string
		v := views.NewDashboardList(views.DashboardListCallbacks{
			InspectDelete: func(id string) tea.Cmd { inspected = id; return nil },
			RetryDelete:   func(id string) tea.Cmd { retried = id; return nil },
		})
		v.SetSize(80, 21)
		v.SetDashboards([]config.Dashboard{{ID: "a", Name: "A", State: config.DashboardDeleting}})
		v.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		if inspected != "a" {
			t.Fatalf("inspected %q", inspected)
		}
		v.Update(messages.DashboardDeletionStatusMsg{DashboardID: "a", Summary: []string{"Session s: removed=false trust cleanup failed"}})
		if !strings.Contains(v.View(), "trust cleanup failed") {
			t.Fatal("per-item failure not shown")
		}
		v.Update(tea.KeyPressMsg{Code: key})
		if retried != "a" {
			t.Fatalf("retried %q", retried)
		}
	}
}

func TestDashboardListRetriesDeletionWithTheDeleteKey(t *testing.T) {
	for _, key := range []rune{tea.KeyDelete, tea.KeyBackspace} {
		var preflights, retries int
		callbacks := views.DashboardListCallbacks{
			PreflightDelete: func(id string) tea.Cmd {
				if id != "a" {
					t.Fatalf("preflight targeted %q", id)
				}
				preflights++
				return nil
			},
			RetryDelete: func(id string) tea.Cmd {
				if id != "a" {
					t.Fatalf("retry targeted %q", id)
				}
				retries++
				return nil
			},
		}
		v := views.NewDashboardList(callbacks)
		v.SetSize(80, 24)
		v.SetDashboards([]config.Dashboard{{ID: "a", Name: "A", State: config.DashboardActive}})
		v.Update(tea.KeyPressMsg{Code: key})
		v.Update(messages.DashboardDeletionPreparedMsg{DashboardID: "a", Err: errors.New("prepare failed")})
		v.Update(tea.KeyPressMsg{Code: key})
		if preflights != 2 {
			t.Fatalf("delete did not retry failed preparation: %d calls", preflights)
		}
		v.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
		v.SetDashboards([]config.Dashboard{{ID: "a", Name: "A", State: config.DashboardDeleting}})
		v.Update(tea.KeyPressMsg{Text: "r", Code: 'r'})
		if retries != 0 {
			t.Fatal("obsolete retry binding remains active")
		}
		v.Update(tea.KeyPressMsg{Code: key})
		if retries != 1 || preflights != 2 {
			t.Fatalf("delete did not resume deletion: retries=%d preflights=%d", retries, preflights)
		}
		if !strings.Contains(v.View(), "Delete: retry") || strings.Contains(v.View(), "r: retry") {
			t.Fatal("retry legend does not match the delete action")
		}
	}
}
