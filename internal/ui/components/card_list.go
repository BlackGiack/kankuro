package components

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"kankuro/internal/config"
	"kankuro/internal/ui/component"
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
	searchQuery string
	callbacks   CardListCallbacks
	delegate    cardDelegate
}

// NewCardList creates a new card list component.
func NewCardList(callbacks CardListCallbacks) *CardList {
	delegate := newCardDelegate(nil, nil)

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

// SetCards updates the card list data.
func (c *CardList) SetCards(cards []config.IDCard, skills []config.Skill, personas []config.Persona) {
	c.cards = cards
	c.skills = skills
	c.personas = personas
	c.delegate = newCardDelegate(skills, personas)
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
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		keyCmd := c.handleKey(keyMsg)
		return c, tea.Batch(cmd, keyCmd)
	}

	return c, cmd
}

func (c *CardList) handleKey(keyMsg tea.KeyMsg) tea.Cmd {
	//nolint:exhaustive // not all keys need to be handled
	switch keyMsg.Type {
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
	case tea.KeyRunes:
		if keyMsg.String() == "e" {
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
	if c.searchQuery == "" {
		return c.cards
	}

	q := strings.ToLower(c.searchQuery)
	var filtered []config.IDCard
	for _, card := range c.cards {
		if strings.Contains(strings.ToLower(card.Name), q) ||
			strings.Contains(strings.ToLower(card.Driver), q) ||
			strings.Contains(strings.ToLower(card.Model), q) {
			filtered = append(filtered, card)
		}
	}
	return filtered
}

// cardDelegate handles rendering custom card items.
type cardDelegate struct {
	height   int
	skills   []config.Skill
	personas []config.Persona
}

func newCardDelegate(skills []config.Skill, personas []config.Persona) cardDelegate {
	return cardDelegate{
		height:   CardDelegateHeight,
		skills:   skills,
		personas: personas,
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
	content := RenderCardItem(cardItem.Card, d.skills, d.personas, selected, m.Width())
	fmt.Fprint(w, content)
}

// RenderCardItem renders a single card's display.
func RenderCardItem(card config.IDCard, skills []config.Skill, personas []config.Persona, selected bool, listWidth int) string {
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

	puppetStr := RenderAgentPuppet(card.Puppet)

	driverLine := fmt.Sprintf("%s | %s", card.Driver, card.Model)

	// Colors - should come from shared styles
	primaryColor := lipgloss.Color("#8E44AD")
	highlightColor := lipgloss.Color("#16A085")
	gray := lipgloss.Color("240")
	white := lipgloss.Color("#FFFFFF")
	deepCharcoal := lipgloss.Color("#1A1A1A")

	titleStyle := lipgloss.NewStyle().Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(gray)

	if selected {
		titleStyle = titleStyle.Foreground(highlightColor)
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

	faceBg := deepCharcoal
	if selected {
		faceBg = primaryColor
	}

	faceBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(primaryColor).
		Background(deepCharcoal).
		Align(lipgloss.Center, lipgloss.Center)

	faceBox := faceBoxStyle.
		Width(FaceBoxWidth).
		Height(FaceBoxHeight).
		Background(faceBg).
		Foreground(white).
		Render(puppetStr)

	content := lipgloss.JoinHorizontal(lipgloss.Top, infoBox, faceBox)

	borderColor := gray
	if selected {
		borderColor = primaryColor
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
		gray := lipgloss.Color("240")
		return []string{lipgloss.NewStyle().Foreground(gray).Render("(no capabilities)")}
	}

	highlightColor := lipgloss.Color("#16A085")
	personaColor := lipgloss.Color("#C0392B")
	white := lipgloss.Color("#FFFFFF")

	for _, s := range skillNames {
		if len(pills) >= MaxPillsShown {
			break
		}
		pills = append(pills, pillStyle(highlightColor, white).Render(truncatePill(s)))
	}
	for _, s := range personaNames {
		if len(pills) >= MaxPillsShown {
			break
		}
		pills = append(pills, pillStyle(personaColor, white).Render(truncatePill(s)))
	}
	return pills
}

func pillStyle(bg, fg lipgloss.Color) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(fg).
		Background(bg).
		Padding(0, 1).
		MarginRight(1)
}

func truncatePill(s string) string {
	if len(s) > PillTextLimit {
		return s[:PillTextLimit-1] + "."
	}
	return s
}

func truncate(s string, w int) string {
	if len(s) > w && w > 3 {
		return s[:w-3] + "..."
	}
	return s
}
