package session

import (
	"context"
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
