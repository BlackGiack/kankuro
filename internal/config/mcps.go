package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"sort"
)

// ListMCPs returns the MCP server registry sorted by name.
func (r *FileRepository) ListMCPs() ([]MCPServer, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.listMCPsLocked()
}

func (r *FileRepository) listMCPsLocked() ([]MCPServer, error) {
	f, err := r.LoadMCPsFile()
	if err != nil {
		return nil, err
	}
	sort.Slice(f.MCPs, func(i, j int) bool { return f.MCPs[i].Name < f.MCPs[j].Name })
	return f.MCPs, nil
}

// UpsertMCP creates or updates one MCP server entry.
func (r *FileRepository) UpsertMCP(mcp MCPServer) (MCPServer, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if mcp.Name == "" {
		return MCPServer{}, errors.New("mcp name is required")
	}
	if mcp.Command == "" && mcp.URL == "" {
		return MCPServer{}, errors.New("mcp command or url is required")
	}

	f, err := r.LoadMCPsFile()
	if err != nil {
		return MCPServer{}, err
	}

	f.MCPs, mcp = UpsertInSlice(
		f.MCPs,
		mcp,
		func(m MCPServer) string { return m.ID },
		func(m *MCPServer, id string) { m.ID = id },
	)

	if err = r.WriteMCPsFile(f); err != nil {
		return MCPServer{}, err
	}
	return mcp, nil
}

// DeleteMCP removes an MCP server unless it is referenced by existing ID cards.
func (r *FileRepository) DeleteMCP(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cards, err := r.listIDCardsLocked()
	if err != nil {
		return err
	}
	for _, card := range cards {
		if slices.Contains(card.MCPIDs, id) {
			return fmt.Errorf("mcp %s is referenced by id card %s", id, card.ID)
		}
	}
	f, err := r.LoadMCPsFile()
	if err != nil {
		return err
	}
	orig := len(f.MCPs)
	f.MCPs = slices.DeleteFunc(f.MCPs, func(m MCPServer) bool { return m.ID == id })
	if len(f.MCPs) == orig {
		return fmt.Errorf("mcp not found: %s", id)
	}
	return r.WriteMCPsFile(f)
}

// LoadMCPsFile reads and normalizes the MCP-registry YAML payload.
func (r *FileRepository) LoadMCPsFile() (MCPsFile, error) {
	path := filepath.Join(r.dir, MCPsFilename)
	f := MCPsFile{}
	if err := ReadYAML(path, &f); err != nil {
		return MCPsFile{}, err
	}
	if f.SchemaVersion == 0 {
		f.SchemaVersion = SchemaVersion
	}
	if f.MCPs == nil {
		f.MCPs = []MCPServer{}
	}
	return f, nil
}

// WriteMCPsFile writes MCP data using the current schema version.
func (r *FileRepository) WriteMCPsFile(f MCPsFile) error {
	f.SchemaVersion = SchemaVersion
	return r.WriteAtomic(filepath.Join(r.dir, MCPsFilename), f)
}
