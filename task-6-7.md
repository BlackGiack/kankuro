# Task 6+7: Extract Driver Sync Logic + DI for DriverConstructors

**Priority:** Medium  
**Effort:** Medium  
**Estimated Time:** 2-3 hours

## Overview

This task addresses two related issues:
1. Duplicate driver synchronization logic between `SwitchAgent` and `startSession`
2. Global mutable state via `DriverConstructors` map

Both improvements make testing easier and reduce code duplication.

---

## Part 1: Extract Driver Sync Logic

### Current State

**File:** `internal/session/lifecycle.go`

Duplicate code exists between:

- `SwitchAgent()` lines 62-85:
```go
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

- `startSession()` lines 157-169:
```go
d := driver.NewCLIDriver(card.Driver)
if errAgentsMD := d.SetAgentsMD(card, personas, mcps, workingDir, isNew); errAgentsMD != nil {
    return errAgentsMD
}
if errSkills := d.SetSkills(skills, workingDir); errSkills != nil {
    return errSkills
}
if errPersonas := d.SetPersonas(personas, workingDir); errPersonas != nil {
    return errPersonas
}
if errMCPs := d.SelectMCPs(mcps, workingDir); errMCPs != nil {
    return errMCPs
}
```

And also in both places, the temporary directory clearing logic is duplicated.

### Proposed Solution

#### Step 1: Create Driver Sync Helper

Add to `internal/session/lifecycle.go` (near top, after imports):

```go
import (
    // ... existing imports ...
    "path/filepath"
)

// syncDriverFiles clears temp directories (if needed) and syncs skill/persona/MCP files.
func (m *Manager) syncDriverFiles(
    d driver.AgentDriver,
    card config.IDCard,
    skills []config.Skill,
    personas []config.Persona,
    mcps []config.MCPServer,
    workingDir string,
    clearTmp bool,
) error {
    if clearTmp {
        // Clear existing .agents directories in tmp workspace
        skillsDir := filepath.Join(workingDir, ".agents", "skills")
        if err := os.RemoveAll(skillsDir); err != nil {
            return fmt.Errorf("clear tmp skills dir: %w", err)
        }
        personasDir := filepath.Join(workingDir, ".agents", "agents")
        if err := os.RemoveAll(personasDir); err != nil {
            return fmt.Errorf("clear tmp personas dir: %w", err)
        }
    }
    
    if err := d.SetAgentsMD(card, personas, mcps, workingDir, clearTmp); err != nil {
        return fmt.Errorf("set agents.md: %w", err)
    }
    if err := d.SetSkills(skills, workingDir); err != nil {
        return fmt.Errorf("set skills: %w", err)
    }
    if err := d.SetPersonas(personas, workingDir); err != nil {
        return fmt.Errorf("set personas: %w", err)
    }
    if err := d.SelectMCPs(mcps, workingDir); err != nil {
        return fmt.Errorf("select mcps: %w", err)
    }
    return nil
}
```

#### Step 2: Refactor SwitchAgent

Update `SwitchAgent()` to use the helper:

```go
func (m *Manager) SwitchAgent(
    id string,
    card config.IDCard,
    skills []config.Skill,
    personas []config.Persona,
    mcps []config.MCPServer,
) error {
    m.mu.RLock()
    s, ok := m.sessions[id]
    m.mu.RUnlock()
    if !ok {
        return fmt.Errorf("session not found: %s", id)
    }
    if s.Status != StatusTerminated {
        return errors.New("session must be terminated before switching agents")
    }

    finalWD := s.WorkingDir
    if finalWD == "" {
        return errors.New("session workspace is empty")
    }
    if _, err := os.Stat(finalWD); err != nil {
        return fmt.Errorf("session workspace missing: %w", err)
    }

    d := driver.NewCLIDriver(card.Driver)
    
    // Clear tmp directories if using tmp workspace
    clearTmp := s.UseTmpWorkspace
    if err := m.syncDriverFiles(d, card, skills, personas, mcps, finalWD, clearTmp); err != nil {
        return err
    }

    args := d.SetModel(card.Model)
    h, err := d.Spawn(driver.AgentConfig{
        Driver:            card.Driver,
        Model:             card.Model,
        Args:              args,
        WorkingDir:        finalWD,
        SystemInstruction: card.SystemInstruction,
        Env:               card.Env,
    })
    if err != nil {
        return err
    }

    now := time.Now().UTC()
    // ... rest of session update logic ...
}
```

#### Step 3: Refactor startSession

Update `startSession()` to use the helper:

```go
func (m *Manager) startSession(
    sid string,
    card config.IDCard,
    skills []config.Skill,
    personas []config.Persona,
    mcps []config.MCPServer,
    isNew bool,
    existingWD string,
) error {
    // ... workspace preparation logic ...

    var workingDir string
    if card.UseTmpWorkspace {
        var err error
        workingDir, err = m.prepareWorkspace(sid, card, wd, isNew)
        if err != nil {
            return err
        }
    } else {
        workingDir = wd
    }

    d := driver.NewCLIDriver(card.Driver)
    
    // Sync driver files (clear for new tmp workspaces)
    clearTmp := card.UseTmpWorkspace && isNew
    if err := m.syncDriverFiles(d, card, skills, personas, mcps, workingDir, clearTmp); err != nil {
        return err
    }

    args := d.SetModel(card.Model)
    h, err := d.Spawn(driver.AgentConfig{
        Driver:            card.Driver,
        Model:             card.Model,
        Args:              args,
        WorkingDir:        workingDir,
        SystemInstruction: card.SystemInstruction,
        Env:               card.Env,
    })
    if err != nil {
        return err
    }

    // ... rest of session creation logic ...
}
```

---

## Part 2: Replace Global DriverConstructors with DI

### Current State

**File:** `internal/driver/factory.go`

```go
//nolint:gochecknoglobals // This map serves as a central registry for agent drivers.
var DriverConstructors = map[string]func() AgentDriver{}

func RegisterDriver(name string, ctor func() AgentDriver) {
    if ctor == nil {
        return
    }
    DriverConstructors[strings.ToLower(strings.TrimSpace(name))] = ctor
}

func NewCLIDriver(driver string) AgentDriver {
    if ctor, ok := DriverConstructors[strings.ToLower(strings.TrimSpace(driver))]; ok {
        return ctor()
    }
    return NewGenericCLIDriver()
}
```

Issues:
- Mutable global state
- Must be initialized before use
- Tests must reset/clear it
- Implicit dependency

### Proposed Solution

#### Step 1: Create DriverRegistry Type

Create `internal/driver/registry.go`:

```go
package driver

import "strings"

// DriverRegistry maps driver names to constructors.
type DriverRegistry interface {
    Get(name string) AgentDriver
}

// MutableDriverRegistry allows registration of drivers.
type MutableDriverRegistry interface {
    DriverRegistry
    Register(name string, ctor func() AgentDriver)
}

// MapDriverRegistry is an in-memory driver registry.
type MapDriverRegistry struct {
    constructors map[string]func() AgentDriver
}

// NewDriverRegistry creates an empty registry.
func NewDriverRegistry() *MapDriverRegistry {
    return &MapDriverRegistry{
        constructors: make(map[string]func() AgentDriver),
    }
}

// Register adds a driver constructor to the registry.
func (r *MapDriverRegistry) Register(name string, ctor func() AgentDriver) {
    if ctor == nil {
        return
    }
    r.constructors[strings.ToLower(strings.TrimSpace(name))] = ctor
}

// Get returns a driver by name, falling back to generic.
func (r *MapDriverRegistry) Get(name string) AgentDriver {
    key := strings.ToLower(strings.TrimSpace(name))
    if ctor, ok := r.constructors[key]; ok {
        return ctor()
    }
    return NewGenericCLIDriver()
}
```

#### Step 2: Deprecate Global Registry

Update `internal/driver/factory.go`:

```go
package driver

import "strings"

// DriverConstructors is deprecated. Use DriverRegistry instead.
// Kept for backward compatibility.
var DriverConstructors = map[string]func() AgentDriver{}

// RegisterDriver is deprecated. Use DriverRegistry.Register instead.
func RegisterDriver(name string, ctor func() AgentDriver) {
    if ctor == nil {
        return
    }
    DriverConstructors[strings.ToLower(strings.TrimSpace(name))] = ctor
}

// NewCLIDriver uses the global registry. Deprecated: use DriverRegistry.Get.
func NewCLIDriver(driverName string) AgentDriver {
    if ctor, ok := DriverConstructors[strings.ToLower(strings.TrimSpace(driverName))]; ok {
        return ctor()
    }
    return NewGenericCLIDriver()
}

// NewCLIDriverFromRegistry creates a driver using an explicit registry.
func NewCLIDriverFromRegistry(reg DriverRegistry, driverName string) AgentDriver {
    return reg.Get(driverName)
}
```

#### Step 3: Update Session Manager

Add registry to `internal/session/manager.go`:

```go
type Manager struct {
    configDir string
    repo      config.Repository
    drivers   driver.DriverRegistry  // NEW
    sessions  map[string]*Session
    mu        sync.RWMutex
}

// Option configures the Manager.
type Option func(*Manager)

// WithDriverRegistry sets a custom driver registry.
func WithDriverRegistry(reg driver.DriverRegistry) Option {
    return func(m *Manager) {
        m.drivers = reg
    }
}

func NewManager(configDir string, repo config.Repository, opts ...Option) *Manager {
    m := &Manager{
        configDir: configDir,
        repo:      repo,
        sessions:  make(map[string]*Session),
        drivers:   driver.NewDriverRegistry(), // default empty registry
    }
    for _, opt := range opts {
        opt(m)
    }
    return m
}
```

#### Step 4: Update lifecycle.go to Use Registry

Update `syncDriverFiles` and session methods:

```go
func (m *Manager) syncDriverFiles(
    card config.IDCard,
    skills []config.Skill,
    personas []config.Persona,
    mcps []config.MCPServer,
    workingDir string,
    clearTmp bool,
) (driver.AgentDriver, error) {
    d := m.drivers.Get(card.Driver)
    
    if clearTmp {
        // ... clear directories ...
    }
    
    if err := d.SetAgentsMD(card, personas, mcps, workingDir, clearTmp); err != nil {
        return nil, fmt.Errorf("set agents.md: %w", err)
    }
    // ... other sync ...
    
    return d, nil
}

func (m *Manager) startSession(...) error {
    // ... prep ...
    
    d, err := m.syncDriverFiles(card, skills, personas, mcps, workingDir, clearTmp)
    if err != nil {
        return err
    }
    
    args := d.SetModel(card.Model)
    h, err := d.Spawn(driver.AgentConfig{...})
    // ...
}
```

#### Step 5: Update main.go

Update `cmd/kankuro/main.go`:

```go
func main() {
    // Create and configure driver registry
    registry := driver.NewDriverRegistry()
    registry.Register("claude", func() driver.AgentDriver { 
        return driver.NewClaudeCodeDriver() 
    })
    
    // Keep global for backward compat during transition
    driver.RegisterDriver("claude", func() driver.AgentDriver { 
        return driver.NewClaudeCodeDriver() 
    })
    
    if err := run(registry); err != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        os.Exit(1)
    }
}

func run(registry driver.DriverRegistry) error {
    // ... flag parsing ...
    
    repo, err := config.NewFileRepository(cfgDir)
    if err != nil {
        return fmt.Errorf("config init error: %w", err)
    }

    sm := session.NewManager(
        repo.Dir(), 
        repo,
        session.WithDriverRegistry(registry),
    )
    
    // ... rest unchanged ...
}
```

#### Step 6: Update Tests

Update `internal/driver/factory_test.go`:

```go
package driver_test

import (
    "testing"
    
    "kankuro/internal/driver"
)

func TestNewDriverRegistry(t *testing.T) {
    reg := driver.NewDriverRegistry()
    reg.Register("claude", func() driver.AgentDriver {
        return driver.NewClaudeCodeDriver()
    })
    
    d := reg.Get("claude")
    if d == nil {
        t.Fatal("expected driver")
    }
    
    // Verify it's a ClaudeCodeDriver
    if _, ok := d.(*driver.ClaudeCodeDriver); !ok {
        t.Fatal("expected ClaudeCodeDriver")
    }
}

func TestRegistryFallbackToGeneric(t *testing.T) {
    reg := driver.NewDriverRegistry()
    d := reg.Get("unknown")
    if _, ok := d.(*driver.GenericCLIDriver); !ok {
        t.Fatal("expected GenericCLIDriver fallback")
    }
}

// Remove tests that mutate global state - use registry instead
```

---

## Verification Steps

```bash
# Build
go build ./...

# Test driver package
go test ./internal/driver/... -v

# Test session package
go test ./internal/session/... -v

# Run full test suite
go test ./... -v

# Manual testing:
# 1. Launch app
# 2. Create a new session with claude driver
# 3. Kill session
# 4. Switch agent on terminated session
# 5. Relaunch session
# All should use the driver registry correctly
```

---

## Files Modified

| File | Changes |
|------|---------|
| `internal/driver/registry.go` | **NEW** - DriverRegistry interface and implementation |
| `internal/driver/factory.go` | Add deprecation notices, new function |
| `internal/session/manager.go` | Add registry field, Option pattern |
| `internal/session/lifecycle.go` | Refactor sync, use registry |
| `cmd/kankuro/main.go` | Create and pass registry |
| `internal/driver/factory_test.go` | Update tests to use registry |

---

## Benefits

### Part 1 (Driver Sync)
- Eliminates ~30 lines of duplicate code
- Single source of truth for driver setup
- Easier to modify sync behavior
- Better error messages with wrapped errors

### Part 2 (DI for Registry)
- Removes mutable global state
- Makes testing easier (inject mock registry)
- Explicit dependencies in constructors
- No need to reset globals in tests
- Can eventually remove `nolint:gochecknoglobals`

---

## Migration Path

The changes are backward compatible:
1. Global `DriverConstructors` still works
2. `NewCLIDriver()` still works
3. Tests can migrate incrementally to use registry
4. Production code receives registry via dependency injection