package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

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

// FileStamp is a cheap, content-agnostic marker of a file's on-disk state:
// its modification time and size. Two stamps that compare equal mean "very
// likely unchanged" - see Fingerprint's doc for the residual mtime-precision
// caveat this accepts instead of hashing file contents.
type FileStamp struct {
	ModTime time.Time
	Size    int64
}

// registryFiles lists the files refreshConfig actually reads. Kept next to
// Fingerprint (not in the caller) so the coordinator never has to know
// filenames - see debt_impl_plan_5.md step 1.
func registryFiles() []string {
	return []string{
		IDCardsFilename,
		SkillsFilename,
		PersonasFilename,
		MCPsFilename,
		WorkspacesFilename,
		PuppetsFilename,
	}
}

// Fingerprint stats every registry file refreshConfig reads and returns a
// mtime+size marker for each, keyed by filename. It never opens or parses a
// file - just one [os.Stat] per registry, so a caller can decide whether a
// refresh is worth doing without paying for the read+unmarshal+sort it might
// skip.
//
// Two Fingerprints comparing equal (see FileStamp) is a "probably unchanged"
// signal, not a guarantee: mtime granularity is 1s on some filesystems, so a
// write landing in the same second as the previous stat can be missed. That
// residual is accepted - every in-app mutation already calls Fingerprint
// itself and sets the resulting baseline directly (see
// coordinator.refreshDataSeeded), so a same-second write there is reflected
// immediately regardless of the stat; only out-of-band hand-edits actually
// depend on a later tick's stat noticing the change, and picking those up a
// tick late is fine.
//
// A missing registry file (EnsureFiles only runs at startup, so a user can
// hand-delete one while the app is running) records a stable zero FileStamp
// rather than failing the whole call: an error here used to return a nil
// stamps map, which never compares equal to anything, silently disabling
// this optimization - every registry got re-read on every tick, for the rest
// of the session, with no signal that it had happened. A zero stamp compares
// equal to itself tick to tick and differs from a real stat the moment the
// file reappears. Any other stat failure (e.g. a permissions error) is still
// returned as a real error.
func (r *FileRepository) Fingerprint() (map[string]FileStamp, error) {
	files := registryFiles()
	stamps := make(map[string]FileStamp, len(files))
	for _, name := range files {
		info, err := os.Stat(filepath.Join(r.dir, name))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				stamps[name] = FileStamp{}
				continue
			}
			return nil, fmt.Errorf("stat %s: %w", name, err)
		}
		stamps[name] = FileStamp{ModTime: info.ModTime(), Size: info.Size()}
	}
	return stamps, nil
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
