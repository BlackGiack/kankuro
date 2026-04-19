package views

import (
	"github.com/charmbracelet/bubbles/list"
	bubbletea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"kankuro/internal/config"
	"kankuro/internal/ui/component"
)

// SwitchModal allows selecting a new agent for a terminated session.
type SwitchModal struct {
	component.Base

	list      list.Model
	cards     []config.IDCard
	confirmed bool
	cancelled bool
	open      bool
}

// NewSwitchModal creates a new switch modal.
func NewSwitchModal() *SwitchModal {
	delegate := list.NewDefaultDelegate()
	delegate.SetHeight(4)

	l := list.New([]list.Item{}, delegate, 40, 10)
	l.SetShowTitle(true)
	l.Title = "Select New Agent"
	l.SetShowStatusBar(false)

	return &SwitchModal{
		list: l,
	}
}

// SetCards populates the card selection list.
func (m *SwitchModal) SetCards(cards []config.IDCard, skills []config.Skill, personas []config.Persona) {
	m.cards = cards
	items := make([]list.Item, len(cards))
	for i, card := range cards {
		items[i] = switchCardItem{
			card:     card,
			skills:   skills,
			personas: personas,
		}
	}
	m.list.SetItems(items)
}

// Open shows the modal.
func (m *SwitchModal) Open() {
	m.confirmed = false
	m.cancelled = false
	m.open = true
}

// Close hides the modal.
func (m *SwitchModal) Close() {
	m.open = false
}

// IsOpen returns whether modal is active.
func (m *SwitchModal) IsOpen() bool {
	return m.open
}

// Confirmed returns true if user confirmed selection.
func (m *SwitchModal) Confirmed() bool {
	return m.confirmed
}

// Cancelled returns true if user cancelled.
func (m *SwitchModal) Cancelled() bool {
	return m.cancelled
}

// SelectedCardID returns the selected card ID.
func (m *SwitchModal) SelectedCardID() string {
	if item := m.list.SelectedItem(); item != nil {
		if si, ok := item.(switchCardItem); ok {
			return si.card.ID
		}
	}
	return ""
}

// Init initializes the modal.
func (m *SwitchModal) Init() bubbletea.Cmd {
	return nil
}

// Update handles modal input.
func (m *SwitchModal) Update(msg bubbletea.Msg) (component.Model, bubbletea.Cmd) {
	// Handle modal-level keys first
	if keyMsg, ok := msg.(bubbletea.KeyMsg); ok {
		switch keyMsg.String() {
		case "enter":
			m.confirmed = true
			m.Close()
			return m, nil
		case "esc":
			m.cancelled = true
			m.Close()
			return m, nil
		}
	}

	// Then handle list navigation
	var cmd bubbletea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the modal.
func (m *SwitchModal) View() string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#8E44AD")).
		Padding(1, 2).
		Width(50)

	return boxStyle.Render(m.list.View())
}

// switchCardItem wraps a card for display in switch modal.
type switchCardItem struct {
	card     config.IDCard
	skills   []config.Skill
	personas []config.Persona
}

func (i switchCardItem) Title() string       { return i.card.Name }
func (i switchCardItem) Description() string { return i.card.Driver + " | " + i.card.Model }
func (i switchCardItem) FilterValue() string { return i.card.Name }
