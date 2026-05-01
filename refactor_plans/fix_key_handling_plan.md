# Fix Key Handling Plan: Migrate to tea.KeyType Constants

## Goal
Replace all `keyMsg.String()` comparisons with `keyMsg.Type` comparisons using Bubble Tea's native `tea.KeyType` constants. Rune-based keys (single characters like 'e', 'n', 'd') should use `tea.KeyRunes` with inner string comparison.

## Background
The codebase currently uses string constants from `constants.go` (e.g., `keyEsc = "esc"`) and compares them against `keyMsg.String()`. This is fragile because string representations can vary across terminal emulators. Bubble Tea provides proper `tea.KeyType` constants that should be used instead.

## Key Mapping

| Current Constant | Bubble Tea KeyType |
|-----------------|-------------------|
| `keyEnter` | `tea.KeyEnter` |
| `keyEsc` | `tea.KeyEsc` |
| `keyTab` | `tea.KeyTab` |
| `keyBackspace` | `tea.KeyBackspace` |
| `keyDelete` | `tea.KeyDelete` |
| `keyCtrlC` | `tea.KeyCtrlC` |
| `keySpace` | `tea.KeySpace` |
| Character keys ('e', 'n', etc.) | `tea.KeyRunes` + `keyMsg.String()` |

## New Pattern

```go
// For Update methods handling tea.Msg:
if keyMsg, ok := msg.(tea.KeyMsg); ok {
    switch keyMsg.Type {
    case tea.KeyEsc:
        // handle escape
    case tea.KeyEnter:
        // handle enter
    case tea.KeyTab:
        // handle tab
    case tea.KeyBackspace:
        // handle backspace
    case tea.KeyDelete:
        // handle delete
    case tea.KeyCtrlC:
        // handle ctrl+c
    case tea.KeySpace:
        // handle space
    case tea.KeyRunes:
        switch keyMsg.String() {
        case "e":
            // handle 'e' key
        case "n":
            // handle 'n' key
        case "m":
            // handle 'm' key
        // ... other character keys
        }
    }
}

// For functions that already receive tea.KeyMsg:
switch keyMsg.Type {
    // same pattern as above
}
```

## Files to Modify

### Phase 1: Views (11 files)
1. **`internal/ui/views/dashboard.go`**
   - Line 310: `switch keyMsg.String()` in `handleKey`
   - Line 409-418: `key := keyMsg.String()` in `handleSearch`
   - Line 644-647: `updateDeleteModal` switch (already partially fixed, needs full migration)

2. **`internal/ui/views/agent_switch.go`**
   - Line 119: `switch keyMsg.String()` in `handleKey`

3. **`internal/ui/views/ide_config.go`**
   - Line 118: `switch keyMsg.String()` in `Update` (display mode)
   - Line 145: `switch keyMsg.String()` in `handleKey` (edit mode)

4. **`internal/ui/views/session_launcher.go`**
   - Line 81: `switch keyMsg.String()` in `Update` (card selection)
   - Line 88: `switch keyMsg.String()` in `Update` (form mode)

5. **`internal/ui/views/mcp_form.go`**
   - Line 124: `switch keyMsg.String()` in `handleKey`

6. **`internal/ui/views/card_editor.go`**
   - Line 308: `switch keyMsg.String()` in `handleKey`
   - Line 588: direct `keyMsg.String() == keyBackspace` in `handleCardBackspace`

7. **`internal/ui/views/puppet_selector_modal.go`**
   - Line 64: `switch keyMsg.String()` in `Update`

8. **`internal/ui/views/capability_list.go`**
   - Line 152: `switch keyMsg.String()` in `Update`

9. **`internal/ui/views/agentsmd_editor_modal.go`**
   - Line 90: `switch keyMsg.String()` in `handleKey`

10. **`internal/ui/views/workspace_picker_modal.go`**
    - Line 158: `switch keyMsg.String()` in `handleKey`
    - Line 154: direct `"q"` check

11. **`internal/ui/views/source_manager.go`**
    - Line 144: direct `keyMsg.String() == keyEsc` comparison

### Phase 2: Components (3 files)
1. **`internal/ui/components/source_list.go`**
   - Line 133: `switch keyMsg.String()` in `Update`
   - Line 181: `switch keyMsg.String()` in `handleKey`

2. **`internal/ui/components/card_list.go`**
   - Line 127: `switch keyMsg.String()` in `Update`

3. **`internal/ui/components/session_list.go`**
   - Line 171: `switch keyMsg.String()` in `Update`

### Phase 3: Cleanup
- **`internal/ui/views/constants.go`**: Remove or deprecate key string constants (`keyEnter`, `keyEsc`, `keyTab`, `keyBackspace`, `keyDelete`, `keyCtrlC`, `keySpace`)

## Implementation Steps per File

1. Identify all key handling locations in the file
2. Replace `switch keyMsg.String()` with `switch keyMsg.Type`
3. Move special keys to use `tea.KeyType` constants
4. Move character keys to `case tea.KeyRunes:` with inner `switch keyMsg.String()`
5. Replace direct comparisons like `keyMsg.String() == keyEsc` with `keyMsg.Type == tea.KeyEsc`
6. Update imports to include `tea "github.com/charmbracelet/bubbletea"` if not present

## Verification

After each file modification:
1. Run tests for the modified package:
   ```bash
   go test ./internal/ui/views/... -v
   go test ./internal/ui/components/... -v
   ```
2. Run full test suite:
   ```bash
   go test ./... -v
   ```
3. Run linter:
   ```bash
   golangci-lint run ./...
   ```
4. Ensure build succeeds:
   ```bash
   go build ./...
   ```

## Final Verification
After all files are updated:
```bash
# Run all tests
go test ./...

# Run linter
golangci-lint run

# Build
go build ./...
```

## Example Transformations

### Example 1: Simple switch
**Before:**
```go
switch keyMsg.String() {
case keyEsc:
    return nil, true
case keyEnter:
    return d.handleSubmit()
case "e":
    return d.handleEdit()
}
```

**After:**
```go
switch keyMsg.Type {
case tea.KeyEsc:
    return nil, true
case tea.KeyEnter:
    return d.handleSubmit()
case tea.KeyRunes:
    switch keyMsg.String() {
    case "e":
        return d.handleEdit()
    }
}
```

### Example 2: Mixed special and rune keys in same case
**Before:**
```go
switch keyMsg.String() {
case keyEnter, "y":
    return d.confirmDelete()
case keyEsc, "n":
    d.deleteModal.Close()
    return d, nil
}
```

**After:**
```go
switch keyMsg.Type {
case tea.KeyEnter:
    return d.confirmDelete()
case tea.KeyEsc:
    d.deleteModal.Close()
    return d, nil
case tea.KeyRunes:
    switch keyMsg.String() {
    case "y":
        return d.confirmDelete()
    case "n":
        d.deleteModal.Close()
        return d, nil
    }
}
```

### Example 3: Direct comparison
**Before:**
```go
if keyMsg.String() == keyBackspace && len(form.Query) > 0 {
    form.Query = form.Query[:len(form.Query)-1]
}
```

**After:**
```go
if keyMsg.Type == tea.KeyBackspace && len(form.Query) > 0 {
    form.Query = form.Query[:len(form.Query)-1]
}
```

## Notes
- The `updateDeleteModal` in dashboard.go was already partially fixed but should follow the full pattern
- Test files use `tea.KeyMsg{Type: tea.KeyEsc}` which will continue to work
- Some rune-based tests may need adjustment to use `Type: tea.KeyRunes, Runes: []rune{'x'}`
