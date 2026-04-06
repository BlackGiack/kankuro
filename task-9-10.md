# Task 9+10: Extract Filter/Window Utils + Minor nolint Cleanup

**Priority:** Low  
**Effort:** Low  
**Estimated Time:** 1-2 hours

## Overview

This task handles quick wins and minor cleanups:
1. Extract duplicate filter and window calculation patterns
2. Fix minor nolint issues that are quick wins

---

## Part 1: Extract Filter/Window Utilities

### Current State

**File:** `internal/ui/app_card_form.go`

Four filter functions:
```go
func (a *App) filteredSkills() []config.Skill {
    if a.cardForm.SkillSearchQuery == "" {
        return a.skills
    }
    var filtered []config.Skill
    query := strings.ToLower(a.cardForm.SkillSearchQuery)
    for _, s := range a.skills {
        if strings.Contains(strings.ToLower(s.Name), query) ||
           strings.Contains(strings.ToLower(s.Description), query) {
            filtered = append(filtered, s)
        }
    }
    return filtered
}

func (a *App) filteredPersonas() []config.Persona { /* similar */ }
func (a *App) filteredMCPs() []config.MCPServer { /* similar */ }
```

Plus `filteredWorkspaces()` in `internal/ui/app_workspaces.go`.

Four window calculation functions:
```go
func (a *App) skillWindow() (int, int) {
    total := len(a.filteredSkills())
    height := a.skillListHeight()
    if total <= height {
        return 0, total
    }
    if a.cardForm.SkillCursor < height/2 {
        return 0, height
    }
    if a.cardForm.SkillCursor > total-height/2 {
        return total - height, total
    }
    return a.cardForm.SkillCursor - height/2, a.cardForm.SkillCursor + height/2
}

func (a *App) personaWindow() (int, int) { /* identical logic */ }
func (a *App) mcpWindow() (int, int) { /* identical logic */ }
// workspaceWindow in app_workspaces.go with same logic
```

### Proposed Solution

#### Step 1: Create UI Utilities File

Create `internal/ui/utils.go`:

```go
package ui

import "strings"

// FilterByQuery filters items by a case-insensitive query match.
func FilterByQuery[T any](items []T, query string, matcher func(T, string) bool) []T {
    if query == "" {
        return items
    }
    var filtered []T
    lowerQuery := strings.ToLower(query)
    for _, item := range items {
        if matcher(item, lowerQuery) {
            filtered = append(filtered, item)
        }
    }
    return filtered
}

// CalculateWindow returns the visible window (start, end) for a list.
func CalculateWindow(cursor, total, viewHeight int) (start, end int) {
    if total <= viewHeight {
        return 0, total
    }
    if cursor < viewHeight/2 {
        return 0, viewHeight
    }
    if cursor > total-viewHeight/2 {
        return total - viewHeight, total
    }
    start = cursor - viewHeight/2
    return start, start + viewHeight
}
```

#### Step 2: Update Filter Functions

Update `internal/ui/app_card_form.go`:

```go
func (a *App) filteredSkills() []config.Skill {
    return FilterByQuery(a.skills, a.cardForm.SkillSearchQuery,
        func(s config.Skill, q string) bool {
            return strings.Contains(strings.ToLower(s.Name), q) ||
                strings.Contains(strings.ToLower(s.Description), q)
        })
}

func (a *App) filteredPersonas() []config.Persona {
    return FilterByQuery(a.personas, a.cardForm.PersonaSearchQuery,
        func(p config.Persona, q string) bool {
            return strings.Contains(strings.ToLower(p.Name), q)
        })
}

func (a *App) filteredMCPs() []config.MCPServer {
    return FilterByQuery(a.mcps, a.cardForm.MCPSearchQuery,
        func(m config.MCPServer, q string) bool {
            return strings.Contains(strings.ToLower(m.Name), q) ||
                strings.Contains(strings.ToLower(m.Description), q)
        })
}
```

Update `internal/ui/app_workspaces.go`:

```go
func (a *App) filteredWorkspaces() []config.Workspace {
    return FilterByQuery(a.workspaces, a.workspaces_.SearchQuery,
        func(w config.Workspace, q string) bool {
            return strings.Contains(strings.ToLower(w.Name), q) ||
                strings.Contains(strings.ToLower(w.Path), q)
        })
}
```

#### Step 3: Update Window Functions

Update `internal/ui/app_card_form.go`:

```go
func (a *App) skillWindow() (int, int) {
    return CalculateWindow(a.cardForm.SkillCursor, len(a.filteredSkills()), a.skillListHeight())
}

func (a *App) personaWindow() (int, int) {
    return CalculateWindow(a.cardForm.PersonaCursor, len(a.filteredPersonas()), personaListHeight)
}

func (a *App) mcpWindow() (int, int) {
    return CalculateWindow(a.cardForm.MCPCursor, len(a.filteredMCPs()), mcpListHeight)
}
```

Update `internal/ui/app_workspaces.go`:

```go
func (a *App) workspaceWindow() (int, int) {
    return CalculateWindow(a.workspaceCursor, len(a.filteredWorkspaces()), a.workspaceListHeight())
}
```

---

## Part 2: Minor nolint Cleanup

### Issue 1: Unnecessary modernize nolint

**File:** `internal/ui/app_mcp.go:69`

Current:
```go
//nolint:modernize // stick to Split for broader compatibility
for _, pair := range strings.Split(input, ",") {
```

Analysis: The linter likely complained about using `strings.Split` instead of `strings.Cut` or similar. However:
- This code is already using `strings.Cut` inside the loop (`strings.Cut(pair, "=")`)
- The `strings.Split` for comma-separated values is appropriate here
- If the linter still complains, add a specific explanation

Fix: Remove the `//nolint:modernize` comment as it's not needed, or replace with a more specific one:

```go
// Split is appropriate here for parsing comma-separated key=value pairs
for _, pair := range strings.Split(input, ",") {
    if k, v, ok := strings.Cut(pair, "="); ok {
        result[strings.TrimSpace(k)] = strings.TrimSpace(v)
    }
}
```

### Issue 2: Remove TODO in types.go

**File:** `internal/config/types.go:3-4`

Current:
```go
// TODO: Split this file declaring each type in the file it's used

import (
```

Since we've now split the Repository interface (Task 1), this TODO is less relevant. The types are well-organized in a central file, which is a valid pattern.

Fix: Remove the TODO comment:

```go
package config

import (
    "time"
)
```

---

## Additional Quick Wins

### Consolidate Magic Numbers

Some magic numbers in `internal/ui/app.go` could be grouped:

```go
const (
    // Layout dimensions
    cardListMinWidth  = 20
    cardListMinHeight = 10
    cardsListPadding  = 34
    cardsListHeight   = 4
    
    sessionListWidth   = 32
    sessionsListHeight = 4
    switchListMinWidth = 40
    
    // Input dimensions
    inputWidthDefault = 60
    editorWidth       = 80
    editorHeight      = 20
    
    // List dimensions
    defaultSkillListHeight = 10
    minSkillListHeight     = 3
    
    // ... etc
)
```

This is optional - the constants are already well-named.

---

## Verification Steps

```bash
# Build
go build ./...

# Test UI package
go test ./internal/ui/... -v

# Run linter to check nolint is resolved
golangci-lint run ./...

# Manual testing of filtered lists:
# 1. Launch app
# 2. Press 'n' to create new card
# 3. Navigate to Skills list
# 4. Type in search box - verify filtering works
# 5. Navigate through list with arrows - verify window scrolling works
# 6. Same for Personas and MCPs
# 7. Test workspace picker filtering
```

---

## Files Modified

| File | Changes |
|------|---------|
| `internal/ui/utils.go` | **NEW** - FilterByQuery, CalculateWindow |
| `internal/ui/app_card_form.go` | Use utility functions for filter/window |
| `internal/ui/app_workspaces.go` | Use utility functions |
| `internal/ui/app_mcp.go` | Remove unnecessary nolint |
| `internal/config/types.go` | Remove TODO comment |

---

## Benefits

### Part 1 (Utilities)
- Eliminates ~80 lines of duplicate code
- Single source of truth for filtering logic
- Single source of truth for window calculation
- Easier to modify behavior (change in one place)

### Part 2 (Cleanup)
- Removes misleading TODO comment
- Removes unnecessary nolint directive
- Cleaner codebase with less noise

---

## Summary

This task is straightforward with minimal risk:
- Utilities are simple, well-tested functions
- No behavioral changes, just refactoring
- Removals are safe (TODO, nolint comments)
- Can be completed in a single session