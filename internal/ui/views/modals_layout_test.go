package views_test

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"kankuro/internal/config"
	"kankuro/internal/ui/messages"
	"kankuro/internal/ui/uitest"
	"kankuro/internal/ui/views"
)

// Contract (step 0 of impl_plan_2.md). The card editor renders a modal as a
// REPLACEMENT for the form, never as a composite over it, so "an overflowing
// modal corrupts the view underneath" really means "the modal is larger than
// the screen, so lipgloss.Place passes it through unpadded and unclipped".
//
//   - AgentsMDEditorModal and PuppetSelectorModal go through
//     CardEditor.renderWithOverlay = lipgloss.Place(W, H, centre, centre, modal):
//     when the modal fits, the editor's View is EXACTLY W x H; when it does not,
//     Place returns the modal as is, larger than W x H.
//   - WorkspacePickerModal is returned bare: the editor's View is the modal's
//     own size (2/3 of the width), which is neither W x H nor centred.
//
// So the contract asserted for all three, through the editor, is BOUNDED
// (nothing wider or taller than W x H), and the strict contract - the editor's
// View is exactly W x H while any modal is open, i.e. the screen does not jump
// when a modal opens - is a skipped test per modal where current code breaks it.
//
// Seen to fail: every skipped test below failed against current code; the
// bounded sweeps and exact-box tests were seen red by making the workspace
// picker's box 2 rows taller, the AGENTS.md text area 10 rows taller, and the
// puppet selector's padding 30 columns wider.

// modalSizes are view sizes for terminals from 80x24 to 160x60 (rows minus
// heightAdjustment).
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
	names := uitest.NamesWithoutTabs()
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

// TestWorkspacePickerModalHasTheEditorsSize: the picker is returned bare, so it
// is 2/3 of the width and not centred: the form disappears and a smaller box
// appears at the top-left.
func TestWorkspacePickerModalHasTheEditorsSize(t *testing.T) {
	t.Skip("bug: workspace picker is not placed in the editor's W x H - see impl_plan_2.md 'Found, not fixed' #10")
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

// TestPuppetSelectorModalClipsUserPuppets: the selector has no SetSize and its
// width comes from the art - every puppet side by side in one row - so a wide
// user puppet (or enough of them) makes the modal wider than the terminal, and
// Place cannot pad it down.
func TestPuppetSelectorModalClipsUserPuppets(t *testing.T) {
	t.Skip("bug: puppet selector width comes from the art and is not clipped - see impl_plan_2.md 'Found, not fixed' #11")
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
				if h < 26 {
					continue // below 26 rows the modal's fixed parts overflow; see TestAgentsMDEditorModalFitsSmallTerminal
				}
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

// TestAgentsMDEditorModalWrapsLongLines: the text area is as wide as 80% of the
// screen (+ its "> " marker) but the box's inner width is 6 columns narrower
// (border and padding), so every text row that fills the area wraps, doubling
// the modal's height past the screen.
func TestAgentsMDEditorModalWrapsLongLines(t *testing.T) {
	t.Skip("bug: AGENTS.md text area is wider than the modal's inner width - see impl_plan_2.md 'Found, not fixed' #12")
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
// the view) the modal's fixed rows plus its 10-row text area are taller than
// the screen, so Place passes it through and the editor's View is 22 rows.
func TestAgentsMDEditorModalFitsSmallTerminal(t *testing.T) {
	t.Skip("bug: AGENTS.md editor modal is taller than a 24-row terminal - see impl_plan_2.md 'Found, not fixed' #12")
	e := newLayoutEditor(true, true, 3, false)
	e.SetSize(80, 20)
	openModal(t, e, views.FocusAgentsMD)
	uitest.AssertExactRect(t, e.View(), 80, 20, "80x24 terminal")
}

// firstLineWidth is the width a bordered box says it is: its first row.
func firstLineWidth(out string) int {
	return ansi.StringWidth(strings.SplitN(out, "\n", 2)[0])
}

// TestWorkspacePickerModalIsAnExactBox tests the picker directly. It is
// returned bare by the editor, so its own box - Width(2/3 of the screen, at
// least 40) x Height(the screen's rows, at least 15), border included - must
// come out exactly that size whatever the list holds.
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
				wantW, wantH := max(40, sz[0]*2/3), max(15, sz[1])
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

// TestAgentsMDEditorModalHonoursItsWidth: the box is declared Width(max(50, 4/5
// of the screen)) but comes out 6 columns wider with an empty or short text
// area - the text area (80% of the screen plus its "> " marker) is wider than
// the box's inner width, and only content that wraps makes the box obey.
func TestAgentsMDEditorModalHonoursItsWidth(t *testing.T) {
	t.Skip("bug: AGENTS.md box is 6 columns wider than its declared width - see impl_plan_2.md 'Found, not fixed' #12")
	for _, sz := range modalSizes {
		m := views.NewAgentsMDEditorModal()
		m.SetSize(sz[0], sz[1])
		m.SetState(views.AgentsMDState{Content: "short", Path: "/a/path"})
		m.Open()
		if got, want := firstLineWidth(m.View()), max(50, sz[0]*4/5); got != want {
			t.Errorf("%dx%d: box is %d wide, want %d", sz[0], sz[1], got, want)
		}
	}
}
