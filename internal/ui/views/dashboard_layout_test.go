package views_test

import (
	"fmt"
	"image/color"
	"strings"
	"sync"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"kankuro/internal/config"
	"kankuro/internal/session"
	"kankuro/internal/ui/messages"
	"kankuro/internal/ui/styles"
	"kankuro/internal/ui/uitest"
	"kankuro/internal/ui/views"
)

// Contract (step 0 of impl_plan_2.md). The coordinator sizes the dashboard to
// (terminal width, terminal rows - 4) - app.views.CoordinatorRows - and
// Dashboard.SetSize's comment promises the full render then fills the terminal:
//
//	row 0               header   "K A N K U R O"   BOUNDED (17 wide; not padded)
//	row 1               blank (header's bottom margin)
//	rows 2 .. T-2       body: three bordered panes side by side, EXACT: every
//	                    row is exactly W wide, there are exactly T-3 = h+1 rows
//	row T-1             footer key legend            BOUNDED
//
// so the whole render is exactly T = h+4 lines (dashboardChrome below pins the
// +4). "Exact" is the dashboard's contract only for the body: the header and
// footer are not padded to W, which is deliberate for the header and a bug for
// a footer wider than the terminal (TestDashboardFooterFitsWidth).
//
// Region contract: the panes keep their columns. With cards = cw, preview = pw
// and sessions = W-cw-pw (Dashboard.CardsWidth / PreviewWidth), a vertical
// border sits at columns cw-1, cw, cw+pw-1, cw+pw and W-1 on every body row,
// and the corners on the first and last body row. A pane that grows by one
// column moves its neighbour's border, which this catches even when the total
// width survives.
//
// Smallest supported terminal: 80x24 (views.MinTerminalWidth/Height). Below that
// the coordinator shows a "terminal too small" notice instead of any view
// (app/size_guard_test.go); the dashboard is still sized then, and only has to
// not go negative (TestDashboardSizeNeverNegative).
//
// Seen to fail: the shape/region sweeps were red at 80 and 120 columns against
// the old code because of the session list help line (fixed since, by
// SetShowHelp(false)); the delete-modal size, footer and long-query tests were
// red until impl_plan_5.md fixed them (they were skipped, with their causes in
// impl_plan_2.md). The sweeps were also seen red by widening the preview pane
// by one column (Width(d.previewWidth + 1)) and by shortening the body by one
// row; the focus test by swapping which pane gets the focused border.

const (
	// dashboardChrome is the rows around the panes: header (1) + its blank
	// margin row (1) + footer (1). With the panes' own h+1 rows that makes the
	// full render h+4 = terminal rows, (h + views.CoordinatorRows).
	dashboardChrome = 3
)

var (
	viewsBuiltinOnce sync.Once       //nolint:gochecknoglobals // test cache: building the repo writes files
	viewsBuiltin     []config.Puppet //nolint:gochecknoglobals // test cache, see viewsBuiltinOnce
)

func builtinPuppetsForViews(t testing.TB) []config.Puppet {
	t.Helper()
	viewsBuiltinOnce.Do(func() {
		repo, err := config.NewFileRepository(t.TempDir())
		if err != nil {
			t.Fatalf("new repo: %v", err)
		}
		viewsBuiltin, err = repo.ListPuppets()
		if err != nil {
			t.Fatalf("list puppets: %v", err)
		}
	})
	return viewsBuiltin
}

type dashboardData struct {
	sessions int
	cards    int
	preview  []byte
}

func dashboardFixtures(t testing.TB, d dashboardData) messages.DataLoadedMsg {
	t.Helper()
	names := uitest.AdversarialNames
	puppets := append(append([]config.Puppet(nil), builtinPuppetsForViews(t)...), uitest.UserPuppet(60, 20))
	statuses := []session.Status{session.StatusRunning, session.StatusWaiting, session.StatusTerminated}
	skillNames := []string{"architecture", "clean-code", "golang", "plan-writing", "日本語のスキル"}

	var skills []config.Skill
	for i, n := range skillNames {
		skills = append(skills, config.Skill{ID: fmt.Sprintf("skill-%d", i), Name: n})
	}
	var cards []config.IDCard
	for i := range d.cards {
		c := config.IDCard{
			ID:        fmt.Sprintf("card-%d", i),
			Name:      names[i%len(names)],
			Driver:    "claude",
			Model:     names[(i+3)%len(names)],
			Puppet:    puppets[i%len(puppets)].Name,
			FaceStyle: []string{"happy", "focused", "thinking", ""}[i%4],
		}
		for j := 0; j <= i%len(skills); j++ {
			c.SkillIDs = append(c.SkillIDs, skills[j].ID)
		}
		cards = append(cards, c)
	}
	var sessions []session.Snapshot
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := range d.sessions {
		sessions = append(sessions, session.Snapshot{
			ID:         fmt.Sprintf("session-%d-0123456789-abcdef", i),
			CardID:     fmt.Sprintf("card-%d", i%max(1, d.cards)),
			Name:       names[i%len(names)],
			PID:        2000 + i,
			WorkingDir: "/home/user/projects/a-rather-long-directory-name/with/segments",
			StartedAt:  base.Add(time.Duration(i) * time.Minute),
			Status:     statuses[i%len(statuses)],
		})
	}
	return messages.DataLoadedMsg{Cards: cards, Sessions: sessions, Skills: skills, Puppets: puppets}
}

var previewContentsForDashboard = [][]byte{ //nolint:gochecknoglobals // shared test fixtures
	[]byte("hello world\nthe last line\n"),
	[]byte(strings.Repeat("漢字", 120) + "\n" + strings.Repeat("🚀 ab ", 80) + "\nlast words\n"),
	[]byte(strings.Repeat("abcdefghij", 1000) + "\nlast words\n"),
	[]byte("\x1b[1;31mred\x1b[0m " + strings.Repeat("\x1b[32mgreen\x1b[0m ", 40) + "\nlast words\n\x1b]8;;https://example.com/never-ends"),
	nil,
}

func newDashboard(t testing.TB, w, rows int, d dashboardData) *views.Dashboard {
	t.Helper()
	dash := views.NewDashboard(views.DashboardCallbacks{})
	dash.SetSize(w, rows-views.CoordinatorRows)
	dash.SetData(dashboardFixtures(t, d))
	if len(d.preview) > 0 && dash.PreviewSessionID() != "" {
		dash.SetPreviewData(messages.PreviewDataMsg{SessionID: dash.PreviewSessionID(), Scrollback: d.preview})
	}
	return dash
}

// assertDashboardShape checks the whole contract.
func assertDashboardShape(t *testing.T, d *views.Dashboard, w, rows int, label string) {
	t.Helper()
	out := d.View()
	lines := strings.Split(out, "\n")
	if len(lines) != rows {
		t.Errorf("%s: %d lines, want %d (terminal rows):\n%s", label, len(lines), rows, ansi.Strip(out))
		return
	}
	header := lines[0]
	if ansi.StringWidth(header) > w || strings.TrimSpace(ansi.Strip(lines[1])) != "" {
		t.Errorf("%s: header rows malformed:\n%s", label, ansi.Strip(out))
	}
	if got := ansi.StringWidth(lines[rows-1]); got > w {
		t.Errorf("%s: footer is %d columns wide:\n%s", label, got, ansi.Strip(lines[rows-1]))
	}
	body := strings.Join(lines[2:rows-1], "\n")
	uitest.AssertExactRect(t, body, w, rows-dashboardChrome, label+" body")
	assertPaneBorders(t, body, d, w, rows-dashboardChrome, label)
}

func assertPaneBorders(t *testing.T, body string, d *views.Dashboard, w, bodyRows int, label string) {
	t.Helper()
	cw, pw := d.CardsWidth(), d.PreviewWidth()
	cols := []int{cw - 1, cw, cw + pw - 1, cw + pw, w - 1}
	for y := range bodyRows {
		for _, x := range cols {
			got := uitest.Region(body, x, y, 1, 1)[0]
			var want string
			switch y {
			case 0:
				want = map[int]string{cw - 1: "┐", cw: "╭", cw + pw - 1: "╮", cw + pw: "┌", w - 1: "┐"}[x]
			case bodyRows - 1:
				want = map[int]string{cw - 1: "┘", cw: "╰", cw + pw - 1: "╯", cw + pw: "└", w - 1: "┘"}[x]
			default:
				want = "│"
			}
			if got != want {
				t.Errorf("%s: body row %d column %d is %q, want pane border %q (cards=%d preview=%d):\n%s",
					label, y, x, got, want, cw, pw, body)
				return
			}
		}
	}
}

// eachSize calls fn for every width x terminal-rows combination.
func eachSize(widths, rowsList []int, fn func(w, rows int)) {
	for _, w := range widths {
		for _, rows := range rowsList {
			fn(w, rows)
		}
	}
}

func TestDashboardLayoutSweep(t *testing.T) {
	n := 0
	eachSize([]int{80, 81, 85, 90, 100, 112, 120, 130, 150, 161, 220}, []int{24, 60}, func(w, rows int) {
		for _, sessions := range []int{0, 1, 20} {
			for _, focus := range []views.Pane{views.PaneCards, views.PaneSessions} {
				data := dashboardData{
					sessions: sessions,
					cards:    []int{0, 3, 12}[n%3],
					preview:  previewContentsForDashboard[n%len(previewContentsForDashboard)],
				}
				n++
				d := newDashboard(t, w, rows, data)
				if focus == views.PaneSessions {
					d.TogglePane()
				}
				label := fmt.Sprintf("%dx%d sessions=%d cards=%d focus=%d case#%d", w, rows, sessions, data.cards, focus, n)
				assertDashboardShape(t, d, w, rows, label)
			}
		}
	})
}

// TestDashboardSearchStates: the search line replaces the blank spacer under
// each pane title, so it must not change the render's size or move a border.
func TestDashboardSearchStates(t *testing.T) {
	eachSize([]int{80, 90, 100, 120, 161}, []int{24, 40}, func(w, rows int) {
		for _, focus := range []views.Pane{views.PaneCards, views.PaneSessions} {
			for _, query := range []string{"a", "漢字", "zzzz", strings.Repeat("long query ", 10)} {
				assertSearchState(t, w, rows, focus, query)
			}
		}
	})
}

func assertSearchState(t *testing.T, w, rows int, focus views.Pane, query string) {
	t.Helper()
	d := newDashboard(t, w, rows, dashboardData{sessions: 8, cards: 6})
	if focus == views.PaneSessions {
		d.TogglePane()
	}
	d.Update(tea.KeyPressMsg{Code: 0, Text: "/"})
	for _, r := range query {
		d.Update(tea.KeyPressMsg{Code: 0, Text: string(r)})
	}
	label := fmt.Sprintf("%dx%d focus=%d searching %q", w, rows, focus, query)
	assertDashboardShape(t, d, w, rows, label)
	if focus == views.PaneCards && !strings.Contains(ansi.Strip(d.View()), "Search: ") {
		t.Errorf("%s: no search line in the cards pane title", label)
	}
	d.Update(tea.KeyPressMsg{Code: tea.KeyEnter}) // confirm: "Filter: ..." replaces "Search: ..."
	assertDashboardShape(t, d, w, rows, label+" confirmed")
}

// TestDashboardLongSearchQueryFits: the "Search: ..." / "Filter: ..." line under
// a pane title used to be written untruncated, so a query wider than the pane
// wrapped inside it and the dashboard grew by one to four rows (2 to 4 at 100
// columns with an 80-character query). The label stays and the head of the
// query is clipped: the end is where the user is typing.
func TestDashboardLongSearchQueryFits(t *testing.T) {
	for _, w := range []int{80, 100, 130} {
		for _, focus := range []views.Pane{views.PaneCards, views.PaneSessions} {
			d := newDashboard(t, w, 30, dashboardData{sessions: 3, cards: 3})
			if focus == views.PaneSessions {
				d.TogglePane()
			}
			d.Update(tea.KeyPressMsg{Code: 0, Text: "/"})
			for range 80 {
				d.Update(tea.KeyPressMsg{Code: 0, Text: "q"})
			}
			assertDashboardShape(t, d, w, 30, fmt.Sprintf("%d focus=%d", w, focus))
		}
	}
}

// TestDashboardFocusedBorderColour: the focused pane's border is drawn in the
// primary colour and the other outer pane's in grey, and Tab moves the focus.
// It reads a body row below the lists' items, where the only "│" runes are the
// outer panes' left and right borders (a selected list item draws its own
// primary-coloured bar higher up).
func TestDashboardFocusedBorderColour(t *testing.T) {
	bar := func(c color.Color) string { return lipgloss.NewStyle().Foreground(c).Render("│") }
	focused, plain := bar(styles.PrimaryColor), bar(styles.Gray)

	for _, rows := range []int{24, 40} {
		d := newDashboard(t, 100, rows, dashboardData{sessions: 2, cards: 2})
		for _, tc := range []struct {
			pane                views.Pane
			wantLeft, wantRight string
			description         string
		}{
			{views.PaneCards, focused, plain, "cards focused"},
			{views.PaneSessions, plain, focused, "sessions focused"},
		} {
			if d.FocusedPane() != tc.pane {
				d.TogglePane()
			}
			lines := strings.Split(d.View(), "\n")
			row := lines[len(lines)-5] // a blank row inside the panes, above the help line
			if !strings.HasPrefix(row, tc.wantLeft) || !strings.HasSuffix(row, tc.wantRight) {
				t.Errorf("rows=%d %s: cards border/sessions border colours wrong in %q", rows, tc.description, row)
			}
		}
	}
}

// TestDashboardDeleteModalFillsTheScreen: opening the delete modal replaces
// the dashboard with lipgloss.Place(d.Width, d.Height, ...), and d.Height is
// the terminal rows minus 4, so the screen shrinks by four rows when the modal
// opens and grows back when it closes.
func TestDashboardDeleteModalFillsTheScreen(t *testing.T) {
	for _, w := range []int{80, 100, 160} {
		for _, rows := range []int{24, 40} {
			d := newDashboard(t, w, rows, dashboardData{sessions: 2, cards: 2})
			d.Update(tea.KeyPressMsg{Code: tea.KeyBackspace}) // focus is on cards: delete-card modal
			if !d.DeleteModal().IsOpen() {
				t.Fatal("delete modal did not open")
			}
			uitest.AssertExactRect(t, d.View(), w, rows, fmt.Sprintf("%dx%d modal open", w, rows))
		}
	}
}

// TestDashboardDeleteModalIsCentredAndIntact: the overlay is exactly the
// screen's size, the modal's text is there and it sits in the vertical middle.
func TestDashboardDeleteModalIsCentredAndIntact(t *testing.T) {
	eachSize([]int{80, 100, 160}, []int{24, 40}, func(w, rows int) {
		d := newDashboard(t, w, rows, dashboardData{sessions: 2, cards: 2})
		d.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
		if !d.DeleteModal().IsOpen() {
			t.Fatal("delete modal did not open")
		}
		out := d.View()
		label := fmt.Sprintf("%dx%d modal open", w, rows)
		uitest.AssertExactRect(t, out, w, rows, label)
		if !strings.Contains(ansi.Strip(out), "Delete this ID card?") {
			t.Errorf("%s: modal content missing:\n%s", label, ansi.Strip(out))
		}
		if gapTop, gapBottom := verticalGaps(out); gapTop-gapBottom > 1 || gapBottom-gapTop > 1 {
			t.Errorf("%s: modal off-centre vertically (gap above %d, below %d)", label, gapTop, gapBottom)
		}
	})
}

// verticalGaps returns the blank rows above and below the non-blank content.
func verticalGaps(out string) (int, int) {
	rows := strings.Split(ansi.Strip(out), "\n")
	top, bottom := -1, -1
	for i, r := range rows {
		if strings.TrimSpace(r) == "" {
			continue
		}
		if top < 0 {
			top = i
		}
		bottom = i
	}
	return top, len(rows) - 1 - bottom
}

// TestDashboardFooterFitsWidth: the footer key legend used to be a fixed 150
// columns wide, so it was cut off (or wrapped) on anything narrower. It now
// shows the navigation group, the focused pane's group and the global group,
// and gives up groups, not characters, as the terminal narrows.
func TestDashboardFooterFitsWidth(t *testing.T) {
	for w := 1; w <= 200; w++ {
		for _, focus := range []views.Pane{views.PaneCards, views.PaneSessions} {
			d := newDashboard(t, w, 30, dashboardData{sessions: 1, cards: 1})
			if focus == views.PaneSessions {
				d.TogglePane()
			}
			lines := strings.Split(d.View(), "\n")
			if got := ansi.StringWidth(lines[len(lines)-1]); got > w {
				t.Errorf("width %d focus %d: footer is %d columns wide", w, focus, got)
			}
		}
	}
}

// TestDashboardFooterShowsFocusedPaneKeys: at every width from the 80-column
// minimum the footer carries the focused pane's keys, never the other pane's,
// and always q=Quit.
func TestDashboardFooterShowsFocusedPaneKeys(t *testing.T) {
	for w := views.MinTerminalWidth; w <= 200; w++ {
		for _, tc := range []struct {
			pane        views.Pane
			want, other string
		}{
			{views.PaneCards, "[Card] Enter=Launch", "[Session]"},
			{views.PaneSessions, "[Session] Enter=Attach/Relaunch", "[Card]"},
		} {
			d := newDashboard(t, w, 30, dashboardData{sessions: 1, cards: 1})
			if d.FocusedPane() != tc.pane {
				d.TogglePane()
			}
			lines := strings.Split(ansi.Strip(d.View()), "\n")
			footer := lines[len(lines)-1]
			if !strings.Contains(footer, tc.want) || strings.Contains(footer, tc.other) || !strings.Contains(footer, "q=Quit") {
				t.Errorf("width %d focus %d: footer %q, want %q, not %q, and q=Quit", w, tc.pane, footer, tc.want, tc.other)
			}
		}
	}
	// With room to spare, the whole legend of the focused pane is there.
	d := newDashboard(t, 200, 30, dashboardData{sessions: 1, cards: 1})
	lines := strings.Split(ansi.Strip(d.View()), "\n")
	footer := lines[len(lines)-1]
	for _, want := range []string{"[Nav] Tab", "[Card] Enter=Launch n=New e=Edit", "[Global] m=MCP w=Workspaces c=Settings q=Quit"} {
		if !strings.Contains(footer, want) {
			t.Errorf("wide footer %q is missing %q", footer, want)
		}
	}
}

// TestDashboardSizeNeverNegative: below the minimum terminal size the
// dashboard is not drawn (the coordinator shows a notice) but it is still
// sized; its panes must not get a negative width, which would panic the
// renderer.
func TestDashboardSizeNeverNegative(t *testing.T) {
	for w := 1; w < views.MinTerminalWidth; w++ {
		d := newDashboard(t, w, views.MinTerminalHeight, dashboardData{sessions: 1, cards: 1})
		if d.PreviewWidth() < 0 || d.CardsWidth() < 0 {
			t.Errorf("width %d: cards %d preview %d", w, d.CardsWidth(), d.PreviewWidth())
		}
		_ = d.View()
	}
}
