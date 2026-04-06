package config

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/google/uuid"
)

// ListWorkspaces returns the workspace registry sorted by name.
func (r *FileRepository) ListWorkspaces() ([]Workspace, error) {
	f, err := r.LoadWorkspacesFile()
	if err != nil {
		return nil, err
	}
	sort.Slice(f.Workspaces, func(i, j int) bool { return f.Workspaces[i].Name < f.Workspaces[j].Name })
	return f.Workspaces, nil
}

// ListWorkspaceSources returns the configured workspace sources.
func (r *FileRepository) ListWorkspaceSources() ([]WorkspaceSource, error) {
	path := filepath.Join(r.dir, WorkspaceSourcesFilename)
	f := WorkspaceSourcesFile{}
	if err := ReadYAML(path, &f); err != nil {
		return nil, err
	}
	if f.SchemaVersion == 0 {
		f.SchemaVersion = SchemaVersion
	}
	if f.Sources == nil {
		f.Sources = []WorkspaceSource{}
	}
	return f.Sources, nil
}

// SaveWorkspaceSources writes the list of workspace sources to disk and syncs.
func (r *FileRepository) SaveWorkspaceSources(sources []WorkspaceSource) error {
	for i := range sources {
		if sources[i].ID == "" {
			sources[i].ID = uuid.NewString()
		}
	}
	err := r.WriteWorkspaceSourcesFile(WorkspaceSourcesFile{
		SchemaVersion: SchemaVersion,
		Sources:       sources,
	})
	if err != nil {
		return err
	}
	return r.SyncWorkspacesFromSources()
}

// DeleteWorkspaceSource removes a source and triggers a sync.
func (r *FileRepository) DeleteWorkspaceSource(id string) error {
	sources, err := r.ListWorkspaceSources()
	if err != nil {
		return err
	}
	orig := len(sources)
	sources = slices.DeleteFunc(sources, func(s WorkspaceSource) bool { return s.ID == id })
	if len(sources) == orig {
		return fmt.Errorf("workspace source not found: %s", id)
	}
	return r.SaveWorkspaceSources(sources)
}

// SyncWorkspacesFromSources rebuilds the workspace registry by scanning configured sources.
func (r *FileRepository) SyncWorkspacesFromSources() error {
	sources, listErr := r.ListWorkspaceSources()
	if listErr != nil {
		return listErr
	}
	var workspaces []Workspace
	for _, source := range sources {
		if source.IsRoot {
			rootPaths := []string{source.Path}
			scanErr := ScanEntriesWithSkip(rootPaths,
				func(entry os.DirEntry) bool { return !shouldSkipWorkspaceEntry(entry) },
				func(_, expanded string, entry os.DirEntry) error {
					fullPath := filepath.Join(expanded, entry.Name())
					workspaces = append(workspaces, Workspace{
						ID:     WorkspaceIDFromPath(fullPath),
						Name:   entry.Name(),
						Path:   fullPath,
						Source: source.ID,
					})
					return nil
				},
			)
			if scanErr != nil {
				return scanErr
			}
		} else {
			expanded, expErr := expandPath(source.Path)
			if expErr != nil {
				continue
			}
			info, statErr := os.Stat(expanded)
			if statErr != nil || !info.IsDir() {
				continue
			}
			workspaces = append(workspaces, Workspace{
				ID:     WorkspaceIDFromPath(expanded),
				Name:   filepath.Base(expanded),
				Path:   expanded,
				Source: source.ID,
			})
		}
	}

	return r.WriteWorkspacesFile(WorkspacesFile{
		SchemaVersion: SchemaVersion,
		Workspaces:    dedupeWorkspacesByPath(workspaces),
	})
}

func shouldSkipWorkspaceEntry(entry os.DirEntry) bool {
	if !entry.IsDir() {
		return true
	}
	return strings.HasPrefix(entry.Name(), ".")
}

func dedupeWorkspacesByPath(workspaces []Workspace) []Workspace {
	unique := make([]Workspace, 0, len(workspaces))
	seen := make(map[string]struct{}, len(workspaces))
	for _, workspace := range workspaces {
		if _, ok := seen[workspace.Path]; ok {
			continue
		}
		seen[workspace.Path] = struct{}{}
		unique = append(unique, workspace)
	}
	return unique
}

// GetWorkspaceByPath finds a workspace by its absolute path.
func (r *FileRepository) GetWorkspaceByPath(path string) (Workspace, bool) {
	expanded, err := expandPath(path)
	if err != nil {
		return Workspace{}, false
	}
	f, err := r.LoadWorkspacesFile()
	if err != nil {
		return Workspace{}, false
	}
	for _, w := range f.Workspaces {
		if w.Path == expanded {
			return w, true
		}
	}
	return Workspace{}, false
}

func WorkspaceIDFromPath(path string) string {
	sum := sha256.Sum256([]byte(path))
	return hex.EncodeToString(sum[:])
}

// LoadWorkspacesFile reads the workspace registry YAML.
func (r *FileRepository) LoadWorkspacesFile() (WorkspacesFile, error) {
	path := filepath.Join(r.dir, WorkspacesFilename)
	f := WorkspacesFile{}
	if err := ReadYAML(path, &f); err != nil {
		return WorkspacesFile{}, err
	}
	if f.SchemaVersion == 0 {
		f.SchemaVersion = SchemaVersion
	}
	if f.Workspaces == nil {
		f.Workspaces = []Workspace{}
	}
	return f, nil
}

// WriteWorkspacesFile writes workspaces data.
func (r *FileRepository) WriteWorkspacesFile(f WorkspacesFile) error {
	f.SchemaVersion = SchemaVersion
	return r.WriteAtomic(filepath.Join(r.dir, WorkspacesFilename), f)
}

// WriteWorkspaceSourcesFile writes workspace sources data.
func (r *FileRepository) WriteWorkspaceSourcesFile(f WorkspaceSourcesFile) error {
	f.SchemaVersion = SchemaVersion
	return r.WriteAtomic(filepath.Join(r.dir, WorkspaceSourcesFilename), f)
}
