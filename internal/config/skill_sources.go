package config

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
)

// ListSkillSources returns the configured root folders for skill discovery.
func (r *FileRepository) ListSkillSources() ([]string, error) {
	path := filepath.Join(r.dir, SkillsSourcesFilename)
	f := SkillsSourcesFile{}
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

// SyncSkillsFromSources rebuilds the skills registry by scanning configured source folders.
func (r *FileRepository) SyncSkillsFromSources() error {
	sources, err := r.ListSkillSources()
	if err != nil {
		return err
	}
	var skills []Skill
	err = ScanEntries(sources,
		func(entry os.DirEntry) bool {
			return entry.IsDir() && !strings.HasPrefix(entry.Name(), ".")
		},
		func(_, expanded string, entry os.DirEntry) error {
			fullPath := filepath.Join(expanded, entry.Name())
			skills = append(skills, Skill{
				ID:     SkillIDFromSource(fullPath),
				Name:   entry.Name(),
				Source: fullPath,
			})
			return nil
		},
	)
	if err != nil {
		return err
	}
	return r.WriteSkillsFile(SkillsFile{SchemaVersion: SchemaVersion, Skills: skills})
}

func SkillIDFromSource(path string) string {
	sum := sha256.Sum256([]byte(path))
	return hex.EncodeToString(sum[:])
}

// SaveSkillSources writes the list of skill sources to disk.
func (r *FileRepository) SaveSkillSources(paths []string) error {
	if paths == nil {
		paths = []string{}
	}
	return r.WriteSkillSourcesFile(SkillsSourcesFile{
		SchemaVersion: SchemaVersion,
		Paths:         paths,
	})
}

// WriteSkillSourcesFile writes skill sources data using the current schema version.
func (r *FileRepository) WriteSkillSourcesFile(f SkillsSourcesFile) error {
	f.SchemaVersion = SchemaVersion
	return r.WriteAtomic(filepath.Join(r.dir, SkillsSourcesFilename), f)
}
