package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	IDCardsFilename          = "id_cards.yaml"
	SkillsFilename           = "skills.yaml"
	SkillsSourcesFilename    = "skills_sources.yaml"
	PersonasFilename         = "personas.yaml"
	PersonaSourcesFilename   = "persona_sources.yaml"
	MCPsFilename             = "mcps.yaml"
	WorkspacesFilename       = "workspaces.yaml"
	WorkspaceSourcesFilename = "workspace_sources.yaml"
)

// CardRepository defines persistence operations for ID cards.
type CardRepository interface {
	ListIDCards() ([]IDCard, error)
	CreateIDCard(card IDCard) (IDCard, error)
	UpdateIDCard(card IDCard) (IDCard, error)
	DeleteIDCard(id string) error
}

// WorkspaceRepository defines persistence operations for workspaces.
type WorkspaceRepository interface {
	ListWorkspaceSources() ([]WorkspaceSource, error)
	SaveWorkspaceSources(sources []WorkspaceSource) error
	SyncWorkspacesFromSources() error
	ListWorkspaces() ([]Workspace, error)
	GetWorkspaceByPath(path string) (Workspace, bool)
	DeleteWorkspaceSource(id string) error
}

// SkillRepository defines persistence operations for skills.
type SkillRepository interface {
	ListSkillSources() ([]string, error)
	SaveSkillSources(paths []string) error
	SyncSkillsFromSources() error
	ListSkills() ([]Skill, error)
	UpsertSkill(skill Skill) (Skill, error)
	DeleteSkill(id string) error
}

// PersonaRepository defines persistence operations for personas.
type PersonaRepository interface {
	ListPersonaSources() ([]string, error)
	SavePersonaSources(paths []string) error
	SyncPersonasFromSources() error
	ListPersonas() ([]Persona, error)
	UpsertPersona(persona Persona) (Persona, error)
	DeletePersona(id string) error
}

// MCPRepository defines persistence operations for MCP servers.
type MCPRepository interface {
	ListMCPs() ([]MCPServer, error)
	UpsertMCP(mcp MCPServer) (MCPServer, error)
	DeleteMCP(id string) error
}

// Repository composes all domain repositories for convenience.
// Use individual sub-interfaces (CardRepository, etc.) when you only need
// operations for a specific domain.
type Repository interface {
	CardRepository
	WorkspaceRepository
	SkillRepository
	PersonaRepository
	MCPRepository
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
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("create config dir: %w", err)
	}
	r := &FileRepository{dir: dir}
	if err := r.EnsureFiles(); err != nil {
		return nil, err
	}
	return r, nil
}

// Dir returns the effective repository directory path.
func (r *FileRepository) Dir() string { return r.dir }

// EnsureFiles guarantees that all expected config files exist.
func (r *FileRepository) EnsureFiles() error {
	idCardsPath := filepath.Join(r.dir, IDCardsFilename)
	idCardsDefault := IDCardsFile{SchemaVersion: SchemaVersion, IDCards: []IDCard{}}
	if err := r.EnsureYAML(idCardsPath, idCardsDefault); err != nil {
		return err
	}

	skillsPath := filepath.Join(r.dir, SkillsFilename)
	skillsDefault := SkillsFile{SchemaVersion: SchemaVersion, Skills: []Skill{}}
	if err := r.EnsureYAML(skillsPath, skillsDefault); err != nil {
		return err
	}

	sourcesPath := filepath.Join(r.dir, SkillsSourcesFilename)
	sourcesDefault := SkillsSourcesFile{SchemaVersion: SchemaVersion, Paths: []string{}}
	if err := r.EnsureYAML(sourcesPath, sourcesDefault); err != nil {
		return err
	}

	personasPath := filepath.Join(r.dir, PersonasFilename)
	personasDefault := PersonasFile{SchemaVersion: SchemaVersion, Personas: []Persona{}}
	if err := r.EnsureYAML(personasPath, personasDefault); err != nil {
		return err
	}

	personaSourcesPath := filepath.Join(r.dir, PersonaSourcesFilename)
	personaSourcesDefault := PersonaSourcesFile{SchemaVersion: SchemaVersion, Paths: []string{}}
	if err := r.EnsureYAML(personaSourcesPath, personaSourcesDefault); err != nil {
		return err
	}

	mcpsPath := filepath.Join(r.dir, MCPsFilename)
	mcpsDefault := MCPsFile{SchemaVersion: SchemaVersion, MCPs: []MCPServer{}}
	if err := r.EnsureYAML(mcpsPath, mcpsDefault); err != nil {
		return err
	}

	workspacesPath := filepath.Join(r.dir, WorkspacesFilename)
	workspacesDefault := WorkspacesFile{SchemaVersion: SchemaVersion, Workspaces: []Workspace{}}
	if err := r.EnsureYAML(workspacesPath, workspacesDefault); err != nil {
		return err
	}

	workspaceSourcesPath := filepath.Join(r.dir, WorkspaceSourcesFilename)
	workspaceSourcesDefault := WorkspaceSourcesFile{SchemaVersion: SchemaVersion, Sources: []WorkspaceSource{}}
	if err := r.EnsureYAML(workspaceSourcesPath, workspaceSourcesDefault); err != nil {
		return err
	}

	return nil
}
