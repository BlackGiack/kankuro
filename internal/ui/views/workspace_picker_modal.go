package views

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
	bubbletea "github.com/charmbracelet/bubbletea"
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
	searchQuery string
}

// NewWorkspacePickerModal creates a new workspace picker.
func NewWorkspacePickerModal() *WorkspacePickerModal {
	delegate := list.NewDefaultDelegate()
	delegate.SetHeight(3)

	l := list.New([]list.Item{}, delegate, 50, 10)
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
	m.confirmed = false
	m.cancelled = false
}

func (m *WorkspacePickerModal) Close() {
	m.confirmed = false
	m.cancelled = false
}

func (m *WorkspacePickerModal) IsOpen() bool    { return !m.cancelled }
func (m *WorkspacePickerModal) Confirmed() bool { return m.confirmed }
func (m *WorkspacePickerModal) Cancelled() bool { return m.cancelled }

func (m *WorkspacePickerModal) Init() bubbletea.Cmd { return nil }

func (m *WorkspacePickerModal) Update(msg bubbletea.Msg) (component.Model, bubbletea.Cmd) {
	if keyMsg, ok := msg.(bubbletea.KeyMsg); ok {
		switch keyMsg.String() {
		case "enter":
			m.confirmed = true
			return m, nil
		case "esc":
			m.cancelled = true
			return m, nil
		}
	}

	var cmd bubbletea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *WorkspacePickerModal) View() string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#8E44AD")).
		Padding(1, 2).
		Width(60)

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
