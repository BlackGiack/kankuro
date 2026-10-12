package components_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/ui/components"
	"github.com/BlackGiack/kankuro/internal/ui/uitest"
)

// renderedCardHeight is the line count of a correctly laid-out card: the face
// box (FaceBoxHeight) plus the top and bottom border.
const renderedCardHeight = components.FaceBoxHeight + 2

func builtinPuppets(t *testing.T) []config.Puppet {
	t.Helper()
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	puppets, err := repo.ListPuppets()
	if err != nil {
		t.Fatalf("list puppets: %v", err)
	}
	return puppets
}

func cardWithSkills(names []string) (config.IDCard, []config.Skill) {
	card := config.IDCard{Name: "Master mind", Driver: "claude", Puppet: "cat"}
	var skills []config.Skill
	for i, n := range names {
		id := fmt.Sprintf("skill-%d", i)
		skills = append(skills, config.Skill{ID: id, Name: n})
		card.SkillIDs = append(card.SkillIDs, id)
	}
	return card, skills
}

// assertIntactCard fails if the rendered card is not a clean rectangle of the
// expected height. A row wider than the card's inner width makes the outer
// border wrap it, which shows up as ragged line widths and extra lines - the
// face box's art spilling onto new lines at the left edge.
func assertIntactCard(t *testing.T, out, label string) {
	t.Helper()
	uitest.AssertExactRect(t, out, ansi.StringWidth(strings.SplitN(out, "\n", 2)[0]), renderedCardHeight, label)
}

// assertSinglePillRow fails if the capability pills spill past their one row.
// The info box holds name, driver, a blank line and the pill row (lines 1-4
// inside the border); line 5 is only there because the face box is taller,
// so its info-box half must stay blank. Pills wrapping onto it is how a pill
// ends up split in two ("api-" / "pat...").
func assertSinglePillRow(t *testing.T, out, label string) {
	t.Helper()
	lines := strings.Split(ansi.Strip(out), "\n")
	if len(lines) != renderedCardHeight {
		return // assertIntactCard reports this
	}
	// "│ info box │ face box │" - the info box is the first cell.
	cells := strings.Split(lines[components.InfoBoxHeight+1], "│")
	if len(cells) < 3 {
		t.Errorf("%s: unexpected row shape %q", label, lines[components.InfoBoxHeight+1])
		return
	}
	if strings.TrimSpace(cells[1]) != "" {
		t.Errorf("%s: pills spilled onto a second row %q:\n%s", label, cells[1], ansi.Strip(out))
	}
}

// TestRenderCardItemStaysIntactAcrossWidths is the regression test for the
// card that shattered only for particular skill sets at particular list
// widths: each pill ends in two columns of whitespace, lipgloss word-wrap
// leaves whitespace hanging past the box edge, and when the edge fell inside
// those two columns the pill row came out 1-2 columns too wide.
//
// The first set is the exact card that was reported (it broke at list widths
// 50, 51 and 63); the others shift the pill boundaries so the sweep covers
// many edge positions, not just one.
func TestRenderCardItemStaysIntactAcrossWidths(t *testing.T) {
	puppets := builtinPuppets(t)
	sets := [][]string{
		{"architecture", "clean-code", "golang", "plan-writing"},
		{"architecture", "clean-code", "api-patterns", "plan-writing"},
		{"architecture", "golang", "plan-writing"},
		{"architecture", "clean-code", "golang", "plan-writing", "api-patterns"},
		{"a", "bb", "ccc", "dddd", "eeeee", "ffffff"},
		{"golang"},
		{"日本語のスキル", "clean-code", "golang"},
		nil,
	}
	for _, names := range sets {
		card, skills := cardWithSkills(names)
		for _, selected := range []bool{false, true} {
			for width := components.CardListMinWidth; width <= 160; width++ {
				out := components.RenderCardItem(card, skills, nil, puppets, selected, width)
				label := fmt.Sprintf("skills=%v selected=%v width=%d", names, selected, width)
				assertIntactCard(t, out, label)
				assertSinglePillRow(t, out, label)
			}
		}
	}
}

// TestRenderCardItemSummarisesPillsThatDoNotFit checks the "+N" summary:
// capabilities that do not fit on the row are counted, not silently dropped,
// and the count accounts for everything hidden.
func TestRenderCardItemSummarisesPillsThatDoNotFit(t *testing.T) {
	puppets := builtinPuppets(t)
	names := []string{"architecture", "clean-code", "golang", "plan-writing", "api-patterns", "rust-pro"}
	card, skills := cardWithSkills(names)

	narrow := ansi.Strip(components.RenderCardItem(card, skills, nil, puppets, false, components.CardListMinWidth))
	wide := ansi.Strip(components.RenderCardItem(card, skills, nil, puppets, false, 160))

	for label, out := range map[string]string{"narrow": narrow, "wide": wide} {
		shown := 0
		for _, n := range []string{"archite", "clean-code", "golang", "plan-wr", "api-pat", "rust-pro"} {
			if strings.Contains(out, n) {
				shown++
			}
		}
		hidden := len(names) - shown
		if hidden == 0 {
			continue
		}
		if want := fmt.Sprintf("+%d", hidden); !strings.Contains(out, want) {
			t.Errorf("%s: %d of %d skills shown, expected a %q summary pill:\n%s", label, shown, len(names), want, out)
		}
	}

	// MaxPillsShown still caps a wide card, so six skills can never all show.
	if !strings.Contains(wide, "+") {
		t.Errorf("wide card should summarise skills beyond MaxPillsShown (%d):\n%s", components.MaxPillsShown, wide)
	}
}

// TestRenderCardItemClipsOversizedPuppet covers the face box's half of the
// row: puppets can come from user puppet sources, and art wider than the box
// used to widen the row and shatter the card exactly like the pill overflow.
func TestRenderCardItemClipsOversizedPuppet(t *testing.T) {
	wide := config.Puppet{
		ID:    "wide",
		Name:  "wide",
		Lines: []string{"<==================>", "|  much too wide   |", "<==================>"},
	}
	card, skills := cardWithSkills([]string{"golang"})
	card.Puppet = "wide"
	for width := components.CardListMinWidth; width <= 120; width++ {
		out := components.RenderCardItem(card, skills, nil, []config.Puppet{wide}, false, width)
		assertIntactCard(t, out, fmt.Sprintf("oversized puppet width=%d", width))
	}
}

// TestRenderCardItemSanitisesUserText: tabs, line breaks and escapes in a
// card's name, driver, model and skill names must not change the card's shape.
// lipgloss expands a tab to four columns after the text was fitted, so a name
// of 18 tab-separated letters is 18 columns to the fitting code and 57 to the
// terminal. Seen to fail with DisplayText removed from the name line.
func TestRenderCardItemSanitisesUserText(t *testing.T) {
	puppets := builtinPuppets(t)
	messy := "a\tb\tc\td\te\tf\tg\th\ti\tj\nk\r\x1b[31ml\x1b[0m"
	card, skills := cardWithSkills([]string{messy, "clean-code"})
	card.Name, card.Driver, card.Model = messy, messy, messy
	for width := components.CardListMinWidth; width <= 160; width++ {
		out := components.RenderCardItem(card, skills, nil, puppets, false, width)
		assertIntactCard(t, out, fmt.Sprintf("width=%d", width))
		if strings.ContainsAny(ansi.Strip(out), "\t\r") {
			t.Fatalf("width=%d: a control character reached the output: %q", width, ansi.Strip(out))
		}
	}
}
