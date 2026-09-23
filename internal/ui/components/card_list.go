package components

import (
	"fmt"
	"image/color"
	"io"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"kankuro/internal/config"
	"kankuro/internal/ui/component"
	"kankuro/internal/ui/styles"
)

// Card styling .
const (
	CardDelegateHeight = 8
	CardListMinWidth   = 46
	CardListMinHeight  = 10

	InfoBoxHeight = 4
	FaceBoxWidth  = 11
	FaceBoxHeight = 5

	MaxPillsShown = 4
	PillTextLimit = 10
)

// CardListItem wraps IDCard for display in the list.
type CardListItem struct {
	Card     config.IDCard
	Skills   []config.Skill
	Personas []config.Persona
}

func (c CardListItem) FilterValue() string {
	return c.Card.Name + " " + c.Card.Driver + " " + c.Card.Model
}

// CardListCallbacks defines callbacks for card actions.
type CardListCallbacks struct {
	OnSelect func(config.IDCard) tea.Cmd
	OnEdit   func(config.IDCard) tea.Cmd
	OnDelete func(string)
}

// CardList is a scrollable list of ID cards.
type CardList struct {
	component.Base

	list        list.Model
	cards       []config.IDCard
	skills      []config.Skill
	personas    []config.Persona
	puppets     []config.Puppet
	searchQuery string
	callbacks   CardListCallbacks
	delegate    cardDelegate
}

// NewCardList creates a new card list component.
func NewCardList(callbacks CardListCallbacks) *CardList {
	delegate := newCardDelegate(nil, nil, nil)

	l := list.New([]list.Item{}, delegate, 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	return &CardList{
		list:      l,
		callbacks: callbacks,
		delegate:  delegate,
	}
}

// SetPuppets updates the puppet catalog.
func (c *CardList) SetPuppets(puppets []config.Puppet) {
	c.puppets = puppets
	c.delegate = newCardDelegate(c.skills, c.personas, c.puppets)
	c.list.SetDelegate(c.delegate)
}

// SetCards updates the card list data.
func (c *CardList) SetCards(cards []config.IDCard, skills []config.Skill, personas []config.Persona) {
	c.cards = cards
	c.skills = skills
	c.personas = personas
	c.delegate = newCardDelegate(skills, personas, c.puppets)
	c.list.SetDelegate(c.delegate)
	c.refreshItems()
}

// SetSearchQuery filters the displayed cards.
func (c *CardList) SetSearchQuery(query string) {
	c.searchQuery = query
	c.refreshItems()
}

// Selected returns the currently selected card.
func (c *CardList) Selected() (config.IDCard, bool) {
	item := c.list.SelectedItem()
	if item == nil {
		return config.IDCard{}, false
	}

	cardItem, ok := item.(CardListItem)
	if !ok {
		return config.IDCard{}, false
	}
	return cardItem.Card, true
}

// Init initializes the component.
func (c *CardList) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (c *CardList) Update(msg tea.Msg) (component.Model, tea.Cmd) {
	var cmd tea.Cmd
	c.list, cmd = c.list.Update(msg)

	// Handle selection events
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		keyCmd := c.handleKey(keyMsg)
		return c, tea.Batch(cmd, keyCmd)
	}

	return c, cmd
}

func (c *CardList) handleKey(keyMsg tea.KeyPressMsg) tea.Cmd {
	switch keyMsg.Code {
	case tea.KeyEnter:
		if card, ok := c.Selected(); ok {
			if c.callbacks.OnSelect != nil {
				return c.callbacks.OnSelect(card)
			}
		}
	case tea.KeyDelete, tea.KeyBackspace:
		if card, ok := c.Selected(); ok {
			if c.callbacks.OnDelete != nil {
				c.callbacks.OnDelete(card.ID)
			}
		}
	default:
		if keyMsg.Text == "e" {
			if card, ok := c.Selected(); ok {
				if c.callbacks.OnEdit != nil {
					return c.callbacks.OnEdit(card)
				}
			}
		}
	}
	return nil
}

// View renders the card list.
func (c *CardList) View() string {
	return c.list.View()
}

// SetSize implements component.Model.
func (c *CardList) SetSize(width, height int) {
	c.Base.SetSize(width, height)
	c.list.SetSize(width, height)
}

// Cards returns the underlying cards.
func (c *CardList) Cards() []config.IDCard {
	return c.cards
}

// FilteredCards returns the currently filtered cards.
func (c *CardList) FilteredCards() []config.IDCard {
	return c.filterCards()
}

// Select selects the item at the given index.
func (c *CardList) Select(index int) {
	c.list.Select(index)
}

// refreshItems rebuilds the list items from filtered cards.
func (c *CardList) refreshItems() {
	filtered := c.filterCards()
	items := make([]list.Item, 0, len(filtered))

	for _, card := range filtered {
		skillList := getSkillsForCard(card, c.skills)
		personaList := getPersonasForCard(card, c.personas)

		items = append(items, CardListItem{
			Card:     card,
			Skills:   skillList,
			Personas: personaList,
		})
	}

	c.list.SetItems(items)
}

// filterCards applies the search query filter.
func (c *CardList) filterCards() []config.IDCard {
	return Filter(c.cards, c.searchQuery, func(card config.IDCard) string {
		return card.Name + " " + card.Driver + " " + card.Model
	})
}

// cardDelegate handles rendering custom card items.
type cardDelegate struct {
	height   int
	skills   []config.Skill
	personas []config.Persona
	puppets  []config.Puppet
}

func newCardDelegate(skills []config.Skill, personas []config.Persona, puppets []config.Puppet) cardDelegate {
	return cardDelegate{
		height:   CardDelegateHeight,
		skills:   skills,
		personas: personas,
		puppets:  puppets,
	}
}

func (d cardDelegate) Height() int                             { return d.height }
func (d cardDelegate) Spacing() int                            { return 0 }
func (d cardDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d cardDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	cardItem, ok := item.(CardListItem)
	if !ok {
		return
	}
	selected := index == m.Index()
	content := RenderCardItem(cardItem.Card, d.skills, d.personas, d.puppets, selected, m.Width())
	fmt.Fprint(w, content)
}

// RenderCardItem renders a single card's display.
func RenderCardItem(card config.IDCard, skills []config.Skill, personas []config.Persona,
	puppets []config.Puppet, selected bool, listWidth int) string {
	const bordersPadding = 2
	cardItemWidth := max(36, listWidth-bordersPadding-2) //nolint:mnd // to not stick to borders
	infoBoxWidth := cardItemWidth - FaceBoxWidth - bordersPadding
	truncateLen := max(20, cardItemWidth-FaceBoxWidth-bordersPadding-1) //nolint:mnd // puppet separator

	skillNames := getSelectedNames(card.SkillIDs, func(id string) (string, bool) {
		for _, s := range skills {
			if s.ID == id {
				return s.Name, true
			}
		}
		return "", false
	})

	personaNames := getSelectedNames(card.PersonaIDs, func(id string) (string, bool) {
		for _, p := range personas {
			if p.ID == id {
				return p.Name, true
			}
		}
		return "", false
	})

	pills := renderPills(skillNames, personaNames)
	skillsLine := lipgloss.JoinHorizontal(lipgloss.Top, pills...)

	puppetStr := RenderAgentPuppet(card.Puppet, puppets)

	driverLine := fmt.Sprintf("%s | %s", card.Driver, card.Model)

	titleStyle := lipgloss.NewStyle().Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(styles.Gray)

	if selected {
		titleStyle = titleStyle.Foreground(styles.HighlightColor)
	}

	infoLines := []string{
		titleStyle.Render(truncate(card.Name, truncateLen)),
		descStyle.Render(truncate(driverLine, truncateLen)),
		"",
		skillsLine,
	}

	infoBox := lipgloss.NewStyle().
		Width(infoBoxWidth).
		Height(InfoBoxHeight).
		Render(strings.Join(infoLines, "\n"))

	faceBg := styles.DeepCharcoal
	if selected {
		faceBg = styles.PrimaryColor
	}

	faceBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(styles.PrimaryColor).
		Background(styles.DeepCharcoal).
		Align(lipgloss.Center, lipgloss.Center)

	faceBox := faceBoxStyle.
		Width(FaceBoxWidth).
		Height(FaceBoxHeight).
		Background(faceBg).
		Foreground(styles.White).
		Render(puppetStr)

	content := lipgloss.JoinHorizontal(lipgloss.Top, infoBox, faceBox)

	borderColor := styles.Gray
	if selected {
		borderColor = styles.PrimaryColor
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(cardItemWidth).
		Render(content)
}

// Helper functions

func getSkillsForCard(card config.IDCard, allSkills []config.Skill) []config.Skill {
	if len(card.SkillIDs) == 0 {
		return nil
	}
	selected := make(map[string]struct{}, len(card.SkillIDs))
	for _, id := range card.SkillIDs {
		selected[id] = struct{}{}
	}
	var result []config.Skill
	for _, s := range allSkills {
		if _, ok := selected[s.ID]; ok {
			result = append(result, s)
		}
	}
	return result
}

func getPersonasForCard(card config.IDCard, allPersonas []config.Persona) []config.Persona {
	if len(card.PersonaIDs) == 0 {
		return nil
	}
	selected := make(map[string]struct{}, len(card.PersonaIDs))
	for _, id := range card.PersonaIDs {
		selected[id] = struct{}{}
	}
	var result []config.Persona
	for _, p := range allPersonas {
		if _, ok := selected[p.ID]; ok {
			result = append(result, p)
		}
	}
	return result
}

func getSelectedNames(ids []string, resolver func(string) (string, bool)) []string {
	var names []string
	for _, id := range ids {
		if name, ok := resolver(id); ok {
			names = append(names, name)
		}
	}
	return names
}

func renderPills(skillNames, personaNames []string) []string {
	var pills []string
	if len(skillNames) == 0 && len(personaNames) == 0 {
		return []string{lipgloss.NewStyle().Foreground(styles.Gray).Render("(no capabilities)")}
	}

	personaColor := styles.SecondaryColor

	for _, s := range skillNames {
		if len(pills) >= MaxPillsShown {
			break
		}
		pills = append(pills, pillStyle(styles.HighlightColor, styles.White).Render(truncatePill(s)))
	}
	for _, s := range personaNames {
		if len(pills) >= MaxPillsShown {
			break
		}
		pills = append(pills, pillStyle(personaColor, styles.White).Render(truncatePill(s)))
	}
	return pills
}

func pillStyle(bg, fg color.Color) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(fg).
		Background(bg).
		Padding(0, 1).
		MarginRight(1)
}

func truncatePill(s string) string {
	return truncateToVisibleWidthLeft(s, PillTextLimit)
}

func truncate(s string, w int) string {
	return truncateToVisibleWidthLeft(s, w)
}
