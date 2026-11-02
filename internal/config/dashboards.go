package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/uuid"
)

const (
	DashboardsFilename     = "dashboards.yaml"
	DashboardSchemaVersion = 1
	DashboardActive        = "active"
	DashboardDeleting      = "deleting"
)

// Dashboard is the durable identity and lifecycle state of a dashboard.
type Dashboard struct {
	ID          string `yaml:"id"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	State       string `yaml:"state"`
}

type dashboardsFile struct {
	SchemaVersion int         `yaml:"schema_version"`
	Dashboards    []Dashboard `yaml:"dashboards"`
}

// ListDashboards reads the dashboard inventory. A missing or invalid inventory
// is an error; callers must never infer a default dashboard from it.
func (r *FileRepository) ListDashboards() ([]Dashboard, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	f, err := r.loadDashboardsLocked()
	if err != nil {
		return nil, err
	}
	sort.Slice(f.Dashboards, func(i, j int) bool {
		if f.Dashboards[i].Name == f.Dashboards[j].Name {
			return f.Dashboards[i].ID < f.Dashboards[j].ID
		}
		return f.Dashboards[i].Name < f.Dashboards[j].Name
	})
	return f.Dashboards, nil
}

// CreateDashboard adds an active dashboard to an already initialized inventory.
func (r *FileRepository) CreateDashboard(d Dashboard) (Dashboard, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	f, err := r.loadDashboardsLocked()
	if err != nil {
		return Dashboard{}, err
	}
	if d.ID == "" {
		d.ID = uuid.NewString()
	}
	tombstones, err := r.loadDashboardTombstonesLocked()
	if err != nil {
		return Dashboard{}, err
	}
	for _, item := range tombstones.Tombstones {
		if item.ID == d.ID {
			return Dashboard{}, errors.New("dashboard ID is deleted and cannot be reused")
		}
	}
	d.State = DashboardActive
	if errAt63 := validateDashboard(d); errAt63 != nil {
		return Dashboard{}, errAt63
	}
	for _, existing := range f.Dashboards {
		if existing.ID == d.ID {
			return Dashboard{}, fmt.Errorf("dashboard id already exists: %s", d.ID)
		}
	}
	f.Dashboards = append(f.Dashboards, d)
	if errAt72 := r.WriteAtomic(filepath.Join(r.dir, DashboardsFilename), f); errAt72 != nil {
		return Dashboard{}, errAt72
	}
	return d, nil
}

// UpdateDashboardDetails changes human-readable fields without changing identity
// or lifecycle state. Deletion must be handled by the cascade protocol.
func (r *FileRepository) UpdateDashboardDetails(id, name, description string) (Dashboard, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	f, err := r.loadDashboardsLocked()
	if err != nil {
		return Dashboard{}, err
	}
	for i, d := range f.Dashboards {
		if d.ID != id {
			continue
		}
		if d.State != DashboardActive {
			return Dashboard{}, fmt.Errorf("dashboard %s is deleting", id)
		}
		d.Name, d.Description = name, description
		if errAt95 := validateDashboard(d); errAt95 != nil {
			return Dashboard{}, errAt95
		}
		f.Dashboards[i] = d
		if errAt99 := r.WriteAtomic(filepath.Join(r.dir, DashboardsFilename), f); errAt99 != nil {
			return Dashboard{}, errAt99
		}
		return d, nil
	}
	return Dashboard{}, fmt.Errorf("dashboard not found: %s", id)
}

func (r *FileRepository) loadDashboardsLocked() (dashboardsFile, error) {
	var f dashboardsFile
	if err := ReadYAML(filepath.Join(r.dir, DashboardsFilename), &f); err != nil {
		return dashboardsFile{}, err
	}
	if f.SchemaVersion != DashboardSchemaVersion {
		return dashboardsFile{}, fmt.Errorf("unsupported dashboard schema version: %d", f.SchemaVersion)
	}
	seen := make(map[string]bool, len(f.Dashboards))
	for _, d := range f.Dashboards {
		if err := validateDashboard(d); err != nil {
			return dashboardsFile{}, err
		}
		if seen[d.ID] {
			return dashboardsFile{}, fmt.Errorf("duplicate dashboard id: %s", d.ID)
		}
		seen[d.ID] = true
	}
	return f, nil
}

func validateDashboard(d Dashboard) error {
	if strings.TrimSpace(d.ID) == "" {
		return errors.New("dashboard id is required")
	}
	if strings.TrimSpace(d.Name) == "" {
		return errors.New("dashboard name is required")
	}
	if d.State != DashboardActive && d.State != DashboardDeleting {
		return fmt.Errorf("invalid dashboard state for %s: %q", d.ID, d.State)
	}
	return nil
}
