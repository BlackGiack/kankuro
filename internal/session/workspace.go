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
	_ []config.Skill,
	personas []config.Persona,
	isNew bool,
) (string, error) {
	if !card.UseTmpWorkspace {
		if errSync := syncAgentsMD(wd, card, personas, isNew); errSync != nil {
			return "", errSync
		}
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

	if errSync := syncAgentsMD(targetDir, card, personas, isNew); errSync != nil {
		return "", errSync
	}
	return targetDir, nil
}

func syncAgentsMD(finalWD string, card config.IDCard, personas []config.Persona, appendIfExist bool) error {
	agentsMD := card.AgentsMD
	if card.AgentsMDPreferPath && card.AgentsMDPath != "" {
		data, errRead := os.ReadFile(card.AgentsMDPath)
		if errRead != nil {
			return fmt.Errorf("read agents md path %s: %w", card.AgentsMDPath, errRead)
		}
		agentsMD = string(data)
	}

	if len(personas) == 1 {
		p := personas[0]
		ext := filepath.Ext(p.Source)
		// Logic matches SetPersonas for destination filename
		destName := p.Name + ext
		// Wait, I don't have nameCounts here.
		// But syncAgentsMD is called with the resolved personas.
		// If there are duplicate names in the original selection, SetPersonas will disambiguate.
		// However, here we only care if len(personas) == 1, so there's no ambiguity by definition within this set.
		relPath := filepath.Join(".agents", "agents", destName)
		personaInstruction := fmt.Sprintf(
			"Read the file %s and embody the persona described during this session,",
			relPath,
		)
		if agentsMD != "" {
			agentsMD += "\n\n" + personaInstruction
		} else {
			agentsMD = personaInstruction
		}
	}

	if agentsMD == "" {
		return nil
	}

	path := filepath.Join(finalWD, "AGENTS.md")
	if !appendIfExist {
		// G703: Path traversal via taint analysis.
		errWrite := os.WriteFile(path, []byte(agentsMD), 0o600) //nolint:gosec // path is controlled
		if errWrite != nil {
			return fmt.Errorf("write AGENTS.md: %w", errWrite)
		}
		return nil
	}

	// G703: Path traversal via taint analysis.
	f, errOpen := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600) //nolint:gosec // path is controlled
	if errOpen != nil {
		return fmt.Errorf("open AGENTS.md: %w", errOpen)
	}
	defer f.Close()

	stat, errStat := f.Stat()
	if errStat != nil {
		return fmt.Errorf("stat AGENTS.md: %w", errStat)
	}

	content := agentsMD
	if stat.Size() > 0 {
		content = "\n\n" + agentsMD
	}

	if _, errWrite := f.WriteString(content); errWrite != nil {
		return fmt.Errorf("write AGENTS.md: %w", errWrite)
	}
	return nil
}
