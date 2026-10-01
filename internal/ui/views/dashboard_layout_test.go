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
// (terminal width, terminal rows - 4) - app.heightAdjustment - and
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
// a footer wider than the terminal (TestDashboardFooterFitsWidth, skipped).
//
// Region contract: the panes keep their columns. With cards = cw, preview = pw
// and sessions = W-cw-pw (Dashboard.CardsWidth / PreviewWidth), a vertical
// border sits at columns cw-1, cw, cw+pw-1, cw+pw and W-1 on every body row,
// and the corners on the first and last body row. A pane that grows by one
// column moves its neighbour's border, which this catches even when the total
// width survives.
//
// Smallest supported terminal: 80x24 (what a stock terminal opens at). Below
// that the layout is undefined - there is no "terminal too small" handling and
// the preview width goes negative under ~62 columns; see
// TestDashboardBelowMinimumWidth (skipped, impl_plan_2.md "Found, not fixed").
//
// Seen to fail: the shape/region sweeps were red at 80 and 120 columns against
// the old code because of the session list help line (fixed since, by
// SetShowHelp(false)); the delete-modal size, footer, narrow-width and
// long-query tests fail against current code and are skipped with the reasons
// in impl_plan_2.md. The sweeps were also seen red by widening the preview pane
// by one column (Width(d.previewWidth + 1)) and by shortening the body by one
// row; the focus test by swapping which pane gets the focused border.

const (
	dashboardMinWidth = 80
	dashboardMinRows  = 24
	// dashboardChrome is the rows around the panes: header (1) + its blank
	// margin row (1) + footer (1). With the panes' own h+1 rows that makes the
	// full render h+4 = terminal rows, matching app.heightAdjustment = 4.
	dashboardChrome  = 3
	heightAdjustment = 4 // app.heightAdjustment; the dashboard is sized to rows-4
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
	names := uitest.NamesWithoutTabs()
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
	dash.SetSize(w, rows-heightAdjustment)
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
	// The footer's width is checked by the skipped TestDashboardFooterFitsWidth.
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
			for _, query := range []string{"a", "漢字", "zzzz"} {
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
// a pane title is not truncated, so a query wider than the pane wraps inside
// it and the dashboard grows by one to four rows (2 to 4 at 100 columns with an
// 80-character query).
func TestDashboardLongSearchQueryFits(t *testing.T) {
	t.Skip("bug: long search query wraps in the pane title - see impl_plan_2.md 'Found, not fixed' #7")
	for _, w := range []int{100, 130} {
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
	t.Skip("bug: delete-modal overlay is 4 rows shorter than the dashboard - see impl_plan_2.md 'Found, not fixed' #6")
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

// TestDashboardDeleteModalIsCentredAndIntact is what holds today: the overlay
// is exactly as wide as the screen, so the shape is right even though its
// height is not (above).
func TestDashboardDeleteModalIsCentredAndIntact(t *testing.T) {
	eachSize([]int{80, 100, 160}, []int{24, 40}, func(w, rows int) {
		d := newDashboard(t, w, rows, dashboardData{sessions: 2, cards: 2})
		d.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
		if !d.DeleteModal().IsOpen() {
			t.Fatal("delete modal did not open")
		}
		out := d.View()
		label := fmt.Sprintf("%dx%d modal open", w, rows)
		uitest.AssertExactRect(t, out, w, rows-heightAdjustment, label)
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

// TestDashboardFooterFitsWidth: the footer key legend is a fixed 150 columns
// wide, so it is cut off (or wrapped, depending on the terminal) on anything
// narrower.
func TestDashboardFooterFitsWidth(t *testing.T) {
	t.Skip("bug: footer legend is 150 columns wide - see impl_plan_2.md 'Found, not fixed' #3")
	for _, w := range []int{80, 100, 120} {
		d := newDashboard(t, w, 30, dashboardData{sessions: 1, cards: 1})
		lines := strings.Split(d.View(), "\n")
		if got := ansi.StringWidth(lines[len(lines)-1]); got > w {
			t.Errorf("width %d: footer is %d columns wide", w, got)
		}
	}
}

// TestDashboardBelowMinimumWidth probes the narrow end. Below ~62 columns
// SetSize forces the cards pane to 46 and takes the difference from the
// preview, whose width then goes negative; between 62 and 79 the preview pane
// is under 14 columns and the render grows a row. Whatever the fix (a "terminal
// too small" screen, or a proper responsive layout), the render must fit.
func TestDashboardBelowMinimumWidth(t *testing.T) {
	t.Skip("bug: no behaviour below 80 columns, preview width goes negative - see impl_plan_2.md 'Found, not fixed' #5")
	for w := 40; w < dashboardMinWidth; w++ {
		d := newDashboard(t, w, dashboardMinRows, dashboardData{sessions: 1, cards: 1})
		out := d.View()
		uitest.AssertWithin(t, out, w, dashboardMinRows, fmt.Sprintf("width %d", w))
	}
}
