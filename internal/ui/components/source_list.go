package components

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/ui/component"
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
	// Returns any error and an optional command.
	OnSave func(item SourceItem, editingIndex int) (tea.Cmd, error)

	// OnDelete is called when user deletes a source.
	// Returns any error and an optional command.
	OnDelete func(index int) (tea.Cmd, error)

	// OnRefresh is called after save/delete to reload data.
	OnRefresh func() tea.Cmd

	// OnReload is called to reload items into the list after save/delete.
	OnReload func()

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
	PrimaryColor   color.Color
	HighlightColor color.Color
	Gray           color.Color
}

// NewSourceList creates a new source list component.
func NewSourceList(config SourceListConfig, callbacks SourceListCallbacks) *SourceList {
	// Create inputs based on config
	numInputs := config.FormFields
	numInputs = max(numInputs, 1)

	inputs := make([]textinput.Model, numInputs)
	for i := range inputs {
		inputs[i] = textinput.New()
		inputs[i].SetWidth(defaultInputWidth)
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
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return s, nil
	}

	switch keyMsg.Code {
	case tea.KeyUp:
		s.moveUp()
	case tea.KeyDown:
		s.moveDown()
	case tea.KeyEnter, tea.KeySpace:
		s.OpenForm(s.Cursor)
	case tea.KeyBackspace:
		return s, s.HandleDelete()
	case tea.KeyEsc:
		if s.Callbacks.OnClose != nil {
			s.Callbacks.OnClose()
		}
	default:
		return s, s.handleListRune(keyMsg.Text)
	}

	return s, nil
}

func (s *SourceList) moveUp() {
	if s.Cursor > 0 {
		s.Cursor--
	}
}

func (s *SourceList) moveDown() {
	if s.Cursor < len(s.Items) {
		s.Cursor++
	}
}

func (s *SourceList) handleListRune(r string) tea.Cmd {
	switch r {
	case "k":
		s.moveUp()
	case "j":
		s.moveDown()
	case "n":
		s.OpenForm(-1) // New item
	case "d":
		return s.HandleDelete()
	case "r":
		var refreshCmd tea.Cmd
		if s.Callbacks.OnRefresh != nil {
			refreshCmd = s.Callbacks.OnRefresh()
		}
		if s.Callbacks.OnReload != nil {
			s.Callbacks.OnReload()
		}
		return refreshCmd
	}
	return nil
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

	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return s, cmd
	}

	switch keyMsg.Code {
	case tea.KeyEnter:
		saveCmd := s.HandleSave()
		// Mode is usually switched by callback/parent or here if successful
		// For simplicity in generic component, we return to list mode
		s.Mode = SourceListModeList
		return s, tea.Batch(cmd, saveCmd)
	case tea.KeyTab:
		s.cycleInputFocus()
		return s, cmd
	case tea.KeyEsc:
		s.Mode = SourceListModeList
	}

	return s, cmd
}

// OpenForm opens the form for new or edit.
func (s *SourceList) OpenForm(index int) {
	s.EditingIndex = index

	if index == len(s.Items) || index < 0 {
		s.initNewForm()
	} else {
		s.initEditForm(index)
	}

	for i := range s.Inputs {
		s.Inputs[i].Blur()
	}
	s.Inputs[0].Focus()
	s.Mode = SourceListModeForm
}

func (s *SourceList) initNewForm() {
	// New item
	s.EditingIndex = -1
	for i := range s.Inputs {
		s.Inputs[i].SetValue("")
	}
}

func (s *SourceList) initEditForm(index int) {
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

// HandleSave processes form submission.
func (s *SourceList) HandleSave() tea.Cmd {
	path := strings.TrimSpace(s.Inputs[0].Value())
	if path == "" {
		return nil
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

	var cmd tea.Cmd
	if s.Callbacks.OnSave != nil {
		saveCmd, err := s.Callbacks.OnSave(item, s.EditingIndex)
		if err != nil {
			// Error handling is done by callback (e.g. set error in App)
			return nil
		}
		cmd = saveCmd
	}

	if s.Callbacks.OnRefresh != nil {
		refreshCmd := s.Callbacks.OnRefresh()
		cmd = tea.Batch(cmd, refreshCmd)
	}

	if s.Callbacks.OnReload != nil {
		s.Callbacks.OnReload()
	}

	return cmd
}

// HandleDelete processes item deletion.
func (s *SourceList) HandleDelete() tea.Cmd {
	if s.Cursor >= len(s.Items) {
		return nil // Can't delete "New" option
	}

	var cmd tea.Cmd
	if s.Callbacks.OnDelete != nil {
		delCmd, err := s.Callbacks.OnDelete(s.Cursor)
		if err != nil {
			return nil
		}
		cmd = delCmd
	}

	// Adjust cursor
	if s.Cursor >= len(s.Items)-1 && s.Cursor > 0 {
		s.Cursor--
	}

	if s.Callbacks.OnRefresh != nil {
		refreshCmd := s.Callbacks.OnRefresh()
		cmd = tea.Batch(cmd, refreshCmd)
	}

	if s.Callbacks.OnReload != nil {
		s.Callbacks.OnReload()
	}

	return cmd
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
	lines = append(lines, "Enter/Space=edit n=new d=delete r=refresh Esc=back")

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
	const (
		maxInputWidth = 60
		inputMargin   = 10
	)
	for i := range s.Inputs {
		s.Inputs[i].SetWidth(min(maxInputWidth, width-inputMargin))
	}
}
