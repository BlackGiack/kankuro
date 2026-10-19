package session_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/driver"
	"github.com/BlackGiack/kankuro/internal/session"
)

//nolint:gocognit,govet // Exercises credential rotation, deployment preservation, and retry after a persisted reload.
func TestRelaunchRefreshesOnlyKankuroMCPAfterRestart(t *testing.T) {
	for _, useTmp := range []bool{false, true} {
		t.Run(map[bool]string{false: "original", true: "temporary"}[useTmp], func(t *testing.T) {
			repo, workspaceID := workspaceValidationRepository(t)
			manager := newTestManager(t, repo.Dir(), repo)
			manager.SetMCPEndpoint("http://127.0.0.1:43210/mcp")
			card := config.IDCard{
				ID: "relaunch", Name: "relaunch", Driver: "cat", WorkspaceID: workspaceID,
				UseTmpWorkspace: useTmp, AgentsMD: "initial role",
			}
			selected := append(config.BuiltinMCPs(), config.MCPServer{ID: "ordinary", Name: "ordinary", Command: "initial"})
			id, err := manager.Launch(card, nil, nil, selected)
			if err != nil {
				t.Fatal(err)
			}
			before, found := manager.GetSession(id)
			if !found || !before.KankuroMCPEnabled {
				t.Fatal("missing deployed Kankuro selection")
			}
			configPath := driver.MCPConfigPath(before.WorkingDir)
			firstToken := readRelaunchToken(t, configPath)
			if err := manager.Kill(id); err != nil {
				t.Fatal(err)
			}
			manager.Shutdown()
			deployed := `{"mcpServers":{"ordinary":{"command":"user-edited","extra":{"keep":true}}},"metadata":{"keep":true}}`
			writeTestFile(t, configPath, deployed)
			instructionPath := filepath.Join(before.WorkingDir, "AGENTS.md")
			writeTestFile(t, instructionPath, "manual instructions")
			restored := newTestManager(t, repo.Dir(), repo)
			restored.SetMCPEndpoint("http://127.0.0.1:43211/mcp")
			card.MCPIDs = nil
			card.AgentsMD = "must not be deployed"
			card.AgentsMDPreferPath = true
			card.AgentsMDPath = filepath.Join(t.TempDir(), "missing-instructions")
			// A malformed deployed config must fail without replacing it or starting a process.
			writeTestFile(t, configPath, "malformed config")
			if err := restored.Relaunch(id, card); err == nil {
				t.Fatal("expected invalid deployed MCP config to block relaunch")
			}
			failed, found := restored.GetSession(id)
			if !found || failed.Status != session.StatusTerminated {
				t.Fatal("failed relaunch started an agent")
			}
			assertFileContent(t, configPath, "malformed config")
			assertFileContent(t, instructionPath, "manual instructions")
			writeTestFile(t, configPath, deployed)
			if err := restored.Relaunch(id, card); err != nil {
				t.Fatal(err)
			}
			assertFileContent(t, instructionPath, "manual instructions")
			nextToken := readRelaunchToken(t, configPath)
			if nextToken == firstToken || restored.HasActivationCredential(firstToken) || !restored.HasActivationCredential(nextToken) {
				t.Fatal("relaunch failed to rotate activation credentials")
			}
			data, err := os.ReadFile(configPath)
			if err != nil {
				t.Fatal(err)
			}
			var root struct {
				MCPServers map[string]json.RawMessage `json:"mcpServers"`
				Metadata   json.RawMessage            `json:"metadata"`
			}
			if err := json.Unmarshal(data, &root); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(root.MCPServers["ordinary"]), "user-edited") ||
				!strings.Contains(string(root.MCPServers["ordinary"]), "extra") || !strings.Contains(string(root.Metadata), "true") {
				t.Fatalf("relaunch replaced workspace-authored MCP configuration: %s", data)
			}
			if !strings.Contains(string(root.MCPServers["kankuro"]), "43211") {
				t.Fatal("relaunch kept stale MCP endpoint")
			}
		})
	}
}

func readRelaunchToken(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var cfg struct {
		MCPServers map[string]struct {
			Headers map[string]string `json:"headers"`
		} `json:"mcpServers"`
	}
	if parseErr := json.Unmarshal(data, &cfg); parseErr != nil {
		t.Fatal(parseErr)
	}
	token := strings.TrimPrefix(cfg.MCPServers["kankuro"].Headers["Authorization"], "Bearer ")
	if token == "" {
		t.Fatal("Kankuro credential missing")
	}
	return token
}
