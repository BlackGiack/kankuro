package dashboardmigration

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/BlackGiack/kankuro/internal/fsutil"
)

const stageDirectory = ".dashboard-migration"
const stageSentinel = "staging-owner"

// StagedFile identifies an output and, when present, its exact original bytes.
type StagedFile struct {
	Path           string `json:"path"`
	OriginalSHA256 string `json:"original_sha256,omitempty"`
	OriginalMode   uint32 `json:"original_mode,omitempty"`
	OutputSHA256   string `json:"output_sha256"`
	Existed        bool   `json:"existed"`
	Published      bool   `json:"published,omitempty"`
}

// StageManifest describes complete staged data. Publishing and rollback are
// separate operations and must never assume staging itself changed the layout.
type StageManifest struct {
	Version     int               `json:"version"`
	Files       []StagedFile      `json:"files"`
	Guards      map[string]string `json:"guards"`
	Complete    bool              `json:"complete,omitempty"`
	RollingBack bool              `json:"rolling_back,omitempty"`
}

// StageLegacy validates and stages migration outputs and exact backups without
// replacing source files. The caller must hold the config-root lock, and must
// ensure application startup uses the same lock before exposing this operation.
//
//nolint:gocognit,nonamedreturns,funlen // Keep ordered staging visible and join cleanup errors.
func StageLegacy(dir string) (manifest StageManifest, err error) {
	if _, errAt39 := InspectLegacy(dir); errAt39 != nil {
		return StageManifest{}, errAt39
	}
	guards, guardErr := historyGuards(dir)
	if guardErr != nil {
		return StageManifest{}, guardErr
	}
	outputs, originals, err := prepareOutputs(dir)
	if err != nil {
		return StageManifest{}, err
	}
	stage := filepath.Join(dir, stageDirectory)
	if errAt47 := os.Mkdir(stage, 0o700); errAt47 != nil {
		return StageManifest{}, fmt.Errorf("create migration staging directory: %w", errAt47)
	}
	complete := false
	defer func() {
		if !complete {
			err = errors.Join(err, os.RemoveAll(stage))
		}
	}()
	if sentinelErr := fsutil.WriteFileAtomic(
		filepath.Join(stage, stageSentinel),
		[]byte("kankuro-dashboard-migration-v1\n"),
		privateFileMode,
	); sentinelErr != nil {
		return StageManifest{}, sentinelErr
	}
	manifest.Version = 1
	manifest.Guards = guards
	paths := make([]string, 0, len(outputs))
	for path := range outputs {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		original, readErr := os.ReadFile(filepath.Join(dir, path))
		existed := readErr == nil
		if readErr != nil && !errors.Is(readErr, os.ErrNotExist) {
			return StageManifest{}, readErr
		}
		expected, wasPresent := originals[path]
		if existed != wasPresent || (existed && digest(original) != digest(expected)) {
			return StageManifest{}, fmt.Errorf("source changed while staging: %s", path)
		}
		item := StagedFile{Path: path, Existed: existed, OutputSHA256: digest(outputs[path])}
		if existed {
			info, statErr := os.Lstat(filepath.Join(dir, path))
			if statErr != nil {
				return StageManifest{}, statErr
			}
			if !info.Mode().IsRegular() {
				return StageManifest{}, fmt.Errorf("source %s is not a regular file", path)
			}
			item.OriginalMode = uint32(info.Mode().Perm())
			item.OriginalSHA256 = digest(original)
			if errAt75 := stageFile(stage, "backup", path, original); errAt75 != nil {
				return StageManifest{}, errAt75
			}
		}
		if errAt79 := stageFile(stage, "output", path, outputs[path]); errAt79 != nil {
			return StageManifest{}, errAt79
		}
		manifest.Files = append(manifest.Files, item)
	}
	if _, inspectErr := InspectLegacy(dir); inspectErr != nil {
		return StageManifest{}, inspectErr
	}
	if verifyErr := verifyHistoryGuards(dir, guards); verifyErr != nil {
		return StageManifest{}, verifyErr
	}
	raw, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return StageManifest{}, err
	}
	if errAt88 := fsutil.WriteFileAtomic(filepath.Join(stage, "manifest.json"), raw, privateFileMode); errAt88 != nil {
		return StageManifest{}, errAt88
	}
	complete = true
	return manifest, nil
}

func prepareOutputs(dir string) (map[string][]byte, map[string][]byte, error) {
	outputs := map[string][]byte{
		tombstonesFileName:          []byte("schema_version: 1\ntombstones: []\n"),
		dashboardTombstonesFileName: []byte("schema_version: 1\ntombstones: []\n"),
		dashboardFileName: []byte(
			"schema_version: 1\ndashboards:\n  - id: default\n    name: Default\n    description: \"\"\n    state: active\n",
		),
	}
	originals := make(map[string][]byte)
	cardPath := filepath.Join(dir, "id_cards.yaml")
	cards, err := os.ReadFile(cardPath)
	switch {
	case errors.Is(err, os.ErrNotExist):
		cards = []byte("schema_version: 1\nid_cards: []\n")
	case err != nil:
		return nil, nil, err
	default:
		originals["id_cards.yaml"] = cards
	}
	outputs["id_cards.yaml"], err = TransformCards(cards)
	if err != nil {
		return nil, nil, fmt.Errorf("prepare cards: %w", err)
	}
	entries, err := os.ReadDir(filepath.Join(dir, "sessions"))
	if errors.Is(err, os.ErrNotExist) {
		return outputs, originals, nil
	}
	if err != nil {
		return nil, nil, err
	}
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != jsonExtension {
			continue
		}
		path := filepath.Join("sessions", entry.Name())
		raw, errAt127 := os.ReadFile(filepath.Join(dir, path))
		if errAt127 != nil {
			return nil, nil, errAt127
		}
		originals[path] = raw
		outputs[path], errAt127 = TransformSession(raw)
		if errAt127 != nil {
			return nil, nil, fmt.Errorf("prepare %s: %w", path, errAt127)
		}
	}
	return outputs, originals, nil
}

func stageFile(stage, kind, path string, data []byte) error {
	destination := filepath.Join(stage, kind, path)
	if err := os.MkdirAll(filepath.Dir(destination), 0o700); err != nil {
		return err
	}
	return fsutil.WriteFileAtomic(destination, data, privateFileMode)
}

func digest(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
