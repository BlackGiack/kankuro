# Task 1-2: Split Repository Interface + Extract Source Sync Pattern

**Priority:** High  
**Effort:** Medium  
**Estimated Time:** 2-3 hours

## Objective

1. Split the monolithic `Repository` interface (27 methods) into focused sub-interfaces following the Interface Segregation Principle
2. Extract the duplicated source synchronization pattern into a reusable generic function

---

## Part 1: Split Repository Interface

### Current State

**File:** `internal/config/repository.go`

```go
type Repository interface {
    // 27 methods mixed together
    ListIDCards() ([]IDCard, error)
    CreateIDCard(card IDCard) (IDCard, error)
    // ... 25 more methods
}
```

### Proposed Changes

#### Step 1: Add Sub-Interfaces

Add before the main `Repository` interface (around line 20):

```go
// CardRepository defines persistence operations for ID cards.
type CardRepository interface {
    ListIDCards() ([]IDCard, error)
    CreateIDCard(card IDCard) (IDCard, error)
    UpdateIDCard(card IDCard) (IDCard, error)
    DeleteIDCard(id string) error
}

// WorkspaceRepository defines persistence operations for workspaces.
type WorkspaceRepository interface {
    ListWorkspaceSources() ([]WorkspaceSource, error)
    SaveWorkspaceSources(sources []WorkspaceSource) error
    SyncWorkspacesFromSources() error
    ListWorkspaces() ([]Workspace, error)
    GetWorkspaceByPath(path string) (Workspace, bool)
    DeleteWorkspaceSource(id string) error
}

// SkillRepository defines persistence operations for skills.
type SkillRepository interface {
    ListSkillSources() ([]string, error)
    SaveSkillSources(paths []string) error
    SyncSkillsFromSources() error
    ListSkills() ([]Skill, error)
    UpsertSkill(skill Skill) (Skill, error)
    DeleteSkill(id string) error
}

// PersonaRepository defines persistence operations for personas.
type PersonaRepository interface {
    ListPersonaSources() ([]string, error)
    SavePersonaSources(paths []string) error
    SyncPersonasFromSources() error
    ListPersonas() ([]Persona, error)
    UpsertPersona(persona Persona) (Persona, error)
    DeletePersona(id string) error
}

// MCPRepository defines persistence operations for MCP servers.
type MCPRepository interface {
    ListMCPs() ([]MCPServer, error)
    UpsertMCP(mcp MCPServer) (MCPServer, error)
    DeleteMCP(id string) error
}

// Repository composes all domain repositories for convenience.
// Individual interfaces can be used where only specific domains are needed.
type Repository interface {
    CardRepository
    WorkspaceRepository
    SkillRepository
    PersonaRepository
    MCPRepository
    Dir() string
}
```

#### Step 2: Remove Old Interface Methods

Delete the original method declarations (lines 22-52), keeping only the composite interface.

#### Step 3: Update Documentation

Update the docstring for `Repository`:

```go
// Repository composes all domain repositories for convenience.
// Use individual sub-interfaces (CardRepository, etc.) when you only need
// operations for a specific domain.
type Repository interface {
    CardRepository
    WorkspaceRepository
    SkillRepository
    PersonaRepository
    MCPRepository
    Dir() string
}
```

### Verification

No code changes needed elsewhere - `FileRepository` already implements all methods. The composite interface maintains backward compatibility.

```bash
go build ./...
go test ./internal/config/...
```

---

## Part 2: Extract Source Sync Pattern

### Current State

Three nearly identical sync functions exist:

1. `internal/config/skill_sources.go:30-86` - `SyncSkillsFromSources()`
2. `internal/config/persona_sources.go:29-70` - `SyncPersonasFromSources()`
3. `internal/config/workspaces.go:74-160` - `SyncWorkspacesFromSources()`

Each follows the same pattern:
1. Load source paths
2. Expand each path (handle `~`)
3. Read directory entries
4. Filter and build items from entries
5. Write to file

### Proposed Changes

#### Step 1: Create Generic Sync Helper

Create new file `internal/config/sync.go`:

```go
package config

import (
    "os"
    "path/filepath"
    "strings"
)

// expandPath handles ~ expansion and returns absolute path.
func expandPath(path string) (string, error) {
    if strings.HasPrefix(path, "~") {
        home, err := os.UserHomeDir()
        if err != nil {
            return "", err
        }
        path = filepath.Join(home, strings.TrimPrefix(path, "~"))
    }
    return filepath.Abs(path)
}

// EntryScanner defines how to scan a source directory for items.
type EntryScanner[T any] interface {
    ScanEntry(source, expanded string, entry os.DirEntry) (T, error)
}

// SyncFromSources reads items from multiple source directories.
// It filters entries using shouldInclude and builds items using scanner.
func SyncFromSources[T any](
    sources []string,
    scanner EntryScanner[T],
    shouldInclude func(os.DirEntry) bool,
) ([]T, error) {
    var items []T
    for _, source := range sources {
        expanded, err := expandPath(source)
        if err != nil {
            return nil, err
        }
        entries, err := os.ReadDir(expanded)
        if err != nil {
            return nil, err
        }
        for _, entry := range entries {
            if !shouldInclude(entry) {
                continue
            }
            item, err := scanner.ScanEntry(source, expanded, entry)
            if err != nil {
                return nil, err
            }
            items = append(items, item)
        }
    }
    return items, nil
}
```

#### Step 2: Create Skill Scanner

Update `internal/config/skill_sources.go`:

```go
// Remove nolint:dupl comment at top

// skillScanner implements EntryScanner for skills.
type skillScanner struct{}

func (s skillScanner) ScanEntry(source, expanded string, entry os.DirEntry) (Skill, error) {
    skillPath := filepath.Join(expanded, entry.Name())
    skill, err := ReadSkill(skillPath)
    if err != nil {
        return Skill{}, err
    }
    skill.Source = source
    return skill, nil
}

func (r *FileRepository) SyncSkillsFromSources() error {
    sources, err := r.ListSkillSources()
    if err != nil {
        return err
    }
    scanner := skillScanner{}
    shouldInclude := func(entry os.DirEntry) bool {
        return entry.IsDir() && !strings.HasPrefix(entry.Name(), ".")
    }
    skills, err := SyncFromSources(sources, scanner, shouldInclude)
    if err != nil {
        return err
    }
    return r.WriteSkillsFile(SkillsFile{
        SchemaVersion: SchemaVersion,
        Skills:        skills,
    })
}
```

#### Step 3: Create Persona Scanner

Update `internal/config/persona_sources.go`:

```go
// Remove nolint:dupl comment at top

// personaScanner implements EntryScanner for personas.
type personaScanner struct{}

func (s personaScanner) ScanEntry(source, expanded string, entry os.DirEntry) (Persona, error) {
    if !strings.HasSuffix(strings.ToLower(entry.Name()), ".md") {
        return Persona{}, nil // skip non-markdown
    }
    persona, err := ReadPersona(expanded, entry)
    if err != nil {
        return Persona{}, err
    }
    persona.Source = source
    return persona, nil
}

func (r *FileRepository) SyncPersonasFromSources() error {
    sources, err := r.ListPersonaSources()
    if err != nil {
        return err
    }
    scanner := personaScanner{}
    shouldInclude := func(entry os.DirEntry) bool {
        return !entry.IsDir() && !strings.HasPrefix(entry.Name(), ".")
    }
    personas, err := SyncFromSources(sources, scanner, shouldInclude)
    if err != nil {
        return err
    }
    return r.WritePersonasFile(PersonasFile{
        SchemaVersion: SchemaVersion,
        Personas:      personas,
    })
}
```

#### Step 4: Create Workspace Scanner

Update `internal/config/workspaces.go`:

```go
// workspaceScanner implements EntryScanner for workspaces.
type workspaceScanner struct {
    isRoot bool
}

func (s workspaceScanner) ScanEntry(source, expanded string, entry os.DirEntry) (Workspace, error) {
    entryPath := filepath.Join(expanded, entry.Name())
    return Workspace{
        ID:     generateID(entryPath),
        Name:   entry.Name(),
        Path:   entryPath,
        Source: source,
    }, nil
}

func (r *FileRepository) SyncWorkspacesFromSources() error {
    sources, err := r.ListWorkspaceSources()
    if err != nil {
        return err
    }
    var workspaces []Workspace
    for _, source := range sources {
        expanded, err := expandPath(source.Path)
        if err != nil {
            return err
        }
        if source.IsRoot {
            // Scan children as workspaces
            scanner := workspaceScanner{isRoot: true}
            shouldInclude := func(entry os.DirEntry) bool {
                return entry.IsDir() && !strings.HasPrefix(entry.Name(), ".")
            }
            ws, err := SyncFromSources([]string{source.Path}, scanner, shouldInclude)
            if err != nil {
                return err
            }
            workspaces = append(workspaces, ws...)
        } else {
            // The path itself is a workspace
            workspaces = append(workspaces, Workspace{
                ID:     generateID(expanded),
                Name:   filepath.Base(expanded),
                Path:   expanded,
                Source: source.ID,
            })
        }
    }
    return r.WriteWorkspacesFile(WorkspacesFile{
        SchemaVersion: SchemaVersion,
        Workspaces:    workspaces,
    })
}
```

#### Step 5: Update Repository Interface

Move `expandPath` to be a package-level helper in `sync.go` since it's now used by all sync functions.

---

## Verification Steps

```bash
# Build
go build ./...

# Test all config tests
go test ./internal/config/... -v

# Run app to verify sync works
./kankuro --debug
# In app: press 'w' to open workspace manager, 'r' to refresh
# Verify workspaces appear correctly
```

---

## Files Modified

| File | Changes |
|------|---------|
| `internal/config/repository.go` | Add sub-interfaces, simplify main interface |
| `internal/config/sync.go` | **NEW** - Generic sync helper |
| `internal/config/skill_sources.go` | Use SyncFromSources, remove dupl nolint |
| `internal/config/persona_sources.go` | Use SyncFromSources, remove dupl nolint |
| `internal/config/workspaces.go` | Use SyncFromSources |

---

## Benefits

### Part 1 (Repository Split)
- Follows Interface Segregation Principle
- Enables targeted mocking in tests (mock only what you need)
- Documents domain boundaries clearly
- No breaking changes (backward compatible)

### Part 2 (Source Sync)
- Eliminates ~150 lines of duplicate code
- Single source of truth for sync logic
- Easier to add new source types
- Removes `nolint:dupl` annotations