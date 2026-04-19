package views

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	bubbletea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"kankuro/internal/config"
	"kankuro/internal/ui/commands"
	"kankuro/internal/ui/component"
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

	FocusCount = 11
)

// CardEditorCallbacks defines callbacks for card editor actions.
type CardEditorCallbacks struct {
	OnSave           func(card config.IDCard) (config.IDCard, error)
	OnCancel         func()
	OnOpenSources    func(sourceType string)
	OnOpenMCPForm    func()
	OnOpenWorkspaces func()
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

	// Data
	workspaces []config.Workspace
	skills     []config.Skill
	personas   []config.Persona
	mcps       []config.MCPServer

	callbacks CardEditorCallbacks
	cmds      *commands.Commands
}

// AgentsMDState holds the agents.md configuration.
type AgentsMDState struct {
	Content    string
	Path       string
	PreferPath bool
}

// NewCardEditor creates a new card editor.
func NewCardEditor(callbacks CardEditorCallbacks, cmds *commands.Commands) *CardEditor {
	// Initialize text inputs
	inputs := make([]textinput.Model, 5)
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
		inputs[i].Width = 60
	}

	// Initialize capability lists
	skillList := NewCapabilityList(CapabilityListConfig{
		Title:      "Skills",
		SourceHint: "ctrl+e=sources",
		MaxVisible: 5,
	})

	personaList := NewCapabilityList(CapabilityListConfig{
		Title:      "Personas",
		SourceHint: "ctrl+e=sources",
		MaxVisible: 5,
	})

	mcpList := NewCapabilityList(CapabilityListConfig{
		Title:      "MCPs",
		SourceHint: "ctrl+e=add",
		MaxVisible: 5,
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
		callbacks:       callbacks,
		cmds:            cmds,
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
		e.puppet = "kankuro"
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
	e.skillList.SetSelected(nil)
	e.personaList.SetSelected(nil)
	e.mcpList.SetSelected(nil)
	e.focusIndex = 0
	e.setFocus(0)
}

// SetData sets the available options.
func (e *CardEditor) SetData(
	workspaces []config.Workspace,
	skills []config.Skill,
	personas []config.Persona,
	mcps []config.MCPServer,
) {
	e.workspaces = workspaces
	e.skills = skills
	e.personas = personas
	e.mcps = mcps

	// Convert to capability items
	skillItems := make([]CapabilityItem, len(skills))
	for i, s := range skills {
		skillItems[i] = skillItem{skill: s}
	}
	e.skillList.SetItems(skillItems)

	personaItems := make([]CapabilityItem, len(personas))
	for i, p := range personas {
		personaItems[i] = personaItem{persona: p}
	}
	e.personaList.SetItems(personaItems)

	mcpItems := make([]CapabilityItem, len(mcps))
	for i, m := range mcps {
		mcpItems[i] = mcpItem{mcp: m}
	}
	e.mcpList.SetItems(mcpItems)

	e.workspacePicker.SetWorkspaces(workspaces)
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
		PersonaIDs:         e.personaList.SelectedIDs(),
		MCPIDs:             e.mcpList.SelectedIDs(),
	}
}

// Init initializes the editor.
func (e *CardEditor) Init() bubbletea.Cmd {
	return nil
}

// Update handles messages.
func (e *CardEditor) Update(msg bubbletea.Msg) (component.Model, bubbletea.Cmd) {
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
	if keyMsg, ok := msg.(bubbletea.KeyMsg); ok {
		if cmd, handled := e.handleKey(keyMsg); handled {
			return e, cmd
		}
	}

	// Route to focused input
	return e.updateFocusedInput(msg)
}

// handleKey processes global key events.
func (e *CardEditor) handleKey(keyMsg bubbletea.KeyMsg) (bubbletea.Cmd, bool) {
	switch keyMsg.String() {
	case "ctrl+c":
		return bubbletea.Quit, true

	case "esc":
		e.handleEscape()
		return nil, true

	case "tab":
		e.clearSearches()
		e.cycleFocus(1)
		return nil, true

	case "shift+tab":
		e.clearSearches()
		e.cycleFocus(-1)
		return nil, true

	case "up":
		e.handleNav(false)
		return nil, true

	case "down":
		e.handleNav(true)
		return nil, true

	case " ":
		return nil, e.handleSpace()

	case "enter":
		e.handleSubmit()
		return nil, true

	case "delete", "del":
		if e.focusIndex == FocusMCPs {
			e.deleteSelectedMCP()
			return nil, true
		}

	case "ctrl+e":
		return nil, e.handleCtrlE()
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

	// Otherwise cancel
	if e.callbacks.OnCancel != nil {
		e.callbacks.OnCancel()
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
		e.personaList.ToggleSelected()
		return true

	case FocusMCPs:
		e.mcpList.ToggleSelected()
		// Auto-select mcp-cli skill
		e.autoSelectMCPCLISkill()
		return true
	}

	return false
}

// handleCtrlE handles ctrl+e for opening source editors.
func (e *CardEditor) handleCtrlE() bool {
	switch e.focusIndex {
	case FocusSkills:
		if e.callbacks.OnOpenSources != nil {
			e.callbacks.OnOpenSources("skills")
		}
		return true

	case FocusPersonas:
		if e.callbacks.OnOpenSources != nil {
			e.callbacks.OnOpenSources("personas")
		}
		return true

	case FocusMCPs:
		if e.callbacks.OnOpenMCPForm != nil {
			e.callbacks.OnOpenMCPForm()
		}
		return true
	}

	return false
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
		if down {
			e.personaList.SetCursor(e.personaList.Cursor() + 1)
		} else {
			e.personaList.SetCursor(e.personaList.Cursor() - 1)
		}

	case FocusMCPs:
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
func (e *CardEditor) handleSubmit() {
	card := e.Card()

	if e.callbacks.OnSave != nil {
		_, err := e.callbacks.OnSave(card)
		if err != nil {
			// Emit error message
			return
		}
	}

	e.editingID = ""
	if e.callbacks.OnCancel != nil {
		e.callbacks.OnCancel()
	}
}

// clearSearches clears all search queries.
func (e *CardEditor) clearSearches() {
	e.skillList.SetSearchQuery("")
	e.personaList.SetSearchQuery("")
	e.mcpList.SetSearchQuery("")
}

// cycleFocus moves focus forward or backward.
func (e *CardEditor) cycleFocus(dir int) {
	e.focusIndex = (e.focusIndex + dir + FocusCount) % FocusCount
	e.setFocus(e.focusIndex)
}

// setFocus updates focus state for inputs.
func (e *CardEditor) setFocus(index int) {
	e.focusIndex = index
	e.skillList.SetFocused(index == FocusSkills)
	e.personaList.SetFocused(index == FocusPersonas)
	e.mcpList.SetFocused(index == FocusMCPs)

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
		return 3
	case FocusUseTmp:
		return 4
	default:
		return -1
	}
}

// updateFocusedInput routes update to the focused element.
func (e *CardEditor) updateFocusedInput(msg bubbletea.Msg) (component.Model, bubbletea.Cmd) {
	var cmd bubbletea.Cmd

	switch e.focusIndex {
	case FocusSkills:
		// Handle search input
		if keyMsg, ok := msg.(bubbletea.KeyMsg); ok {
			if len(keyMsg.Runes) > 0 {
				ch := keyMsg.Runes[0]
				if ch >= 33 && ch <= 126 {
					e.skillList.SetSearchQuery(e.skillList.searchQuery + string(ch))
					return e, nil
				}
			}
			if keyMsg.String() == "backspace" {
				if len(e.skillList.searchQuery) > 0 {
					e.skillList.SetSearchQuery(e.skillList.searchQuery[:len(e.skillList.searchQuery)-1])
				}
				return e, nil
			}
		}
		_, cmd = e.skillList.Update(msg)

	case FocusPersonas:
		if keyMsg, ok := msg.(bubbletea.KeyMsg); ok {
			if len(keyMsg.Runes) > 0 {
				ch := keyMsg.Runes[0]
				if ch >= 33 && ch <= 126 {
					e.personaList.SetSearchQuery(e.personaList.searchQuery + string(ch))
					return e, nil
				}
			}
			if keyMsg.String() == "backspace" {
				if len(e.personaList.searchQuery) > 0 {
					e.personaList.SetSearchQuery(e.personaList.searchQuery[:len(e.personaList.searchQuery)-1])
				}
				return e, nil
			}
		}
		_, cmd = e.personaList.Update(msg)

	case FocusMCPs:
		if keyMsg, ok := msg.(bubbletea.KeyMsg); ok {
			if len(keyMsg.Runes) > 0 {
				ch := keyMsg.Runes[0]
				if ch >= 33 && ch <= 126 {
					e.mcpList.SetSearchQuery(e.mcpList.searchQuery + string(ch))
					return e, nil
				}
			}
			if keyMsg.String() == "backspace" {
				if len(e.mcpList.searchQuery) > 0 {
					e.mcpList.SetSearchQuery(e.mcpList.searchQuery[:len(e.mcpList.searchQuery)-1])
				}
				return e, nil
			}
		}
		_, cmd = e.mcpList.Update(msg)

	default:
		// Update text inputs
		cmds := make([]bubbletea.Cmd, len(e.inputs))
		for i := range e.inputs {
			e.inputs[i], cmds[i] = e.inputs[i].Update(msg)
		}
		cmd = bubbletea.Batch(cmds...)
	}

	return e, cmd
}

// updateWorkspacePicker handles workspace picker modal.
func (e *CardEditor) updateWorkspacePicker(msg bubbletea.Msg) (component.Model, bubbletea.Cmd) {
	model, cmd := e.workspacePicker.Update(msg)
	e.workspacePicker = model.(*WorkspacePickerModal) //nolint:errcheck,forcetypeassert // safe type assertion

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
func (e *CardEditor) updatePuppetSelector(msg bubbletea.Msg) (component.Model, bubbletea.Cmd) {
	model, cmd := e.puppetSelector.Update(msg)
	e.puppetSelector = model.(*PuppetSelectorModal) //nolint:errcheck,forcetypeassert // safe type assertion

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
func (e *CardEditor) updateAgentsMDEditor(msg bubbletea.Msg) (component.Model, bubbletea.Cmd) {
	model, cmd := e.agentsMDEditor.Update(msg)
	e.agentsMDEditor = model.(*AgentsMDEditorModal) //nolint:errcheck,forcetypeassert // safe type assertion

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
func (e *CardEditor) deleteSelectedMCP() {
	if item, ok := e.mcpList.SelectedItem(); ok {
		id := item.GetID()
		// This should emit a message for App to handle
		// For now, just remove from selection
		delete(e.mcpList.selected, id)
	}
}

// View renders the card editor.
func (e *CardEditor) View() string {
	// Handle modal overlays
	if e.workspacePicker.IsOpen() {
		return e.renderWithOverlay(e.workspacePicker.View())
	}
	if e.puppetSelector.IsOpen() {
		return e.renderWithOverlay(e.puppetSelector.View())
	}
	if e.agentsMDEditor.IsOpen() {
		return e.renderWithOverlay(e.agentsMDEditor.View())
	}

	return e.renderForm()
}

// renderForm renders the main form.
func (e *CardEditor) renderForm() string {
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

	// Capability lists
	sb.WriteString(e.skillList.View() + "\n\n")
	sb.WriteString(e.personaList.View() + "\n\n")
	sb.WriteString(e.mcpList.View() + "\n\n")

	// Help
	sb.WriteString("Enter=save Tab=next Esc=cancel")

	return sb.String()
}

// renderWorkspaceLine renders the workspace picker line.
func (e *CardEditor) renderWorkspaceLine() string {
	summary := e.workspaceSummary()
	line := "Workspace: " + summary

	if e.focusIndex == FocusWorkspace {
		line = "> " + line
		return lipgloss.NewStyle().Foreground(highlightColor).Bold(true).Render(line)
	}
	return line
}

// workspaceSummary returns a summary of the selected workspace.
func (e *CardEditor) workspaceSummary() string {
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
	contentSummary := "empty"
	if e.agentsMD.Content != "" {
		lines := strings.Split(e.agentsMD.Content, `\n`)
		contentSummary = lines[0]
		if len(lines) > 1 || len(contentSummary) > 27 {
			if len(contentSummary) > 27 {
				contentSummary = contentSummary[:24] + "..."
			} else {
				contentSummary += "..."
			}
		}
	}

	pathSummary := "empty"
	if e.agentsMD.Path != "" {
		pathSummary = e.agentsMD.Path
		if len(pathSummary) > 27 {
			pathSummary = "..." + pathSummary[len(pathSummary)-24:]
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
		Padding(1, 2)

	return form + "\n" + overlayStyle.Render(modalContent)
}

// SetSize implements component.Model.
func (e *CardEditor) SetSize(width, height int) {
	e.Base.SetSize(width, height)
	e.skillList.SetSize(width, 5)
	e.personaList.SetSize(width, 5)
	e.mcpList.SetSize(width, 5)
}

// Capability item types

type skillItem struct {
	skill config.Skill
}

func (i skillItem) GetID() string          { return i.skill.ID }
func (i skillItem) GetLabel() string       { return i.skill.Name }
func (i skillItem) GetDescription() string { return i.skill.Description }

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
