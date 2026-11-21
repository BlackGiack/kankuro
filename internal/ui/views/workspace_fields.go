package views

import (
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/ui/components"
	"github.com/BlackGiack/kankuro/internal/ui/models"
	"github.com/BlackGiack/kankuro/internal/ui/styles"
)

// workspaceFields owns the workspace draft and picker shared by editing and launch.
type workspaceFields struct {
	workspacePicker    *WorkspacePickerModal
	workspaceID        string
	subWorkspaceIDs    []string
	subWorkspaceEdited bool
	workspaces         []models.WorkspaceOption
}

func newWorkspaceFields() *workspaceFields {
	return &workspaceFields{workspacePicker: NewWorkspacePickerModal()}
}

func (w *workspaceFields) setChoices(choices contracts.ReadResult[[]contracts.WorkspaceOption]) {
	if !choices.Attempted || choices.Err != nil {
		return
	}
	w.workspaces = make([]models.WorkspaceOption, 0, len(choices.Value))
	for _, v := range choices.Value {
		w.workspaces = append(w.workspaces, models.WorkspaceOption{ID: v.ID, Name: v.Name, Path: v.Path})
	}
	w.workspacePicker.SetWorkspaces(w.workspaces)
}
func (w *workspaceFields) open(multiple bool) {
	if multiple {
		w.workspacePicker.SetMultiple(true)
		w.workspacePicker.SetExcludedID(w.workspaceID)
		w.workspacePicker.SetSelectedIDs(w.subWorkspaceIDs)
	} else {
		w.workspacePicker.SetSelectedID(w.workspaceID)
	}
	w.workspacePicker.Open()
}
func (w *workspaceFields) updatePicker(msg tea.Msg, multiple bool) tea.Cmd {
	model, cmd := w.workspacePicker.Update(msg)
	w.workspacePicker, _ = model.(*WorkspacePickerModal)
	if w.workspacePicker.Confirmed() {
		if multiple {
			w.subWorkspaceIDs = w.workspacePicker.SelectedIDs()
			w.subWorkspaceEdited = true
			w.workspacePicker.SetMultiple(false)
		} else {
			w.workspaceID = w.workspacePicker.SelectedID()
			if slices.Contains(w.subWorkspaceIDs, w.workspaceID) {
				w.subWorkspaceIDs = slices.DeleteFunc(w.subWorkspaceIDs, func(id string) bool { return id == w.workspaceID })
				w.subWorkspaceEdited = true
			}
		}
		w.workspacePicker.Close()
	}
	if w.workspacePicker.Cancelled() {
		w.workspacePicker.Close()
	}
	return cmd
}
func (w *workspaceFields) summary() string {
	if w.workspaceID == "" {
		return "(none selected)"
	}
	for _, ws := range w.workspaces {
		if ws.ID == w.workspaceID {
			return ws.Name
		}
	}
	return w.workspaceID
}
func (w *workspaceFields) renderWorkspaceLine(width int, focused bool) string {
	value := components.DisplayText(w.summary() + " (ctrl+e=sources)")
	if width > 0 {
		value = components.TruncateHead(value, width-len(focusMarker)-len(workspaceLabel))
	}
	line := workspaceLabel + value
	if focused {
		return lipgloss.NewStyle().Foreground(styles.HighlightColor).Bold(true).Render(focusMarker + line)
	}
	return line
}
func (w *workspaceFields) renderSubWorkspacesLine(width int, focused bool) string {
	names := make([]string, 0, len(w.subWorkspaceIDs))
	for _, id := range w.subWorkspaceIDs {
		name := id
		for _, ws := range w.workspaces {
			if ws.ID == id {
				name = ws.Name
				break
			}
		}
		names = append(names, components.DisplayText(name))
	}
	value := "none"
	if len(names) > 0 {
		value = strings.Join(names, ", ")
	}
	label, hint := "Sub Workspaces: ", " (ctrl+e=sources)"
	if focused {
		hint = " (ctrl+e=sources ctrl+x=clear)"
	}
	if width > 0 {
		value = components.TruncateHead(components.DisplayText(value), width-len(focusMarker)-len(label)-lipgloss.Width(hint))
	}
	line := label + value + hint
	if focused {
		return lipgloss.NewStyle().Foreground(styles.HighlightColor).Bold(true).Render(focusMarker + line)
	}
	return line
}
