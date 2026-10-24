package components_test

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/BlackGiack/kankuro/internal/ui/components"
	"github.com/BlackGiack/kankuro/internal/ui/uitest"
)

// Contract (step 0 of impl_plan_2.md): SessionPreview.View() is BOUNDED. The
// header block is as wide as its text and the scrollback block below it is
// exactly the preview width, so the render is not a uniform rectangle; the
// dashboard pane pads it. What the preview owes its parent is: no line wider
// than the width it was given and no more lines than the height it was given.
// It owes that by itself - Dashboard.renderPreviewPane also clamps, but that
// clamp is a backstop that must not be what makes the preview fit (see
// TestSessionPreviewFitsWithoutParentClamp).
//
// Region contract: below the header (name, status badge, ID, WS, Started,
// Card, blank, side by side with the puppet) comes a full-width rule, then the
// tail of the scrollback. When the pane is tall enough to show the body, that
// rule must be on the screen and the last meaningful line must be visible.
//
// Seen to fail: with the clampToSize call removed from renderPreview the
// sweep goes red (long lines and oversized puppets spill) while the dashboard
// would still look fine - see TestSessionPreviewFitsWithoutParentClamp.

// previewBodyMinHeight is the height from which the header (at most 7 rows)
// leaves room for a scrollback rule and at least one content row.
const previewBodyMinHeight = 12

// The narrowest preview the dashboard hands its centre pane on the smallest
// terminal it supports (80 columns) is 12: the pane is 14 columns and its
// border takes 2. previewWidths and previewHeights hit every column near the smallest
// supported size and the boundaries where header text starts to truncate,
// then step coarsely; content variety matters more than size density.
func previewWidths() []int {
	return []int{12, 13, 14, 15, 16, 17, 19, 20, 24, 31, 38, 39, 60, 93, 126, 160}
}

func previewHeights() []int {
	return []int{5, 6, 8, 11, 12, 13, 14, 19, 20, 24, 30, 49}
}

func manyLines(n int) []byte {
	var b strings.Builder
	for i := range n {
		fmt.Fprintf(&b, "line %d done\n", i)
	}
	return []byte(b.String())
}

// previewContents is the adversarial scrollback set. tail is the text the
// preview should show last, where there is one ("" = no check).
type previewContent struct {
	name string
	data []byte
	tail string
}

func previewContents() []previewContent {
	return []previewContent{
		{"plain", []byte("hello world\nsecond line here\nthe last line\n"), "the last line"},
		{"wide-runes-at-edge", []byte(strings.Repeat("漢字", 120) + "\n" + strings.Repeat("🚀 ab ", 80) + "\nlast words\n"), "last words"},
		{"10000-col-line", []byte(strings.Repeat("abcdefghij", 1000) + "\nlast words\n"), "last words"},
		{"cr-overwrite", []byte("progress 10%\rprogress 50%\rprogress 100% done\nlast words\n"), "last words"},
		{"cursor-movement", []byte("\x1b[2J\x1b[H\x1b[5;75Hhello world\x1b[1A\x1b[Kmore text here\nlast words\n"), "last words"},
		{
			"osc8-hyperlink",
			[]byte("\x1b]8;;https://example.com/a/very/long/link/target\x1b\\click here for text\x1b]8;;\x1b\\ trailing\nlast words\n"),
			"last words",
		},
		{"unterminated-csi", []byte("some visible text\nlast words\n\x1b[38;2;12"), ""},
		{"unterminated-osc", []byte("some visible text\nlast words\n\x1b]8;;https://example.com/never-ends"), ""},
		{
			"ansi-heavy",
			[]byte(
				"\x1b[1;31mred bold\x1b[0m \x1b[48;5;21mblue bg\x1b[0m " + strings.Repeat("\x1b[32mgreen\x1b[0m ", 40) + "\nlast words\n",
			),
			"last words",
		},
		{"5000-lines", manyLines(5000), "line 4999"},
		{"empty", nil, ""},
	}
}

func previewContentByName(t *testing.T, name string) previewContent {
	t.Helper()
	for _, c := range previewContents() {
		if c.name == name {
			return c
		}
	}
	t.Fatalf("no preview content named %q", name)
	return previewContent{}
}

func previewSnapshot(name string) *session.Snapshot {
	return &session.Snapshot{
		ID:         "session-0123456789-abcdef-0123456789-abcdef",
		Name:       name,
		Status:     session.StatusRunning,
		WorkingDir: "/home/user/projects/a-rather-long-directory-name/with/segments",
		StartedAt:  time.Date(2026, 1, 1, 13, 14, 15, 0, time.UTC),
	}
}

var (
	builtinOnce sync.Once       //nolint:gochecknoglobals // test cache: building the repo writes files
	builtinList []config.Puppet //nolint:gochecknoglobals // test cache, see builtinOnce
)

func cachedBuiltinPuppets(t *testing.T) []config.Puppet {
	t.Helper()
	builtinOnce.Do(func() { builtinList = builtinPuppets(t) })
	return builtinList
}

func newPreview(t *testing.T, snap *session.Snapshot, puppet string, extra ...config.Puppet) *components.SessionPreview {
	t.Helper()
	p := components.NewSessionPreview()
	p.SetPuppets(append(append([]config.Puppet(nil), cachedBuiltinPuppets(t)...), extra...))
	var card *config.IDCard
	if puppet != "" {
		card = &config.IDCard{ID: "c1", Name: "Master mind", Puppet: puppet}
	}
	p.SetSession(snap, card)
	return p
}

func TestSessionPreviewLayoutSweep(t *testing.T) {
	user := uitest.UserPuppet(60, 20)
	for _, content := range previewContents() {
		for _, puppet := range []string{"", "cat", user.Name} {
			p := newPreview(t, previewSnapshot("my session"), puppet, user)
			p.SetScrollback(content.data)
			bodyMin := previewBodyMinHeight
			if puppet == user.Name {
				bodyMin += 20 // the header is as tall as the puppet
			}
			for _, w := range previewWidths() {
				for _, h := range previewHeights() {
					p.SetSize(w, h)
					label := fmt.Sprintf("content=%s puppet=%q %dx%d", content.name, puppet, w, h)
					out := p.View()
					uitest.AssertWithin(t, out, w, h, label)
					if h >= bodyMin {
						assertPreviewBody(t, out, content, w, label)
					}
				}
			}
		}
	}
}

// assertPreviewBody is the region check: the scrollback rule is on screen, and
// the newest meaningful line is visible below it.
func assertPreviewBody(t *testing.T, out string, content previewContent, w int, label string) {
	t.Helper()
	rows := strings.Split(ansi.Strip(out), "\n")
	rule := -1
	for i, r := range rows {
		if strings.TrimSpace(r) != "" && strings.Trim(r, "─") == "" && ansi.StringWidth(r) == w {
			rule = i
			break
		}
	}
	if rule < 0 {
		t.Errorf("%s: no full-width scrollback rule in view:\n%s", label, ansi.Strip(out))
		return
	}
	if content.tail != "" && w >= 30 && !strings.Contains(strings.Join(rows[rule+1:], "\n"), content.tail) {
		t.Errorf("%s: newest line %q not visible below the rule:\n%s", label, content.tail, ansi.Strip(out))
	}
}

func TestSessionPreviewAdversarialSessionNames(t *testing.T) {
	for _, name := range uitest.AdversarialNames {
		for _, w := range previewWidths() {
			for _, h := range []int{5, 12, 19, 40} {
				p := newPreview(t, previewSnapshot(name), "cat")
				p.SetScrollback([]byte("hello world\nlast words\n"))
				p.SetSize(w, h)
				label := fmt.Sprintf("name=%q %dx%d", name, w, h)
				out := p.View()
				uitest.AssertWithin(t, out, w, h, label)
			}
		}
	}
}

// TestSessionPreviewHeaderKeepsItsRows: whatever the scrollback holds, the
// first rows stay the title block - name, status badge, ID, WS, Started.
func TestSessionPreviewHeaderKeepsItsRows(t *testing.T) {
	prefixes := []string{"my session", " Running", "ID: ", "WS: ", "Started: "}
	for _, content := range previewContents() {
		p := newPreview(t, previewSnapshot("my session"), "cat")
		p.SetScrollback(content.data)
		for _, w := range []int{40, 80, 120} {
			p.SetSize(w, 30)
			rows := strings.Split(ansi.Strip(p.View()), "\n")
			for i, prefix := range prefixes {
				if !strings.HasPrefix(rows[i], prefix) {
					t.Errorf("content=%s width=%d: row %d = %q, want prefix %q", content.name, w, i, rows[i], prefix)
				}
			}
		}
	}
}

// TestSessionPreviewEmptyStates covers the two views that are not a preview of
// anything: no session at all (the placeholder) and a session with no output.
func TestSessionPreviewEmptyStates(t *testing.T) {
	for _, w := range previewWidths() {
		for _, h := range previewHeights() {
			label := fmt.Sprintf("%dx%d", w, h)

			none := components.NewSessionPreview()
			none.SetSize(w, h)
			uitest.AssertWithin(t, none.View(), w, h, "placeholder "+label)

			quiet := newPreview(t, previewSnapshot("quiet"), "cat")
			quiet.SetSize(w, h)
			out := quiet.View()
			uitest.AssertWithin(t, out, w, h, "no output "+label)
			if h >= previewBodyMinHeight && w >= 20 && !strings.Contains(ansi.Strip(out), "(no output yet)") {
				t.Errorf("no output %s: missing the empty-output notice:\n%s", label, ansi.Strip(out))
			}
		}
	}
}

func TestSessionPreviewNarrowPlaceholderKeepsReadableWords(t *testing.T) {
	preview := components.NewSessionPreview()
	preview.SetSize(8, 19)
	out := preview.View()
	uitest.AssertWithin(t, out, 8, 19, "narrow empty preview")
	plain := ansi.Strip(out)
	for _, word := range []string{"session", "selected"} {
		if !strings.Contains(plain, word) {
			t.Errorf("eight-column placeholder should preserve %q as a whole word:\n%s", word, plain)
		}
	}
}

// TestSessionPreviewFitsWithoutParentClamp makes the lesson of §5 executable:
// SessionPreview must fit by itself, not because the dashboard clamps its
// result afterwards. It feeds the content most likely to overflow (an
// oversized user puppet, long header text) straight into View().
func TestSessionPreviewFitsWithoutParentClamp(t *testing.T) {
	user := uitest.UserPuppet(120, 40)
	p := newPreview(t, previewSnapshot(strings.Repeat("w", 300)), user.Name, user)
	p.SetScrollback(previewContentByName(t, "10000-col-line").data)
	for _, w := range previewWidths() {
		for _, h := range previewHeights() {
			p.SetSize(w, h)
			uitest.AssertWithin(t, p.View(), w, h, fmt.Sprintf("oversized puppet %dx%d", w, h))
		}
	}
}

func TestSessionPreviewReservesStablePuppetColumn(t *testing.T) {
	marker := config.Puppet{ID: "marker", Name: "marker", Lines: []string{"PUPPET", "art"}}
	p := newPreview(t, previewSnapshot("review"), "marker", marker)
	p.SetSize(42, 30)
	out := ansi.Strip(p.View())
	firstRow, _, _ := strings.Cut(out, "\n")
	textBeforePuppet, _, foundPuppet := strings.Cut(firstRow, "PUPPET")
	if !foundPuppet {
		t.Fatalf("complete six-cell puppet missing:\n%s", out)
	}
	if got, want := ansi.StringWidth(textBeforePuppet), 36; got != want {
		t.Fatalf("puppet starts at cell %d, want stable text slot+gap %d:\n%s", got, want, out)
	}
	if got := ansi.StringWidth(firstRow); got != 42 {
		t.Fatalf("composed header row width = %d, want 42:\n%s", got, out)
	}
}

func TestSessionPreviewFitsUnicodeMetadataAndOmitsUnfittingArt(t *testing.T) {
	wideArt := uitest.UserPuppet(60, 20)
	snapshot := previewSnapshot(strings.Repeat("名", 30))
	snapshot.ID = strings.Repeat("0123456789abcdef", 4)
	snapshot.WorkingDir = "/workspace/" + strings.Repeat("very-long-directory/", 5) + "プロジェクト/終端"
	card := &config.IDCard{ID: "card", Name: strings.Repeat("審査", 20), Puppet: wideArt.Name}
	preview := newPreview(t, snapshot, card.Puppet, wideArt)
	preview.SetSession(snapshot, card)
	preview.SetSize(42, 30)
	out := ansi.Strip(preview.View())
	rows := strings.Split(out, "\n")
	if strings.Contains(out, "<===") {
		t.Fatalf("oversized art was partially shown instead of omitted:\n%s", out)
	}
	for i, row := range rows[:min(6, len(rows))] {
		if width := ansi.StringWidth(row); width > 42 {
			t.Errorf("metadata row %d is %d cells wide, want <=42: %q", i, width, row)
		}
	}
	if !strings.Contains(rows[0], "...") || !strings.Contains(rows[2], "...") ||
		!strings.HasPrefix(rows[3], "WS: ...") || !strings.HasSuffix(rows[3], "プロジェクト/終端") ||
		!strings.Contains(rows[5], "...") {
		t.Fatalf("long Unicode name, ID, path, and card were not fitted with ellipses:\n%s", out)
	}

	marker := config.Puppet{ID: "marker", Name: "marker", Lines: []string{"PUPPET", "art"}}
	withArt := newPreview(t, snapshot, marker.Name, marker)
	withArt.SetSession(snapshot, &config.IDCard{ID: card.ID, Name: card.Name, Puppet: marker.Name})
	withArt.SetSize(42, 30)
	artRows := strings.Split(ansi.Strip(withArt.View()), "\n")
	if !strings.Contains(artRows[0], "PUPPET") || !strings.Contains(artRows[1], "art") {
		t.Fatalf("fitting art was not preserved intact:\n%s", strings.Join(artRows, "\n"))
	}
	for _, field := range []struct {
		row   int
		label string
	}{{0, "名"}, {2, "ID: "}, {3, "WS: "}, {4, "Started: "}, {5, "Card: "}} {
		row := artRows[field.row]
		if !strings.Contains(row, field.label) || ansi.StringWidth(ansi.Truncate(row, 33, "")) != 33 {
			t.Errorf("metadata row %d did not use the reserved 33-cell text slot: %q", field.row, row)
		}
	}
	if !strings.Contains(artRows[0], "...") || !strings.Contains(artRows[2], "...") ||
		!strings.HasPrefix(artRows[3], "WS: ...") || !strings.Contains(artRows[5], "...") {
		t.Fatalf("art reservation did not ellipsis-fit long metadata in the same text slot:\n%s", strings.Join(artRows, "\n"))
	}

	preview.SetSize(32, 30)
	narrow := ansi.Strip(preview.View())
	if strings.Contains(narrow, "<===") || strings.Contains(narrow, "|################") {
		t.Fatalf("narrow layout should omit the oversized puppet entirely:\n%s", narrow)
	}
	uitest.AssertWithin(t, narrow, 32, 30, "narrow metadata-only layout")
}

func TestSessionPreviewKeepsBuiltinPuppetIntactWhenItFits(t *testing.T) {
	puppets := cachedBuiltinPuppets(t)
	want := ansi.Strip(components.RenderAgentPuppet("kankuro", puppets))
	preview := newPreview(t, previewSnapshot("ordinary session"), "kankuro")
	preview.SetSize(42, 30)
	out := ansi.Strip(preview.View())
	rows, artRows := strings.Split(out, "\n"), strings.Split(want, "\n")
	for i, art := range artRows {
		if i >= len(rows) || !strings.Contains(rows[i], art) {
			t.Fatalf("ordinary puppet row %d was clipped:\nwant %q\ngot:\n%s", i, art, out)
		}
	}
}
