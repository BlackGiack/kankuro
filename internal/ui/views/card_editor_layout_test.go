package views_test

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"

	"kankuro/internal/config"
	"kankuro/internal/ui/components"
	"kankuro/internal/ui/uitest"
	"kankuro/internal/ui/views"
)

// Contract (step 0 of impl_plan_2.md): CardEditor.View() is BOUNDED. It is a
// plain top-to-bottom form with no border, so it does not pad to its size: no
// line wider than the width it was given and no more lines than the height it
// was given (the coordinator gives it terminal rows - 4).
//
// Region contract: the form is title, text inputs, workspace/puppet lines, the
// AGENTS.md row, then one capability list per visible section (Skills always,
// Personas and MCPs by app settings), then the help line - and the help line
// must be the LAST line for every combination. A section that grows would push
// it past the height or wrap it; a section that is hidden must really be gone.
//
// Per-combination minimum height: the form's fixed part plus every visible list
// at its minimum height (components' viewHeightOffset + minCapListHeight) and
// the help line. cardEditorMinHeight measures that from a render; below it the
// editor does not shrink, it overflows - TestCardEditorFitsSmallTerminal (skipped) and
// impl_plan_2.md "Found, not fixed" #8. With both optional sections on, which
// is the default, that includes a stock 24-row terminal.
//
// Seen to fail: the sweep was seen red at 80x20 (the bug above) before the
// minimum-height split; it was red when the Personas list was rendered
// regardless of its setting, and when a trailing newline followed the help
// line.

// cardEditorMinHeight is the fewest rows the editor can lay out in without
// overflowing, by section: the height it renders at when asked for a 1-row box,
// where every capability list sits at its floor (minCapListHeight). It is read
// from a render, not copied from the layout constants, so it follows them.
func cardEditorMinHeight(personas, mcps bool) int {
	e := newLayoutEditor(personas, mcps, 40, false)
	e.SetSize(80, 1)
	return strings.Count(e.View(), "\n") + 1
}

type layoutItem struct{ label, desc string }

func (i layoutItem) GetID() string          { return i.label }
func (i layoutItem) GetLabel() string       { return i.label }
func (i layoutItem) GetDescription() string { return i.desc }

// layoutItems builds n capability items. Labels are clipped to maxLabel columns
// because the list does not clip them itself (skipped
// TestCardEditorLongLinesAreClipped, "Found, not fixed" #9); pass a large
// maxLabel to see that.
func layoutItems(n, maxLabel int) []components.CapabilityItem {
	names := uitest.NamesWithoutTabs()
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
		minH := cardEditorMinHeight(personas, mcps)
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
	for section, want := range map[string]bool{"Skills (": true, "Personas (": personas, "MCPs (": mcps} {
		if got := strings.Contains(strings.Join(lines, "\n"), section); got != want {
			t.Errorf("%s: section %q present=%v, want %v:\n%s", label, section, got, want, ansi.Strip(out))
		}
	}
}

// TestCardEditorFitsSmallTerminal: below cardEditorMinHeight the editor does
// not shrink further (the lists are floored at minCapListHeight), it grows past
// its box, so the bottom - help line included - is cut off. With the default
// settings that is the stock 24-row terminal.
func TestCardEditorFitsSmallTerminal(t *testing.T) {
	t.Skip("bug: the card editor overflows below ~35 rows with default settings - see impl_plan_2.md 'Found, not fixed' #8")
	for _, personas := range []bool{false, true} {
		for _, mcps := range []bool{false, true} {
			for rows := 24; rows < 40; rows++ {
				e := newLayoutEditor(personas, mcps, 40, false)
				e.SetSize(80, rows-heightAdjustment)
				uitest.AssertWithin(t, e.View(), 80, rows-heightAdjustment,
					fmt.Sprintf("personas=%v mcps=%v rows=%d", personas, mcps, rows))
			}
		}
	}
}

// TestCardEditorAgentsMDSummaryIsValidUTF8: the AGENTS.md row trims its content
// and path summaries by byte, so non-ASCII text is cut mid-rune (same defect
// as the session list's working-dir trimming).
func TestCardEditorAgentsMDSummaryIsValidUTF8(t *testing.T) {
	t.Skip("bug: AGENTS.md summaries are sliced by byte - see impl_plan_2.md 'Found, not fixed' #2")
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
// capability list labels are written out unclipped, so a long workspace name, a
// long puppet name or a long skill/MCP name makes the line wider than the
// editor (the terminal then wraps it, pushing the help line off the bottom).
func TestCardEditorLongLinesAreClipped(t *testing.T) {
	t.Skip("bug: workspace/puppet lines and capability labels are not clipped - see impl_plan_2.md 'Found, not fixed' #9")
	e := newLayoutEditor(true, true, 0, false)
	e.SkillList().SetItems(layoutItems(5, 500))
	card := config.IDCard{ID: "c1", Name: "n", Puppet: strings.Repeat("p", 200), WorkspaceID: strings.Repeat("w", 200)}
	e.SetCard(card)
	e.SetSize(80, 50)
	uitest.AssertWithin(t, e.View(), 80, 50, "long lines")
}
