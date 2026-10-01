package views

import (
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"kankuro/internal/config"
	"kankuro/internal/ui/component"
	"kankuro/internal/ui/components"
)

// WorkspacePickerModal allows selecting a workspace.
type WorkspacePickerModal struct {
	component.Base
	components.ModalState

	list        list.Model
	workspaces  []config.Workspace
	selectedID  string
	searchQuery string
	modalWidth  int
	modalHeight int
	listWidth   int
}

// NewWorkspacePickerModal creates a new workspace picker.
func NewWorkspacePickerModal() *WorkspacePickerModal {
	delegate := list.NewDefaultDelegate()
	delegate.SetHeight(workspacePickerDelegateHeight)

	l := list.New([]list.Item{}, delegate, 40, 30) //nolint:mnd // responsive sizing
	l.SetShowTitle(true)
	l.Title = "Select Workspace"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false) // Use manual filtering

	m := &WorkspacePickerModal{
		list:        l,
		modalWidth:  40, //nolint:mnd // responsive sizing
		modalHeight: 30, //nolint:mnd // responsive sizing
	}
	m.applyListSize()
	return m
}

// applyListSize sizes the list to fit within the box and leaves room for the
// search row and the legend. listWidth is the width View gives the list's block.
func (m *WorkspacePickerModal) applyListSize() {
	m.listWidth = max(1, m.modalWidth-widthPadding*2-workspacePickerListWidthPadding)
	listHeight := max(1, m.modalHeight-heightPadding*2-workspacePickerListHeightPadding)
	m.list.SetSize(m.listWidth, listHeight)
}

// resetTransientState clears all transient state that should not survive between openings.
func (m *WorkspacePickerModal) resetTransientState() {
	m.searchQuery = ""
	// Reset the list to clear any stale cursor/selection/paginator state
	m.rebuildList()
}

// rebuildList recreates the list.Model to ensure clean state.
func (m *WorkspacePickerModal) rebuildList() {
	delegate := list.NewDefaultDelegate()
	delegate.SetHeight(workspacePickerDelegateHeight)

	items := make([]list.Item, len(m.workspaces))
	for i, ws := range m.workspaces {
		items[i] = workspaceItem{workspace: ws}
	}

	l := list.New(items, delegate, 40, 30) //nolint:mnd // responsive sizing
	l.SetShowTitle(true)
	l.Title = "Select Workspace"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false) // Use manual filtering

	m.list = l

	// Restore the current size to preserve terminal dimensions
	if m.modalWidth > 0 && m.modalHeight > 0 {
		m.applyListSize()
	}

	// Restore selection if possible
	m.SetSelectedID(m.selectedID)
}

func (m *WorkspacePickerModal) SetWorkspaces(workspaces []config.Workspace) {
	m.workspaces = workspaces
	m.refreshItems()
}

func (m *WorkspacePickerModal) SetSelectedID(id string) {
	m.selectedID = id
	if id == "" {
		return
	}

	// Find the workspace in the current list items (respects filtering)
	items := m.list.Items()
	for i, item := range items {
		if wi, ok := item.(workspaceItem); ok {
			if wi.workspace.ID == id {
				m.list.Select(i)
				return
			}
		}
	}
	// If not found in visible items, select first item if available
	if len(items) > 0 {
		m.list.Select(0)
	}
}

func (m *WorkspacePickerModal) SelectedID() string {
	if item := m.list.SelectedItem(); item != nil {
		if wi, ok := item.(workspaceItem); ok {
			return wi.workspace.ID
		}
	}
	return ""
}

func (m *WorkspacePickerModal) Open() {
	m.ModalState.Open()
	// Reset transient state to ensure clean state on reopen
	m.resetTransientState()
}

func (m *WorkspacePickerModal) Init() tea.Cmd { return nil }

func (m *WorkspacePickerModal) SetSize(width, height int) {
	m.Base.SetSize(width, height)

	// Never larger than the space given, or Place cannot centre it and the
	// editor's view stops being the size of its box.
	m.modalWidth = min(width, max(40, width*2/3)) //nolint:mnd // responsive sizing
	m.modalHeight = height

	m.applyListSize()
}

func (m *WorkspacePickerModal) Update(msg tea.Msg) (component.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch keyMsg.Code {
		case tea.KeyEnter:
			m.Confirm()
			// Save the currently selected workspace before resetting state
			m.selectedID = m.SelectedID()
			// Reset transient state before closing to prepare for next open
			m.resetTransientState()
			return m, nil
		case tea.KeyEsc:
			if m.searchQuery != "" {
				m.searchQuery = ""
				m.refreshItems()
				return m, nil
			}
			m.Cancel()
			// Reset transient state before closing to prepare for next open
			m.resetTransientState()
			return m, nil
		case tea.KeyBackspace:
			if len(m.searchQuery) > 0 {
				m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
				m.refreshItems()
			}
			return m, nil
		default:
			if len(keyMsg.Text) > 0 {
				m.searchQuery += keyMsg.Text
				m.refreshItems()
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *WorkspacePickerModal) View() string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#8E44AD")).
		Padding(heightPadding, widthPadding).
		Width(m.modalWidth).
		Height(m.modalHeight)
	pickerHelpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	help := "Enter=select • Esc=back • Type to search"
	if m.searchQuery != "" {
		help = "Enter=select • Esc=clear search • Type to search"
	}

	// The search row is always there, blank until a query is typed, so the
	// legend's row does not depend on whether a search is active.
	searchRow := ""
	if m.searchQuery != "" {
		searchRow = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ADF")).
			Render("Search: " + components.TruncateTail(
				components.DisplayText(m.searchQuery),
				max(0, m.modalWidth-pickerBoxChrome-len("Search: ")),
			))
	}

	// Nothing here is positioned against its neighbours' content. The list is
	// padded to the width it was given (bubbles pads only to its widest line,
	// which changes from page to page) and everything is joined left-aligned,
	// so the legend is left-aligned with the list and sits at a fixed place.
	return boxStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.NewStyle().Width(m.listWidth).Render(m.list.View()),
			"",
			searchRow,
			"",
			"",
			pickerHelpStyle.Render(help),
		),
	)
}

func (m *WorkspacePickerModal) refreshItems() {
	filtered := m.filterWorkspaces()
	items := make([]list.Item, len(filtered))
	for i, ws := range filtered {
		items[i] = workspaceItem{workspace: ws}
	}
	m.list.SetItems(items)
	// Normalize selection after refreshing items
	m.normalizeSelection()
}

func (m *WorkspacePickerModal) filterWorkspaces() []config.Workspace {
	return components.Filter(m.workspaces, m.searchQuery, func(ws config.Workspace) string {
		return ws.Name + " " + ws.Path
	})
}

// normalizeSelection ensures the selected index is valid for the current filtered list.
func (m *WorkspacePickerModal) normalizeSelection() {
	items := m.list.Items()
	if len(items) == 0 {
		return
	}

	currentIndex := m.list.Index()
	if currentIndex < 0 || currentIndex >= len(items) {
		m.list.Select(0)
	}
}

// pickerBoxChrome is the columns the box takes from its width: the border and
// the horizontal padding on both sides.
const pickerBoxChrome = 2 + 2*widthPadding

type workspaceItem struct {
	workspace config.Workspace
}

func (i workspaceItem) Title() string       { return components.DisplayText(i.workspace.Name) }
func (i workspaceItem) Description() string { return components.DisplayText(i.workspace.Path) }
func (i workspaceItem) FilterValue() string { return i.workspace.Name }
