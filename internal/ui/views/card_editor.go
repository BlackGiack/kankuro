package views

import (
	"slices"
	"strings"
	"unicode"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/ui/component"
	"github.com/BlackGiack/kankuro/internal/ui/components"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
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
	visible     func(config.AppSettings) bool // nil means always visible
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
		{focus: FocusDashboardControl, inputIdx: notAnInput},
		{focus: FocusAgentsMD, inputIdx: notAnInput},
		{focus: FocusSkills, inputIdx: notAnInput},
		{
			focus: FocusPersonas, inputIdx: notAnInput,
			visible: func(s config.AppSettings) bool { return s.ShowPersonasInCardEditor },
		},
		{
			focus: FocusMCPs, inputIdx: notAnInput,
			visible: func(s config.AppSettings) bool { return s.ShowMCPsInCardEditor },
		},
	}
}

// CardEditorCallbacks defines callbacks for card editor actions.
type CardEditorCallbacks struct {
	OnSave              func(card config.IDCard) tea.Cmd
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
	skillList   *components.CapabilityList
	personaList *components.CapabilityList
	mcpList     *components.CapabilityList

	// Modals
	workspacePicker *WorkspacePickerModal
	puppetSelector  *PuppetSelectorModal
	agentsMDEditor  *AgentsMDEditorModal

	// State
	loaded           config.IDCard // card as loaded by SetCard; Card() overlays edits onto this
	focusIndex       int
	puppet           string
	workspaceID      string
	subWorkspaceIDs  []string
	useTmpWorkspace  bool
	dashboardControl bool
	agentsMD         AgentsMDState

	// scrollOffset is the first form line shown when the form is taller than
	// the editor. It is derived state: layout() keeps the focused block inside
	// the window and moves it as little as possible.
	scrollOffset int

	// Data
	workspaces []config.Workspace
	skills     []config.Skill
	personas   []config.Persona
	mcps       []config.MCPServer
	puppets    []config.Puppet
	settings   config.AppSettings

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
		settings:        config.DefaultAppSettings(),
		callbacks:       callbacks,
	}
}

func (e *CardEditor) SkillList() *components.CapabilityList { return e.skillList }

func (e *CardEditor) PersonaList() *components.CapabilityList { return e.personaList }

func (e *CardEditor) MCPList() *components.CapabilityList { return e.mcpList }

func (e *CardEditor) Settings() config.AppSettings { return e.settings }

func (e *CardEditor) Workspaces() []config.Workspace { return e.workspaces }

func (e *CardEditor) WorkspaceID() string               { return e.workspaceID }
func (e *CardEditor) SetWorkspaceID(workspaceID string) { e.workspaceID = workspaceID }

func (e *CardEditor) Puppet() string          { return e.puppet }
func (e *CardEditor) SetPuppet(puppet string) { e.puppet = puppet }

func (e *CardEditor) EditingID() string { return e.loaded.ID }

func (e *CardEditor) FocusIndex() int { return e.focusIndex }

func (e *CardEditor) Inputs() []textinput.Model { return e.inputs }

func (e *CardEditor) SetSettings(settings config.AppSettings) {
	// A section going from visible to hidden must capture whatever is
	// currently selected into e.loaded, since Card() falls back to e.loaded's
	// value for a hidden section instead of reading the list.
	//
	// Guard on the list having items rather than the selection being
	// non-empty: SelectedIDs() only returns IDs whose items are currently
	// loaded (finding #17), so before SetData has populated the list,
	// "nothing selected" is indistinguishable from "list not populated yet"
	// and must not wipe a real selection. Checking Items() makes that
	// distinction, so a genuine deselect-everything edit on a populated list
	// is still captured (TestCardEditor_HiddenSectionsCaptureDeselection).
	if e.settings.ShowPersonasInCardEditor && !settings.ShowPersonasInCardEditor {
		if len(e.personaList.Items()) > 0 {
			e.loaded.PersonaIDs = e.personaList.SelectedIDs()
		}
	}
	if e.settings.ShowMCPsInCardEditor && !settings.ShowMCPsInCardEditor {
		if len(e.mcpList.Items()) > 0 {
			e.loaded.MCPIDs = e.mcpList.SelectedIDs()
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

// SetCard populates the form for editing an existing card. The full card is
// retained in e.loaded so that Card() can overlay edits onto it without
// dropping fields the form does not expose (e.g. SystemInstruction, Env).
func (e *CardEditor) SetCard(card config.IDCard) {
	e.loaded = card
	e.inputs[nameInputIdx].SetValue(card.Name)
	e.inputs[driverInputIdx].SetValue(card.Driver)
	e.inputs[modelInputIdx].SetValue(card.Model)
	e.workspaceID = card.WorkspaceID
	e.inputs[faceStyleInputIdx].SetValue(card.FaceStyle)
	e.puppet = card.Puppet
	if e.puppet == "" {
		e.puppet = defaultPuppetStyle
	}
	e.useTmpWorkspace = card.UseTmpWorkspace
	e.dashboardControl = card.DashboardControl
	e.subWorkspaceIDs = slices.Clone(card.SubWorkspaceIDs)
	e.skillList.SetSelectedOnly(false)
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
	e.loaded = config.IDCard{}
	e.workspaceID = ""
	e.subWorkspaceIDs = nil
	e.useTmpWorkspace = false
	e.dashboardControl = false
	e.skillList.SetSelectedOnly(false)
	e.puppet = "kankuro"
	e.agentsMD = AgentsMDState{}
	e.skillList.SetSelected(nil)
	e.personaList.SetSelected(nil)
	e.mcpList.SetSelected(nil)
	e.focusIndex = 0
	e.setFocus(0)
}

// SetData sets the available options. Per messages.DataLoadedMsg's contract,
// a nil field means "that registry could not be read this round" - not "the
// caller has no data" - so each field is only replaced when non-nil,
// preserving whatever the editor already has (and, critically, whatever is
// currently selected in the capability lists) across a transient read
// failure. See review_dp_4.md finding 1: this used to assign unconditionally,
// which meant a corrupt skills.yaml wiped the editor's skill selection on the
// next refresh and a save would then persist the card with none.
func (e *CardEditor) SetData(msg messages.DataLoadedMsg) {
	if msg.Workspaces != nil {
		e.workspaces = msg.Workspaces
		e.workspacePicker.SetWorkspaces(e.workspaces)
	}

	if msg.Skills != nil {
		e.skills = msg.Skills
		skillItems := make([]components.CapabilityItem, len(e.skills))
		for i, s := range e.skills {
			skillItems[i] = SkillItem{Skill: s}
		}
		e.skillList.SetItems(skillItems)
	}

	if msg.Personas != nil {
		e.personas = msg.Personas
		personaItems := make([]components.CapabilityItem, len(e.personas))
		for i, p := range e.personas {
			personaItems[i] = personaItem{persona: p}
		}
		e.personaList.SetItems(personaItems)
	}

	if msg.MCPs != nil {
		e.mcps = msg.MCPs
		mcpItems := make([]components.CapabilityItem, len(e.mcps))
		for i, m := range e.mcps {
			mcpItems[i] = mcpItem{mcp: m}
		}
		e.mcpList.SetItems(mcpItems)
	}

	if msg.Puppets != nil {
		e.puppets = msg.Puppets
		e.puppetSelector.SetPuppets(e.puppets)
	}
}

// Card returns the current card data, overlaid onto the card most recently
// loaded via SetCard (or the zero value for a new card). Every field the
// form does not expose — SystemInstruction, Env, CreatedAt, ID, and anything
// added later — is carried through from e.loaded unchanged instead of being
// silently dropped. Only the fields the form actually edits are assigned
// below.
func (e *CardEditor) Card() config.IDCard {
	card := e.loaded
	card.Name = strings.TrimSpace(e.inputs[nameInputIdx].Value())
	card.Driver = strings.TrimSpace(e.inputs[driverInputIdx].Value())
	card.Model = strings.TrimSpace(e.inputs[modelInputIdx].Value())
	card.WorkspaceID = e.workspaceID
	card.FaceStyle = strings.TrimSpace(e.inputs[faceStyleInputIdx].Value())
	card.Puppet = e.puppet
	card.UseTmpWorkspace = e.useTmpWorkspace
	card.DashboardControl = e.dashboardControl
	card.SubWorkspaceIDs = slices.Clone(e.subWorkspaceIDs)
	card.AgentsMD = e.agentsMD.Content
	card.AgentsMDPath = e.agentsMD.Path
	card.AgentsMDPreferPath = e.agentsMD.PreferPath
	card.SkillIDs = e.skillList.SelectedIDs()
	// A hidden persona/MCP section is not rendered, so its list can't be
	// read; leave e.loaded's value (set by SetCard, and kept current by
	// SetSettings when a section transitions to hidden) untouched.
	if e.settings.ShowPersonasInCardEditor {
		card.PersonaIDs = e.personaList.SelectedIDs()
	}
	if e.settings.ShowMCPsInCardEditor {
		card.MCPIDs = e.mcpList.SelectedIDs()
	}
	return card
}

// Init initializes the editor.
func (e *CardEditor) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (e *CardEditor) Update(msg tea.Msg) (component.Model, tea.Cmd) {
	// Handle data refresh messages
	if msg, ok := msg.(messages.MCPDeletedMsg); ok {
		// Remove the deleted MCP from selection
		e.mcpList.Deselect(msg.MCPID)
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
		e.subWorkspaceIDs = nil
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
		return e.callbacks.OnSave(card)
	}

	e.loaded.ID = ""
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
	if e.loaded.ID != "" {
		title = "Edit ID Card"
	}

	return []formBlock{
		blockOf(FocusName, styles.TitleStyle.Render(title)+"\n\n"+"Name: "+e.inputs[nameInputIdx].View()+"\n"),
		blockOf(FocusDriver, "Driver: "+e.inputs[driverInputIdx].View()+"\n"),
		blockOf(FocusModel, "Model: "+e.inputs[modelInputIdx].View()+"\n"),
		blockOf(FocusWorkspace, e.renderWorkspaceLine()+"\n"),
		blockOf(FocusSubWorkspaces, e.renderSubWorkspacesLine()+"\n"),
		blockOf(FocusFaceStyle, "Face Style: "+e.inputs[faceStyleInputIdx].View()+"\n"),
		blockOf(FocusPuppet, e.renderPuppetLine()+"\n"),
		blockOf(FocusUseTmp, e.renderTmpWorkspaceLine()+"\n\n"),
		blockOf(FocusDashboardControl, e.renderDashboardControlLine()+"\n"),
		blockOf(FocusAgentsMD, "AGENTS.md:\n"+e.renderAgentsMDRow()+"\n\n"),
	}
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
	if e.settings.ShowPersonasInCardEditor {
		blocks = append(blocks, listBlock(FocusPersonas, e.personaList))
	}
	if e.settings.ShowMCPsInCardEditor {
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
	const help = "Enter=save Tab=next Esc=cancel"

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
	mark := "[ ]"
	if e.useTmpWorkspace {
		mark = "[x]"
	}
	line := "Use Tmp Workspace: " + mark
	if e.focusIndex == FocusUseTmp {
		line = focusMarker + line
		return lipgloss.NewStyle().Foreground(styles.HighlightColor).Bold(true).Render(line)
	}
	return line
}

func (e *CardEditor) renderDashboardControlLine() string {
	mark := "[ ]"
	if e.dashboardControl {
		mark = "[x]"
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
	if e.settings.ShowPersonasInCardEditor {
		part++
	}
	if e.settings.ShowMCPsInCardEditor {
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
