package session

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"kankuro/internal/config"
)

func (m *Manager) prepareWorkspace(
	sid string,
	card config.IDCard,
	wd string,
	isNew bool,
) (string, error) {
	if !card.UseTmpWorkspace {
		return wd, nil
	}

	tmpRoot := os.Getenv("KANKURO_TMP_DIR")
	if tmpRoot == "" {
		tmpRoot = filepath.Join(m.configDir, "tmp")
	}
	sessionRoot := filepath.Join(tmpRoot, sid)
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
