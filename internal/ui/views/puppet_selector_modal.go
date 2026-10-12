package views

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/BlackGiack/kankuro/internal/ascii"
	"github.com/BlackGiack/kankuro/internal/ascii/characters"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/ui/component"
	"github.com/BlackGiack/kankuro/internal/ui/components"
)

// PuppetSelectorModal allows selecting a puppet style.
type PuppetSelectorModal struct {
	component.Base
	components.ModalState

	puppets []config.Puppet
	cursor  int

	// cellW, cellH are the art area of every cell, fixed at construction.
	cellW, cellH int
}

// NewPuppetSelectorModal creates a new puppet selector.
func NewPuppetSelectorModal() *PuppetSelectorModal {
	cellW, cellH := builtinCellSize()
	return &PuppetSelectorModal{
		puppets: builtinPuppetOptions(),
		cellW:   cellW,
		cellH:   cellH,
	}
}

func builtinPuppetOptions() []config.Puppet {
	defs := characters.BuiltinPuppetLines()
	puppets := make([]config.Puppet, 0, len(defs))
	for _, d := range defs {
		lines := make([]string, len(d.Lines))
		copy(lines, d.Lines)
		puppets = append(puppets, config.Puppet{
			Name:    d.Name,
			Builtin: true,
			Lines:   lines,
		})
	}
	return puppets
}

func (m *PuppetSelectorModal) SetPuppets(puppets []config.Puppet) {
	if len(puppets) == 0 {
		m.puppets = builtinPuppetOptions()
	} else {
		m.puppets = puppets
	}
	if m.cursor >= len(m.puppets) {
		m.cursor = 0
	}
}

func (m *PuppetSelectorModal) Cursor() int { return m.cursor }

func (m *PuppetSelectorModal) SetSelected(name string) {
	for i, p := range m.puppets {
		if p.Name == name {
			m.cursor = i
			return
		}
	}
	m.cursor = 0
}

func (m *PuppetSelectorModal) Selected() string {
	if m.cursor >= 0 && m.cursor < len(m.puppets) {
		return m.puppets[m.cursor].Name
	}
	return defaultPuppetStyle
}

func (m *PuppetSelectorModal) Init() tea.Cmd { return nil }

func (m *PuppetSelectorModal) Update(msg tea.Msg) (component.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch keyMsg.Code {
		case tea.KeyLeft:
			if m.cursor > 0 {
				m.cursor--
			} else {
				m.cursor = len(m.puppets) - 1
			}
		case tea.KeyRight:
			if m.cursor < len(m.puppets)-1 {
				m.cursor++
			} else {
				m.cursor = 0
			}
		case tea.KeyEnter:
			m.Confirm()
		case tea.KeyEsc:
			m.Cancel()
		default:
			return m, nil
		}
	}
	return m, nil
}

// Layout of the selector. Every puppet is drawn in a cell of the same size -
// the largest built-in puppet's - with a border ring that is only visible on
// the selected one, so moving the cursor never shifts the others. Cells that
// do not fit the modal's width scroll horizontally, with an arrow at the edge
// that has more.
const (
	puppetArrowWidth = 2 // "‹ " / " ›"
	puppetArrows     = 2 // the left and the right one
	puppetCellRing   = 2 // the border ring around a cell
	puppetBoxChrome  = 2 + 2*widthPadding
)

const puppetHint = "← → to navigate, Enter to confirm, Esc=back"

// builtinCellSize is the largest built-in puppet's bounds: the art area of
// every cell.
func builtinCellSize() (int, int) {
	var cw, ch int
	for _, d := range characters.BuiltinPuppetLines() {
		w, h := characters.NewPuppet(func(c *ascii.Config) { c.Set("style", d.Name) }).Bounds()
		cw, ch = max(cw, w), max(ch, h)
	}
	return cw, ch
}

// puppetArt returns a puppet's art. User puppets come from files, so their
// lines are made safe for one row each and are clipped by the caller.
func puppetArt(p config.Puppet) string {
	if p.Builtin {
		puppet := characters.NewPuppet(func(c *ascii.Config) { c.Set("style", p.Name) })
		w, h := puppet.Bounds()
		canvas := ascii.NewCanvas(w, h)
		puppet.Render(canvas, 0, 0)
		return canvas.String()
	}
	lines := make([]string, len(p.Lines))
	for i, l := range p.Lines {
		lines[i] = components.DisplayText(l)
	}
	return strings.Join(lines, "\n")
}

// padBlock pads a block with spaces on the right and blank lines at the bottom
// to exactly w columns and h lines. The block must already be clipped to them.
func padBlock(block string, w, h int) string {
	lines := strings.Split(block, "\n")
	for len(lines) < h {
		lines = append(lines, "")
	}
	for i, l := range lines {
		lines[i] = l + strings.Repeat(" ", max(0, w-ansi.StringWidth(l)))
	}
	return strings.Join(lines, "\n")
}

// renderCell draws one puppet in its fixed cell, with a visible border only
// when selected. Art is clipped to the cell, never wrapped (invariant 9).
func (m *PuppetSelectorModal) renderCell(p config.Puppet, selected bool) string {
	art := padBlock(components.ClipBlock(puppetArt(p), m.cellW, m.cellH), m.cellW, m.cellH)

	style := lipgloss.NewStyle().Border(lipgloss.HiddenBorder())
	if selected {
		style = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#8E44AD"))
	}
	return style.Render(art)
}

// visibleWindow returns the first puppet shown and how many are shown. The
// whole row is shown when it fits; otherwise as many cells as fit beside the
// two arrows, with the cursor's cell kept as near the middle as the ends allow.
func (m *PuppetSelectorModal) visibleWindow(cell int) (int, int) {
	n := len(m.puppets)
	avail := m.Width - puppetBoxChrome
	if m.Width <= 0 || n*cell <= avail {
		return 0, n
	}
	count := min(max(1, (avail-puppetArrows*puppetArrowWidth)/cell), n)
	first := min(max(m.cursor-count/2, 0), n-count)
	return first, count
}

// arrowColumn is a column of the row's height with arrow on its middle line.
func arrowColumn(arrow string, height int, show bool) string {
	lines := make([]string, height)
	for i := range lines {
		lines[i] = strings.Repeat(" ", puppetArrowWidth)
	}
	if show {
		lines[height/2] = arrow
	}
	return strings.Join(lines, "\n")
}

func (m *PuppetSelectorModal) View() string {
	cellOuterW, cellOuterH := m.cellW+puppetCellRing, m.cellH+puppetCellRing

	first, count := m.visibleWindow(cellOuterW)
	cells := make([]string, 0, count+puppetArrows)
	scrolls := count < len(m.puppets)
	if scrolls {
		cells = append(cells, arrowColumn("‹ ", cellOuterH, first > 0))
	}
	for i := first; i < first+count; i++ {
		cells = append(cells, m.renderCell(m.puppets[i], i == m.cursor))
	}
	if scrolls {
		cells = append(cells, arrowColumn(" ›", cellOuterH, first+count < len(m.puppets)))
	}
	row := lipgloss.JoinHorizontal(lipgloss.Top, cells...)

	hint := puppetHint
	if m.Width > 0 {
		hint = components.TruncateHead(hint, m.Width-puppetBoxChrome)
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#8E44AD")).
		Padding(heightPadding, widthPadding).
		Render("Select Puppet:\n\n" + row + "\n\n" + hint)
}
