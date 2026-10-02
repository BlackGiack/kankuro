package views

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"kankuro/internal/ascii"
	"kankuro/internal/ascii/characters"
	"kankuro/internal/config"
	"kankuro/internal/ui/component"
	"kankuro/internal/ui/components"
)

// PuppetSelectorModal allows selecting a puppet style.
type PuppetSelectorModal struct {
	component.Base
	components.ModalState

	puppets []config.Puppet
	cursor  int
}

// NewPuppetSelectorModal creates a new puppet selector.
func NewPuppetSelectorModal() *PuppetSelectorModal {
	return &PuppetSelectorModal{
		puppets: builtinPuppetOptions(),
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

func (m *PuppetSelectorModal) View() string {
	var puppets []string
	for i, p := range m.puppets {
		var puppetStr string
		if p.Builtin {
			puppet := characters.NewPuppet(func(c *ascii.Config) { c.Set("style", p.Name) })
			w, h := puppet.Bounds()
			canvas := ascii.NewCanvas(w, h)
			puppet.Render(canvas, 0, 0)
			puppetStr = canvas.String()
		} else {
			w := maxLineWidthStatic(p.Lines)
			canvas := ascii.NewCanvas(w, len(p.Lines))
			for li, line := range p.Lines {
				canvas.DrawText(0, li, line)
			}
			puppetStr = canvas.String()
		}

		if i == m.cursor {
			puppetStr = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#8E44AD")).
				Render(puppetStr)
		}
		puppets = append(puppets, puppetStr)
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, puppets...)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#8E44AD")).
		Padding(heightPadding, widthPadding).
		Render("Select Puppet:\n\n" + row + "\n\n← → to navigate, Enter to confirm, Esc=back")
}

func maxLineWidthStatic(lines []string) int {
	w := 0
	for _, l := range lines {
		if len(l) > w {
			w = len(l)
		}
	}
	return w
}
