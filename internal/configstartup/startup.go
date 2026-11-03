// Package configstartup prepares and validates configuration for application startup.
package configstartup

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BlackGiack/kankuro/internal/fsutil"
)

const (
	layoutMarker    = "layout-version"
	currentMarker   = "2-ready\n"
	jsonExtension   = ".json"
	privateFileMode = 0o600
)

// PrepareStartup validates a current layout or initializes a genuinely empty root.
// The caller must hold the config-root lock throughout startup and runtime.
func PrepareStartup(dir string) error {
	_, err := os.Lstat(filepath.Join(dir, layoutMarker))
	if err == nil {
		return ValidateLayout(dir)
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
				"configuration has no current layout marker; select an empty config directory or restore a complete current-format configuration",
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
	return ValidateLayout(dir)
}
