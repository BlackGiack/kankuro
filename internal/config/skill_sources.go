package config

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ListSkillSources returns the configured root folders for skill discovery.
func (r *FileRepository) ListSkillSources() ([]string, error) {
	path := filepath.Join(r.dir, skillsSourcesFilename)
	f := skillsSourcesFile{}
	if err := readYAML(path, &f); err != nil {
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
	skills := make([]Skill, 0)
	for _, source := range sources {
		expanded, err := expandPath(source)
		if err != nil {
			return fmt.Errorf("expand skill source %q: %w", source, err)
		}
		entries, err := os.ReadDir(expanded)
		if err != nil {
			return fmt.Errorf("read skill source %q: %w", expanded, err)
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			fullPath := filepath.Join(expanded, entry.Name())
			skills = append(skills, Skill{
				ID:     skillIDFromSource(fullPath),
				Name:   entry.Name(),
				Source: fullPath,
			})
		}
	}
	return r.writeSkillsFile(skillsFile{SchemaVersion: SchemaVersion, Skills: skills})
}

func expandPath(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("path is empty")
	}
	value = os.ExpandEnv(value)
	if strings.HasPrefix(value, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		suffix := strings.TrimPrefix(value, "~")
		suffix = strings.TrimPrefix(suffix, string(os.PathSeparator))
		value = filepath.Join(home, suffix)
	}
	return filepath.Abs(value)
}

func skillIDFromSource(path string) string {
	sum := sha256.Sum256([]byte(path))
	return hex.EncodeToString(sum[:])
}
