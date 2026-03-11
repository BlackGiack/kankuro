package config

// TODO: Split this file declaring each type in the file it's used

import "time"

const SchemaVersion = 1

type IDCard struct {
	ID                string            `yaml:"id"`
	Name              string            `yaml:"name"`
	Driver            string            `yaml:"driver"`
	Model             string            `yaml:"model"`
	WorkingDir        string            `yaml:"working_dir"`
	UseTmpWorkspace   bool              `yaml:"use_tmp_workspace"`
	SystemInstruction string            `yaml:"system_instruction"`
	SkillIDs          []string          `yaml:"skill_ids"`
	MCPIDs            []string          `yaml:"mcp_ids"`
	Env               map[string]string `yaml:"env"`
	FaceStyle         string            `yaml:"face_style"`
	CreatedAt         time.Time         `yaml:"created_at"`
	UpdatedAt         time.Time         `yaml:"updated_at"`
}

type Skill struct {
	ID          string `yaml:"id"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Version     string `yaml:"version"`
	Source      string `yaml:"source"`
}

type MCPServer struct {
	ID          string            `yaml:"id"`
	Name        string            `yaml:"name"`
	Command     string            `yaml:"command"`
	Args        []string          `yaml:"args"`
	Env         map[string]string `yaml:"env"`
	Description string            `yaml:"description"`
}

type idCardsFile struct {
	SchemaVersion int      `yaml:"schema_version"`
	IDCards       []IDCard `yaml:"id_cards"`
}

type skillsFile struct {
	SchemaVersion int     `yaml:"schema_version"`
	Skills        []Skill `yaml:"skills"`
}

type mcpsFile struct {
	SchemaVersion int         `yaml:"schema_version"`
	MCPs          []MCPServer `yaml:"mcps"`
}
