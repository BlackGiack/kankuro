# Implementation Plan: Phase 4 & 5 - App Coordinator Refactor and Testing

## Goal

Refactor the monolithic `App` struct into a coordinator pattern that delegates to composite views, then verify the entire refactor with comprehensive testing.

---

## Overview

### Phase 4: App Coordinator Refactor
- Replace 17-case view mode switch with view delegation
- Wire callbacks to repository/session operations
- Maintain backwards compatibility with existing public API

### Phase 5: Testing & Verification
- Comprehensive unit tests for all components
- Integration tests for main user flows
- Manual visual regression check

---

## Phase 4: App Coordinator Refactor

### Architecture After Phase 3

```
App (current monolith)
├── repo, session (data layer)
├── 17 viewMode cases in Update()
├── 17 viewMode cases in View()
└── Direct state management (cards, sessions, skills, etc.)

↓ After Phase 4

Coordinator
├── repo, session (data layer)
├── views map[ViewType]component.Model
│   ├── ViewDashboard → Dashboard composite
│   ├── ViewCardEditor → CardEditor composite
│   ├── ViewSourceManager → SourceManager composite
│   ├── ViewIDEConfig → IDEConfig composite
│   ├── ViewMCPForm → MCPForm (simple)
│   └── ViewNaming → SessionLauncher (simple)
├── currentView ViewType
└── commands Commands (wired to repo/session)

App (wrapper - backwards compatible)
└── coordinator *Coordinator (delegates all methods)
```

### View Mode Mapping

| Old viewMode | New ViewType | Composite View |
|--------------|--------------|----------------|
| viewDashboard | ViewDashboard | Dashboard.base |
| viewDeleteConfirm | ViewDashboard | Dashboard.modal.deleteConfirm |
| viewSwitchAgent | ViewDashboard | Dashboard.modal.switchAgent |
| viewCardForm | ViewCardEditor | CardEditor.base |
| viewPuppetSelector | ViewCardEditor | CardEditor.modal.puppetSelector |
| viewWorkspacePicker | ViewCardEditor | CardEditor.modal.workspacePicker |
| viewAgentsEditor | ViewCardEditor | CardEditor.modal.agentsMDEditor |
| viewAgentsPathInput | ViewCardEditor | CardEditor.modal.agentsMDPath |
| viewSkillSources | ViewSourceManager | SourceManager.skills |
| viewSkillSourceForm | ViewSourceManager | SourceManager.skills.form |
| viewPersonaSources | ViewSourceManager | SourceManager.personas |
| viewPersonaSourceForm | ViewSourceManager | SourceManager.personas.form |
| viewWorkspaces | ViewSourceManager | SourceManager.workspaces |
| viewWorkspaceSourceForm | ViewSourceManager | SourceManager.workspaces.form |
| viewIDE | ViewIDEConfig | IDEConfig.base |
| viewIDEForm | ViewIDEConfig | IDEConfig.form |
| viewMCPForm | ViewMCPForm | MCPForm |
| viewNaming | ViewNaming | SessionLauncher |

---

## Task 4.1: Create View Coordinator

### Step 4.1.1: Define ViewType Enum and Coordinator Struct

**File**: `internal/ui/coordinator.go` (NEW)

```go
package ui

import (
    "errors"
    
    tea "github.com/charmbracelet/bubbletea"
    
    "kankuro/internal/config"
    "kankuro/internal/session"
    "kankuro/internal/ui/commands"
    "kankuro/internal/ui/component"
    "kankuro/internal/ui/messages"
    "kankuro/internal/ui/views"
)

type ViewType int

const (
    ViewDashboard ViewType = iota
    ViewCardEditor
    ViewSourceManager
    ViewIDEConfig
    ViewMCPForm
    ViewNaming
)

type Coordinator struct {
    currentView ViewType
    prevView    ViewType
    
    views map[ViewType]component.Model
    
    repo    config.Repository
    session *session.Manager
    cmds    *commands.Commands
    
    width  int
    height int
    err    error
    
    pendingNavigation *ViewType
}

func NewCoordinator(repo config.Repository, sm *session.Manager) *Coordinator {
    cmds := wireCommands(repo, sm)
    
    c := &Coordinator{
        repo:      repo,
        session:   sm,
        cmds:      cmds,
        views:     make(map[ViewType]component.Model),
        currentView: ViewDashboard,
    }
    
    c.initViews()
    return c
}

func (c *Coordinator) initViews() {
    c.views[ViewDashboard] = views.NewDashboard(c.dashboardCallbacks(), c.cmds)
    c.views[ViewCardEditor] = views.NewCardEditor(c.cardEditorCallbacks(), c.cmds)
    c.views[ViewSourceManager] = views.NewSourceManager(c.sourceManagerCallbacks(), c.cmds)
    c.views[ViewIDEConfig] = views.NewIDEConfig(c.ideCallbacks(), c.cmds)
    c.views[ViewMCPForm] = views.NewMCPForm(c.mcpCallbacks(), c.cmds)
    c.views[ViewNaming] = views.NewSessionLauncher(c.namingCallbacks(), c.cmds)
}

func (c *Coordinator) Init() tea.Cmd {
    return tea.Batch(
        c.refreshData(),
        c.tick(),
    )
}

func (c *Coordinator) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch m := msg.(type) {
    case tea.WindowSizeMsg:
        c.width = m.Width
        c.height = m.Height
        c.resizeAll()
        return c, nil
        
    case tickMsg:
        return c, tea.Batch(c.refreshData(), c.tick())
        
    case messages.NavigateMsg:
        return c, c.navigateTo(m.View)
        
    case messages.ErrorMsg:
        c.err = m.Err
        return c, nil
        
    case messages.ClearErrorMsg:
        c.err = nil
        return c, nil
    }
    
    return c.updateCurrentView(msg)
}

func (c *Coordinator) View() string {
    if view, ok := c.views[c.currentView]; ok {
        return view.View()
    }
    return "Unknown view"
}

func (c *Coordinator) updateCurrentView(msg tea.Msg) (tea.Model, tea.Cmd) {
    view, ok := c.views[c.currentView]
    if !ok {
        return c, nil
    }
    
    updated, cmd := view.Update(msg)
    c.views[c.currentView] = updated.(component.Model)
    return c, cmd
}

func (c *Coordinator) navigateTo(vt ViewType) tea.Cmd {
    c.prevView = c.currentView
    c.currentView = vt
    c.err = nil
    return c.views[vt].Init()
}

func (c *Coordinator) resizeAll() {
    for _, view := range c.views {
        view.SetSize(c.width, c.height)
    }
}

func (c *Coordinator) refreshData() tea.Cmd {
    cards, _ := c.repo.ListIDCards()
    sessions := c.session.List()
    skills, _ := c.repo.ListSkills()
    personas, _ := c.repo.ListPersonas()
    mcps, _ := c.repo.ListMCPServers()
    workspaces, _ := c.repo.ListWorkspaces()
    
    c.broadcastData(messages.DataLoadedMsg{
        Cards:      cards,
        Sessions:   sessions,
        Skills:     skills,
        Personas:   personas,
        MCPs:       mcps,
        Workspaces: workspaces,
    })
    
    return nil
}

func (c *Coordinator) broadcastData(msg messages.DataLoadedMsg) {
    for _, view := range c.views {
        if dataReceiver, ok := view.(messages.DataReceiver); ok {
            dataReceiver.SetData(msg)
        }
    }
}
```

---

### Step 4.1.2: Define Callback Interfaces

**File**: `internal/ui/callbacks.go` (NEW)

```go
package ui

import (
    tea "github.com/charmbracelet/bubbletea"
    
    "kankuro/internal/config"
    "kankuro/internal/session"
    "kankuro/internal/ui/views"
)

func (c *Coordinator) dashboardCallbacks() views.DashboardCallbacks {
    return views.DashboardCallbacks{
        OnEditCard: func(card config.IDCard) {
            editor := c.views[ViewCardEditor].(*views.CardEditor)
            editor.SetCard(card)
            c.navigateTo(ViewCardEditor)
        },
        
        OnLaunchCard: func(card config.IDCard) {
            launcher := c.views[ViewNaming].(*views.SessionLauncher)
            launcher.SetCard(card)
            c.navigateTo(ViewNaming)
        },
        
        OnAttach: func(sessionID string) tea.Cmd {
            return c.session.Attach(sessionID)
        },
        
        OnKill: func(sessionID string) error {
            return c.session.Kill(sessionID)
        },
        
        OnRemove: func(sessionID string) error {
            return c.session.Remove(sessionID)
        },
        
        OnSwitch: func(sessionID, newCardID string) error {
            card, err := c.repo.GetIDCard(newCardID)
            if err != nil {
                return err
            }
            skills := c.skillsForCard(card)
            personas := c.personasForCard(card)
            mcps := c.mcpsForCard(card)
            return c.session.Relaunch(sessionID, card, skills, personas, mcps)
        },
        
        OnDeleteCard: func(cardID string) error {
            return c.repo.DeleteIDCard(cardID)
        },
        
        OnOpenMCP: func() {
            c.navigateTo(ViewMCPForm)
        },
        
        OnOpenWorkspaces: func() {
            sm := c.views[ViewSourceManager].(*views.SourceManager)
            sm.SetType(views.SourceTypeWorkspaces)
            c.navigateTo(ViewSourceManager)
        },
        
        OnOpenIDE: func() {
            c.navigateTo(ViewIDEConfig)
        },
        
        OnOpenCardEditor: func() {
            editor := c.views[ViewCardEditor].(*views.CardEditor)
            editor.Reset()
            c.navigateTo(ViewCardEditor)
        },
        
        OnOpenSkillSources: func() {
            sm := c.views[ViewSourceManager].(*views.SourceManager)
            sm.SetType(views.SourceTypeSkills)
            c.navigateTo(ViewSourceManager)
        },
        
        OnOpenPersonaSources: func() {
            sm := c.views[ViewSourceManager].(*views.SourceManager)
            sm.SetType(views.SourceTypePersonas)
            c.navigateTo(ViewSourceManager)
        },
    }
}

func (c *Coordinator) cardEditorCallbacks() views.CardEditorCallbacks {
    return views.CardEditorCallbacks{
        OnSave: func(card config.IDCard) (config.IDCard, error) {
            if card.ID == "" {
                return c.repo.CreateIDCard(card)
            }
            return card, c.repo.UpdateIDCard(card)
        },
        
        OnCancel: func() {
            c.navigateTo(ViewDashboard)
        },
        
        OnOpenWorkspacePicker: func() {
        },
        
        OnOpenPuppetSelector: func() {
        },
        
        OnOpenAgentsMDEditor: func() {
        },
    }
}

func (c *Coordinator) sourceManagerCallbacks() views.SourceManagerCallbacks {
    return views.SourceManagerCallbacks{
        OnSaveSkillSource: func(src config.SkillSource) error {
            sources, _ := c.repo.ListSkillSources()
            sources = append(sources, src)
            return c.repo.SaveSkillSources(sources)
        },
        
        OnDeleteSkillSource: func(id string) error {
            sources, _ := c.repo.ListSkillSources()
            var filtered []config.SkillSource
            for _, s := range sources {
                if s.ID != id {
                    filtered = append(filtered, s)
                }
            }
            return c.repo.SaveSkillSources(filtered)
        },
        
        OnSavePersonaSource: func(src config.PersonaSource) error {
            sources, _ := c.repo.ListPersonaSources()
            sources = append(sources, src)
            return c.repo.SavePersonaSources(sources)
        },
        
        OnDeletePersonaSource: func(id string) error {
            sources, _ := c.repo.ListPersonaSources()
            var filtered []config.PersonaSource
            for _, s := range sources {
                if s.ID != id {
                    filtered = append(filtered, s)
                }
            }
            return c.repo.SavePersonaSources(filtered)
        },
        
        OnSaveWorkspaceSource: func(src config.WorkspaceSource) error {
            sources, _ := c.repo.ListWorkspaceSources()
            sources = append(sources, src)
            return c.repo.SaveWorkspaceSources(sources)
        },
        
        OnDeleteWorkspaceSource: func(id string) error {
            sources, _ := c.repo.ListWorkspaceSources()
            var filtered []config.WorkspaceSource
            for _, s := range sources {
                if s.ID != id {
                    filtered = append(filtered, s)
                }
            }
            return c.repo.SaveWorkspaceSources(filtered)
        },
        
        OnBack: func() {
            c.navigateTo(ViewDashboard)
        },
    }
}

func (c *Coordinator) ideCallbacks() views.IDEConfigCallbacks {
    return views.IDEConfigCallbacks{
        OnSave: func(cfg config.IDEConfig) error {
            return c.repo.SaveIDEConfig(cfg)
        },
        
        OnLaunch: func(sessionID string) error {
            return c.session.OpenIDE(sessionID)
        },
        
        OnBack: func() {
            c.navigateTo(ViewDashboard)
        },
    }
}

func (c *Coordinator) mcpCallbacks() views.MCPFormCallbacks {
    return views.MCPFormCallbacks{
        OnSave: func(mcp config.MCPServer) (config.MCPServer, error) {
            if mcp.ID == "" {
                return c.repo.CreateMCP(mcp)
            }
            return mcp, c.repo.UpdateMCP(mcp)
        },
        
        OnDelete: func(mcpID string) error {
            return c.repo.DeleteMCP(mcpID)
        },
        
        OnBack: func() {
            c.navigateTo(ViewDashboard)
        },
    }
}

func (c *Coordinator) namingCallbacks() views.SessionLauncherCallbacks {
    return views.SessionLauncherCallbacks{
        OnLaunch: func(card config.IDCard, name string) (string, error) {
            skills := c.skillsForCard(card)
            personas := c.personasForCard(card)
            mcps := c.mcpsForCard(card)
            return c.session.Launch(card, skills, personas, mcps, name)
        },
        
        OnCancel: func() {
            c.navigateTo(ViewDashboard)
        },
    }
}

func (c *Coordinator) skillsForCard(card config.IDCard) []config.Skill {
    var result []config.Skill
    allSkills, _ := c.repo.ListSkills()
    for _, skill := range allSkills {
        for _, id := range card.SkillIDs {
            if skill.ID == id {
                result = append(result, skill)
            }
        }
    }
    return result
}

func (c *Coordinator) personasForCard(card config.IDCard) []config.Persona {
    var result []config.Persona
    allPersonas, _ := c.repo.ListPersonas()
    for _, persona := range allPersonas {
        for _, id := range card.PersonaIDs {
            if persona.ID == id {
                result = append(result, persona)
            }
        }
    }
    return result
}

func (c *Coordinator) mcpsForCard(card config.IDCard) []config.MCPServer {
    var result []config.MCPServer
    allMCPs, _ := c.repo.ListMCPServers()
    for _, mcp := range allMCPs {
        for _, id := range card.MCPIDs {
            if mcp.ID == id {
                result = append(result, mcp)
            }
        }
    }
    return result
}
```

---

## Task 4.2: Wire Commands to Repository

### Step 4.2.1: Define Commands Registry

**File**: `internal/ui/commands/commands.go` (MODIFY from Phase 1)

```go
package commands

import (
    "kankuro/internal/config"
    "kankuro/internal/session"
)

type Commands struct {
    LaunchSession func(card config.IDCard, skills []config.Skill, personas []config.Persona, mcps []config.MCPServer, name string) (string, error)
    AttachSession  func(sessionID string) error
    KillSession    func(sessionID string) error
    RemoveSession  func(sessionID string) error
    RelaunchSession func(sessionID string, card config.IDCard, skills []config.Skill, personas []config.Persona, mcps []config.MCPServer) error
    
    SaveCard       func(card config.IDCard) (config.IDCard, error)
    DeleteCard     func(cardID string) error
    GetCard        func(cardID string) (config.IDCard, error)
    
    SaveMCP        func(mcp config.MCPServer) (config.MCPServer, error)
    DeleteMCP      func(mcpID string) error
    
    SaveSkillSource   func(src config.SkillSource) error
    DeleteSkillSource func(id string) error
    
    SavePersonaSource   func(src config.PersonaSource) error
    DeletePersonaSource func(id string) error
    
    SaveWorkspaceSource   func(src config.WorkspaceSource) error
    DeleteWorkspaceSource func(id string) error
    
    SaveIDEConfig  func(config.IDEConfig) error
    OpenIDE        func(sessionID string) error
    
    RefreshData    func()
}
```

### Step 4.2.2: Wire Commands to Repository

**File**: `internal/ui/wiring.go` (NEW)

```go
package ui

import (
    "kankuro/internal/config"
    "kankuro/internal/session"
    "kankuro/internal/ui/commands"
)

func wireCommands(repo config.Repository, sm *session.Manager) *commands.Commands {
    return &commands.Commands{
        LaunchSession: func(card config.IDCard, skills []config.Skill, personas []config.Persona, mcps []config.MCPServer, name string) (string, error) {
            return sm.Launch(card, skills, personas, mcps, name)
        },
        
        AttachSession: func(sessionID string) error {
            return sm.Attach(sessionID)
        },
        
        KillSession: func(sessionID string) error {
            return sm.Kill(sessionID)
        },
        
        RemoveSession: func(sessionID string) error {
            return sm.Remove(sessionID)
        },
        
        RelaunchSession: func(sessionID string, card config.IDCard, skills []config.Skill, personas []config.Persona, mcps []config.MCPServer) error {
            return sm.Relaunch(sessionID, card, skills, personas, mcps)
        },
        
        SaveCard: func(card config.IDCard) (config.IDCard, error) {
            if card.ID == "" {
                return repo.CreateIDCard(card)
            }
            return card, repo.UpdateIDCard(card)
        },
        
        DeleteCard: func(cardID string) error {
            return repo.DeleteIDCard(cardID)
        },
        
        GetCard: func(cardID string) (config.IDCard, error) {
            return repo.GetIDCard(cardID)
        },
        
        SaveMCP: func(mcp config.MCPServer) (config.MCPServer, error) {
            if mcp.ID == "" {
                return repo.CreateMCP(mcp)
            }
            return mcp, repo.UpdateMCP(mcp)
        },
        
        DeleteMCP: func(mcpID string) error {
            return repo.DeleteMCP(mcpID)
        },
        
        SaveSkillSource: func(src config.SkillSource) error {
            sources, _ := repo.ListSkillSources()
            sources = append(sources, src)
            return repo.SaveSkillSources(sources)
        },
        
        DeleteSkillSource: func(id string) error {
            sources, _ := repo.ListSkillSources()
            var filtered []config.SkillSource
            for _, s := range sources {
                if s.ID != id {
                    filtered = append(filtered, s)
                }
            }
            return repo.SaveSkillSources(filtered)
        },
        
        SavePersonaSource: func(src config.PersonaSource) error {
            sources, _ := repo.ListPersonaSources()
            sources = append(sources, src)
            return repo.SavePersonaSources(sources)
        },
        
        DeletePersonaSource: func(id string) error {
            sources, _ := repo.ListPersonaSources()
            var filtered []config.PersonaSource
            for _, s := range sources {
                if s.ID != id {
                    filtered = append(filtered, s)
                }
            }
            return repo.SavePersonaSources(filtered)
        },
        
        SaveWorkspaceSource: func(src config.WorkspaceSource) error {
            sources, _ := repo.ListWorkspaceSources()
            sources = append(sources, src)
            return repo.SaveWorkspaceSources(sources)
        },
        
        DeleteWorkspaceSource: func(id string) error {
            sources, _ := repo.ListWorkspaceSources()
            var filtered []config.WorkspaceSource
            for _, s := range sources {
                if s.ID != id {
                    filtered = append(filtered, s)
                }
            }
            return repo.SaveWorkspaceSources(filtered)
        },
        
        SaveIDEConfig: func(cfg config.IDEConfig) error {
            return repo.SaveIDEConfig(cfg)
        },
        
        OpenIDE: func(sessionID string) error {
            return sm.OpenIDE(sessionID)
        },
        
        RefreshData: func() {},
    }
}
```

---

## Task 4.3: Create App Wrapper (Backwards Compatible)

### Step 4.3.1: Refactor App to Delegate to Coordinator

**File**: `internal/ui/app.go` (MODIFY)

```go
package ui

import (
    tea "github.com/charmbracelet/bubbletea"
    
    "kankuro/internal/config"
    "kankuro/internal/session"
)

type App struct {
    coordinator *Coordinator
}

func New(repo config.Repository, sm *session.Manager) (*App, error) {
    coord := NewCoordinator(repo, sm)
    return &App{coordinator: coord}, nil
}

func (a *App) Init() tea.Cmd {
    return a.coordinator.Init()
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    return a.coordinator.Update(msg)
}

func (a *App) View() string {
    return a.coordinator.View()
}

func (a *App) SetSize(width, height int) {
    a.coordinator.width = width
    a.coordinator.height = height
    a.coordinator.resizeAll()
}
```

### Step 4.3.2: Preserve Backward Compatibility

The `App` struct maintains the same public interface:
- `New(repo, sm) (*App, error)` - unchanged
- `Init() tea.Cmd` - unchanged
- `Update(msg) (tea.Model, tea.Cmd)` - unchanged
- `View() string` - unchanged

All existing code that uses `App` continues to work without modification.

---

## Phase 5: Testing & Verification

### Testing Strategy

```
Phase 5 Tests
├── Unit Tests (per component)
│   ├── components/card_list_test.go
│   ├── components/session_list_test.go
│   ├── components/session_preview_test.go
│   ├── components/source_list_test.go
│   ├── components/modal_test.go
│   ├── views/dashboard_test.go
│   ├── views/card_editor_test.go
│   ├── views/source_manager_test.go
│   ├── views/ide_config_test.go
│   ├── views/mcp_form_test.go
│   └── views/session_launcher_test.go
│
├── Integration Tests
│   ├── integration_test.go
│   │   ├── TestFullSessionLaunchFlow
│   │   ├── TestCardCRUDFlow
│   │   ├── TestSourceManagementFlow
│   │   ├── TestSessionManagementFlow
│   │   └── TestNavigationFlow
│   └── coordinator_test.go
│       ├── TestCoordinatorNavigation
│       ├── TestCoordinatorDataPropagation
│       └── TestCoordinatorErrorHandling
│
└── Manual Verification
    ├── Dashboard navigation
    ├── Card creation/editing
    ├── Session management
    ├── Source management
    ├── MCP configuration
    └── IDE integration
```

---

## Task 5.1: Verify and Complete Unit Tests

Unit tests should have been created alongside each component in Phases 2 and 3. This task verifies their existence and fills any gaps.

### Step 5.1.1: Check Component Tests

Verify test files exist and coverage is adequate:

| Component | Test File | Expected Tests | Coverage Target |
|-----------|-----------|----------------|-----------------|
| CardList | `internal/ui/components/card_list_test.go` | New, SetCards, Selection, Search, Callbacks | ≥ 70% |
| SessionList | `internal/ui/components/session_list_test.go` | New, SetSessions, StatusIndicators | ≥ 70% |
| SessionPreview | `internal/ui/components/session_preview_test.go` | New, SetSession, Render, Clear | ≥ 70% |
| SourceList | `internal/ui/components/source_list_test.go` | New, SetSources, CRUD | ≥ 70% |
| Modal | `internal/ui/components/modal_test.go` | Open, Close, IsOpen | ≥ 70% |

**Verification Command:**
```bash
go test ./internal/ui/components/... -cover

# If coverage < 70% for any package:
go test ./internal/ui/components/... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

**If test file missing:** Create minimal tests for:
- Primary data setter (`SetCards`, `SetSessions`, etc.)
- Selection/navigation
- View rendering (non-empty output)

### Step 5.1.2: Check View Tests

| View | Test File | Expected Tests | Coverage Target |
|------|-----------|----------------|-----------------|
| Dashboard | `internal/ui/views/dashboard_test.go` | New, SetData, TogglePane, Search, Modals, SetSize, View | ≥ 70% |
| CardEditor | `internal/ui/views/card_editor_test.go` | New, SetCard, Reset, Save, Validation, FocusCycle | ≥ 70% |
| SourceManager | `internal/ui/views/source_manager_test.go` | New, SetType, CRUD operations | ≥ 70% |
| IDEConfig | `internal/ui/views/ide_config_test.go` | New, Save, Launch | ≥ 70% |
| MCPForm | `internal/ui/views/mcp_form_test.go` | New, Save, Validation | ≥ 70% |
| SessionLauncher | `internal/ui/views/session_launcher_test.go` | New, SetCard, SetName, Launch | ≥ 70% |

**Verification Command:**
```bash
go test ./internal/ui/views/... -cover

# If coverage < 70% for any package:
go test ./internal/ui/views/... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

**If test file missing:** Create minimal tests for:
- Constructor
- Data setters
- Key navigation
- View rendering (non-empty output)

### Step 5.1.3: Fill Coverage Gaps

Only create tests for untested scenarios. Follow existing test patterns from:
- `internal/ui/app_dashboard_test.go`
- `internal/ui/app_card_form_test.go`

---

## Task 5.2: Integration Tests

### Step 5.2.1: Create Integration Test Suite

**File**: `internal/ui/integration_test.go` (NEW)

```go
package ui

import (
    "testing"
    
    tea "github.com/charmbracelet/bubbletea"
    
    "kankuro/internal/config"
    "kankuro/internal/session"
)

func TestFullSessionLaunchFlow(t *testing.T) {
    repo, _ := config.NewFileRepository(t.TempDir())
    sm := session.NewManager(repo.Dir(), repo)
    
    wd := t.TempDir()
    repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
    
    card, _ := repo.CreateIDCard(config.IDCard{
        Name:   "Test Card",
        Driver: "echo",
        Model:  "test",
    })
    
    coord := NewCoordinator(repo, sm)
    coord.Init()
    
    if coord.currentView != ViewDashboard {
        t.Errorf("Expected ViewDashboard, got %v", coord.currentView)
    }
    
    dashboard := coord.views[ViewDashboard]
    dashboard.SetSize(120, 40)
    
    msg := tea.KeyMsg{Runes: []rune{'n'}}
    coord.Update(msg)
    
    if coord.currentView != ViewCardEditor {
        t.Errorf("Expected ViewCardEditor after 'n', got %v", coord.currentView)
    }
}

func TestCardCRUDFlow(t *testing.T) {
    repo, _ := config.NewFileRepository(t.TempDir())
    sm := session.NewManager(repo.Dir(), repo)
    
    wd := t.TempDir()
    repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
    
    coord := NewCoordinator(repo, sm)
    coord.Init()
    
    editor := coord.views[ViewCardEditor]
    
    testCard := config.IDCard{
        Name:   "New Card",
        Driver: "claude",
        Model:  "opus",
    }
    
    editor.(*views.CardEditor).SetCard(testCard)
    
    saved, err := editor.(*views.CardEditor).Save()
    if err != nil {
        t.Fatalf("Save failed: %v", err)
    }
    
    if saved.ID == "" {
        t.Error("Saved card should have an ID")
    }
    
    cards, _ := repo.ListIDCards()
    if len(cards) != 1 {
        t.Errorf("Expected 1 card in repo, got %d", len(cards))
    }
}

func TestSourceManagementFlow(t *testing.T) {
    repo, _ := config.NewFileRepository(t.TempDir())
    sm := session.NewManager(repo.Dir(), repo)
    
    coord := NewCoordinator(repo, sm)
    coord.Init()
    
    coord.navigateTo(ViewSourceManager)
    sm := coord.views[ViewSourceManager].(*views.SourceManager)
    sm.SetType(views.SourceTypeSkills)
    
    sm.AddSource("/path/to/skills")
    
    sources, _ := repo.ListSkillSources()
    if len(sources) != 1 {
        t.Errorf("Expected 1 skill source, got %d", len(sources))
    }
}

func TestSessionManagementFlow(t *testing.T) {
    repo, _ := config.NewFileRepository(t.TempDir())
    sm := session.NewManager(repo.Dir(), repo)
    
    wd := t.TempDir()
    repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
    workspaces, _ := repo.ListWorkspaces()
    
    card, _ := repo.CreateIDCard(config.IDCard{
        Name:        "Test",
        Driver:      "echo",
        Model:       "test",
        WorkspaceID: workspaces[0].ID,
    })
    
    coord := NewCoordinator(repo, sm)
    coord.Init()
    
    launcher := coord.views[ViewNaming].(*views.SessionLauncher)
    launcher.SetCard(card)
    launcher.SetName("Test Session")
    
    sessionID, err := launcher.Launch()
    if err != nil {
        t.Fatalf("Launch failed: %v", err)
    }
    
    sessions := sm.List()
    if len(sessions) != 1 {
        t.Errorf("Expected 1 session, got %d", len(sessions))
    }
    
    if sessions[0].ID != sessionID {
        t.Error("Session ID mismatch")
    }
}

func TestNavigationFlow(t *testing.T) {
    repo, _ := config.NewFileRepository(t.TempDir())
    sm := session.NewManager(repo.Dir(), repo)
    
    coord := NewCoordinator(repo, sm)
    coord.Init()
    
    views := []ViewType{
        ViewDashboard,
        ViewCardEditor,
        ViewSourceManager,
        ViewIDEConfig,
        ViewMCPForm,
    }
    
    for _, v := range views {
        coord.navigateTo(v)
        if coord.currentView != v {
            t.Errorf("Expected %v, got %v", v, coord.currentView)
        }
    }
}

func TestCoordinatorErrorHandling(t *testing.T) {
    repo, _ := config.NewFileRepository(t.TempDir())
    sm := session.NewManager(repo.Dir(), repo)
    
    coord := NewCoordinator(repo, sm)
    coord.Init()
    
    coord.Update(messages.ErrorMsg{Err: errors.New("test error")})
    
    if coord.err == nil {
        t.Error("Expected error to be set")
    }
    
    coord.Update(messages.ClearErrorMsg{})
    
    if coord.err != nil {
        t.Error("Expected error to be cleared")
    }
}

func TestCoordinatorDataPropagation(t *testing.T) {
    repo, _ := config.NewFileRepository(t.TempDir())
    sm := session.NewManager(repo.Dir(), repo)
    
    wd := t.TempDir()
    repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
    
    card, _ := repo.CreateIDCard(config.IDCard{
        Name:   "Test",
        Driver: "claude",
        Model:  "opus",
    })
    
    coord := NewCoordinator(repo, sm)
    coord.Init()
    
    coord.refreshData()
    
    dashboard := coord.views[ViewDashboard].(*views.Dashboard)
    if len(dashboard.Cards()) != 1 {
        t.Errorf("Expected 1 card in dashboard, got %d", len(dashboard.Cards()))
    }
}
```

### Step 5.2.2: Create Coordinator Unit Tests

**File**: `internal/ui/coordinator_test.go` (NEW)

```go
package ui

import (
    "errors"
    "testing"
    
    "kankuro/internal/config"
    "kankuro/internal/session"
)

func TestCoordinator_New(t *testing.T) {
    repo, _ := config.NewFileRepository(t.TempDir())
    sm := session.NewManager(repo.Dir(), repo)
    
    coord := NewCoordinator(repo, sm)
    
    if coord == nil {
        t.Fatal("NewCoordinator returned nil")
    }
    
    if coord.currentView != ViewDashboard {
        t.Errorf("Expected initial view ViewDashboard, got %v", coord.currentView)
    }
    
    if len(coord.views) != 6 {
        t.Errorf("Expected 6 views, got %d", len(coord.views))
    }
}

func TestCoordinator_Navigate(t *testing.T) {
    repo, _ := config.NewFileRepository(t.TempDir())
    sm := session.NewManager(repo.Dir(), repo)
    
    coord := NewCoordinator(repo, sm)
    coord.Init()
    
    coord.navigateTo(ViewCardEditor)
    
    if coord.currentView != ViewCardEditor {
        t.Errorf("Expected ViewCardEditor, got %v", coord.currentView)
    }
    if coord.prevView != ViewDashboard {
        t.Errorf("Expected prevView ViewDashboard, got %v", coord.prevView)
    }
}

func TestCoordinator_ResizeAll(t *testing.T) {
    repo, _ := config.NewFileRepository(t.TempDir())
    sm := session.NewManager(repo.Dir(), repo)
    
    coord := NewCoordinator(repo, sm)
    coord.Init()
    
    coord.width = 120
    coord.height = 40
    coord.resizeAll()
    
    for vt, view := range coord.views {
        if view.Width() != 120 {
            t.Errorf("View %v width should be 120, got %d", vt, view.Width())
        }
        if view.Height() != 40 {
            t.Errorf("View %v height should be 40, got %d", vt, view.Height())
        }
    }
}

func TestCoordinator_WireCommands(t *testing.T) {
    repo, _ := config.NewFileRepository(t.TempDir())
    sm := session.NewManager(repo.Dir(), repo)
    
    cmds := wireCommands(repo, sm)
    
    if cmds == nil {
        t.Fatal("wireCommands returned nil")
    }
    
    if cmds.LaunchSession == nil {
        t.Error("LaunchSession should not be nil")
    }
    if cmds.SaveCard == nil {
        t.Error("SaveCard should not be nil")
    }
}

func TestApp_BackwardsCompatibility(t *testing.T) {
    repo, _ := config.NewFileRepository(t.TempDir())
    sm := session.NewManager(repo.Dir(), repo)
    
    app, err := New(repo, sm)
    if err != nil {
        t.Fatalf("New failed: %v", err)
    }
    
    if app.coordinator == nil {
        t.Error("App.coordinator should not be nil")
    }
    
    cmd := app.Init()
    if cmd == nil {
        t.Error("App.Init should return a command")
    }
    
    app.SetSize(120, 40)
    if app.coordinator.width != 120 {
        t.Error("SetSize should propagate to coordinator")
    }
    
    view := app.View()
    if view == "" {
        t.Error("View should not be empty")
    }
}
```

---

## Task 5.3: Visual Regression Check

### Step 5.3.1: Manual Testing Checklist

**Dashboard Navigation**
- [ ] Tab switches between cards and sessions panes
- [ ] Arrow keys navigate within lists
- [ ] Search (/) activates search mode
- [ ] Esc clears search or exits
- [ ] Header and footer render correctly
- [ ] Empty sessions state shows crow ASCII art
- [ ] Focus indicators (border colors) update correctly

**Card Creation/Editing**
- [ ] `n` opens card editor with empty form
- [ ] `e` on selected card opens editor with pre-filled data
- [ ] Tab cycles through form fields
- [ ] Enter saves card and returns to dashboard
- [ ] Esc cancels and returns to dashboard
- [ ] Workspace picker opens with Space
- [ ] Puppet selector shows available faces
- [ ] Skill/Persona/MCP lists allow selection

**Session Management**
- [ ] Enter on card launches session naming
- [ ] Enter on session attaches (if running)
- [ ] `x` kills selected session
- [ ] Delete/Backspace on terminated session shows delete confirmation
- [ ] `s` on terminated session opens agent switcher
- [ ] Status indicators show correct colors (running/waiting/terminated)

**Source Management**
- [ ] `w` opens workspace sources
- [ ] `n` adds new source
- [ ] `e` edits selected source
- [ ] `d` deletes selected source
- [ ] Esc returns to dashboard
- [ ] Skill/Persona sources work identically

**MCP Configuration**
- [ ] `m` opens MCP form
- [ ] Tab cycles through MCP fields
- [ ] Enter saves MCP
- [ ] Esc cancels
- [ ] Delete removes MCP

**IDE Integration**
- [ ] `i` opens IDE config
- [ ] Path input works correctly
- [ ] Save persists config
- [ ] Launch opens IDE for session

**Error Handling**
- [ ] Errors display at top of screen
- [ ] Errors clear on navigation
- [ ] Form validation errors show inline

---

## Verification Commands

```bash
# Build all packages
go build ./...

# Run all unit tests
go test ./internal/ui/... -v -cover

# Run integration tests specifically
go test ./internal/ui/integration_test.go -v

# Run coordinator tests
go test ./internal/ui/coordinator_test.go -v

# Check test coverage
go test ./internal/ui/... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Run all tests with race detection
go test ./internal/ui/... -race

# Vet
go vet ./internal/ui/...
```

---

## Success Criteria

### Phase 4
- [ ] `Coordinator` struct defined in `internal/ui/coordinator.go`
- [ ] 6 views registered in views map
- [ ] Navigation between views works via `navigateTo()`
- [ ] Data propagation to views via `broadcastData()`
- [ ] Commands wired to repository operations
- [ ] App struct delegates to Coordinator
- [ ] Existing tests pass without modification
- [ ] No changes to public API of App

### Phase 5
- [ ] All component unit tests pass (≥70% coverage)
- [ ] Integration tests cover main user flows
- [ ] Coordinator unit tests pass
- [ ] App backwards compatibility verified
- [ ] Manual testing checklist completed
- [ ] No visual regressions detected
- [ ] All existing tests still pass

---

## Dependencies

```
Phase 3 Complete
├── components/card_list.go ✓
├── components/session_list.go ✓
├── components/session_preview.go ✓
├── components/source_list.go ✓
├── components/modal.go ✓
├── views/dashboard.go ✓
├── views/card_editor.go ✓
├── views/source_manager.go ✓
├── views/ide_config.go ✓
├── views/mcp_form.go ✓
└── views/session_launcher.go ✓

Phase 4
├── internal/ui/coordinator.go ← NEW
├── internal/ui/callbacks.go ← NEW
├── internal/ui/wiring.go ← NEW
├── internal/ui/app.go ← MODIFY
└── internal/ui/commands/commands.go ← MODIFY

Phase 5
├── internal/ui/components/*_test.go ← VERIFY EXISTS / FILL GAPS
├── internal/ui/views/*_test.go ← VERIFY EXISTS / FILL GAPS
├── internal/ui/integration_test.go ← NEW
└── internal/ui/coordinator_test.go ← NEW
```

---

## Estimated Time

| Phase | Task | Time |
|-------|------|------|
| 4 | 4.1 Coordinator struct | 60 min |
| 4 | 4.1.2 Callback interfaces | 45 min |
| 4 | 4.2 Wire commands | 30 min |
| 4 | 4.3 App wrapper | 30 min |
| 4 | Verification | 20 min |
| **Phase 4 Total** | | **~3 hours** |
| 5 | 5.1 Verify/complete unit tests | 45 min |
| 5 | 5.2 Integration tests | 60 min |
| 5 | 5.3 Manual verification | 45 min |
| **Phase 5 Total** | | **~2.5 hours** |
| **Grand Total** | | **~5.5 hours** |

---

## Notes

1. **Sequential Execution Required**: Phase 4 tasks must be executed in order (4.1 → 4.2 → 4.3)

2. **Parallel Testing Possible**: Phase 5.1 (unit tests verification) and 5.2 (integration tests) can run in parallel after Phase 4 is complete

3. **Unit Tests Already Exist**: Unit tests for components and views should have been created in Phases 2 and 3. Task 5.1 only verifies and fills gaps.

3. **Migration Strategy**: The App wrapper ensures zero breaking changes. Existing code that imports `ui.App` continues to work.

4. **Test Coverage Target**: ≥70% per component, ≥60% overall for the ui package

5. **Future Extensibility**: New views can be added by:
   - Defining a new ViewType
   - Creating the view component in `internal/ui/views/`
   - Registering in `initViews()`
   - Adding callbacks if needed
