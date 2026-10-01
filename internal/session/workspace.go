package session

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"kankuro/internal/config"
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
		cmd := exec.CommandContext(context.Background(), "cp", "-r", wd, targetDir)
		if out, errCopy := cmd.CombinedOutput(); errCopy != nil {
			return "", fmt.Errorf("copy workspace: %s: %w", string(out), errCopy)
		}
	} else {
		if _, errStat := os.Stat(targetDir); errStat != nil {
			return "", fmt.Errorf("tmp workspace missing: %w", errStat)
		}
	}

	return targetDir, nil
}
