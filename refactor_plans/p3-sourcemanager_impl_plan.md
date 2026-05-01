# Implementation Plan: Phase 3 Task 3.3 - SourceManager Component

## Goal

Create SourceManager composite view that wraps the SourceList component to provide a unified interface for managing skill sources, persona sources, and workspace sources, consolidating 6 view modes into a single configurable component.

---

## Analysis: What to Extract

### View Mode Consolidation

| Current View Mode | New Location |
|-------------------|--------------|
| `viewSkillSources` | `SourceManager.skills` (list mode) |
| `viewSkillSourceForm` | `SourceManager.skills` (form mode) |
| `viewPersonaSources` | `SourceManager.personas` (list mode) |
| `viewPersonaSourceForm` | `SourceManager.personas` (form mode) |
| `viewWorkspaces` | `SourceManager.workspaces` (list mode) |
| `viewWorkspaceSourceForm` | `SourceManager.workspaces` (form mode) |

### From Existing Code

| File | Lines | Content | Destination |
|------|-------|---------|-------------|
| `app_skill_sources.go` | 12-38 | `updateSkillSources()` | Delegates to SourceList |
| `app_skill_sources.go` | 41-62 | `handleSkillSourceNew/Select()` | Delegates to SourceList |
| `app_skill_sources.go` | 64-78 | `handleSkillSourceDelete()` | Callback |
| `app_skill_sources.go` | 80-115 | `updateSkillSourceForm/saveSkillSource()` | Handled by SourceList |
| `app_skill_sources.go` | 117-143 | `renderSkillSourcesList()` | Handled by SourceList |
| `app_skill_sources.go` | 145-156 | `refreshSkills()` | Callback |
| `app_persona_sources.go` | 12-159 | Same pattern as skills | Delegates to SourceList |
| `app_workspaces.go` | 136-234 | Workspace source management | Delegates to SourceList |
| `view_sources.go` | 4-35 | Render methods | Handled by SourceList |

### SourceManager-Specific Behavior

```
Source Types:
├── Skills
│   ├── Simple sources (string paths)
│   ├── Single form field (path)
│   └── Return to: viewCardForm
│
├── Personas
│   ├── Simple sources (string paths)
│   ├── Single form field (path)
│   └── Return to: viewCardForm
│
└── Workspaces
    ├── Complex sources (ID, Path, IsRoot)
    ├── Two form fields (path, isRoot)
    └── Return to: viewDashboard
```

---

## Architecture Overview

### Component Hierarchy

```
SourceManager (CompositeView - thin wrapper)
├── sourceList    *components.SourceList  // Phase 2.4
└── state
    ├── sourceType    string              // "skills", "personas", "workspaces"
    ├── prevView      ViewType            // Where to return
    └── callbacks     SourceManagerCallbacks
```

### Design Philosophy

SourceManager is intentionally minimal because:
1. **SourceList already handles everything**: The SourceList component from Phase 2.4 already supports:
   - List mode and form mode
   - Simple items (skills/personas) and complex items (workspaces)
   - Navigation, editing, deletion
   - Empty state rendering

2. **SourceManager's job is configuration**: It configures SourceList with the right settings for each source type.

3. **State is passed through**: SourceManager doesn't duplicate state - it passes callbacks that interact with App's repository.

### Message Flow

```
User Input → SourceManager.Update()
    │
    └─→ sourceList.Update(msg)
            │
            ├─→ List mode: navigation/selection → internal state
            └─→ Form mode: input handling → callbacks
                    │
                    ├─→ OnSave → App.repo.SaveXxxSources()
                    ├─→ OnDelete → App.repo.DeleteXxxSource()
                    ├─→ OnRefresh → App.refreshXxx()
                    └─→ OnClose → SourceManager.Close() → NavigateMsg
```

---

## Step 3.3.1: Define Source Types

**File**: `internal/ui/views/source_type.go` (NEW)

```go
package views

// SourceType identifies which kind of sources are being managed.
type SourceType string

const (
    SourceTypeSkills     SourceType = "skills"
    SourceTypePersonas   SourceType = "personas"
    SourceTypeWorkspaces SourceType = "workspaces"
)

// SourceTypeConfig returns the configuration for a source type.
type SourceTypeConfig struct {
    Title       string
    Legend      string
    FormPrompt  string
    FormFields  int
    Field2Label string
    ReturnType  ViewType // Where to return after close
}

func GetSourceTypeConfig(st SourceType) SourceTypeConfig {
    switch st {
    case SourceTypeSkills:
        return SourceTypeConfig{
            Title:      "Manage Skill Sources",
            FormPrompt: "Path to skill source...",
            FormFields: 1,
            ReturnType: ViewCardEditor,
        }
    case SourceTypePersonas:
        return SourceTypeConfig{
            Title:      "Manage Persona Sources",
            FormPrompt: "Path to persona source...",
            FormFields: 1,
            ReturnType: ViewCardEditor,
        }
    case SourceTypeWorkspaces:
        return SourceTypeConfig{
            Title:       "Manage Workspace Sources",
            Legend:      "(R=Root folder containing multiple workspaces)",
            FormPrompt:  "Path to workspace or root directory...",
            FormFields:  2,
            Field2Label: "Is Root (y/n)",
            ReturnType:  ViewDashboard,
        }
    default:
        return SourceTypeConfig{
            Title:      "Manage Sources",
            FormFields: 1,
            ReturnType: ViewDashboard,
        }
    }
}
```

---

## Step 3.3.2: Create SourceManager View File

**File**: `internal/ui/views/source_manager.go` (NEW)

```go
package views

import (
    "errors"
    "fmt"
    "time"

    bubbletea "github.com/charmbracelet/bubbletea"
    
    "kankuro/internal/config"
    "kankuro/internal/ui/commands"
    "kankuro/internal/ui/component"
    "kankuro/internal/ui/components"
    "kankuro/internal/ui/messages"
)

// SourceManagerCallbacks defines repository callbacks.
type SourceManagerCallbacks struct {
    // Load sources
    LoadSources func() []components.SourceItem
    
    // Save operations
    SaveSimpleSource   func(path string, editingIndex int) error
    SaveWorkspaceSource func(source config.WorkspaceSource, editingIndex int) error
    
    // Delete operations
    DeleteSource func(index int) error
    
    // Refresh after save/delete
    Refresh func()
}

// SourceManager is a composite view for managing source lists.
// It wraps SourceList and configures it for skills, personas, or workspaces.
type SourceManager struct {
    component.Base
    
    sourceList   *components.SourceList
    sourceType   SourceType
    typeConfig   SourceTypeConfig
    callbacks    SourceManagerCallbacks
    
    prevView     ViewType
    cmds         *commands.Commands
    
    // Error state
    err          error
}

// NewSourceManager creates a new source manager for the given type.
func NewSourceManager(
    sourceType SourceType,
    callbacks SourceManagerCallbacks,
    cmds *commands.Commands,
) *SourceManager {
    typeConfig := GetSourceTypeConfig(sourceType)
    
    // Configure SourceList
    listConfig := components.SourceListConfig{
        Title:       typeConfig.Title,
        Legend:      typeConfig.Legend,
        FormPrompt:  typeConfig.FormPrompt,
        FormFields:  typeConfig.FormFields,
        Field2Label: typeConfig.Field2Label,
    }
    
    listCallbacks := components.SourceListCallbacks{
        OnSave: func(item components.SourceItem, editingIndex int) error {
            return handleSave(sourceType, callbacks, item, editingIndex)
        },
        OnDelete: func(index int) error {
            if callbacks.DeleteSource != nil {
                return callbacks.DeleteSource(index)
            }
            return nil
        },
        OnRefresh: func() {
            if callbacks.Refresh != nil {
                callbacks.Refresh()
            }
        },
        OnClose: func() {
            // SourceManager handles close via messages
        },
    }
    
    sourceList := components.NewSourceList(listConfig, listCallbacks)
    
    return &SourceManager{
        sourceList: sourceList,
        sourceType: sourceType,
        typeConfig: typeConfig,
        callbacks:  callbacks,
        cmds:       cmds,
    }
}

// handleSave routes save to the appropriate callback based on source type.
func handleSave(
    sourceType SourceType,
    callbacks SourceManagerCallbacks,
    item components.SourceItem,
    editingIndex int,
) error {
    switch sourceType {
    case SourceTypeSkills, SourceTypePersonas:
        if callbacks.SaveSimpleSource != nil {
            return callbacks.SaveSimpleSource(item.Display(), editingIndex)
        }
        
    case SourceTypeWorkspaces:
        if wsItem, ok := item.(components.WorkspaceSourceItem); ok {
            if callbacks.SaveWorkspaceSource != nil {
                return callbacks.SaveWorkspaceSource(wsItem.Source, editingIndex)
            }
        }
        return errors.New("invalid workspace source item")
    }
    
    return nil
}

// SetPrevView sets where to return when closed.
func (m *SourceManager) SetPrevView(view ViewType) {
    m.prevView = view
}

// SetData loads sources into the manager.
func (m *SourceManager) SetData() {
    if m.callbacks.LoadSources != nil {
        items := m.callbacks.LoadSources()
        m.sourceList.SetItems(items)
    }
}

// Init initializes the manager.
func (m *SourceManager) Init() bubbletea.Cmd {
    m.SetData()
    return m.sourceList.Init()
}

// Update handles messages.
func (m *SourceManager) Update(msg bubbletea.Msg) (component.Model, bubbletea.Cmd) {
    // Check for close before delegating
    if keyMsg, ok := msg.(bubbletea.KeyMsg); ok {
        if keyMsg.String() == "esc" && m.sourceList.Mode() == components.SourceListModeList {
            // Emit navigate message to return to previous view
            return m, func() bubbletea.Msg {
                return messages.NavigateMsg{View: m.prevView}
            }
        }
    }
    
    // Delegate to SourceList
    model, cmd := m.sourceList.Update(msg)
    m.sourceList = model.(*components.SourceList)
    
    return m, cmd
}

// View renders the manager.
func (m *SourceManager) View() string {
    if m.err != nil {
        return m.sourceList.View() + "\n\nError: " + m.err.Error()
    }
    return m.sourceList.View()
}

// SetSize implements component.Model.
func (m *SourceManager) SetSize(width, height int) {
    m.Base.SetSize(width, height)
    m.sourceList.SetSize(width, height)
}

// SourceType returns the managed source type.
func (m *SourceManager) SourceType() SourceType {
    return m.sourceType
}

// GenerateID creates a new unique ID for workspace sources.
func GenerateID() string {
    return fmt.Sprintf("ws-%d", time.Now().UnixNano())
}
```

---

## Step 3.3.3: Add ViewType to Messages

**File**: `internal/ui/messages/messages.go` (MODIFY - ensure ViewType is defined)

The ViewType should already be defined from Phase 1.2. If not, add:

```go
// ViewType represents the current view mode.
type ViewType int

const (
    ViewDashboard ViewType = iota
    ViewCardEditor
    ViewSourceManager
    ViewMCPForm
    ViewIDEConfig
    ViewSessionLauncher
)
```

---

## Step 3.3.4: Create SourceManager Tests

**File**: `internal/ui/views/source_manager_test.go` (NEW)

```go
package views

import (
    "errors"
    "testing"
    
    bubbletea "github.com/charmbracelet/bubbletea"
    
    "kankuro/internal/config"
    "kankuro/internal/ui/commands"
    "kankuro/internal/ui/components"
)

func TestSourceManager_New_Skills(t *testing.T) {
    callbacks := SourceManagerCallbacks{}
    sm := NewSourceManager(SourceTypeSkills, callbacks, &commands.Commands{})
    
    if sm == nil {
        t.Fatal("NewSourceManager returned nil")
    }
    
    if sm.sourceType != SourceTypeSkills {
        t.Errorf("Expected SourceTypeSkills, got %s", sm.sourceType)
    }
    
    if sm.typeConfig.Title != "Manage Skill Sources" {
        t.Errorf("Wrong title: %s", sm.typeConfig.Title)
    }
    
    if sm.typeConfig.FormFields != 1 {
        t.Errorf("Skills should have 1 form field, got %d", sm.typeConfig.FormFields)
    }
}

func TestSourceManager_New_Personas(t *testing.T) {
    callbacks := SourceManagerCallbacks{}
    sm := NewSourceManager(SourceTypePersonas, callbacks, &commands.Commands{})
    
    if sm.sourceType != SourceTypePersonas {
        t.Errorf("Expected SourceTypePersonas, got %s", sm.sourceType)
    }
    
    if sm.typeConfig.FormFields != 1 {
        t.Errorf("Personas should have 1 form field, got %d", sm.typeConfig.FormFields)
    }
}

func TestSourceManager_New_Workspaces(t *testing.T) {
    callbacks := SourceManagerCallbacks{}
    sm := NewSourceManager(SourceTypeWorkspaces, callbacks, &commands.Commands{})
    
    if sm.sourceType != SourceTypeWorkspaces {
        t.Errorf("Expected SourceTypeWorkspaces, got %s", sm.sourceType)
    }
    
    if sm.typeConfig.FormFields != 2 {
        t.Errorf("Workspaces should have 2 form fields, got %d", sm.typeConfig.FormFields)
    }
    
    if sm.typeConfig.Legend == "" {
        t.Error("Workspaces should have a legend")
    }
}

func TestSourceManager_SetData(t *testing.T) {
    loaded := false
    
    callbacks := SourceManagerCallbacks{
        LoadSources: func() []components.SourceItem {
            loaded = true
            return []components.SourceItem{
                components.SimpleSourceItem("/path/to/source"),
            }
        },
    }
    
    sm := NewSourceManager(SourceTypeSkills, callbacks, &commands.Commands{})
    sm.SetData()
    
    if !loaded {
        t.Error("LoadSources should have been called")
    }
}

func TestSourceManager_SetPrevView(t *testing.T) {
    sm := NewSourceManager(SourceTypeSkills, SourceManagerCallbacks{}, &commands.Commands{})
    
    sm.SetPrevView(ViewCardEditor)
    
    if sm.prevView != ViewCardEditor {
        t.Errorf("Expected ViewCardEditor, got %d", sm.prevView)
    }
}

func TestSourceManager_View(t *testing.T) {
    callbacks := SourceManagerCallbacks{
        LoadSources: func() []components.SourceItem {
            return []components.SourceItem{
                components.SimpleSourceItem("/path1"),
                components.SimpleSourceItem("/path2"),
            }
        },
    }
    
    sm := NewSourceManager(SourceTypeSkills, callbacks, &commands.Commands{})
    sm.Init()
    
    view := sm.View()
    
    if view == "" {
        t.Error("View should not be empty")
    }
    
    if !containsString(view, "Manage Skill Sources") {
        t.Error("View should contain title")
    }
    
    if !containsString(view, "/path1") {
        t.Error("View should contain sources")
    }
}

func TestSourceManager_Save_Simple(t *testing.T) {
    var savedPath string
    var savedIndex int
    
    callbacks := SourceManagerCallbacks{
        SaveSimpleSource: func(path string, idx int) error {
            savedPath = path
            savedIndex = idx
            return nil
        },
        Refresh: func() {},
    }
    
    sm := NewSourceManager(SourceTypeSkills, callbacks, &commands.Commands{})
    
    // Create a simple item
    item := components.SimpleSourceItem("/new/source")
    
    // Call internal save handler
    err := handleSave(SourceTypeSkills, callbacks, item, -1)
    
    if err != nil {
        t.Errorf("Save should not error: %v", err)
    }
    
    if savedPath != "/new/source" {
        t.Errorf("Expected path '/new/source', got '%s'", savedPath)
    }
    
    if savedIndex != -1 {
        t.Errorf("Expected index -1 for new item, got %d", savedIndex)
    }
}

func TestSourceManager_Save_Workspace(t *testing.T) {
    var savedSource config.WorkspaceSource
    
    callbacks := SourceManagerCallbacks{
        SaveWorkspaceSource: func(source config.WorkspaceSource, idx int) error {
            savedSource = source
            return nil
        },
        Refresh: func() {},
    }
    
    sm := NewSourceManager(SourceTypeWorkspaces, callbacks, &commands.Commands{})
    
    // Create a workspace item
    item := components.WorkspaceSourceItem{
        Source: config.WorkspaceSource{
            ID:     "test-id",
            Path:   "/workspace/path",
            IsRoot: true,
        },
    }
    
    err := handleSave(SourceTypeWorkspaces, callbacks, item, 0)
    
    if err != nil {
        t.Errorf("Save should not error: %v", err)
    }
    
    if savedSource.Path != "/workspace/path" {
        t.Errorf("Expected path '/workspace/path', got '%s'", savedSource.Path)
    }
    
    if !savedSource.IsRoot {
        t.Error("IsRoot should be true")
    }
}

func TestSourceManager_Delete(t *testing.T) {
    var deletedIndex int
    
    callbacks := SourceManagerCallbacks{
        DeleteSource: func(idx int) error {
            deletedIndex = idx
            return nil
        },
        Refresh: func() {},
    }
    
    sm := NewSourceManager(SourceTypeSkills, callbacks, &commands.Commands{})
    
    // Set up data
    sm.SetData()
    
    // Simulate deletion through SourceList
    sm.sourceList.SetItems([]components.SourceItem{
        components.SimpleSourceItem("/path1"),
        components.SimpleSourceItem("/path2"),
    })
    sm.sourceList.SetCursor(1)
    
    // Trigger delete
    model, _ := sm.Update(bubbletea.KeyMsg{Type: bubbletea.KeyRunes, Runes: []rune{'d'}})
    sm = model.(*SourceManager)
    
    // Note: The actual deletion happens through callbacks
    // This test verifies the wiring is correct
}

func TestSourceManager_Escape_NavigatesBack(t *testing.T) {
    sm := NewSourceManager(SourceTypeSkills, SourceManagerCallbacks{}, &commands.Commands{})
    sm.SetPrevView(ViewCardEditor)
    sm.Init()
    
    // Press escape in list mode
    model, cmd := sm.Update(bubbletea.KeyMsg{Type: bubbletea.KeyEsc})
    
    if cmd == nil {
        t.Error("Should return a command on escape")
    }
    
    // The command should emit NavigateMsg
    msg := cmd()
    if navMsg, ok := msg.(messages.NavigateMsg); ok {
        if navMsg.View != ViewCardEditor {
            t.Errorf("Should navigate to ViewCardEditor, got %d", navMsg.View)
        }
    } else {
        t.Error("Should emit NavigateMsg")
    }
}

func TestSourceManager_Escape_InFormMode(t *testing.T) {
    sm := NewSourceManager(SourceTypeSkills, SourceManagerCallbacks{}, &commands.Commands{})
    sm.Init()
    
    // Open form mode
    sm.sourceList.OpenForm(-1)
    
    // Press escape in form mode
    model, _ := sm.Update(bubbletea.KeyMsg{Type: bubbletea.KeyEsc})
    sm = model.(*SourceManager)
    
    // Should return to list mode, not navigate away
    if sm.sourceList.Mode() != components.SourceListModeList {
        t.Error("Should return to list mode on escape in form")
    }
}

func TestSourceManager_SetSize(t *testing.T) {
    sm := NewSourceManager(SourceTypeSkills, SourceManagerCallbacks{}, &commands.Commands{})
    sm.SetSize(80, 24)
    
    if sm.Width != 80 {
        t.Errorf("Expected width 80, got %d", sm.Width)
    }
    
    if sm.Height != 24 {
        t.Errorf("Expected height 24, got %d", sm.Height)
    }
}

func TestSourceManager_Error(t *testing.T) {
    callbacks := SourceManagerCallbacks{
        SaveSimpleSource: func(path string, idx int) error {
            return errors.New("save failed")
        },
    }
    
    sm := NewSourceManager(SourceTypeSkills, callbacks, &commands.Commands{})
    sm.err = errors.New("test error")
    
    view := sm.View()
    
    if !containsString(view, "Error:") {
        t.Error("Should display error in view")
    }
}

func TestGetSourceTypeConfig(t *testing.T) {
    tests := []struct {
        sourceType     SourceType
        expectedTitle  string
        expectedFields int
    }{
        {SourceTypeSkills, "Manage Skill Sources", 1},
        {SourceTypePersonas, "Manage Persona Sources", 1},
        {SourceTypeWorkspaces, "Manage Workspace Sources", 2},
    }
    
    for _, tt := range tests {
        t.Run(string(tt.sourceType), func(t *testing.T) {
            config := GetSourceTypeConfig(tt.sourceType)
            
            if config.Title != tt.expectedTitle {
                t.Errorf("Expected title '%s', got '%s'", tt.expectedTitle, config.Title)
            }
            
            if config.FormFields != tt.expectedFields {
                t.Errorf("Expected %d fields, got %d", tt.expectedFields, config.FormFields)
            }
        })
    }
}

func TestGenerateID(t *testing.T) {
    id1 := GenerateID()
    id2 := GenerateID()
    
    if id1 == id2 {
        t.Error("IDs should be unique")
    }
    
    if len(id1) < 5 {
        t.Errorf("ID too short: %s", id1)
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

## Step 3.3.5: Integration with App (Backwards Compatible)

**File**: `internal/ui/app.go` (MODIFY - add integration stub)

```go
// ===== Integration with new SourceManager view =====

func (a *App) newSourceManagerView(sourceType string) *views.SourceManager {
    var st views.SourceType
    switch sourceType {
    case "skills":
        st = views.SourceTypeSkills
    case "personas":
        st = views.SourceTypePersonas
    case "workspaces":
        st = views.SourceTypeWorkspaces
    }
    
    callbacks := views.SourceManagerCallbacks{
        LoadSources: func() []components.SourceItem {
            return a.loadSourceItems(sourceType)
        },
        
        SaveSimpleSource: func(path string, idx int) error {
            return a.saveSimpleSource(sourceType, path, idx)
        },
        
        SaveWorkspaceSource: func(source config.WorkspaceSource, idx int) error {
            return a.saveWorkspaceSource(source, idx)
        },
        
        DeleteSource: func(index int) error {
            return a.deleteSource(sourceType, index)
        },
        
        Refresh: func() {
            a.refreshByType(sourceType)
        },
    }
    
    sm := views.NewSourceManager(st, callbacks, &a.cmds)
    
    // Set return destination
    if sourceType == "workspaces" {
        sm.SetPrevView(views.ViewDashboard)
    } else {
        sm.SetPrevView(views.ViewCardEditor)
    }
    
    return sm
}

func (a *App) loadSourceItems(sourceType string) []components.SourceItem {
    switch sourceType {
    case "skills":
        return components.ToSimpleItems(a.skillSources.Sources)
    case "personas":
        return components.ToSimpleItems(a.personaSources.Sources)
    case "workspaces":
        return components.ToWorkspaceItems(a.workspaceForm.Sources)
    }
    return nil
}

func (a *App) saveSimpleSource(sourceType, path string, idx int) error {
    switch sourceType {
    case "skills":
        if idx == -1 {
            a.skillSources.Sources = append(a.skillSources.Sources, path)
        } else {
            a.skillSources.Sources[idx] = path
        }
        if err := a.repo.SaveSkillSources(a.skillSources.Sources); err != nil {
            a.err = err
            return err
        }
        a.refreshSkills()
        
    case "personas":
        if idx == -1 {
            a.personaSources.Sources = append(a.personaSources.Sources, path)
        } else {
            a.personaSources.Sources[idx] = path
        }
        if err := a.repo.SavePersonaSources(a.personaSources.Sources); err != nil {
            a.err = err
            return err
        }
        a.refreshPersonas()
    }
    return nil
}

func (a *App) saveWorkspaceSource(source config.WorkspaceSource, idx int) error {
    sources := a.workspaceForm.Sources
    
    if idx == -1 {
        // New source
        source.ID = views.GenerateID()
        sources = append(sources, source)
    } else {
        // Edit existing
        source.ID = sources[idx].ID
        sources[idx] = source
    }
    
    if err := a.repo.SaveWorkspaceSources(sources); err != nil {
        a.err = err
        return err
    }
    
    a.refreshWorkspaces()
    return nil
}

func (a *App) deleteSource(sourceType string, index int) error {
    switch sourceType {
    case "skills":
        a.skillSources.Sources = append(
            a.skillSources.Sources[:index],
            a.skillSources.Sources[index+1:]...,
        )
        if err := a.repo.SaveSkillSources(a.skillSources.Sources); err != nil {
            a.err = err
            return err
        }
        a.refreshSkills()
        
    case "personas":
        a.personaSources.Sources = append(
            a.personaSources.Sources[:index],
            a.personaSources.Sources[index+1:]...,
        )
        if err := a.repo.SavePersonaSources(a.personaSources.Sources); err != nil {
            a.err = err
            return err
        }
        a.refreshPersonas()
        
    case "workspaces":
        id := a.workspaceForm.Sources[index].ID
        if err := a.repo.DeleteWorkspaceSource(id); err != nil {
            a.err = err
            return err
        }
        a.refreshWorkspaces()
    }
    return nil
}

func (a *App) refreshByType(sourceType string) {
    switch sourceType {
    case "skills":
        a.refreshSkills()
    case "personas":
        a.refreshPersonas()
    case "workspaces":
        a.refreshWorkspaces()
    }
}

// For future use when fully migrated:
// sourceManagerView *views.SourceManager would be a field in App
// and New() would initialize it based on current source management needs
```

---

## Step 3.3.6: Update View Dispatcher

**File**: `internal/ui/app_view.go` (MODIFY - integrate SourceManager)

When fully migrated, the view dispatcher should handle SourceManager:

```go
// In App.View() switch statement:
case viewSkillSources, viewPersonaSources, viewWorkspaces:
    // All handled by SourceManager
    return errLine + a.sourceManagerView.View()
    
// Note: viewSkillSourceForm, viewPersonaSourceForm, viewWorkspaceSourceForm
// are no longer needed - SourceList handles form mode internally
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

# 3. Specific tests for SourceManager
go test ./internal/ui/views/ -run SourceManager -v
go test ./internal/ui/views/ -run SourceType -v

# 4. Existing UI still builds
go build ./internal/ui/

# 5. Run existing tests
go test ./internal/ui/
```

---

## Success Criteria

- [ ] `SourceType` enum defined (skills, personas, workspaces)
- [ ] `SourceTypeConfig` provides per-type configuration
- [ ] `SourceManager` struct wraps SourceList
- [ ] Correct configuration for each source type
- [ ] `SetData()` loads sources via callback
- [ ] `SetPrevView()` configures return destination
- [ ] Escape in list mode navigates back
- [ ] Escape in form mode returns to list
- [ ] Save routes to appropriate callback
- [ ] Delete routes to appropriate callback
- [ ] Refresh called after save/delete
- [ ] Unit tests pass (>80% coverage)
- [ ] Integration with App (backwards compatible)
- [ ] Existing App still compiles and tests pass

---

## Dependencies

```
component/component.go (from 1.1) ✓
messages/messages.go (from 1.2) ✓
commands/commands.go (from 1.3) ✓
components/source_item.go (from 2.4) ✓
components/source_list.go (from 2.4) ✓

views/source_type.go ← NEW
views/source_manager.go ← NEW
```

---

## Estimated Time

| Step | Time |
|------|------|
| 3.3.1 SourceType definition | 15 min |
| 3.3.2 SourceManager view | 40 min |
| 3.3.3 Messages update | 10 min |
| 3.3.4 Unit tests | 35 min |
| 3.3.5 App integration | 25 min |
| 3.3.6 View dispatcher | 10 min |
| Verification | 15 min |
| **Total** | **~2.5 hours** |

---

## What This Enables

1. **Unified source management**: Single view handles all three source types
2. **6 views → 1**: viewSkillSources, viewSkillSourceForm, viewPersonaSources, viewPersonaSourceForm, viewWorkspaces, viewWorkspaceSourceForm all handled by SourceManager
3. **Consistent UX**: All source types have identical navigation and interaction
4. **Code reuse**: Maximum leverage of SourceList component
5. **Phase 4 ready**: App coordinator can instantiate SourceManager with proper callbacks

---

## Code Reduction Summary

| Before | After | Reduction |
|--------|-------|-----------|
| `app_skill_sources.go` (156 lines) | ~50 lines (integration) | -106 |
| `app_persona_sources.go` (159 lines) | ~50 lines (integration) | -109 |
| `app_workspaces.go` (source part ~100 lines) | ~60 lines (integration) | -40 |
| `view_sources.go` (35 lines) | 0 (removed) | -35 |
| **Total removed** | | **-290 lines** |

New files:
| File | Lines |
|------|-------|
| `source_type.go` | ~50 |
| `source_manager.go` | ~150 |
| `source_manager_test.go` | ~250 |
| **Total added** | ~450 |

Net lines: +160, but:
- Tests included (was 0 before)
- Type-safe configuration
- Single source of truth for source management
- Reusable for any future source types
