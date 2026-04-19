package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"kankuro/internal/config"
	"kankuro/internal/ui/component"
)

// SourceListMode determines if showing list or form.
type SourceListMode int

const (
	SourceListModeList SourceListMode = iota
	SourceListModeForm
)

// SourceListCallbacks defines callbacks for source actions.
type SourceListCallbacks struct {
	// OnSave is called when user saves a source.
	// Returns any error.
	OnSave func(item SourceItem, editingIndex int) error

	// OnDelete is called when user deletes a source.
	OnDelete func(index int) error

	// OnRefresh is called after save/delete to reload data.
	OnRefresh func()

	// OnClose is called when user presses Escape.
	OnClose func()
}

// SourceListConfig configures the source list behavior.
type SourceListConfig struct {
	Title       string
	Legend      string // Shown at bottom, e.g., "(R=Root folder...)"
	FormPrompt  string // Label for input, e.g., "Path:"
	FormFields  int    // Number of form fields (1 or 2)
	Field2Label string // Label for second field, e.g., "Is Root (y/n):"
}

// SourceList is a reusable component for managing source lists.
type SourceList struct {
	component.Base

	Mode      SourceListMode
	Config    SourceListConfig
	Callbacks SourceListCallbacks

	Items  []SourceItem
	Cursor int

	Inputs       []textinput.Model
	EditingIndex int // -1 for new, >= 0 for edit

	// Theme
	PrimaryColor   lipgloss.Color
	HighlightColor lipgloss.Color
	Gray           lipgloss.Color
}

// NewSourceList creates a new source list component.
func NewSourceList(config SourceListConfig, callbacks SourceListCallbacks) *SourceList {
	// Create inputs based on config
	numInputs := config.FormFields
	if numInputs < 1 {
		numInputs = 1
	}

	inputs := make([]textinput.Model, numInputs)
	for i := range inputs {
		inputs[i] = textinput.New()
		inputs[i].Width = 60
	}

	if config.FormPrompt != "" {
		inputs[0].Placeholder = config.FormPrompt
	}

	return &SourceList{
		Config:         config,
		Callbacks:      callbacks,
		Inputs:         inputs,
		EditingIndex:   -1,
		PrimaryColor:   lipgloss.Color("#8E44AD"),
		HighlightColor: lipgloss.Color("#16A085"),
		Gray:           lipgloss.Color("240"),
	}
}

// SetItems updates the source list data.
func (s *SourceList) SetItems(items []SourceItem) {
	s.Items = items

	// Adjust cursor if needed
	maxCursor := len(items) // index of "New" option
	if s.Cursor > maxCursor {
		s.Cursor = maxCursor
	}
	if s.Cursor < 0 {
		s.Cursor = 0
	}
}

// Init initializes the component.
func (s *SourceList) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (s *SourceList) Update(msg tea.Msg) (component.Model, tea.Cmd) {
	if s.Mode == SourceListModeForm {
		return s.updateForm(msg)
	}
	return s.updateList(msg)
}

// updateList handles list mode input.
func (s *SourceList) updateList(msg tea.Msg) (component.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return s, nil
	}

	switch keyMsg.String() {
	case "up", "k":
		if s.Cursor > 0 {
			s.Cursor--
		}
	case "down", "j":
		if s.Cursor < len(s.Items) {
			s.Cursor++
		}
	case "n":
		s.OpenForm(-1) // New item
	case " ", "enter":
		s.OpenForm(s.Cursor)
	case "d", "backspace":
		s.HandleDelete()
	case "esc":
		if s.Callbacks.OnClose != nil {
			s.Callbacks.OnClose()
		}
	}

	return s, nil
}

// updateForm handles form mode input.
func (s *SourceList) updateForm(msg tea.Msg) (component.Model, tea.Cmd) {
	// Update all inputs
	var cmd tea.Cmd
	cmds := make([]tea.Cmd, len(s.Inputs))
	for i := range s.Inputs {
		s.Inputs[i], cmds[i] = s.Inputs[i].Update(msg)
	}
	cmd = tea.Batch(cmds...)

	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return s, cmd
	}

	switch keyMsg.String() {
	case "enter":
		s.HandleSave()
		// Mode is usually switched by callback/parent or here if successful
		// For simplicity in generic component, we return to list mode
		s.Mode = SourceListModeList
	case "tab":
		s.cycleInputFocus()
	case "esc":
		s.Mode = SourceListModeList
	}

	return s, cmd
}

// OpenForm opens the form for new or edit.
func (s *SourceList) OpenForm(index int) {
	s.EditingIndex = index

	if index == len(s.Items) || index < 0 {
		// New item
		s.EditingIndex = -1
		for i := range s.Inputs {
			s.Inputs[i].SetValue("")
		}
	} else {
		// Edit existing
		item := s.Items[index]
		if ws, ok := item.(WorkspaceSourceItem); ok {
			s.Inputs[0].SetValue(ws.Source.Path)
			if len(s.Inputs) > 1 {
				val := "n"
				if ws.Source.IsRoot {
					val = "y"
				}
				s.Inputs[1].SetValue(val)
			}
		} else {
			s.Inputs[0].SetValue(item.Display())
		}
	}

	for i := range s.Inputs {
		s.Inputs[i].Blur()
	}
	s.Inputs[0].Focus()
	s.Mode = SourceListModeForm
}

// HandleSave processes form submission.
func (s *SourceList) HandleSave() {
	path := strings.TrimSpace(s.Inputs[0].Value())
	if path == "" {
		return
	}

	var item SourceItem

	if len(s.Inputs) > 1 {
		// Complex item (workspace)
		isRoot := strings.ToLower(strings.TrimSpace(s.Inputs[1].Value())) == "y"

		// Get existing ID if editing
		existingID := ""
		if s.EditingIndex >= 0 && s.EditingIndex < len(s.Items) {
			existingID = s.Items[s.EditingIndex].ID()
		}

		item = WorkspaceSourceItem{
			Source: config.WorkspaceSource{
				ID:     existingID,
				Path:   path,
				IsRoot: isRoot,
			},
		}
	} else {
		// Simple item (skill/persona)
		item = SimpleSourceItem(path)
	}

	if s.Callbacks.OnSave != nil {
		if err := s.Callbacks.OnSave(item, s.EditingIndex); err != nil {
			// Error handling is done by callback (e.g. set error in App)
			return
		}
	}

	if s.Callbacks.OnRefresh != nil {
		s.Callbacks.OnRefresh()
	}
}

// HandleDelete processes item deletion.
func (s *SourceList) HandleDelete() {
	if s.Cursor >= len(s.Items) {
		return // Can't delete "New" option
	}

	if s.Callbacks.OnDelete != nil {
		if err := s.Callbacks.OnDelete(s.Cursor); err != nil {
			return
		}
	}

	// Adjust cursor
	if s.Cursor >= len(s.Items)-1 && s.Cursor > 0 {
		s.Cursor--
	}

	if s.Callbacks.OnRefresh != nil {
		s.Callbacks.OnRefresh()
	}
}

// cycleInputFocus cycles between form inputs.
func (s *SourceList) cycleInputFocus() {
	if len(s.Inputs) <= 1 {
		return
	}

	// Find current focus
	for i := range s.Inputs {
		if s.Inputs[i].Focused() {
			s.Inputs[i].Blur()
			next := (i + 1) % len(s.Inputs)
			s.Inputs[next].Focus()
			return
		}
	}
	s.Inputs[0].Focus()
}

// View renders the component.
func (s *SourceList) View() string {
	if s.Mode == SourceListModeForm {
		return s.renderForm()
	}
	return s.renderList()
}

// renderList renders the source list.
func (s *SourceList) renderList() string {
	var lines []string

	lines = append(lines, lipgloss.NewStyle().
		Bold(true).
		Foreground(s.PrimaryColor).
		Render(s.Config.Title))
	lines = append(lines, "")
	lines = append(lines, "Defined sources:")

	if len(s.Items) == 0 {
		lines = append(lines, lipgloss.NewStyle().
			Foreground(s.Gray).
			Render("  (no sources defined)"))
	} else {
		for i, item := range s.Items {
			prefix := "  "
			if i == s.Cursor {
				prefix = "> "
			}

			line := fmt.Sprintf("%s%s", prefix, item.Display())

			if i == s.Cursor {
				line = lipgloss.NewStyle().
					Foreground(s.HighlightColor).
					Render(line)
			}
			lines = append(lines, line)
		}
	}

	// "New" option
	prefix := "  "
	if s.Cursor == len(s.Items) {
		prefix = "> "
	}
	newLine := fmt.Sprintf("%s[ New Source ]", prefix)
	if s.Cursor == len(s.Items) {
		newLine = lipgloss.NewStyle().
			Foreground(s.HighlightColor).
			Render(newLine)
	}
	lines = append(lines, newLine)

	// Legend
	if s.Config.Legend != "" {
		lines = append(lines, "")
		lines = append(lines, lipgloss.NewStyle().
			Foreground(s.Gray).
			Render(s.Config.Legend))
	}

	// Help text
	lines = append(lines, "")
	lines = append(lines, "Enter/Space=edit n=new d=delete Esc=back")

	return strings.Join(lines, "\n")
}

// renderForm renders the input form.
func (s *SourceList) renderForm() string {
	title := "New Source"
	if s.EditingIndex >= 0 {
		title = "Edit Source"
	}

	var lines []string

	lines = append(lines, lipgloss.NewStyle().
		Bold(true).
		Foreground(s.PrimaryColor).
		Render(title))
	lines = append(lines, "")

	// First input
	prompt := s.Config.FormPrompt
	if prompt == "" {
		prompt = "Path"
	}
	lines = append(lines, fmt.Sprintf("%s: %s", prompt, s.Inputs[0].View()))

	// Second input (if configured)
	if len(s.Inputs) > 1 && s.Config.Field2Label != "" {
		lines = append(lines, fmt.Sprintf("%s: %s", s.Config.Field2Label, s.Inputs[1].View()))
	}

	// Help text
	lines = append(lines, "")
	if len(s.Inputs) > 1 {
		lines = append(lines, "Enter=save Tab=next Esc=cancel")
	} else {
		lines = append(lines, "Enter=save Esc=cancel")
	}

	return strings.Join(lines, "\n")
}

// SetSize implements component.Model.
func (s *SourceList) SetSize(width, height int) {
	s.Base.SetSize(width, height)
	// Adjust input width
	for i := range s.Inputs {
		s.Inputs[i].Width = min(60, width-10)
	}
}
