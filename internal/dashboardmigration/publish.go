package dashboardmigration

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BlackGiack/kankuro/internal/fsutil"
	"github.com/BlackGiack/kankuro/internal/layout"
)

const layoutMarker = "layout-version"

// PublishStaged verifies staged files and current sources, then atomically
// replaces each source. Repeating it resumes after an interrupted replacement.
// The caller must hold the shared lock and startup must reject incomplete stages.
//
//nolint:gocognit // Publication ordering and its durable checkpoints must remain visible.
func PublishStaged(dir string) error {
	stage := filepath.Join(dir, stageDirectory)
	manifest, err := readManifest(stage)
	if err != nil {
		return err
	}
	if manifest.RollingBack {
		return errors.New("migration rollback is in progress")
	}
	if guardErr := verifyHistoryGuards(dir, manifest.Guards); guardErr != nil {
		return guardErr
	}
	if marker, errAt25 := os.ReadFile(filepath.Join(dir, layoutMarker)); errAt25 == nil {
		if string(marker) != layout.CurrentMarker && string(marker) != "2\n" {
			return errors.New("unexpected layout marker")
		}
	} else if !errors.Is(errAt25, os.ErrNotExist) {
		return errAt25
	}
	if errAt32 := validateStagedFiles(dir, stage, manifest); errAt32 != nil {
		return errAt32
	}
	for i := range manifest.Files {
		item := &manifest.Files[i]
		if item.Published {
			continue
		}
		output, errAt40 := os.ReadFile(filepath.Join(stage, "output", item.Path))
		if errAt40 != nil {
			return errAt40
		}
		if errAt44 := os.MkdirAll(filepath.Dir(filepath.Join(dir, item.Path)), 0o700); errAt44 != nil {
			return errAt44
		}
		if errAt47 := fsutil.WriteFileAtomic(filepath.Join(dir, item.Path), output, privateFileMode); errAt47 != nil {
			return errAt47
		}
		item.Published = true
		if errAt51 := writeManifest(stage, manifest); errAt51 != nil {
			return errAt51
		}
	}
	// The manifest checkpoint precedes the permanent fence. Runtime never needs
	// to read migration backups to distinguish a completed publication.
	manifest.Complete = true
	if checkpointErr := writeManifest(stage, manifest); checkpointErr != nil {
		return checkpointErr
	}
	marker := filepath.Join(dir, layoutMarker)
	//nolint:nestif // Upgrade the old fence only after the checkpoint is durable.
	if existing, errAt56 := os.ReadFile(marker); errAt56 == nil {
		if string(existing) != layout.CurrentMarker && string(existing) != "2\n" {
			return errors.New("unexpected layout marker")
		}
		if string(existing) == "2\n" {
			if markerErr := fsutil.WriteFileAtomic(marker, []byte(layout.CurrentMarker), privateFileMode); markerErr != nil {
				return markerErr
			}
		}
	} else if !errors.Is(errAt56, os.ErrNotExist) {
		return errAt56
	} else if errAt62 := fsutil.WriteFileAtomic(marker, []byte(layout.CurrentMarker), privateFileMode); errAt62 != nil {
		return errAt62
	}
	return nil
}

func readManifest(stage string) (StageManifest, error) {
	raw, err := os.ReadFile(filepath.Join(stage, "manifest.json"))
	if err != nil {
		return StageManifest{}, err
	}
	if validationErr := validateJSON(raw); validationErr != nil {
		return StageManifest{}, fmt.Errorf("invalid migration manifest: %w", validationErr)
	}
	var manifest StageManifest
	if errAt75 := json.Unmarshal(raw, &manifest); errAt75 != nil {
		return StageManifest{}, errAt75
	}
	if manifest.Version != 1 || len(manifest.Files) == 0 {
		return StageManifest{}, errors.New("invalid migration manifest")
	}
	return manifest, nil
}

func writeManifest(stage string, manifest StageManifest) error {
	raw, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return fsutil.WriteFileAtomic(filepath.Join(stage, "manifest.json"), raw, privateFileMode)
}

//nolint:gocognit // Verify every input before allowing the first publication effect.
func validateStagedFiles(dir, stage string, manifest StageManifest) error {
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
			backup, errAt104 := os.ReadFile(filepath.Join(stage, "backup", item.Path))
			if errAt104 != nil || digest(backup) != item.OriginalSHA256 {
				return fmt.Errorf("invalid backup %s", item.Path)
			}
		}
		current, err := os.ReadFile(filepath.Join(dir, item.Path))
		if errors.Is(err, os.ErrNotExist) && !item.Existed && !item.Published {
			continue
		}
		if err != nil {
			return fmt.Errorf("read source %s: %w", item.Path, err)
		}
		currentHash := digest(current)
		if currentHash != item.OutputSHA256 && (item.Published || !item.Existed || currentHash != item.OriginalSHA256) {
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

func validStagePath(path string) bool {
	if path == "id_cards.yaml" || path == dashboardFileName || path == tombstonesFileName || path == dashboardTombstonesFileName {
		return true
	}
	return filepath.IsLocal(path) && filepath.Dir(path) == "sessions" && strings.HasSuffix(path, jsonExtension)
}
