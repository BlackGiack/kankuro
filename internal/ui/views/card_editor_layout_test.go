package views_test

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/ui/components"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/uitest"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

// Contract (step 0 of impl_plan_2.md, reworked by impl_plan_5.md): CardEditor.View()
// is BOUNDED. It is a plain top-to-bottom form with no border, so it does not
// pad to its size: no line wider than the width it was given and no more lines
// than the height it was given (the coordinator gives it terminal rows - 4, so
// 20 rows on the 80x24 minimum terminal).
//
// Region contract: the form is title, text inputs, workspace/puppet lines, the
// AGENTS.md row, then one capability list per visible section (Skills always,
// Personas and MCPs by app settings), then the help line - and the help line
// must be the LAST line for every combination. A section that is hidden must
// really be gone while the form is not scrolled.
//
// When the form is taller than its box it scrolls: the focused block stays on
// screen and a dim "more" row marks each cut edge
// (TestCardEditorFocusedFieldAlwaysVisible). The sweep therefore only checks
// the section list on forms that fit; the scroll tests check the rest.
//
// Seen to fail: the sweep was seen red at 80x20 before the form scrolled (the
// form was 30 lines tall), when the Personas list was rendered regardless of
// its setting, and when a trailing newline followed the help line.

const moreAbove, moreBelow = "↑ more", "↓ more"

// cardEditorMinRows is the editor height on the smallest supported terminal.
const cardEditorMinRows = views.MinTerminalHeight - views.CoordinatorRows

type layoutItem struct{ label, desc string }

func (i layoutItem) GetID() string          { return i.label }
func (i layoutItem) GetLabel() string       { return i.label }
func (i layoutItem) GetDescription() string { return i.desc }

// layoutItems builds n capability items. Labels are clipped to maxLabel columns
// here; pass a large maxLabel to see the list clip them itself.
func layoutItems(n, maxLabel int) []components.CapabilityItem {
	names := uitest.AdversarialNames
	items := make([]components.CapabilityItem, n)
	for i := range items {
		items[i] = layoutItem{
			label: fmt.Sprintf("%d-%s", i, ansi.Truncate(names[i%len(names)], maxLabel, "")),
			desc:  strings.Repeat("desc ", 30),
		}
	}
	return items
}

const sweepLabelWidth = 30

func newLayoutEditor(personas, mcps bool, items int, long bool) *views.CardEditor {
	e := views.NewCardEditor(views.CardEditorCallbacks{})
	settings := config.DefaultAppSettings()
	settings.ShowPersonasInCardEditor = personas
	settings.ShowMCPsInCardEditor = mcps
	e.SetSettings(settings)
	e.SkillList().SetItems(layoutItems(items, sweepLabelWidth))
	e.PersonaList().SetItems(layoutItems(items, sweepLabelWidth))
	e.MCPList().SetItems(layoutItems(items, sweepLabelWidth))
	card := config.IDCard{ID: "c1", Name: "Master mind", Driver: "claude", Model: "opus", Puppet: "cat"}
	if long {
		card = config.IDCard{
			ID:           "c1",
			Name:         strings.Repeat("n", 200),
			Driver:       "漢字のドライバー" + strings.Repeat("d", 100),
			Model:        "🚀" + strings.Repeat("m", 300),
			Puppet:       "cat",
			FaceStyle:    strings.Repeat("f", 100),
			AgentsMD:     "first line of content\nsecond line",
			AgentsMDPath: "/a/very/long/path/to/an/AGENTS.md/file/that/does/not/fit/in/the/summary",
		}
	}
	e.SetCard(card)
	return e
}

type cardEditorCase struct {
	personas, mcps, long bool
	items, w, h, focus   int
}

func (c cardEditorCase) label() string {
	return fmt.Sprintf("personas=%v mcps=%v items=%d long=%v %dx%d focus=%d",
		c.personas, c.mcps, c.items, c.long, c.w, c.h, c.focus)
}

// cardEditorCases is the sweep: every combination of the optional sections,
// content and size, with every focus target in turn (the focused row renders
// differently) at one size and focus 0 at the others.
func cardEditorCases() []cardEditorCase {
	var cases []cardEditorCase
	for _, sections := range [][2]bool{{false, false}, {false, true}, {true, false}, {true, true}} {
		personas, mcps := sections[0], sections[1]
		minH := cardEditorMinRows
		for _, content := range []struct {
			items int
			long  bool
		}{{0, false}, {0, true}, {40, false}, {40, true}} {
			for _, size := range [][2]int{{80, minH}, {80, minH + 1}, {80, 60}, {81, minH}, {160, minH}, {160, 60}} {
				foci := 1
				if size == [2]int{80, minH + 1} {
					foci = views.FocusCount
				}
				for focus := range foci {
					cases = append(cases, cardEditorCase{personas, mcps, content.long, content.items, size[0], size[1], focus})
				}
			}
		}
	}
	return cases
}

func TestCardEditorLayoutSweep(t *testing.T) {
	for _, c := range cardEditorCases() {
		e := newLayoutEditor(c.personas, c.mcps, c.items, c.long)
		e.SetSize(c.w, c.h)
		for i := 0; e.FocusIndex() != c.focus && i < views.FocusCount*2; i++ {
			e.CycleFocus(1)
		}
		if e.FocusIndex() != c.focus {
			continue // hidden section: Tab cannot land there
		}
		assertCardEditorShape(t, e, c.personas, c.mcps, c.w, c.h, c.label())
	}
}

func assertCardEditorShape(t *testing.T, e *views.CardEditor, personas, mcps bool, w, h int, label string) {
	t.Helper()
	out := e.View()
	uitest.AssertWithin(t, out, w, h, label)
	lines := strings.Split(ansi.Strip(out), "\n")
	if last := lines[len(lines)-1]; last != "Enter=save Tab=next Esc=cancel" {
		t.Errorf("%s: last line is %q, want the help line:\n%s", label, last, ansi.Strip(out))
	}
	if strings.Contains(out, moreAbove) || strings.Contains(out, moreBelow) {
		return // scrolled: which sections are on screen depends on the focus
	}
	for section, want := range map[string]bool{"Skills (": true, "Personas (": personas, "MCPs (": mcps} {
		if got := strings.Contains(strings.Join(lines, "\n"), section); got != want {
			t.Errorf("%s: section %q present=%v, want %v:\n%s", label, section, got, want, ansi.Strip(out))
		}
	}
}

// TestCardEditorFitsSmallTerminal: the editor must fit its box on every
// terminal from the 24-row minimum up. The capability lists are floored at
// minCapListHeight, so on a short terminal the form is taller than its box and
// has to scroll instead of overflowing (it used to be 30 lines in a 20-line
// box at 80x24 with the default settings, help line cut off).
func TestCardEditorFitsSmallTerminal(t *testing.T) {
	for _, personas := range []bool{false, true} {
		for _, mcps := range []bool{false, true} {
			for rows := views.MinTerminalHeight; rows < 40; rows++ {
				e := newLayoutEditor(personas, mcps, 40, false)
				e.SetSize(80, rows-views.CoordinatorRows)
				uitest.AssertWithin(t, e.View(), 80, rows-views.CoordinatorRows,
					fmt.Sprintf("personas=%v mcps=%v rows=%d", personas, mcps, rows))
			}
		}
	}
}

// TestCardEditorAgentsMDSummaryIsValidUTF8: the AGENTS.md row used to trim its
// content and path summaries by byte, so non-ASCII text was cut mid-rune (same
// defect as the session list's working-dir trimming).
func TestCardEditorAgentsMDSummaryIsValidUTF8(t *testing.T) {
	e := newLayoutEditor(false, false, 0, false)
	card := config.IDCard{
		ID:           "c1",
		AgentsMD:     strings.Repeat("漢字のコンテンツ", 5),
		AgentsMDPath: "/ホーム/ユーザー/プロジェクト/とても長い/パス/AGENTS.md",
	}
	e.SetCard(card)
	e.SetSize(100, 40)
	if out := e.View(); !utf8.ValidString(out) {
		t.Errorf("view is not valid UTF-8:\n%q", out)
	}
}

// TestCardEditorLongLinesAreClipped: the workspace and puppet lines and the
// capability list labels used to be written out unclipped, so a long workspace
// name, a long puppet name or a long skill/MCP name made the line wider than
// the editor (the terminal then wraps it, pushing the help line off the bottom).
func TestCardEditorLongLinesAreClipped(t *testing.T) {
	e := newLayoutEditor(true, true, 0, false)
	e.SkillList().SetItems(layoutItems(5, 500))
	card := config.IDCard{ID: "c1", Name: "n", Puppet: strings.Repeat("p", 200), WorkspaceID: strings.Repeat("w", 200)}
	e.SetCard(card)
	e.SetSize(80, 50)
	uitest.AssertWithin(t, e.View(), 80, 50, "long lines")
}

func TestFocusedSubWorkspaceClearHintFitsLongSummaries(t *testing.T) {
	for _, width := range []int{views.MinTerminalWidth, 120, 200} {
		e := newLayoutEditor(true, true, 0, false)
		workspaces := []config.Workspace{
			{ID: "main", Name: "Main"},
			{ID: "sub-a", Name: strings.Repeat("Alpha Workspace ", 8)},
			{ID: "sub-b", Name: strings.Repeat("Beta Workspace ", 8)},
		}
		e.SetData(messages.DataLoadedMsg{Workspaces: workspaces})
		e.SetCard(config.IDCard{ID: "long-sub-workspaces", WorkspaceID: "main", SubWorkspaceIDs: []string{"sub-a", "sub-b"}})
		e.CycleFocus(views.FocusSubWorkspaces)
		height := views.MinTerminalHeight - views.CoordinatorRows
		e.SetSize(width, height)
		out := e.View()
		uitest.AssertWithin(t, out, width, height, fmt.Sprintf("width=%d", width))
		var field string
		for line := range strings.SplitSeq(ansi.Strip(out), "\n") {
			if strings.HasPrefix(line, "> Sub Workspaces: ") {
				field = line
				break
			}
		}
		if field == "" || !strings.Contains(field, "ctrl+x=clear") {
			t.Fatalf("width %d focused sub workspace field omitted clear hint:\n%s", width, ansi.Strip(out))
		}
		if got := ansi.StringWidth(field); got > width {
			t.Fatalf("width %d focused sub workspace row is %d columns", width, got)
		}
	}
}

// focusMarkerLine returns what the view must show for the focused target: the
// text of the line that marks focus. Text inputs have no "> " marker (the
// cursor is the marker), so their label is checked - it is the line's start,
// and is present only when the line is on screen.
func focusMarkerLine(focus int) []string {
	switch focus {
	case views.FocusName:
		return []string{"Name: "}
	case views.FocusDriver:
		return []string{"Driver: "}
	case views.FocusModel:
		return []string{"Model: "}
	case views.FocusFaceStyle:
		return []string{"Face Style: "}
	case views.FocusUseTmp:
		return []string{"Use Tmp Workspace: "}
	case views.FocusDashboardControl:
		return []string{"> Dashboard Control: "}
	case views.FocusWorkspace:
		return []string{"> Workspace: "}
	case views.FocusSubWorkspaces:
		return []string{"> Sub Workspaces: "}
	case views.FocusPuppet:
		return []string{"> Puppet: "}
	case views.FocusAgentsMD:
		return []string{"> x [ Content: ", "> x [ Path: ", ">   [ Content: ", ">   [ Path: "}
	case views.FocusSkills, views.FocusPersonas, views.FocusMCPs:
		return []string{"> [", "> (no matching"} // the list's cursor row
	}
	return nil
}

// TestCardEditorFocusedFieldAlwaysVisible: on short terminals the form
// scrolls, and Tab through every target must keep the focused one on screen
// and the view inside its box. Below the 24-row minimum (20 and 16 editor rows)
// is covered too: the editor is still sized then, behind the "too small" screen.
//
// Seen to fail: with scrollOffset frozen at 0 (syncScroll returning early) every
// target below the AGENTS.md block went red at 80x20 and at 80x24.
func TestCardEditorFocusedFieldAlwaysVisible(t *testing.T) {
	for _, sections := range [][2]bool{{false, false}, {false, true}, {true, false}, {true, true}} {
		personas, mcps := sections[0], sections[1]
		for _, rows := range []int{20, views.MinTerminalHeight} {
			for _, items := range []int{0, 40} {
				e := newLayoutEditor(personas, mcps, items, false)
				e.SetSize(80, rows-views.CoordinatorRows)
				visited := 0
				for {
					assertFocusedVisible(t, e, fmt.Sprintf("personas=%v mcps=%v rows=%d items=%d focus=%d",
						personas, mcps, rows, items, e.FocusIndex()), rows-views.CoordinatorRows)
					visited++
					e.CycleFocus(1)
					if e.FocusIndex() == views.FocusName {
						break
					}
				}
				if want := 11 + btoi(personas) + btoi(mcps); visited != want {
					t.Fatalf("visited %d focus targets, want %d", visited, want)
				}
			}
		}
	}
}

func assertFocusedVisible(t *testing.T, e *views.CardEditor, label string, h int) {
	t.Helper()
	out := e.View()
	uitest.AssertWithin(t, out, 80, h, label)
	plain := ansi.Strip(out)
	for _, marker := range focusMarkerLine(e.FocusIndex()) {
		if strings.Contains(plain, marker) {
			return
		}
	}
	t.Errorf("%s: the focused target is not on screen:\n%s", label, plain)
}

// TestCardEditorScrollIndicators: a form cut at an edge says so on that edge
// only, in the row the cut content would have used, and the help line stays last.
func TestCardEditorScrollIndicators(t *testing.T) {
	e := newLayoutEditor(true, true, 40, false)
	e.SetSize(80, cardEditorMinRows)
	top := strings.Split(ansi.Strip(e.View()), "\n")
	if strings.Contains(top[0], moreAbove) || !strings.Contains(top[len(top)-2], moreBelow) {
		t.Errorf("at the top want no %q and %q on the row above the help line:\n%s", moreAbove, moreBelow, strings.Join(top, "\n"))
	}
	for e.FocusIndex() != views.FocusMCPs {
		e.CycleFocus(1)
	}
	bottom := strings.Split(ansi.Strip(e.View()), "\n")
	if !strings.Contains(bottom[0], moreAbove) || strings.Contains(bottom[len(bottom)-2], moreBelow) {
		t.Errorf("at the bottom want %q on top and no %q:\n%s", moreAbove, moreBelow, strings.Join(bottom, "\n"))
	}
	if len(bottom) != cardEditorMinRows {
		t.Errorf("scrolled form is %d lines, want exactly %d", len(bottom), cardEditorMinRows)
	}
}

func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

// TestCardEditorFitsEveryWidth sweeps each width from the minimum to well past
// it with long and adversarial content: the text inputs sit after a label, so
// their width is derived from the editor's, and a derivation that is off by one
// shows up at some widths only (the way the card's pill row did).
func TestCardEditorFitsEveryWidth(t *testing.T) {
	for w := views.MinTerminalWidth; w <= 200; w++ {
		e := newLayoutEditor(true, true, 40, true)
		e.SkillList().SetItems(layoutItems(5, 500))
		e.SetSize(w, cardEditorMinRows)
		for range views.FocusCount {
			uitest.AssertWithin(t, e.View(), w, cardEditorMinRows, fmt.Sprintf("width=%d focus=%d", w, e.FocusIndex()))
			e.CycleFocus(1)
		}
	}
}

// rowOf returns the first row of the plain view that starts with prefix, or -1.
func rowOf(plain, prefix string) int {
	for i, line := range strings.Split(plain, "\n") {
		if strings.HasPrefix(line, prefix) {
			return i
		}
	}
	return -1
}

func typeInto(e *views.CardEditor, text string) {
	for _, ch := range text {
		e.Update(tea.KeyPressMsg{Code: ch, Text: string(ch)})
	}
}

func pressKey(e *views.CardEditor, code rune, n int) {
	for range n {
		e.Update(tea.KeyPressMsg{Code: code})
	}
}

// TestCardEditorListsKeepTheirPlace: the capability lists are stacked, so a
// list that changes height moves everything below it. On a terminal tall
// enough that the form does not scroll, the rows of every list header and of
// the help line must not change when a list scrolls or is searched.
//
// Seen to fail: scrolling a list into its middle pushed the next header down a
// row (both "..." rows), and a search matching 1 item or none pulled it up.
func TestCardEditorListsKeepTheirPlace(t *testing.T) {
	// 15 rows of fields + 3 lists of 7 rows + 2 separator rows each + the help
	// line = 41 rows: the form fits without scrolling. (30 terminal rows would
	// floor the lists at minCapListHeight and scroll.)
	const terminalRows = 47
	h := terminalRows - views.CoordinatorRows
	e := newLayoutEditor(true, true, 0, false)
	for _, l := range []struct {
		name string
		list *components.CapabilityList
	}{{"skill", e.SkillList()}, {"persona", e.PersonaList()}, {"mcp", e.MCPList()}} {
		items := make([]components.CapabilityItem, 20)
		for i := range items {
			items[i] = layoutItem{label: fmt.Sprintf("%s-%02d", l.name, i)}
		}
		l.list.SetItems(items)
	}
	e.SetSize(80, h)

	places := func() [4]int {
		plain := ansi.Strip(e.View())
		if strings.Contains(plain, moreAbove) || strings.Contains(plain, moreBelow) {
			t.Fatalf("the form scrolls at %d rows; pick a taller terminal:\n%s", terminalRows, plain)
		}
		return [4]int{rowOf(plain, "Skills ("), rowOf(plain, "Personas ("), rowOf(plain, "MCPs ("), rowOf(plain, "Enter=save")}
	}
	base := places()
	for i, row := range base {
		if row < 0 {
			t.Fatalf("element %d is not on screen:\n%s", i, ansi.Strip(e.View()))
		}
	}

	for _, target := range []struct {
		name   string
		prefix string
	}{{"skills", "skill"}, {"personas", "persona"}} {
		if target.name == "personas" {
			e.CycleFocus(1)
		} else {
			for e.FocusIndex() != views.FocusSkills {
				e.CycleFocus(1)
			}
		}
		check := func(state string) {
			t.Helper()
			if got := places(); got != base {
				t.Errorf("%s, %s: rows of [skills personas mcps help] are %v, want %v:\n%s",
					target.name, state, got, base, ansi.Strip(e.View()))
			}
		}
		pressKey(e, tea.KeyDown, 10)
		check("cursor in the middle")
		pressKey(e, tea.KeyDown, 10)
		check("cursor at the bottom")
		typeInto(e, target.prefix+"-07")
		check("search matching 1 item")
		pressKey(e, tea.KeyBackspace, len(target.prefix+"-07"))
		typeInto(e, "zzz")
		check("search matching nothing")
		pressKey(e, tea.KeyBackspace, len("zzz"))
		check("search cleared")
	}
}
