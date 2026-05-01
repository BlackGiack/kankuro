# Misplaced Implementation Plan

This plan addresses the cleanup of hardcoded sizing defaults and the refactoring of misplaced logic identified in the `internal/ui` package, including the creation of a dedicated `app` package for orchestration.

## Objective
1.  **Refactor**: Move non-UI logic (IDE launching) out of `internal/ui`.
2.  **Architectural Refactor**: Create `internal/app` to house the `Coordinator` and `Wiring` logic, separating orchestration from UI rendering.
3.  **Sizing**: Clean up remaining hardcoded default dimensions in UI component constructors to align with the responsive architecture.

## Proposed Changes

### 1. Refactor IDE Launcher
- **Move**: `internal/ui/views/ide_launcher.go` -> `internal/session/ide_launcher.go`
- **Move**: `internal/ui/views/ide_launcher_test.go` -> `internal/session/ide_launcher_test.go`
- **Action**: Update package name to `session`.
- **Action**: Update all references to `IDELauncher` to point to the `session` package.

### 2. Create `internal/app` Package
- **Create**: `internal/app` directory.
- **Move**: `internal/ui/coordinator.go` -> `internal/app/coordinator.go`
- **Move**: `internal/ui/coordinator_test.go` -> `internal/app/coordinator_test.go`
- **Move**: `internal/ui/wiring.go` -> `internal/app/wiring.go`
- **Action**: Update package name to `app`.
- **Action**: Update imports in moved files (e.g., `internal/ui/views`, `internal/ui/messages`).
- **Action**: Update `cmd/kankuro/main.go` (if it exists) or wherever the TUI is initialized to use `app.NewCoordinator`.

### 3. Cleanup Sizing Defaults
- **File**: `internal/ui/views/agentsmd_editor_modal.go`
    - Update `NewAgentsMDEditorModal` to use safer initial defaults:
        - `editorWidth`: 40 (was 60)
        - `editorHeight`: 10 (unchanged)
        - `pathInput.Width`: 30 (was 50)
- **File**: `internal/ui/components/modal.go`
    - Update `NewModal` default `modalWidth` to 40 (was 50).
- **File**: `internal/ui/views/workspace_picker_modal.go`
    - Update `NewWorkspacePickerModal` default `modalWidth` to 40 (was 50).
    - Update `NewWorkspacePickerModal` default `modalHeight` to 15 (was 30).
- **File**: `internal/ui/views/agent_switch.go`
    - Update `NewAgentSwitchView` default `modalWidth` to 35 (was 40).
    - Update `NewAgentSwitchView` default `modalHeight` to 8 (was 10).

## Verification & Testing
1.  **Build**: `go build ./...`
2.  **Lint**: `golangci-lint run`
3.  **Tests**:
    - `go test ./internal/session/...`
    - `go test ./internal/app/...`
    - `go test ./internal/ui/...`
4.  **Manual Check**: Verify the application launches and basic navigation/IDE launching works.
