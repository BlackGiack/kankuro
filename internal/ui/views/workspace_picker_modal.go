package views

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"kankuro/internal/config"
	"kankuro/internal/ui/component"
)

// WorkspacePickerModal allows selecting a workspace.
type WorkspacePickerModal struct {
	component.Base

	list        list.Model
	workspaces  []config.Workspace
	selectedID  string
	confirmed   bool
	cancelled   bool
	open        bool
	searchQuery string
	modalWidth  int
	modalHeight int
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

	return &WorkspacePickerModal{
		list:        l,
		modalWidth:  40, //nolint:mnd // responsive sizing
		modalHeight: 30, //nolint:mnd // responsive sizing
	}
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
		listWidth := max(1, m.modalWidth-widthPadding*2-workspacePickerListWidthPadding)
		listHeight := max(1, m.modalHeight-heightPadding*2-workspacePickerListHeightPadding)
		m.list.SetSize(listWidth, listHeight)
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
	m.open = true
	m.confirmed = false
	m.cancelled = false
	// Reset transient state to ensure clean state on reopen
	m.resetTransientState()
}

func (m *WorkspacePickerModal) Close() {
	m.open = false
	m.confirmed = false
	m.cancelled = false
}

func (m *WorkspacePickerModal) IsOpen() bool    { return m.open }
func (m *WorkspacePickerModal) Confirmed() bool { return m.confirmed }
func (m *WorkspacePickerModal) Cancelled() bool { return m.cancelled }

func (m *WorkspacePickerModal) Init() tea.Cmd { return nil }

func (m *WorkspacePickerModal) SetSize(width, height int) {
	m.Base.SetSize(width, height)

	m.modalWidth = max(40, width*2/3) //nolint:mnd // responsive sizing
	m.modalHeight = max(15, height)   //nolint:mnd // responsive sizing

	// Adjust list size to fit within the box and leave room for help/search
	listWidth := max(1, m.modalWidth-widthPadding*2-workspacePickerListWidthPadding)
	listHeight := max(1, m.modalHeight-heightPadding*2-workspacePickerListHeightPadding)
	m.list.SetSize(listWidth, listHeight)
}

func (m *WorkspacePickerModal) Update(msg tea.Msg) (component.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.Type { //nolint:exhaustive // the picker only consumes navigation and confirm/cancel keys
		case tea.KeyEnter:
			m.confirmed = true
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
			m.cancelled = true
			// Reset transient state before closing to prepare for next open
			m.resetTransientState()
			return m, nil
		case tea.KeyBackspace:
			if len(m.searchQuery) > 0 {
				m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
				m.refreshItems()
			}
			return m, nil
		case tea.KeyRunes:
			m.searchQuery += keyMsg.String()
			m.refreshItems()
			return m, nil
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

	searchDisplay := ""
	if m.searchQuery != "" {
		searchDisplay = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ADF")).
			Render("\nSearch: " + m.searchQuery)
	}

	return boxStyle.Render(
		lipgloss.JoinVertical(lipgloss.Center,
			m.list.View(),
			searchDisplay,
			"\n",
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
	if m.searchQuery == "" {
		return m.workspaces
	}
	q := strings.ToLower(m.searchQuery)
	var filtered []config.Workspace
	for _, ws := range m.workspaces {
		if strings.Contains(strings.ToLower(ws.Name), q) ||
			strings.Contains(strings.ToLower(ws.Path), q) {
			filtered = append(filtered, ws)
		}
	}
	return filtered
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

type workspaceItem struct {
	workspace config.Workspace
}

func (i workspaceItem) Title() string       { return i.workspace.Name }
func (i workspaceItem) Description() string { return i.workspace.Path }
func (i workspaceItem) FilterValue() string { return i.workspace.Name }
