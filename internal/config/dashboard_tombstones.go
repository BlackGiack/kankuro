package config

import (
	"errors"
	"os"
	"path/filepath"
	"time"
)

const DashboardTombstonesFilename = "dashboard_tombstones.yaml"

// DashboardTombstone retains the immutable identity of a deleted dashboard.
type DashboardTombstone struct {
	ID        string    `yaml:"id"`
	DeletedAt time.Time `yaml:"deleted_at"`
}

type dashboardTombstonesFile struct {
	SchemaVersion int                  `yaml:"schema_version"`
	Tombstones    []DashboardTombstone `yaml:"tombstones"`
}

// ListDashboardTombstones returns retained deleted-dashboard identities.
func (r *FileRepository) ListDashboardTombstones() ([]DashboardTombstone, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	file, err := r.loadDashboardTombstonesLocked()
	if err != nil {
		return nil, err
	}
	return file.Tombstones, nil
}

func (r *FileRepository) loadDashboardTombstonesLocked() (dashboardTombstonesFile, error) {
	var file dashboardTombstonesFile
	if err := ReadYAML(filepath.Join(r.dir, DashboardTombstonesFilename), &file); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return dashboardTombstonesFile{SchemaVersion: 1, Tombstones: []DashboardTombstone{}}, nil
		}
		return dashboardTombstonesFile{}, err
	}
	if file.SchemaVersion != 1 {
		return dashboardTombstonesFile{}, errors.New("unsupported dashboard tombstone schema")
	}
	seen := make(map[string]bool)
	for _, item := range file.Tombstones {
		if item.ID == "" || seen[item.ID] {
			return dashboardTombstonesFile{}, errors.New("invalid or duplicate dashboard tombstone")
		}
		seen[item.ID] = true
	}
	return file, nil
}

func (r *FileRepository) retainDashboardTombstoneLocked(id string) error {
	file, err := r.loadDashboardTombstonesLocked()
	if err != nil {
		return err
	}
	for _, item := range file.Tombstones {
		if item.ID == id {
			return nil
		}
	}
	file.Tombstones = append(file.Tombstones, DashboardTombstone{ID: id, DeletedAt: time.Now().UTC()})
	return r.WriteAtomic(filepath.Join(r.dir, DashboardTombstonesFilename), file)
}
