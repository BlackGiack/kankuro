package session_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/BlackGiack/kankuro/internal/api"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/driver"
	"github.com/BlackGiack/kankuro/internal/mcp"
	"github.com/BlackGiack/kankuro/internal/session"
)

const (
	workspaceExitTimeout = 5 * time.Second
	workspaceStatusPoll  = 20 * time.Millisecond
)

//nolint:gocognit,govet // Exercises natural-exit cleanup and workspace config assertions in one manager lifecycle.
func TestNaturalExitRevokesOnlyItsWorkspaceKankuroEntry(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	workspace := t.TempDir()
	if err := repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: workspace}}); err != nil {
		t.Fatal(err)
	}
	workspaces, err := repo.ListWorkspaces()
	if err != nil || len(workspaces) != 1 {
		t.Fatalf("workspaces: %v", err)
	}
	driverPath := filepath.Join(t.TempDir(), "exit-agent")
	if err := os.WriteFile(driverPath, []byte("#!/bin/sh\nexit 0\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	manager := newTestManager(t, repo.Dir(), repo)
	runtime, err := mcp.Start(api.New(repo, manager))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), workspaceExitTimeout)
		defer cancel()
		if err := runtime.Shutdown(ctx); err != nil {
			t.Errorf("shutdown MCP runtime: %v", err)
		}
	})
	t.Cleanup(func() { manager.Shutdown() })
	manager.SetMCPEndpoint(runtime.Endpoint())
	selected := append(config.BuiltinMCPs(), config.MCPServer{ID: "ordinary", Name: "ordinary", URL: "https://ordinary"})
	card := config.IDCard{ID: "card", Name: "natural-exit", Driver: driverPath, WorkspaceID: workspaces[0].ID}
	sessionID, err := manager.Launch(card, nil, nil, selected)
	if err != nil {
		t.Fatalf("launch: %v", err)
	}
	deadline := time.Now().Add(workspaceExitTimeout)
	for time.Now().Before(deadline) {
		snapshot, found := manager.GetSession(sessionID)
		if found && snapshot.Status == session.StatusTerminated {
			path := driver.MCPConfigPath(snapshot.WorkingDir)
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var result struct {
				MCPServers map[string]json.RawMessage `json:"mcpServers"`
			}
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatal(err)
			}
			if len(result.MCPServers) != 1 || result.MCPServers["ordinary"] == nil || result.MCPServers["kankuro"] != nil {
				t.Fatalf("natural exit cleanup left wrong workspace MCP config: %s", data)
			}
			return
		}
		time.Sleep(workspaceStatusPoll)
	}
	t.Fatal("natural exit was not observed before cleanup timeout")
}

//nolint:govet // Each filesystem operation has an independent error assertion.
func TestWorkspaceMCPWriteFailureAbortsBeforeSpawn(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	workspace := t.TempDir()
	if err := repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: workspace}}); err != nil {
		t.Fatal(err)
	}
	workspaces, err := repo.ListWorkspaces()
	if err != nil || len(workspaces) != 1 {
		t.Fatalf("workspaces: %v", err)
	}
	driverScript := filepath.Join(t.TempDir(), "spawn-marker")
	spawnMarker := filepath.Join(t.TempDir(), "spawned")
	script := "#!/bin/sh\nprintf spawned > " + testShellQuote(spawnMarker) + "\nexec cat\n"
	if err := os.WriteFile(driverScript, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	manager := newTestManager(t, repo.Dir(), repo)
	card := config.IDCard{ID: "card", Name: "card", Driver: driverScript, WorkspaceID: workspaces[0].ID}
	configPath := driver.MCPConfigPath(workspace)
	previous := []byte(`{"mcpServers":{"old":{"url":"https://old"}}}`)
	if err := os.MkdirAll(filepath.Dir(configPath), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configPath, previous, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filepath.Dir(configPath), 0o500); err != nil {
		t.Fatal(err)
	}
	_, launchErr := manager.Launch(card, nil, nil, nil)
	if err := os.Chmod(filepath.Dir(configPath), 0o700); err != nil {
		t.Fatal(err)
	}
	if launchErr == nil {
		t.Skip("filesystem permissions are not enforced for this test user")
	}
	if !strings.Contains(launchErr.Error(), "select MCPs") {
		t.Fatalf("launch failure did not come from MCP config publication: %v", launchErr)
	}
	if _, err := os.Stat(spawnMarker); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("agent spawned after MCP config write failure: %v", err)
	}
	after, err := os.ReadFile(configPath)
	if err != nil || !bytes.Equal(after, previous) {
		t.Fatalf("failed workspace MCP write changed the previous config: %s, %v", after, err)
	}
}
