// Package layout owns the current configuration layout and its runtime fence.
package layout

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BlackGiack/kankuro/internal/fsutil"
)

const (
	MarkerFilename     = "layout-version"
	UsedMarkerFilename = ".dashboard-layout-used"
	CurrentMarker      = "2-ready\n"
	layoutMarker       = MarkerFilename
	layoutUsedMarker   = UsedMarkerFilename
	jsonExtension      = ".json"
	currentMarker      = CurrentMarker
	privateFileMode    = 0o600
)

// PrepareStartup validates a current layout or initializes a genuinely empty root.
// The caller must hold the config-root lock throughout startup and runtime.
func PrepareStartup(dir string) error {
	_, err := os.Lstat(filepath.Join(dir, layoutMarker))
	if err == nil {
		if err = ValidateLayout(dir); err != nil {
			return err
		}
		return MarkLayoutUsed(dir)
	}
	if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.Name() != ".dashboard-layout.lock" {
			return errors.New(
				"configuration has no completed layout marker; migrate a legacy root offline or repair an interrupted initialization",
			)
		}
	}
	initial := map[string]string{
		"id_cards.yaml":             "schema_version: 2\nid_cards: []\n",
		"dashboards.yaml":           "schema_version: 1\ndashboards:\n  - id: default\n    name: Default\n    description: \"\"\n    state: active\n",
		"card_tombstones.yaml":      "schema_version: 1\ntombstones: []\n",
		"dashboard_tombstones.yaml": "schema_version: 1\ntombstones: []\n",
	}
	for name, content := range initial {
		if err = fsutil.WriteFileAtomic(filepath.Join(dir, name), []byte(content), privateFileMode); err != nil {
			return fmt.Errorf("initialize %s: %w", name, err)
		}
	}
	if err = fsutil.WriteFileAtomic(filepath.Join(dir, layoutMarker), []byte(currentMarker), privateFileMode); err != nil {
		return err
	}
	if err = ValidateLayout(dir); err != nil {
		return err
	}
	return MarkLayoutUsed(dir)
}

// MarkLayoutUsed permanently prevents rollback after a current layout is used.
func MarkLayoutUsed(dir string) error {
	path := filepath.Join(dir, layoutUsedMarker)
	if info, err := os.Lstat(path); err == nil {
		if !info.Mode().IsRegular() {
			return errors.New("invalid dashboard layout use marker")
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		if string(data) != "used\n" {
			return errors.New("invalid dashboard layout use marker")
		}
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return fsutil.WriteFileAtomic(path, []byte("used\n"), privateFileMode)
}
