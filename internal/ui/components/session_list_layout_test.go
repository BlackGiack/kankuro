package components_test

import (
	"fmt"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/BlackGiack/kankuro/internal/ui/components"
	"github.com/BlackGiack/kankuro/internal/ui/uitest"
)

// Contract (step 0 of impl_plan_2.md): SessionList.View() is a bare
// bubbles/list render, so it is BOUNDED, not exact: no line wider than the
// width it was given (the dashboard pane pads it) and exactly the height it
// was given or less. Observed: it renders exactly `height` lines because the
// list pads its page, but only the "at most" half is part of the contract.
//
// Region contract: every item occupies SessionDelegateHeight rows plus one
// spacing row. Its third row (index 2) is the status line "<status> pid=N",
// so a status row must sit at 5*k+2 for every item k on the page and nowhere
// else - an item whose name or card wrapped would push it out of place.
//
// Seen to fail: the two tests that were skipped here (working dir sliced by
// byte, tabs in names) failed against the code before impl_plan_5.md fixed them
// ("Found, not fixed" #2 and #4 in impl_plan_2.md); the sweeps were seen red
// before those cases were split out, and again with the list's help line
// re-enabled (they then measure the raw View(), trailing padding included); the
// item-row test was seen red by shifting the expected status row by one. The
// sweeps use every adversarial name, tabs included.

const sessionItemRows = components.SessionDelegateHeight + 1 // item rows + spacing row

// sessionListMinHeight is the shortest list the dashboard hands its sessions
// pane on a 24-row terminal: the coordinator gives the dashboard 24-4 = 20,
// the pane is 21 rows, minus its border (2) and title block (2). Below ~9 rows
// the list's own chrome (pagination) plus one item no longer fits;
// that is out of range, since the dashboard is never that short.
const sessionListMinHeight = 17

// sessionListMinWidth is the narrowest list the dashboard hands its sessions
// pane on the smallest terminal it supports (80 columns, see
// views/dashboard_layout_test.go): 80/4 = 20 pane columns minus the two border
// columns. TestSessionListFitsEveryWidth covers narrower lists too.
const sessionListMinWidth = 18

func sweepNames() []string { return uitest.AdversarialNames }

func sessionFixtures(n int, names []string) ([]session.Snapshot, []config.IDCard) {
	faces := []string{"", "happy", "focused", "thinking", "neutral"}
	statuses := []session.Status{session.StatusRunning, session.StatusWaiting, session.StatusTerminated}
	dirs := []string{
		"/home/user/projects/a-rather-long-directory-name/with/many/segments",
		"/home/用户/项目/日本語のディレクトリ/とても長い名前",
		"/tmp/x",
		"",
		"/home/user/🚀/🚀/🚀/🚀/🚀/🚀/🚀/🚀",
	}
	var cards []config.IDCard
	var sessions []session.Snapshot
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := range n {
		cardID := fmt.Sprintf("card-%d", i)
		if i%7 != 6 { // every seventh session's card was deleted
			cards = append(cards, config.IDCard{
				ID:        cardID,
				Name:      names[(i+1)%len(names)],
				FaceStyle: faces[i%len(faces)],
			})
		}
		sessions = append(sessions, session.Snapshot{
			ID:              fmt.Sprintf("session-%d-%s", i, strings.Repeat("0123456789", 4)),
			CardID:          cardID,
			Name:            names[i%len(names)],
			PID:             1000 + i,
			WorkingDir:      dirs[i%len(dirs)],
			UseTmpWorkspace: i%3 == 0,
			StartedAt:       base.Add(time.Duration(i) * time.Minute),
			Status:          statuses[i%len(statuses)],
			Pinged:          i%4 == 0,
		})
	}
	return sessions, cards
}

func TestSessionListStaysWithinItsBox(t *testing.T) {
	for _, count := range []int{0, 1, 20} {
		for _, query := range []string{"", "s"} {
			sessions, cards := sessionFixtures(count, sweepNames())
			for w := sessionListMinWidth; w <= 80; w += 3 {
				for h := sessionListMinHeight; h <= 30; h += 3 {
					sl := components.NewSessionList(components.SessionListCallbacks{})
					sl.SetSize(w, h)
					sl.SetSessions(sessions, cards)
					sl.SetSearchQuery(query)
					label := fmt.Sprintf("sessions=%d query=%q %dx%d", count, query, w, h)
					uitest.AssertWithin(t, sl.View(), w, h, label)
				}
			}
		}
	}
}

func TestSessionListRegistrationDetailsStayWithinTheirRows(t *testing.T) {
	sessions, cards := sessionFixtures(1, []string{strings.Repeat("n", 200)})
	sessions[0].Registration = session.RegistrationPending
	sessions[0].RegistrationErr = strings.Repeat("registration failure ", 20)
	sessions[0].ActiveConversationID = strings.Repeat("native-id-", 20)
	sessions[0].Conversations = map[string][]session.ConversationBinding{
		sessions[0].CardID: {{ID: sessions[0].ActiveConversationID, Name: strings.Repeat("conversation name ", 10), Driver: "codex"}},
	}
	for _, width := range []int{sessionListMinWidth, 23, 40, 80} {
		list := components.NewSessionList(components.SessionListCallbacks{})
		list.SetSize(width, sessionListMinHeight)
		list.SetSessions(sessions, cards)
		uitest.AssertWithin(
			t,
			list.View(),
			width,
			sessionListMinHeight,
			fmt.Sprintf("registration details %dx%d", width, sessionListMinHeight),
		)
		assertItemRows(t, list, sessionListMinHeight, fmt.Sprintf("registration rows %d", width))
	}
}

// TestSessionListKeepsItemRowsInPlace checks each item against its own region.
// Every name from AdversarialNames gets its own sweep, since a single wrapped
// or overlong title shifts every row below it.
func TestSessionListKeepsItemRowsInPlace(t *testing.T) {
	for i, name := range sweepNames() {
		sessions, cards := sessionFixtures(12, []string{name})
		for _, w := range []int{sessionListMinWidth, 23, 29, 40, 80} {
			for _, h := range []int{17, 24, 30} {
				sl := components.NewSessionList(components.SessionListCallbacks{})
				sl.SetSize(w, h)
				sl.SetSessions(sessions, cards)
				label := fmt.Sprintf("name#%d %dx%d", i, w, h)
				assertItemRows(t, sl, h, label)
			}
		}
	}
}

func assertItemRows(t *testing.T, sl *components.SessionList, h int, label string) {
	t.Helper()
	out := sl.View()
	rows := strings.Split(ansi.Strip(out), "\n")
	onPage := sl.List().Paginator.ItemsOnPage(len(sl.List().Items()))
	if onPage*sessionItemRows-1 > h {
		t.Errorf("%s: %d items of %d rows cannot fit in %d:\n%s", label, onPage, sessionItemRows, h, ansi.Strip(out))
		return
	}
	for k := range onPage {
		row := sessionItemRows*k + 2
		if row >= len(rows) || !strings.Contains(rows[row], "pid=") && !strings.Contains(rows[row], "wai") &&
			!strings.Contains(rows[row], "run") && !strings.Contains(rows[row], "ter") {
			t.Errorf("%s: item %d has no status row at line %d:\n%s", label, k, row, ansi.Strip(out))
			return
		}
	}
	for i, r := range rows[:min(len(rows), onPage*sessionItemRows)] {
		if i%sessionItemRows == sessionItemRows-1 && strings.TrimSpace(r) != "" {
			t.Errorf("%s: spacing row %d is not blank %q:\n%s", label, i, r, ansi.Strip(out))
			return
		}
	}
}

// TestSessionListEmptyAndFilteredAway covers 0 items and a search that matches
// nothing: the filter changes what is drawn, and the result must still fit.
func TestSessionListEmptyAndFilteredAway(t *testing.T) {
	sessions, cards := sessionFixtures(5, []string{"alpha"})
	sl := components.NewSessionList(components.SessionListCallbacks{})
	sl.SetSize(30, sessionListMinHeight)
	sl.SetSessions(sessions, cards)
	sl.SetSearchQuery("zzz-matches-nothing")
	uitest.AssertWithin(t, sl.View(), 30, sessionListMinHeight, "filtered away")
	if len(sl.List().Items()) != 0 {
		t.Fatalf("expected the query to match nothing, got %d items", len(sl.List().Items()))
	}
}

// TestSessionListDescriptionIsValidUTF8 guards the path trimming in
// Description(): it slices a byte offset, so a path with multibyte runes near
// the cut produced invalid UTF-8, which terminals draw as replacement boxes
// and which breaks width measurement.
func TestSessionListDescriptionIsValidUTF8(t *testing.T) {
	for _, dir := range []string{
		"/home/用户/项目/日本語のディレクトリ/とても長い名前",
		"/home/user/🚀/🚀/🚀/🚀/🚀/🚀/🚀/🚀",
		"/ééééééééééééééééééééééééééééé",
	} {
		item := components.SessionListItem{Session: session.Snapshot{ID: "s", WorkingDir: dir}}
		desc := item.Description()
		if !utf8.ValidString(desc) {
			t.Errorf("working dir %q: Description() is not valid UTF-8: %q", dir, desc)
		}
	}
}

// TestSessionListFitsEveryWidth is the regression test for the list's built-in
// help line ("up / down / q quit / ? more"), which was wider than the list at
// list widths 17-20, 26-28 and 35-36 (and below 17 for an empty list) and made
// the dashboard's sessions pane wrap it and grow two rows - on a stock 80- and
// 120-column terminal. It sweeps every width, narrower than the dashboard ever
// asks for, so re-enabling the help (or any other fixed-width row) fails here.
// Seen to fail: with SetShowHelp(false) removed from NewSessionList.
func TestSessionListFitsEveryWidth(t *testing.T) {
	for _, count := range []int{0, 3, 20} {
		sessions, cards := sessionFixtures(count, []string{"alpha"})
		for w := 10; w <= 80; w++ {
			sl := components.NewSessionList(components.SessionListCallbacks{})
			sl.SetSize(w, sessionListMinHeight)
			sl.SetSessions(sessions, cards)
			uitest.AssertWithin(t, sl.View(), w, sessionListMinHeight, fmt.Sprintf("sessions=%d width=%d", count, w))
		}
	}
}

// TestSessionListNameWithTabsFitsWidth: lipgloss expands a tab to four spaces
// after the delegate has measured and truncated the title, so a name with tabs
// ends up wider than the list.
func TestSessionListNameWithTabsFitsWidth(t *testing.T) {
	sessions, cards := sessionFixtures(3, []string{"name\twith\ttabs and more text"})
	sl := components.NewSessionList(components.SessionListCallbacks{})
	sl.SetSize(30, sessionListMinHeight)
	sl.SetSessions(sessions, cards)
	uitest.AssertWithin(t, sl.View(), 30, sessionListMinHeight, "tab name")
}
