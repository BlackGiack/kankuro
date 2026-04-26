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
	l.SetFilteringEnabled(false)

	return &WorkspacePickerModal{
		list: l,
	}
}

func (m *WorkspacePickerModal) SetWorkspaces(workspaces []config.Workspace) {
	m.workspaces = workspaces
	m.refreshItems()
}

func (m *WorkspacePickerModal) SetSelectedID(id string) {
	m.selectedID = id
	for i, ws := range m.workspaces {
		if ws.ID == id {
			m.list.Select(i)
			return
		}
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
	m.list.SetSize(width-widthPadding*2-4, height-heightPadding*2-8)
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
			return m, nil
		case keyEsc:
			if m.searchQuery != "" {
				m.searchQuery = ""
				m.refreshItems()
				return m, nil
			}
			m.cancelled = true
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

	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
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
		m.list.View() +
			searchDisplay +
			"\n\n" +
			helpStyle.Render(help),
	)
}

func (m *WorkspacePickerModal) refreshItems() {
	filtered := m.filterWorkspaces()
	items := make([]list.Item, len(filtered))
	for i, ws := range filtered {
		items[i] = workspaceItem{workspace: ws}
	}
	m.list.SetItems(items)
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

type workspaceItem struct {
	workspace config.Workspace
}

func (i workspaceItem) Title() string       { return i.workspace.Name }
func (i workspaceItem) Description() string { return i.workspace.Path }
func (i workspaceItem) FilterValue() string { return i.workspace.Name }
