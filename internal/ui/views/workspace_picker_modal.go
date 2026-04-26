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

func (m *WorkspacePickerModal) Update(msg tea.Msg) (component.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case keyEnter:
			m.confirmed = true
			return m, nil
		case keyEsc:
			m.cancelled = true
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
		Width(workspacePickerDialogWidth)

	return boxStyle.Render(m.list.View())
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
