# Task 3+8: Extract Upsert Pattern + Simplify Validation

**Priority:** Medium  
**Effort:** Medium  
**Estimated Time:** 2-3 hours

## Overview

This task combines two related refactorings:
1. Extract the duplicated upsert pattern across domain types
2. Simplify the complex validation logic in `ValidateCard()`

---

## Part 1: Extract Upsert Pattern

### Current State

Four nearly identical upsert implementations exist:

| File | Function | Lines |
|------|----------|-------|
| `internal/config/id_cards.go` | `UpdateIDCard()` | 53-81 |
| `internal/config/skills.go` | `UpsertSkill()` | 25-51 |
| `internal/config/personas.go` | `UpsertPersona()` | 25-51 |
| `internal/config/mcps.go` | `UpsertMCP()` | 24-52 |

### Pattern Analysis

All implementations share this structure:

```go
func (r *FileRepository) UpsertX(item X) (X, error) {
    // 1. Validate required field
    if item.Name == "" { // or other validation
        return X{}, errors.New("name is required")
    }
    
    // 2. Generate ID if missing
    if item.ID == "" {
        item.ID = uuid.NewString()
    }
    
    // 3. Load existing items
    f, err := r.LoadXFile()
    if err != nil {
        return X{}, err
    }
    
    // 4. Find and update, or append
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
    
    // 5. Write back
    if err = r.WriteXFile(f); err != nil {
        return X{}, err
    }
    
    return item, nil
}
```

### Proposed Solution

#### Step 1: Create Generic Upsert Helper

Add to `internal/config/storage.go`:

```go
package config

import "github.com/google/uuid"

// IDSetter defines how to get and set an ID on an item.
type IDSetter[T any] interface {
    GetID() string
    SetID(id string)
}

// UpsertItem adds or updates an item in a slice by ID.
// Returns the modified slice and the item (with ID set if new).
func UpsertItem[T IDSetter[T]](items []T, item T) ([]T, T) {
    if item.GetID() == "" {
        item.SetID(uuid.NewString())
    }
    for i := range items {
        if items[i].GetID() == item.GetID() {
            items[i] = item
            return items, item
        }
    }
    return append(items, item), item
}
```

#### Step 2: Implement IDSetter for Domain Types

Add methods to each domain type. For existing types in `internal/config/types.go`:

```go
// IDCard
func (c *IDCard) GetID() string { return c.ID }
func (c *IDCard) SetID(id string) { c.ID = id }

// Skill
func (s *Skill) GetID() string { return s.ID }
func (s *Skill) SetID(id string) { s.ID = id }

// Persona
func (p *Persona) GetID() string { return p.ID }
func (p *Persona) SetID(id string) { p.ID = id }

// MCPServer
func (m *MCPServer) GetID() string { return m.ID }
func (m *MCPServer) SetID(id string) { m.ID = id }
```

#### Step 3: Refactor Upsert Functions

Update each upsert function to use the helper:

**internal/config/skills.go:**

```go
func (r *FileRepository) UpsertSkill(skill Skill) (Skill, error) {
    if skill.Name == "" {
        return Skill{}, errors.New("skill name is required")
    }
    
    f, err := r.LoadSkillsFile()
    if err != nil {
        return Skill{}, err
    }
    
    f.Skills, skill = UpsertItem[Skill](f.Skills, &skill)
    
    if err = r.WriteSkillsFile(f); err != nil {
        return Skill{}, err
    }
    return skill, nil
}
```

Wait - there's an issue. The helper expects `T` but we're passing `*Skill`. Let's revise.

#### Alternative: Simpler Generic Helper

Use a functional approach instead:

```go
// internal/config/storage.go

// UpsertInSlice adds or updates an item in a slice.
// getID and setID functions provide ID access.
func UpsertInSlice[T any](
    items []T,
    item T,
    getID func(T) string,
    setID func(*T, string),
) ([]T, T) {
    if getID(item) == "" {
        setID(&item, uuid.NewString())
    }
    for i := range items {
        if getID(items[i]) == getID(item) {
            items[i] = item
            return items, item
        }
    }
    return append(items, item), item
}
```

Usage:

```go
// internal/config/skills.go
func (r *FileRepository) UpsertSkill(skill Skill) (Skill, error) {
    if skill.Name == "" {
        return Skill{}, errors.New("skill name is required")
    }
    
    f, err := r.LoadSkillsFile()
    if err != nil {
        return Skill{}, err
    }
    
    f.Skills, skill = UpsertInSlice(
        f.Skills,
        skill,
        func(s Skill) string { return s.ID },
        func(s *Skill, id string) { s.ID = id },
    )
    
    if err = r.WriteSkillsFile(f); err != nil {
        return Skill{}, err
    }
    return skill, nil
}
```

#### Apply to All Upserts

**internal/config/personas.go:**

```go
func (r *FileRepository) UpsertPersona(persona Persona) (Persona, error) {
    if persona.Name == "" {
        return Persona{}, errors.New("persona name is required")
    }
    
    f, err := r.LoadPersonasFile()
    if err != nil {
        return Persona{}, err
    }
    
    f.Personas, persona = UpsertInSlice(
        f.Personas,
        persona,
        func(p Persona) string { return p.ID },
        func(p *Persona, id string) { p.ID = id },
    )
    
    if err = r.WritePersonasFile(f); err != nil {
        return Persona{}, err
    }
    return persona, nil
}
```

**internal/config/mcps.go:**

```go
func (r *FileRepository) UpsertMCP(mcp MCPServer) (MCPServer, error) {
    if mcp.Name == "" {
        return MCPServer{}, errors.New("mcp name is required")
    }
    
    f, err := r.LoadMCPsFile()
    if err != nil {
        return MCPServer{}, err
    }
    
    f.MCPs, mcp = UpsertInSlice(
        f.MCPs,
        mcp,
        func(m MCPServer) string { return m.ID },
        func(m *MCPServer, id string) { m.ID = id },
    )
    
    if err = r.WriteMCPsFile(f); err != nil {
        return MCPServer{}, err
    }
    return mcp, nil
}
```

**internal/config/id_cards.go:**

Note: IDCard uses Create/Update rather than Upsert. The UpdateIDCard can use the helper:

```go
func (r *FileRepository) UpdateIDCard(card IDCard) (IDCard, error) {
    if card.Name == "" {
        return IDCard{}, errors.New("id card name is required")
    }
    if card.ID == "" {
        return IDCard{}, errors.New("id card id is required for update")
    }
    if err := r.ValidateCard(card); err != nil {
        return IDCard{}, err
    }
    card.UpdatedAt = time.Now().UTC()
    
    f, err := r.LoadIDCardsFile()
    if err != nil {
        return IDCard{}, err
    }
    
    f.IDCards, card = UpsertInSlice(
        f.IDCards,
        card,
        func(c IDCard) string { return c.ID },
        func(c *IDCard, id string) { c.ID = id }, // won't be called since ID required
    )
    
    if err = r.WriteIDCardsFile(f); err != nil {
        return IDCard{}, err
    }
    return card, nil
}
```

---

## Part 2: Simplify Validation Logic

### Current State

**File:** `internal/config/validation.go`

```go
//nolint:gocognit // validation logic is inherently complex
func (r *FileRepository) ValidateCard(card IDCard) error {
    // 62 lines of mixed validation
}
```

Cognitive complexity is high due to:
- Multiple sequential validations
- Multiple map lookups
- Nested loops for reference checking

### Proposed Solution

#### Step 1: Extract Validation Helpers

Create `internal/config/validation_helpers.go`:

```go
package config

import (
    "errors"
    "fmt"
)

// validateRequiredFields checks that mandatory card fields are set.
func validateCardFields(card IDCard) error {
    if card.Name == "" {
        return errors.New("id card name is required")
    }
    if card.Driver == "" {
        return errors.New("driver is required")
    }
    if card.WorkspaceID == "" {
        return errors.New("workspace is required")
    }
    return nil
}

// idSet builds a set of IDs from items.
func idSet[T any](items []T, getID func(T) string) map[string]struct{} {
    set := make(map[string]struct{})
    for _, item := range items {
        set[getID(item)] = struct{}{}
    }
    return set
}

// validateRefs checks that all IDs in refs exist in the set.
func validateRefs(refs []string, set map[string]struct{}, typeName string) error {
    for _, id := range refs {
        if _, ok := set[id]; !ok {
            return fmt.Errorf("%s reference not found: %s", typeName, id)
        }
    }
    return nil
}
```

#### Step 2: Refactor ValidateCard

Update `internal/config/validation.go`:

```go
package config

// ValidateCard enforces required fields and checks skill/MCP references.
func (r *FileRepository) ValidateCard(card IDCard) error {
    if err := validateCardFields(card); err != nil {
        return err
    }
    
    if err := r.validateWorkspaceRef(card.WorkspaceID); err != nil {
        return err
    }
    
    if err := r.validateSkillRefs(card.SkillIDs); err != nil {
        return err
    }
    
    if err := r.validatePersonaRefs(card.PersonaIDs); err != nil {
        return err
    }
    
    return r.validateMCPRefs(card.MCPIDs)
}

func (r *FileRepository) validateWorkspaceRef(workspaceID string) error {
    workspaces, err := r.ListWorkspaces()
    if err != nil {
        return err
    }
    set := idSet(workspaces, func(w Workspace) string { return w.ID })
    if _, ok := set[workspaceID]; !ok {
        return fmt.Errorf("workspace reference not found: %s", workspaceID)
    }
    return nil
}

func (r *FileRepository) validateSkillRefs(skillIDs []string) error {
    if len(skillIDs) == 0 {
        return nil
    }
    skills, err := r.ListSkills()
    if err != nil {
        return err
    }
    return validateRefs(skillIDs, idSet(skills, func(s Skill) string { return s.ID }), "skill")
}

func (r *FileRepository) validatePersonaRefs(personaIDs []string) error {
    if len(personaIDs) == 0 {
        return nil
    }
    personas, err := r.ListPersonas()
    if err != nil {
        return err
    }
    return validateRefs(personaIDs, idSet(personas, func(p Persona) string { return p.ID }), "persona")
}

func (r *FileRepository) validateMCPRefs(mcpIDs []string) error {
    if len(mcpIDs) == 0 {
        return nil
    }
    mcps, err := r.ListMCPs()
    if err != nil {
        return err
    }
    return validateRefs(mcpIDs, idSet(mcps, func(m MCPServer) string { return m.ID }), "mcp")
}
```

#### Step 3: Remove nolint Comment

Remove the `//nolint:gocognit` comment as the function is now simplified.

---

## Verification Steps

```bash
# Build
go build ./...

# Run config tests
go test ./internal/config/... -v

# Test specific validation scenarios
go test ./internal/config/... -v -run TestValidateCard

# Verify upsert behavior
go test ./internal/config/... -v -run TestUpsert
```

---

## Files Modified

| File | Changes |
|------|---------|
| `internal/config/storage.go` | Add UpsertInSlice helper |
| `internal/config/validation_helpers.go` | **NEW** - Validation helpers |
| `internal/config/validation.go` | Refactor to use helpers, remove nolint |
| `internal/config/skills.go` | Use UpsertInSlice |
| `internal/config/personas.go` | Use UpsertInSlice |
| `internal/config/mcps.go` | Use UpsertInSlice |
| `internal/config/id_cards.go` | Use UpsertInSlice in UpdateIDCard |

---

## Benefits

### Part 1 (Upsert)
- Eliminates ~100 lines of duplicate code
- Single source of truth for upsert logic
- Easier to modify behavior (e.g., add timestamps)
- Makes `nolint:dupl` removal possible

### Part 2 (Validation)
- Reduces cognitive complexity from 14 to ~5
- Each validation type is isolated
- Easier to test each reference type independently
- Removes need for `nolint:gocognit`