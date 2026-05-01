# Implementation Plan: Phase 3 Task 3.2 - CardEditor Component

## Goal

Create CardEditor composite view that composes form inputs, capability lists, and modals for editing ID Cards, consolidating 6 view modes (viewCardForm, viewPuppetSelector, viewWorkspacePicker, viewAgentsEditor, viewAgentsPathInput, and parts of viewMCPForm) into a single component.

---

## Analysis: What to Extract

### View Mode Consolidation

| Current View Mode | New Location |
|-------------------|--------------|
| `viewCardForm` | `CardEditor.base` |
| `viewPuppetSelector` | `CardEditor.modal.puppetSelector` |
| `viewWorkspacePicker` | `CardEditor.modal.workspacePicker` |
| `viewAgentsEditor` | `CardEditor.modal.agentsMDEditor` |
| `viewAgentsPathInput` | `CardEditor.modal.agentsMDPath` |
| `viewMCPForm` (partial) | Shared MCP form (called from CardEditor and Dashboard) |

### From Existing Code

| File | Lines | Content | Destination |
|------|-------|---------|-------------|
| `app_card_form.go` | 14-32 | `filteredSkills/Personas/MCPs()` | `views/card_editor.go` |
| `app_card_form.go` | 34-51 | `handleCardFormEscape()` | `views/card_editor.go` |
| `app_card_form.go` | 53-111 | Search input handlers | `views/card_editor.go` |
| `app_card_form.go` | 113-128 | `updateCardForm()` | `views/card_editor.go` |
| `app_card_form.go` | 130-163 | `handleCardFormKeyMsg()` | `views/card_editor.go` |
| `app_card_form.go` | 165-187 | `handleCardFormSpace()` | `views/card_editor.go` |
| `app_card_form.go` | 189-205 | `handleCardFormCtrlE()` | `views/card_editor.go` |
| `app_card_form.go` | 242-269 | Navigation handlers | `views/card_editor.go` |
| `app_card_form.go` | 271-291 | Toggle handlers | `views/card_editor.go` |
| `app_card_form.go` | 293-324 | `handleCardFormSubmit()` | `views/card_editor.go` |
| `app_card_form.go` | 326-353 | `resetCardForm()` | `views/card_editor.go` |
| `app_card_form.go` | 355-417 | Selection helpers | `views/card_editor.go` |
| `app_card_form.go` | 436-467 | Focus management | `views/card_editor.go` |
| `app_card_form.go` | 469-553 | Toggle methods | `views/card_editor.go` |
| `app_card_form.go` | 483-513 | `updatePuppetSelector()` | `views/puppet_selector.go` |
| `view_card_form.go` | 5-43 | `viewCardFormRender()` | `views/card_editor.go` |
| `app_render.go` | 14-55 | `renderAgentsMDRow()` | `views/card_editor.go` |
| `app_render.go` | 57-88 | `renderSkillList/PersonaList/MCPList()` | `views/capability_list.go` |
| `app_render.go` | 90-141 | `renderRegistryList()` (generic) | `views/capability_list.go` |
| `app_state.go` | 11-41 | `CardFormState` | `views/card_editor.go` (internal) |

### CardEditor-Specific Constants

```go
// Focus indices
cardWorkspaceIndex   = 3
cardFaceStyleIndex   = 4
cardPuppetIndex      = 5
cardUseTmpIndex      = 6
cardAgentsMDIndex    = 7
cardSkillListIndex   = 8
cardPersonaListIndex = 9
cardMCPListIndex     = 10

// Layout
cardInputCount  = 5
inputWidthDefault = 60
```

---

## Architecture Overview

### Component Hierarchy

```
CardEditor (CompositeView)
├── inputs         []textinput.Model     // Name, Driver, Model, FaceStyle, UseTmp
├── workspacePicker *WorkspacePickerModal (NEW)
├── puppetSelector *PuppetSelectorModal  (NEW)
├── agentsMDEditor *AgentsMDEditorModal  (NEW)
├── capabilities
│   ├── skillList   *CapabilityList
│   ├── personaList *CapabilityList
│   └── mcpList     *CapabilityList
└── state
    ├── editingID      string
    ├── focusIndex     int
    ├── selectedIDs    map[string]bool
    └── searchQueries  map[string]string
```

### Focus Cycle

```
Index 0: Name (textinput)
Index 1: Driver (textinput)
Index 2: Model (textinput)
Index 3: Workspace (picker - Space to open)
Index 4: Face Style (textinput)
Index 5: Puppet (selector - Space to open)
Index 6: Use Tmp Workspace (textinput)
Index 7: Agents.md (options - Space to open)
Index 8: Skills (list - Space to toggle)
Index 9: Personas (list - Space to toggle)
Index 10: MCPs (list - Space to toggle)
```

### Message Flow

```
User Input → CardEditor.Update()
    │
    ├─→ If modal open → Modal.Update()
    │       └─→ Return ModalResult (Select/Cancel)
    │
    ├─→ Tab/Shift+Tab → Cycle focus
    │
    ├─→ Space on picker/selector → Open modal
    │
    ├─→ Space on capability list → Toggle selection
    │
    ├─→ Up/Down on capability list → Move cursor
    │
    ├─→ ctrl+e on capability list → Open source editor
    │
    ├─→ Enter → Submit form
    │
    └─→ Esc → Cancel (or clear search)
```

---

## Step 3.2.1: Create CapabilityList Component

**File**: `internal/ui/views/capability_list.go` (NEW)

This is a reusable component for skills, personas, and MCPs.

```go
package views

import (
    "fmt"
    "strings"

    bubbletea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    
    "kankuro/internal/ui/component"
)

// CapabilityItem represents an item in a capability list.
type CapabilityItem interface {
    GetID() string
    GetLabel() string
    GetDescription() string
}

// CapabilityListConfig configures a capability list.
type CapabilityListConfig struct {
    Title        string
    SourceHint   string // e.g., "ctrl+e=sources"
    MaxVisible   int
}

// CapabilityList is a selectable list of capabilities.
type CapabilityList struct {
    component.Base
    
    title      string
    sourceHint string
    maxVisible int
    
    items      []CapabilityItem
    cursor     int
    selected   map[string]bool
    searchQuery string
    focused    bool
}

// NewCapabilityList creates a new capability list.
func NewCapabilityList(config CapabilityListConfig) *CapabilityList {
    return &CapabilityList{
        title:      config.Title,
        sourceHint: config.SourceHint,
        maxVisible: config.MaxVisible,
        selected:   make(map[string]bool),
    }
}

// SetItems updates the list items.
func (c *CapabilityList) SetItems(items []CapabilityItem) {
    c.items = items
}

// SetSelected sets the selected IDs.
func (c *CapabilityList) SetSelected(ids []string) {
    c.selected = make(map[string]bool)
    for _, id := range ids {
        c.selected[id] = true
    }
}

// SelectedIDs returns the selected IDs.
func (c *CapabilityList) SelectedIDs() []string {
    ids := make([]string, 0, len(c.selected))
    for _, item := range c.items {
        if c.selected[item.GetID()] {
            ids = append(ids, item.GetID())
        }
    }
    return ids
}

// SetFocused controls focus state.
func (c *CapabilityList) SetFocused(focused bool) {
    c.focused = focused
}

// SetSearchQuery filters the displayed items.
func (c *CapabilityList) SetSearchQuery(query string) {
    c.searchQuery = query
    c.cursor = 0
}

// Cursor returns the current cursor position.
func (c *CapabilityList) Cursor() int {
    return c.cursor
}

// SetCursor sets the cursor position.
func (c *CapabilityList) SetCursor(pos int) {
    filtered := c.filteredItems()
    if pos < 0 {
        pos = 0
    }
    if pos >= len(filtered) {
        pos = len(filtered) - 1
    }
    c.cursor = pos
}

// ToggleSelected toggles selection at cursor.
func (c *CapabilityList) ToggleSelected() {
    filtered := c.filteredItems()
    if c.cursor < 0 || c.cursor >= len(filtered) {
        return
    }
    
    id := filtered[c.cursor].GetID()
    if c.selected[id] {
        delete(c.selected, id)
    } else {
        c.selected[id] = true
    }
}

// SelectedItem returns the item at cursor.
func (c *CapabilityList) SelectedItem() (CapabilityItem, bool) {
    filtered := c.filteredItems()
    if c.cursor < 0 || c.cursor >= len(filtered) {
        return nil, false
    }
    return filtered[c.cursor], true
}

// Init initializes the component.
func (c *CapabilityList) Init() bubbletea.Cmd {
    return nil
}

// Update handles messages.
func (c *CapabilityList) Update(msg bubbletea.Msg) (component.Model, bubbletea.Cmd) {
    if keyMsg, ok := msg.(bubbletea.KeyMsg); ok {
        filtered := c.filteredItems()
        
        switch keyMsg.String() {
        case "up":
            if c.cursor > 0 {
                c.cursor--
            }
        case "down":
            if c.cursor < len(filtered)-1 {
                c.cursor++
            }
        case " ":
            c.ToggleSelected()
        case "backspace":
            if len(c.searchQuery) > 0 {
                c.searchQuery = c.searchQuery[:len(c.searchQuery)-1]
                c.cursor = 0
            }
        default:
            // Handle search input
            if len(keyMsg.Runes) > 0 {
                ch := keyMsg.Runes[0]
                if ch >= 33 && ch <= 126 {
                    c.searchQuery += string(ch)
                    c.cursor = 0
                }
            }
        }
    }
    
    return c, nil
}

// View renders the capability list.
func (c *CapabilityList) View() string {
    var sb strings.Builder
    
    // Header
    header := lipgloss.NewStyle().Foreground(gray).Render(c.title)
    if c.sourceHint != "" {
        header += lipgloss.NewStyle().Foreground(gray).Render(" (" + c.sourceHint + ")")
    }
    if c.searchQuery != "" {
        header += lipgloss.NewStyle().Foreground(gray).Render(" | Search: " + c.searchQuery)
    }
    sb.WriteString(header + "\n")
    
    filtered := c.filteredItems()
    
    if len(filtered) == 0 {
        text := "(no matching items)"
        if c.focused {
            sb.WriteString(lipgloss.NewStyle().Foreground(highlightColor).Bold(true).Render("> " + text))
        } else {
            sb.WriteString(text)
        }
        return sb.String()
    }
    
    // Calculate visible window
    start, end := c.calculateWindow(len(filtered))
    
    if start > 0 {
        sb.WriteString("  ...\n")
    }
    
    for i := start; i < end; i++ {
        item := filtered[i]
        prefix := "  "
        if c.focused && i == c.cursor {
            prefix = "> "
        }
        
        mark := " "
        if c.selected[item.GetID()] {
            mark = "x"
        }
        
        line := fmt.Sprintf("%s[%s] %s", prefix, mark, item.GetLabel())
        
        if c.focused && i == c.cursor {
            line = lipgloss.NewStyle().Foreground(highlightColor).Render(line)
        }
        
        sb.WriteString(line + "\n")
    }
    
    if end < len(filtered) {
        sb.WriteString("  ...\n")
    }
    
    return sb.String()
}

// filteredItems returns items matching the search query.
func (c *CapabilityList) filteredItems() []CapabilityItem {
    if c.searchQuery == "" {
        return c.items
    }
    
    q := strings.ToLower(c.searchQuery)
    var filtered []CapabilityItem
    for _, item := range c.items {
        if strings.Contains(strings.ToLower(item.GetLabel()), q) ||
           strings.Contains(strings.ToLower(item.GetDescription()), q) {
            filtered = append(filtered, item)
        }
    }
    return filtered
}

// calculateWindow returns the visible window based on cursor position.
func (c *CapabilityList) calculateWindow(total int) (start, end int) {
    if c.maxVisible <= 0 {
        return 0, total
    }
    
    half := c.maxVisible / 2
    start = c.cursor - half
    if start < 0 {
        start = 0
    }
    
    end = start + c.maxVisible
    if end > total {
        end = total
        start = end - c.maxVisible
        if start < 0 {
            start = 0
        }
    }
    
    return start, end
}

// SetSize implements component.Model.
func (c *CapabilityList) SetSize(width, height int) {
    c.Base.SetSize(width, height)
    c.maxVisible = height
}

// Theme colors
var (
    highlightColor = lipgloss.Color("#16A085")
    gray           = lipgloss.Color("240")
)
```

---

## Step 3.2.2: Create CardEditor View File

**File**: `internal/ui/views/card_editor.go` (NEW)

```go
package views

import (
    "strings"

    bubbletea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/bubbles/textinput"
    "github.com/charmbracelet/bubbles/textarea"
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
    editingID    string
    focusIndex   int
    puppet       string
    workspaceID  string
    agentsMD     AgentsMDState
    
    // Data
    workspaces []config.Workspace
    skills     []config.Skill
    personas   []config.Persona
    mcps       []config.MCPServer
    
    callbacks  CardEditorCallbacks
    cmds       *commands.Commands
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
        filtered := e.skillList.filteredItems()
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
    e.workspacePicker = model.(*WorkspacePickerModal)
    
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
    e.puppetSelector = model.(*PuppetSelectorModal)
    
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
    e.agentsMDEditor = model.(*AgentsMDEditorModal)
    
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

// Theme colors
var (
    primaryColor   = lipgloss.Color("#8E44AD")
    highlightColor = lipgloss.Color("#16A085")
    gray           = lipgloss.Color("240")
    titleStyle     = lipgloss.NewStyle().Bold(true).Foreground(primaryColor)
)

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
func (i personaItem) GetDescription() string { return i.persona.Description }

type mcpItem struct {
    mcp config.MCPServer
}

func (i mcpItem) GetID() string          { return i.mcp.ID }
func (i mcpItem) GetLabel() string       { return i.mcp.Name }
func (i mcpItem) GetDescription() string { return i.mcp.Description }
```

---

## Step 3.2.3: Create Modal Components

### WorkspacePickerModal

**File**: `internal/ui/views/workspace_picker_modal.go` (NEW)

```go
package views

import (
    "strings"

    bubbletea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/bubbles/list"
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
            return ws.workspace.ID
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
```

### PuppetSelectorModal

**File**: `internal/ui/views/puppet_selector_modal.go` (NEW)

```go
package views

import (
    bubbletea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    
    "kankuro/internal/ascii"
    "kankuro/internal/ascii/characters"
    "kankuro/internal/ui/component"
)

var puppetStyles = []string{"default", "kankuro", "cat", "strawhat", "darth", "totoro"}

// PuppetSelectorModal allows selecting a puppet style.
type PuppetSelectorModal struct {
    component.Base
    styles    []string
    cursor    int
    prevStyle string
    confirmed bool
    cancelled bool
}

// NewPuppetSelectorModal creates a new puppet selector.
func NewPuppetSelectorModal() *PuppetSelectorModal {
    return &PuppetSelectorModal{
        styles: puppetStyles,
    }
}

func (m *PuppetSelectorModal) SetSelected(style string) {
    m.prevStyle = style
    for i, s := range m.styles {
        if s == style {
            m.cursor = i
            return
        }
    }
    m.cursor = 0
}

func (m *PuppetSelectorModal) Selected() string {
    if m.cursor >= 0 && m.cursor < len(m.styles) {
        return m.styles[m.cursor]
    }
    return "kankuro"
}

func (m *PuppetSelectorModal) Open()    { m.confirmed = false; m.cancelled = false }
func (m *PuppetSelectorModal) Close()   { m.confirmed = false; m.cancelled = false }
func (m *PuppetSelectorModal) IsOpen() bool    { return !m.cancelled }
func (m *PuppetSelectorModal) Confirmed() bool { return m.confirmed }
func (m *PuppetSelectorModal) Cancelled() bool { return m.cancelled }

func (m *PuppetSelectorModal) Init() bubbletea.Cmd { return nil }

func (m *PuppetSelectorModal) Update(msg bubbletea.Msg) (component.Model, bubbletea.Cmd) {
    if keyMsg, ok := msg.(bubbletea.KeyMsg); ok {
        switch keyMsg.String() {
        case "left":
            if m.cursor > 0 {
                m.cursor--
            } else {
                m.cursor = len(m.styles) - 1
            }
        case "right":
            if m.cursor < len(m.styles)-1 {
                m.cursor++
            } else {
                m.cursor = 0
            }
        case "enter":
            m.confirmed = true
        case "esc":
            m.cancelled = true
        }
    }
    return m, nil
}

func (m *PuppetSelectorModal) View() string {
    var puppets []string
    for i, style := range m.styles {
        puppet := characters.NewPuppet(func(c *ascii.Config) { c.Set("style", style) })
        w, h := puppet.Bounds()
        canvas := ascii.NewCanvas(w, h)
        puppet.Render(canvas, 0, 0)
        puppetStr := canvas.String()
        
        if i == m.cursor {
            puppetStr = lipgloss.NewStyle().
                Border(lipgloss.RoundedBorder()).
                BorderForeground(lipgloss.Color("#8E44AD")).
                Render(puppetStr)
        }
        puppets = append(puppets, puppetStr)
    }
    
    row := lipgloss.JoinHorizontal(lipgloss.Top, puppets...)
    
    return lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("#8E44AD")).
        Padding(1, 2).
        Render("Select Puppet:\n\n" + row + "\n\n← → to navigate, Enter to confirm")
}
```

### AgentsMDEditorModal

**File**: `internal/ui/views/agentsmd_editor_modal.go` (NEW)

```go
package views

import (
    bubbletea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/bubbles/textarea"
    "github.com/charmbracelet/bubbles/textinput"
    "github.com/charmbracelet/lipgloss"
    
    "kankuro/internal/ui/component"
)

// AgentsMDEditorModal allows editing agents.md content or path.
type AgentsMDEditorModal struct {
    component.Base
    mode       int // 0 = content, 1 = path
    editor     textarea.Model
    pathInput  textinput.Model
    preferPath bool
    confirmed  bool
    cancelled  bool
}

// NewAgentsMDEditorModal creates a new agents.md editor.
func NewAgentsMDEditorModal() *AgentsMDEditorModal {
    editor := textarea.New()
    editor.SetWidth(60)
    editor.SetHeight(10)
    editor.Placeholder = "Enter AGENTS.md content..."
    
    pathInput := textinput.New()
    pathInput.Placeholder = "Path to AGENTS.md file..."
    pathInput.Width = 50
    
    return &AgentsMDEditorModal{
        editor:    editor,
        pathInput: pathInput,
    }
}

func (m *AgentsMDEditorModal) SetState(state AgentsMDState) {
    m.editor.SetValue(state.Content)
    m.pathInput.SetValue(state.Path)
    m.preferPath = state.PreferPath
    if m.preferPath {
        m.mode = 1
        m.pathInput.Focus()
    } else {
        m.mode = 0
        m.editor.Focus()
    }
}

func (m *AgentsMDEditorModal) State() AgentsMDState {
    return AgentsMDState{
        Content:    m.editor.Value(),
        Path:       m.pathInput.Value(),
        PreferPath: m.preferPath,
    }
}

func (m *AgentsMDEditorModal) Open()    { m.confirmed = false; m.cancelled = false }
func (m *AgentsMDEditorModal) Close()   { m.confirmed = false; m.cancelled = false }
func (m *AgentsMDEditorModal) IsOpen() bool    { return !m.cancelled }
func (m *AgentsMDEditorModal) Confirmed() bool { return m.confirmed }
func (m *AgentsMDEditorModal) Cancelled() bool { return m.cancelled }

func (m *AgentsMDEditorModal) Init() bubbletea.Cmd { return nil }

func (m *AgentsMDEditorModal) Update(msg bubbletea.Msg) (component.Model, bubbletea.Cmd) {
    if keyMsg, ok := msg.(bubbletea.KeyMsg); ok {
        switch keyMsg.String() {
        case "enter":
            if m.mode == 1 { // Path mode - enter confirms
                m.preferPath = true
                m.confirmed = true
                return m, nil
            }
        case "ctrl+enter":
            m.preferPath = m.mode == 1
            m.confirmed = true
            return m, nil
        case "esc":
            m.cancelled = true
            return m, nil
        case "tab":
            m.mode = 1 - m.mode
            if m.mode == 0 {
                m.editor.Focus()
                m.pathInput.Blur()
            } else {
                m.editor.Blur()
                m.pathInput.Focus()
            }
            return m, nil
        }
    }
    
    var cmd bubbletea.Cmd
    if m.mode == 0 {
        m.editor, cmd = m.editor.Update(msg)
    } else {
        m.pathInput, cmd = m.pathInput.Update(msg)
    }
    return m, cmd
}

func (m *AgentsMDEditorModal) View() string {
    boxStyle := lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("#8E44AD")).
        Padding(1, 2).
        Width(70)
    
    var content string
    content += "AGENTS.md Editor\n\n"
    content += "Content:\n"
    if m.mode == 0 {
        content += lipgloss.NewStyle().Foreground(highlightColor).Render("> " + m.editor.View()) + "\n"
    } else {
        content += m.editor.View() + "\n"
    }
    content += "\nPath:\n"
    if m.mode == 1 {
        content += lipgloss.NewStyle().Foreground(highlightColor).Render("> " + m.pathInput.View()) + "\n"
    } else {
        content += m.pathInput.View() + "\n"
    }
    content += "\nTab=switch Ctrl+Enter=confirm Esc=cancel"
    
    return boxStyle.Render(content)
}
```

---

## Step 3.2.4: Create CardEditor Tests

**File**: `internal/ui/views/card_editor_test.go` (NEW)

```go
package views

import (
    "testing"
    
    "kankuro/internal/config"
    "kankuro/internal/ui/commands"
)

func TestCardEditor_New(t *testing.T) {
    callbacks := CardEditorCallbacks{
        OnSave: func(config.IDCard) (config.IDCard, error) {
            return config.IDCard{}, nil
        },
    }
    
    e := NewCardEditor(callbacks, &commands.Commands{})
    
    if e == nil {
        t.Fatal("NewCardEditor returned nil")
    }
    
    if e.focusIndex != 0 {
        t.Error("CardEditor should start with focus on first field")
    }
}

func TestCardEditor_SetCard(t *testing.T) {
    e := NewCardEditor(CardEditorCallbacks{}, &commands.Commands{})
    
    card := config.IDCard{
        ID:          "test-1",
        Name:        "Test Card",
        Driver:      "claude",
        Model:       "claude-3-opus",
        WorkspaceID: "ws-1",
        Puppet:      "cat",
        SkillIDs:    []string{"s1", "s2"},
    }
    
    e.SetCard(card)
    
    if e.editingID != "test-1" {
        t.Errorf("Expected editingID 'test-1', got '%s'", e.editingID)
    }
    if e.inputs[0].Value() != "Test Card" {
        t.Errorf("Expected name 'Test Card', got '%s'", e.inputs[0].Value())
    }
    if e.puppet != "cat" {
        t.Errorf("Expected puppet 'cat', got '%s'", e.puppet)
    }
}

func TestCardEditor_Reset(t *testing.T) {
    e := NewCardEditor(CardEditorCallbacks{}, &commands.Commands{})
    
    e.SetCard(config.IDCard{
        ID:     "test-1",
        Name:   "Test Card",
        Puppet: "cat",
    })
    
    e.Reset()
    
    if e.editingID != "" {
        t.Error("editingID should be empty after reset")
    }
    if e.inputs[0].Value() != "" {
        t.Error("Name input should be empty after reset")
    }
    if e.puppet != "kankuro" {
        t.Errorf("Puppet should reset to default, got '%s'", e.puppet)
    }
}

func TestCardEditor_Card(t *testing.T) {
    e := NewCardEditor(CardEditorCallbacks{}, &commands.Commands{})
    
    e.inputs[0].SetValue("My Card")
    e.inputs[1].SetValue("claude")
    e.inputs[2].SetValue("claude-3-opus")
    e.workspaceID = "ws-1"
    e.puppet = "darth"
    e.skillList.selected["s1"] = true
    
    card := e.Card()
    
    if card.Name != "My Card" {
        t.Errorf("Expected name 'My Card', got '%s'", card.Name)
    }
    if card.Driver != "claude" {
        t.Errorf("Expected driver 'claude', got '%s'", card.Driver)
    }
    if card.Puppet != "darth" {
        t.Errorf("Expected puppet 'darth', got '%s'", card.Puppet)
    }
    if len(card.SkillIDs) != 1 {
        t.Errorf("Expected 1 skill, got %d", len(card.SkillIDs))
    }
}

func TestCardEditor_CycleFocus(t *testing.T) {
    e := NewCardEditor(CardEditorCallbacks{}, &commands.Commands{})
    
    if e.focusIndex != 0 {
        t.Error("Should start on first field")
    }
    
    e.cycleFocus(1)
    if e.focusIndex != 1 {
        t.Errorf("Expected focus 1, got %d", e.focusIndex)
    }
    
    // Cycle to end
    for i := 0; i < FocusCount-1; i++ {
        e.cycleFocus(1)
    }
    if e.focusIndex != 0 {
        t.Errorf("Should wrap to 0, got %d", e.focusIndex)
    }
    
    // Cycle backwards
    e.cycleFocus(-1)
    if e.focusIndex != FocusCount-1 {
        t.Errorf("Should wrap to last, got %d", e.focusIndex)
    }
}

func TestCardEditor_SetData(t *testing.T) {
    e := NewCardEditor(CardEditorCallbacks{}, &commands.Commands{})
    
    workspaces := []config.Workspace{
        {ID: "ws-1", Name: "Workspace 1"},
    }
    skills := []config.Skill{
        {ID: "s1", Name: "Skill 1"},
    }
    personas := []config.Persona{
        {ID: "p1", Name: "Persona 1"},
    }
    mcps := []config.MCPServer{
        {ID: "m1", Name: "MCP 1"},
    }
    
    e.SetData(workspaces, skills, personas, mcps)
    
    if len(e.workspaces) != 1 {
        t.Errorf("Expected 1 workspace, got %d", len(e.workspaces))
    }
    if len(e.skillList.items) != 1 {
        t.Errorf("Expected 1 skill item, got %d", len(e.skillList.items))
    }
}

func TestCardEditor_WorkspaceSummary(t *testing.T) {
    e := NewCardEditor(CardEditorCallbacks{}, &commands.Commands{})
    e.SetData([]config.Workspace{
        {ID: "ws-1", Name: "My Workspace"},
    }, nil, nil, nil)
    
    // No selection
    if e.workspaceSummary() != "(none selected)" {
        t.Errorf("Expected '(none selected)', got '%s'", e.workspaceSummary())
    }
    
    // With selection
    e.workspaceID = "ws-1"
    if e.workspaceSummary() != "My Workspace" {
        t.Errorf("Expected 'My Workspace', got '%s'", e.workspaceSummary())
    }
}

func TestCapabilityList_Toggle(t *testing.T) {
    list := NewCapabilityList(CapabilityListConfig{Title: "Test"})
    list.SetItems([]CapabilityItem{
        skillItem{skill: config.Skill{ID: "s1", Name: "Skill 1"}},
        skillItem{skill: config.Skill{ID: "s2", Name: "Skill 2"}},
    })
    
    // Toggle first item
    list.cursor = 0
    list.ToggleSelected()
    
    if !list.selected["s1"] {
        t.Error("s1 should be selected after toggle")
    }
    
    // Toggle again
    list.ToggleSelected()
    if list.selected["s1"] {
        t.Error("s1 should be deselected after second toggle")
    }
}

func TestCapabilityList_Search(t *testing.T) {
    list := NewCapabilityList(CapabilityListConfig{Title: "Test"})
    list.SetItems([]CapabilityItem{
        skillItem{skill: config.Skill{ID: "s1", Name: "Alpha"}},
        skillItem{skill: config.Skill{ID: "s2", Name: "Beta"}},
        skillItem{skill: config.Skill{ID: "s3", Name: "Gamma"}},
    })
    
    filtered := list.filteredItems()
    if len(filtered) != 3 {
        t.Errorf("Expected 3 items without filter, got %d", len(filtered))
    }
    
    list.SetSearchQuery("be")
    filtered = list.filteredItems()
    if len(filtered) != 1 {
        t.Errorf("Expected 1 item matching 'be', got %d", len(filtered))
    }
    if filtered[0].GetID() != "s2" {
        t.Errorf("Expected 's2', got '%s'", filtered[0].GetID())
    }
}

func TestCardEditor_View(t *testing.T) {
    e := NewCardEditor(CardEditorCallbacks{}, &commands.Commands{})
    e.SetSize(80, 40)
    
    view := e.View()
    
    if view == "" {
        t.Error("View should not be empty")
    }
    if !containsString(view, "New ID Card") {
        t.Error("View should contain title")
    }
    if !containsString(view, "Name:") {
        t.Error("View should contain Name field")
    }
    if !containsString(view, "Skills") {
        t.Error("View should contain Skills section")
    }
}

func TestPuppetSelector_Navigation(t *testing.T) {
    m := NewPuppetSelectorModal()
    m.Open()
    
    if m.cursor != 0 {
        t.Error("Should start at cursor 0")
    }
    
    // Right arrow
    m.Update(bubbletea.KeyMsg{Runes: []rune{}, Type: bubbletea.KeyRight})
    if m.cursor != 1 {
        t.Errorf("Expected cursor 1, got %d", m.cursor)
    }
    
    // Wrap around
    for i := 0; i < len(puppetStyles)-1; i++ {
        m.Update(bubbletea.KeyMsg{Runes: []rune{}, Type: bubbletea.KeyRight})
    }
    if m.cursor != 0 {
        t.Errorf("Should wrap to 0, got %d", m.cursor)
    }
}

func containsString(s, substr string) bool {
    for i := 0; i <= len(s)-len(substr); i++ {
        if s[i:i+len(substr)] == substr {
            return true
        }
    }
    return false
}
```

---

## Step 3.2.5: Integration with App (Backwards Compatible)

**File**: `internal/ui/app_card_form.go` (MODIFY - add integration stub)

```go
// ===== Integration with new CardEditor view =====
// This shows how App will eventually delegate to CardEditor.
// Once fully migrated, the existing updateCardForm and viewCardFormRender
// methods can be removed.

func (a *App) newCardEditorView() *views.CardEditor {
    callbacks := views.CardEditorCallbacks{
        OnSave: func(card config.IDCard) (config.IDCard, error) {
            if card.ID == "" {
                return a.repo.CreateIDCard(card)
            }
            return a.repo.UpdateIDCard(card)
        },
        OnCancel: func() {
            a.mode = viewDashboard
        },
        OnOpenSources: func(sourceType string) {
            switch sourceType {
            case "skills":
                a.openSkillSources()
            case "personas":
                a.openPersonaSources()
            }
        },
        OnOpenMCPForm: func() {
            a.cardForm.MCPFormPrevMode = viewCardForm
            a.mode = viewMCPForm
        },
        OnOpenWorkspaces: func() {
            a.refreshWorkspaces()
            a.mode = viewWorkspaces
        },
    }
    
    return views.NewCardEditor(callbacks, &a.cmds)
}

// For future use when fully migrated:
// cardEditorView *views.CardEditor would be a field in App
// and New() would initialize it
```

---

## Verification Checklist

```bash
# 1. Build all packages
go build ./internal/ui/component/
go build ./internal/ui/messages/
go build ./internal/ui/commands/
go build ./internal/ui/components/
go build ./internal/ui/views/

# 2. Run all tests
go test ./internal/ui/views/ -v

# 3. Existing UI still builds
go build ./internal/ui/

# 4. Run existing tests
go test ./internal/ui/
```

---

## Success Criteria

- [ ] `CardEditor` struct defined in `internal/ui/views/card_editor.go`
- [ ] Composes text inputs and capability lists
- [ ] Focus cycling (Tab/Shift+Tab) works correctly
- [ ] Search/filter in capability lists works
- [ ] Workspace picker modal works
- [ ] Puppet selector modal works
- [ ] Agents.md editor modal works
- [ ] Form validation and submission
- [ ] Cancel/reset behavior
- [ ] `SetCard()` populates for editing
- [ ] `Reset()` clears for new card
- [ ] Unit tests pass (>80% coverage)
- [ ] Integration with App (backwards compatible)
- [ ] Existing App still compiles and tests pass

---

## Dependencies

```
component/component.go (from 1.1) ✓
messages/messages.go (from 1.2) ✓
commands/commands.go (from 1.3) ✓
components/modal.go (from 1.4) ✓

views/capability_list.go ← NEW
views/card_editor.go ← NEW
views/workspace_picker_modal.go ← NEW
views/puppet_selector_modal.go ← NEW
views/agentsmd_editor_modal.go ← NEW
```

---

## Estimated Time

| Step | Time |
|------|------|
| 3.2.1 CapabilityList | 30 min |
| 3.2.2 CardEditor view | 75 min |
| 3.2.3 Modal components | 45 min |
| 3.2.4 Unit tests | 45 min |
| 3.2.5 App integration | 20 min |
| Verification | 15 min |
| **Total** | **~3.5 hours** |

---

## What This Enables

1. **Card editing complete**: Full CRUD for ID cards in one component
2. **Modal pattern complete**: Picker, selector, editor modals established
3. **Focus management**: Complex focus cycling with non-input elements
4. **Capability selection**: Reusable CapabilityList for skills/personas/MCPs
5. **Phase 4 ready**: App coordinator can delegate to CardEditor
