package views

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/BlackGiack/kankuro/internal/ui/component"
	"github.com/BlackGiack/kankuro/internal/ui/components"
	"github.com/BlackGiack/kankuro/internal/ui/models"
	"github.com/BlackGiack/kankuro/internal/ui/styles"
)

// AgentSwitchCallbacks defines callbacks for agent switch actions.
type AgentSwitchCallbacks struct {
	OnSwitch func(sessionID, cardID, driver string) tea.Cmd
	OnCancel func() tea.Cmd
}

// AgentSwitchView allows selecting a new agent for a terminated session.
type AgentSwitchView struct {
	component.Base

	list        list.Model
	cards       []models.CardRow
	sessionID   string
	sessionName string
	searchQuery string
	modalWidth  int
	modalHeight int

	callbacks AgentSwitchCallbacks
}

// NewAgentSwitchView creates a new agent switch view.
func NewAgentSwitchView(callbacks AgentSwitchCallbacks) *AgentSwitchView {
	delegate := list.NewDefaultDelegate()
	delegate.SetHeight(switchModalDelegateHeight)

	l := list.New([]list.Item{}, delegate, 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	return &AgentSwitchView{
		list:        l,
		callbacks:   callbacks,
		modalWidth:  35, //nolint:mnd // responsive sizing
		modalHeight: 8,  //nolint:mnd // responsive sizing
	}
}

// SetSession sets the target session context.
func (v *AgentSwitchView) SetSession(id, name string) {
	v.sessionID = id
	v.sessionName = name
}

// SetRows replaces switch destinations with focused card rows.
func (v *AgentSwitchView) SetRows(cards []models.CardRow) {
	selectedID := ""
	if item, ok := v.list.SelectedItem().(switchCardItem); ok {
		selectedID = item.card.ID
	}
	v.cards = models.CloneCardRows(cards)
	for index := range v.cards {
		if v.cards[index].FilterKey == "" {
			v.cards[index].FilterKey = v.cards[index].Name + " " + v.cards[index].Driver + " " + v.cards[index].Model
		}
	}
	v.refreshItems()
	if selectedID != "" {
		for index, item := range v.list.Items() {
			if row, ok := item.(switchCardItem); ok && row.card.ID == selectedID {
				v.list.Select(index)
				return
			}
		}
	}
}

func (v *AgentSwitchView) refreshItems() {
	filtered := v.filterCards()
	items := make([]list.Item, len(filtered))
	for i, card := range filtered {
		items[i] = switchCardItem{card: card}
	}
	v.list.SetItems(items)
}

func (v *AgentSwitchView) filterCards() []models.CardRow {
	available := make([]models.CardRow, 0, len(v.cards))
	for _, card := range v.cards {
		if card.Launch.Enabled {
			available = append(available, card)
		}
	}
	return components.Filter(available, v.searchQuery, func(card models.CardRow) string {
		return card.FilterKey
	})
}

// Init initializes the view.
func (v *AgentSwitchView) Init() tea.Cmd {
	return nil
}

// Update handles input.
func (v *AgentSwitchView) Update(msg tea.Msg) (component.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		if cmd, handled := v.handleKey(keyMsg); handled {
			return v, cmd
		}
	}

	var cmd tea.Cmd
	v.list, cmd = v.list.Update(msg)
	return v, cmd
}

func (v *AgentSwitchView) handleKey(keyMsg tea.KeyPressMsg) (tea.Cmd, bool) {
	switch keyMsg.Code {
	case tea.KeyEsc:
		if v.searchQuery != "" {
			v.searchQuery = ""
			v.refreshItems()
			return nil, true
		}
		if v.callbacks.OnCancel != nil {
			return v.callbacks.OnCancel(), true
		}
		return nil, true
	case tea.KeyEnter:
		if item := v.list.SelectedItem(); item != nil {
			if si, ok := item.(switchCardItem); ok && v.callbacks.OnSwitch != nil {
				return v.callbacks.OnSwitch(v.sessionID, si.card.ID, si.card.Driver), true
			}
		}
		return nil, true
	case tea.KeyBackspace:
		if len(v.searchQuery) > 0 {
			v.searchQuery = v.searchQuery[:len(v.searchQuery)-1]
			v.refreshItems()
		}
		return nil, true
	default:
		if len(keyMsg.Text) > 0 {
			if keyMsg.String() == "q" && v.searchQuery == "" {
				// Block the list's default quit binding while in agent-switch view.
				return nil, true
			}
			v.searchQuery += keyMsg.Text
			v.refreshItems()
			return nil, true
		}
	}

	return nil, false
}

// View renders the view.
func (v *AgentSwitchView) View() string {
	var sb strings.Builder

	headerText := "Switch Agent"
	if v.sessionName != "" {
		headerText = fmt.Sprintf("Switch Agent (session: %s)", v.sessionName)
	}
	sb.WriteString(styles.TitleStyle.Render(headerText) + "\n\n")

	var searchDisplay string
	if v.searchQuery != "" {
		searchDisplay = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ADF")).
			Render("Search: " + v.searchQuery)
	} else {
		searchDisplay = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("Type to search...")
	}
	sb.WriteString(searchDisplay + "\n\n")

	sb.WriteString(v.list.View() + "\n\n")

	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	help := "Enter=switch • Esc=cancel"
	if v.searchQuery != "" {
		help = "Enter=switch • Esc=clear search"
	}
	sb.WriteString(helpStyle.Render(help))

	return lipgloss.NewStyle().
		Padding(heightPadding, widthPadding).
		Width(v.modalWidth).
		Height(v.modalHeight).
		Render(sb.String())
}

// SetSize implements component.Model.
func (v *AgentSwitchView) SetSize(width, height int) {
	v.Base.SetSize(width, height)
	v.modalWidth = max(35, width/2) //nolint:mnd // responsive sizing
	v.modalHeight = max(8, height)  //nolint:mnd // responsive sizing
	v.list.SetSize(v.modalWidth-widthPadding*2, v.modalHeight-heightPadding*2-agentSwitchListHeightPadding)
}

// switchCardItem wraps a card for display in switch modal.
type switchCardItem struct {
	card models.CardRow
}

func (i switchCardItem) Title() string       { return i.card.Name }
func (i switchCardItem) Description() string { return i.card.Driver + " | " + i.card.Model }
func (i switchCardItem) FilterValue() string { return i.card.Name }
