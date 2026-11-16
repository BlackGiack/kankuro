package session

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ErrCommandNotConfigured is returned when a workspace command is empty.
var ErrCommandNotConfigured = &ideError{"workspace command not configured"}

// LaunchWorkspaceCommand starts a configured command in a workspace directory.
func LaunchWorkspaceCommand(command, workspacePath string) error {
	if command == "" {
		return ErrCommandNotConfigured
	}
	if workspacePath == "" {
		return ErrWorkspacePathEmpty
	}
	absPath, err := filepath.Abs(workspacePath)
	if err != nil {
		return err
	}
	if _, err = os.Stat(absPath); err != nil {
		return err
	}
	parts := SplitCommand(strings.ReplaceAll(command, PathPlaceholder, absPath))
	if len(parts) == 0 {
		return ErrInvalidCommand
	}
	child := exec.CommandContext(context.Background(), parts[0], parts[1:]...) //nolint:gosec // configured command
	child.Dir = absPath
	child.Stdin, child.Stdout, child.Stderr = nil, nil, nil
	if startErr := child.Start(); startErr != nil {
		return startErr
	}
	go func() { _ = child.Wait() }()
	return nil
}

// OpenWorkspaceChecked protects session identity and workspace cleanup until
// the configured child has started. The child outlives the manager lock and
// request. Cancellation is checked before start; it does not stop a started child.
// check runs under the manager lock and must not call the manager.
func (m *Manager) OpenWorkspaceChecked(ctx context.Context, id, command string, check func([]Snapshot) error) error {
	if check == nil {
		return errors.New("workspace scope check is required")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return ErrManagerClosed
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := check(m.listLocked()); err != nil {
		return err
	}
	s, ok := m.sessions[id]
	if !ok {
		return fmt.Errorf("%w: %s", ErrSessionNotFound, id)
	}
	if s.deleting {
		return fmt.Errorf("%w: session is being removed", ErrLifecycleConflict)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return LaunchWorkspaceCommand(command, s.WorkingDir)
}
