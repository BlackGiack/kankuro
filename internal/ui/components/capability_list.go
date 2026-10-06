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
	viewHeightOffset    = 2 // header and potential footer (ellipsis)
	focusPrefix         = "> "
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
func (c *CapabilityList) View() string {
	var sb strings.Builder

	// Header
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
	sb.WriteString(lipgloss.NewStyle().Foreground(styles.Gray).Render(header) + "\n")

	filtered := c.FilteredItems()

	if len(filtered) == 0 {
		text := c.fit("(no matching items)", len(focusPrefix))
		if c.focused {
			sb.WriteString(lipgloss.NewStyle().Foreground(styles.HighlightColor).Bold(true).Render(focusPrefix + text))
		} else {
			sb.WriteString(text)
		}
		return sb.String()
	}

	// Calculate visible window
	start, end := c.calculateWindow(len(filtered))

	if start > 0 {
		sb.WriteString("  ...\n")
	}

	for i := start; i < end; i++ {
		item := filtered[i]
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

		sb.WriteString(line + "\n")
	}

	if end < len(filtered) {
		sb.WriteString("  ...\n")
	}

	return sb.String()
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

// calculateWindow returns the visible window based on cursor position.
func (c *CapabilityList) calculateWindow(total int) (int, int) {
	half := c.maxVisible / cursorCenterDivisor // Divide by 2 to center the cursor
	start := max(c.cursor-half, 0)

	end := start + c.maxVisible
	if end > total {
		end = total
		start = max(end-c.maxVisible, 0)
	}

	return start, end
}

// SetSize implements component.Model.
func (c *CapabilityList) SetSize(width, height int) {
	c.Base.SetSize(width, height)
	// Clamp to ensure a finite window. We subtract viewHeightOffset to
	// account for the header and a conservative budget for at least one
	// ellipsis ("...") line to prevent overflow.
	c.maxVisible = max(1, height-viewHeightOffset)
}
