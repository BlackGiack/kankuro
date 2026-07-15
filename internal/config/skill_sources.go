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
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.listSkillSourcesLocked()
}

func (r *FileRepository) listSkillSourcesLocked() ([]string, error) {
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

// SyncSkillsFromSources rebuilds the skills registry by scanning configured source folders
// and merging built-in skills so they are never removed by a refresh.
func (r *FileRepository) SyncSkillsFromSources() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.syncSkillsFromSourcesLocked()
}

func (r *FileRepository) syncSkillsFromSourcesLocked() error {
	sources, err := r.listSkillSourcesLocked()
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
	// Merge built-in skills so they persist across refreshes
	skills = MergeBuiltinSkills(skills)
	return r.WriteSkillsFile(SkillsFile{SchemaVersion: SchemaVersion, Skills: skills})
}

func SkillIDFromSource(path string) string {
	sum := sha256.Sum256([]byte(path))
	return hex.EncodeToString(sum[:])
}

// SaveSkillSources writes the list of skill sources to disk.
func (r *FileRepository) SaveSkillSources(paths []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if paths == nil {
		paths = []string{}
	}
	return r.WriteSkillSourcesFile(SkillsSourcesFile{
		SchemaVersion: SchemaVersion,
		Paths:         paths,
	})
}

// UpdateSkillSources reads the current skill sources, applies fn, persists
// the result, and re-syncs the skill registry - all under one lock
// acquisition. This closes the read-modify-write race a caller doing
// ListSkillSources -> mutate -> SaveSkillSources across three separate lock
// acquisitions is exposed to: two concurrent callers can each read the same
// starting list and one save clobbers the other's.
//
// If fn's mutated list itself fails to persist, that error is returned
// as-is. If the write succeeds but the subsequent registry resync fails
// (e.g. a newly added source directory does not exist), the error is
// returned wrapped in *SyncFailedError so callers can tell "nothing was
// saved" from "the source list saved but skills.yaml is now stale" and
// still refresh the UI in the latter case.
func (r *FileRepository) UpdateSkillSources(fn func([]string) ([]string, error)) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	current, err := r.listSkillSourcesLocked()
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
	if err = r.WriteSkillSourcesFile(SkillsSourcesFile{SchemaVersion: SchemaVersion, Paths: updated}); err != nil {
		return err
	}
	if err = r.syncSkillsFromSourcesLocked(); err != nil {
		return &SyncFailedError{Err: err}
	}
	return nil
}

// WriteSkillSourcesFile writes skill sources data using the current schema version.
func (r *FileRepository) WriteSkillSourcesFile(f SkillsSourcesFile) error {
	f.SchemaVersion = SchemaVersion
	return r.WriteAtomic(filepath.Join(r.dir, SkillsSourcesFilename), f)
}
