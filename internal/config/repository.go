package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
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
	AppSettingsFilename      = "app_settings.yaml"
	PuppetsFilename          = "puppets.yaml"
	PuppetSourcesFilename    = "puppet_sources.yaml"
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
	// UpdateWorkspaceSources applies fn to the current source list and
	// persists + syncs the result atomically. Prefer this over
	// ListWorkspaceSources+SaveWorkspaceSources for any read-modify-write.
	UpdateWorkspaceSources(fn func([]WorkspaceSource) ([]WorkspaceSource, error)) error
	SyncWorkspacesFromSources() error
	ListWorkspaces() ([]Workspace, error)
	GetWorkspaceByPath(path string) (Workspace, bool)
	DeleteWorkspaceSource(id string) error
}

// SkillRepository defines persistence operations for skills.
type SkillRepository interface {
	ListSkillSources() ([]string, error)
	SaveSkillSources(paths []string) error
	// UpdateSkillSources applies fn to the current source list and persists
	// + syncs the result atomically. Prefer this over
	// ListSkillSources+SaveSkillSources for any read-modify-write.
	UpdateSkillSources(fn func([]string) ([]string, error)) error
	SyncSkillsFromSources() error
	ListSkills() ([]Skill, error)
	UpsertSkill(skill Skill) (Skill, error)
	DeleteSkill(id string) error
}

// PersonaRepository defines persistence operations for personas.
type PersonaRepository interface {
	ListPersonaSources() ([]string, error)
	SavePersonaSources(paths []string) error
	// UpdatePersonaSources applies fn to the current source list and
	// persists + syncs the result atomically. Prefer this over
	// ListPersonaSources+SavePersonaSources for any read-modify-write.
	UpdatePersonaSources(fn func([]string) ([]string, error)) error
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

// SettingsRepository defines persistence operations for global app settings.
type SettingsRepository interface {
	GetAppSettings() (AppSettings, error)
	SaveAppSettings(cfg AppSettings) error
}

// PuppetRepository defines persistence operations for puppets.
type PuppetRepository interface {
	ListPuppetSources() ([]string, error)
	SavePuppetSources(paths []string) error
	// UpdatePuppetSources applies fn to the current source list and persists
	// + syncs the result atomically. Prefer this over
	// ListPuppetSources+SavePuppetSources for any read-modify-write.
	UpdatePuppetSources(fn func([]string) ([]string, error)) error
	SyncPuppetsFromSources() error
	ListPuppets() ([]Puppet, error)
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
	SettingsRepository
	PuppetRepository
	Dir() string
}

// FileRepository stores all configuration state in YAML files under one directory.
//
// mu serializes read-modify-write sequences (list, mutate in memory, write
// back) across all domains. It guards logical consistency on top of the
// crash/interleaving safety WriteAtomic already provides for a single write.
// It is a plain Mutex, not an RWMutex: several public readers are called
// internally by mutators that already hold the lock (e.g. ValidateCard reads
// workspaces/skills/personas/mcps while Create/UpdateIDCard hold it), and
// Go's locks are not reentrant, so every exported method that touches disk
// takes mu itself and internal call sites use the unexported *Locked helpers
// that assume the lock is already held.
type FileRepository struct {
	dir string
	mu  sync.Mutex
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

	settingsPath := filepath.Join(r.dir, AppSettingsFilename)
	settingsDefault := AppSettingsFile{SchemaVersion: SchemaVersion, AppSettings: DefaultAppSettings()}
	if err := r.EnsureYAML(settingsPath, settingsDefault); err != nil {
		return err
	}

	puppetsPath := filepath.Join(r.dir, PuppetsFilename)
	puppetsDefault := PuppetsFile{SchemaVersion: SchemaVersion, Puppets: []Puppet{}}
	if err := r.EnsureYAML(puppetsPath, puppetsDefault); err != nil {
		return err
	}

	puppetSourcesPath := filepath.Join(r.dir, PuppetSourcesFilename)
	puppetSourcesDefault := PuppetSourcesFile{SchemaVersion: SchemaVersion, Paths: []string{}}
	if err := r.EnsureYAML(puppetSourcesPath, puppetSourcesDefault); err != nil {
		return err
	}

	return nil
}

// GetAppSettings loads app settings from disk. Fields absent from the file
// keep DefaultAppSettings' value (so a field added after the file was first
// written defaults sanely instead of zeroing); fields present are decoded
// straight from the real AppSettings struct via reflection, so a field added
// to AppSettings is picked up automatically without hand-editing a second,
// divergent schema here (see TestGetAppSettingsCoversEveryField for the
// regression guard).
func (r *FileRepository) GetAppSettings() (AppSettings, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	path := filepath.Join(r.dir, AppSettingsFilename)
	var raw struct {
		SchemaVersion int                  `yaml:"schema_version"`
		AppSettings   map[string]yaml.Node `yaml:"app_settings"`
	}
	if err := ReadYAML(path, &raw); err != nil {
		return AppSettings{}, err
	}
	settings := DefaultAppSettings()
	if err := overlayYAMLFields(&settings, raw.AppSettings); err != nil {
		return AppSettings{}, fmt.Errorf("decode app_settings: %w", err)
	}
	return settings, nil
}

// SaveAppSettings persists app settings to disk.
func (r *FileRepository) SaveAppSettings(cfg AppSettings) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	path := filepath.Join(r.dir, AppSettingsFilename)
	f := AppSettingsFile{SchemaVersion: SchemaVersion, AppSettings: cfg}
	return r.WriteAtomic(path, f)
}

// overlayYAMLFields decodes each field present in values onto the matching
// yaml-tagged field of dst (a pointer to a struct), leaving fields absent
// from values untouched. It is used to apply a partial YAML document onto a
// struct that already carries defaults.
func overlayYAMLFields(dst any, values map[string]yaml.Node) error {
	v := reflect.ValueOf(dst).Elem()
	t := v.Type()
	for i := range t.NumField() {
		tag := strings.Split(t.Field(i).Tag.Get("yaml"), ",")[0]
		if tag == "" || tag == "-" {
			continue
		}
		node, ok := values[tag]
		if !ok {
			continue
		}
		if err := node.Decode(v.Field(i).Addr().Interface()); err != nil {
			return fmt.Errorf("field %q: %w", tag, err)
		}
	}
	return nil
}
