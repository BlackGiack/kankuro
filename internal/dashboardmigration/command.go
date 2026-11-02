package dashboardmigration

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/BlackGiack/kankuro/internal/layout"
)

// Apply stages and publishes a legacy layout or resumes a staged publication.
// The caller must hold the shared config-root lock for the entire operation.
//
//nolint:gocognit // Handle old and new publication checkpoints without replaying a used layout.
func Apply(dir string) (string, bool, error) {
	backup := filepath.Join(dir, stageDirectory, "backup")
	//nolint:nestif // Resume or verify an already published migration.
	if _, err := os.Lstat(filepath.Join(dir, layoutMarker)); err == nil {
		manifest, errAt14 := readManifest(filepath.Join(dir, stageDirectory))
		if errAt14 != nil {
			return "", false, errAt14
		}
		marker, markerErr := os.ReadFile(filepath.Join(dir, layoutMarker))
		if markerErr != nil {
			return "", false, markerErr
		}
		_, usedErr := os.Lstat(filepath.Join(dir, layoutUsedMarker))
		if usedErr != nil && !errors.Is(usedErr, os.ErrNotExist) {
			return "", false, usedErr
		}
		if usedErr == nil && !manifest.Complete {
			return "", false, errors.New("used layout has an incomplete migration manifest")
		}
		if !manifest.Complete || (string(marker) == "2\n" && errors.Is(usedErr, os.ErrNotExist)) {
			if errAt19 := PublishStaged(dir); errAt19 != nil {
				return "", false, errAt19
			}
			if errAt22 := layout.ValidateLayout(dir); errAt22 != nil {
				return "", false, errAt22
			}
			return backup, false, nil
		}
		if errAt27 := layout.ValidateLayout(dir); errAt27 != nil {
			return "", false, errAt27
		}
		return backup, true, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", false, err
	}
	if _, err := os.Lstat(filepath.Join(dir, stageDirectory)); errors.Is(err, os.ErrNotExist) {
		if _, errAt35 := StageLegacy(dir); errAt35 != nil {
			return "", false, errAt35
		}
	} else if err != nil {
		return "", false, err
	}
	if err := PublishStaged(dir); err != nil {
		return "", false, err
	}
	if err := layout.ValidateLayout(dir); err != nil {
		return "", false, err
	}
	return backup, false, nil
}

// Rollback restores a published or partially published layout before first use,
// or discards an incomplete stage that never had a manifest.
func Rollback(dir string) error {
	if _, err := os.Lstat(filepath.Join(dir, stageDirectory, "manifest.json")); err == nil {
		return RollbackStaged(dir)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return DiscardIncompleteStage(dir)
}
