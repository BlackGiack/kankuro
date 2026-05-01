# UI Component Architecture: internal/ui Package

## Current State: No Formal Component Architecture

The `internal/ui` package does **not** implement a component-based architecture. Instead, it uses a monolithic struct with methods attached directly to it.

---

## Existing Structure Types

### Core Struct: App (God Object)

**File**: `app.go:116-155`

```go
type App struct {
    repo    config.Repository
    session *session.Manager
    mode    viewMode
    cards      []config.IDCard
    skills     []config.Skill
    personas   []config.Persona
    mcps       []config.MCPServer
    sessions   []session.Snapshot
    workspaces []config.Workspace
    cardsList    list.Model
    sessionsList list.Model
    switchList   list.Model
    cardForm    CardFormState
    sessionView SessionViewState
    naming      NamingState
    skillSources   SourceFormState
    personaSources SourceFormState
    workspaceForm  WorkspaceFormState
    mcpInputs []textinput.Model
    ideConfig config.IDEConfig
    ideInput  textinput.Model
    help help.Model
    width  int
    height int
    err    error
}
```

**Role**: Single entry point for all UI. Acts as:
- State container
- Event dispatcher
- View renderer
- Business logic handler

---

### State Containers (Data Only, Not Components)

#### CardFormState

**File**: `app_state.go:11-51`

```go
type CardFormState struct {
    Inputs              []textinput.Model
    FocusIndex          int
    Puppet              string
    PuppetSelectorIndex int
    PuppetSelectorPrev  string
    EditingID           string
    Editor              textarea.Model
    Content             string
    Path                string
    PreferPath          bool
    PathInput           textinput.Model
    SelectedWorkspaceID string
    SelectedSkillIDs    map[string]bool
    SelectedPersonaIDs  map[string]bool
    SelectedMCPIDs      map[string]bool
    SkillCursor         int
    PersonaCursor       int
    MCPCursor           int
    WorkspaceCursor     int
    SkillSearchQuery    string
    PersonaSearchQuery  string
    MCPSearchQuery      string
    WorkspaceSearchQuery string
    MCPFormPrevMode     viewMode
}
```

**Role**: Data container for ID card form. Has no methods except constructor `NewCardFormState()`.

---

#### SessionViewState

**File**: `app_state.go:54-71`

```go
type SessionViewState struct {
    ActiveID            string
    PendingDeleteID     string
    PendingDeleteCardID string
    SwitchID            string
    FocusedPane         dashboardPane
    CardsSearchQuery    string
    SessionsSearchQuery string
    IsSearching         bool
    PreviewSessionID    string
    PreviewScrollback   []byte
    CardsWidth          int
    SessionsWidth       int
    PreviewWidth        int
}
```

**Role**: Dashboard state. No methods.

---

#### NamingState

**File**: `app_state.go:74-88`

```go
type NamingState struct {
    SessionID  string
    Card       config.IDCard
    Input      textinput.Model
    LaunchDone bool
    LaunchErr  error
    Submitted  bool
}
```

**Role**: Session naming prompt state. Constructor only.

---

#### SourceFormState

**File**: `app_state.go:91-105`

```go
type SourceFormState struct {
    Sources    []string
    Cursor     int
    Input      textinput.Model
    EditingIdx int
}
```

**Role**: Generic source list form state. Used for skill and persona sources.

---

#### WorkspaceFormState

**File**: `app_state.go:108-125`

```go
type WorkspaceFormState struct {
    Sources     []config.WorkspaceSource
    Cursor      int
    Inputs      []textinput.Model
    EditingID   string
    SearchQuery string
}
```

**Role**: Workspace source form state.

---

### Item Types (Data Transfer Objects)

#### sessionListItem

**File**: `app_items.go:10-42`

```go
type sessionListItem struct {
    sessionName   string
    cardNameFace  string
    desc          string
    sessionDetail string
    id            string
    status        session.Status
}

func (i sessionListItem) Title() string
func (i sessionListItem) Description() string
func (i sessionListItem) FilterValue() string
```

**Role**: Implements `list.Item` interface for sessions list.

---

#### cardItem

**File**: `app_items.go:44-53`

```go
type cardItem struct {
    card     config.IDCard
    skills   []config.Skill
    personas []config.Persona
    status   session.Status
}

func (i cardItem) Title() string
func (i cardItem) Description() string
func (i cardItem) FilterValue() string
```

**Role**: Implements `list.Item` interface for cards list.

---

### Delegate Types (List Renderers)

#### cardDelegate

**File**: `app_lists.go:16-42`

```go
type cardDelegate struct {
    height   int
    skills   []config.Skill
    personas []config.Persona
}

func newCardDelegate(skills []config.Skill, personas []config.Persona) cardDelegate
func (d cardDelegate) Height() int
func (d cardDelegate) Spacing() int
func (d cardDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd
func (d cardDelegate) Render(w io.Writer, m list.Model, index int, item list.Item)
```

**Role**: Implements `list.DefaultDelegate` interface for custom card rendering.

---

### Enum Types

#### viewMode

**File**: `app.go:12-35`

```go
const (
    viewDashboard viewMode = iota
    viewCardForm
    viewPuppetSelector
    viewMCPForm
    viewDeleteConfirm
    viewSwitchAgent
    viewAgentsEditor
    viewAgentsPathInput
    viewSkillSources
    viewSkillSourceForm
    viewPersonaSources
    viewPersonaSourceForm
    viewWorkspaces
    viewWorkspaceSourceForm
    viewWorkspacePicker
    viewNaming
    viewIDE
    viewIDEForm
)
```

**Role**: View state machine state.

---

#### dashboardPane

**File**: `app.go:41-44`

```go
const (
    paneCards dashboardPane = iota
    paneSessions
)
```

**Role**: Dashboard focus state.

---

## View/Update Pattern

Instead of components, the codebase organizes by feature:

### View Files (Rendering)

| File | Purpose |
|------|---------|
| `view_card_form.go` | Card form rendering |
| `view_preview.go` | Session preview rendering |
| `view_naming.go` | Session naming prompt |
| `view_mcp.go` | MCP form rendering |
| `view_switch.go` | Agent switch dialog |
| `view_delete.go` | Delete confirmation |
| `view_puppet_selector.go` | Puppet style picker |
| `view_workspaces.go` | Workspace management |
| `view_sources.go` | Skill/persona source lists |
| `view_ide.go` | IDE configuration |
| `view_agents.go` | AGENTS.md editor |
| `app_view.go` | Main view dispatcher |
| `app_render.go` | Shared render helpers |

### Update Files (Event Handling)

| File | Purpose |
|------|---------|
| `app_update.go` | Central update dispatcher |
| `app_dashboard.go` | Dashboard key handling |
| `app_card_form.go` | Card form updates |
| `app_naming.go` | Naming prompt updates |
| `app_mcp.go` | MCP form updates |
| `app_switch.go` | Agent switch updates |
| `app_workspaces.go` | Workspace management |
| `app_skill_sources.go` | Skill source management |
| `app_persona_sources.go` | Persona source management |
| `app_agents.go` | AGENTS.md editor updates |
| `app_ide.go` | IDE config updates |
| `app_new.go` | App initialization |
| `app_lists.go` | List management methods |
| `app_inputs.go` | Focus cycling helpers |

---

## Architecture Pattern Summary

**Pattern**: Feature-based flat structure with monolithic state

**Structure**:
```
internal/ui/
├── app.go              # God struct + constants
├── app_state.go        # State containers (data only)
├── app_*.go            # Update logic (attached to App)
├── view_*.go           # Render logic (attached to App)
├── attach.go           # PTY attach logic
├── styles.go           # Global style definitions
├── utils.go            # Helper functions
└── app_items.go        # List item types
```

**Key Characteristics**:
- No component hierarchy
- State and behavior mixed in App
- Views are methods, not independent entities
- Shared state accessed directly
- Bubble Tea models (list, textinput) embedded in App

---

## Missing Abstractions

1. **Component Interface**: No `Component` or `Model` interface for composable UI elements
2. **Props/State Pattern**: No prop passing, everything reads from App
3. **Lifecycle Hooks**: No init/destroy pattern for views
4. **Event Bubbling**: No parent-child event propagation
5. **Layout Components**: No dedicated layout types (Box, Grid, Stack)
6. **Reusable Widgets**: No reusable UI widgets beyond Bubble Tea primitives

---

## Recommended Architecture

```go
// Proposed component interface
type Component interface {
    Init() tea.Cmd
    Update(tea.Msg) (Component, tea.Cmd)
    View() string
}

// Example component
type CardList struct {
    list      list.Model
    cards     []config.IDCard
    onSelect  func(config.IDCard)
    onEdit    func(config.IDCard)
}

func (c CardList) Update(msg tea.Msg) (Component, tea.Cmd) {
    // Handle own messages, emit commands
}

func (c CardList) View() string {
    return c.list.View()
}

// Dashboard composes components
type Dashboard struct {
    cardList    CardList
    sessionList SessionList
    preview     SessionPreview
    focused     FocusState
}
```

This would enable:
- Independent testing
- Reusable components
- Clear data flow
- Easier extensions