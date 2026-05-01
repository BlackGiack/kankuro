# Component Architecture Refactor Plan

## Goal

Decompose monolithic `App` struct into composable Bubble Tea components with callback-based data access, consolidating 17 view modes into 6 composite views.

---

## Architecture Overview

### Component Hierarchy

```
App (Coordinator)
├── Dashboard (CompositeView)
│   ├── cardList (CardList)
│   ├── sessionList (SessionList)
│   ├── preview (SessionPreview)
│   └── modals
│       ├── deleteConfirm (Modal)
│       └── switchAgent (CardPicker)
│
├── CardEditor (CompositeView)
│   ├── formInputs (FormInputs)
│   ├── capabilityLists (CapabilityLists)
│   └── modals
│       ├── puppetSelector (PuppetSelector)
│       ├── workspacePicker (WorkspacePicker)
│       └── agentsMDEditor (AgentsMDEditor)
│
├── SourceManager (CompositeView)
│   └── sourceList (SourceList) [reusable for skills/personas/workspaces]
│
├── MCPForm (SimpleView)
├── IDEConfig (CompositeView)
└── SessionLauncher (SimpleView)
```

### View Mode Consolidation

| Current (17 modes) | New (6 views) |
|-------------------|---------------|
| viewDashboard | Dashboard.base |
| viewDeleteConfirm | Dashboard.modal.deleteConfirm |
| viewSwitchAgent | Dashboard.modal.switchAgent |
| viewCardForm | CardEditor.base |
| viewPuppetSelector | CardEditor.modal.puppetSelector |
| viewWorkspacePicker | CardEditor.modal.workspacePicker |
| viewAgentsEditor | CardEditor.modal.agentsMDEditor |
| viewAgentsPathInput | CardEditor.modal.agentsMDPath |
| viewSkillSources | SourceManager.skills |
| viewSkillSourceForm | SourceManager.skills.form |
| viewPersonaSources | SourceManager.personas |
| viewPersonaSourceForm | SourceManager.personas.form |
| viewWorkspaces | SourceManager.workspaces |
| viewWorkspaceSourceForm | SourceManager.workspaces.form |
| viewNaming | SessionLauncher |
| viewMCPForm | MCPForm |
| viewIDE + viewIDEForm | IDEConfig |

### Message Flow Pattern

```
User Input → App.Update()
    │
    ├─→ Current View.Update() → View-specific handling
    │       │
    │       └─→ Component.Update() → Internal state
    │               │
    │               └─→ Return tea.Cmd (callback)
    │
    └─→ App handles commands → Mutate repo/session
            │
            └─→ Emit data refresh messages → Components update
```

---

## Phase 1: Foundation

**Goal**: Create component infrastructure without breaking existing code.

### Task 1.1: Define Component Interface

**File**: `internal/ui/component/component.go`

```go
package component

type Model interface {
    Init() tea.Cmd
    Update(tea.Msg) (Model, tea.Cmd)
    View() string
    SetSize(width, height int)
}

type Base struct {
    Width  int
    Height int
    Focused bool
}
```

**Verify**: `go build ./internal/ui/component/`

---

### Task 1.2: Create Messages Package

**File**: `internal/ui/messages/messages.go`

```go
package messages

// Data messages (from App to components)
type CardsLoadedMsg []config.IDCard
type SessionsLoadedMsg []session.Snapshot
type SkillsLoadedMsg []config.Skill
type PersonasLoadedMsg []config.Persona
type MCPsLoadedMsg []config.MCPServer
type WorkspacesLoadedMsg []config.Workspace

// Action messages (from components to App)
type LaunchSessionMsg struct {
    CardID string
    Name   string
}
type AttachSessionMsg struct{ SessionID string }
type KillSessionMsg struct{ SessionID string }
type DeleteCardMsg struct{ CardID string }
type SaveCardMsg struct{ Card config.IDCard }
type DeleteSessionMsg struct{ SessionID string }

// Navigation messages
type NavigateMsg struct{ View ViewType }
type ShowModalMsg struct{ Modal ModalType }
type CloseModalMsg struct{}
```

**Verify**: `go build ./internal/ui/messages/`

---

### Task 1.3: Create Commands Registry

**File**: `internal/ui/commands.go`

```go
package ui

type Commands struct {
    LaunchSession func(card config.IDCard, skills []config.Skill, personas []config.Persona, mcps []config.MCPServer) (string, error)
    AttachSession  func(sessionID string) tea.Cmd
    KillSession    func(sessionID string) error
    RemoveSession  func(sessionID string) error
    
    SaveCard       func(card config.IDCard) (config.IDCard, error)
    DeleteCard     func(cardID string) error
    
    SaveMCP        func(mcp config.MCPServer) (config.MCPServer, error)
    DeleteMCP      func(mcpID string) error
    
    SaveIDEConfig  func(config.IDEConfig) error
    
    RefreshData    func()
}
```

**Verify**: `go build ./internal/ui/`

---

### Task 1.4: Create Modal Component

**File**: `internal/ui/components/modal.go`

```go
package components

type Modal struct {
    component.Model
    title   string
    content Model
    open    bool
}

func NewModal(title string, content Model) *Modal
func (m *Modal) Open()
func (m *Modal) Close()
func (m *Modal) IsOpen() bool
```

**Verify**: Unit test `TestModal_OpenClose`

---

## Phase 2: Extract Independent Components

**Goal**: Extract components that don't depend on App state directly.

### Task 2.1: Create CardList Component

**File**: `internal/ui/components/card_list.go`

**Extract from**: `app_lists.go` (cardDelegate), `app_render.go` (renderCardItem)

```go
package components

type CardList struct {
    component.Base
    list    list.Model
    cards   []config.IDCard
    skills  []config.Skill
    personas []config.Persona
    
    onSelect func(config.IDCard)
    onEdit   func(config.IDCard)
    onDelete func(config.IDCard)
}

func NewCardList(onSelect, onEdit, onDelete func(config.IDCard)) *CardList
func (c *CardList) SetCards(cards []config.IDCard)
func (c *CardList) Selected() (config.IDCard, bool)
```

**Tests**: `TestCardList_Selection`, `TestCardList_SetCards`

**Parallel with**: Task 2.2, 2.3

---

### Task 2.2: Create SessionList Component

**File**: `internal/ui/components/session_list.go`

**Extract from**: `app_lists.go` (sessionsDelegate, sessionListItem), `app_render.go`

```go
package components

type SessionList struct {
    component.Base
    list       list.Model
    sessions   []session.Snapshot
    cards      []config.IDCard
    
    onSelect   func(session.Snapshot)
    onAttach   func(string)
    onKill     func(string)
    onDelete   func(string)
    onSwitch   func(string)
}

func NewSessionList(callbacks SessionCallbacks) *SessionList
func (s *SessionList) SetSessions(sessions []session.Snapshot, cards []config.IDCard)
func (s *SessionList) Selected() (session.Snapshot, bool)
```

**Tests**: `TestSessionList_Callbacks`

**Parallel with**: Task 2.1, 2.3

---

### Task 2.3: Create SessionPreview Component

**File**: `internal/ui/components/session_preview.go`

**Extract from**: `view_preview.go`

```go
package components

type SessionPreview struct {
    component.Base
    sessionID   string
    snapshot    session.Snapshot
    card        *config.IDCard
    scrollback  []byte
}

func NewSessionPreview() *SessionPreview
func (p *SessionPreview) SetSession(snap session.Snapshot, card *config.IDCard)
func (p *SessionPreview) SetScrollback(data []byte)
func (p *SessionPreview) Clear()
```

**Tests**: `TestSessionPreview_Render`

**Parallel with**: Task 2.1, 2.2

---

### Task 2.4: Create SourceList Component (Generic)

**File**: `internal/ui/components/source_list.go`

**Extract from**: `app_skill_sources.go`, `app_persona_sources.go`, `app_workspaces.go`

```go
package components

type SourceList struct {
    component.Base
    sources    []SourceItem
    cursor     int
    input      textinput.Model
    editingIdx int
    
    onSave   func(source SourceItem)
    onDelete func(id string)
}

type SourceItem interface {
    Display() string
    ID() string
}

func NewSourceList() *SourceList
func (s *SourceList) SetSources(sources []SourceItem)
```

**Tests**: `TestSourceList_CRUD`

---

## Phase 3: Create Composite Views

**Goal**: Build views that compose components together.

### Task 3.1: Create Dashboard Component

**File**: `internal/ui/views/dashboard.go`

**Composes**: CardList, SessionList, SessionPreview, Modals

```go
package views

type Dashboard struct {
    component.Base
    
    cardList    *components.CardList
    sessionList *components.SessionList
    preview     *components.SessionPreview
    
    deleteModal *components.Modal
    switchModal *components.Modal
    
    focusedPane dashboardPane
    searchQuery string
    
    cmds        *ui.Commands
}

func NewDashboard(cmds *ui.Commands) *Dashboard
func (d *Dashboard) SetData(cards []config.IDCard, sessions []session.Snapshot, ...)
func (d *Dashboard) View() string
```

**Tests**: `TestDashboard_FocusSwitch`, `TestDashboard_ModalOverlay`

---

### Task 3.2: Create CardEditor Component

**File**: `internal/ui/views/card_editor.go`

**Composes**: FormInputs, CapabilityLists, Modals

```go
package views

type CardEditor struct {
    component.Base
    
    inputs         []textinput.Model
    capabilities   *components.CapabilityLists
    
    puppetModal    *components.Modal
    workspaceModal *components.Modal
    agentsMDModal  *components.Modal
    
    editingID      string
    focusIndex     int
    
    cmds           *ui.Commands
}

func NewCardEditor(cmds *ui.Commands) *CardEditor
func (e *CardEditor) SetCard(card config.IDCard)
func (e *CardEditor) Reset()
func (e *CardEditor) Card() config.IDCard
```

**Tests**: `TestCardEditor_Validation`, `TestCardEditor_FocusCycle`

---

### Task 3.3: Create SourceManager Component

**File**: `internal/ui/views/source_manager.go`

```go
package views

type SourceManager struct {
    component.Base
    sourceList    *components.SourceList
    
    sourceType    string
    prevView      ViewType
    
    cmds          *ui.Commands
}

func NewSourceManager(sourceType string, cmds *ui.Commands) *SourceManager
```

**Tests**: `TestSourceManager_CRUD`

---

### Task 3.4: Create IDEConfig Component

**File**: `internal/ui/views/ide_config.go`

```go
package views

type IDEConfig struct {
    component.Base
    input       textinput.Model
    editing     bool
    config      config.IDEConfig
    
    cmds        *ui.Commands
}
```

**Tests**: `TestIDEConfig_Save`

---

## Phase 4: Refactor App Coordinator

**Goal**: Replace monolithic App with coordinator pattern.

### Task 4.1: Create View Coordinator

**File**: `internal/ui/coordinator.go`

```go
package ui

type Coordinator struct {
    currentView ViewType
    
    views map[ViewType]component.Model
    
    cards     []config.IDCard
    sessions  []session.Snapshot
    skills    []config.Skill
    personas  []config.Persona
    mcps      []config.MCPServer
    
    repo      config.Repository
    session   *session.Manager
    cmds      Commands
}
```

**Depends on**: Phase 3 complete

---

### Task 4.2: Wire Commands to Repository

**File**: `internal/ui/wiring.go`

```go
func WireCommands(repo config.Repository, sm *session.Manager) Commands {
    return Commands{
        LaunchSession: func(card config.IDCard, ...) (string, error) {
            return sm.Launch(card, ...)
        },
        // ... other commands
    }
}
```

**Verify**: Integration test with real repo

---

### Task 4.3: Create App Wrapper (Backwards Compatible)

**File**: `internal/ui/app.go` (refactor)

Keep same public API but delegate to coordinator.

```go
type App struct {
    coordinator *Coordinator
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    return a.coordinator.Update(msg)
}
```

**Verify**: All existing tests pass

---

## Phase 5: Testing & Verification

### Task 5.1: Unit Tests for Each Component

Create `internal/ui/components/*_test.go` for each component.

**Coverage target**: 80% per component

---

### Task 5.2: Integration Tests

Create `internal/ui/integration_test.go`:

- Full session launch flow
- Card CRUD flow
- Source management flow

---

### Task 5.3: Visual Regression Check

Run app manually through all flows:
- Dashboard navigation
- Card creation/editing
- Session management
- Source management
- MCP form
- IDE config

---

## Task Dependencies

```
Phase 1 (Foundation)
├── 1.1 Component Interface ─┬──> Phase 2
├── 1.2 Messages             ─┤
├── 1.3 Commands Registry     ─┤
└── 1.4 Modal Component       ─┘

Phase 2 (Independent Components - PARALLEL)
├── 2.1 CardList      ───┬──> Phase 3
├── 2.2 SessionList   ───┤
├── 2.3 SessionPreview ───┤
└── 2.4 SourceList    ───┘

Phase 3 (Composite Views)
├── 3.1 Dashboard    ───┬──> Phase 4
├── 3.2 CardEditor   ───┤
├── 3.3 SourceManager ──┤
└── 3.4 IDEConfig    ───┘

Phase 4 (App Refactor - SEQUENTIAL)
├── 4.1 Coordinator  ───> 
├── 4.2 Wiring      ───> 
└── 4.3 App Wrapper ───> Phase 5

Phase 5 (Testing - SEQUENTIAL)
├── 5.1 Unit Tests   ───>
├── 5.2 Integration   ───>
└── 5.3 Manual Check  ───> DONE
```

---

## Parallelization Opportunities

**Phase 2**: All 4 tasks can run in parallel

**Phase 3**: Tasks 3.1, 3.3, 3.4 can run in parallel after Phase 2 complete

---

## Estimated Time

| Phase | Tasks | Time |
|-------|-------|------|
| Phase 1 | 4 tasks | 2-3 hours |
| Phase 2 | 4 tasks (parallel) | 3-4 hours |
| Phase 3 | 4 tasks | 4-5 hours |
| Phase 4 | 3 tasks | 3-4 hours |
| Phase 5 | 3 tasks | 2-3 hours |
| **Total** | | **14-19 hours** |

---

## Done When

- [ ] All 4 independent components extracted with unit tests
- [ ] Dashboard composite view renders correctly
- [ ] CardEditor composite view with all modals works
- [ ] SourceManager supports skills/personas/workspaces
- [ ] App coordinator delegates to views
- [ ] All existing tests pass
- [ ] New component unit tests pass (>80% coverage)
- [ ] Integration tests cover main flows
- [ ] Manual testing confirms no regressions