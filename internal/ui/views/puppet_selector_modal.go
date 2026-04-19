package views

import (
	bubbletea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"kankuro/internal/ascii"
	"kankuro/internal/ascii/characters"
	"kankuro/internal/ui/component"
)

// puppetStyles contains available puppet style options.
var puppetStyles = []string{"default", "kankuro", "cat", "strawhat", "darth", "totoro"} //nolint:gochecknoglobals // puppet style options

// PuppetSelectorModal allows selecting a puppet style.
type PuppetSelectorModal struct {
	component.Base

	styles    []string
	cursor    int
	prevStyle string
	confirmed bool
	cancelled bool
}

// NewPuppetSelectorModal creates a new puppet selector.
func NewPuppetSelectorModal() *PuppetSelectorModal {
	return &PuppetSelectorModal{
		styles: puppetStyles,
	}
}

func (m *PuppetSelectorModal) SetSelected(style string) {
	m.prevStyle = style
	for i, s := range m.styles {
		if s == style {
			m.cursor = i
			return
		}
	}
	m.cursor = 0
}

func (m *PuppetSelectorModal) Selected() string {
	if m.cursor >= 0 && m.cursor < len(m.styles) {
		return m.styles[m.cursor]
	}
	return "kankuro"
}

func (m *PuppetSelectorModal) Open()           { m.confirmed = false; m.cancelled = false }
func (m *PuppetSelectorModal) Close()          { m.confirmed = false; m.cancelled = false }
func (m *PuppetSelectorModal) IsOpen() bool    { return !m.cancelled }
func (m *PuppetSelectorModal) Confirmed() bool { return m.confirmed }
func (m *PuppetSelectorModal) Cancelled() bool { return m.cancelled }

func (m *PuppetSelectorModal) Init() bubbletea.Cmd { return nil }

func (m *PuppetSelectorModal) Update(msg bubbletea.Msg) (component.Model, bubbletea.Cmd) {
	if keyMsg, ok := msg.(bubbletea.KeyMsg); ok {
		switch keyMsg.String() {
		case "left":
			if m.cursor > 0 {
				m.cursor--
			} else {
				m.cursor = len(m.styles) - 1
			}
		case "right":
			if m.cursor < len(m.styles)-1 {
				m.cursor++
			} else {
				m.cursor = 0
			}
		case "enter":
			m.confirmed = true
		case "esc":
			m.cancelled = true
		}
	}
	return m, nil
}

func (m *PuppetSelectorModal) View() string {
	var puppets []string
	for i, style := range m.styles {
		puppet := characters.NewPuppet(func(c *ascii.Config) { c.Set("style", style) })
		w, h := puppet.Bounds()
		canvas := ascii.NewCanvas(w, h)
		puppet.Render(canvas, 0, 0)
		puppetStr := canvas.String()

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
		Padding(1, 2).
		Render("Select Puppet:\n\n" + row + "\n\n← → to navigate, Enter to confirm")
}
