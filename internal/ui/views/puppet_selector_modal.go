package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"kankuro/internal/ascii"
	"kankuro/internal/ascii/characters"
	"kankuro/internal/ui/component"
)

// PuppetStyles contains available puppet style options.
var PuppetStyles = []string{"default", "kankuro", "cat", "strawhat", "darth", "totoro"} //nolint:gochecknoglobals // puppet style options

// PuppetSelectorModal allows selecting a puppet style.
type PuppetSelectorModal struct {
	component.Base

	styles    []string
	cursor    int
	prevStyle string
	confirmed bool
	cancelled bool
	open      bool
}

// NewPuppetSelectorModal creates a new puppet selector.
func NewPuppetSelectorModal() *PuppetSelectorModal {
	return &PuppetSelectorModal{
		styles: PuppetStyles,
	}
}

func (m *PuppetSelectorModal) Cursor() int { return m.cursor }

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
	if m.cursor >= 0 && m.cursor < len(PuppetStyles) {
		return PuppetStyles[m.cursor]
	}
	return defaultPuppetStyle
}

func (m *PuppetSelectorModal) Open()           { m.open = true; m.confirmed = false; m.cancelled = false }
func (m *PuppetSelectorModal) Close()          { m.open = false; m.confirmed = false; m.cancelled = false }
func (m *PuppetSelectorModal) IsOpen() bool    { return m.open }
func (m *PuppetSelectorModal) Confirmed() bool { return m.confirmed }
func (m *PuppetSelectorModal) Cancelled() bool { return m.cancelled }

func (m *PuppetSelectorModal) Init() tea.Cmd { return nil }

func (m *PuppetSelectorModal) Update(msg tea.Msg) (component.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
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
		case keyEnter:
			m.confirmed = true
		case keyEsc:
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
		Padding(heightPadding, widthPadding).
		Render("Select Puppet:\n\n" + row + "\n\n← → to navigate, Enter to confirm, Esc=back")
}
