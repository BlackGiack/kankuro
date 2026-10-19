package config_test

import (
	"strings"
	"testing"

	"github.com/BlackGiack/kankuro/internal/config"
)

//nolint:govet // Each repository mutation asserts its own failure independently.
func TestKankuroMCPIsOfferedOnceAndProtected(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	mcps, err := repo.ListMCPs()
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, mcp := range mcps {
		if mcp.ID == config.BuiltinMCPKankuro {
			count++
			if !mcp.Builtin {
				t.Fatal("Kankuro MCP must be marked built-in")
			}
		}
	}
	if count != 1 {
		t.Fatalf("Kankuro MCP count = %d, want one", count)
	}
	if _, err := repo.UpsertMCP(config.MCPServer{ID: config.BuiltinMCPKankuro, Name: "kankuro", URL: "http://attacker"}); err == nil {
		t.Fatal("built-in MCP edit unexpectedly succeeded")
	}
	if err := repo.DeleteMCP(config.BuiltinMCPKankuro); err == nil {
		t.Fatal("built-in MCP deletion unexpectedly succeeded")
	}
	if _, err := repo.UpsertMCP(
		config.MCPServer{Name: "kankuro", URL: "http://attacker"},
	); err == nil ||
		!strings.Contains(err.Error(), "shadowed") {
		t.Fatalf("built-in MCP name collision error = %v", err)
	}
}
