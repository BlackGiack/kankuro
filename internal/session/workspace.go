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
		// G703: Path traversal via taint analysis.
		if err := os.MkdirAll(sessionRoot, 0o700); err != nil { //nolint:gosec // sessionRoot is derived from managed ID
			return "", fmt.Errorf("create tmp session dir: %w", err)
		}
		// G702: Command injection via taint analysis.
		// Used for local copy with verified paths.
		cmd := exec.CommandContext(context.Background(), "cp", "-r", wd, targetDir) //nolint:gosec // cp is safe here
		if out, errCopy := cmd.CombinedOutput(); errCopy != nil {
			return "", fmt.Errorf("copy workspace: %s: %w", string(out), errCopy)
		}
	} else {
		// G703: Path traversal via taint analysis.
		if _, errStat := os.Stat(targetDir); errStat != nil { //nolint:gosec // targetDir is derived from managed ID
			return "", fmt.Errorf("tmp workspace missing: %w", errStat)
		}
	}

	skillsDir := filepath.Join(targetDir, ".agents", "skills")
	// G703: Path traversal via taint analysis.
	if errRemove := os.RemoveAll(skillsDir); errRemove != nil { //nolint:gosec // skillsDir is derived from managed path
		return "", fmt.Errorf("clear tmp skills dir: %w", errRemove)
	}

	personasDir := filepath.Join(targetDir, ".agents", "agents")
	// G703: Path traversal via taint analysis.
	if errRemove := os.RemoveAll(personasDir); errRemove != nil { //nolint:gosec // personasDir is derived from managed path
		return "", fmt.Errorf("clear tmp personas dir: %w", errRemove)
	}

	return targetDir, nil
}
