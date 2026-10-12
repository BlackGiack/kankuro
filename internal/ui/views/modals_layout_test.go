package views_test

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/uitest"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

// Contract (step 0 of impl_plan_2.md, reworked by impl_plan_5.md D). The card
// editor renders a modal as a REPLACEMENT for the form, never as a composite
// over it, so "an overflowing modal corrupts the view underneath" really means
// "the modal is larger than the screen, so lipgloss.Place passes it through
// unpadded and unclipped".
//
// All three modals go through CardEditor.renderWithOverlay =
// lipgloss.Place(W, H, centre, centre, modal), so the contract, asserted through
// the editor, is EXACT: while any modal is open the editor's View is exactly
// W x H - the screen does not jump when a modal opens. A modal larger than the
// screen would come out larger than W x H, which is what the *StaysWithin*
// sweeps catch (and what the user puppet, the long AGENTS.md line and the
// 20-row editor used to do). The modals' own boxes are exact too, at the width
// each one declares (TestWorkspacePickerModalIsAnExactBox,
// TestAgentsMDEditorModalHonoursItsWidth).
//
// Seen to fail: before impl_plan_5.md the workspace picker (returned bare),
// puppet selector (width from the art) and the three AGENTS.md tests (box 6
// columns too wide, wrapped rows, 22 rows tall) all failed; the bounded sweeps
// and exact-box tests were seen red by making the workspace picker's box 2 rows
// taller, the AGENTS.md text area 10 rows taller, and the puppet selector's
// padding 30 columns wider.

// modalSizes are view sizes for terminals from 80x24 to 160x60 (rows minus
// views.CoordinatorRows).
var modalSizes = [][2]int{ //nolint:gochecknoglobals // shared test fixture
	{80, 20},
	{81, 21},
	{100, 26},
	{120, 36},
	{160, 56},
}

func openModal(t *testing.T, e *views.CardEditor, focus int) {
	t.Helper()
	for i := 0; e.FocusIndex() != focus && i < views.FocusCount*2; i++ {
		e.CycleFocus(1)
	}
	if e.FocusIndex() != focus {
		t.Fatalf("could not focus target %d", focus)
	}
	e.Update(tea.KeyPressMsg{Code: tea.KeySpace})
}

func workspaceFixtures(n int) []config.Workspace {
	names := uitest.AdversarialNames
	paths := []string{
		"/home/user/projects/a-rather-long-directory-name/with/many/segments/and/more/segments/still",
		"/home/用户/项目/日本語のディレクトリ/とても長い名前",
		"/tmp/x",
	}
	ws := make([]config.Workspace, n)
	for i := range ws {
		ws[i] = config.Workspace{ID: fmt.Sprintf("ws-%d", i), Name: names[i%len(names)], Path: paths[i%len(paths)]}
	}
	return ws
}

func newModalEditor(t *testing.T, w, h int, data messages.DataLoadedMsg) *views.CardEditor {
	t.Helper()
	e := newLayoutEditor(true, true, 3, false)
	e.SetData(data)
	e.SetSize(w, h)
	return e
}

func TestWorkspacePickerModalStaysWithinTheScreen(t *testing.T) {
	for _, n := range []int{0, 1, 200} {
		for _, typed := range []string{"", "a", "漢"} {
			for _, sz := range modalSizes {
				w, h := sz[0], sz[1]
				e := newModalEditor(t, w, h, messages.DataLoadedMsg{Workspaces: workspaceFixtures(n)})
				openModal(t, e, views.FocusWorkspace)
				if typed != "" {
					e.Update(tea.KeyPressMsg{Code: 0, Text: typed})
				}
				label := fmt.Sprintf("workspaces=%d typed=%q %dx%d", n, typed, w, h)
				if !strings.Contains(ansi.Strip(e.View()), "Select Workspace") {
					t.Fatalf("%s: picker is not open:\n%s", label, ansi.Strip(e.View()))
				}
				uitest.AssertWithin(t, e.View(), w, h, label)
			}
		}
	}
}

// TestWorkspacePickerModalHasTheEditorsSize: the picker used to be returned
// bare, so it was 2/3 of the width and not centred: the form disappeared and a
// smaller box appeared at the top-left.
func TestWorkspacePickerModalHasTheEditorsSize(t *testing.T) {
	for _, sz := range modalSizes {
		e := newModalEditor(t, sz[0], sz[1], messages.DataLoadedMsg{Workspaces: workspaceFixtures(5)})
		openModal(t, e, views.FocusWorkspace)
		uitest.AssertExactRect(t, e.View(), sz[0], sz[1], fmt.Sprintf("%dx%d", sz[0], sz[1]))
	}
}

func TestPuppetSelectorModalStaysWithinTheScreen(t *testing.T) {
	for _, sz := range modalSizes {
		w, h := sz[0], sz[1]
		e := newModalEditor(t, w, h, messages.DataLoadedMsg{Puppets: builtinPuppetsForViews(t)})
		openModal(t, e, views.FocusPuppet)
		label := fmt.Sprintf("builtin puppets %dx%d", w, h)
		if !strings.Contains(ansi.Strip(e.View()), "Select Puppet:") {
			t.Fatalf("%s: selector is not open:\n%s", label, ansi.Strip(e.View()))
		}
		// The built-in row fits, so Place pads it to exactly the screen.
		uitest.AssertExactRect(t, e.View(), w, h, label)
	}
}

// TestPuppetSelectorModalClipsUserPuppets: the selector's width used to come
// from the art - every puppet side by side in one row - so a wide user puppet
// (or enough of them) made the modal wider than the terminal, and Place could
// not pad it down. Every puppet now has a fixed cell and the row scrolls.
func TestPuppetSelectorModalClipsUserPuppets(t *testing.T) {
	for _, sz := range modalSizes {
		puppets := append(append([]config.Puppet(nil), builtinPuppetsForViews(t)...), uitest.UserPuppet(60, 12))
		e := newModalEditor(t, sz[0], sz[1], messages.DataLoadedMsg{Puppets: puppets})
		openModal(t, e, views.FocusPuppet)
		uitest.AssertWithin(t, e.View(), sz[0], sz[1], fmt.Sprintf("user puppet %dx%d", sz[0], sz[1]))
	}
}

func TestAgentsMDEditorModalStaysWithinTheScreen(t *testing.T) {
	contents := map[string]string{
		"empty":      "",
		"many-lines": strings.Repeat("a line of the file\n", 300),
	}
	for name, content := range contents {
		for _, path := range []bool{false, true} {
			for _, sz := range modalSizes {
				w, h := sz[0], sz[1]
				e := newLayoutEditor(true, true, 3, false)
				e.SetCard(config.IDCard{
					ID:                 "c1",
					AgentsMD:           content,
					AgentsMDPath:       strings.Repeat("/long/path", 20),
					AgentsMDPreferPath: path,
				})
				e.SetSize(w, h)
				openModal(t, e, views.FocusAgentsMD)
				label := fmt.Sprintf("content=%s path=%v %dx%d", name, path, w, h)
				if !strings.Contains(ansi.Strip(e.View()), "AGENTS.md Editor") {
					t.Fatalf("%s: editor is not open:\n%s", label, ansi.Strip(e.View()))
				}
				uitest.AssertExactRect(t, e.View(), w, h, label)
			}
		}
	}
}

// TestAgentsMDEditorModalWrapsLongLines: the text area used to be as wide as 80%
// of the screen (+ its "> " marker) but the box's inner width is 6 columns
// narrower (border and padding), so every text row that filled the area wrapped,
// doubling the modal's height past the screen.
func TestAgentsMDEditorModalWrapsLongLines(t *testing.T) {
	for name, content := range map[string]string{
		"long-line":  strings.Repeat("word ", 200),
		"wide-runes": strings.Repeat("漢字のコンテンツ🚀", 60),
	} {
		for _, sz := range modalSizes[2:] {
			e := newLayoutEditor(true, true, 3, false)
			e.SetCard(config.IDCard{ID: "c1", AgentsMD: content})
			e.SetSize(sz[0], sz[1])
			openModal(t, e, views.FocusAgentsMD)
			uitest.AssertExactRect(t, e.View(), sz[0], sz[1], fmt.Sprintf("%s %dx%d", name, sz[0], sz[1]))
		}
	}
}

// TestAgentsMDEditorModalFitsSmallTerminal: at the 80x24 minimum (20 rows for
// the view) the modal's fixed rows plus a 10-row text area were taller than the
// screen, so Place passed it through and the editor's View was 22 rows. The text
// area now takes the rows that are left.
func TestAgentsMDEditorModalFitsSmallTerminal(t *testing.T) {
	e := newLayoutEditor(true, true, 3, false)
	e.SetSize(80, 20)
	openModal(t, e, views.FocusAgentsMD)
	uitest.AssertExactRect(t, e.View(), 80, 20, "80x24 terminal")
}

// firstLineWidth is the width a bordered box says it is: its first row.
func firstLineWidth(out string) int {
	return ansi.StringWidth(strings.SplitN(out, "\n", 2)[0])
}

// TestWorkspacePickerModalIsAnExactBox tests the picker directly: its own box -
// Width(2/3 of the screen, at least 40, at most the screen) x Height(the
// screen's rows), border included - must come out exactly that size whatever
// the list holds.
func TestWorkspacePickerModalIsAnExactBox(t *testing.T) {
	for _, n := range []int{0, 1, 200} {
		for _, typed := range []string{"", "a"} {
			for _, sz := range modalSizes {
				m := views.NewWorkspacePickerModal()
				m.SetWorkspaces(workspaceFixtures(n))
				m.SetSize(sz[0], sz[1])
				m.Open()
				if typed != "" {
					m.Update(tea.KeyPressMsg{Code: 0, Text: typed})
				}
				wantW, wantH := min(sz[0], max(40, sz[0]*2/3)), sz[1]
				uitest.AssertExactRect(t, m.View(), wantW, wantH, fmt.Sprintf("workspaces=%d typed=%q %dx%d", n, typed, sz[0], sz[1]))
			}
		}
	}
}

// TestPuppetSelectorModalIsAnIntactBox: the selector has no size of its own;
// whatever width the art gives it, the box must still be a clean rectangle -
// a ragged edge means the row wrapped.
func TestPuppetSelectorModalIsAnIntactBox(t *testing.T) {
	for name, puppets := range map[string][]config.Puppet{
		"builtin":    builtinPuppetsForViews(t),
		"with-user":  append(append([]config.Puppet(nil), builtinPuppetsForViews(t)...), uitest.UserPuppet(30, 8)),
		"user-alone": {uitest.UserPuppet(15, 4)},
	} {
		for cursor := range puppets {
			m := views.NewPuppetSelectorModal()
			m.SetPuppets(puppets)
			m.SetSelected(puppets[cursor].Name)
			m.Open()
			out := m.View()
			uitest.AssertExactRect(t, out, firstLineWidth(out), strings.Count(out, "\n")+1, fmt.Sprintf("%s cursor=%d", name, cursor))
		}
	}
}

// TestAgentsMDEditorModalIsAnIntactBox: the box around the editor stays a clean
// rectangle (every row the width of its first) for content that fits.
func TestAgentsMDEditorModalIsAnIntactBox(t *testing.T) {
	for _, sz := range modalSizes {
		for _, content := range []string{"", "short\ncontent", strings.Repeat("a line of the file\n", 100)} {
			m := views.NewAgentsMDEditorModal()
			m.SetSize(sz[0], sz[1])
			m.SetState(views.AgentsMDState{Content: content, Path: "/a/path"})
			m.Open()
			out := m.View()
			uitest.AssertExactRect(
				t,
				out,
				firstLineWidth(out),
				strings.Count(out, "\n")+1,
				fmt.Sprintf("content=%q %dx%d", content[:min(5, len(content))], sz[0], sz[1]),
			)
		}
	}
}

// TestAgentsMDEditorModalHonoursItsWidth: the box is declared
// min(W, max(50, 4/5 of W)) wide; it used to come out 6 columns wider with an
// empty or short text area - the text area was wider than the box's inner
// width, and only content that wrapped made the box obey.
func TestAgentsMDEditorModalHonoursItsWidth(t *testing.T) {
	for _, sz := range modalSizes {
		m := views.NewAgentsMDEditorModal()
		m.SetSize(sz[0], sz[1])
		m.SetState(views.AgentsMDState{Content: "short", Path: "/a/path"})
		m.Open()
		if got, want := firstLineWidth(m.View()), min(sz[0], max(50, sz[0]*4/5)); got != want {
			t.Errorf("%dx%d: box is %d wide, want %d", sz[0], sz[1], got, want)
		}
	}
}

// TestPuppetSelectorScrollsToTheCursor: with more puppets than fit, moving the
// cursor past the visible window scrolls it. The selected puppet is always
// drawn with its highlight border, an arrow marks each edge that has more, and
// the box keeps its shape and the editor's size.
func TestPuppetSelectorScrollsToTheCursor(t *testing.T) {
	var puppets []config.Puppet
	for i := range 14 {
		p := uitest.UserPuppet(10, 5)
		p.Name = fmt.Sprintf("user-%02d", i)
		p.Lines[2] = fmt.Sprintf("|%-8d|", i) // tells the cells apart
		puppets = append(puppets, p)
	}
	puppets = append(builtinPuppetsForViews(t), puppets...)

	for _, sz := range modalSizes {
		e := newModalEditor(t, sz[0], sz[1], messages.DataLoadedMsg{Puppets: puppets})
		openModal(t, e, views.FocusPuppet)
		sawLeft, sawRight := false, false
		for i := range puppets {
			label := fmt.Sprintf("%dx%d cursor=%d", sz[0], sz[1], i)
			out := e.View()
			uitest.AssertExactRect(t, out, sz[0], sz[1], label)
			plain := ansi.Strip(out)
			if got := strings.Count(plain, "╭"); got < 2 { // the modal's own corner and the cursor's
				t.Fatalf("%s: the selected puppet has no highlight border:\n%s", label, plain)
			}
			if i >= len(builtinPuppetsForViews(t)) {
				want := fmt.Sprintf("|%-8d|", i-len(builtinPuppetsForViews(t)))
				if !strings.Contains(plain, want) {
					t.Fatalf("%s: the selected puppet %q is not on screen:\n%s", label, want, plain)
				}
			}
			sawLeft = sawLeft || strings.Contains(plain, "‹")
			sawRight = sawRight || strings.Contains(plain, "›")
			e.Update(tea.KeyPressMsg{Code: tea.KeyRight})
		}
		if !sawLeft || !sawRight {
			t.Errorf("%dx%d: arrows seen left=%v right=%v, want both", sz[0], sz[1], sawLeft, sawRight)
		}
	}
}

// Contract (bonus_impl_plan.md, part A): the workspace picker keeps its
// elements in place. The column of the title, of the first item, of the page
// dots and of the legend (measured from the box's left border) and the row of
// the legend depend on the size of the picker only - not on which page is shown
// or on whether a search is active. Shape tests above check that the box fits;
// these check where things are.
//
// Seen to fail: the list was padded to its widest line and joined centred, so
// the list block and the legend moved sideways from page to page (columns
// 7/4/6/2 became 5/2/4/2 and the legend 4, 5, 7, 8, 10), and the legend moved
// down a row when a search started.

// mixedWorkspaces has short paths on the first and last pages and long ones in
// the middle, so the widest line on a page changes as it is paged through.
func mixedWorkspaces() []config.Workspace {
	ws := make([]config.Workspace, 40)
	for i := range ws {
		path := "/tmp/x"
		if i >= 10 && i < 30 {
			path = "/home/user/projects/" + strings.Repeat("a-rather-long-segment/", 6) + "end"
		}
		ws[i] = config.Workspace{ID: fmt.Sprintf("ws-%02d", i), Name: fmt.Sprintf("name-%02d", i), Path: path}
	}
	return ws
}

// pickerPlace is where the picker's elements sit: columns from the left border
// (-1 when absent) and the legend's row.
type pickerPlace struct {
	title, item, dots, legend, legendRow int
}

func measurePicker(out string) pickerPlace {
	p := pickerPlace{-1, -1, -1, -1, -1}
	col := func(line, needle string) int {
		beforeBorder, _, okBorder := strings.Cut(line, "│")
		beforeNeedle, _, okNeedle := strings.Cut(line, needle)
		if !okBorder || !okNeedle {
			return -1
		}
		return ansi.StringWidth(beforeNeedle) - ansi.StringWidth(beforeBorder)
	}
	for row, line := range strings.Split(ansi.Strip(out), "\n") {
		if c := col(line, "Select Workspace"); c >= 0 {
			p.title = c
		}
		if c := col(line, "name-"); c >= 0 && p.item < 0 {
			p.item = c
		}
		// The page dots are the line that holds nothing but dots.
		if inner := strings.Trim(line, "│ "); inner != "" && strings.Trim(inner, "•") == "" {
			p.dots = col(line, "•")
		}
		if c := col(line, "Enter=select"); c >= 0 {
			p.legend, p.legendRow = c, row
		}
	}
	return p
}

// pickerViews builds the picker the two ways it is shown: the modal on its own
// and through the card editor, for terminals of 80x24 and 120x30.
func pickerViews(t *testing.T, sz [2]int, ws []config.Workspace) map[string]struct {
	view func() string
	send func(tea.Msg)
} {
	t.Helper()
	type pv = struct {
		view func() string
		send func(tea.Msg)
	}
	m := views.NewWorkspacePickerModal()
	m.SetWorkspaces(ws)
	m.SetSize(sz[0], sz[1])
	m.Open()

	e := newModalEditor(t, sz[0], sz[1], messages.DataLoadedMsg{Workspaces: ws})
	openModal(t, e, views.FocusWorkspace)
	return map[string]pv{
		"modal":  {m.View, func(msg tea.Msg) { m.Update(msg) }},
		"editor": {e.View, func(msg tea.Msg) { e.Update(msg) }},
	}
}

var pickerTerminalSizes = [][2]int{{80, 20}, {120, 26}} //nolint:gochecknoglobals // shared test fixture

// firstWorkspaceName is the name of the first item on the page.
func firstWorkspaceName(plain string) string {
	_, rest, _ := strings.Cut(plain, "name-")
	return "name-" + rest[:len("00")]
}

type pickerPage struct {
	firstName, plain string
	place            pickerPlace
}

// pagePicker pages through the picker with → until the first page comes back
// and returns what each page looked like.
func pagePicker(view func() string, send func(tea.Msg)) []pickerPage {
	var pages []pickerPage
	seen := map[string]bool{}
	for range 40 {
		out := view()
		plain := ansi.Strip(out)
		name := firstWorkspaceName(plain)
		if seen[name] {
			break
		}
		seen[name] = true
		pages = append(pages, pickerPage{name, plain, measurePicker(out)})
		send(tea.KeyPressMsg{Code: tea.KeyRight})
	}
	return pages
}

func TestWorkspacePickerKeepsColumnsAcrossPages(t *testing.T) {
	for _, sz := range pickerTerminalSizes {
		for name, pv := range pickerViews(t, sz, mixedWorkspaces()) {
			label := fmt.Sprintf("%s %dx%d", name, sz[0], sz[1])
			pages := pagePicker(pv.view, pv.send)
			if len(pages) < 6 {
				t.Fatalf("%s: only %d pages, the test needs at least 6", label, len(pages))
			}
			first := pages[0].place
			if first.title < 0 || first.item < 0 || first.dots < 0 || first.legend < 0 {
				t.Fatalf("%s: an element is missing %+v:\n%s", label, first, pages[0].plain)
			}
			for i, page := range pages {
				if page.place != first {
					t.Errorf("%s page %d (%s): elements at %+v, want %+v as on page 0:\n%s",
						label, i, page.firstName, page.place, first, page.plain)
				}
			}
		}
	}
}

func TestWorkspacePickerLegendRowIsStable(t *testing.T) {
	for _, sz := range pickerTerminalSizes {
		for name, pv := range pickerViews(t, sz, mixedWorkspaces()) {
			want := measurePicker(pv.view()).legendRow
			if want < 0 {
				t.Fatalf("%s %dx%d: no legend:\n%s", name, sz[0], sz[1], ansi.Strip(pv.view()))
			}
			for _, typed := range []string{"n", "zzz"} {
				pv.send(tea.KeyPressMsg{Code: 0, Text: typed})
				if got := measurePicker(pv.view()).legendRow; got != want {
					t.Errorf("%s %dx%d after typing %q: legend on row %d, want %d:\n%s",
						name, sz[0], sz[1], typed, got, want, ansi.Strip(pv.view()))
				}
			}
			pv.send(tea.KeyPressMsg{Code: tea.KeyEsc})
			if got := measurePicker(pv.view()).legendRow; got != want {
				t.Errorf("%s %dx%d after clearing the search: legend on row %d, want %d", name, sz[0], sz[1], got, want)
			}
		}
	}
}
