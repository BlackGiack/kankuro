package views

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/BlackGiack/kankuro/internal/ui/component"
	"github.com/BlackGiack/kankuro/internal/ui/components"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/styles"
)

const (
	conversationChoiceItemHeight     = 2
	conversationChoiceMinWidth       = 42
	conversationChoiceMinHeight      = 10
	conversationChoiceListMinWidth   = 30
	conversationChoiceListMinHeight  = 2
	conversationChoiceChromeRows     = 7
	conversationChoiceMinDetailWidth = 20
	conversationChoiceOuterPadding   = 4
)

type ConversationChoiceCallbacks struct {
	OnSelect func(sessionID, cardID string, choice session.ConversationChoice) tea.Cmd
	OnCancel func(returnView messages.ViewType) tea.Cmd
}

type ConversationChoiceView struct {
	component.Base

	list         list.Model
	sessionID    string
	cardID       string
	returnView   messages.ViewType
	driver       string
	bindings     []session.ConversationBinding
	query        string
	contextErr   string
	detailOffset int
	detailRows   int
	detailKey    string
	callbacks    ConversationChoiceCallbacks
}

type conversationChoiceItem struct {
	binding    *session.ConversationBinding
	new        bool
	compatible bool
}

func (i conversationChoiceItem) Title() string {
	if i.new {
		return "New conversation"
	}
	return components.DisplayText(i.binding.Name)
}

func (i conversationChoiceItem) Description() string {
	if i.new {
		return "Start fresh; saved conversations remain available"
	}
	if !i.compatible {
		return fmt.Sprintf("%s | %s | incompatible driver", components.DisplayText(i.binding.ID), components.DisplayText(i.binding.Driver))
	}
	return fmt.Sprintf("%s | %s", components.DisplayText(i.binding.ID), components.DisplayText(i.binding.Driver))
}

func (i conversationChoiceItem) FilterValue() string {
	if i.new {
		return "new conversation"
	}
	return components.DisplayText(
		i.binding.Name,
	) + " " + components.DisplayText(
		i.binding.ID,
	) + " " + components.DisplayText(
		i.binding.Driver,
	)
}

func NewConversationChoiceView(callbacks ConversationChoiceCallbacks) *ConversationChoiceView {
	d := list.NewDefaultDelegate()
	d.SetHeight(conversationChoiceItemHeight)
	l := list.New([]list.Item{}, d, 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	return &ConversationChoiceView{list: l, callbacks: callbacks}
}

func (v *ConversationChoiceView) SetContext(
	sessionID, cardID, driver string,
	returnView messages.ViewType,
	bindings []session.ConversationBinding,
) {
	v.sessionID, v.cardID, v.returnView = sessionID, cardID, returnView
	v.driver = driver
	v.query = ""
	v.bindings = append([]session.ConversationBinding(nil), bindings...)
	v.sortBindings()
	v.contextErr = ""
	v.refreshItems()
}

// RefreshContext updates live card/session state while preserving the selected
// saved identity when it remains present and compatible.
func (v *ConversationChoiceView) RefreshContext(
	sessionID, cardID, driver string,
	bindings []session.ConversationBinding,
) {
	previous := v.selectedChoice()
	v.sessionID, v.cardID = sessionID, cardID
	v.driver = driver
	v.bindings = append([]session.ConversationBinding(nil), bindings...)
	v.sortBindings()
	v.contextErr = ""
	v.refreshItemsPreserving(previous)
}

func (v *ConversationChoiceView) Target() (string, string, messages.ViewType) {
	return v.sessionID, v.cardID, v.returnView
}

func (v *ConversationChoiceView) SelectedBinding() (session.ConversationBinding, bool) {
	item, ok := v.list.SelectedItem().(conversationChoiceItem)
	if !ok || item.binding == nil {
		return session.ConversationBinding{}, false
	}
	return *item.binding, true
}

func (v *ConversationChoiceView) SetContextError(message string) {
	v.contextErr = message
}

func (v *ConversationChoiceView) sortBindings() {
	sort.Slice(v.bindings, func(i, j int) bool {
		if v.bindings[i].Name != v.bindings[j].Name {
			return v.bindings[i].Name < v.bindings[j].Name
		}
		if v.bindings[i].ID != v.bindings[j].ID {
			return v.bindings[i].ID < v.bindings[j].ID
		}
		return v.bindings[i].Driver < v.bindings[j].Driver
	})
}

func (v *ConversationChoiceView) selectedChoice() *conversationChoiceItem {
	item, ok := v.list.SelectedItem().(conversationChoiceItem)
	if !ok {
		return nil
	}
	selected := item
	return &selected
}

func (v *ConversationChoiceView) Init() tea.Cmd { return nil }

func (v *ConversationChoiceView) Update(msg tea.Msg) (component.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyPressMsg); ok {
		if cmd, handled := v.handleKey(key); handled {
			return v, cmd
		}
	}
	var cmd tea.Cmd
	v.list, cmd = v.list.Update(msg)
	v.resizeList()
	return v, cmd
}

func (v *ConversationChoiceView) handleKey(key tea.KeyPressMsg) (tea.Cmd, bool) {
	switch key.Code {
	case tea.KeyPgUp:
		v.scrollDetails(-max(1, v.detailRows))
		return nil, true
	case tea.KeyPgDown:
		v.scrollDetails(max(1, v.detailRows))
		return nil, true
	case tea.KeyEsc:
		if v.query != "" {
			v.query = ""
			v.refreshItems()
			return nil, true
		}
		if v.callbacks.OnCancel != nil {
			return v.callbacks.OnCancel(v.returnView), true
		}
		return nil, true
	case tea.KeyEnter:
		return v.handleEnter(), true
	case tea.KeyBackspace:
		if v.query != "" {
			v.query = v.query[:len(v.query)-1]
			v.refreshItems()
		}
		return nil, true
	default:
		if key.Text != "" {
			v.query += key.Text
			v.refreshItems()
			return nil, true
		}
	}
	return nil, false
}

func (v *ConversationChoiceView) handleEnter() tea.Cmd {
	if v.contextErr != "" {
		return nil
	}
	item, ok := v.list.SelectedItem().(conversationChoiceItem)
	if !ok || v.callbacks.OnSelect == nil || (!item.new && !item.compatible) {
		return nil
	}
	choice := session.ConversationChoice{New: item.new}
	if item.binding != nil {
		choice.ID = item.binding.ID
	}
	cardID := v.cardID
	if v.returnView == messages.ViewDashboard {
		cardID = ""
	}
	return v.callbacks.OnSelect(v.sessionID, cardID, choice)
}

func (v *ConversationChoiceView) refreshItems() {
	v.refreshItemsPreserving(nil)
}

func (v *ConversationChoiceView) refreshItemsPreserving(preferred *conversationChoiceItem) {
	items := v.filteredItems()
	v.list.SetItems(items)
	if !v.selectPreferred(items, preferred) {
		v.selectFirstCompatible(items)
	}
	v.resizeList()
}

func (v *ConversationChoiceView) filteredItems() []list.Item {
	items := make([]list.Item, 0, len(v.bindings)+1)
	for i := range v.bindings {
		item := conversationChoiceItem{binding: &v.bindings[i], compatible: v.bindings[i].Driver == v.driver}
		if v.query == "" || strings.Contains(strings.ToLower(item.FilterValue()), strings.ToLower(v.query)) {
			items = append(items, item)
		}
	}
	return append(items, conversationChoiceItem{new: true})
}

func (v *ConversationChoiceView) selectPreferred(items []list.Item, preferred *conversationChoiceItem) bool {
	if preferred == nil {
		return false
	}
	for i, candidate := range items {
		item, ok := candidate.(conversationChoiceItem)
		if !ok || (!item.new && !item.compatible) || item.new != preferred.new {
			continue
		}
		if item.new || (item.binding.ID == preferred.binding.ID && item.binding.Driver == preferred.binding.Driver) {
			v.list.Select(i)
			return true
		}
	}
	return false
}

func (v *ConversationChoiceView) selectFirstCompatible(items []list.Item) {
	for i, candidate := range items {
		item, ok := candidate.(conversationChoiceItem)
		if ok && (item.new || item.compatible) {
			v.list.Select(i)
			return
		}
	}
}

func (v *ConversationChoiceView) View() string {
	header := styles.TitleStyle.Render("Choose conversation")
	search := v.searchLine()
	helpText := "Enter=choose • Esc=clear filter/cancel"
	fullDetails := v.selectedDetailLines()
	if len(fullDetails) > v.detailRows && v.detailRows > 0 {
		helpText = "↑/↓=choose • PgUp/PgDn=scroll ID • Enter=choose • Esc=cancel"
	} else if len(fullDetails) > 0 {
		helpText = "↑/↓=choose • Enter=choose • Esc=cancel"
	}
	if v.contextErr != "" {
		helpText = components.TruncateTail(
			components.DisplayText(v.contextErr),
			max(1, v.Width-widthPadding*4),
		)
	}
	help := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(helpText)
	selected := ""
	if len(fullDetails) > 0 && v.detailRows > 0 {
		start := min(v.detailOffset, max(0, len(fullDetails)-v.detailRows))
		end := min(len(fullDetails), start+v.detailRows)
		selected = strings.Join(fullDetails[start:end], "\n")
	}
	sections := []string{header, search, v.list.View()}
	if selected != "" {
		sections = append(sections, selected)
	}
	sections = append(sections, help)
	return lipgloss.NewStyle().
		Padding(heightPadding, widthPadding).
		Width(max(conversationChoiceMinWidth, v.Width-conversationChoiceOuterPadding)).
		Height(max(conversationChoiceMinHeight, v.Height)).
		Render(strings.Join(sections, "\n"))
}

func (v *ConversationChoiceView) selectedDetailLines() []string {
	item, ok := v.list.SelectedItem().(conversationChoiceItem)
	if !ok || item.binding == nil {
		return nil
	}
	details := "Selected ID: " + components.DisplayText(item.binding.ID) + " (" + components.DisplayText(item.binding.Driver) + ")"
	wrapped := ansi.Wrap(details, max(conversationChoiceMinDetailWidth, v.Width-widthPadding*4), "")
	return strings.Split(wrapped, "\n")
}

func (v *ConversationChoiceView) scrollDetails(delta int) {
	if len(v.selectedDetailLines()) == 0 || v.detailRows == 0 {
		return
	}
	maxOffset := max(0, len(v.selectedDetailLines())-v.detailRows)
	v.detailOffset = min(maxOffset, max(0, v.detailOffset+delta))
}

func (v *ConversationChoiceView) searchLine() string {
	if v.query == "" {
		return "Type to filter..."
	}
	const prefix = "Filter: "
	width := max(1, v.Width-widthPadding*4-lipgloss.Width(prefix))
	return prefix + components.TruncateTail(components.DisplayText(v.query), width)
}

func (v *ConversationChoiceView) SetSize(width, height int) {
	v.Base.SetSize(width, height)
	v.resizeList()
}

func (v *ConversationChoiceView) resizeList() {
	selectionKey := "new"
	if item, ok := v.list.SelectedItem().(conversationChoiceItem); ok && item.binding != nil {
		selectionKey = item.binding.Driver + "\x00" + item.binding.ID
	}
	if v.detailKey != selectionKey {
		v.detailKey = selectionKey
		v.detailOffset = 0
	}
	detailLines := v.selectedDetailLines()
	availableRows := max(0, v.Height-2*heightPadding-conversationChoiceChromeRows)
	v.detailRows = min(len(detailLines), max(0, availableRows-conversationChoiceListMinHeight))
	listRows := max(conversationChoiceListMinHeight, availableRows-v.detailRows)
	if len(detailLines) == 0 {
		listRows = max(conversationChoiceListMinHeight, availableRows+1)
	}
	maxOffset := max(0, len(detailLines)-v.detailRows)
	v.detailOffset = min(maxOffset, v.detailOffset)
	v.list.SetSize(
		max(conversationChoiceListMinWidth, v.Width-widthPadding*2),
		listRows,
	)
}
