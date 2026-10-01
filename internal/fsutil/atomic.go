// Package fsutil holds small, dependency-free file-system helpers shared by
// the config and session packages.
package fsutil

import (
	"fmt"
	"os"
	"path/filepath"
)

// WriteFileAtomic writes data to path using temp-file + rename for crash
// safety: a crash at any point leaves either the old file or the complete new
// one under path, never a truncated one.
//
// It writes to a uniquely named temp file in the same directory (so two
// concurrent writers never share a temp file), fsyncs the data before the
// rename so a crash cannot leave a truncated file visible under the real
// name, and removes the temp file on any error path. The temp file is named
// ".<base>.<random>.tmp", so a leftover from a crash does not match a
// "*.<ext>" glob of the real files.
//
// The directory must already exist; creating it is the caller's job.
func WriteFileAtomic(path string, data []byte, perm os.FileMode) error {
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

	if err = f.Chmod(perm); err != nil {
		_ = f.Close()
		return fmt.Errorf("chmod temp %s: %w", filepath.Base(path), err)
	}
	if _, err = f.Write(data); err != nil {
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
