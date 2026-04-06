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
