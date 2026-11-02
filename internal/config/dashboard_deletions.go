package config

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/BlackGiack/kankuro/internal/fsutil"
)

const dashboardDeletionsDir = "dashboard_deletions"

// DashboardDeletionSession records cleanup identity and observed effects.
type DashboardDeletionSession struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Status           string `json:"status"`
	WorkingDir       string `json:"working_dir"`
	TemporaryRoot    string `json:"temporary_root,omitempty"`
	UseTmpWorkspace  bool   `json:"use_tmp_workspace"`
	Stopped          bool   `json:"stopped"`
	Removed          bool   `json:"removed"`
	RecordDeleted    bool   `json:"record_deleted"`
	TemporaryRemoved bool   `json:"temporary_removed"`
	Error            string `json:"error,omitempty"`
}

// DashboardDeletionReference describes an affected current or saved card reference.
type DashboardDeletionReference struct {
	DashboardID string `json:"dashboard_id"`
	SessionID   string `json:"session_id"`
	Current     bool   `json:"current"`
	Binding     bool   `json:"binding"`
	Pending     bool   `json:"pending"`
}

// DashboardDeletionCard records a card owned by the deleting dashboard.
type DashboardDeletionCard struct {
	ID         string                       `json:"id"`
	Name       string                       `json:"name"`
	Revision   string                       `json:"revision"`
	Shared     bool                         `json:"shared"`
	References []DashboardDeletionReference `json:"references"`
	Deleted    bool                         `json:"deleted"`
	Error      string                       `json:"error,omitempty"`
}

// DashboardDeletionPending identifies an activation reserved before the intent.
type DashboardDeletionPending struct {
	CardID        string `json:"card_id"`
	DashboardID   string `json:"dashboard_id"`
	SessionID     string `json:"session_id"`
	WorkingDir    string `json:"working_dir,omitempty"`
	TemporaryRoot string `json:"temporary_root,omitempty"`
	Cleaned       bool   `json:"cleaned"`
	Error         string `json:"error,omitempty"`
}

// DashboardDeletion records the durable work needed to finish a cascade.
type DashboardDeletion struct {
	SchemaVersion     int                        `json:"schema_version"`
	DashboardID       string                     `json:"dashboard_id"`
	Name              string                     `json:"name"`
	Description       string                     `json:"description"`
	Sessions          []DashboardDeletionSession `json:"sessions"`
	Cards             []DashboardDeletionCard    `json:"cards"`
	Pending           []DashboardDeletionPending `json:"pending"`
	AdmissionsDrained bool                       `json:"admissions_drained"`
	HistoryRemoved    bool                       `json:"history_removed"`
}

// ListDashboardDeletions rejects corrupt manifests rather than hiding them.
func (r *FileRepository) ListDashboardDeletions() ([]DashboardDeletion, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.listDashboardDeletionsLocked()
}

func (r *FileRepository) listDashboardDeletionsLocked() ([]DashboardDeletion, error) {
	dir := filepath.Join(r.dir, dashboardDeletionsDir)
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return []DashboardDeletion{}, nil
	}
	if err != nil {
		return nil, err
	}
	var manifests []DashboardDeletion
	seen := make(map[string]bool)
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		if !entry.Type().IsRegular() {
			return nil, fmt.Errorf("dashboard deletion %s is not a regular file", entry.Name())
		}
		raw, readErr := os.ReadFile(filepath.Join(dir, entry.Name()))
		if readErr != nil {
			return nil, readErr
		}
		var manifest DashboardDeletion
		if decodeErr := json.Unmarshal(raw, &manifest); decodeErr != nil {
			return nil, fmt.Errorf("invalid dashboard deletion %s: %w", entry.Name(), decodeErr)
		}
		if validationErr := validateDashboardDeletion(manifest); validationErr != nil {
			return nil, validationErr
		}
		if seen[manifest.DashboardID] || entry.Name() != dashboardDeletionFilename(manifest.DashboardID) {
			return nil, errors.New("duplicate or mismatched dashboard deletion")
		}
		seen[manifest.DashboardID] = true
		manifests = append(manifests, manifest)
	}
	return manifests, nil
}

// BeginDashboardDeletion persists a new intent. The caller must coordinate
// authorization and admission with the manager before writing it.
func (r *FileRepository) BeginDashboardDeletion(manifest DashboardDeletion) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := validateDashboardDeletion(manifest); err != nil {
		return err
	}
	manifests, err := r.listDashboardDeletionsLocked()
	if err != nil {
		return err
	}
	for _, existing := range manifests {
		if existing.DashboardID == manifest.DashboardID {
			return errors.New("dashboard deletion already exists")
		}
	}
	return r.writeDashboardDeletionLocked(manifest)
}

// SaveDashboardDeletion updates an existing durable deletion manifest.
func (r *FileRepository) SaveDashboardDeletion(manifest DashboardDeletion) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := validateDashboardDeletion(manifest); err != nil {
		return err
	}
	manifests, err := r.listDashboardDeletionsLocked()
	if err != nil {
		return err
	}
	for _, existing := range manifests {
		if existing.DashboardID == manifest.DashboardID {
			return r.writeDashboardDeletionLocked(manifest)
		}
	}
	return errors.New("dashboard deletion manifest not found")
}

func (r *FileRepository) writeDashboardDeletionLocked(manifest DashboardDeletion) error {
	dir := filepath.Join(r.dir, dashboardDeletionsDir)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return fsutil.WriteFileAtomic(filepath.Join(dir, dashboardDeletionFilename(manifest.DashboardID)), raw, configFileMode)
}

// FinalizeDashboardDeletion removes the dashboard record only after every
// recorded cascade effect is durable, then removes its recovery manifest.
//
//nolint:gocognit // Validate each durable cleanup checkpoint before finalization.
func (r *FileRepository) FinalizeDashboardDeletion(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	manifests, err := r.listDashboardDeletionsLocked()
	if err != nil {
		return err
	}
	var selected *DashboardDeletion
	for index := range manifests {
		if manifests[index].DashboardID == id {
			selected = &manifests[index]
			break
		}
	}
	if selected == nil {
		return errors.New("dashboard deletion manifest not found")
	}
	if !selected.AdmissionsDrained || !selected.HistoryRemoved {
		return errors.New("dashboard deletion is incomplete")
	}
	for _, item := range selected.Sessions {
		if !item.Removed || !item.Stopped || !item.RecordDeleted || !item.TemporaryRemoved {
			return errors.New("dashboard session cleanup is incomplete")
		}
	}
	for _, item := range selected.Cards {
		if !item.Deleted {
			return errors.New("dashboard card cleanup is incomplete")
		}
	}
	for _, item := range selected.Pending {
		if !item.Cleaned {
			return errors.New("pending activation cleanup is incomplete")
		}
	}
	if errAt211 := r.retainDashboardTombstoneLocked(id); errAt211 != nil {
		return errAt211
	}
	dashboards, err := r.loadDashboardsLocked()
	if err != nil {
		return err
	}
	dashboards.Dashboards = slices.DeleteFunc(dashboards.Dashboards, func(item Dashboard) bool { return item.ID == id })
	if errAt219 := r.WriteAtomic(filepath.Join(r.dir, DashboardsFilename), dashboards); errAt219 != nil {
		return errAt219
	}
	dir := filepath.Join(r.dir, dashboardDeletionsDir)
	if errAt223 := os.Remove(filepath.Join(dir, dashboardDeletionFilename(id))); errAt223 != nil {
		return errAt223
	}
	directory, err := os.Open(dir)
	syncErr := err
	if err == nil {
		syncErr = errors.Join(directory.Sync(), directory.Close())
	}
	if syncErr == nil {
		return nil
	}
	raw, marshalErr := json.MarshalIndent(selected, "", "  ")
	if marshalErr != nil {
		return errors.Join(syncErr, marshalErr)
	}
	restoreErr := fsutil.WriteFileAtomic(filepath.Join(dir, dashboardDeletionFilename(id)), raw, configFileMode)
	return errors.Join(syncErr, restoreErr)
}

func validateDashboardDeletion(manifest DashboardDeletion) error {
	if manifest.SchemaVersion != 1 || strings.TrimSpace(manifest.DashboardID) == "" || strings.TrimSpace(manifest.Name) == "" {
		return errors.New("invalid dashboard deletion manifest")
	}
	sessions := make(map[string]bool)
	for _, item := range manifest.Sessions {
		if item.ID == "" || sessions[item.ID] ||
			(item.UseTmpWorkspace && !validTemporaryIdentity(item.ID, item.TemporaryRoot, item.WorkingDir)) {
			return errors.New("invalid dashboard deletion session")
		}
		sessions[item.ID] = true
	}
	cards := make(map[string]bool)
	for _, item := range manifest.Cards {
		if item.ID == "" || cards[item.ID] {
			return errors.New("invalid dashboard deletion card")
		}
		cards[item.ID] = true
	}
	for _, item := range manifest.Pending {
		if item.CardID == "" || item.DashboardID == "" || item.SessionID == "" || ((item.WorkingDir == "") != (item.TemporaryRoot == "")) ||
			(item.WorkingDir != "" && !validTemporaryIdentity(item.SessionID, item.TemporaryRoot, item.WorkingDir)) {
			return errors.New("invalid pending dashboard activation")
		}
	}
	return nil
}

func validTemporaryIdentity(id, root, workingDir string) bool {
	return id != "" && id != "." && id != ".." && filepath.IsAbs(root) && filepath.IsAbs(workingDir) &&
		filepath.Base(filepath.Clean(root)) == id && filepath.Dir(filepath.Clean(workingDir)) == filepath.Clean(root)
}

func dashboardDeletionFilename(id string) string {
	hash := sha256.Sum256([]byte(id))
	return hex.EncodeToString(hash[:]) + ".json"
}
