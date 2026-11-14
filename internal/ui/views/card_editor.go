package views

import (
	"slices"
	"strings"
	"unicode"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/ui/component"
	"github.com/BlackGiack/kankuro/internal/ui/components"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/models"
	"github.com/BlackGiack/kankuro/internal/ui/styles"
)

// Focus indices for the card form. focusCount is a sentinel that must stay
// last in the block so FocusCount tracks the iota automatically.
const (
	FocusName = iota
	FocusDriver
	FocusModel
	FocusWorkspace
	FocusSubWorkspaces
	FocusFaceStyle
	FocusPuppet
	FocusUseTmp
	FocusShared
	FocusDashboardControl
	FocusAgentsMD
	FocusSkills
	FocusPersonas
	FocusMCPs
	focusCount
)

// FocusCount is the total number of potential focusable elements.
const FocusCount = focusCount

// UI constants.
const (
	hPadding = 2
	vPadding = 1
)

// Input array indices for the focus targets that are text inputs. Named so
// focusTargets doesn't scatter unexplained integer literals.
const (
	nameInputIdx = iota
	driverInputIdx
	modelInputIdx
	faceStyleInputIdx
)

// notAnInput marks a focusTarget that has no corresponding text input.
const notAnInput = -1

// focusTarget is the single source of truth for a focusable element: which
// input (if any) it maps to, what placeholder that input shows, and whether
// it is visible given the current settings. mapFocusToInput,
// visibleFocusOrder, and the input slice built in NewCardEditor are all
// derived from this table instead of being hand-kept in sync.
type focusTarget struct {
	focus       int
	inputIdx    int // index into CardEditor.inputs, or notAnInput
	placeholder string
	visible     func(models.EditorSettings) bool // nil means always visible
}

// focusTargetTable returns the focus target table. It is a function rather
// than a package-level var so each CardEditor gets its own slice/closures
// without a shared global.
func focusTargetTable() []focusTarget {
	return []focusTarget{
		{focus: FocusName, inputIdx: nameInputIdx, placeholder: "Card name..."},
		{focus: FocusDriver, inputIdx: driverInputIdx, placeholder: "Driver (e.g., claude, gemini)..."},
		{focus: FocusModel, inputIdx: modelInputIdx, placeholder: "Model (e.g., claude-3-opus)..."},
		{focus: FocusWorkspace, inputIdx: notAnInput},
		{focus: FocusSubWorkspaces, inputIdx: notAnInput},
		{focus: FocusFaceStyle, inputIdx: faceStyleInputIdx, placeholder: "Face style..."},
		{focus: FocusPuppet, inputIdx: notAnInput},
		{focus: FocusUseTmp, inputIdx: notAnInput},
		{focus: FocusShared, inputIdx: notAnInput},
		{focus: FocusDashboardControl, inputIdx: notAnInput},
		{focus: FocusAgentsMD, inputIdx: notAnInput},
		{focus: FocusSkills, inputIdx: notAnInput},
		{
			focus: FocusPersonas, inputIdx: notAnInput,
			visible: func(s models.EditorSettings) bool { return s.ShowPersonas },
		},
		{
			focus: FocusMCPs, inputIdx: notAnInput,
			visible: func(s models.EditorSettings) bool { return s.ShowMCPs },
		},
	}
}

// CardEditorCallbacks defines callbacks for card editor actions.
const (
	uncheckedMark = "[ ]"
	checkedMark   = "[x]"
)

type CardEditorCallbacks struct {
	OnSave              func(draft models.CardEditorDraft) tea.Cmd
	OnCancel            func() tea.Cmd
	OnReload            func() tea.Cmd
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
	skillList   *components.CapabilityList
	personaList *components.CapabilityList
	mcpList     *components.CapabilityList

	// Modals
	workspacePicker *WorkspacePickerModal
	puppetSelector  *PuppetSelectorModal
	agentsMDEditor  *AgentsMDEditorModal

	// State
	loaded                                                    contracts.CardEditorData
	focusIndex                                                int
	puppet                                                    string
	workspaceID                                               string
	subWorkspaceIDs                                           []string
	subWorkspaceEdited, skillEdited, personaEdited, mcpEdited bool
	useTmpWorkspace                                           bool
	dashboardControl                                          bool
	sharingEnabled                                            bool
	readOnly                                                  bool
	saveCompleted                                             bool
	saving                                                    bool
	shared                                                    bool
	dashboardID                                               string
	agentsMD                                                  AgentsMDState

	// scrollOffset is the first form line shown when the form is taller than
	// the editor. It is derived state: layout() keeps the focused block inside
	// the window and moves it as little as possible.
	scrollOffset int

	// Data
	workspaces []models.WorkspaceOption
	skills     []models.EditorOption
	personas   []models.EditorOption
	mcps       []models.EditorOption
	puppets    []models.PuppetOption
	settings   models.EditorSettings

	callbacks CardEditorCallbacks
}

// AgentsMDState holds the agents.md configuration.
type AgentsMDState struct {
	Content    string
	Path       string
	PreferPath bool
}

const (
	maxVisibleCaps = 5
	summaryLimit   = 27
)

// NewCardEditor creates a new card editor.
func NewCardEditor(callbacks CardEditorCallbacks) *CardEditor {
	// Initialize text inputs, sized and placed from focusTargets.
	inputs := make([]textinput.Model, inputCountFromTargets())
	for _, t := range focusTargetTable() {
		if t.inputIdx < 0 {
			continue
		}
		inputs[t.inputIdx] = textinput.New()
		inputs[t.inputIdx].Placeholder = t.placeholder
		inputs[t.inputIdx].SetWidth(mediumInputWidth) // Default width, will be updated in SetSize
	}

	// Initialize capability lists
	skillList := components.NewCapabilityList(components.CapabilityListConfig{
		Title:      "Skills",
		SourceHint: "ctrl+e=sources ctrl+s=selected",
		MaxVisible: maxVisibleCaps,
	})

	personaList := components.NewCapabilityList(components.CapabilityListConfig{
		Title:      "Personas",
		SourceHint: "ctrl+e=sources",
		MaxVisible: maxVisibleCaps,
	})

	mcpList := components.NewCapabilityList(components.CapabilityListConfig{
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
		settings:        models.EditorSettings{ShowPersonas: true, ShowMCPs: true},
		callbacks:       callbacks,
	}
}

func (e *CardEditor) SkillList() *components.CapabilityList { return e.skillList }

func (e *CardEditor) PersonaList() *components.CapabilityList { return e.personaList }

func (e *CardEditor) MCPList() *components.CapabilityList { return e.mcpList }

func (e *CardEditor) Settings() models.EditorSettings { return e.settings }

func (e *CardEditor) Workspaces() []models.WorkspaceOption { return slices.Clone(e.workspaces) }

func (e *CardEditor) WorkspaceID() string               { return e.workspaceID }
func (e *CardEditor) SetWorkspaceID(workspaceID string) { e.workspaceID = workspaceID }

func (e *CardEditor) Puppet() string          { return e.puppet }
func (e *CardEditor) SetPuppet(puppet string) { e.puppet = puppet }

func (e *CardEditor) EditingID() string { return e.loaded.CardID }

func (e *CardEditor) FocusIndex() int { return e.focusIndex }

func (e *CardEditor) Inputs() []textinput.Model { return e.inputs }

func (e *CardEditor) SetSettings(settings models.EditorSettings) {
	e.settings = settings
	if !e.isFocusVisible(e.focusIndex) {
		e.setFocus(e.visibleFocusOrder()[0])
	}
	if e.Width > 0 && e.Height > 0 {
		e.SetSize(e.Width, e.Height)
	}
}

// LoadEditor replaces the draft and its read metadata.
func (e *CardEditor) LoadEditor(data contracts.CardEditorData) {
	e.Reset()
	data.Values.SubWorkspaceIDs = slices.Clone(data.Values.SubWorkspaceIDs)
	data.Values.SkillIDs = slices.Clone(data.Values.SkillIDs)
	data.Values.PersonaIDs = slices.Clone(data.Values.PersonaIDs)
	data.Values.MCPIDs = slices.Clone(data.Values.MCPIDs)
	e.loaded = data
	card := data.Values
	e.inputs[nameInputIdx].SetValue(card.Name)
	e.inputs[driverInputIdx].SetValue(card.Driver)
	e.inputs[modelInputIdx].SetValue(card.Model)
	e.workspaceID = card.WorkspaceID
	e.inputs[faceStyleInputIdx].SetValue(card.FaceStyle)
	e.puppet = card.Puppet
	e.useTmpWorkspace = card.UseTmpWorkspace
	e.dashboardControl = card.DashboardControl
	e.shared = card.Shared
	e.subWorkspaceIDs = slices.Clone(card.SubWorkspaceIDs)
	e.agentsMD = AgentsMDState{Content: card.AgentsMD, Path: card.AgentsMDPath, PreferPath: card.AgentsMDPreferPath}
	e.skillList.SetSelected(card.SkillIDs)
	e.personaList.SetSelected(card.PersonaIDs)
	e.mcpList.SetSelected(card.MCPIDs)
	e.SetDashboardContext(data.DashboardID, !data.CanEdit)
}

// SetDashboardContext enables shared-card controls and optionally makes a
// consumer's view read-only. Call it after Reset or SetCard.
func (e *CardEditor) SetDashboardContext(dashboardID string, readOnly bool) {
	e.dashboardID = dashboardID
	e.sharingEnabled = dashboardID != ""
	e.readOnly = readOnly
	if readOnly {
		e.workspacePicker.Close()
		e.puppetSelector.Close()
		e.agentsMDEditor.Close()
	}
	if !e.isFocusVisible(e.focusIndex) {
		e.setFocus(FocusName)
	}
}

// Reset clears the form for creating a new card.
func (e *CardEditor) Reset() {
	for i := range e.inputs {
		e.inputs[i].SetValue("")
		e.inputs[i].Blur()
	}
	e.loaded = contracts.CardEditorData{}
	e.saveCompleted = false
	e.saving = false
	e.workspacePicker.Close()
	e.puppetSelector.Close()
	e.agentsMDEditor.Close()
	e.clearSearches()
	e.sharingEnabled, e.readOnly, e.dashboardID = false, false, ""
	e.workspaceID = ""
	e.subWorkspaceIDs = nil
	e.subWorkspaceEdited, e.skillEdited, e.personaEdited, e.mcpEdited = false, false, false, false
	e.useTmpWorkspace = false
	e.dashboardControl = false
	e.shared = false
	e.skillList.SetSelectedOnly(false)
	e.puppet = "kankuro"
	e.agentsMD = AgentsMDState{}
	e.skillList.SetSelected(nil)
	e.personaList.SetSelected(nil)
	e.mcpList.SetSelected(nil)
	e.focusIndex = 0
	e.setFocus(0)
}

// SetSaving prevents draft mutations while a submitted snapshot is in flight.
func (e *CardEditor) SetSaving(saving bool) { e.saving = saving }

// MarkSaveCompleted preserves the draft while requiring an explicit reload
// before further saves after an offscreen completion.
func (e *CardEditor) MarkSaveCompleted() {
	e.saveCompleted = true
	e.workspacePicker.Close()
	e.puppetSelector.Close()
	e.agentsMDEditor.Close()
}

// SetChoices applies independently successful option reads; failures retain the previous options.
func (e *CardEditor) SetChoices(choices contracts.CardEditorChoices) {
	if choices.Workspaces.Attempted && choices.Workspaces.Err == nil {
		e.workspaces = make([]models.WorkspaceOption, 0, len(choices.Workspaces.Value))
		for _, v := range choices.Workspaces.Value {
			e.workspaces = append(e.workspaces, models.WorkspaceOption{ID: v.ID, Name: v.Name, Path: v.Path})
		}
		e.workspacePicker.SetWorkspaces(e.workspaces)
	}
	if choices.Skills.Attempted && choices.Skills.Err == nil {
		e.skills = editorOptions(choices.Skills.Value)
		e.skillList.SetItems(capabilityItems(e.skills))
	}
	if choices.Personas.Attempted && choices.Personas.Err == nil {
		e.personas = editorOptions(choices.Personas.Value)
		e.personaList.SetItems(capabilityItems(e.personas))
	}
	if choices.MCPs.Attempted && choices.MCPs.Err == nil {
		e.mcps = editorOptions(choices.MCPs.Value)
		e.mcpList.SetItems(capabilityItems(e.mcps))
	}
	if choices.Puppets.Attempted && choices.Puppets.Err == nil {
		e.puppets = make([]models.PuppetOption, 0, len(choices.Puppets.Value))
		for _, v := range choices.Puppets.Value {
			e.puppets = append(e.puppets, models.PuppetOption{ID: v.ID, Name: v.Name, Lines: slices.Clone(v.Lines), Builtin: v.Builtin})
		}
		e.puppetSelector.SetPuppets(e.puppets)
	}
}
func editorOptions(values []contracts.EditorOption) []models.EditorOption {
	out := make([]models.EditorOption, 0, len(values))
	for _, v := range values {
		out = append(out, models.EditorOption{ID: v.ID, Name: v.Name, Description: v.Description})
	}
	return out
}
func capabilityItems(options []models.EditorOption) []components.CapabilityItem {
	items := make([]components.CapabilityItem, 0, len(options))
	for _, option := range options {
		items = append(items, editorOptionItem{option})
	}
	return items
}

type editorOptionItem struct{ option models.EditorOption }

func (i editorOptionItem) GetID() string          { return i.option.ID }
func (i editorOptionItem) GetLabel() string       { return i.option.Name }
func (i editorOptionItem) GetDescription() string { return i.option.Description }

// Draft captures detached editable values and explicit changes.
func (e *CardEditor) Draft() models.CardEditorDraft {
	values := contracts.CardValues{
		Name:               strings.TrimSpace(e.inputs[nameInputIdx].Value()),
		Driver:             strings.TrimSpace(e.inputs[driverInputIdx].Value()),
		Model:              strings.TrimSpace(e.inputs[modelInputIdx].Value()),
		WorkspaceID:        e.workspaceID,
		SubWorkspaceIDs:    slices.Clone(e.subWorkspaceIDs),
		FaceStyle:          strings.TrimSpace(e.inputs[faceStyleInputIdx].Value()),
		Puppet:             e.puppet,
		UseTmpWorkspace:    e.useTmpWorkspace,
		DashboardControl:   e.dashboardControl,
		Shared:             e.shared,
		AgentsMD:           e.agentsMD.Content,
		AgentsMDPath:       e.agentsMD.Path,
		AgentsMDPreferPath: e.agentsMD.PreferPath,
		SkillIDs:           e.skillList.AllSelectedIDs(),
		PersonaIDs:         e.personaList.AllSelectedIDs(),
		MCPIDs:             e.mcpList.AllSelectedIDs(),
	}
	if e.skillEdited && len(values.SkillIDs) == 0 {
		values.SkillIDs = []string{}
	}
	if e.personaEdited && len(values.PersonaIDs) == 0 {
		values.PersonaIDs = []string{}
	}
	if e.mcpEdited && len(values.MCPIDs) == 0 {
		values.MCPIDs = []string{}
	}
	old := e.loaded.Values
	changes := contracts.CardChanges{}
	changes.Name = contracts.Change[string]{Present: values.Name != old.Name, Value: values.Name}
	changes.Driver = contracts.Change[string]{Present: values.Driver != old.Driver, Value: values.Driver}
	changes.Model = contracts.Change[string]{Present: values.Model != old.Model, Value: values.Model}
	changes.WorkspaceID = contracts.Change[string]{Present: values.WorkspaceID != old.WorkspaceID, Value: values.WorkspaceID}
	changes.SubWorkspaceIDs = contracts.Change[[]string]{
		Present: e.subWorkspaceEdited || !slices.Equal(values.SubWorkspaceIDs, old.SubWorkspaceIDs),
		Value:   values.SubWorkspaceIDs,
	}
	changes.FaceStyle = contracts.Change[string]{Present: values.FaceStyle != old.FaceStyle, Value: values.FaceStyle}
	changes.Puppet = contracts.Change[string]{Present: values.Puppet != old.Puppet, Value: values.Puppet}
	changes.UseTmpWorkspace = contracts.Change[bool]{Present: values.UseTmpWorkspace != old.UseTmpWorkspace, Value: values.UseTmpWorkspace}
	changes.DashboardControl = contracts.Change[bool]{
		Present: values.DashboardControl != old.DashboardControl,
		Value:   values.DashboardControl,
	}
	changes.Shared = contracts.Change[bool]{Present: values.Shared != old.Shared, Value: values.Shared}
	changes.AgentsMD = contracts.Change[string]{Present: values.AgentsMD != old.AgentsMD, Value: values.AgentsMD}
	changes.AgentsMDPath = contracts.Change[string]{Present: values.AgentsMDPath != old.AgentsMDPath, Value: values.AgentsMDPath}
	changes.AgentsMDPreferPath = contracts.Change[bool]{
		Present: values.AgentsMDPreferPath != old.AgentsMDPreferPath,
		Value:   values.AgentsMDPreferPath,
	}
	changes.SkillIDs = contracts.Change[[]string]{
		Present: e.skillEdited || !slices.Equal(values.SkillIDs, old.SkillIDs),
		Value:   values.SkillIDs,
	}
	changes.PersonaIDs = contracts.Change[[]string]{
		Present: e.personaEdited || !slices.Equal(values.PersonaIDs, old.PersonaIDs),
		Value:   values.PersonaIDs,
	}
	changes.MCPIDs = contracts.Change[[]string]{Present: e.mcpEdited || !slices.Equal(values.MCPIDs, old.MCPIDs), Value: values.MCPIDs}
	return models.CardEditorDraft{Values: values, Changes: changes}
}

// Init initializes the editor.
func (e *CardEditor) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (e *CardEditor) Update(msg tea.Msg) (component.Model, tea.Cmd) {
	if cmd, handled := e.handlePendingSave(msg); handled {
		return e, cmd
	}
	if e.readOnly {
		return e, e.handleReadOnly(msg)
	}
	// Handle data refresh messages
	if msg, ok := msg.(messages.MCPDeletedMsg); ok {
		// Remove the deleted MCP from selection
		e.mcpList.Deselect(msg.MCPID)
		e.mcpEdited = true
		// Adjust cursor if needed
		if e.mcpList.Cursor() >= len(e.mcpList.Items()) && e.mcpList.Cursor() > 0 {
			e.mcpList.SetCursor(e.mcpList.Cursor() - 1)
		}
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
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		if cmd, handled := e.handleKey(keyMsg); handled {
			return e, cmd
		}
	}

	// Route to focused input
	return e.updateFocusedInput(msg)
}

func (e *CardEditor) handleReadOnly(msg tea.Msg) tea.Cmd {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}
	switch key.Code {
	case tea.KeyEsc:
		if e.callbacks.OnCancel != nil {
			return e.callbacks.OnCancel()
		}
	case tea.KeyTab:
		e.CycleFocus(1)
	case tea.KeyUp:
		e.handleNav(false)
	case tea.KeyDown:
		e.handleNav(true)
	}
	return nil
}

// handlePendingSave keeps an in-flight or completed snapshot from being edited.
func (e *CardEditor) handlePendingSave(msg tea.Msg) (tea.Cmd, bool) {
	if !e.saving && !e.saveCompleted {
		return nil, false
	}
	key, isKey := msg.(tea.KeyPressMsg)
	if !isKey {
		return nil, true
	}
	if e.saving {
		return nil, true
	}
	if key.Code == tea.KeyEsc && e.callbacks.OnCancel != nil {
		return e.callbacks.OnCancel(), true
	}
	if e.saveCompleted && key.String() == "ctrl+r" && e.callbacks.OnReload != nil {
		return e.callbacks.OnReload(), true
	}
	return nil, true
}

// handleKey processes global key events.
func (e *CardEditor) handleKey(keyMsg tea.KeyPressMsg) (tea.Cmd, bool) {
	switch editorKeyIdentity(keyMsg) {
	case keyCtrlC:
		return tea.Quit, true
	case keyShiftTab:
		e.clearSearches()
		e.CycleFocus(-1)
		return nil, true
	case "esc":
		e.handleEscape()
		if e.callbacks.OnCancel != nil {
			return e.callbacks.OnCancel(), true
		}
		return nil, true

	case "ctrl+r":
		if e.callbacks.OnReload != nil {
			return e.callbacks.OnReload(), true
		}
		return nil, true

	case "tab":
		e.clearSearches()
		e.CycleFocus(1)
		return nil, true

	case "up":
		e.handleNav(false)
		return nil, true

	case "down":
		e.handleNav(true)
		return nil, true

	case "space":
		return nil, e.handleSpace()

	case "enter":
		return e.handleSubmit(), true

	case "delete":
		if e.focusIndex != FocusMCPs {
			return nil, false
		}
		return e.deleteSelectedMCP(), true
	case keyCtrlS:
		if e.focusIndex != FocusSkills {
			return nil, false
		}
		e.skillList.SetSelectedOnly(!e.skillList.SelectedOnly())
		e.skillList.SetCursor(0)
		return nil, true
	case keyCtrlX:
		if e.focusIndex != FocusSubWorkspaces {
			return nil, false
		}
		e.subWorkspaceIDs = []string{}
		e.subWorkspaceEdited = true
		return nil, true
	case keyCtrlE:
		return e.handleCtrlE()
	}
	return nil, false
}

func editorKeyIdentity(keyMsg tea.KeyPressMsg) string {
	identity := keyMsg.String()
	switch identity {
	case keyCtrlC, keyCtrlS, keyCtrlE, keyCtrlX, keyShiftTab:
		return identity
	}
	switch keyMsg.Code {
	case tea.KeyEsc:
		return "esc"
	case tea.KeyTab:
		return "tab"
	case tea.KeyUp:
		return "up"
	case tea.KeyDown:
		return "down"
	case tea.KeySpace:
		return "space"
	case tea.KeyEnter:
		return "enter"
	case tea.KeyDelete:
		return "delete"
	default:
		return identity
	}
}

// handleEscape handles escape key based on current state.
func (e *CardEditor) handleEscape() {
	// Clear search if active
	switch e.focusIndex {
	case FocusSkills:
		if e.skillList.SearchQuery != "" {
			e.skillList.SetSearchQuery("")
			return
		}
	case FocusPersonas:
		if e.personaList.SearchQuery != "" {
			e.personaList.SetSearchQuery("")
			return
		}
	case FocusMCPs:
		if e.mcpList.SearchQuery != "" {
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
	case FocusSubWorkspaces:
		e.workspacePicker.SetMultiple(true)
		e.workspacePicker.SetExcludedID(e.workspaceID)
		e.workspacePicker.SetSelectedIDs(e.subWorkspaceIDs)
		e.workspacePicker.Open()
		return true
	case FocusUseTmp:
		e.useTmpWorkspace = !e.useTmpWorkspace
		return true
	case FocusShared:
		if e.sharingEnabled {
			e.shared = !e.shared
		}
		return true
	case FocusDashboardControl:
		e.dashboardControl = !e.dashboardControl
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
		if _, ok := e.skillList.SelectedItem(); ok {
			e.skillList.ToggleSelected()
			e.skillEdited = true
		}
		return true

	case FocusPersonas:
		if !e.settings.ShowPersonas {
			return false
		}
		if _, ok := e.personaList.SelectedItem(); ok {
			e.personaList.ToggleSelected()
			e.personaEdited = true
		}
		return true

	case FocusMCPs:
		if !e.settings.ShowMCPs {
			return false
		}
		if _, ok := e.mcpList.SelectedItem(); ok {
			e.mcpList.ToggleSelected()
			e.mcpEdited = true
		}
		// Auto-select mcp-cli skill
		e.autoSelectMCPCLISkill()
		return true
	}

	return false
}

// handleCtrlE handles ctrl+e for opening source editors.
func (e *CardEditor) handleCtrlE() (tea.Cmd, bool) {
	switch e.focusIndex {
	case FocusWorkspace, FocusSubWorkspaces:
		if e.callbacks.OnOpenWorkspaces != nil {
			return e.callbacks.OnOpenWorkspaces(), true
		}
		return nil, true
	case FocusSkills:
		if e.callbacks.OnOpenSources != nil {
			return e.callbacks.OnOpenSources("skills"), true
		}
		return nil, true

	case FocusPersonas:
		if !e.settings.ShowPersonas {
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
		if !e.settings.ShowMCPs {
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
		if !e.settings.ShowPersonas {
			return
		}
		if down {
			e.personaList.SetCursor(e.personaList.Cursor() + 1)
		} else {
			e.personaList.SetCursor(e.personaList.Cursor() - 1)
		}

	case FocusMCPs:
		if !e.settings.ShowMCPs {
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
	draft := e.Draft()

	if e.callbacks.OnSave != nil {
		return e.callbacks.OnSave(draft)
	}

	e.loaded.CardID = ""
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
	e.personaList.SetFocused(index == FocusPersonas && e.settings.ShowPersonas)
	e.mcpList.SetFocused(index == FocusMCPs && e.settings.ShowMCPs)
	defer e.syncScroll()

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

// mapFocusToInput maps logical focus index to input array index, looked up
// from focusTargets.
func (e *CardEditor) mapFocusToInput(index int) int {
	for _, t := range focusTargetTable() {
		if t.focus == index {
			return t.inputIdx
		}
	}
	return -1
}

// inputCountFromTargets returns how many focusTargets map to a text input.
func inputCountFromTargets() int {
	n := 0
	for _, t := range focusTargetTable() {
		if t.inputIdx >= 0 {
			n++
		}
	}
	return n
}

// updateFocusedInput routes update to the focused element.
func (e *CardEditor) updateFocusedInput(msg tea.Msg) (component.Model, tea.Cmd) {
	switch e.focusIndex {
	case FocusSkills:
		return e, e.updateSkillList(msg)
	case FocusPersonas:
		if !e.settings.ShowPersonas {
			return e, nil
		}
		return e, e.updatePersonaList(msg)
	case FocusMCPs:
		if !e.settings.ShowMCPs {
			return e, nil
		}
		return e, e.updateMCPList(msg)
	default:
		return e, e.updateInputs(msg)
	}
}

func (e *CardEditor) updateSkillList(msg tea.Msg) tea.Cmd {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		if e.handleListSearch(e.skillList, keyMsg) {
			return nil
		}
	}
	_, cmd := e.skillList.Update(msg)
	return cmd
}

func (e *CardEditor) updatePersonaList(msg tea.Msg) tea.Cmd {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		if e.handleListSearch(e.personaList, keyMsg) {
			return nil
		}
	}
	_, cmd := e.personaList.Update(msg)
	return cmd
}

func (e *CardEditor) updateMCPList(msg tea.Msg) tea.Cmd {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		if e.handleListSearch(e.mcpList, keyMsg) {
			return nil
		}
	}
	_, cmd := e.mcpList.Update(msg)
	return cmd
}

func (e *CardEditor) handleListSearch(list *components.CapabilityList, keyMsg tea.KeyPressMsg) bool {
	if len(keyMsg.Text) > 0 {
		for _, ch := range keyMsg.Text {
			if unicode.IsPrint(ch) && !unicode.IsSpace(ch) {
				list.SetSearchQuery(list.SearchQuery + string(ch))
				return true
			}
		}
	}
	if keyMsg.Code == tea.KeyBackspace {
		if len(list.SearchQuery) > 0 {
			list.SetSearchQuery(list.SearchQuery[:len(list.SearchQuery)-1])
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
		if e.focusIndex == FocusSubWorkspaces {
			e.subWorkspaceIDs = e.workspacePicker.SelectedIDs()
			e.subWorkspaceEdited = true
			e.workspacePicker.SetMultiple(false)
		} else {
			e.workspaceID = e.workspacePicker.SelectedID()
		}
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
			e.skillList.Select(s.ID)
			e.skillEdited = true
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

// formBlock is a run of form lines that scroll as one unit. focus is the
// focus target the block contains, so the editor can keep the focused block on
// screen when the form is taller than its box.
type formBlock struct {
	focus int
	lines []string
}

// blockOf splits text (every block ends with a newline; capability lists end
// with several) into a block's lines.
func blockOf(focus int, text string) formBlock {
	return formBlock{focus: focus, lines: strings.Split(strings.TrimSuffix(text, "\n"), "\n")}
}

// fieldBlocks returns the blocks above the capability lists: one per field
// line and one for the AGENTS.md row. The title belongs to the first block so
// that scrolling to the first field scrolls all the way to the top.
func (e *CardEditor) fieldBlocks() []formBlock {
	title := "New ID Card"
	if e.loaded.CardID != "" {
		title = "Edit ID Card"
	}
	if e.readOnly {
		title = "ID Card (read-only)"
	}
	if e.saveCompleted {
		title = "Card saved — reload to continue"
	}
	if e.sharingEnabled {
		owner := e.loaded.OwnerDashboardID
		if owner == "" {
			owner = e.dashboardID
		}
		title += "\nOwner: " + components.DisplayText(owner)
		if e.shared {
			title += "\nShared card edits affect every dashboard."
		}
	}

	blocks := []formBlock{
		blockOf(FocusName, styles.TitleStyle.Render(title)+"\n\n"+"Name: "+e.inputs[nameInputIdx].View()+"\n"),
		blockOf(FocusDriver, "Driver: "+e.inputs[driverInputIdx].View()+"\n"),
		blockOf(FocusModel, "Model: "+e.inputs[modelInputIdx].View()+"\n"),
		blockOf(FocusWorkspace, e.renderWorkspaceLine()+"\n"),
		blockOf(FocusSubWorkspaces, e.renderSubWorkspacesLine()+"\n"),
		blockOf(FocusFaceStyle, "Face Style: "+e.inputs[faceStyleInputIdx].View()+"\n"),
		blockOf(FocusPuppet, e.renderPuppetLine()+"\n"),
		blockOf(FocusUseTmp, e.renderTmpWorkspaceLine()+"\n\n"),
	}
	if e.sharingEnabled {
		blocks = append(blocks, blockOf(FocusShared, e.renderSharedLine()+"\n"))
	}
	return append(blocks,
		blockOf(FocusDashboardControl, e.renderDashboardControlLine()+"\n"),
		blockOf(FocusAgentsMD, "AGENTS.md:\n"+e.renderAgentsMDRow()+"\n\n"),
	)
}

func (e *CardEditor) renderSharedLine() string {
	mark := uncheckedMark
	if e.shared {
		mark = checkedMark
	}
	line := "Shared: " + mark + " (available across dashboards)"
	if e.focusIndex == FocusShared {
		return lipgloss.NewStyle().Foreground(styles.HighlightColor).Bold(true).Render(focusMarker + line)
	}
	return line
}

// listBlock is a capability list followed by the blank rows that separate it
// from what is below. A sized list renders exactly its height with no trailing
// newline; an unsized one ends with one.
func listBlock(focus int, l *components.CapabilityList) formBlock {
	view := l.View()
	if l.Height <= 0 {
		// Only the unsized list's newline is a terminator. A sized list's last
		// row may itself be blank and must stay.
		view = strings.TrimSuffix(view, "\n")
	}
	return blockOf(focus, view+"\n\n\n")
}

// listBlocks returns one block per visible capability list.
func (e *CardEditor) listBlocks() []formBlock {
	blocks := []formBlock{listBlock(FocusSkills, e.skillList)}
	if e.settings.ShowPersonas {
		blocks = append(blocks, listBlock(FocusPersonas, e.personaList))
	}
	if e.settings.ShowMCPs {
		blocks = append(blocks, listBlock(FocusMCPs, e.mcpList))
	}
	return blocks
}

// formLayout is the whole form as lines, with where the focused block sits.
type formLayout struct {
	rows       []string
	focusStart int // first line of the focused block
	focusEnd   int // one past its last non-blank line
}

// layout flattens the blocks. When the form is taller than win the blank
// lines that trail it are dropped: they would only be scrolled to.
func (e *CardEditor) layout(win int) formLayout {
	var l formLayout
	for _, b := range append(e.fieldBlocks(), e.listBlocks()...) {
		start := len(l.rows)
		l.rows = append(l.rows, b.lines...)
		if b.focus == e.focusIndex {
			l.focusStart = start
			l.focusEnd = start + len(b.lines)
			for l.focusEnd > l.focusStart+1 && b.lines[l.focusEnd-start-1] == "" {
				l.focusEnd--
			}
		}
	}
	if win > 0 && len(l.rows) > win {
		for len(l.rows) > win && l.rows[len(l.rows)-1] == "" {
			l.rows = l.rows[:len(l.rows)-1]
		}
	}
	return l
}

// formWindow is how many form lines fit above the pinned help line, or 0 when
// the editor has not been given a size (then nothing scrolls).
func (e *CardEditor) formWindow() int {
	if e.Height <= 0 {
		return 0
	}
	return max(e.Height-helpLineHeight, 1)
}

// minIndicatorWindow is the smallest window that can spare two rows for the
// "more" indicators and still show a line of the form.
const minIndicatorWindow = 3

// syncScroll moves scrollOffset the least that puts the focused block fully
// inside the window, leaving the rows the "↑ more" / "↓ more" indicators
// occupy out of it. It returns the layout it computed against.
func (e *CardEditor) syncScroll() (formLayout, int) {
	win := e.formWindow()
	l := e.layout(win)
	if win <= 0 || len(l.rows) <= win {
		e.scrollOffset = 0
		return l, win
	}

	indicator := 0
	if win >= minIndicatorWindow {
		indicator = 1
	}
	maxOff := len(l.rows) - win
	off := min(max(e.scrollOffset, 0), maxOff)
	bottom := func(off int) int {
		if off+win < len(l.rows) {
			return off + win - indicator
		}
		return off + win
	}
	top := func(off int) int {
		if off > 0 {
			return off + indicator
		}
		return 0
	}

	if l.focusEnd > bottom(off) {
		off = min(l.focusEnd-win+indicator, maxOff)
	}
	// Checked second so that a block taller than the window shows its start.
	if l.focusStart < top(off) {
		off = max(l.focusStart-indicator, 0)
	}
	e.scrollOffset = min(max(off, 0), maxOff)
	return l, win
}

// renderForm renders the main form: the visible slice of the form lines, and
// the help line pinned to the last row.
func (e *CardEditor) renderForm() string {
	help := "Enter=save Tab=next Ctrl+R=reload Esc=cancel"
	if e.readOnly {
		help = "Read-only Tab=next Esc=back"
	}
	if e.saveCompleted {
		help = "Saved: Ctrl+R=reload Esc=back"
	}
	if e.saving {
		help = "Saving..."
	}

	l, win := e.syncScroll()
	rows := e.visibleRows(l, win)
	if len(rows) == 0 {
		return help
	}
	return strings.Join(rows, "\n") + "\n" + help
}

// visibleRows is the slice of the form shown in the window, with the dim
// "more" indicators standing in for the first and last row when content is cut.
func (e *CardEditor) visibleRows(l formLayout, win int) []string {
	if win <= 0 || len(l.rows) <= win {
		return l.rows
	}
	rows := slices.Clone(l.rows[e.scrollOffset:min(e.scrollOffset+win, len(l.rows))])
	if win < minIndicatorWindow {
		return rows
	}
	more := lipgloss.NewStyle().Foreground(styles.Gray)
	if e.scrollOffset > 0 {
		rows[0] = more.Render("↑ more")
	}
	if e.scrollOffset+win < len(l.rows) {
		rows[len(rows)-1] = more.Render("↓ more")
	}
	return rows
}

// fitField clips user text sitting after a label of the given display width
// to the editor's width, so a long value cannot widen the line. An editor
// that has no size yet does not clip.
func (e *CardEditor) fitField(text string, labelWidth int) string {
	text = components.DisplayText(text)
	if e.Width <= 0 {
		return text
	}
	return components.TruncateHead(text, e.Width-labelWidth)
}

const (
	workspaceLabel = "Workspace: "
	puppetLabel    = "Puppet: "
	focusMarker    = "> "
)

// renderWorkspaceLine renders the workspace picker line.
func (e *CardEditor) renderWorkspaceLine() string {
	// Budget for the focused prefix in both states, so the line keeps its
	// width when focus moves onto it.
	line := workspaceLabel + e.fitField(e.WorkspaceSummary()+" (ctrl+e=sources)", len(focusMarker)+len(workspaceLabel))

	if e.focusIndex == FocusWorkspace {
		line = focusMarker + line
		return lipgloss.NewStyle().Foreground(styles.HighlightColor).Bold(true).Render(line)
	}
	return line
}

func (e *CardEditor) renderTmpWorkspaceLine() string {
	mark := uncheckedMark
	if e.useTmpWorkspace {
		mark = checkedMark
	}
	line := "Use Tmp Workspace: " + mark
	if e.focusIndex == FocusUseTmp {
		line = focusMarker + line
		return lipgloss.NewStyle().Foreground(styles.HighlightColor).Bold(true).Render(line)
	}
	return line
}

func (e *CardEditor) renderDashboardControlLine() string {
	mark := uncheckedMark
	if e.dashboardControl {
		mark = checkedMark
	}
	line := "Dashboard Control: " + mark + " (linked-dashboard session access; requires Kankuro MCP)"
	if e.focusIndex == FocusDashboardControl {
		line = focusMarker + line
		return lipgloss.NewStyle().Foreground(styles.HighlightColor).Bold(true).Render(line)
	}
	return line
}

func (e *CardEditor) renderSubWorkspacesLine() string {
	names := make([]string, 0, len(e.subWorkspaceIDs))
	for _, id := range e.subWorkspaceIDs {
		name := id
		for _, ws := range e.workspaces {
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
	label := "Sub Workspaces: "
	hint := " (ctrl+e=sources)"
	focused := e.focusIndex == FocusSubWorkspaces
	if focused {
		hint = " (ctrl+e=sources ctrl+x=clear)"
	}
	if e.Width > 0 {
		value = components.TruncateHead(components.DisplayText(value),
			e.Width-len(focusMarker)-len(label)-lipgloss.Width(hint))
	}
	line := label + value + hint
	if focused {
		line = focusMarker + line
		return lipgloss.NewStyle().Foreground(styles.HighlightColor).Bold(true).Render(line)
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
	line := puppetLabel + e.fitField(e.puppet, len(focusMarker)+len(puppetLabel))

	if e.focusIndex == FocusPuppet {
		line = focusMarker + line
		return lipgloss.NewStyle().Foreground(styles.HighlightColor).Bold(true).Render(line)
	}
	return line
}

// renderAgentsMDRow renders the agents.md options.
func (e *CardEditor) renderAgentsMDRow() string {
	contentSummary := "empty"
	if e.agentsMD.Content != "" {
		lines := strings.Split(e.agentsMD.Content, "\n")
		first := components.DisplayText(lines[0])
		contentSummary = components.TruncateHead(first, summaryLimit)
		if len(lines) > 1 && contentSummary == first {
			contentSummary += "..."
		}
	}

	pathSummary := "empty"
	if e.agentsMD.Path != "" {
		pathSummary = components.TruncateTail(components.DisplayText(e.agentsMD.Path), summaryLimit)
	}

	contentOption := "[ Content: " + contentSummary + " ]"
	pathOption := "[ Path: " + pathSummary + " ]"

	if e.focusIndex == FocusAgentsMD {
		prefix := focusMarker
		if e.agentsMD.PreferPath {
			return lipgloss.NewStyle().Foreground(styles.HighlightColor).Bold(true).Render(
				prefix + "  " + contentOption + "\n" + prefix + "x " + pathOption,
			)
		}
		return lipgloss.NewStyle().Foreground(styles.HighlightColor).Bold(true).Render(
			prefix + "x " + contentOption + "\n" + prefix + "  " + pathOption,
		)
	}

	if e.agentsMD.PreferPath {
		return lipgloss.NewStyle().Foreground(styles.Gray).Render("    " + contentOption + "\n  x " + pathOption)
	}
	return lipgloss.NewStyle().Foreground(styles.Gray).Render("  x " + contentOption + "\n    " + pathOption)
}

// renderWithOverlay renders form with modal overlay.
func (e *CardEditor) renderWithOverlay(modalContent string) string {
	return lipgloss.Place(e.Width, e.Height, lipgloss.Center, lipgloss.Center, modalContent)
}

const (
	helpLineHeight    = 1
	listSeparatorRows = 2
	minCapListHeight  = 2
)

// SetSize implements component.Model.
func (e *CardEditor) SetSize(width, height int) {
	e.Base.SetSize(width, height)

	headerHeight := 0
	for _, b := range e.fieldBlocks() {
		headerHeight += len(b.lines)
	}

	part := 1
	if e.settings.ShowPersonas {
		part++
	}
	if e.settings.ShowMCPs {
		part++
	}

	separatorsAndHelp := (part * listSeparatorRows) + helpLineHeight
	available := height - headerHeight - separatorsAndHelp

	// Floored: when the lists do not fit, the form scrolls instead of
	// squeezing them further.
	capListHeight := max(minCapListHeight, available/part)

	e.skillList.SetSize(width, capListHeight)
	e.personaList.SetSize(width, capListHeight)
	e.mcpList.SetSize(width, capListHeight)

	// Give modals full dimensions
	e.workspacePicker.SetSize(width, height)
	e.agentsMDEditor.SetSize(width, height)
	e.puppetSelector.SetSize(width, height)

	// Update input widths. A text input draws its prompt and cursor on top of
	// the width it is given, and sits after the widest label.
	calculatedWidth := min(
		width-widestInputLabel-inputChrome,
		max(40, width*3/5), //nolint:mnd // responsive sizing
	)
	for i := range e.inputs {
		e.inputs[i].SetWidth(max(1, calculatedWidth))
	}

	e.syncScroll()
}

// widestInputLabel is the display width of the longest label in front of a
// text input ("Use Tmp Workspace: "); inputChrome is what textinput draws
// beyond its width (the "> " prompt and the cursor cell).
const (
	widestInputLabel = len("Use Tmp Workspace: ")
	inputChrome      = 3
)

// visibleFocusOrder returns the display/tab order of currently visible focus
// targets, derived from focusTargets so a new entry there is automatically
// included (and automatically reachable by Tab/wrap-around).
func (e *CardEditor) visibleFocusOrder() []int {
	targets := focusTargetTable()
	order := make([]int, 0, len(targets))
	for _, t := range targets {
		if t.focus == FocusShared && !e.sharingEnabled {
			continue
		}
		if t.visible == nil || t.visible(e.settings) {
			order = append(order, t.focus)
		}
	}
	return order
}

func (e *CardEditor) isFocusVisible(index int) bool {
	return slices.Contains(e.visibleFocusOrder(), index)
}

// Capability item types
