package components

import (
	"fmt"
	"strings"
	"unicode"

	bubbletea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"kankuro/internal/ui/component"
	"kankuro/internal/ui/styles"
)

const (
	cursorCenterDivisor = 2
	focusPrefix         = "> "
	ellipsisRow         = "  ..."

	// minIndicatorSlots is the smallest body that can spare two rows for the
	// "..." indicators and still show an item.
	minIndicatorSlots = 3
)

// CapabilityItem represents an item in a capability list.
type CapabilityItem interface {
	GetID() string
	GetLabel() string
	GetDescription() string
}

// CapabilityListConfig configures a capability list.
type CapabilityListConfig struct {
	Title      string
	SourceHint string // e.g., "ctrl+e=sources"
	// MaxVisible is how many items an unsized list shows. Once the list has a
	// height (SetSize) the height decides.
	MaxVisible int
}

// CapabilityList is a selectable list of capabilities.
type CapabilityList struct {
	component.Base

	title      string
	sourceHint string
	maxVisible int

	items       []CapabilityItem
	cursor      int
	selected    map[string]bool
	SearchQuery string
	focused     bool
}

// NewCapabilityList creates a new capability list.
func NewCapabilityList(config CapabilityListConfig) *CapabilityList {
	return &CapabilityList{
		title:      config.Title,
		sourceHint: config.SourceHint,
		maxVisible: max(1, config.MaxVisible),
		selected:   make(map[string]bool),
	}
}

func (c *CapabilityList) Items() []CapabilityItem { return c.items }

// SetItems updates the list items.
func (c *CapabilityList) SetItems(items []CapabilityItem) {
	c.items = items
}

// IsSelected reports whether id is currently selected.
func (c *CapabilityList) IsSelected(id string) bool {
	return c.selected[id]
}

// Select marks id as selected.
func (c *CapabilityList) Select(id string) {
	c.selected[id] = true
}

// Deselect clears id's selection.
func (c *CapabilityList) Deselect(id string) {
	delete(c.selected, id)
}

// Toggle flips id's selection state.
func (c *CapabilityList) Toggle(id string) {
	if c.selected[id] {
		delete(c.selected, id)
	} else {
		c.selected[id] = true
	}
}

// SetSelected sets the selected IDs.
func (c *CapabilityList) SetSelected(ids []string) {
	c.selected = make(map[string]bool)
	for _, id := range ids {
		c.selected[id] = true
	}
}

// SelectedIDs returns the selected IDs.
func (c *CapabilityList) SelectedIDs() []string {
	ids := make([]string, 0, len(c.selected))
	for _, item := range c.items {
		if c.selected[item.GetID()] {
			ids = append(ids, item.GetID())
		}
	}
	return ids
}

// SetFocused controls focus state.
func (c *CapabilityList) SetFocused(focused bool) {
	c.focused = focused
}

// SetSearchQuery filters the displayed items.
func (c *CapabilityList) SetSearchQuery(query string) {
	c.SearchQuery = query
	c.cursor = 0
}

// Cursor returns the current cursor position.
func (c *CapabilityList) Cursor() int {
	return c.cursor
}

// SetCursor sets the cursor position.
func (c *CapabilityList) SetCursor(pos int) {
	filtered := c.FilteredItems()
	if pos < 0 {
		pos = 0
	}
	if pos >= len(filtered) {
		pos = len(filtered) - 1
	}
	c.cursor = pos
}

// ToggleSelected toggles selection at cursor.
func (c *CapabilityList) ToggleSelected() {
	filtered := c.FilteredItems()
	if c.cursor < 0 || c.cursor >= len(filtered) {
		return
	}

	c.Toggle(filtered[c.cursor].GetID())
}

// SelectedItem returns the item at cursor.
func (c *CapabilityList) SelectedItem() (CapabilityItem, bool) {
	filtered := c.FilteredItems()
	if c.cursor < 0 || c.cursor >= len(filtered) {
		return nil, false
	}
	return filtered[c.cursor], true
}

// Init initializes the component.
func (c *CapabilityList) Init() bubbletea.Cmd {
	return nil
}

// Update handles messages.
func (c *CapabilityList) Update(msg bubbletea.Msg) (component.Model, bubbletea.Cmd) {
	if keyMsg, ok := msg.(bubbletea.KeyPressMsg); ok {
		c.handleKey(keyMsg)
	}
	return c, nil
}

func (c *CapabilityList) handleKey(keyMsg bubbletea.KeyPressMsg) {
	filtered := c.FilteredItems()

	switch keyMsg.Code {
	case bubbletea.KeyUp:
		if c.cursor > 0 {
			c.cursor--
		}
	case bubbletea.KeyDown:
		if c.cursor < len(filtered)-1 {
			c.cursor++
		}
	case bubbletea.KeySpace:
		c.ToggleSelected()
	case bubbletea.KeyBackspace:
		if len(c.SearchQuery) > 0 {
			c.SearchQuery = c.SearchQuery[:len(c.SearchQuery)-1]
			c.cursor = 0
		}
	default:
		if len(keyMsg.Text) > 0 {
			for _, ch := range keyMsg.Text {
				if unicode.IsPrint(ch) && !unicode.IsSpace(ch) {
					c.SearchQuery += string(ch)
					c.cursor = 0
				}
			}
		}
	}
}

// View renders the capability list.
//
// A list that was given a height renders exactly that many lines, with no
// trailing newline, and every element keeps its row whatever the cursor or the
// search does (see viewSized). An unsized list renders what its items need.
func (c *CapabilityList) View() string {
	if c.Height > 0 {
		return c.viewSized()
	}
	return c.viewUnsized()
}

func (c *CapabilityList) renderHeader() string {
	header := c.title
	if c.sourceHint != "" {
		header += " (" + c.sourceHint + ")"
	}
	header = c.fit(header, 0)
	if c.SearchQuery != "" {
		// The end of the query is where the user is typing, so clip its head.
		const label = " | Search: "
		query := DisplayText(c.SearchQuery)
		if c.Width > 0 {
			query = TruncateTail(query, c.Width-lipgloss.Width(header)-len(label))
		}
		header += label + query
		header = c.fit(header, 0)
	}
	return lipgloss.NewStyle().Foreground(styles.Gray).Render(header)
}

func (c *CapabilityList) renderEmpty() string {
	text := c.fit("(no matching items)", len(focusPrefix))
	if c.focused {
		return lipgloss.NewStyle().Foreground(styles.HighlightColor).Bold(true).Render(focusPrefix + text)
	}
	return text
}

func (c *CapabilityList) renderItem(i int, item CapabilityItem) string {
	prefix := "  "
	if c.focused && i == c.cursor {
		prefix = focusPrefix
	}

	mark := " "
	if c.IsSelected(item.GetID()) {
		mark = "x"
	}

	// The prefix and the "[x] " box are fixed; the label gets what is left.
	label := c.fit(item.GetLabel(), len(prefix)+len("[x] "))
	line := fmt.Sprintf("%s[%s] %s", prefix, mark, label)

	if c.focused && i == c.cursor {
		line = lipgloss.NewStyle().Foreground(styles.HighlightColor).Render(line)
	}
	return line
}

// viewUnsized is the layout of a list that has no height: the header, a
// "..." row on each cut edge and up to maxVisible items. Components that are
// built before SetSize runs use it.
func (c *CapabilityList) viewUnsized() string {
	var sb strings.Builder
	sb.WriteString(c.renderHeader() + "\n")

	filtered := c.FilteredItems()
	if len(filtered) == 0 {
		sb.WriteString(c.renderEmpty())
		return sb.String()
	}

	start, end := c.calculateWindow(len(filtered), c.maxVisible)
	if start > 0 {
		sb.WriteString(ellipsisRow + "\n")
	}
	for i := start; i < end; i++ {
		sb.WriteString(c.renderItem(i, filtered[i]) + "\n")
	}
	if end < len(filtered) {
		sb.WriteString(ellipsisRow + "\n")
	}
	return sb.String()
}

// viewSized renders exactly c.Height lines in a layout that does not depend on
// the scroll position or on the search:
//
//	row 0   the header
//	rows 1+ the body, Height-1 rows
//
// The body layout is decided from the unfiltered item count. When every item
// fits, items start on row 1. When they do not (and the body has room for
// them), the first and last body rows belong to the "..." indicators and are
// blank when nothing is cut off on that side. Rows the filtered items do not
// fill are plain empty lines.
func (c *CapabilityList) viewSized() string {
	lines := make([]string, c.Height)
	lines[0] = c.renderHeader()

	slots := c.Height - 1
	if slots <= 0 {
		return lines[0]
	}

	itemRows, firstItem, indicators := slots, 1, false
	if len(c.items) > slots && slots >= minIndicatorSlots {
		itemRows, firstItem, indicators = slots-2, 2, true //nolint:mnd // one indicator row on each side
	}

	filtered := c.FilteredItems()
	if len(filtered) == 0 {
		lines[firstItem] = c.renderEmpty()
		return strings.Join(lines, "\n")
	}

	start, end := c.calculateWindow(len(filtered), itemRows)
	for i := start; i < end; i++ {
		lines[firstItem+i-start] = c.renderItem(i, filtered[i])
	}
	if indicators {
		if start > 0 {
			lines[1] = ellipsisRow
		}
		if end < len(filtered) {
			lines[c.Height-1] = ellipsisRow
		}
	}
	return strings.Join(lines, "\n")
}

// fit makes user text safe for one line of this list and clips it to the
// list's width minus the fixed text already on that line. An unsized list
// (width 0) does not clip.
func (c *CapabilityList) fit(s string, fixed int) string {
	s = DisplayText(s)
	if c.Width <= 0 {
		return s
	}
	return TruncateHead(s, c.Width-fixed)
}

// FilteredItems returns items matching the search query.
func (c *CapabilityList) FilteredItems() []CapabilityItem {
	if c.SearchQuery == "" {
		return c.items
	}

	q := strings.ToLower(c.SearchQuery)
	var filtered []CapabilityItem
	for _, item := range c.items {
		if strings.Contains(strings.ToLower(item.GetLabel()), q) ||
			strings.Contains(strings.ToLower(item.GetDescription()), q) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// calculateWindow returns the window of rows items that keeps the cursor
// roughly centred, as indexes into the filtered items.
func (c *CapabilityList) calculateWindow(total, rows int) (int, int) {
	rows = max(rows, 1)
	half := rows / cursorCenterDivisor // Divide by 2 to center the cursor
	start := max(c.cursor-half, 0)

	end := start + rows
	if end > total {
		end = total
		start = max(end-rows, 0)
	}

	return start, end
}
