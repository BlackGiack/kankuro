package views

import (
	"slices"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/ui/component"
	"github.com/BlackGiack/kankuro/internal/ui/components"
)

// WorkspacePickerModal allows selecting a workspace.
type WorkspacePickerModal struct {
	component.Base
	components.ModalState

	list                  list.Model
	workspaces            []config.Workspace
	selectedID            string
	searchQuery           string
	modalWidth            int
	modalHeight           int
	listWidth             int
	multiple              bool
	selectedIDs           []string
	checked               map[string]bool
	excludedID            string
	selectedWorkspaceInfo map[string]config.Workspace
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

	workspaces := m.filterWorkspaces()
	items := make([]list.Item, len(workspaces))
	for i, ws := range workspaces {
		items[i] = workspaceItem{workspace: ws, checked: m.multiple && m.checked[ws.ID], multiple: m.multiple}
	}

	l := list.New(items, delegate, 40, 30) //nolint:mnd // responsive sizing
	l.SetShowTitle(true)
	l.Title = m.title()
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false) // Use manual filtering

	m.list = l

	// Restore the current size to preserve terminal dimensions
	if m.modalWidth > 0 && m.modalHeight > 0 {
		m.applyListSize()
	}

	if !m.multiple {
		m.restoreSingleSelection()
	}
}

func (m *WorkspacePickerModal) title() string {
	if m.multiple {
		return "Select Sub Workspaces"
	}
	return "Select Workspace"
}

func (m *WorkspacePickerModal) SetWorkspaces(workspaces []config.Workspace) {
	for _, workspace := range m.workspaces {
		if m.checked[workspace.ID] {
			m.rememberWorkspace(workspace)
		}
	}
	m.workspaces = workspaces
	m.refreshItems()
}

func (m *WorkspacePickerModal) SetExcludedID(id string) { m.excludedID = id; m.refreshItems() }

func (m *WorkspacePickerModal) SetSelectedID(id string) {
	m.multiple = false
	m.selectedID = id
	m.restoreSingleSelection()
}

func (m *WorkspacePickerModal) restoreSingleSelection() {
	id := m.selectedID
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

func (m *WorkspacePickerModal) SetMultiple(enabled bool) { m.multiple = enabled; m.rebuildList() }
func (m *WorkspacePickerModal) SetSelectedIDs(ids []string) {
	m.multiple = true
	m.selectedIDs = append([]string(nil), ids...)
	m.checked = make(map[string]bool)
	m.selectedWorkspaceInfo = make(map[string]config.Workspace)
	for _, id := range ids {
		m.checked[id] = true
		for _, workspace := range m.workspaces {
			if workspace.ID == id {
				m.rememberWorkspace(workspace)
				break
			}
		}
	}
	m.refreshItems()
}
func (m *WorkspacePickerModal) SelectedIDs() []string {
	out := make([]string, 0, len(m.checked))
	for _, ws := range m.workspaces {
		if m.checked[ws.ID] {
			out = append(out, ws.ID)
		}
	}
	for _, id := range m.selectedIDs {
		if m.checked[id] {
			found := false
			for _, ws := range m.workspaces {
				if ws.ID == id {
					found = true
					break
				}
			}
			if !found {
				out = append(out, id)
			}
		}
	}
	return out
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

//nolint:gocognit,nestif // modal key handling preserves distinct confirm/cancel/search behavior.
func (m *WorkspacePickerModal) Update(msg tea.Msg) (component.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch keyMsg.Code {
		case tea.KeyEnter:
			if m.multiple {
				m.Confirm()
				m.selectedIDs = m.SelectedIDs()
				m.resetTransientState()
				return m, nil
			}
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
		case tea.KeySpace:
			if m.multiple {
				if item := m.list.SelectedItem(); item != nil {
					wi, itemOK := item.(workspaceItem)
					if !itemOK {
						return m, nil
					}
					m.toggleChecked(wi.workspace.ID)
					m.refreshItems()
				}
				return m, nil
			}
			if keyMsg.Text != "" {
				m.searchQuery += keyMsg.Text
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

func (m *WorkspacePickerModal) toggleChecked(id string) {
	if m.checked[id] {
		delete(m.checked, id)
		m.selectedIDs = removeID(m.selectedIDs, id)
		return
	}
	m.checked[id] = true
	if item := m.list.SelectedItem(); item != nil {
		if workspace, ok := item.(workspaceItem); ok {
			m.rememberWorkspace(workspace.workspace)
		}
	}
	if !slices.Contains(m.selectedIDs, id) {
		m.selectedIDs = append(m.selectedIDs, id)
	}
}

func (m *WorkspacePickerModal) rememberWorkspace(workspace config.Workspace) {
	if m.selectedWorkspaceInfo == nil {
		m.selectedWorkspaceInfo = make(map[string]config.Workspace)
	}
	m.selectedWorkspaceInfo[workspace.ID] = workspace
}

func removeID(ids []string, target string) []string {
	for i, id := range ids {
		if id == target {
			return append(ids[:i], ids[i+1:]...)
		}
	}
	return ids
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
	if m.multiple {
		help = "Space=toggle • Enter=confirm • Esc=back • Type to search"
	}
	if m.searchQuery != "" {
		if m.multiple {
			help = "Space=toggle • Enter=confirm • Esc=clear search • Type to search"
		} else {
			help = "Enter=select • Esc=clear search • Type to search"
		}
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
		items[i] = workspaceItem{workspace: ws, checked: m.multiple && m.checked[ws.ID], multiple: m.multiple}
	}
	m.list.SetItems(items)
	// Normalize selection after refreshing items
	m.normalizeSelection()
}

func (m *WorkspacePickerModal) filterWorkspaces() []config.Workspace {
	workspaces := make([]config.Workspace, 0, len(m.workspaces)+len(m.checked))
	registered := make(map[string]bool, len(m.workspaces))
	for _, ws := range m.workspaces {
		registered[ws.ID] = true
		if m.multiple && ws.ID == m.excludedID && !m.checked[ws.ID] {
			continue
		}
		workspaces = append(workspaces, ws)
	}
	if m.multiple {
		for _, id := range m.selectedIDs {
			if registered[id] || !m.checked[id] {
				continue
			}
			workspace := m.selectedWorkspaceInfo[id]
			name := "[missing] " + id
			path := "Not in workspace registry"
			if workspace.Name != "" {
				name = "[missing] " + workspace.Name + " (" + id + ")"
				path = workspace.Path
			}
			workspaces = append(workspaces, config.Workspace{ID: id, Name: name, Path: path})
		}
	}
	return components.Filter(workspaces, m.searchQuery, func(ws config.Workspace) string {
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
	checked   bool
	multiple  bool
}

func (i workspaceItem) Title() string {
	if !i.multiple {
		return components.DisplayText(i.workspace.Name)
	}
	mark := ""
	if i.checked {
		mark = "[x] "
	} else if i.workspace.Name != "" {
		mark = "[ ] "
	}
	return mark + components.DisplayText(i.workspace.Name)
}
func (i workspaceItem) Description() string { return components.DisplayText(i.workspace.Path) }
func (i workspaceItem) FilterValue() string { return i.workspace.Name }
