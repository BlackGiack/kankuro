package config

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ListPersonaSources returns the configured root folders for persona discovery.
func (r *FileRepository) ListPersonaSources() ([]string, error) {
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
	sources, err := r.ListPersonaSources()
	if err != nil {
		return err
	}
	personas := make([]Persona, 0)
	for _, source := range sources {
		expanded, errExp := expandPath(source)
		if errExp != nil {
			return fmt.Errorf("expand persona source %q: %w", source, errExp)
		}
		entries, errRead := os.ReadDir(expanded)
		if errRead != nil {
			return fmt.Errorf("read persona source %q: %w", expanded, errRead)
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			if !strings.HasSuffix(strings.ToLower(entry.Name()), ".md") {
				continue
			}
			fullPath := filepath.Join(expanded, entry.Name())
			personas = append(personas, Persona{
				ID:     PersonaIDFromSource(fullPath),
				Name:   strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())),
				Source: fullPath,
			})
		}
	}
	return r.WritePersonasFile(PersonasFile{SchemaVersion: SchemaVersion, Personas: personas})
}

func PersonaIDFromSource(path string) string {
	sum := sha256.Sum256([]byte(path))
	return hex.EncodeToString(sum[:])
}

// SavePersonaSources writes the list of persona sources to disk.
func (r *FileRepository) SavePersonaSources(paths []string) error {
	if paths == nil {
		paths = []string{}
	}
	return r.WritePersonaSourcesFile(PersonaSourcesFile{
		SchemaVersion: SchemaVersion,
		Paths:         paths,
	})
}

// WritePersonaSourcesFile writes persona sources data using the current schema version.
func (r *FileRepository) WritePersonaSourcesFile(f PersonaSourcesFile) error {
	f.SchemaVersion = SchemaVersion
	return r.WriteAtomic(filepath.Join(r.dir, PersonaSourcesFilename), f)
}
