# Technical Debt Analysis - Kankuro

**Date:** 2026-04-05

## Executive Summary

This document identifies technical debt in the Kankuro codebase across 10 distinct areas. The analysis covers code smells, architectural issues, and DRY violations that should be addressed to improve maintainability and testability.

---

## 1. Nolint Annotations Audit

### Summary Table

| Location | nolint | Verdict | Action |
|----------|--------|---------|--------|
| `ascii/canvas.go:74` | `mnd` | **Keep** | Bresenham's algorithm magic number is intentional |
| `app_view.go:15` | `funlen,gocognit` | **Remove** | Refactor into view renderers |
| `app_new.go:96` | `funlen` | **Remove** | Part of App struct refactoring |
| `app_update.go:23` | `funlen` | **Acceptable** | Dispatch function by design, well-structured |
| `validation.go:10` | `gocognit` | **Remove** | Simplify validation logic |
| `app_skill_sources.go:1` | `dupl` | **Remove** | Extract source sync pattern |
| `app_persona_sources.go:1` | `dupl` | **Remove** | Extract source sync pattern |
| `skills.go:1` | `dupl` | **Remove** | Extract source sync pattern |
| `personas.go:1` | `dupl` | **Remove** | Extract source sync pattern |
| `attach.go:147` | `gocognit` | **Keep** | Complex PTY event loop is inherent complexity |
| `workspace.go:32-56` | `gosec` (multiple) | **Keep** | All paths are derived from managed IDs |
| `driver.go:220` | `gosec` | **Keep** | User-defined command execution is intentional |
| `factory.go:5` | `gochecknoglobals` | **Remove** | Replace with dependency injection |
| `factory_test.go:36,45` | `reassign` | **Keep** | Test needs to reset global registry |
| `driver_registration_test.go:5` | `gochecknoinits` | **Keep** | Test mirrors main registration |
| `session/driver_registration_test.go:5` | `gochecknoinits` | **Keep** | Test mirrors main registration |
| `styles.go:5` | `gochecknoglobals` | **Keep** | Immutable style constants are acceptable |
| `app_mcp.go:69` | `modernize` | **Remove** | Already using `strings.Cut`, nolint unnecessary |
| Test files (3) | `testpackage` | **Keep** | Needed for testing private UI state |

---

## 2. Repository Interface - Interface Segregation Violation

**Location:** `internal/config/repository.go:21-53`

**Issue:** The `Repository` interface has 27 methods covering 5 distinct domains:
- ID Cards (4 methods)
- Workspaces (6 methods)
- Skills (6 methods)
- Personas (6 methods)
- MCPs (4 methods)
- Common (1 method: `Dir()`)

This violates the Interface Segregation Principle (ISP). Clients that only need skill operations shouldn't depend on workspace methods.

**Current State:**
```go
type Repository interface {
    ListIDCards() ([]IDCard, error)
    CreateIDCard(card IDCard) (IDCard, error)
    UpdateIDCard(card IDCard) (IDCard, error)
    DeleteIDCard(id string) error
    
    ListWorkspaceSources() ([]WorkspaceSource, error)
    SaveWorkspaceSources(sources []WorkspaceSource) error
    SyncWorkspacesFromSources() error
    ListWorkspaces() ([]Workspace, error)
    GetWorkspaceByPath(path string) (Workspace, bool)
    DeleteWorkspaceSource(id string) error
    
    ListSkillSources() ([]string, error)
    SaveSkillSources(paths []string) error
    SyncSkillsFromSources() error
    ListSkills() ([]Skill, error)
    UpsertSkill(skill Skill) (Skill, error)
    DeleteSkill(id string) error
    
    ListPersonaSources() ([]string, error)
    SavePersonaSources(paths []string) error
    SyncPersonasFromSources() error
    ListPersonas() ([]Persona, error)
    UpsertPersona(persona Persona) (Persona, error)
    DeletePersona(id string) error
    
    ListMCPs() ([]MCPServer, error)
    UpsertMCP(mcp MCPServer) (MCPServer, error)
    DeleteMCP(id string) error
    
    Dir() string
}
```

---

## 3. God Object: App Struct

**Location:** `internal/ui/app.go:117-197`

**Issue:** The `App` struct contains 80+ fields mixing multiple concerns:

**Current Fields Breakdown:**

| Category | Fields | Count |
|----------|--------|-------|
| Dependencies | `repo`, `session` | 2 |
| View Mode | `mode`, `focusedPane` | 2 |
| Domain Data | `cards`, `skills`, `personas`, `mcps`, `sessions`, `workspaces` | 6 |
| UI Components | `cardsList`, `sessionsList`, `switchList` | 3 |
| Card Form Inputs | `cardInputs` (5), `mcpInputs` (7), `workspaceInputs` (2), `agentsEditor` | 1 struct |
| Card Form State | `cardFocusIndex`, `selectedWorkspaceID`, `workspaceCursor`, etc. | 10+ |
| Skill/Persona Selection | `skillCursor`, `selectedSkillIDs`, `skillSearchQuery`, etc. | 8+ |
| MCP Selection | `mcpCursor`, `selectedMCPIDs`, `mcpSearchQuery`, `mcpFormPrevMode` | 4 |
| Source Management | `skillSources`, `skillSourceCursor`, `skillSourceInput`, etc. | 12 |
| Edit State | `editingCardID`, `editingSkillSourceIndex`, etc. | 5 |
| Session State | `activeSessionID`, `pendingDeleteID`, `switchSessionID` | 3 |
| Naming State | `namingSessionID`, `namingCard`, `namingInput`, etc. | 5 |
| Puppet State | `cardPuppet`, `puppetSelectorIndex`, `puppetSelectorPrev` | 3 |
| Layout | `width`, `height`, `err` | 3 |
| Help | `help` | 1 |

**Impact:** 
- Hard to test individual features
- Tight coupling between unrelated features
- Difficult to understand which state belongs to which view

---

## 4. God Function: View()

**Location:** `internal/ui/app_view.go:16-156`

**Issue:** The `View()` method is a 156-line switch statement handling 15+ view modes. Each case contains significant inline rendering logic.

**View Modes Handled:**
1. `viewDeleteConfirm`
2. `viewCardForm`
3. `viewSkillSources`
4. `viewSkillSourceForm`
5. `viewPuppetSelector`
6. `viewPersonaSources`
7. `viewPersonaSourceForm`
8. `viewWorkspaces`
9. `viewWorkspaceSourceForm`
10. `viewWorkspacePicker`
11. `viewAgentsEditor`
12. `viewAgentsPathInput`
13. `viewMCPForm`
14. `viewSwitchAgent`
15. `viewNaming`
16. `viewDashboard`

**Note:** A TODO comment already acknowledges this debt (line 13).

---

## 5. DRY Violation: Source Sync Pattern

**Locations:**
- `internal/config/skill_sources.go:30-86`
- `internal/config/persona_sources.go:29-70`
- `internal/config/workspaces.go:74-160`

**Issue:** All three sync functions follow nearly identical patterns:

```go
func (r *FileRepository) SyncXFromSources() error {
    sources, err := r.ListXSources()
    if err != nil {
        return err
    }
    var items []X
    for _, source := range sources {
        expanded, err := expandPath(source)
        if err != nil {
            return err
        }
        entries, err := os.ReadDir(expanded)
        if err != nil {
            return err
        }
        for _, entry := range entries {
            // Filter hidden files
            // Build item from directory/file
            items = append(items, item)
        }
    }
    return r.WriteXFile(XFile{SchemaVersion: SchemaVersion, Items: items})
}
```

---

## 6. DRY Violation: Upsert Pattern

**Locations:**
- `internal/config/id_cards.go:53-81`
- `internal/config/skills.go:25-51`
- `internal/config/personas.go:25-51`
- `internal/config/mcps.go:24-52`

**Issue:** All upsert implementations share identical logic:

```go
func (r *FileRepository) UpsertX(item X) (X, error) {
    if item.Name == "" {
        return X{}, errors.New("name is required")
    }
    if item.ID == "" {
        item.ID = uuid.NewString()
    }
    f, err := r.LoadXFile()
    if err != nil {
        return X{}, err
    }
    updated := false
    for i := range f.Items {
        if f.Items[i].ID == item.ID {
            f.Items[i] = item
            updated = true
            break
        }
    }
    if !updated {
        f.Items = append(f.Items, item)
    }
    if err = r.WriteXFile(f); err != nil {
        return X{}, err
    }
    return item, nil
}
```

---

## 7. DRY Violation: Filter Pattern

**Location:** `internal/ui/app_card_form.go:14-56`

**Issue:** Four nearly identical filter functions:

```go
func (a *App) filteredSkills() []config.Skill { ... }
func (a *App) filteredPersonas() []config.Persona { ... }
func (a *App) filteredMCPs() []config.MCPServer { ... }
```

Plus `filteredWorkspaces()` in `app_workspaces.go:28-41`.

---

## 8. DRY Violation: Window Calculation

**Location:** `internal/ui/app_card_form.go:582-631`

**Issue:** Four identical window calculation functions:

```go
func (a *App) skillWindow() (int, int) {
    total := len(a.filteredSkills())
    height := a.skillListHeight()
    if total <= height {
        return 0, total
    }
    if a.skillCursor < height/2 {
        return 0, height
    }
    if a.skillCursor > total-height/2 {
        return total - height, total
    }
    return a.skillCursor - height/2, a.skillCursor + height/2
}

func (a *App) personaWindow() (int, int) { /* same logic */ }
func (a *App) mcpWindow() (int, int) { /* same logic */ }
// workspaceWindow in app_workspaces.go - same logic
```

---

## 9. Duplicate Logic: Driver Sync

**Location:** `internal/session/lifecycle.go`

**Issue:** `SwitchAgent()` (lines 62-85) and `startSession()` (lines 157-169) share almost identical driver file synchronization:

```go
// Both blocks contain:
d := driver.NewCLIDriver(card.Driver)
if errAgentsMD := d.SetAgentsMD(card, personas, mcps, finalWD, false); errAgentsMD != nil {
    return errAgentsMD
}
if errSkills := d.SetSkills(skills, finalWD); errSkills != nil {
    return errSkills
}
if errPersonas := d.SetPersonas(personas, finalWD); errPersonas != nil {
    return errPersonas
}
if errMCPs := d.SelectMCPs(mcps, finalWD); errMCPs != nil {
    return errMCPs
}
```

---

## 10. Global Mutable State: DriverConstructors

**Location:** `internal/driver/factory.go:6`

**Issue:**
```go
var DriverConstructors = map[string]func() AgentDriver{}
```

This is:
- A mutable global variable
- Modified during init by `RegisterDriver()`
- Makes testing harder (must reset between tests)
- Implicit dependency (not injected)

**Current Usage:**
- `main.go:19` - Registers claude driver
- `factory.go:18` - Read to create drivers
- `factory_test.go:37,46` - Reset in tests

---

## 11. Complex Validation Logic

**Location:** `internal/config/validation.go:11-73`

**Issue:** `ValidateCard()` has high cognitive complexity (14 according to gocognit) mixing:
- Field validation
- Workspace reference checking
- Skill reference checking
- Persona reference checking
- MCP reference checking

The function builds multiple lookup maps and iterates through reference lists.

---

## 12. Large Constants Block

**Location:** `internal/ui/app.go:40-115`

**Issue:** 75 lines of constants with many magic numbers:
- Layout dimensions (widths, heights, paddings)
- Input counts
- Focus indices
- Key bindings

Could be grouped into semantic structs:
```go
type LayoutConstants struct {
    CardListMinWidth int
    SessionListWidth int
    // ...
}
```

---

## 13. Types File Organization

**Location:** `internal/config/types.go`

**Issue:** Contains type definitions for all domain objects (IDCard, Workspace, Skill, Persona, MCPServer) plus all file types.

The TODO on line 3 suggests: "Split this file declaring each type in the file it's used"

---

## Package Metrics Summary

| Package | Lines of Code | Largest File | Issue |
|---------|---------------|--------------|-------|
| `internal/ui` | 3,786 | `app_card_form.go` (647) | God object + god function |
| `internal/config` | 1,470 | `workspaces.go` (210) | DRY violations |
| `internal/driver` | 1,176 | `driver.go` (587) | Global state |
| `internal/session` | 1,287 | `lifecycle.go` (319) | Duplicate logic |
| `internal/ascii` | 268 | `canvas.go` (127) | Clean |

**Total:** ~8,000 lines across codebase

---

## Test Coverage Notes

- 3 UI test files use `package ui` instead of `package ui_test` to access private state
- All other tests correctly use `*_test` package convention
- Test coverage is good, but relies on white-box testing for UI

---

## Architecture Assessment

**Positive aspects:**
- No circular dependencies detected
- Clean layered architecture: `config → driver → session → ui`
- Clear separation between data/presentation/process layers
- Good use of interfaces elsewhere (AgentDriver)

**Areas for improvement:**
- Interface segregation
- State management in UI
- Code reuse patterns
- Dependency injection over globals

---

## Execution Order & Parallel Execution

### Task Dependencies Graph

```
task-1-2.md (Repository Split + Source Sync)
    ↓
    ├── task-3-8.md (Upsert + Validation) [depends on: source sync changes]
    │
    ├── task-6-7.md (Driver Sync + DI) [depends on: source sync concept]
    │
    └── task-9-10.md (Utils + Cleanup) [independent]

task-5.md (View Split) [depends on: App struct field naming]
    ↓
task-4.md (App Split) [must come after View Split]
```

### Parallel Execution Analysis

| Phase | Tasks | Can Run in Parallel? | Reason |
|-------|-------|---------------------|--------|
| **Phase 1** | task-1-2 | No - must complete first | Foundational changes to config package |
| **Phase 2a** | task-6-7, task-9-10 | **YES** - totally independent | Different packages (session/driver vs ui) |
| **Phase 2b** | task-3-8 | **YES** with Phase 2a | Only touches config/, no overlap |
| **Phase 3** | task-5 | No - wait for task-2a completion | Needs stable App struct field names |
| **Phase 4** | task-4 | No - must be last | Largest refactor, needs all others stable |

### Detailed Analysis

#### Can Run in Parallel (No Merge Conflicts)

1. **task-6-7 + task-9-10 + task-3-8 (Phase 2)**
   - `task-6-7`: `internal/session/*.go`, `internal/driver/*.go`, `cmd/kankuro/main.go`
   - `task-9-10`: `internal/ui/*.go` (utils, filters), `internal/config/types.go`
   - `task-3-8`: `internal/config/storage.go`, `internal/config/validation*.go`, upsert functions
   - **No file overlap** → Safe to run on separate branches

#### Cannot Run in Parallel (High Rebase Risk)

1. **task-1-2 must complete first**
   - Creates `internal/config/sync.go` (new file)
   - Modifies `internal/config/repository.go`
   - Other tasks may reference the new sync pattern or interfaces

2. **task-5 must wait for task-6-7 and task-9-10**
   - task-6-7 may change `internal/session/lifecycle.go` method signatures
   - task-9-10 modifies `internal/ui/app_card_form.go`
   - task-5 modifies same files: `internal/ui/app_view.go`, `internal/ui/app_card_form.go`
   - **High conflict risk** on `app_card_form.go`

3. **task-4 must run last**
   - Changes field names across ALL `internal/ui/*.go` files
   - Any earlier task touching UI files would need rebase
   - `app_new.go`, `app_view.go`, `app_card_form.go` would have massive conflicts

### Recommended Execution Strategy

```
Week 1:
  Day 1: task-1-2 (complete) → merge to main
  Day 2-3: Start in parallel on separate branches:
           - Branch A: task-6-7
           - Branch B: task-9-10  
           - Branch C: task-3-8
  Day 4: Merge A, B, C (resolve any minor conflicts)

Week 2:
  Day 1: task-5 (complete) → merge to main
  Day 2-3: task-4 (complete) → merge to main
```

### Merge Conflict Risk Assessment

| Task Pair | Overlap Files | Conflict Risk |
|-----------|---------------|---------------|
| 1-2 vs 6-7 | None | **None** |
| 1-2 vs 9-10 | `config/types.go` | **Low** (different lines) |
| 1-2 vs 3-8 | `config/sync.go` concept | **Medium** (same pattern) |
| 6-7 vs 9-10 | None | **None** |
| 6-7 vs 3-8 | None | **None** |
| 9-10 vs 5 | `app_card_form.go` | **High** (same functions) |
| 5 vs 4 | All UI files | **Certain** (field renames) |

### Git Branch Strategy

```bash
# Phase 1: Foundation
git checkout -b refactor/repository-split
# ... complete task-1-2 ...
git merge main

# Phase 2: Independent branches
git checkout -b refactor/di-registry          # task-6-7
git checkout -b refactor/utils-cleanup        # task-9-10
git checkout -b refactor/upsert-validation    # task-3-8

# After testing each, merge:
git checkout main && git merge refactor/di-registry
git checkout main && git merge refactor/utils-cleanup
git checkout main && git merge refactor/upsert-validation

# Phase 3: View refactor
git checkout -b refactor/view-split           # task-5
git checkout main && git merge refactor/view-split

# Phase 4: Final refactor
git checkout -b refactor/app-state            # task-4
git checkout main && git merge refactor/app-state
```