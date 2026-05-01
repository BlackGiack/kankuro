# Technical Debt Report: internal/ui Package

## Executive Summary

The `internal/ui` package exhibits significant architectural debt, primarily centered around a monolithic `App` struct that violates separation of concerns. The codebase lacks true component abstraction, making it difficult to maintain, test, and extend. Sizing conventions are predominantly fixed rather than relative, which limits responsiveness across different terminal dimensions.

---

## Priority 1: Architectural Debt

### 1.1 Monolithic God Struct

**Location**: `app.go:116-155`

The `App` struct serves as a god object containing:
- Repository and session manager references
- 17+ view modes (viewMode enum)
- 6 data slice collections (cards, skills, personas, mcps, sessions, workspaces)
- 3 list models (cardsList, sessionsList, switchList)
- 6 form state structs
- Input arrays, IDE config, help model
- UI dimensions (width, height, err)

**Problems**:
- Violates Single Responsibility Principle
- Makes testing impossible without full mock setup
- Any UI change requires modifying the central struct
- No isolation between features

**Recommendation**: Decompose into focused components:
- `Dashboard` component (owns card/session lists)
- `CardForm` component (owns form inputs and validation)
- `SessionPreview` component
- `SourceManager` component (for skills/personas/workspaces)
- Use dependency injection for shared services

### 1.2 Missing Component Hierarchy

**Location**: All `view_*.go` and `app_*.go` files

Every view function is a method on `*App`, directly accessing and mutating global state. There are no reusable component types, no component lifecycle, and no component composition.

**Problems**:
- No way to compose smaller UI elements
- Views cannot be tested in isolation
- No prop passing pattern - everything accesses App directly
- Violates encapsulation

**Recommendation**: Introduce component types following Bubble Tea patterns:
```go
type CardListComponent struct {
    list     list.Model
    cards    []config.IDCard
    onSelect func(config.IDCard)
    // ... own state
}

func (c *CardListComponent) Update(msg tea.Msg) tea.Cmd
func (c *CardListComponent) View() string
```

### 1.3 State Machine Without Formalism

**Location**: `app_update.go:24-86`

The view mode switching uses a manual switch statement with 17 cases. Each view's update logic is scattered across files (app_dashboard.go, app_card_form.go, etc.) without formal state transitions.

**Problems**:
- No enforced state machine model
- View transitions can happen anywhere (e.g., `a.mode = viewCardForm` appears in 15+ locations)
- No transition guards or validation
- Difficult to track state flow

**Recommendation**: Implement a formal state machine pattern or finite state machine library.

### 1.4 Delegate Definitions Inline

**Location**: `app_lists.go:16-42`

`cardDelegate` and session list items are defined inline as struct types rather than as proper components:

```go
type cardDelegate struct {
    height   int
    skills   []config.Skill
    personas []config.Persona
}
```

**Problems**:
- Not reusable
- Tightly coupled to App's data
- Missing standard component interface

### 1.5 Massive Update Dispatchers

**Location**: `app_dashboard.go:14-87`, `app_card_form.go:113-128`

Update functions are god methods handling all key events for entire views. `updateDashboard` is 87 lines with nested conditionals.

**Recommendation**: Break into focused handlers per input type or action.

### 1.6 Scattered State in app_state.go

**Location**: `app_state.go`

While state structs exist (`CardFormState`, `SessionViewState`, etc.), they're tightly coupled to App and don't own their own behavior. They're effectively data containers without methods.

**Recommendation**: Move view-specific logic into state structs (e.g., `CardFormState.Update()`).

---

## Priority 2: Sizing Debt

### 2.1 Fixed Size Constants

**Location**: `app.go:47-114`

The following fixed constants limit responsiveness:

| Constant | Value | Usage | Should Be Relative |
|----------|-------|-------|-------------------|
| `cardListMinWidth` | 46 | Cards list minimum | Keep as min, OK |
| `cardItemWidth` | 42 | Card item | Relative to list width |
| `infoBoxWidth` | 30 | Card info section | Relative to item width |
| `infoBoxHeight` | 4 | Card info section | Fixed (OK for 4 lines) |
| `faceBoxWidth` | 11 | Puppet face | Fixed (OK for ASCII art) |
| `faceBoxHeight` | 5 | Puppet face | Fixed (OK for ASCII art) |
| `cardDelegateHeight` | 8 | List item height | Fixed (OK for consistent rows) |
| `sessionsDelegateHeight` | 4 | List item height | Fixed (OK) |
| `inputWidthDefault` | 60 | Form inputs | Relative to screen width |
| `editorWidth` | 80 | Text area | Relative to screen width |
| `editorHeight` | 20 | Text area | Relative to screen height |
| `truncateLength` | 28 | Text truncation | Relative to container |
| `pathTruncateLimit` | 25 | Path display | Relative to container |

### 2.2 Non-Responsive Editors

**Location**: `app_new.go:156-159`

```go
a.cardForm.Editor.SetWidth(editorWidth)  // Fixed 80
a.cardForm.Editor.SetHeight(editorHeight) // Fixed 20
```

**Problem**: Editors don't adapt to terminal size. On small terminals, they overflow. On large terminals, they waste space.

**Recommendation**: Calculate editor dimensions as percentages:
```go
a.cardForm.Editor.SetWidth(max(40, a.width*8/10))
a.cardForm.Editor.SetHeight(max(10, a.height/3))
```

### 2.3 Good Pattern: Dashboard Columns

**Location**: `app_lists.go:44-68`

Dashboard resize uses relative sizing correctly:
```go
sessionListWidth := a.width / 4              // 25%
previewWidth := (a.width/4)*2 - borderPadding // 50%
cardsWidth := a.width - totalSideWidths       // Remaining
```

This pattern should be applied more broadly.

### 2.4 Inconsistent Sizing in Views

**Location**: `view_ide.go:28`

```go
Width(min(a.width-ideCurrentBoxWidthOffset, inputWidthDefault+ideCurrentBoxWidthOffset))
```

This correctly uses screen width but caps at a fixed value. Consider:
```go
Width(min(a.width*9/10, 80)) // 90% of screen, max 80
```

---

## Other Technical Debt

### 3.1 Duplicate Code

**Location**: `app_skill_sources.go` and `app_persona_sources.go`

Both files are nearly identical (156 vs 159 lines), flagged with `//nolint:dupl`. 

**Recommendation**: Extract generic source manager component.

### 3.2 Magic Numbers

**Location**: Multiple files with `//nolint:mnd` comments

Example: `app_view.go:59`:
```go
fmt.Print("\x1b[=0u") //nolint:forbidigo 
```

Many numeric constants lack semantic naming.

### 3.3 Inline Style Construction

**Location**: Various render functions

Styles are often built inline:
```go
lipgloss.NewStyle().Foreground(highlightColor).Bold(true).Render(line)
```

**Recommendation**: Pre-define reusable styles in `styles.go`.

### 3.4 Partial Generic Abstraction

**Location**: `app_render.go:90-141`

`renderRegistryList` exists as a generic helper but isn't used consistently. Some list renders use it, others don't.

---

## Recommendations Priority Order

1. **Decompose App struct** into component types
2. **Apply relative sizing** to editors and input widths
3. **Extract reusable styles** to styles.go
4. **Genericize source managers** to eliminate duplicate code
5. **Formalize state machine** for view modes

---

## Metrics

| Metric | Value |
|--------|-------|
| Total files | 40 |
| App struct fields | 22+ |
| View modes | 17 |
| Fixed size constants | 20+ |
| Duplicate code blocks | 2 (acknowledged) |