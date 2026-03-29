package session

import (
	"context"
	"encoding/json"
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
	mcps []config.MCPServer,
	isNew bool,
) (string, error) {
	if !card.UseTmpWorkspace {
		if errSync := syncAgentsMD(wd, card, personas, mcps, isNew); errSync != nil {
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

	if errSync := syncAgentsMD(targetDir, card, personas, mcps, isNew); errSync != nil {
		return "", errSync
	}
	return targetDir, nil
}

func syncAgentsMD(finalWD string, card config.IDCard, personas []config.Persona, mcps []config.MCPServer, appendIfExist bool) error {
	agentsMD, err := resolveAgentsMDContent(card)
	if err != nil {
		return err
	}

	if len(personas) == 1 {
		agentsMD = appendPersonaInstruction(agentsMD, personas[0])
	}

	if len(mcps) > 0 {
		agentsMD = appendMCPInstruction(agentsMD, mcps)
	}

	if agentsMD == "" {
		return nil
	}

	return writeAgentsMD(finalWD, agentsMD, appendIfExist)
}

func resolveAgentsMDContent(card config.IDCard) (string, error) {
	if card.AgentsMDPreferPath && card.AgentsMDPath != "" {
		data, err := os.ReadFile(card.AgentsMDPath)
		if err != nil {
			return "", fmt.Errorf("read agents md path %s: %w", card.AgentsMDPath, err)
		}
		return string(data), nil
	}
	return card.AgentsMD, nil
}

func appendPersonaInstruction(content string, p config.Persona) string {
	ext := filepath.Ext(p.Source)
	relPath := filepath.Join(".agents", "agents", p.Name+ext)
	instruction := fmt.Sprintf("Read the file %s and embody the persona described during this session,", relPath)
	if content != "" {
		return content + "\n\n" + instruction
	}
	return instruction
}

func appendMCPInstruction(content string, mcps []config.MCPServer) string {
	mcpNames := make([]string, 0, len(mcps))
	for _, mcp := range mcps {
		mcpNames = append(mcpNames, mcp.Name)
	}
	mcpJSON, _ := json.Marshal(mcpNames)
	mcpLine := fmt.Sprintf("Available mcp: %s", string(mcpJSON))
	if content != "" {
		return content + "\n" + mcpLine
	}
	return mcpLine
}

func writeAgentsMD(dir, content string, appendIfExist bool) error {
	path := filepath.Join(dir, "AGENTS.md")
	if !appendIfExist {
		// G703: Path traversal via taint analysis.
		return os.WriteFile(path, []byte(content), 0o600) //nolint:gosec // path is controlled
	}

	// G703: Path traversal via taint analysis.
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600) //nolint:gosec // path is controlled
	if err != nil {
		return fmt.Errorf("open AGENTS.md: %w", err)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return fmt.Errorf("stat AGENTS.md: %w", err)
	}

	if stat.Size() > 0 {
		content = "\n\n" + content
	}

	_, err = f.WriteString(content)
	return err
}
