package driver_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/driver"
)

type testMCPConfig struct {
	MCPServers map[string]struct {
		Command string            `json:"command"`
		Args    []string          `json:"args"`
		Env     map[string]string `json:"env"`
		CWD     string            `json:"cwd"`
		URL     string            `json:"url"`
		Headers map[string]string `json:"headers"`
	} `json:"mcpServers"`
}

//nolint:gocognit,govet // Table covers complete server-set replacement and unrelated metadata preservation.
func TestSelectMCPsReplacesCompleteWorkspaceSelection(t *testing.T) {
	root := t.TempDir()
	d := driver.NewGenericCLIDriver()
	builtin := config.MCPServer{
		ID: config.BuiltinMCPKankuro, Name: "kankuro", URL: "http://127.0.0.1:1234/mcp",
		Headers: map[string]string{"Authorization": "Bearer activation-token"},
	}
	ordinary := config.MCPServer{
		ID: "ordinary", Name: "ordinary", Command: "serve", Args: []string{"--stdio"},
		Env: map[string]string{"MCP_KEY": "ordinary-secret"}, CWD: "/workspace", URL: "https://ordinary",
		Headers: map[string]string{"X-Auth": "ordinary-header"},
	}
	metadataDir := filepath.Join(root, ".kankuro")
	if err := os.MkdirAll(metadataDir, 0o700); err != nil {
		t.Fatal(err)
	}
	metadataPath := filepath.Join(metadataDir, "sub-workspaces.json")
	if err := os.WriteFile(metadataPath, []byte(`{"keep":true}`), 0o600); err != nil {
		t.Fatal(err)
	}

	for _, test := range []struct {
		name    string
		servers []config.MCPServer
		want    int
	}{
		{name: "built-in only", servers: []config.MCPServer{builtin}, want: 1},
		{name: "ordinary only", servers: []config.MCPServer{ordinary}, want: 1},
		{name: "combined", servers: []config.MCPServer{ordinary, builtin}, want: 2},
		{name: "empty clears old selection", servers: []config.MCPServer{}, want: 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := d.SelectMCPs(test.servers, root); err != nil {
				t.Fatalf("SelectMCPs: %v", err)
			}
			data, err := os.ReadFile(driver.MCPConfigPath(root))
			if err != nil {
				t.Fatal(err)
			}
			var got testMCPConfig
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("decode workspace MCP config: %v", err)
			}
			if len(got.MCPServers) != test.want {
				t.Fatalf("server count = %d, want %d: %#v", len(got.MCPServers), test.want, got.MCPServers)
			}
			if entry, ok := got.MCPServers["ordinary"]; ok {
				if entry.Command != ordinary.Command || entry.URL != ordinary.URL || entry.Args[0] != ordinary.Args[0] ||
					entry.Env["MCP_KEY"] != "ordinary-secret" || entry.CWD != ordinary.CWD || entry.Headers["X-Auth"] != "ordinary-header" {
					t.Fatalf("ordinary MCP fields changed: %#v", entry)
				}
			}
			if entry, ok := got.MCPServers["kankuro"]; ok {
				if entry.URL != builtin.URL || entry.Headers["Authorization"] != builtin.Headers["Authorization"] {
					t.Fatalf("resolved Kankuro MCP fields changed: %#v", entry)
				}
			}
			info, err := os.Stat(driver.MCPConfigPath(root))
			if err != nil || info.Mode().Perm() != 0o600 {
				t.Fatalf("config permissions = %v, %v; want 0600", info, err)
			}
			metadata, err := os.ReadFile(metadataPath)
			if err != nil || string(metadata) != `{"keep":true}` {
				t.Fatalf("unrelated workspace metadata changed: %q, %v", metadata, err)
			}
		})
	}
}

func TestSelectMCPsRejectsUnresolvedBuiltinWithoutReplacingFile(t *testing.T) {
	root := t.TempDir()
	d := driver.NewGenericCLIDriver()
	if err := d.SelectMCPs([]config.MCPServer{{ID: "ordinary", Name: "ordinary", URL: "https://ordinary"}}, root); err != nil {
		t.Fatal(err)
	}
	path := driver.MCPConfigPath(root)
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	err = d.SelectMCPs([]config.MCPServer{{ID: config.BuiltinMCPKankuro, Name: "kankuro"}}, root)
	if err == nil {
		t.Fatal("unresolved built-in MCP unexpectedly serialized")
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatalf("failed config publication replaced prior file: before=%s after=%s", before, after)
	}
}

//nolint:govet // The test deliberately checks each individual filesystem operation.
func TestSelectMCPsWriteFailureLeavesPreviousCompleteConfig(t *testing.T) {
	root := t.TempDir()
	d := driver.NewGenericCLIDriver()
	if err := d.SelectMCPs([]config.MCPServer{{ID: "before", Name: "before", URL: "https://before"}}, root); err != nil {
		t.Fatal(err)
	}
	path := driver.MCPConfigPath(root)
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	metadataDir := filepath.Dir(path)
	if err := os.Chmod(metadataDir, 0o500); err != nil {
		t.Fatal(err)
	}
	writeErr := d.SelectMCPs([]config.MCPServer{{ID: "after", Name: "after", URL: "https://after"}}, root)
	if err := os.Chmod(metadataDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if writeErr == nil {
		t.Skip("filesystem permissions are not enforced for this test user")
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatalf("failed atomic publication changed prior config: before=%s after=%s", before, after)
	}
}

//nolint:govet // Independent cleanup and read failures are asserted at each step.
func TestRemoveKankuroEntryOnlyWhenCredentialMatches(t *testing.T) {
	root := t.TempDir()
	d := driver.NewGenericCLIDriver()
	servers := []config.MCPServer{
		{
			ID:      config.BuiltinMCPKankuro,
			Name:    "kankuro",
			URL:     "http://127.0.0.1:1234/mcp",
			Headers: map[string]string{"Authorization": "Bearer token-a"},
		},
		{ID: "ordinary", Name: "ordinary", Command: "server", Args: []string{"--stdio"}},
	}
	if err := d.SelectMCPs(servers, root); err != nil {
		t.Fatal(err)
	}
	path := driver.MCPConfigPath(root)
	metadataPath := filepath.Join(filepath.Dir(path), "sub-workspaces.json")
	if err := os.WriteFile(metadataPath, []byte(`{"kept":true}`), 0o600); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := driver.RemoveKankuroMCPIfAuthorizationMatches(path, "token-b"); err != nil {
		t.Fatal(err)
	}
	unchanged, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(before, unchanged) {
		t.Fatalf("another activation's config changed: %v", err)
	}
	if err := driver.RemoveKankuroMCPIfAuthorizationMatches(path, "token-a"); err != nil {
		t.Fatal(err)
	}
	var result testMCPConfig
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}
	if len(result.MCPServers) != 1 || result.MCPServers["ordinary"].Command != "server" {
		t.Fatalf("conditional cleanup did not preserve ordinary MCP: %#v", result.MCPServers)
	}
	metadata, err := os.ReadFile(metadataPath)
	if err != nil || string(metadata) != `{"kept":true}` {
		t.Fatalf("conditional cleanup changed other workspace metadata: %q, %v", metadata, err)
	}
}

//nolint:govet // Separate filesystem and parse operations each have a focused assertion.
func TestRemoveKankuroEntryKeepsEmptyConfigAndHandlesMissingOrMalformedFiles(t *testing.T) {
	root := t.TempDir()
	d := driver.NewGenericCLIDriver()
	server := config.MCPServer{
		ID: config.BuiltinMCPKankuro, Name: "kankuro", URL: "http://127.0.0.1:1234/mcp",
		Headers: map[string]string{"Authorization": "Bearer token"},
	}
	if err := d.SelectMCPs([]config.MCPServer{server}, root); err != nil {
		t.Fatal(err)
	}
	path := driver.MCPConfigPath(root)
	if err := driver.RemoveKankuroMCPIfAuthorizationMatches(path, "token"); err != nil {
		t.Fatal(err)
	}
	var empty testMCPConfig
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &empty); err != nil || len(empty.MCPServers) != 0 {
		t.Fatalf("Kankuro-only cleanup did not retain an empty map: %#v, %v", empty, err)
	}
	if err := driver.RemoveKankuroMCPIfAuthorizationMatches(filepath.Join(root, "missing.json"), "token"); err != nil {
		t.Fatalf("missing config cleanup: %v", err)
	}
	if err := os.WriteFile(path, []byte("not-json"), 0o600); err != nil {
		t.Fatal(err)
	}
	err = driver.RemoveKankuroMCPIfAuthorizationMatches(path, "token")
	if err == nil {
		t.Fatal("malformed config cleanup unexpectedly succeeded")
	}
	malformed, readErr := os.ReadFile(path)
	if readErr != nil || string(malformed) != "not-json" {
		t.Fatalf("malformed config was modified: %q, %v", malformed, readErr)
	}
}
