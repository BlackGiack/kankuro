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
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.listWorkspacesLocked()
}

func (r *FileRepository) listWorkspacesLocked() ([]Workspace, error) {
	f, err := r.LoadWorkspacesFile()
	if err != nil {
		return nil, err
	}
	sort.Slice(f.Workspaces, func(i, j int) bool { return f.Workspaces[i].Name < f.Workspaces[j].Name })
	return f.Workspaces, nil
}

// ListWorkspaceSources returns the configured workspace sources.
func (r *FileRepository) ListWorkspaceSources() ([]WorkspaceSource, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.listWorkspaceSourcesLocked()
}

func (r *FileRepository) listWorkspaceSourcesLocked() ([]WorkspaceSource, error) {
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
// IDs are assigned here (not by the caller) so there is exactly one scheme
// for workspace-source identifiers.
func (r *FileRepository) SaveWorkspaceSources(sources []WorkspaceSource) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.saveWorkspaceSourcesLocked(sources)
}

func (r *FileRepository) saveWorkspaceSourcesLocked(sources []WorkspaceSource) error {
	sources = assignWorkspaceSourceIDs(sources)
	err := r.WriteWorkspaceSourcesFile(WorkspaceSourcesFile{
		SchemaVersion: SchemaVersion,
		Sources:       sources,
	})
	if err != nil {
		return err
	}
	return r.syncWorkspacesFromSourcesLocked()
}

// assignWorkspaceSourceIDs fills in an ID for any source that does not
// already have one, so there is exactly one scheme for workspace-source
// identifiers (the repository's, not a caller's).
func assignWorkspaceSourceIDs(sources []WorkspaceSource) []WorkspaceSource {
	for i := range sources {
		if sources[i].ID == "" {
			sources[i].ID = uuid.NewString()
		}
	}
	return sources
}

// UpdateWorkspaceSources reads the current workspace sources, applies fn,
// persists the result (assigning IDs to any new entries), and re-syncs the
// workspace registry - all under one lock acquisition. This closes the
// read-modify-write race a caller doing ListWorkspaceSources -> mutate ->
// SaveWorkspaceSources across two separate lock acquisitions is exposed to.
//
// If the write itself fails, that error is returned as-is - nothing was
// saved. If the write succeeds but the subsequent registry resync fails, the
// error is returned wrapped in *SyncFailedError so callers can still refresh
// the UI to show the saved source list while surfacing that the workspace
// registry built from it is now stale.
func (r *FileRepository) UpdateWorkspaceSources(fn func([]WorkspaceSource) ([]WorkspaceSource, error)) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	current, err := r.listWorkspaceSourcesLocked()
	if err != nil {
		return err
	}
	updated, err := fn(current)
	if err != nil {
		return err
	}
	if updated == nil {
		updated = []WorkspaceSource{}
	}
	updated = assignWorkspaceSourceIDs(updated)
	if err = r.WriteWorkspaceSourcesFile(WorkspaceSourcesFile{SchemaVersion: SchemaVersion, Sources: updated}); err != nil {
		return err
	}
	if err = r.syncWorkspacesFromSourcesLocked(); err != nil {
		return &SyncFailedError{Err: err}
	}
	return nil
}

// DeleteWorkspaceSource removes a source and triggers a sync.
func (r *FileRepository) DeleteWorkspaceSource(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	sources, err := r.listWorkspaceSourcesLocked()
	if err != nil {
		return err
	}
	orig := len(sources)
	sources = slices.DeleteFunc(sources, func(s WorkspaceSource) bool { return s.ID == id })
	if len(sources) == orig {
		return fmt.Errorf("workspace source not found: %s", id)
	}
	return r.saveWorkspaceSourcesLocked(sources)
}

// SyncWorkspacesFromSources rebuilds the workspace registry by scanning configured sources.
func (r *FileRepository) SyncWorkspacesFromSources() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.syncWorkspacesFromSourcesLocked()
}

func (r *FileRepository) syncWorkspacesFromSourcesLocked() error {
	sources, listErr := r.listWorkspaceSourcesLocked()
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
	r.mu.Lock()
	defer r.mu.Unlock()

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
