package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

// EnsureYAML creates a file with default content if it does not already exist.
func (r *FileRepository) EnsureYAML(path string, data any) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return r.WriteAtomic(path, data)
}

// ReadYAML loads and unmarshals a YAML document into out.
func ReadYAML(path string, out any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", filepath.Base(path), err)
	}
	if err = yaml.Unmarshal(data, out); err != nil {
		return fmt.Errorf("parse %s: %w", filepath.Base(path), err)
	}
	return nil
}

// WriteAtomic persists YAML content using temp-file + rename for crash safety.
func (r *FileRepository) WriteAtomic(path string, data any) error {
	raw, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal %s: %w", filepath.Base(path), err)
	}
	tmp := path + ".tmp"
	if err = os.WriteFile(tmp, raw, 0o600); err != nil {
		return fmt.Errorf("write temp %s: %w", filepath.Base(path), err)
	}
	if err = os.Rename(tmp, path); err != nil {
		return fmt.Errorf("rename temp %s: %w", filepath.Base(path), err)
	}
	return nil
}

// UpsertInSlice adds or updates an item in a slice.
// getID and setID functions provide ID access.
func UpsertInSlice[T any](
	items []T,
	item T,
	getID func(T) string,
	setID func(*T, string),
) ([]T, T) {
	if getID(item) == "" {
		setID(&item, uuid.NewString())
	}
	for i := range items {
		if getID(items[i]) == getID(item) {
			items[i] = item
			return items, item
		}
	}
	return append(items, item), item
}
