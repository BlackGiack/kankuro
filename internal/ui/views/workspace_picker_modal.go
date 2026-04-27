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
}

// NewWorkspacePickerModal creates a new workspace picker.
func NewWorkspacePickerModal() *WorkspacePickerModal {
	delegate := list.NewDefaultDelegate()
	delegate.SetHeight(workspacePickerDelegateHeight)

	l := list.New([]list.Item{}, delegate, workspacePickerWidth, workspacePickerHeight)
	l.SetShowTitle(true)
	l.Title = "Select Workspace"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false) // Use manual filtering

	return &WorkspacePickerModal{
		list: l,
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

	l := list.New(items, delegate, workspacePickerWidth, workspacePickerHeight)
	l.SetShowTitle(true)
	l.Title = "Select Workspace"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false) // Use manual filtering

	m.list = l

	// Restore the current size to preserve terminal dimensions
	if m.Width > 0 && m.Height > 0 {
		listWidth := max(1, m.Width-widthPadding*2-workspacePickerListWidthPadding)
		listHeight := max(1, m.Height-heightPadding*2-workspacePickerListHeightPadding)
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
	// Adjust list size to fit within the box and leave room for help/search
	listWidth := max(1, width-widthPadding*2-workspacePickerListWidthPadding)
	listHeight := max(1, height-heightPadding*2-workspacePickerListHeightPadding)
	m.list.SetSize(listWidth, listHeight)
}

func (m *WorkspacePickerModal) Update(msg tea.Msg) (component.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		// Intercept 'q' to prevent quitting
		if keyMsg.String() == "q" {
			return m, nil
		}

		switch keyMsg.String() {
		case keyEnter:
			m.confirmed = true
			// Save the currently selected workspace before resetting state
			m.selectedID = m.SelectedID()
			// Reset transient state before closing to prepare for next open
			m.resetTransientState()
			return m, nil
		case keyEsc:
			if m.searchQuery != "" {
				m.searchQuery = ""
				m.refreshItems()
				return m, nil
			}
			m.cancelled = true
			// Reset transient state before closing to prepare for next open
			m.resetTransientState()
			return m, nil
		case keyBackspace:
			if len(m.searchQuery) > 0 {
				m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
				m.refreshItems()
			}
			return m, nil
		}

		// Handle search input
		if len(keyMsg.Runes) > 0 {
			m.searchQuery += string(keyMsg.Runes)
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
		Width(m.Width).
		Height(m.Height)
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
