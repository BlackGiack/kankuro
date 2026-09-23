package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

// configFileMode is the permission mode for every config file this package writes.
const configFileMode = 0o600

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
// It writes to a uniquely named temp file in the same directory (so two
// concurrent writers never share a temp file), fsyncs the data before the
// rename so a crash cannot leave a truncated file visible under the real
// name, and removes the temp file on any error path.
func (r *FileRepository) WriteAtomic(path string, data any) error {
	raw, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal %s: %w", filepath.Base(path), err)
	}

	dir := filepath.Dir(path)
	f, err := os.CreateTemp(dir, "."+filepath.Base(path)+".*.tmp")
	if err != nil {
		return fmt.Errorf("create temp %s: %w", filepath.Base(path), err)
	}
	tmpName := f.Name()
	committed := false
	defer func() {
		if !committed {
			_ = os.Remove(tmpName)
		}
	}()

	if err = f.Chmod(configFileMode); err != nil {
		_ = f.Close()
		return fmt.Errorf("chmod temp %s: %w", filepath.Base(path), err)
	}
	if _, err = f.Write(raw); err != nil {
		_ = f.Close()
		return fmt.Errorf("write temp %s: %w", filepath.Base(path), err)
	}
	if err = f.Sync(); err != nil {
		_ = f.Close()
		return fmt.Errorf("sync temp %s: %w", filepath.Base(path), err)
	}
	if err = f.Close(); err != nil {
		return fmt.Errorf("close temp %s: %w", filepath.Base(path), err)
	}
	if err = os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("rename temp %s: %w", filepath.Base(path), err)
	}
	committed = true

	if dirFile, dirErr := os.Open(dir); dirErr == nil {
		_ = dirFile.Sync()
		_ = dirFile.Close()
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
