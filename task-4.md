# Task 4: Split App Struct

**Priority:** High  
**Effort:** High  
**Estimated Time:** 4-6 hours

## Objective

Refactor the `App` struct (80+ fields) into focused state objects to improve testability, reduce coupling, and make the codebase more maintainable.

---

## Current State

**File:** `internal/ui/app.go`

The `App` struct contains all state for the entire TUI application mixed together:
- Domain data (cards, skills, personas, mcps, sessions, workspaces)
- Multiple form states (card form, MCP form, workspace form, etc.)
- Navigation state (cursors, selections, focus)
- Session management state
- Source management state

---

## Proposed Changes

### Step 1: Create State Sub-Types

Create new file `internal/ui/app_state.go`:

```go
package ui

import (
    "github.com/charmbracelet/bubbles/textarea"
    "github.com/charmbracelet/bubbles/textinput"
    
    "kankuro/internal/config"
)

// CardFormState holds all state for the ID card form view.
type CardFormState struct {
    Inputs              []textinput.Model
    FocusIndex          int
    Puppet              string
    PuppetSelectorIndex int
    PuppetSelectorPrev  string
    EditingID           string
    
    // AgentsMD
    Editor              textarea.Model
    Content             string
    Path                string
    PreferPath          bool
    PathInput           textinput.Model
    
    // Selections
    SelectedWorkspaceID string
    SelectedSkillIDs    map[string]bool
    SelectedPersonaIDs  map[string]bool
    SelectedMCPIDs      map[string]bool
    
    // Cursors
    SkillCursor         int
    PersonaCursor       int
    MCPCursor           int
    
    // Search
    SkillSearchQuery    string
    PersonaSearchQuery  string
    MCPSearchQuery      string
    MCPFormPrevMode     viewMode
}

func NewCardFormState() CardFormState {
    return CardFormState{
        Inputs:           make([]textinput.Model, cardInputCount),
        SelectedSkillIDs: make(map[string]bool),
        SelectedPersonaIDs: make(map[string]bool),
        SelectedMCPIDs:   make(map[string]bool),
        Puppet:           defaultPuppet,
    }
}

// SessionViewState holds session-related UI state.
type SessionViewState struct {
    ActiveID        string
    PendingDeleteID string
    SwitchID        string
    FocusedPane     dashboardPane
}

// NamingState holds the session naming prompt state.
type NamingState struct {
    SessionID  string
    Card       config.IDCard
    Input      textinput.Model
    LaunchDone bool
    LaunchErr  error
    Submitted  bool
}

func NewNamingState() NamingState {
    ni := textinput.New()
    ni.Placeholder = "Session name (Enter to confirm)"
    ni.Width = inputWidthDefault
    return NamingState{Input: ni}
}

// SourceFormState holds generic source form state.
type SourceFormState struct {
    Sources    []string
    Cursor     int
    Input      textinput.Model
    EditingIdx int
}

func NewSourceFormState() SourceFormState {
    ti := textinput.New()
    ti.Width = inputWidthDefault
    return SourceFormState{
        EditingIdx: -1,
        Input:      ti,
    }
}

// WorkspaceFormState holds workspace source form state.
type WorkspaceFormState struct {
    Sources     []config.WorkspaceSource
    Cursor      int
    Inputs      []textinput.Model
    EditingID   string
    SearchQuery string
}

func NewWorkspaceFormState() WorkspaceFormState {
    inputs := make([]textinput.Model, workspaceInputCount)
    inputs[0] = textinput.New()
    inputs[0].Placeholder = "Path to workspace or root directory..."
    inputs[0].Width = inputWidthDefault
    inputs[1] = textinput.New()
    inputs[1].Placeholder = "Is root directory (y/n)"
    inputs[1].Width = inputWidthDefault
    return WorkspaceFormState{Inputs: inputs}
}
```

### Step 2: Refactor App Struct

Update `internal/ui/app.go`:

```go
type App struct {
    // Dependencies
    repo    config.Repository
    session *session.Manager
    
    // View mode
    mode viewMode
    
    // Domain data (loaded from repository)
    cards      []config.IDCard
    skills     []config.Skill
    personas   []config.Persona
    mcps       []config.MCPServer
    sessions   []session.Snapshot
    workspaces []config.Workspace
    
    // UI lists (Bubble Tea list models)
    cardsList    list.Model
    sessionsList list.Model
    switchList   list.Model
    
    // View-specific state (split from original god object)
    cardForm     CardFormState
    sessionView  SessionViewState
    naming       NamingState
    
    // Source management
    skillSources   SourceFormState
    personaSources SourceFormState
    workspaces_    WorkspaceFormState  // underscore to avoid conflict with domain field
    
    // MCP form (remains separate as it's simple)
    mcpInputs     []textinput.Model
    mcpFormPrev   viewMode
    
    // Layout
    help   help.Model
    width  int
    height int
    err    error
}
```

### Step 3: Update New() Constructor

Update `internal/ui/app_new.go`:

```go
func New(repo config.Repository, sm *session.Manager) (*App, error) {
    cards, err := repo.ListIDCards()
    if err != nil {
        return nil, err
    }
    skills, err := repo.ListSkills()
    if err != nil {
        return nil, err
    }
    mcps, err := repo.ListMCPs()
    if err != nil {
        return nil, err
    }
    personas, err := repo.ListPersonas()
    if err != nil {
        return nil, err
    }

    app := &App{
        repo:            repo,
        session:         sm,
        mode:            viewDashboard,
        cards:           cards,
        skills:          skills,
        personas:        personas,
        mcps:            mcps,
        sessions:        sm.List(),
        cardForm:        NewCardFormState(),
        sessionView:     SessionViewState{FocusedPane: paneCards},
        naming:          NewNamingState(),
        skillSources:    NewSourceFormState(),
        personaSources:  NewSourceFormState(),
        workspaces_:     NewWorkspaceFormState(),
        help:            help.New(),
    }
    
    app.initMCPInputs()

    app.initLists()
    app.refreshWorkspaces()
    app.refreshLists()

    return app, nil
}

func (a *App) initMCPInputs() {
    a.mcpInputs = make([]textinput.Model, mcpInputCount)
    placeholders := []string{
        "mcp name",
        "mcp command (local only)",
        "args (comma separated, local only)",
        "cwd (local only)",
        "url (remote only)",
        "env (KEY=VALUE, local only)",
        "headers (KEY=VALUE, remote only)",
    }
    for i := range a.mcpInputs {
        a.mcpInputs[i] = textinput.New()
        a.mcpInputs[i].Placeholder = placeholders[i]
        a.mcpInputs[i].Width = inputWidthDefault
    }
}
```

### Step 4: Update All Field Access

Search and replace throughout `internal/ui/*.go`:

| Old Access | New Access |
|------------|------------|
| `a.cardInputs` | `a.cardForm.Inputs` |
| `a.cardFocusIndex` | `a.cardForm.FocusIndex` |
| `a.cardPuppet` | `a.cardForm.Puppet` |
| `a.puppetSelectorIndex` | `a.cardForm.PuppetSelectorIndex` |
| `a.puppetSelectorPrev` | `a.cardForm.PuppetSelectorPrev` |
| `a.editingCardID` | `a.cardForm.EditingID` |
| `a.agentsEditor` | `a.cardForm.Editor` |
| `a.agentsMD` | `a.cardForm.Content` |
| `a.agentsMDPath` | `a.cardForm.Path` |
| `a.agentsMDPreferPath` | `a.cardForm.PreferPath` |
| `a.agentsPathInput` | `a.cardForm.PathInput` |
| `a.selectedWorkspaceID` | `a.cardForm.SelectedWorkspaceID` |
| `a.selectedSkillIDs` | `a.cardForm.SelectedSkillIDs` |
| `a.selectedPersonaIDs` | `a.cardForm.SelectedPersonaIDs` |
| `a.selectedMCPIDs` | `a.cardForm.SelectedMCPIDs` |
| `a.skillCursor` | `a.cardForm.SkillCursor` |
| `a.personaCursor` | `a.cardForm.PersonaCursor` |
| `a.mcpCursor` | `a.cardForm.MCPCursor` |
| `a.skillSearchQuery` | `a.cardForm.SkillSearchQuery` |
| `a.personaSearchQuery` | `a.cardForm.PersonaSearchQuery` |
| `a.mcpSearchQuery` | `a.cardForm.MCPSearchQuery` |
| `a.mcpFormPrevMode` | `a.cardForm.MCPFormPrevMode` |
| `a.activeSessionID` | `a.sessionView.ActiveID` |
| `a.pendingDeleteID` | `a.sessionView.PendingDeleteID` |
| `a.switchSessionID` | `a.sessionView.SwitchID` |
| `a.focusedPane` | `a.sessionView.FocusedPane` |
| `a.namingSessionID` | `a.naming.SessionID` |
| `a.namingCard` | `a.naming.Card` |
| `a.namingInput` | `a.naming.Input` |
| `a.namingLaunchDone` | `a.naming.LaunchDone` |
| `a.namingLaunchErr` | `a.naming.LaunchErr` |
| `a.namingSubmitted` | `a.naming.Submitted` |
| `a.skillSources` | `a.skillSources.Sources` |
| `a.skillSourceCursor` | `a.skillSources.Cursor` |
| `a.skillSourceInput` | `a.skillSources.Input` |
| `a.editingSkillSourceIndex` | `a.skillSources.EditingIdx` |
| `a.personaSources` | `a.personaSources.Sources` |
| `a.personaSourceCursor` | `a.personaSources.Cursor` |
| `a.personaSourceInput` | `a.personaSources.Input` |
| `a.editingPersonaSourceIndex` | `a.personaSources.EditingIdx` |
| `a.workspaceSources` | `a.workspaces_.Sources` |
| `a.workspaceSourceCursor` | `a.workspaces_.Cursor` |
| `a.workspaceInputs` | `a.workspaces_.Inputs` |
| `a.editingWorkspaceSourceID` | `a.workspaces_.EditingID` |
| `a.workspaceSearchQuery` | `a.workspaces_.SearchQuery` |

### Step 5: Update initInputs()

Replace the monolithic `initInputs()` with focused initialization:

```go
func (a *App) initInputs() {
    // Card form inputs
    placeholders := []string{
        "name",
        "driver command (e.g. claude)",
        "model",
        "face_style (neutral/happy/focused/thinking)",
        "use tmp workspace (y/n)",
    }
    for i := range a.cardForm.Inputs {
        a.cardForm.Inputs[i] = textinput.New()
        a.cardForm.Inputs[i].Placeholder = placeholders[i]
        a.cardForm.Inputs[i].Width = inputWidthDefault
    }
    a.cardForm.Inputs[0].Focus()
    
    // Initialize puppet
    for i, p := range puppetStyles() {
        if p == a.cardForm.Puppet {
            a.cardForm.PuppetSelectorIndex = i
            break
        }
    }
    
    // AgentsMD editor
    a.cardForm.Editor = textarea.New()
    a.cardForm.Editor.Placeholder = "Enter AGENTS.md content here..."
    a.cardForm.Editor.SetWidth(editorWidth)
    a.cardForm.Editor.SetHeight(editorHeight)
    
    a.cardForm.PathInput = textinput.New()
    a.cardForm.PathInput.Placeholder = "Path to AGENTS.md file..."
    a.cardForm.PathInput.Width = inputWidthDefault
    
    // Source form inputs already initialized in NewSourceFormState()
    // Workspace form inputs already initialized in NewWorkspaceFormState()
}
```

---

## Verification Steps

1. **Build:** `go build ./cmd/kankuro`
2. **Test:** `go test ./internal/ui/...`
3. **Run:** Manually test all UI flows:
   - Create/edit ID card
   - Select skills/personas/MCPs
   - Puppet selector
   - Session management
   - Source management

---

## Files Modified

| File | Changes |
|------|---------|
| `internal/ui/app_state.go` | **NEW** - State type definitions |
| `internal/ui/app.go` | Refactored App struct |
| `internal/ui/app_new.go` | Updated New() and init methods |
| `internal/ui/app_card_form.go` | Field access updates |
| `internal/ui/app_view.go` | Field access updates |
| `internal/ui/app_update.go` | Field access updates |
| `internal/ui/app_dashboard.go` | Field access updates |
| `internal/ui/app_skill_sources.go` | Field access updates |
| `internal/ui/app_persona_sources.go` | Field access updates |
| `internal/ui/app_workspaces.go` | Field access updates |
| `internal/ui/app_naming.go` | Field access updates |
| `internal/ui/app_mcp.go` | Field access updates |
| `internal/ui/app_lists.go` | Field access updates |
| `internal/ui/app_items.go` | Field access updates |
| `internal/ui/app_inputs.go` | Field access updates |
| `internal/ui/app_render.go` | Field access updates |
| `internal/ui/app_switch.go` | Field access updates |
| `internal/ui/app_agents.go` | Field access updates |

---

## Benefits

- Each state object has a clear purpose
- Easier to test individual features
- Reduces mental overhead when reading code
- Makes future refactoring (e.g., splitting into multiple packages) easier
- Removes god object anti-pattern