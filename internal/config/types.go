package config

// TODO: Split this file declaring each type in the file it's used

import (
	"time"
)

const SchemaVersion = 1

type IDCard struct {
	ID                 string            `yaml:"id"`
	Name               string            `yaml:"name"`
	Driver             string            `yaml:"driver"`
	Model              string            `yaml:"model"`
	WorkspaceID        string            `yaml:"workspace_id"`
	UseTmpWorkspace    bool              `yaml:"use_tmp_workspace"`
	SystemInstruction  string            `yaml:"system_instruction"`
	SkillIDs           []string          `yaml:"skill_ids"`
	PersonaIDs         []string          `yaml:"persona_ids"`
	MCPIDs             []string          `yaml:"mcp_ids"`
	Env                map[string]string `yaml:"env"`
	FaceStyle          string            `yaml:"face_style"`
	Puppet             string            `yaml:"puppet"`
	AgentsMD           string            `yaml:"agents_md"`
	AgentsMDPath       string            `yaml:"agents_md_path"`
	AgentsMDPreferPath bool              `yaml:"agents_md_prefer_path"`
	CreatedAt          time.Time         `yaml:"created_at"`
	UpdatedAt          time.Time         `yaml:"updated_at"`
}

type Workspace struct {
	ID     string `yaml:"id"`
	Name   string `yaml:"name"`
	Path   string `yaml:"path"`
	Source string `yaml:"source"`
}

type WorkspaceSource struct {
	ID     string `yaml:"id"`
	Path   string `yaml:"path"`
	IsRoot bool   `yaml:"is_root"`
}

type Skill struct {
	ID          string `yaml:"id"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Version     string `yaml:"version"`
	Source      string `yaml:"source"`
}

type Persona struct {
	ID     string `yaml:"id"`
	Name   string `yaml:"name"`
	Source string `yaml:"source"`
}

type MCPServer struct {
	ID          string            `yaml:"id"`
	Name        string            `yaml:"name"`
	Command     string            `yaml:"command"`
	Args        []string          `yaml:"args"`
	Env         map[string]string `yaml:"env"`
	CWD         string            `yaml:"cwd"`
	URL         string            `yaml:"url"`
	Headers     map[string]string `yaml:"headers"`
	Description string            `yaml:"description"`
}

type IDCardsFile struct {
	SchemaVersion int      `yaml:"schema_version"`
	IDCards       []IDCard `yaml:"id_cards"`
}

type SkillsFile struct {
	SchemaVersion int     `yaml:"schema_version"`
	Skills        []Skill `yaml:"skills"`
}

type SkillsSourcesFile struct {
	SchemaVersion int      `yaml:"schema_version"`
	Paths         []string `yaml:"paths"`
}

type PersonasFile struct {
	SchemaVersion int       `yaml:"schema_version"`
	Personas      []Persona `yaml:"personas"`
}

type PersonaSourcesFile struct {
	SchemaVersion int      `yaml:"schema_version"`
	Paths         []string `yaml:"paths"`
}

type MCPsFile struct {
	SchemaVersion int         `yaml:"schema_version"`
	MCPs          []MCPServer `yaml:"mcps"`
}

type WorkspacesFile struct {
	SchemaVersion int         `yaml:"schema_version"`
	Workspaces    []Workspace `yaml:"workspaces"`
}

type WorkspaceSourcesFile struct {
	SchemaVersion int               `yaml:"schema_version"`
	Sources       []WorkspaceSource `yaml:"sources"`
}

// IDEConfig holds the user's IDE launch command configuration.
type IDEConfig struct {
	Command string `yaml:"command"`
}

// IDEConfigFile represents the persisted IDE configuration file.
type IDEConfigFile struct {
	SchemaVersion int       `yaml:"schema_version"`
	IDEConfig     IDEConfig `yaml:"ide_config"`
}
