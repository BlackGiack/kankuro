package views

import (
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"kankuro/internal/config"
	"kankuro/internal/ui/commands"
	"kankuro/internal/ui/component"
	"kankuro/internal/ui/messages"
)

// Focus indices for the card form.
const (
	FocusName = iota
	FocusDriver
	FocusModel
	FocusWorkspace
	FocusFaceStyle
	FocusPuppet
	FocusUseTmp
	FocusAgentsMD
	FocusSkills
	FocusPersonas
	FocusMCPs
)

// FocusCount is the total number of potential focusable elements.
const FocusCount = 11

// Focus field constants for card editor layout.
const (
	faceStyleIndex = 3
	useTmpIndex    = 4

	// UI constants.
	hPadding = 2
	vPadding = 1
)

// CardEditorCallbacks defines callbacks for card editor actions.
type CardEditorCallbacks struct {
	OnSave              func(card config.IDCard) (config.IDCard, tea.Cmd, error)
	OnCancel            func() tea.Cmd
	OnOpenSources       func(sourceType string) tea.Cmd
	OnOpenMCPForm       func() tea.Cmd
	OnOpenWorkspaces    func() tea.Cmd
	OnOpenPuppetSources func() tea.Cmd
}

// CardEditor is the form for creating/editing ID cards.
type CardEditor struct {
	component.Base

	// Text inputs
	inputs []textinput.Model

	// Capability lists
	skillList   *CapabilityList
	personaList *CapabilityList
	mcpList     *CapabilityList

	// Modals
	workspacePicker *WorkspacePickerModal
	puppetSelector  *PuppetSelectorModal
	agentsMDEditor  *AgentsMDEditorModal

	// State
	editingID   string
	focusIndex  int
	puppet      string
	workspaceID string
	agentsMD    AgentsMDState
	personaIDs  []string
	mcpIDs      []string

	// Data
	workspaces []config.Workspace
	skills     []config.Skill
	personas   []config.Persona
	mcps       []config.MCPServer
	puppets    []config.Puppet
	settings   config.AppSettings

	callbacks CardEditorCallbacks
	cmds      *commands.Commands
}

// AgentsMDState holds the agents.md configuration.
type AgentsMDState struct {
	Content    string
	Path       string
	PreferPath bool
}

const (
	inputCount     = 5
	maxVisibleCaps = 5
	summaryLimit   = 27
)

// NewCardEditor creates a new card editor.
func NewCardEditor(callbacks CardEditorCallbacks, cmds *commands.Commands) *CardEditor {
	// Initialize text inputs
	inputs := make([]textinput.Model, inputCount)
	placeholders := []string{
		"Card name...",
		"Driver (e.g., claude, gemini)...",
		"Model (e.g., claude-3-opus)...",
		"Face style...",
		"Use tmp workspace (y/n)...",
	}
	for i := range inputs {
		inputs[i] = textinput.New()
		inputs[i].Placeholder = placeholders[i]
		inputs[i].Width = 40 // Default width, will be updated in SetSize
	}

	// Initialize capability lists
	skillList := NewCapabilityList(CapabilityListConfig{
		Title:      "Skills",
		SourceHint: "ctrl+e=sources",
		MaxVisible: maxVisibleCaps,
	})

	personaList := NewCapabilityList(CapabilityListConfig{
		Title:      "Personas",
		SourceHint: "ctrl+e=sources",
		MaxVisible: maxVisibleCaps,
	})

	mcpList := NewCapabilityList(CapabilityListConfig{
		Title:      "MCPs",
		SourceHint: "ctrl+e=add",
		MaxVisible: maxVisibleCaps,
	})

	return &CardEditor{
		inputs:          inputs,
		skillList:       skillList,
		personaList:     personaList,
		mcpList:         mcpList,
		workspacePicker: NewWorkspacePickerModal(),
		puppetSelector:  NewPuppetSelectorModal(),
		agentsMDEditor:  NewAgentsMDEditorModal(),
		puppet:          "kankuro",
		settings:        config.DefaultAppSettings(),
		callbacks:       callbacks,
		cmds:            cmds,
	}
}

func (e *CardEditor) SkillList() *CapabilityList { return e.skillList }

func (e *CardEditor) PersonaList() *CapabilityList { return e.personaList }

func (e *CardEditor) MCPList() *CapabilityList { return e.mcpList }

func (e *CardEditor) Settings() config.AppSettings { return e.settings }

func (e *CardEditor) Workspaces() []config.Workspace { return e.workspaces }

func (e *CardEditor) WorkspaceID() string               { return e.workspaceID }
func (e *CardEditor) SetWorkspaceID(workspaceID string) { e.workspaceID = workspaceID }

func (e *CardEditor) Puppet() string          { return e.puppet }
func (e *CardEditor) SetPuppet(puppet string) { e.puppet = puppet }

func (e *CardEditor) EditingID() string { return e.editingID }

func (e *CardEditor) FocusIndex() int { return e.focusIndex }

func (e *CardEditor) Inputs() []textinput.Model { return e.inputs }

func (e *CardEditor) SetSettings(settings config.AppSettings) {
	if e.settings.ShowPersonasInCardEditor && !settings.ShowPersonasInCardEditor {
		if selected := e.personaList.SelectedIDs(); len(selected) > 0 {
			e.personaIDs = selected
		}
	}
	if e.settings.ShowMCPsInCardEditor && !settings.ShowMCPsInCardEditor {
		if selected := e.mcpList.SelectedIDs(); len(selected) > 0 {
			e.mcpIDs = selected
		}
	}
	e.settings = settings
	if !e.isFocusVisible(e.focusIndex) {
		e.setFocus(e.visibleFocusOrder()[0])
	}
	if e.Width > 0 && e.Height > 0 {
		e.SetSize(e.Width, e.Height)
	}
}

// SetCard populates the form for editing an existing card.
func (e *CardEditor) SetCard(card config.IDCard) {
	e.editingID = card.ID
	e.inputs[0].SetValue(card.Name)
	e.inputs[1].SetValue(card.Driver)
	e.inputs[2].SetValue(card.Model)
	e.workspaceID = card.WorkspaceID
	e.inputs[3].SetValue(card.FaceStyle)
	e.puppet = card.Puppet
	if e.puppet == "" {
		e.puppet = defaultPuppetStyle
	}
	tmpVal := "n"
	if card.UseTmpWorkspace {
		tmpVal = "y"
	}
	e.inputs[4].SetValue(tmpVal)
	e.agentsMD.Content = card.AgentsMD
	e.agentsMD.Path = card.AgentsMDPath
	e.agentsMD.PreferPath = card.AgentsMDPreferPath

	e.skillList.SetSelected(card.SkillIDs)
	e.personaList.SetSelected(card.PersonaIDs)
	e.mcpList.SetSelected(card.MCPIDs)
	e.personaIDs = append([]string(nil), card.PersonaIDs...)
	e.mcpIDs = append([]string(nil), card.MCPIDs...)

	e.focusIndex = 0
	e.setFocus(0)
}

// Reset clears the form for creating a new card.
func (e *CardEditor) Reset() {
	for i := range e.inputs {
		e.inputs[i].SetValue("")
		e.inputs[i].Blur()
	}
	e.editingID = ""
	e.workspaceID = ""
	e.puppet = "kankuro"
	e.agentsMD = AgentsMDState{}
	e.personaIDs = nil
	e.mcpIDs = nil
	e.skillList.SetSelected(nil)
	e.personaList.SetSelected(nil)
	e.mcpList.SetSelected(nil)
	e.focusIndex = 0
	e.setFocus(0)
}

// SetData sets the available options.
func (e *CardEditor) SetData(msg messages.DataLoadedMsg) {
	e.workspaces = msg.Workspaces
	e.skills = msg.Skills
	e.personas = msg.Personas
	e.mcps = msg.MCPs

	// Convert to capability items
	skillItems := make([]CapabilityItem, len(e.skills))
	for i, s := range e.skills {
		skillItems[i] = SkillItem{Skill: s}
	}
	e.skillList.SetItems(skillItems)

	personaItems := make([]CapabilityItem, len(e.personas))
	for i, p := range e.personas {
		personaItems[i] = personaItem{persona: p}
	}
	e.personaList.SetItems(personaItems)

	mcpItems := make([]CapabilityItem, len(e.mcps))
	for i, m := range e.mcps {
		mcpItems[i] = mcpItem{mcp: m}
	}
	e.mcpList.SetItems(mcpItems)

	e.puppets = msg.Puppets
	e.puppetSelector.SetPuppets(e.puppets)

	e.workspacePicker.SetWorkspaces(e.workspaces)
}

// Card returns the current card data.
func (e *CardEditor) Card() config.IDCard {
	return config.IDCard{
		ID:                 e.editingID,
		Name:               strings.TrimSpace(e.inputs[0].Value()),
		Driver:             strings.TrimSpace(e.inputs[1].Value()),
		Model:              strings.TrimSpace(e.inputs[2].Value()),
		WorkspaceID:        e.workspaceID,
		FaceStyle:          strings.TrimSpace(e.inputs[3].Value()),
		Puppet:             e.puppet,
		UseTmpWorkspace:    strings.ToLower(strings.TrimSpace(e.inputs[4].Value())) == "y",
		AgentsMD:           e.agentsMD.Content,
		AgentsMDPath:       e.agentsMD.Path,
		AgentsMDPreferPath: e.agentsMD.PreferPath,
		SkillIDs:           e.skillList.SelectedIDs(),
		PersonaIDs:         e.selectedPersonaIDs(),
		MCPIDs:             e.selectedMCPIDs(),
	}
}

// Init initializes the editor.
func (e *CardEditor) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (e *CardEditor) Update(msg tea.Msg) (component.Model, tea.Cmd) {
	// Handle data refresh messages
	switch msg := msg.(type) {
	case messages.MCPDeletedMsg:
		// Remove the deleted MCP from selection
		delete(e.mcpList.selected, msg.MCPID)
		// Adjust cursor if needed
		if e.mcpList.Cursor() >= len(e.mcpList.items) && e.mcpList.Cursor() > 0 {
			e.mcpList.SetCursor(e.mcpList.Cursor() - 1)
		}
		return e, nil

	case messages.MCPUpdatedMsg:
		// MCP list will be refreshed via DataLoadedMsg
		return e, nil
	}

	// Handle modals first
	if e.workspacePicker.IsOpen() {
		return e.updateWorkspacePicker(msg)
	}
	if e.puppetSelector.IsOpen() {
		return e.updatePuppetSelector(msg)
	}
	if e.agentsMDEditor.IsOpen() {
		return e.updateAgentsMDEditor(msg)
	}

	// Handle key events
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if cmd, handled := e.handleKey(keyMsg); handled {
			return e, cmd
		}
	}

	// Route to focused input
	return e.updateFocusedInput(msg)
}

// handleKey processes global key events.
func (e *CardEditor) handleKey(keyMsg tea.KeyMsg) (tea.Cmd, bool) {
	switch keyMsg.Type { //nolint:exhaustive // the form handles only its explicit shortcuts and navigation keys
	case tea.KeyCtrlC:
		return tea.Quit, true

	case tea.KeyEsc:
		e.handleEscape()
		if e.callbacks.OnCancel != nil {
			return e.callbacks.OnCancel(), true
		}
		return nil, true

	case tea.KeyTab:
		e.clearSearches()
		e.CycleFocus(1)
		return nil, true

	case tea.KeyShiftTab:
		e.clearSearches()
		e.CycleFocus(-1)
		return nil, true

	case tea.KeyUp:
		e.handleNav(false)
		return nil, true

	case tea.KeyDown:
		e.handleNav(true)
		return nil, true

	case tea.KeySpace:
		return nil, e.handleSpace()

	case tea.KeyEnter:
		return e.handleSubmit(), true

	case tea.KeyDelete:
		if e.focusIndex == FocusMCPs {
			return e.deleteSelectedMCP(), true
		}

	case tea.KeyRunes:
		if keyMsg.String() == "del" {
			if e.focusIndex == FocusMCPs {
				return e.deleteSelectedMCP(), true
			}
		}

	case tea.KeyCtrlE:
		return e.handleCtrlE()
	}

	return nil, false
}

// handleEscape handles escape key based on current state.
func (e *CardEditor) handleEscape() {
	// Clear search if active
	switch e.focusIndex {
	case FocusSkills:
		if e.skillList.searchQuery != "" {
			e.skillList.SetSearchQuery("")
			return
		}
	case FocusPersonas:
		if e.personaList.searchQuery != "" {
			e.personaList.SetSearchQuery("")
			return
		}
	case FocusMCPs:
		if e.mcpList.searchQuery != "" {
			e.mcpList.SetSearchQuery("")
			return
		}
	}
}

// handleSpace handles space key based on focus.
func (e *CardEditor) handleSpace() bool {
	switch e.focusIndex {
	case FocusWorkspace:
		e.workspacePicker.SetSelectedID(e.workspaceID)
		e.workspacePicker.Open()
		return true

	case FocusPuppet:
		e.puppetSelector.SetSelected(e.puppet)
		e.puppetSelector.Open()
		return true

	case FocusAgentsMD:
		e.agentsMDEditor.SetState(e.agentsMD)
		e.agentsMDEditor.Open()
		return true

	case FocusSkills:
		e.skillList.ToggleSelected()
		return true

	case FocusPersonas:
		if !e.settings.ShowPersonasInCardEditor {
			return false
		}
		e.personaList.ToggleSelected()
		return true

	case FocusMCPs:
		if !e.settings.ShowMCPsInCardEditor {
			return false
		}
		e.mcpList.ToggleSelected()
		// Auto-select mcp-cli skill
		e.autoSelectMCPCLISkill()
		return true
	}

	return false
}

// handleCtrlE handles ctrl+e for opening source editors.
func (e *CardEditor) handleCtrlE() (tea.Cmd, bool) {
	switch e.focusIndex {
	case FocusSkills:
		if e.callbacks.OnOpenSources != nil {
			return e.callbacks.OnOpenSources("skills"), true
		}
		return nil, true

	case FocusPersonas:
		if !e.settings.ShowPersonasInCardEditor {
			return nil, false
		}
		if e.callbacks.OnOpenSources != nil {
			return e.callbacks.OnOpenSources("personas"), true
		}
		return nil, true

	case FocusPuppet:
		if e.callbacks.OnOpenPuppetSources != nil {
			return e.callbacks.OnOpenPuppetSources(), true
		}
		return nil, true

	case FocusMCPs:
		if !e.settings.ShowMCPsInCardEditor {
			return nil, false
		}
		if e.callbacks.OnOpenMCPForm != nil {
			return e.callbacks.OnOpenMCPForm(), true
		}
		return nil, true
	}

	return nil, false
}

// handleNav handles up/down navigation.
func (e *CardEditor) handleNav(down bool) {
	switch e.focusIndex {
	case FocusSkills:
		if down {
			e.skillList.SetCursor(e.skillList.Cursor() + 1)
		} else {
			e.skillList.SetCursor(e.skillList.Cursor() - 1)
		}

	case FocusPersonas:
		if !e.settings.ShowPersonasInCardEditor {
			return
		}
		if down {
			e.personaList.SetCursor(e.personaList.Cursor() + 1)
		} else {
			e.personaList.SetCursor(e.personaList.Cursor() - 1)
		}

	case FocusMCPs:
		if !e.settings.ShowMCPsInCardEditor {
			return
		}
		if down {
			e.mcpList.SetCursor(e.mcpList.Cursor() + 1)
		} else {
			e.mcpList.SetCursor(e.mcpList.Cursor() - 1)
		}

	case FocusAgentsMD:
		e.agentsMD.PreferPath = down
	}
}

// handleSubmit saves the card.
func (e *CardEditor) handleSubmit() tea.Cmd {
	card := e.Card()

	if e.callbacks.OnSave != nil {
		_, cmd, err := e.callbacks.OnSave(card)
		if err != nil {
			// Emit error message
			return func() tea.Msg {
				return messages.ErrorMsg{Err: err}
			}
		}
		if cmd != nil {
			return cmd
		}
	}

	e.editingID = ""
	if e.callbacks.OnCancel != nil {
		return e.callbacks.OnCancel()
	}
	return nil
}

// clearSearches clears all search queries.
func (e *CardEditor) clearSearches() {
	e.skillList.SetSearchQuery("")
	e.personaList.SetSearchQuery("")
	e.mcpList.SetSearchQuery("")
}

// CycleFocus moves focus forward or backward.
func (e *CardEditor) CycleFocus(dir int) {
	order := e.visibleFocusOrder()
	if len(order) == 0 {
		return
	}
	current := 0
	for i, focus := range order {
		if focus == e.focusIndex {
			current = i
			break
		}
	}
	next := (current + dir + len(order)) % len(order)
	e.setFocus(order[next])
}

// setFocus updates focus state for inputs.
func (e *CardEditor) setFocus(index int) {
	e.focusIndex = index
	e.skillList.SetFocused(index == FocusSkills)
	e.personaList.SetFocused(index == FocusPersonas && e.settings.ShowPersonasInCardEditor)
	e.mcpList.SetFocused(index == FocusMCPs && e.settings.ShowMCPsInCardEditor)

	// Map focus index to input index
	inputIdx := e.mapFocusToInput(index)
	for i := range e.inputs {
		if i == inputIdx {
			e.inputs[i].Focus()
		} else {
			e.inputs[i].Blur()
		}
	}
}

// mapFocusToInput maps logical focus index to input array index.
func (e *CardEditor) mapFocusToInput(index int) int {
	switch index {
	case FocusName, FocusDriver, FocusModel:
		return index
	case FocusFaceStyle:
		return faceStyleIndex
	case FocusUseTmp:
		return useTmpIndex
	default:
		return -1
	}
}

// updateFocusedInput routes update to the focused element.
func (e *CardEditor) updateFocusedInput(msg tea.Msg) (component.Model, tea.Cmd) {
	switch e.focusIndex {
	case FocusSkills:
		return e, e.updateSkillList(msg)
	case FocusPersonas:
		if !e.settings.ShowPersonasInCardEditor {
			return e, nil
		}
		return e, e.updatePersonaList(msg)
	case FocusMCPs:
		if !e.settings.ShowMCPsInCardEditor {
			return e, nil
		}
		return e, e.updateMCPList(msg)
	default:
		return e, e.updateInputs(msg)
	}
}

func (e *CardEditor) updateSkillList(msg tea.Msg) tea.Cmd {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if e.handleListSearch(e.skillList, keyMsg) {
			return nil
		}
	}
	_, cmd := e.skillList.Update(msg)
	return cmd
}

func (e *CardEditor) updatePersonaList(msg tea.Msg) tea.Cmd {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if e.handleListSearch(e.personaList, keyMsg) {
			return nil
		}
	}
	_, cmd := e.personaList.Update(msg)
	return cmd
}

func (e *CardEditor) updateMCPList(msg tea.Msg) tea.Cmd {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if e.handleListSearch(e.mcpList, keyMsg) {
			return nil
		}
	}
	_, cmd := e.mcpList.Update(msg)
	return cmd
}

func (e *CardEditor) handleListSearch(list *CapabilityList, keyMsg tea.KeyMsg) bool {
	if keyMsg.Type == tea.KeyRunes {
		for _, ch := range keyMsg.Runes {
			if ch >= 33 && ch <= 126 {
				list.SetSearchQuery(list.searchQuery + string(ch))
				return true
			}
		}
	}
	if keyMsg.Type == tea.KeyBackspace {
		if len(list.searchQuery) > 0 {
			list.SetSearchQuery(list.searchQuery[:len(list.searchQuery)-1])
		}
		return true
	}
	return false
}

func (e *CardEditor) updateInputs(msg tea.Msg) tea.Cmd {
	// Update text inputs
	cmds := make([]tea.Cmd, len(e.inputs))
	for i := range e.inputs {
		e.inputs[i], cmds[i] = e.inputs[i].Update(msg)
	}
	return tea.Batch(cmds...)
}

// updateWorkspacePicker handles workspace picker modal.
func (e *CardEditor) updateWorkspacePicker(msg tea.Msg) (component.Model, tea.Cmd) {
	model, cmd := e.workspacePicker.Update(msg)
	e.workspacePicker, _ = model.(*WorkspacePickerModal)

	if e.workspacePicker.Confirmed() {
		e.workspaceID = e.workspacePicker.SelectedID()
		e.workspacePicker.Close()
	}

	if e.workspacePicker.Cancelled() {
		e.workspacePicker.Close()
	}

	return e, cmd
}

// updatePuppetSelector handles puppet selector modal.
func (e *CardEditor) updatePuppetSelector(msg tea.Msg) (component.Model, tea.Cmd) {
	model, cmd := e.puppetSelector.Update(msg)
	e.puppetSelector, _ = model.(*PuppetSelectorModal)

	if e.puppetSelector.Confirmed() {
		e.puppet = e.puppetSelector.Selected()
		e.puppetSelector.Close()
	}

	if e.puppetSelector.Cancelled() {
		e.puppetSelector.Close()
	}

	return e, cmd
}

// updateAgentsMDEditor handles agents.md editor modal.
func (e *CardEditor) updateAgentsMDEditor(msg tea.Msg) (component.Model, tea.Cmd) {
	model, cmd := e.agentsMDEditor.Update(msg)
	e.agentsMDEditor, _ = model.(*AgentsMDEditorModal)

	if e.agentsMDEditor.Confirmed() {
		e.agentsMD = e.agentsMDEditor.State()
		e.agentsMDEditor.Close()
	}

	if e.agentsMDEditor.Cancelled() {
		e.agentsMDEditor.Close()
	}

	return e, cmd
}

// autoSelectMCPCLISkill selects the mcp-cli skill when an MCP is added.
func (e *CardEditor) autoSelectMCPCLISkill() {
	for _, s := range e.skills {
		if s.Name == "mcp-cli" {
			e.skillList.selected[s.ID] = true
			return
		}
	}
}

// deleteSelectedMCP deletes the selected MCP.
func (e *CardEditor) deleteSelectedMCP() tea.Cmd {
	if item, ok := e.mcpList.SelectedItem(); ok {
		id := item.GetID()
		// Emit message for coordinator to handle
		return func() tea.Msg {
			return messages.MCPDeleteRequestedMsg{MCPID: id}
		}
	}
	return nil
}

// View renders the card editor.
func (e *CardEditor) View() string {
	// Handle modal overlays
	if e.workspacePicker.IsOpen() {
		return e.workspacePicker.View()
	}
	if e.puppetSelector.IsOpen() {
		return e.renderWithOverlay(e.puppetSelector.View())
	}
	if e.agentsMDEditor.IsOpen() {
		return e.renderWithOverlay(e.agentsMDEditor.View())
	}

	return e.renderForm()
}

// renderFormAboveLists renders the form parts that appear before the capability lists.
func (e *CardEditor) renderFormAboveLists() string {
	title := "New ID Card"
	if e.editingID != "" {
		title = "Edit ID Card"
	}

	var sb strings.Builder
	sb.WriteString(titleStyle.Render(title) + "\n\n")

	// Text inputs
	sb.WriteString("Name: " + e.inputs[0].View() + "\n")
	sb.WriteString("Driver: " + e.inputs[1].View() + "\n")
	sb.WriteString("Model: " + e.inputs[2].View() + "\n")
	sb.WriteString(e.renderWorkspaceLine() + "\n")
	sb.WriteString("Face Style: " + e.inputs[3].View() + "\n")
	sb.WriteString(e.renderPuppetLine() + "\n")
	sb.WriteString("Use Tmp Workspace: " + e.inputs[4].View() + "\n\n")

	// Agents.md
	sb.WriteString("AGENTS.md:\n" + e.renderAgentsMDRow() + "\n\n")

	return sb.String()
}

// renderForm renders the main form.
func (e *CardEditor) renderForm() string {
	var sb strings.Builder
	sb.WriteString(e.renderFormAboveLists())

	// Capability lists
	sb.WriteString(e.skillList.View() + "\n\n")
	if e.settings.ShowPersonasInCardEditor {
		sb.WriteString(e.personaList.View() + "\n\n")
	}
	if e.settings.ShowMCPsInCardEditor {
		sb.WriteString(e.mcpList.View() + "\n\n")
	}

	// Help
	sb.WriteString("Enter=save Tab=next Esc=cancel")

	return sb.String()
}

// renderWorkspaceLine renders the workspace picker line.
func (e *CardEditor) renderWorkspaceLine() string {
	summary := e.WorkspaceSummary()
	line := "Workspace: " + summary

	if e.focusIndex == FocusWorkspace {
		line = "> " + line
		return lipgloss.NewStyle().Foreground(highlightColor).Bold(true).Render(line)
	}
	return line
}

// WorkspaceSummary returns a summary of the selected workspace.
func (e *CardEditor) WorkspaceSummary() string {
	if e.workspaceID == "" {
		return "(none selected)"
	}
	for _, ws := range e.workspaces {
		if ws.ID == e.workspaceID {
			return ws.Name
		}
	}
	return e.workspaceID
}

// renderPuppetLine renders the puppet selector line.
func (e *CardEditor) renderPuppetLine() string {
	line := "Puppet: " + e.puppet

	if e.focusIndex == FocusPuppet {
		line = "> " + line
		return lipgloss.NewStyle().Foreground(highlightColor).Bold(true).Render(line)
	}
	return line
}

// renderAgentsMDRow renders the agents.md options.
func (e *CardEditor) renderAgentsMDRow() string {
	const ellipsisLen = 3
	contentSummary := "empty"
	if e.agentsMD.Content != "" {
		lines := strings.Split(e.agentsMD.Content, `\n`)
		contentSummary = lines[0]
		if len(lines) > 1 || len(contentSummary) > summaryLimit {
			if len(contentSummary) > summaryLimit {
				// Trim content and add ellipsis (3 chars)
				contentSummary = contentSummary[:summaryLimit-ellipsisLen] + "..."
			} else {
				contentSummary += "..."
			}
		}
	}

	pathSummary := "empty"
	if e.agentsMD.Path != "" {
		pathSummary = e.agentsMD.Path
		if len(pathSummary) > summaryLimit {
			// Trim path and add ellipsis (3 chars)
			pathSummary = "..." + pathSummary[len(pathSummary)-(summaryLimit-ellipsisLen):]
		}
	}

	contentOption := "[ Content: " + contentSummary + " ]"
	pathOption := "[ Path: " + pathSummary + " ]"

	if e.focusIndex == FocusAgentsMD {
		prefix := "> "
		if e.agentsMD.PreferPath {
			return lipgloss.NewStyle().Foreground(highlightColor).Bold(true).Render(
				prefix + "  " + contentOption + "\n" + prefix + "x " + pathOption,
			)
		}
		return lipgloss.NewStyle().Foreground(highlightColor).Bold(true).Render(
			prefix + "x " + contentOption + "\n" + prefix + "  " + pathOption,
		)
	}

	if e.agentsMD.PreferPath {
		return lipgloss.NewStyle().Foreground(gray).Render("    " + contentOption + "\n  x " + pathOption)
	}
	return lipgloss.NewStyle().Foreground(gray).Render("  x " + contentOption + "\n    " + pathOption)
}

// renderWithOverlay renders form with modal overlay.
func (e *CardEditor) renderWithOverlay(modalContent string) string {
	form := e.renderForm()

	overlayStyle := lipgloss.NewStyle().
		Padding(vPadding, hPadding)

	return form + "\n" + overlayStyle.Render(modalContent)
}

const (
	helpLineHeight    = 1
	listSeparatorRows = 2
	minCapListHeight  = 2
)

// SetSize implements component.Model.
func (e *CardEditor) SetSize(width, height int) {
	e.Base.SetSize(width, height)

	headerHeight := lipgloss.Height(e.renderFormAboveLists())

	part := 1
	if e.settings.ShowPersonasInCardEditor {
		part++
	}
	if e.settings.ShowMCPsInCardEditor {
		part++
	}

	separatorsAndHelp := (part * listSeparatorRows) + helpLineHeight
	available := height - headerHeight - separatorsAndHelp

	capListHeight := max(minCapListHeight, available/part)

	e.skillList.SetSize(width, capListHeight)
	e.personaList.SetSize(width, capListHeight)
	e.mcpList.SetSize(width, capListHeight)

	// Give modals full dimensions
	e.workspacePicker.SetSize(width, height)
	e.agentsMDEditor.SetSize(width, height)
	e.puppetSelector.SetSize(width, height)

	// Update input widths
	calculatedWidth := max(40, width*3/5) //nolint:mnd // responsive sizing
	for i := range e.inputs {
		e.inputs[i].Width = calculatedWidth
	}
}

func (e *CardEditor) visibleFocusOrder() []int {
	order := []int{
		FocusName,
		FocusDriver,
		FocusModel,
		FocusWorkspace,
		FocusFaceStyle,
		FocusPuppet,
		FocusUseTmp,
		FocusAgentsMD,
		FocusSkills,
	}
	if e.settings.ShowPersonasInCardEditor {
		order = append(order, FocusPersonas)
	}
	if e.settings.ShowMCPsInCardEditor {
		order = append(order, FocusMCPs)
	}
	return order
}

func (e *CardEditor) isFocusVisible(index int) bool {
	return slices.Contains(e.visibleFocusOrder(), index)
}

func (e *CardEditor) selectedPersonaIDs() []string {
	if e.settings.ShowPersonasInCardEditor {
		e.personaIDs = e.personaList.SelectedIDs()
		return e.personaIDs
	}
	return append([]string(nil), e.personaIDs...)
}

func (e *CardEditor) selectedMCPIDs() []string {
	if e.settings.ShowMCPsInCardEditor {
		e.mcpIDs = e.mcpList.SelectedIDs()
		return e.mcpIDs
	}
	return append([]string(nil), e.mcpIDs...)
}

// Capability item types

type SkillItem struct {
	Skill config.Skill
}

func (i SkillItem) GetID() string          { return i.Skill.ID }
func (i SkillItem) GetLabel() string       { return i.Skill.Name }
func (i SkillItem) GetDescription() string { return i.Skill.Description }

type personaItem struct {
	persona config.Persona
}

func (i personaItem) GetID() string          { return i.persona.ID }
func (i personaItem) GetLabel() string       { return i.persona.Name }
func (i personaItem) GetDescription() string { return i.persona.Source }

type mcpItem struct {
	mcp config.MCPServer
}

func (i mcpItem) GetID() string          { return i.mcp.ID }
func (i mcpItem) GetLabel() string       { return i.mcp.Name }
func (i mcpItem) GetDescription() string { return i.mcp.Description }
