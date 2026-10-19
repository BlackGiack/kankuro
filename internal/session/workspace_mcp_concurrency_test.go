package session_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/BlackGiack/kankuro/internal/api"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/driver"
	"github.com/BlackGiack/kankuro/internal/mcp"
	"github.com/BlackGiack/kankuro/internal/session"
)

const workspaceMCPInterleaveTimeout = 3 * time.Second

type killGateDriver struct {
	*driver.GenericCLIDriver

	entered chan struct{}
	release <-chan struct{}
}

type spawnFailGateDriver struct {
	*driver.GenericCLIDriver

	entered chan struct{}
	release <-chan struct{}
}

func (d *spawnFailGateDriver) Spawn(_ driver.AgentConfig) (*driver.SessionHandle, error) {
	d.entered <- struct{}{}
	<-d.release
	return nil, errors.New("controlled older spawn failure")
}

func (d *killGateDriver) Spawn(cfg driver.AgentConfig) (*driver.SessionHandle, error) {
	cfg.Driver = "cat"
	return d.GenericCLIDriver.Spawn(cfg)
}

func (d *killGateDriver) Kill() error {
	select {
	case d.entered <- struct{}{}:
	default:
	}
	<-d.release
	return d.GenericCLIDriver.Kill()
}

//nolint:cyclop,gocognit,gocyclo,govet // Controlled kill interleavings exercise shared-config ownership for each selection mode.
func TestOlderSharedWorkspaceCleanupPreservesNewerSelection(t *testing.T) {
	for _, test := range []struct {
		name         string
		newServers   []config.MCPServer
		wantKankuro  bool
		wantOrdinary bool
	}{
		{
			name:        "new activation selects Kankuro and ordinary",
			newServers:  append(config.BuiltinMCPs(), config.MCPServer{ID: "ordinary", Name: "ordinary", URL: "https://ordinary"}),
			wantKankuro: true, wantOrdinary: true,
		},
		{
			name:         "new activation selects ordinary only",
			newServers:   []config.MCPServer{{ID: "ordinary", Name: "ordinary", URL: "https://ordinary"}},
			wantOrdinary: true,
		},
		{name: "new activation selects nothing"},
	} {
		t.Run(test.name, func(t *testing.T) {
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
			registry := driver.NewBuiltinRegistry()
			killEntered := make(chan struct{}, 1)
			killRelease := make(chan struct{})
			var releaseOnce sync.Once
			releaseKill := func() { releaseOnce.Do(func() { close(killRelease) }) }
			registry.Register("kill-gate", func() driver.AgentDriver {
				return &killGateDriver{
					GenericCLIDriver: driver.NewGenericCLIDriver(),
					entered:          killEntered,
					release:          killRelease,
				}
			})
			manager := newTestManager(t, repo.Dir(), repo, session.WithDriverRegistry(registry))
			runtime, err := mcp.Start(api.New(repo, manager))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := runtime.Shutdown(ctx); err != nil {
					t.Errorf("shutdown MCP runtime: %v", err)
				}
			})
			t.Cleanup(func() { manager.Shutdown() })
			t.Cleanup(releaseKill)
			manager.SetMCPEndpoint(runtime.Endpoint())

			metadataPath := filepath.Join(workspace, ".kankuro", "sub-workspaces.json")
			if err := os.MkdirAll(filepath.Dir(metadataPath), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(metadataPath, []byte(`{"preserve":true}`), 0o600); err != nil {
				t.Fatal(err)
			}
			olderCard := config.IDCard{ID: "older", Name: "older", Driver: "kill-gate", WorkspaceID: workspaces[0].ID}
			olderID, err := manager.Launch(olderCard, nil, nil, config.BuiltinMCPs())
			if err != nil {
				t.Fatalf("launch older activation: %v", err)
			}
			olderToken := tokenFromConfig(t, driver.MCPConfigPath(workspace))

			newerCard := config.IDCard{ID: "newer", Name: "newer", Driver: "cat", WorkspaceID: workspaces[0].ID}
			newerID, err := manager.Launch(newerCard, nil, nil, test.newServers)
			if err != nil {
				t.Fatalf("launch newer activation: %v", err)
			}
			mcpPath := driver.MCPConfigPath(workspace)
			newerBytes, err := os.ReadFile(mcpPath)
			if err != nil {
				t.Fatal(err)
			}
			var newerConfig struct {
				MCPServers map[string]struct {
					Headers map[string]string `json:"headers"`
				} `json:"mcpServers"`
			}
			if err := json.Unmarshal(newerBytes, &newerConfig); err != nil {
				t.Fatal(err)
			}
			if _, found := newerConfig.MCPServers["kankuro"]; found != test.wantKankuro {
				t.Fatalf("newer Kankuro selection = %v, want %v", found, test.wantKankuro)
			}
			if _, found := newerConfig.MCPServers["ordinary"]; found != test.wantOrdinary {
				t.Fatalf("newer ordinary selection = %v, want %v", found, test.wantOrdinary)
			}
			var newerToken string
			if test.wantKankuro {
				newerToken = strings.TrimPrefix(newerConfig.MCPServers["kankuro"].Headers["Authorization"], "Bearer ")
				if newerToken == olderToken {
					t.Fatal("concurrent activation reused the older credential")
				}
			}

			killResult := make(chan error, 1)
			go func() { killResult <- manager.Kill(olderID) }()
			select {
			case <-killEntered:
			case <-time.After(workspaceMCPInterleaveTimeout):
				t.Fatal("older activation kill did not reach the controlled interleaving")
			}
			releaseKill()
			select {
			case err := <-killResult:
				if err != nil {
					t.Fatalf("kill older activation: %v", err)
				}
			case <-time.After(workspaceMCPInterleaveTimeout):
				t.Fatal("older activation cleanup did not finish")
			}
			afterOlderCleanup, err := os.ReadFile(mcpPath)
			if err != nil || !bytes.Equal(afterOlderCleanup, newerBytes) {
				t.Fatalf("older cleanup modified newer config: err=%v before=%s after=%s", err, newerBytes, afterOlderCleanup)
			}
			if _, err := manager.RegisterConversation(olderToken, "stale-id", "stale"); err == nil {
				t.Fatal("older credential remained authorized after kill")
			}
			if test.wantKankuro {
				if _, err := manager.RegisterConversation(newerToken, "newer-id", "newer"); err != nil {
					t.Fatalf("newer credential was revoked by older cleanup: %v", err)
				}
			}
			if err := manager.Kill(newerID); err != nil {
				t.Fatal(err)
			}
			metadata, err := os.ReadFile(metadataPath)
			if err != nil || string(metadata) != `{"preserve":true}` {
				t.Fatalf("cleanup changed sub-workspace metadata: %q, %v", metadata, err)
			}
		})
	}
}

//nolint:govet // Controlled launch synchronization asserts that failed cleanup preserves newer workspace state.
func TestFailedOlderActivationCannotCleanupNewerWorkspaceConfig(t *testing.T) {
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
	spawnEntered := make(chan struct{}, 1)
	spawnRelease := make(chan struct{})
	var releaseOnce sync.Once
	releaseSpawn := func() { releaseOnce.Do(func() { close(spawnRelease) }) }
	registry := driver.NewBuiltinRegistry()
	registry.Register("spawn-fail", func() driver.AgentDriver {
		return &spawnFailGateDriver{
			GenericCLIDriver: driver.NewGenericCLIDriver(),
			entered:          spawnEntered,
			release:          spawnRelease,
		}
	})
	manager := newTestManager(t, repo.Dir(), repo, session.WithDriverRegistry(registry))
	runtime, err := mcp.Start(api.New(repo, manager))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := runtime.Shutdown(ctx); err != nil {
			t.Errorf("shutdown MCP runtime: %v", err)
		}
	})
	t.Cleanup(func() { manager.Shutdown() })
	t.Cleanup(releaseSpawn)
	manager.SetMCPEndpoint(runtime.Endpoint())
	olderCard := config.IDCard{ID: "older", Name: "older", Driver: "spawn-fail", WorkspaceID: workspaces[0].ID}
	olderResult := make(chan error, 1)
	go func() {
		_, launchErr := manager.Launch(olderCard, nil, nil, config.BuiltinMCPs())
		olderResult <- launchErr
	}()
	select {
	case <-spawnEntered:
	case <-time.After(workspaceMCPInterleaveTimeout):
		t.Fatal("older activation did not reach its blocked spawn")
	}
	newerCard := config.IDCard{ID: "newer", Name: "newer", Driver: "cat", WorkspaceID: workspaces[0].ID}
	newerID, err := manager.Launch(newerCard, nil, nil, config.BuiltinMCPs())
	if err != nil {
		t.Fatalf("launch newer activation: %v", err)
	}
	path := driver.MCPConfigPath(workspace)
	newerConfig, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	newerToken := tokenFromConfig(t, path)
	releaseSpawn()
	select {
	case launchErr := <-olderResult:
		if launchErr == nil || !strings.Contains(launchErr.Error(), "controlled older spawn failure") {
			t.Fatalf("older launch error = %v, want controlled spawn failure", launchErr)
		}
	case <-time.After(workspaceMCPInterleaveTimeout):
		t.Fatal("older failed activation did not finish cleanup")
	}
	configAfterCleanup, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(configAfterCleanup, newerConfig) {
		t.Fatalf("failed older cleanup changed newer config: err=%v", err)
	}
	if _, err := manager.RegisterConversation(newerToken, "newer-id", "newer"); err != nil {
		t.Fatalf("newer credential was revoked by failed older cleanup: %v", err)
	}
	if err := manager.Kill(newerID); err != nil {
		t.Fatal(err)
	}
}
