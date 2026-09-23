package config

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
)

// ListPersonaSources returns the configured root folders for persona discovery.
func (r *FileRepository) ListPersonaSources() ([]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.listPersonaSourcesLocked()
}

func (r *FileRepository) listPersonaSourcesLocked() ([]string, error) {
	path := filepath.Join(r.dir, PersonaSourcesFilename)
	f := PersonaSourcesFile{}
	if err := ReadYAML(path, &f); err != nil {
		return nil, err
	}
	if f.SchemaVersion == 0 {
		f.SchemaVersion = SchemaVersion
	}
	if f.Paths == nil {
		f.Paths = []string{}
	}
	return f.Paths, nil
}

// SyncPersonasFromSources rebuilds the personas registry by scanning configured source folders.
func (r *FileRepository) SyncPersonasFromSources() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.syncPersonasFromSourcesLocked()
}

func (r *FileRepository) syncPersonasFromSourcesLocked() error {
	sources, err := r.listPersonaSourcesLocked()
	if err != nil {
		return err
	}
	var personas []Persona
	err = ScanEntries(sources,
		func(entry os.DirEntry) bool {
			return !entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") &&
				strings.HasSuffix(strings.ToLower(entry.Name()), ".md")
		},
		func(_, expanded string, entry os.DirEntry) error {
			fullPath := filepath.Join(expanded, entry.Name())
			personas = append(personas, Persona{
				ID:     PersonaIDFromSource(fullPath),
				Name:   strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())),
				Source: fullPath,
			})
			return nil
		},
	)
	if err != nil {
		return err
	}
	return r.WritePersonasFile(PersonasFile{SchemaVersion: SchemaVersion, Personas: personas})
}

func PersonaIDFromSource(path string) string {
	sum := sha256.Sum256([]byte(path))
	return hex.EncodeToString(sum[:])
}

// SavePersonaSources writes the list of persona sources to disk.
func (r *FileRepository) SavePersonaSources(paths []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if paths == nil {
		paths = []string{}
	}
	return r.WritePersonaSourcesFile(PersonaSourcesFile{
		SchemaVersion: SchemaVersion,
		Paths:         paths,
	})
}

// UpdatePersonaSources reads the current persona sources, applies fn,
// persists the result, and re-syncs the persona registry - all under one
// lock acquisition. See UpdateSkillSources for why this exists and what a
// *SyncFailedError return means.
func (r *FileRepository) UpdatePersonaSources(fn func([]string) ([]string, error)) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	current, err := r.listPersonaSourcesLocked()
	if err != nil {
		return err
	}
	updated, err := fn(current)
	if err != nil {
		return err
	}
	if updated == nil {
		updated = []string{}
	}
	if err = r.WritePersonaSourcesFile(PersonaSourcesFile{SchemaVersion: SchemaVersion, Paths: updated}); err != nil {
		return err
	}
	if err = r.syncPersonasFromSourcesLocked(); err != nil {
		return &SyncFailedError{Err: err}
	}
	return nil
}

// WritePersonaSourcesFile writes persona sources data using the current schema version.
func (r *FileRepository) WritePersonaSourcesFile(f PersonaSourcesFile) error {
	f.SchemaVersion = SchemaVersion
	return r.WriteAtomic(filepath.Join(r.dir, PersonaSourcesFilename), f)
}
