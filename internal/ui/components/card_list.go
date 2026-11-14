package components

import (
	"fmt"
	"image/color"
	"io"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/BlackGiack/kankuro/internal/ui/component"
	"github.com/BlackGiack/kankuro/internal/ui/models"
	"github.com/BlackGiack/kankuro/internal/ui/styles"
)

const (
	CardDelegateHeight        = 8
	CardListMinWidth          = 46
	CardListMinHeight         = 10
	InfoBoxHeight             = 4
	FaceBoxWidth              = 11
	FaceBoxHeight             = 5
	cardListItemGutter        = 2
	MaxPillsShown             = 4
	PillTextLimit             = 10
	minimumCardItemWidth      = 36
	minimumCardTextWidth      = 20
	cardListHorizontalPadding = 2
)

type CardListItem struct{ Card models.CardRow }

func (c CardListItem) FilterValue() string { return c.Card.FilterKey }

type CardListCallbacks struct {
	OnSelect func(string) tea.Cmd
	OnEdit   func(string) tea.Cmd
	OnDelete func(string)
}

type CardList struct {
	component.Base

	list        list.Model
	cards       []models.CardRow
	searchQuery string
	callbacks   CardListCallbacks
	delegate    cardDelegate
}

func NewCardList(callbacks CardListCallbacks) *CardList {
	delegate := newCardDelegate()
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	return &CardList{list: l, callbacks: callbacks, delegate: delegate}
}

// SetRows publishes complete display rows and preserves the selected identity.
func (c *CardList) SetRows(cards []models.CardRow) {
	selectedID := ""
	if selected, ok := c.Selected(); ok {
		selectedID = selected.ID
	}
	c.cards = models.CloneCardRows(cards)
	c.delegate = newCardDelegate()
	c.list.SetDelegate(c.delegate)
	c.refreshItems()
	if selectedID != "" {
		for index, row := range c.list.Items() {
			if item, ok := row.(CardListItem); ok && item.Card.ID == selectedID {
				c.list.Select(index)
				return
			}
		}
	}
}

func (c *CardList) SetSearchQuery(query string) { c.searchQuery = query; c.refreshItems() }
func (c *CardList) Selected() (models.CardRow, bool) {
	item := c.list.SelectedItem()
	if item == nil {
		return models.CardRow{}, false
	}
	row, ok := item.(CardListItem)
	if !ok {
		return models.CardRow{}, false
	}
	return models.CloneCardRow(row.Card), true
}
func (c *CardList) Init() tea.Cmd { return nil }
func (c *CardList) Update(msg tea.Msg) (component.Model, tea.Cmd) {
	var cmd tea.Cmd
	c.list, cmd = c.list.Update(msg)
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		return c, tea.Batch(cmd, c.handleKey(keyMsg))
	}
	return c, cmd
}
func (c *CardList) handleKey(keyMsg tea.KeyPressMsg) tea.Cmd {
	card, ok := c.Selected()
	if !ok {
		return nil
	}
	switch keyMsg.Code {
	case tea.KeyEnter:
		if c.callbacks.OnSelect != nil {
			return c.callbacks.OnSelect(card.ID)
		}
	case tea.KeyDelete, tea.KeyBackspace:
		if c.callbacks.OnDelete != nil {
			c.callbacks.OnDelete(card.ID)
		}
	default:
		if keyMsg.Text == "e" && c.callbacks.OnEdit != nil {
			return c.callbacks.OnEdit(card.ID)
		}
	}
	return nil
}
func (c *CardList) View() string { return c.list.View() }
func (c *CardList) SetSize(width, height int) {
	c.Base.SetSize(width, height)
	c.list.SetSize(width, height)
}
func (c *CardList) Cards() []models.CardRow         { return models.CloneCardRows(c.cards) }
func (c *CardList) FilteredCards() []models.CardRow { return models.CloneCardRows(c.filterCards()) }
func (c *CardList) Select(index int)                { c.list.Select(index) }
func (c *CardList) refreshItems() {
	filtered := c.filterCards()
	items := make([]list.Item, 0, len(filtered))
	for _, card := range filtered {
		items = append(items, CardListItem{Card: card})
	}
	c.list.SetItems(items)
}
func (c *CardList) filterCards() []models.CardRow {
	return Filter(c.cards, c.searchQuery, func(card models.CardRow) string { return card.FilterKey })
}

type cardDelegate struct{ height int }

func newCardDelegate() cardDelegate                            { return cardDelegate{height: CardDelegateHeight} }
func (d cardDelegate) Height() int                             { return d.height }
func (d cardDelegate) Spacing() int                            { return 0 }
func (d cardDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d cardDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	cardItem, ok := item.(CardListItem)
	if !ok {
		return
	}
	content := RenderCardItem(cardItem.Card, index == m.Index(), m.Width())
	fmt.Fprint(w, lipgloss.NewStyle().PaddingLeft(cardListItemGutter).Render(content))
}

func RenderCardItem(card models.CardRow, selected bool, listWidth int) string {
	const bordersPadding = 2
	cardItemWidth := max(minimumCardItemWidth, listWidth-bordersPadding-cardListHorizontalPadding)
	infoBoxWidth := cardItemWidth - FaceBoxWidth - bordersPadding
	truncateLen := max(minimumCardTextWidth, cardItemWidth-FaceBoxWidth-bordersPadding-1)
	pills := renderPills(card.SkillNames, card.PersonaNames, infoBoxWidth)
	skillsLine := lipgloss.JoinHorizontal(lipgloss.Top, pills...)
	puppetStr := ClipBlock(RenderPuppetArtwork(card.PuppetName, card.PuppetBuiltin, card.PuppetLines), FaceBoxWidth-1, FaceBoxHeight)
	driverLine := fmt.Sprintf("%s | %s", card.Driver, card.Model)
	titleStyle := lipgloss.NewStyle().Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(styles.Gray)
	if selected {
		titleStyle = titleStyle.Foreground(styles.HighlightColor)
	}
	infoLines := []string{
		titleStyle.Render(truncate(DisplayText(card.Name), truncateLen)),
		descStyle.Render(truncate(DisplayText(driverLine), truncateLen)),
		"",
		skillsLine,
	}
	infoBox := lipgloss.NewStyle().Width(infoBoxWidth).MaxWidth(infoBoxWidth).Height(InfoBoxHeight).Render(strings.Join(infoLines, "\n"))
	faceBg := styles.DeepCharcoal
	if selected {
		faceBg = styles.PrimaryColor
	}
	faceBox := lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, false, true).BorderForeground(styles.PrimaryColor).
		Background(styles.DeepCharcoal).Align(lipgloss.Center, lipgloss.Center).Width(FaceBoxWidth).MaxWidth(FaceBoxWidth).
		Height(FaceBoxHeight).Background(faceBg).Foreground(styles.White).Render(puppetStr)
	content := lipgloss.JoinHorizontal(lipgloss.Top, infoBox, faceBox)
	borderColor := styles.Gray
	if selected {
		borderColor = styles.PrimaryColor
	}
	return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(borderColor).Width(cardItemWidth).Render(content)
}

// renderPills renders the capability pills for one card row, fitted to
// maxWidth by measured display width.
//
// The row must never be wrapped by lipgloss: every pill ends in two columns
// of whitespace (padding + margin), and a word-wrap leaves whitespace hanging
// past the box edge instead of moving it to the next line. When the edge
// lands inside those two columns the row comes out 1-2 columns too wide,
// which is what shattered the card for particular skill sets at particular
// terminal widths. So pills are added only while they fit, and whatever does
// not fit - including anything past MaxPillsShown - is summarised in a
// trailing "+N" pill.
func renderPills(skillNames, personaNames []string, maxWidth int) []string {
	if len(skillNames) == 0 && len(personaNames) == 0 {
		return []string{lipgloss.NewStyle().Foreground(styles.Gray).Render("(no capabilities)")}
	}

	type pillSpec struct {
		name string
		bg   color.Color
	}
	specs := make([]pillSpec, 0, len(skillNames)+len(personaNames))
	for _, n := range skillNames {
		specs = append(specs, pillSpec{name: n, bg: styles.HighlightColor})
	}
	for _, n := range personaNames {
		specs = append(specs, pillSpec{name: n, bg: styles.SecondaryColor})
	}

	var pills []string
	var widths []int
	used := 0
	for _, sp := range specs {
		if len(pills) >= MaxPillsShown {
			break
		}
		p := pillStyle(sp.bg, styles.White).Render(truncatePill(DisplayText(sp.name)))
		w := lipgloss.Width(p)
		if used+w > maxWidth {
			break
		}
		pills = append(pills, p)
		widths = append(widths, w)
		used += w
	}

	// Make room for the "+N" summary, dropping pills from the end if needed.
	// Dropping one raises N, which can widen the summary by a digit, so this
	// re-measures on every iteration rather than reserving space up front.
	for hidden := len(specs) - len(pills); hidden > 0; hidden = len(specs) - len(pills) {
		more := overflowPill(hidden)
		if used+lipgloss.Width(more) <= maxWidth {
			pills = append(pills, more)
			break
		}
		if len(pills) == 0 {
			break // not even the summary fits; the info box MaxWidth clips
		}
		used -= widths[len(widths)-1]
		pills, widths = pills[:len(pills)-1], widths[:len(widths)-1]
	}
	return pills
}

// ClipBlock truncates a multi-line block to at most width display columns
// and height lines, without wrapping anything onto a new line.
func ClipBlock(block string, width, height int) string {
	lines := strings.Split(block, "\n")
	if len(lines) > height {
		lines = lines[:height]
	}
	for i, l := range lines {
		lines[i] = ansi.Truncate(l, width, "")
	}
	return strings.Join(lines, "\n")
}

// overflowPill summarises capabilities that did not fit on the card row.
func overflowPill(hidden int) string {
	return pillStyle(styles.Gray, styles.White).Render(fmt.Sprintf("+%d", hidden))
}

func pillStyle(bg, fg color.Color) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(fg).
		Background(bg).
		Padding(0, 1).
		MarginRight(1)
}

func truncatePill(s string) string {
	return TruncateHead(s, PillTextLimit)
}

func truncate(s string, w int) string {
	return TruncateHead(s, w)
}
