package driver

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/fsutil"
)

const (
	kankuroMCPName                  = "kankuro"
	authorizationHeader             = "Authorization"
	kankuroDirMode      os.FileMode = 0o700
	mcpConfigFileMode   os.FileMode = 0o600
)

type mcpConfigFile struct {
	MCPServers map[string]mcpConfigEntry `json:"mcpServers"`
}

type mcpConfigEntry struct {
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	CWD     string            `json:"cwd,omitempty"`
	URL     string            `json:"url,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

// MCPConfigPath returns the one MCP config file for an agent workspace.
func MCPConfigPath(workspaceRoot string) string {
	return filepath.Join(workspaceRoot, ".kankuro", "mcp-servers.json")
}

func writeWorkspaceMCPConfig(mcps []config.MCPServer, workspaceRoot string) error {
	if strings.TrimSpace(workspaceRoot) == "" {
		return errors.New("workspace root is required for MCP sync")
	}
	kankuroDir := filepath.Join(workspaceRoot, ".kankuro")
	if err := os.MkdirAll(kankuroDir, kankuroDirMode); err != nil {
		return fmt.Errorf("create .kankuro dir: %w", err)
	}
	entries := make(map[string]mcpConfigEntry, len(mcps))
	for _, server := range mcps {
		if server.ID == config.BuiltinMCPKankuro &&
			(strings.TrimSpace(server.URL) == "" || strings.TrimSpace(server.Headers[authorizationHeader]) == "") {
			return errors.New("built-in kankuro MCP requires a runtime URL and authorization header")
		}
		entries[server.Name] = mcpConfigEntry{
			Command: server.Command,
			Args:    server.Args,
			Env:     server.Env,
			CWD:     server.CWD,
			URL:     server.URL,
			Headers: server.Headers,
		}
	}
	data, err := json.MarshalIndent(mcpConfigFile{MCPServers: entries}, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal MCP config: %w", err)
	}
	if writeErr := fsutil.WriteFileAtomic(MCPConfigPath(workspaceRoot), data, mcpConfigFileMode); writeErr != nil {
		return fmt.Errorf("write mcp-servers.json: %w", writeErr)
	}
	return nil
}

// RemoveKankuroMCPIfAuthorizationMatches removes only the Kankuro entry owned
// by token, preserving ordinary servers and workspace metadata.
func RemoveKankuroMCPIfAuthorizationMatches(path, token string) error {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read workspace MCP config: %w", err)
	}
	var root map[string]json.RawMessage
	if parseRootErr := json.Unmarshal(data, &root); parseRootErr != nil {
		return fmt.Errorf("parse workspace MCP config: %w", parseRootErr)
	}
	serversJSON, ok := root["mcpServers"]
	if !ok {
		return errors.New("workspace MCP config has no mcpServers map")
	}
	var servers map[string]json.RawMessage
	if parseServersErr := json.Unmarshal(serversJSON, &servers); parseServersErr != nil {
		return fmt.Errorf("parse workspace MCP servers: %w", parseServersErr)
	}
	if servers == nil {
		return errors.New("workspace MCP servers is not an object")
	}
	kankuroJSON, ok := servers[kankuroMCPName]
	if !ok {
		return nil
	}
	var kankuro struct {
		Headers map[string]string `json:"headers"`
	}
	if parseEntryErr := json.Unmarshal(kankuroJSON, &kankuro); parseEntryErr != nil {
		return fmt.Errorf("parse workspace Kankuro MCP entry: %w", parseEntryErr)
	}
	if kankuro.Headers[authorizationHeader] != "Bearer "+token {
		return nil
	}
	delete(servers, kankuroMCPName)
	serversJSON, marshalServersErr := json.Marshal(servers)
	if marshalServersErr != nil {
		return fmt.Errorf("marshal workspace MCP servers: %w", marshalServersErr)
	}
	root["mcpServers"] = serversJSON
	data, marshalConfigErr := json.MarshalIndent(root, "", "  ")
	if marshalConfigErr != nil {
		return fmt.Errorf("marshal workspace MCP config: %w", marshalConfigErr)
	}
	if writeErr := fsutil.WriteFileAtomic(path, data, mcpConfigFileMode); writeErr != nil {
		return fmt.Errorf("rewrite workspace MCP config: %w", writeErr)
	}
	return nil
}

// RefreshKankuroMCP rotates the built-in entry without redeploying ordinary servers
// or other workspace configuration on relaunch.
func RefreshKankuroMCP(server config.MCPServer, workspaceRoot string) error {
	path := MCPConfigPath(workspaceRoot)
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read deployed MCP config: %w", err)
	}
	var root map[string]json.RawMessage
	if parseErr := json.Unmarshal(data, &root); parseErr != nil {
		return fmt.Errorf("parse deployed MCP config: %w", parseErr)
	}
	var servers map[string]json.RawMessage
	if parseErr := json.Unmarshal(root["mcpServers"], &servers); parseErr != nil {
		return fmt.Errorf("parse deployed MCP servers: %w", parseErr)
	}
	if servers == nil {
		return errors.New("deployed MCP servers is not an object")
	}
	entry, err := json.Marshal(mcpConfigEntry{URL: server.URL, Headers: server.Headers})
	if err != nil {
		return fmt.Errorf("marshal Kankuro MCP entry: %w", err)
	}
	servers[kankuroMCPName] = entry
	serversData, err := json.Marshal(servers)
	if err != nil {
		return fmt.Errorf("marshal deployed MCP servers: %w", err)
	}
	root["mcpServers"] = serversData
	data, err = json.MarshalIndent(root, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal deployed MCP config: %w", err)
	}
	return fsutil.WriteFileAtomic(path, data, mcpConfigFileMode)
}
