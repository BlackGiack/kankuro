# Task 5: Split View() Renderers

**Priority:** High  
**Effort:** Medium  
**Estimated Time:** 2-3 hours

## Objective

Break down the 156-line `View()` function in `internal/ui/app_view.go` into focused render functions, one per view mode.

---

## Current State

**File:** `internal/ui/app_view.go`

The `View()` function is a massive switch statement with 16 cases, each containing significant inline rendering logic:

```go
//nolint:funlen,gocognit // Main Bubble Tea renderer delegates to many mode-specific branches by design.
func (a *App) View() string {
    errLine := ""
    if a.err != nil {
        errLine = errorStyle.Render(a.err.Error()) + "\n"
    }

    switch a.mode {
    case viewDeleteConfirm:
        return errLine + titleStyle.Render("Delete Session?") + "\n\n" + ...
    case viewCardForm:
        // 30+ lines of inline rendering
    case viewPuppetSelector:
        // 15+ lines
    // ... 13 more cases
    }
}
```

A TODO comment on line 13 acknowledges this debt.

---

## Proposed Changes

### Step 1: Create Render Function Files

Create separate files for logical groupings of render functions.

#### Create `internal/ui/view_delete.go`:

```go
package ui

import "github.com/charmbracelet/lipgloss"

func (a *App) viewDeleteConfirmRender() string {
    return titleStyle.Render("Delete Session?") + "\n\n" +
        "Are you sure you want to delete session " + a.sessionView.PendingDeleteID + "?\n" +
        "This will also delete its temporary workspace if it exists.\n\n" +
        "y=Confirm  n=Cancel"
}
```

#### Create `internal/ui/view_card_form.go`:

```go
package ui

import "github.com/charmbracelet/lipgloss"

func (a *App) viewCardFormRender() string {
    title := "New ID Card"
    if a.cardForm.EditingID != "" {
        title = "Edit ID Card"
    }
    
    workspaceLine := a.renderWorkspaceLine()
    puppetLine := a.renderPuppetLine()
    
    return titleStyle.Render(title) + "\n\n" +
        "Name: " + a.cardForm.Inputs[0].View() + "\n" +
        "Driver: " + a.cardForm.Inputs[1].View() + "\n" +
        "Model: " + a.cardForm.Inputs[2].View() + "\n" +
        workspaceLine + "\n" +
        "Face Style: " + a.cardForm.Inputs[3].View() + "\n" +
        puppetLine + "\n" +
        "Use Tmp Workspace: " + a.cardForm.Inputs[4].View() + "\n\n" +
        "AGENTS.md:\n" + a.renderAgentsMDRow() + "\n\n" +
        a.renderSkillList() + "\n\n" +
        a.renderPersonaList() + "\n\n" +
        a.renderMCPList() + "\n\n" +
        "Enter=save Tab=next Esc=cancel"
}

func (a *App) renderWorkspaceLine() string {
    line := "Workspace: " + a.workspaceSummary()
    if a.cardForm.FocusIndex == cardWorkspaceIndex {
        line = "> " + line
        line = lipgloss.NewStyle().Foreground(highlightColor).Bold(true).Render(line)
    }
    return line
}

func (a *App) renderPuppetLine() string {
    line := "Puppet: " + a.cardForm.Puppet
    if a.isPuppetFocused() {
        line = "> " + line
        line = lipgloss.NewStyle().Foreground(highlightColor).Bold(true).Render(line)
    }
    return line
}
```

#### Create `internal/ui/view_puppet_selector.go`:

```go
package ui

import (
    "kankuro/internal/ascii"
    "kankuro/internal/ascii/characters"
)

func (a *App) viewPuppetSelectorRender() string {
    title := "Select Puppet"
    selected := puppetStyles()[a.cardForm.PuppetSelectorIndex]
    p := characters.NewPuppet(func(c *ascii.Config) { c.Set("style", selected) })
    w, h := p.Bounds()
    canvas := ascii.NewCanvas(w, h)
    p.Render(canvas, 0, 0)
    preview := "\n" + canvas.String() + "\n"
    
    return titleStyle.Render(title) + "\n\n" +
        "Use Left/Right to change puppet, Enter=confirm, Esc=cancel\n\n" +
        "Selected: " + selected + preview
}
```

#### Create `internal/ui/view_sources.go`:

```go
package ui

// Skill Sources
func (a *App) viewSkillSourcesRender() string {
    return titleStyle.Render("Manage Skill Sources") + "\n\n" +
        "Defined sources:\n" + a.renderSkillSourcesList() + "\n\n" +
        "Enter/Space=edit n=new d=delete Esc=back"
}

func (a *App) viewSkillSourceFormRender() string {
    title := "New Skill Source"
    if a.skillSources.EditingIdx != -1 {
        title = "Edit Skill Sources"
    }
    return titleStyle.Render(title) + "\n\n" +
        "Path: " + a.skillSources.Input.View() + "\n\n" +
        "Enter=save Esc=cancel"
}

// Persona Sources
func (a *App) viewPersonaSourcesRender() string {
    return titleStyle.Render("Manage Persona Sources") + "\n\n" +
        "Defined sources:\n" + a.renderPersonaSourcesList() + "\n\n" +
        "Enter/Space=edit n=new d=delete Esc=back"
}

func (a *App) viewPersonaSourceFormRender() string {
    title := "New Persona Sources"
    if a.personaSources.EditingIdx != -1 {
        title = "Edit Persona Sources"
    }
    return titleStyle.Render(title) + "\n\n" +
        "Path: " + a.personaSources.Input.View() + "\n\n" +
        "Enter=save Esc=cancel"
}
```

#### Create `internal/ui/view_workspaces.go`:

```go
package ui

func (a *App) viewWorkspacesRender() string {
    return a.renderWorkspacesList() + "\n\n" +
        "Enter/Space=edit n=new d=delete r=refresh Esc=back"
}

func (a *App) viewWorkspaceSourceFormRender() string {
    title := "New Workspace Sources"
    if a.workspaces_.EditingID != "" {
        title = "Edit Workspace Sources"
    }
    return titleStyle.Render(title) + "\n\n" +
        "Path: " + a.workspaces_.Inputs[0].View() + "\n" +
        "Is Root (y/n): " + a.workspaces_.Inputs[1].View() + "\n\n" +
        "Enter=save Tab=next Esc=cancel"
}

func (a *App) viewWorkspacePickerRender() string {
    return a.renderWorkspacePicker() + "\n\n" +
        "Enter/Space=select Esc=cancel"
}
```

#### Create `internal/ui/view_agents.go`:

```go
package ui

func (a *App) viewAgentsEditorRender() string {
    return titleStyle.Render("Editing AGENTS.md") + "\n\n" +
        a.cardForm.Editor.View() + "\n\n" +
        "Ctrl+s=save Esc=cancel"
}

func (a *App) viewAgentsPathInputRender() string {
    return titleStyle.Render("Enter AGENTS.md Path") + "\n\n" +
        a.cardForm.PathInput.View() + "\n\n" +
        "Enter=save Esc=cancel"
}
```

#### Create `internal/ui/view_mcp.go`:

```go
package ui

func (a *App) viewMCPFormRender() string {
    return titleStyle.Render("Add MCP") + "\n\n" +
        "Name: " + a.mcpInputs[0].View() + "\n" +
        "Command: " + a.mcpInputs[1].View() + "\n" +
        "Args: " + a.mcpInputs[2].View() + "\n" +
        "CWD: " + a.mcpInputs[3].View() + "\n" +
        "URL: " + a.mcpInputs[4].View() + "\n" +
        "Env: " + a.mcpInputs[5].View() + "\n" +
        "Headers: " + a.mcpInputs[6].View() + "\n\n" +
        "Enter=save Tab=next Esc=cancel"
}
```

#### Create `internal/ui/view_switch.go`:

```go
package ui

import "fmt"

func (a *App) viewSwitchAgentRender() string {
    header := "Switch Agent"
    if a.sessionView.SwitchID != "" {
        header = fmt.Sprintf("Switch Agent (session %s)", a.sessionView.SwitchID)
    }
    return titleStyle.Render(header) + "\n\n" +
        a.switchList.View() + "\n" +
        "Enter=switch Esc=cancel"
}
```

#### Create `internal/ui/view_naming.go`:

```go
package ui

func (a *App) viewNamingRender() string {
    var statusLine string
    if a.naming.LaunchDone {
        if a.naming.LaunchErr != nil {
            statusLine = "Launch failed: " + a.naming.LaunchErr.Error()
        } else {
            statusLine = "Session ready. Enter to confirm name and attach."
        }
    } else {
        statusLine = "Launching session..."
    }
    return titleStyle.Render("Name Your Session") + "\n\n" +
        statusLine + "\n\n" +
        "Name: " + a.naming.Input.View() + "\n\n" +
        "Enter=confirm  (Leave empty to use session ID as name)"
}
```

### Step 2: Refactor Main View()

Update `internal/ui/app_view.go`:

```go
package ui

import (
    "github.com/charmbracelet/lipgloss"
)

// View renders the active screen for the Bubble Tea program.
func (a *App) View() string {
    errLine := ""
    if a.err != nil {
        errLine = errorStyle.Render(a.err.Error()) + "\n"
    }

    switch a.mode {
    case viewDeleteConfirm:
        return errLine + a.viewDeleteConfirmRender()
    case viewCardForm:
        return errLine + a.viewCardFormRender()
    case viewPuppetSelector:
        return errLine + a.viewPuppetSelectorRender()
    case viewSkillSources:
        return errLine + a.viewSkillSourcesRender()
    case viewSkillSourceForm:
        return errLine + a.viewSkillSourceFormRender()
    case viewPersonaSources:
        return errLine + a.viewPersonaSourcesRender()
    case viewPersonaSourceForm:
        return errLine + a.viewPersonaSourceFormRender()
    case viewWorkspaces:
        return errLine + a.viewWorkspacesRender()
    case viewWorkspaceSourceForm:
        return errLine + a.viewWorkspaceSourceFormRender()
    case viewWorkspacePicker:
        return errLine + a.viewWorkspacePickerRender()
    case viewAgentsEditor:
        return errLine + a.viewAgentsEditorRender()
    case viewAgentsPathInput:
        return errLine + a.viewAgentsPathInputRender()
    case viewMCPForm:
        return errLine + a.viewMCPFormRender()
    case viewSwitchAgent:
        return errLine + a.viewSwitchAgentRender()
    case viewNaming:
        return errLine + a.viewNamingRender()
    case viewDashboard:
        return errLine + a.renderDashboard()
    default:
        return "Unknown view"
    }
}

// renderDashboard remains in this file as it's complex and well-structured
func (a *App) renderDashboard() string {
    // ... existing dashboard rendering code ...
}
```

### Step 3: Remove nolint and TODO

Remove:
- Line 13: `// TODO: This function is quite long...`
- Line 15: `//nolint:funlen,gocognit // ...`

---

## File Structure After Refactoring

```
internal/ui/
├── app.go
├── app_view.go          # Slimmed down to ~50 lines (switch + dashboard)
├── view_delete.go       # NEW - Delete confirmation render
├── view_card_form.go    # NEW - Card form render
├── view_puppet_selector.go # NEW - Puppet selector render
├── view_sources.go      # NEW - Skill/Persona sources render
├── view_workspaces.go   # NEW - Workspace management render
├── view_agents.go       # NEW - AGENTS.md editor render
├── view_mcp.go          # NEW - MCP form render
├── view_switch.go       # NEW - Agent switch render
├── view_naming.go       # NEW - Session naming render
├── app_dashboard.go     # Existing - dashboard list rendering
├── app_render.go        # Existing - card/puppet ASCII rendering
└── ... other existing files
```

---

## Verification Steps

```bash
# Build
go build ./...

# Test UI package
go test ./internal/ui/... -v

# Manual testing - exercise all view modes:
# 1. Launch app
# 2. Press 'n' to create new card (viewCardForm)
# 3. Navigate through card form fields
# 4. Press 'p' on Puppet field (viewPuppetSelector)
# 5. Press 'm' from dashboard (viewMCPForm)
# 6. Press 'w' from dashboard (viewWorkspaces)
# 7. Launch a session (viewNaming)
# 8. Try to delete a session (viewDeleteConfirm)
```

---

## Benefits

- Each view mode has isolated rendering logic
- Easy to find where a specific view is rendered
- Simpler to test individual views
- Removes 100+ line god function
- Removes need for `nolint:funlen,gocognit`
- Follows single responsibility principle