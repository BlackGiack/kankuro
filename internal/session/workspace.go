package session

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/BlackGiack/kankuro/internal/config"
)

// tmpRoot is the directory holding temporary workspace copies: KANKURO_TMP_DIR
// when set, else <config>/tmp. Creating and removing a copy must agree on it,
// so both go through here.
func (m *Manager) tmpRoot() string {
	if root := os.Getenv("KANKURO_TMP_DIR"); root != "" {
		return root
	}
	return filepath.Join(m.configDir, "tmp")
}

func (m *Manager) prepareWorkspace(
	ctx context.Context,
	sid string,
	card config.IDCard,
	wd string,
	isNew bool,
) (string, error) {
	if !card.UseTmpWorkspace {
		return wd, nil
	}

	sessionRoot := filepath.Join(m.tmpRoot(), sid)
	targetDir := filepath.Join(sessionRoot, filepath.Base(wd))

	if isNew {
		if err := os.MkdirAll(sessionRoot, 0o700); err != nil {
			return "", fmt.Errorf("create tmp session dir: %w", err)
		}
		if err := m.copyWorkspace(ctx, wd, targetDir); err != nil {
			return "", err
		}
		return targetDir, nil
	}
	if _, err := os.Stat(targetDir); err != nil {
		return "", fmt.Errorf("tmp workspace missing: %w", err)
	}
	return targetDir, nil
}

func (m *Manager) copyWorkspace(ctx context.Context, source, destination string) error {
	if m.workspaceCopier != nil {
		if err := m.workspaceCopier(ctx, source, destination); err != nil {
			return fmt.Errorf("copy workspace: %w", err)
		}
		return nil
	}
	cmd := exec.CommandContext(ctx, "cp", "-r", source, destination)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("copy workspace: %s: %w", string(output), err)
	}
	return nil
}
