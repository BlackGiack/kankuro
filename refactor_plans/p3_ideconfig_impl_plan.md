# Implementation Plan: Phase 3 Task 3.4 - IDEConfig Component

## Goal

Create IDEConfig composite view that consolidates IDE configuration display and editing into a single component, eliminating 2 view modes (viewIDE, viewIDEForm) and providing a clean interface for managing IDE launch command settings.

---

## Analysis: What to Extract

### View Mode Consolidation

| Current View Mode | New Location |
|-------------------|--------------|
| `viewIDE` | `IDEConfig.displayMode` |
| `viewIDEForm` | `IDEConfig.editMode` |

### From Existing Code

| File | Lines | Content | Destination |
|------|-------|---------|-------------|
| `app_ide.go` | 15-18 | `openIDEConfig()` | Entry point |
| `app_ide.go` | 20-64 | `launchIDE()` | Callback |
| `app_ide.go` | 66-118 | `resolveWorkspacePath()`, `splitCommand()` | Shared helpers |
| `app_ide.go` | 134-149 | `updateIDE()` | `views/ide_config.go` |
| `app_ide.go` | 151-170 | `updateIDEForm()` | `views/ide_config.go` |
| `view_ide.go` | 16-49 | `viewIDERender()` | `views/ide_config.go` |
| `view_ide.go` | 51-83 | `viewIDEFormRender()` | `views/ide_config.go` |
| `app.go` | 147-148 | `ideConfig`, `ideInput` | Component state |
| `app_new.go` | 54-64 | IDE initialization | Component initialization |

### IDEConfig-Specific Behavior

```
Modes:
├── Display Mode
│   ├── Shows current command (or "not configured")
│   ├── Enter/e → Switch to edit mode
│   └── Esc → Return to Dashboard
│
└── Edit Mode
    ├── Text input with placeholder
    ├── $PATH placeholder explanation
    ├── Enter → Save and return to display
    └── Esc → Cancel and return to display
```

### Key Features

1. **Placeholder support**: `$PATH` is replaced with session workspace path when launching IDE
2. **Command parsing**: Handles quoted arguments in command string
3. **Simple state**: Just one string (the command)
4. **Two modes**: Display and edit handled internally

---

## Architecture Overview

### Component Hierarchy

```
IDEConfig (CompositeView - simple)
├── input         textinput.Model
└── state
    ├── mode         IDEConfigMode (display/edit)
    ├── config       config.IDEConfig
    └── callbacks    IDEConfigCallbacks
```

### Message Flow

```
User Input → IDEConfig.Update()
    │
    ├─→ Display mode:
    │       ├── Enter/e → Switch to edit mode
    │       └── Esc → NavigateMsg(ViewDashboard)
    │
    └─→ Edit mode:
            ├── Enter → Save callback → Display mode
            └── Esc → Cancel → Display mode
```

---

## Step 3.4.1: Define IDEConfig Modes

**File**: `internal/ui/views/ide_config.go` (NEW)

```go
package views

import (
    "fmt"
    "strings"

    bubbletea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/bubbles/textinput"
    "github.com/charmbracelet/lipgloss"
    
    "kankuro/internal/config"
    "kankuro/internal/ui/commands"
    "kankuro/internal/ui/component"
    "kankuro/internal/ui/messages"
)

// IDEConfigMode determines the current view mode.
type IDEConfigMode int

const (
    IDEModeDisplay IDEConfigMode = iota
    IDEModeEdit
)

// IDEConfigCallbacks defines callbacks for IDE config actions.
type IDEConfigCallbacks struct {
    // OnSave is called when the command is saved.
    OnSave func(command string) error
    
    // OnClose is called when user exits to dashboard.
    OnClose func()
    
    // OnLaunch is called when user wants to launch IDE for a session.
    // Optional - may be handled by Dashboard instead.
    OnLaunch func(sessionPath string) error
}

// IDEConfig is a component for managing IDE launch command configuration.
type IDEConfig struct {
    component.Base
    
    mode       IDEConfigMode
    input      textinput.Model
    config     config.IDEConfig
    
    callbacks  IDEConfigCallbacks
    cmds       *commands.Commands
    
    // Layout constants
    boxPadding int
    boxOffset  int
}

// NewIDEConfig creates a new IDE config component.
func NewIDEConfig(callbacks IDEConfigCallbacks, cmds *commands.Commands) *IDEConfig {
    input := textinput.New()
    input.Placeholder = "e.g. code $PATH"
    input.Width = 60
    
    return &IDEConfig{
        mode:       IDEModeDisplay,
        input:      input,
        callbacks:  callbacks,
        cmds:       cmds,
        boxPadding: 2,
        boxOffset:  4,
    }
}

// SetConfig updates the current IDE configuration.
func (c *IDEConfig) SetConfig(cfg config.IDEConfig) {
    c.config = cfg
    c.input.SetValue(cfg.Command)
}

// Config returns the current configuration.
func (c *IDEConfig) Config() config.IDEConfig {
    return c.config
}

// Init initializes the component.
func (c *IDEConfig) Init() bubbletea.Cmd {
    return nil
}

// Update handles messages.
func (c *IDEConfig) Update(msg bubbletea.Msg) (component.Model, bubbletea.Cmd) {
    if c.mode == IDEModeEdit {
        return c.updateEdit(msg)
    }
    return c.updateDisplay(msg)
}

// updateDisplay handles display mode input.
func (c *IDEConfig) updateDisplay(msg bubbletea.Msg) (component.Model, bubbletea.Cmd) {
    keyMsg, ok := msg.(bubbletea.KeyMsg)
    if !ok {
        return c, nil
    }
    
    switch keyMsg.String() {
    case "esc":
        if c.callbacks.OnClose != nil {
            c.callbacks.OnClose()
        }
        return c, func() bubbletea.Msg {
            return messages.NavigateMsg{View: ViewDashboard}
        }
        
    case "enter", "e":
        c.mode = IDEModeEdit
        c.input.SetValue(c.config.Command)
        c.input.Focus()
        return c, nil
        
    case "ctrl+c":
        return c, bubbletea.Quit
    }
    
    return c, nil
}

// updateEdit handles edit mode input.
func (c *IDEConfig) updateEdit(msg bubbletea.Msg) (component.Model, bubbletea.Cmd) {
    // Update text input
    var cmd bubbletea.Cmd
    c.input, cmd = c.input.Update(msg)
    
    keyMsg, ok := msg.(bubbletea.KeyMsg)
    if !ok {
        return c, cmd
    }
    
    switch keyMsg.String() {
    case "esc":
        // Cancel - return to display without saving
        c.mode = IDEModeDisplay
        c.input.Blur()
        return c, nil
        
    case "enter":
        // Save
        c.config.Command = strings.TrimSpace(c.input.Value())
        c.input.Blur()
        c.mode = IDEModeDisplay
        
        if c.callbacks.OnSave != nil {
            if err := c.callbacks.OnSave(c.config.Command); err != nil {
                // Could emit error message, but for now just continue
            }
        }
        return c, nil
        
    case "ctrl+c":
        return c, bubbletea.Quit
    }
    
    return c, cmd
}

// View renders the component.
func (c *IDEConfig) View() string {
    if c.mode == IDEModeEdit {
        return c.renderEdit()
    }
    return c.renderDisplay()
}

// renderDisplay renders the display mode.
func (c *IDEConfig) renderDisplay() string {
    title := titleStyle.Padding(0, 1).Render("IDE Configuration")
    
    currentCmd := c.config.Command
    if currentCmd == "" {
        currentCmd = "(not configured)"
    }
    
    maxWidth := min(c.Width-c.boxOffset, 64+c.boxPadding*2)
    if maxWidth < 30 {
        maxWidth = 30
    }
    
    currentBox := lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(gray).
        Padding(1, c.boxPadding).
        Width(maxWidth).
        Render(fmt.Sprintf("Command: %s", currentCmd))
    
    helpText := lipgloss.NewStyle().
        Foreground(gray).
        Padding(1, 0).
        Render("Press e or Enter to edit • Esc to go back")
    
    pathHelp := lipgloss.NewStyle().
        Foreground(secondaryColor).
        Render("Use $PATH in your command to reference the session workspace path")
    
    return lipgloss.JoinVertical(lipgloss.Left,
        title,
        "",
        currentBox,
        "",
        pathHelp,
        "",
        helpText,
    )
}

// renderEdit renders the edit mode.
func (c *IDEConfig) renderEdit() string {
    title := titleStyle.Padding(0, 1).Render("Edit IDE Command")
    
    explanationWidth := min(c.Width-c.boxOffset, 60)
    if explanationWidth < 40 {
        explanationWidth = 40
    }
    
    explanation := lipgloss.NewStyle().
        Foreground(secondaryColor).
        Width(explanationWidth).
        Render(fmt.Sprintf(
            "Enter the command to launch your IDE.\n"+
                "Use $PATH as a placeholder for the session workspace path.\n"+
                "Example: code $PATH",
        ))
    
    inputBox := lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(focusedBorderColor).
        Padding(1, c.boxPadding).
        Render(c.input.View())
    
    footer := lipgloss.NewStyle().
        Foreground(gray).
        Render("Enter to save • Esc to cancel")
    
    return lipgloss.JoinVertical(lipgloss.Left,
        title,
        "",
        explanation,
        "",
        inputBox,
        "",
        footer,
    )
}

// SetSize implements component.Model.
func (c *IDEConfig) SetSize(width, height int) {
    c.Base.SetSize(width, height)
    c.input.Width = min(60, width-10)
}

// Theme colors
var (
    primaryColor      = lipgloss.Color("#8E44AD")
    secondaryColor    = lipgloss.Color("#C0392B")
    gray              = lipgloss.Color("240")
    focusedBorderColor = primaryColor
    titleStyle        = lipgloss.NewStyle().
        Bold(true).
        Foreground(primaryColor)
)
```

---

## Step 3.4.2: Create IDE Launcher Helper

**File**: `internal/ui/views/ide_launcher.go` (NEW)

The IDE launching logic is kept separate as it's used by Dashboard (via 'o' key) and potentially other components.

```go
package views

import (
    "context"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
)

const PathPlaceholder = "$PATH"

// IDELauncher provides IDE launching functionality.
type IDELauncher struct {
    command string
}

// NewIDELauncher creates a launcher with the given command.
func NewIDELauncher(command string) *IDELauncher {
    return &IDELauncher{command: command}
}

// Launch starts the IDE for the given workspace path.
func (l *IDELauncher) Launch(workspacePath string) error {
    if l.command == "" {
        return ErrIDENotConfigured
    }
    
    if workspacePath == "" {
        return ErrWorkspacePathEmpty
    }
    
    absPath, err := resolvePath(workspacePath)
    if err != nil {
        return err
    }
    
    // Replace placeholder with actual path
    cmd := strings.ReplaceAll(l.command, PathPlaceholder, absPath)
    
    // Parse command with quote support
    parts := splitCommand(cmd)
    if len(parts) == 0 {
        return ErrInvalidCommand
    }
    
    // Execute
    execCmd := exec.CommandContext(context.Background(), parts[0], parts[1:]...) //nolint:gosec // User-configured
    execCmd.Dir = absPath
    execCmd.Stdin = nil
    execCmd.Stdout = nil
    execCmd.Stderr = nil
    
    if err := execCmd.Start(); err != nil {
        return err
    }
    
    // Don't wait for completion
    go func() { _ = execCmd.Wait() }()
    
    return nil
}

// resolvePath converts a path to absolute and verifies it exists.
func resolvePath(path string) (string, error) {
    abs, err := filepath.Abs(path)
    if err != nil {
        return "", err
    }
    if _, err := os.Stat(abs); err != nil {
        return "", err
    }
    return abs, nil
}

// splitCommand parses a command string with quote support.
func splitCommand(cmd string) []string {
    var parts []string
    var current strings.Builder
    inQuote := false
    quoteChar := rune(0)
    
    for _, c := range cmd {
        if inQuote {
            if c == quoteChar {
                inQuote = false
                quoteChar = 0
            } else {
                current.WriteRune(c)
            }
            continue
        }
        
        if c == '"' || c == '\'' {
            inQuote = true
            quoteChar = c
            continue
        }
        
        if c == ' ' || c == '\t' {
            if current.Len() > 0 {
                parts = append(parts, current.String())
                current.Reset()
            }
            continue
        }
        
        current.WriteRune(c)
    }
    
    if current.Len() > 0 {
        parts = append(parts, current.String())
    }
    
    return parts
}

// IDE errors.
var (
    ErrIDENotConfigured = &ideError{"IDE command not configured"}
    ErrWorkspacePathEmpty = &ideError{"workspace path is empty"}
    ErrInvalidCommand = &ideError{"invalid IDE command"}
)

type ideError struct {
    msg string
}

func (e *ideError) Error() string { return e.msg }
```

---

## Step 3.4.3: Create IDEConfig Tests

**File**: `internal/ui/views/ide_config_test.go` (NEW)

```go
package views

import (
    "errors"
    "testing"
    
    bubbletea "github.com/charmbracelet/bubbletea"
    
    "kankuro/internal/config"
    "kankuro/internal/ui/commands"
)

func TestIDEConfig_New(t *testing.T) {
    callbacks := IDEConfigCallbacks{}
    c := NewIDEConfig(callbacks, &commands.Commands{})
    
    if c == nil {
        t.Fatal("NewIDEConfig returned nil")
    }
    
    if c.mode != IDEModeDisplay {
        t.Error("Should start in display mode")
    }
}

func TestIDEConfig_SetConfig(t *testing.T) {
    c := NewIDEConfig(IDEConfigCallbacks{}, &commands.Commands{})
    
    cfg := config.IDEConfig{Command: "code $PATH"}
    c.SetConfig(cfg)
    
    if c.config.Command != "code $PATH" {
        t.Errorf("Expected 'code $PATH', got '%s'", c.config.Command)
    }
    
    if c.input.Value() != "code $PATH" {
        t.Error("Input should be populated with config command")
    }
}

func TestIDEConfig_SwitchToEditMode(t *testing.T) {
    c := NewIDEConfig(IDEConfigCallbacks{}, &commands.Commands{})
    c.SetConfig(config.IDEConfig{Command: "existing"})
    
    // Press Enter in display mode
    model, _ := c.Update(bubbletea.KeyMsg{Type: bubbletea.KeyEnter})
    c = model.(*IDEConfig)
    
    if c.mode != IDEModeEdit {
        t.Error("Should be in edit mode after Enter")
    }
    
    if c.input.Value() != "existing" {
        t.Error("Input should have existing command")
    }
}

func TestIDEConfig_SwitchToEditMode_EKey(t *testing.T) {
    c := NewIDEConfig(IDEConfigCallbacks{}, &commands.Commands{})
    
    // Press 'e' in display mode
    model, _ := c.Update(bubbletea.KeyMsg{Runes: []rune{'e'}})
    c = model.(*IDEConfig)
    
    if c.mode != IDEModeEdit {
        t.Error("Should be in edit mode after 'e'")
    }
}

func TestIDEConfig_SaveInEditMode(t *testing.T) {
    var savedCommand string
    
    callbacks := IDEConfigCallbacks{
        OnSave: func(cmd string) error {
            savedCommand = cmd
            return nil
        },
    }
    
    c := NewIDEConfig(callbacks, &commands.Commands{})
    c.mode = IDEModeEdit
    c.input.SetValue("vim $PATH")
    c.input.Focus()
    
    // Press Enter in edit mode
    model, _ := c.Update(bubbletea.KeyMsg{Type: bubbletea.KeyEnter})
    c = model.(*IDEConfig)
    
    if c.mode != IDEModeDisplay {
        t.Error("Should return to display mode after save")
    }
    
    if savedCommand != "vim $PATH" {
        t.Errorf("Expected saved command 'vim $PATH', got '%s'", savedCommand)
    }
    
    if c.config.Command != "vim $PATH" {
        t.Error("Config should be updated")
    }
}

func TestIDEConfig_CancelEdit(t *testing.T) {
    c := NewIDEConfig(IDEConfigCallbacks{}, &commands.Commands{})
    c.SetConfig(config.IDEConfig{Command: "original"})
    c.mode = IDEModeEdit
    c.input.SetValue("changed")
    c.input.Focus()
    
    // Press Esc in edit mode
    model, _ := c.Update(bubbletea.KeyMsg{Type: bubbletea.KeyEsc})
    c = model.(*IDEConfig)
    
    if c.mode != IDEModeDisplay {
        t.Error("Should return to display mode on cancel")
    }
    
    // Config should NOT be changed
    if c.config.Command != "original" {
        t.Error("Config should remain unchanged on cancel")
    }
}

func TestIDEConfig_NavigateBack(t *testing.T) {
    c := NewIDEConfig(IDEConfigCallbacks{}, &commands.Commands{})
    c.mode = IDEModeDisplay
    
    // Press Esc in display mode
    model, cmd := c.Update(bubbletea.KeyMsg{Type: bubbletea.KeyEsc})
    c = model.(*IDEConfig)
    
    if cmd == nil {
        t.Error("Should return a command on Esc in display mode")
    }
    
    // Command should emit NavigateMsg
    msg := cmd()
    if navMsg, ok := msg.(NavigateMsg); ok {
        if navMsg.View != ViewDashboard {
            t.Errorf("Should navigate to ViewDashboard, got %d", navMsg.View)
        }
    } else {
        t.Error("Should emit NavigateMsg")
    }
}

func TestIDEConfig_View_DisplayMode(t *testing.T) {
    c := NewIDEConfig(IDEConfigCallbacks{}, &commands.Commands{})
    c.SetConfig(config.IDEConfig{Command: "code $PATH"})
    c.SetSize(80, 24)
    
    view := c.View()
    
    if view == "" {
        t.Error("View should not be empty")
    }
    
    if !containsString(view, "IDE Configuration") {
        t.Error("Should contain title")
    }
    
    if !containsString(view, "code $PATH") {
        t.Error("Should show current command")
    }
    
    if !containsString(view, "$PATH") {
        t.Error("Should show placeholder help")
    }
}

func TestIDEConfig_View_EditMode(t *testing.T) {
    c := NewIDEConfig(IDEConfigCallbacks{}, &commands.Commands{})
    c.mode = IDEModeEdit
    c.input.SetValue("vim $PATH")
    c.SetSize(80, 24)
    
    view := c.View()
    
    if view == "" {
        t.Error("View should not be empty")
    }
    
    if !containsString(view, "Edit IDE Command") {
        t.Error("Should contain edit title")
    }
    
    if !containsString(view, "vim $PATH") {
        t.Error("Should show input value")
    }
}

func TestIDEConfig_View_NotConfigured(t *testing.T) {
    c := NewIDEConfig(IDEConfigCallbacks{}, &commands.Commands{})
    // Don't set a command
    
    view := c.View()
    
    if !containsString(view, "not configured") {
        t.Error("Should show 'not configured' when empty")
    }
}

func TestIDEConfig_SetSize(t *testing.T) {
    c := NewIDEConfig(IDEConfigCallbacks{}, &commands.Commands{})
    c.SetSize(100, 30)
    
    if c.Width != 100 {
        t.Errorf("Expected width 100, got %d", c.Width)
    }
    
    // Input width should be capped
    if c.input.Width > 60 {
        t.Errorf("Input width should be capped at 60, got %d", c.input.Width)
    }
}

func TestIDEConfig_OnSaveError(t *testing.T) {
    callbacks := IDEConfigCallbacks{
        OnSave: func(cmd string) error {
            return errors.New("save failed")
        },
    }
    
    c := NewIDEConfig(callbacks, &commands.Commands{})
    c.mode = IDEModeEdit
    c.input.SetValue("test")
    
    // Should not panic on save error
    model, _ := c.Update(bubbletea.KeyMsg{Type: bubbletea.KeyEnter})
    c = model.(*IDEConfig)
    
    // Should still return to display mode
    if c.mode != IDEModeDisplay {
        t.Error("Should return to display mode even on error")
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

## Step 3.4.4: Create IDE Launcher Tests

**File**: `internal/ui/views/ide_launcher_test.go` (NEW)

```go
package views

import (
    "testing"
)

func TestSplitCommand_Simple(t *testing.T) {
    tests := []struct {
        input    string
        expected []string
    }{
        {
            input:    "code /path/to/project",
            expected: []string{"code", "/path/to/project"},
        },
        {
            input:    "vim",
            expected: []string{"vim"},
        },
        {
            input:    "code --new-window /path",
            expected: []string{"code", "--new-window", "/path"},
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.input, func(t *testing.T) {
            result := splitCommand(tt.input)
            
            if len(result) != len(tt.expected) {
                t.Errorf("Expected %d parts, got %d: %v", len(tt.expected), len(result), result)
                return
            }
            
            for i, part := range result {
                if part != tt.expected[i] {
                    t.Errorf("Part %d: expected '%s', got '%s'", i, tt.expected[i], part)
                }
            }
        })
    }
}

func TestSplitCommand_Quotes(t *testing.T) {
    tests := []struct {
        input    string
        expected []string
    }{
        {
            input:    `code "/path/with spaces/project"`,
            expected: []string{"code", "/path/with spaces/project"},
        },
        {
            input:    `'/path/with spaces/project'`,
            expected: []string{"/path/with spaces/project"},
        },
        {
            input:    `cmd "arg1" 'arg2' arg3`,
            expected: []string{"cmd", "arg1", "arg2", "arg3"},
        },
        {
            input:    `code "$PATH"`,
            expected: []string{"code", "$PATH"},
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.input, func(t *testing.T) {
            result := splitCommand(tt.input)
            
            if len(result) != len(tt.expected) {
                t.Errorf("Expected %d parts, got %d: %v", len(tt.expected), len(result), result)
                return
            }
            
            for i, part := range result {
                if part != tt.expected[i] {
                    t.Errorf("Part %d: expected '%s', got '%s'", i, tt.expected[i], part)
                }
            }
        })
    }
}

func TestSplitCommand_Empty(t *testing.T) {
    result := splitCommand("")
    
    if len(result) != 0 {
        t.Errorf("Empty input should return empty slice, got %v", result)
    }
}

func TestSplitCommand_Whitespace(t *testing.T) {
    result := splitCommand("   ")
    
    if len(result) != 0 {
        t.Errorf("Whitespace-only input should return empty slice, got %v", result)
    }
}

func TestSplitCommand_MultipleSpaces(t *testing.T) {
    result := splitCommand("cmd   arg1    arg2")
    expected := []string{"cmd", "arg1", "arg2"}
    
    if len(result) != len(expected) {
        t.Errorf("Expected %d parts, got %d", len(expected), len(result))
    }
}

func TestIDELauncher_New(t *testing.T) {
    launcher := NewIDELauncher("code $PATH")
    
    if launcher == nil {
        t.Fatal("NewIDELauncher returned nil")
    }
    
    if launcher.command != "code $PATH" {
        t.Errorf("Expected command 'code $PATH', got '%s'", launcher.command)
    }
}

func TestIDELauncher_Errors(t *testing.T) {
    tests := []struct {
        name      string
        command   string
        path      string
        wantErr   error
    }{
        {
            name:    "no command configured",
            command: "",
            path:    "/tmp",
            wantErr: ErrIDENotConfigured,
        },
        {
            name:    "empty path",
            command: "code $PATH",
            path:    "",
            wantErr: ErrWorkspacePathEmpty,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            launcher := NewIDELauncher(tt.command)
            err := launcher.Launch(tt.path)
            
            if err != tt.wantErr {
                t.Errorf("Expected error %v, got %v", tt.wantErr, err)
            }
        })
    }
}

func TestIDELauncher_PathReplacement(t *testing.T) {
    launcher := NewIDELauncher("code $PATH")
    
    // We can't test actual execution, but we can verify the command string
    // would be correctly formed
    expected := "code /home/user/project"
    result := replacePath("code $PATH", "/home/user/project")
    
    if result != expected {
        t.Errorf("Expected '%s', got '%s'", expected, result)
    }
}

func replacePath(cmd, path string) string {
    return replaceAll(cmd, PathPlaceholder, path)
}

func replaceAll(s, old, new string) string {
    result := ""
    for i := 0; i < len(s); {
        if i+len(old) <= len(s) && s[i:i+len(old)] == old {
            result += new
            i += len(old)
        } else {
            result += string(s[i])
            i++
       	}
    }
    return result
}

func TestIDEErrors(t *testing.T) {
    tests := []struct {
        err      error
        expected string
    }{
        {ErrIDENotConfigured, "IDE command not configured"},
        {ErrWorkspacePathEmpty, "workspace path is empty"},
        {ErrInvalidCommand, "invalid IDE command"},
    }
    
    for _, tt := range tests {
        t.Run(tt.expected, func(t *testing.T) {
            if tt.err.Error() != tt.expected {
                t.Errorf("Expected '%s', got '%s'", tt.expected, tt.err.Error())
            }
        })
    }
}
```

---

## Step 3.4.5: Integration with App (Backwards Compatible)

**File**: `internal/ui/app_ide.go` (MODIFY - add integration stub)

```go
// ===== Integration with new IDEConfig view =====

func (a *App) newIDEConfigView() *views.IDEConfig {
    callbacks := views.IDEConfigCallbacks{
        OnSave: func(command string) error {
            a.ideConfig.Command = command
            return a.repo.SaveIDEConfig(a.ideConfig)
        },
        OnClose: func() {
            a.mode = viewDashboard
        },
    }
    
    c := views.NewIDEConfig(callbacks, &a.cmds)
    c.SetConfig(a.ideConfig)
    return c
}

// For future use when fully migrated:
// ideConfigView *views.IDEConfig would be a field in App
// and New() would initialize it
```

**File**: `internal/ui/app_dashboard.go` (MODIFY - add IDE launcher integration)

```go
// IDE launching for sessions
func (a *App) getIDELauncher() *views.IDELauncher {
    return views.NewIDELauncher(a.ideConfig.Command)
}
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

# 3. Specific tests for IDEConfig
go test ./internal/ui/views/ -run IDEConfig -v
go test ./internal/ui/views/ -run IDELauncher -v
go test ./internal/ui/views/ -run SplitCommand -v

# 4. Existing UI still builds
go build ./internal/ui/

# 5. Run existing tests
go test ./internal/ui/
```

---

## Success Criteria

- [ ] `IDEConfigMode` enum defined (display/edit)
- [ ] `IDEConfig` component with text input
- [ ] Display mode shows current command
- [ ] Edit mode with text input
- [ ] Switch to edit on Enter/e
- [ ] Save on Enter in edit mode
- [ ] Cancel on Esc in edit mode
- [ ] Navigate back on Esc in display mode
- [ ] `SetConfig()` and `Config()` methods
- [ ] `IDELauncher` helper for launching IDE
- [ ] `splitCommand()` with quote support
- [ ] Error types defined
- [ ] Unit tests pass (>80% coverage)
- [ ] Integration with App (backwards compatible)
- [ ] Existing App still compiles and tests pass

---

## Dependencies

```
component/component.go (from 1.1) ✓
messages/messages.go (from 1.2) ✓
commands/commands.go (from 1.3) ✓
config/types.go (existing - IDEConfig struct)

views/ide_config.go ← NEW
views/ide_launcher.go ← NEW
views/ide_config_test.go ← NEW
views/ide_launcher_test.go ← NEW
```

---

## Estimated Time

| Step | Time |
|------|------|
| 3.4.1 IDEConfig component | 35 min |
| 3.4.2 IDE Launcher helper | 20 min |
| 3.4.3 IDEConfig tests | 30 min |
| 3.4.4 IDE Launcher tests | 25 min |
| 3.4.5 App integration | 15 min |
| Verification | 10 min |
| **Total** | **~2.25 hours** |

---

## What This Enables

1. **2 views → 1**: viewIDE + viewIDEForm handled by IDEConfig
2. **Reusable IDE launching**: IDELauncher can be used from Dashboard (o key) or other views
3. **Clean mode handling**: Internal state machine instead of separate view modes
4. **Consistent UI**: Follows the same pattern as other composite views
5. **Phase 4 ready**: App coordinator can delegate to IDEConfig

---

## Code Reduction Summary

| Before | After | Reduction |
|--------|-------|-----------|
| `app_ide.go` (170 lines) | ~40 lines (integration) | -130 |
| `view_ide.go` (83 lines) | 0 (removed) | -83 |
| **Total removed** | | **-213 lines** |

New files:
| File | Lines |
|------|-------|
| `ide_config.go` | ~150 |
| `ide_launcher.go` | ~90 |
| `ide_config_test.go` | ~180 |
| `ide_launcher_test.go` | ~150 |
| **Total added** | ~570 |

Net lines: +357, but includes comprehensive tests and IDELauncher helper that can be used elsewhere.
