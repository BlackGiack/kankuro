package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"sort"

	"github.com/google/uuid"
)

// ListMCPs returns the MCP server registry sorted by name.
func (r *FileRepository) ListMCPs() ([]MCPServer, error) {
	f, err := r.loadMCPsFile()
	if err != nil {
		return nil, err
	}
	sort.Slice(f.MCPs, func(i, j int) bool { return f.MCPs[i].Name < f.MCPs[j].Name })
	return f.MCPs, nil
}

// UpsertMCP creates or updates one MCP server entry.
func (r *FileRepository) UpsertMCP(mcp MCPServer) (MCPServer, error) {
	if mcp.Name == "" {
		return MCPServer{}, errors.New("mcp name is required")
	}
	if mcp.Command == "" {
		return MCPServer{}, errors.New("mcp command is required")
	}
	if mcp.ID == "" {
		mcp.ID = uuid.NewString()
	}
	f, err := r.loadMCPsFile()
	if err != nil {
		return MCPServer{}, err
	}
	updated := false
	for i := range f.MCPs {
		if f.MCPs[i].ID == mcp.ID {
			f.MCPs[i] = mcp
			updated = true
			break
		}
	}
	if !updated {
		f.MCPs = append(f.MCPs, mcp)
	}
	if err := r.writeMCPsFile(f); err != nil {
		return MCPServer{}, err
	}
	return mcp, nil
}

// DeleteMCP removes an MCP server unless it is referenced by existing ID cards.
func (r *FileRepository) DeleteMCP(id string) error {
	cards, err := r.ListIDCards()
	if err != nil {
		return err
	}
	for _, card := range cards {
		if slices.Contains(card.MCPIDs, id) {
			return fmt.Errorf("mcp %s is referenced by id card %s", id, card.ID)
		}
	}
	f, err := r.loadMCPsFile()
	if err != nil {
		return err
	}
	orig := len(f.MCPs)
	f.MCPs = slices.DeleteFunc(f.MCPs, func(m MCPServer) bool { return m.ID == id })
	if len(f.MCPs) == orig {
		return fmt.Errorf("mcp not found: %s", id)
	}
	return r.writeMCPsFile(f)
}

// loadMCPsFile reads and normalizes the MCP-registry YAML payload.
func (r *FileRepository) loadMCPsFile() (mcpsFile, error) {
	path := filepath.Join(r.dir, mcpsFilename)
	f := mcpsFile{}
	if err := readYAML(path, &f); err != nil {
		return mcpsFile{}, err
	}
	if f.SchemaVersion == 0 {
		f.SchemaVersion = SchemaVersion
	}
	if f.MCPs == nil {
		f.MCPs = []MCPServer{}
	}
	return f, nil
}

// writeMCPsFile writes MCP data using the current schema version.
func (r *FileRepository) writeMCPsFile(f mcpsFile) error {
	f.SchemaVersion = SchemaVersion
	return r.writeAtomic(filepath.Join(r.dir, mcpsFilename), f)
}
