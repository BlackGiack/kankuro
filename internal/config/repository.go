package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	idCardsFilename       = "id_cards.yaml"
	skillsFilename        = "skills.yaml"
	skillsSourcesFilename = "skills_sources.yaml"
	mcpsFilename          = "mcps.yaml"
)

// Repository defines persistence operations for ID cards and capability registries.
type Repository interface {
	ListIDCards() ([]IDCard, error)
	CreateIDCard(card IDCard) (IDCard, error)
	UpdateIDCard(card IDCard) (IDCard, error)
	DeleteIDCard(id string) error

	ListSkillSources() ([]string, error)
	SyncSkillsFromSources() error
	ListSkills() ([]Skill, error)
	UpsertSkill(skill Skill) (Skill, error)
	DeleteSkill(id string) error

	ListMCPs() ([]MCPServer, error)
	UpsertMCP(mcp MCPServer) (MCPServer, error)
	DeleteMCP(id string) error

	Dir() string
}

// FileRepository stores all configuration state in YAML files under one directory.
type FileRepository struct {
	dir string
}

// NewFileRepository creates the repository and bootstraps empty YAML files when missing.
func NewFileRepository(dir string) (*FileRepository, error) {
	if dir == "" {
		cfg, err := os.UserConfigDir()
		if err != nil {
			return nil, fmt.Errorf("resolve user config dir: %w", err)
		}
		dir = filepath.Join(cfg, "kankuro")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create config dir: %w", err)
	}
	r := &FileRepository{dir: dir}
	if err := r.ensureFiles(); err != nil {
		return nil, err
	}
	return r, nil
}

// Dir returns the effective repository directory path.
func (r *FileRepository) Dir() string { return r.dir }

// ensureFiles guarantees that all expected config files exist.
func (r *FileRepository) ensureFiles() error {
	if err := r.ensureYAML(filepath.Join(r.dir, idCardsFilename), idCardsFile{SchemaVersion: SchemaVersion, IDCards: []IDCard{}}); err != nil {
		return err
	}
	if err := r.ensureYAML(filepath.Join(r.dir, skillsFilename), skillsFile{SchemaVersion: SchemaVersion, Skills: []Skill{}}); err != nil {
		return err
	}
	if err := r.ensureYAML(filepath.Join(r.dir, skillsSourcesFilename), skillsSourcesFile{SchemaVersion: SchemaVersion, Paths: []string{}}); err != nil {
		return err
	}
	if err := r.ensureYAML(filepath.Join(r.dir, mcpsFilename), mcpsFile{SchemaVersion: SchemaVersion, MCPs: []MCPServer{}}); err != nil {
		return err
	}
	return nil
}
