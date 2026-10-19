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
	SubWorkspaceIDs    []string          `yaml:"sub_workspace_ids,omitempty"`
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
	// Builtin marks a skill as bundled into the binary. Built-in skills are
	// never removed during source sync and their content is resolved from
	// embedded assets rather than from a filesystem source directory.
	Builtin bool `yaml:"builtin,omitempty"`
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
	Builtin     bool              `yaml:"builtin,omitempty"`
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

// AppSettings holds global user settings.
type AppSettings struct {
	IDECommand               string `yaml:"ide_command"`
	TerminalCommand          string `yaml:"terminal_command"`
	ShowPersonasInCardEditor bool   `yaml:"show_personas_in_card_editor"`
	ShowMCPsInCardEditor     bool   `yaml:"show_mcps_in_card_editor"`
}

func DefaultAppSettings() AppSettings {
	return AppSettings{
		ShowPersonasInCardEditor: true,
		ShowMCPsInCardEditor:     true,
	}
}

// Puppet represents a custom or built-in ASCII puppet character.
type Puppet struct {
	ID      string   `yaml:"id"`
	Name    string   `yaml:"name"`
	Source  string   `yaml:"source"`
	Builtin bool     `yaml:"builtin,omitempty"`
	Lines   []string `yaml:"lines"`
}

type PuppetsFile struct {
	SchemaVersion int      `yaml:"schema_version"`
	Puppets       []Puppet `yaml:"puppets"`
}

type PuppetSourcesFile struct {
	SchemaVersion int      `yaml:"schema_version"`
	Paths         []string `yaml:"paths"`
}

type AppSettingsFile struct {
	SchemaVersion int         `yaml:"schema_version"`
	AppSettings   AppSettings `yaml:"app_settings"`
}
