package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"kankuro/internal/config"
)

func (a *App) refreshWorkspaces() {
	a.err = nil
	ws, err := a.repo.ListWorkspaces()
	if err != nil {
		a.err = err
		return
	}
	a.workspaces = ws
	sources, err := a.repo.ListWorkspaceSources()
	if err != nil {
		a.err = err
		return
	}
	a.workspaceForm.Sources = sources
}

func (a *App) filteredWorkspaces() []config.Workspace {
	return filterByQuery(a.workspaces, a.cardForm.WorkspaceSearchQuery, func(w config.Workspace, q string) bool {
		return strings.Contains(strings.ToLower(w.Name), q) ||
			strings.Contains(strings.ToLower(w.Path), q)
	})
}

func (a *App) workspaceSummary() string {
	if a.cardForm.SelectedWorkspaceID == "" {
		return "none selected (press Space to pick)"
	}
	for _, ws := range a.workspaces {
		if ws.ID == a.cardForm.SelectedWorkspaceID {
			path := ws.Path
			if len(path) > pathTruncateLimit {
				path = "..." + path[len(path)-pathTruncateKeep:]
			}
			return fmt.Sprintf("%s (%s)", ws.Name, path)
		}
	}
	return "unknown workspace (press Space to pick)"
}

func (a *App) openWorkspacePicker() {
	a.cardForm.WorkspaceSearchQuery = ""
	a.cardForm.WorkspaceCursor = 0
	a.mode = viewWorkspacePicker
}

func (a *App) updateWorkspacePicker(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return a, nil
	}

	switch keyMsg.String() {
	case keyEsc:
		a.mode = viewCardForm
	case "up":
		if a.cardForm.WorkspaceCursor > 0 {
			a.cardForm.WorkspaceCursor--
		}
	case keyDown:
		a.handleWorkspacePickerDown()
	case keyEnter, " ":
		a.handleWorkspacePickerSelect()
	case keyBackspace:
		a.handleWorkspacePickerBackspace()
	default:
		a.handleWorkspacePickerTyping(keyMsg)
	}
	return a, nil
}

func (a *App) handleWorkspacePickerDown() {
	filtered := a.filteredWorkspaces()
	if a.cardForm.WorkspaceCursor < len(filtered)-1 {
		a.cardForm.WorkspaceCursor++
	}
}

func (a *App) handleWorkspacePickerSelect() {
	filtered := a.filteredWorkspaces()
	if len(filtered) > 0 && a.cardForm.WorkspaceCursor >= 0 && a.cardForm.WorkspaceCursor < len(filtered) {
		a.cardForm.SelectedWorkspaceID = filtered[a.cardForm.WorkspaceCursor].ID
	}
	a.mode = viewCardForm
}

func (a *App) handleWorkspacePickerBackspace() {
	if len(a.cardForm.WorkspaceSearchQuery) > 0 {
		a.cardForm.WorkspaceSearchQuery = a.cardForm.WorkspaceSearchQuery[:len(a.cardForm.WorkspaceSearchQuery)-1]
		a.cardForm.WorkspaceCursor = 0
	}
}

func (a *App) handleWorkspacePickerTyping(keyMsg tea.KeyMsg) {
	if len(keyMsg.Runes) > 0 {
		ch := keyMsg.Runes[0]
		if ch >= printableCharMin && ch <= printableCharMax {
			a.cardForm.WorkspaceSearchQuery += string(ch)
			a.cardForm.WorkspaceCursor = 0
		}
	}
}

func (a *App) renderWorkspacePicker() string {
	return renderRegistryList("Select Workspace", true, a.cardForm.WorkspaceSearchQuery,
		func() []config.Workspace { return a.filteredWorkspaces() },
		func(total int) (int, int) { return a.workspaceWindow(total) },
		a.cardForm.WorkspaceCursor,
		func(id string) bool { return id == a.cardForm.SelectedWorkspaceID },
		func(item config.Workspace) string {
			path := item.Path
			if len(path) > pathTruncateLimit {
				path = "..." + path[len(path)-pathTruncateKeep:]
			}
			return fmt.Sprintf("%s (%s)", item.Name, path)
		},
		func(item config.Workspace) string { return item.ID },
	)
}

func (a *App) workspaceWindow(total int) (int, int) {
	return calculateWindow(a.cardForm.WorkspaceCursor, total, a.skillListHeight())
}

func (a *App) updateWorkspaces(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return a, nil
	}

	switch keyMsg.String() {
	case keyEsc:
		a.mode = viewDashboard
	case "up":
		if a.workspaceForm.Cursor > 0 {
			a.workspaceForm.Cursor--
		}
	case keyDown:
		if a.workspaceForm.Cursor < len(a.workspaceForm.Sources)-1 {
			a.workspaceForm.Cursor++
		}
	case "n":
		a.openNewWorkspaceSourceForm()
	case "e", " ":
		a.openEditWorkspaceSourceForm()
	case "d":
		a.deleteSelectedWorkspaceSource()
	case "r":
		a.refreshWorkspaceRegistry()
	}
	return a, nil
}

func (a *App) refreshWorkspaceRegistry() {
	if err := a.repo.SyncWorkspacesFromSources(); err != nil {
		a.err = err
		return
	}
	a.refreshWorkspaces()
}

func (a *App) openNewWorkspaceSourceForm() {
	a.workspaceForm.EditingID = ""
	a.workspaceForm.Inputs[0].SetValue("")
	a.workspaceForm.Inputs[1].SetValue("n")
	a.workspaceForm.Inputs[0].Focus()
	a.mode = viewWorkspaceSourceForm
}

func (a *App) openEditWorkspaceSourceForm() {
	if len(a.workspaceForm.Sources) == 0 {
		return
	}
	s := a.workspaceForm.Sources[a.workspaceForm.Cursor]
	a.workspaceForm.EditingID = s.ID
	a.workspaceForm.Inputs[0].SetValue(s.Path)
	isRoot := "n"
	if s.IsRoot {
		isRoot = "y"
	}
	a.workspaceForm.Inputs[1].SetValue(isRoot)
	a.workspaceForm.Inputs[0].Focus()
	a.mode = viewWorkspaceSourceForm
}

func (a *App) deleteSelectedWorkspaceSource() {
	if len(a.workspaceForm.Sources) == 0 {
		return
	}
	id := a.workspaceForm.Sources[a.workspaceForm.Cursor].ID
	if err := a.repo.DeleteWorkspaceSource(id); err != nil {
		a.err = err
		return
	}
	a.refreshWorkspaces()
	if a.workspaceForm.Cursor >= len(a.workspaceForm.Sources) {
		a.workspaceForm.Cursor = max(0, len(a.workspaceForm.Sources)-1)
	}
}

func (a *App) renderWorkspacesList() string {
	header := lipgloss.NewStyle().Foreground(gray).Render("Manage Workspace Sources") + "\n\n"
	if len(a.workspaceForm.Sources) == 0 {
		return header + "(no sources defined)"
	}
	var lines []string
	for i, s := range a.workspaceForm.Sources {
		prefix := "  "
		if i == a.workspaceForm.Cursor {
			prefix = "> "
		}
		rootMarker := " "
		if s.IsRoot {
			rootMarker = "R"
		}
		line := fmt.Sprintf("%s[%s] %s", prefix, rootMarker, s.Path)
		if i == a.workspaceForm.Cursor {
			line = lipgloss.NewStyle().Foreground(highlightColor).Render(line)
		}
		lines = append(lines, line)
	}
	return header + strings.Join(lines, "\n") + "\n\n(R=Root folder containing multiple workspaces)"
}

func (a *App) updateWorkspaceSourceForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if handled := a.handleWorkspaceSourceFormKey(keyMsg); handled {
			return a, nil
		}
	}

	var cmds []tea.Cmd
	for i := range a.workspaceForm.Inputs {
		var cmd tea.Cmd
		a.workspaceForm.Inputs[i], cmd = a.workspaceForm.Inputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}
	return a, tea.Batch(cmds...)
}

func (a *App) handleWorkspaceSourceFormKey(keyMsg tea.KeyMsg) bool {
	switch keyMsg.String() {
	case keyEsc:
		a.mode = viewWorkspaces
		return true
	case keyTab:
		a.toggleWorkspaceSourceInputFocus()
		return true
	case keyEnter:
		a.submitWorkspaceSourceForm()
		return true
	default:
		return false
	}
}

func (a *App) toggleWorkspaceSourceInputFocus() {
	if a.workspaceForm.Inputs[0].Focused() {
		a.workspaceForm.Inputs[0].Blur()
		a.workspaceForm.Inputs[1].Focus()
		return
	}
	a.workspaceForm.Inputs[1].Blur()
	a.workspaceForm.Inputs[0].Focus()
}

func (a *App) submitWorkspaceSourceForm() {
	path := strings.TrimSpace(a.workspaceForm.Inputs[0].Value())
	if path == "" {
		a.mode = viewWorkspaces
		return
	}
	if err := a.saveWorkspaceSource(path, parseWorkspaceRootInput(a.workspaceForm.Inputs[1].Value())); err != nil {
		a.err = err
		return
	}
	a.mode = viewWorkspaces
}

func parseWorkspaceRootInput(value string) bool {
	return strings.ToLower(strings.TrimSpace(value)) == "y"
}

func (a *App) saveWorkspaceSource(path string, isRoot bool) error {
	sources, err := a.repo.ListWorkspaceSources()
	if err != nil {
		return err
	}

	source := config.WorkspaceSource{
		ID:     a.workspaceForm.EditingID,
		Path:   path,
		IsRoot: isRoot,
	}
	sources = upsertWorkspaceSource(sources, source)
	if saveErr := a.repo.SaveWorkspaceSources(sources); saveErr != nil {
		return saveErr
	}

	a.refreshWorkspaces()
	return nil
}

func upsertWorkspaceSource(sources []config.WorkspaceSource, source config.WorkspaceSource) []config.WorkspaceSource {
	if source.ID == "" {
		return append(sources, source)
	}
	for i, existing := range sources {
		if existing.ID == source.ID {
			sources[i] = source
			return sources
		}
	}
	return append(sources, source)
}
