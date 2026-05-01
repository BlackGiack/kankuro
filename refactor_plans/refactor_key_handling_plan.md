# Refactor Key Handling Plan: Use tea.KeyType Constants

## Goal
Replace all `keyMsg.String()` comparisons with `keyMsg.Type` comparisons using Bubble Tea's native `tea.KeyType` constants. For rune-based keys (single characters like 'e', 'n', 'd'), continue using `keyMsg.String()` but only in a dedicated rune handling section.

## Background
Currently, the codebase uses string constants defined in `constants.go` (like `keyEsc = "esc"`, `keyEnter = "enter"`) and compares them against `keyMsg.String()`. This approach is fragile because:
1. It relies on string representation which may vary across terminal emulators
2. It doesn't use Bubble Tea's built-in key type system
3. It mixes special keys (Esc, Enter, Tab) with rune keys ('e', 'n', 'm') in the same switch statements

Bubble Tea provides proper `tea.KeyType` constants (`tea.KeyEsc`, `tea.KeyEnter`, `tea.KeyTab`, etc.) that should be used instead.

## Approach

Each key handler should use a `switch` on `keyMsg.Type` for special keys, with a `case tea.KeyRunes:` branch for character keys.

```go
// NEW PATTERN:
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
case tea.KeyRunes:
    // handle rune-based keys
    switch string(keyMsg.Runes) {
    case "e":
        // handle 'e'
    case "n":
        // handle 'n'
    // ...
    }
}
```

## Files to Refactor

### Phase 1: Views (High Priority)
1. **`internal/ui/views/dashboard.go`**
   - Line 310: `switch keyMsg.String()` - main key handler
   - Line 409: `key := keyMsg.String()` - search handler
   - Line 644-647: `updateDeleteModal` - already partially fixed

2. **`internal/ui/views/agent_switch.go`**
   - Line 119: `switch keyMsg.String()`

3. **`internal/ui/views/ide_config.go`**
   - Line 118: `switch keyMsg.String()`
   - Line 145: `switch keyMsg.String()`

4. **`internal/ui/views/session_launcher.go`**
   - Line 81: `switch keyMsg.String()`
   - Line 88: `switch keyMsg.String()`

5. **`internal/ui/views/mcp_form.go`**
   - Line 124: `switch keyMsg.String()`

6. **`internal/ui/views/card_editor.go`**
   - Line 308: `switch keyMsg.String()`

7. **`internal/ui/views/puppet_selector_modal.go`**
   - Line 64: `switch keyMsg.String()`

8. **`internal/ui/views/capability_list.go`**
   - Line 152: `switch keyMsg.String()`

9. **`internal/ui/views/agentsmd_editor_modal.go`**
   - Line 90: `switch keyMsg.String()`

10. **`internal/ui/views/workspace_picker_modal.go`**
    - Line 158: `switch keyMsg.String()`

11. **`internal/ui/views/source_manager.go`**
    - Has direct `keyMsg.String() == keyEsc` comparison at line 144

### Phase 2: Components
1. **`internal/ui/components/source_list.go`**
   - Line 133: `switch keyMsg.String()`
   - Line 181: `switch keyMsg.String()`

2. **`internal/ui/components/card_list.go`**
   - Line 127: `switch keyMsg.String()`

3. **`internal/ui/components/session_list.go`**
   - Line 171: `switch keyMsg.String()`

## Key Mapping Reference

| Current String Constant | Bubble Tea KeyType | Notes |
|------------------------|-------------------|-------|
| `keyEnter` ("enter") | `tea.KeyEnter` | Concurrent |
| `keyEsc` ("esc") | `tea.KeyEsc` | Concurrent |
| `keyTab` ("tab") | `tea.KeyTab` | Concurrent |
| `keyBackspace` ("backspace") | `tea.KeyBackspace` | Concurrent |
| `keyDelete` ("delete") | `tea.KeyDelete` | Concurrent |
| `keyCtrlC` ("ctrl+c") | `tea.KeyCtrlC` | Concurrent |
| `keySpace` (" ") | `tea.KeySpace` | Concurrent |
| | `tea.KeyRunes` | For character keys like 'e', 'n', 'm', 'd', '/' |

## Implementation Steps

### For each file:
1. Identify all `switch keyMsg.String()` statements
2. Identify all direct `keyMsg.String() == constant` comparisons
3. Replace with `switch keyMsg.Type` pattern:
   - Special keys → use `tea.KeyType` constants
   - Rune-based keys → use `case tea.KeyRunes:` with inner switch on `string(keyMsg.Runes)`
4. Remove unused string constants from `constants.go` (or deprecate them)
5. Update tests to use `tea.KeyMsg{Type: tea.KeyX}` instead of string-based messages

### Example Transformation

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
    switch string(keyMsg.Runes) {
    case "y":
        return d.confirmDelete()
    case "n":
        d.deleteModal.Close()
        return d, nil
    }
}
```

## Testing Strategy
- Run existing tests after each file refactor
- Tests already use `tea.KeyMsg{Type: tea.KeyEsc}` etc., so they should continue to work
- May need to update tests that send rune-based keys to use `Type: tea.KeyRunes, Runes: []rune{'x'}`

## Notes
- The `dashboard.go` `updateDeleteModal` has already been partially fixed but should be fully migrated to the new pattern
- Some files like `card_editor.go` line 588 have direct comparisons that need updating
- The `handleSearch` function in `dashboard.go` uses a hybrid approach that should be fully migrated
