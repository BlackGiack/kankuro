package dashboardmigration

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BlackGiack/kankuro/internal/fsutil"
)

const layoutUsedMarker = ".dashboard-layout-used"

// DiscardIncompleteStage removes only a stage that never published a manifest.
// The caller must hold the shared lock and explicitly request rollback.
func DiscardIncompleteStage(dir string) error {
	stage := filepath.Join(dir, stageDirectory)
	info, err := os.Lstat(stage)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return errors.New("migration stage is not a directory")
	}
	if _, errAt25 := os.Lstat(filepath.Join(dir, layoutMarker)); errAt25 == nil {
		return errors.New("layout marker exists")
	} else if !errors.Is(errAt25, os.ErrNotExist) {
		return errAt25
	}
	if _, errAt30 := os.Lstat(filepath.Join(stage, "manifest.json")); errAt30 == nil {
		return errors.New("migration manifest exists; use staged rollback")
	} else if !errors.Is(errAt30, os.ErrNotExist) {
		return errAt30
	}
	marker, err := os.ReadFile(filepath.Join(stage, stageSentinel))
	if errors.Is(err, os.ErrNotExist) {
		return os.Remove(stage)
	}
	if err != nil {
		return err
	}
	if string(marker) != "kankuro-dashboard-migration-v1\n" {
		return errors.New("migration stage ownership is invalid")
	}
	return os.RemoveAll(stage)
}

// RollbackStaged restores the exact staged backups before any new-layout use.
// The caller must hold the shared lock; startup must record use before writing.
//
//nolint:gocognit // Ordered durable rollback and recovery checks form one transaction.
func RollbackStaged(dir string) error {
	if _, err := os.Lstat(filepath.Join(dir, layoutUsedMarker)); err == nil {
		return errors.New("new layout has been used; automatic rollback is unsafe")
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	stage := filepath.Join(dir, stageDirectory)
	manifest, err := readManifest(stage)
	if err != nil {
		return err
	}
	if guardErr := verifyHistoryGuards(dir, manifest.Guards); guardErr != nil {
		return guardErr
	}
	if errAt29 := validateRollback(dir, stage, manifest); errAt29 != nil {
		return errAt29
	}
	// Close the runtime fence durably before marking rollback in the migration
	// manifest. Runtime intentionally does not inspect that manifest.
	marker := filepath.Join(dir, layoutMarker)
	if _, statErr := os.Lstat(marker); statErr == nil {
		if fenceErr := fsutil.WriteFileAtomic(marker, []byte("rollback\n"), privateFileMode); fenceErr != nil {
			return fenceErr
		}
	} else if !errors.Is(statErr, os.ErrNotExist) {
		return statErr
	}
	manifest.RollingBack = true
	if errAt33 := writeManifest(stage, manifest); errAt33 != nil {
		return errAt33
	}
	if markerErr := os.Remove(marker); markerErr != nil && !errors.Is(markerErr, os.ErrNotExist) {
		return markerErr
	}
	for _, item := range manifest.Files {
		path := filepath.Join(dir, item.Path)
		if !item.Existed {
			if errAt39 := os.Remove(path); errAt39 != nil && !errors.Is(errAt39, os.ErrNotExist) {
				return errAt39
			}
			continue
		}
		backup, errAt44 := os.ReadFile(filepath.Join(stage, "backup", item.Path))
		if errAt44 != nil {
			return errAt44
		}
		if errAt48 := fsutil.WriteFileAtomic(path, backup, os.FileMode(item.OriginalMode)); errAt48 != nil {
			return errAt48
		}
	}
	return os.RemoveAll(stage)
}

//nolint:gocognit // Check all possible interruption states before restoring anything.
func validateRollback(dir, stage string, manifest StageManifest) error {
	seen := make(map[string]bool)
	for _, item := range manifest.Files {
		if !validStagePath(item.Path) || seen[item.Path] {
			return fmt.Errorf("invalid or duplicate staged path %q", item.Path)
		}
		seen[item.Path] = true
		output, err := os.ReadFile(filepath.Join(stage, "output", item.Path))
		if err != nil || digest(output) != item.OutputSHA256 {
			return fmt.Errorf("invalid staged output %s", item.Path)
		}
		if item.Existed {
			backup, errAt68 := os.ReadFile(filepath.Join(stage, "backup", item.Path))
			if errAt68 != nil || digest(backup) != item.OriginalSHA256 {
				return fmt.Errorf("invalid backup %s", item.Path)
			}
		}
		current, err := os.ReadFile(filepath.Join(dir, item.Path))
		if errors.Is(err, os.ErrNotExist) && !item.Existed {
			continue
		}
		if err != nil {
			return fmt.Errorf("read source %s: %w", item.Path, err)
		}
		hash := digest(current)
		if hash != item.OutputSHA256 && (!item.Existed || hash != item.OriginalSHA256) {
			return fmt.Errorf("source changed since staging: %s", item.Path)
		}
	}
	if !seen["id_cards.yaml"] || !seen[dashboardFileName] || !seen[tombstonesFileName] || !seen[dashboardTombstonesFileName] {
		return errors.New("migration manifest missing required files")
	}
	entries, err := os.ReadDir(filepath.Join(dir, "sessions"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == jsonExtension && !seen[filepath.Join("sessions", entry.Name())] {
			return fmt.Errorf("session added since staging: %s", entry.Name())
		}
	}
	return nil
}
