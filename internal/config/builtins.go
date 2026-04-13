package config

import (
	"kankuro"
	"slices"
)

const (
	// BuiltinSkillMCPCLI is the stable ID for the built-in mcp-cli skill.
	BuiltinSkillMCPCLI = "builtin:mcp-cli"
)

// BuiltinSkills returns all built-in skill entries.
func BuiltinSkills() []Skill {
	return []Skill{
		{
			ID:          BuiltinSkillMCPCLI,
			Name:        "mcp-cli",
			Description: "Interface for MCP (Model Context Protocol) servers via CLI",
			Builtin:     true,
		},
	}
}

// IsBuiltinSkill returns true if the skill ID corresponds to a built-in skill.
func IsBuiltinSkill(id string) bool {
	return slices.ContainsFunc(BuiltinSkills(), func(s Skill) bool {
		return s.ID == id
	})
}

// GetBuiltinSkill returns the built-in skill by ID, or zero value if not found.
func GetBuiltinSkill(id string) (Skill, bool) {
	for _, s := range BuiltinSkills() {
		if s.ID == id {
			return s, true
		}
	}
	return Skill{}, false
}

// GetBuiltinSkillContent returns the embedded content for a built-in skill by name.
func GetBuiltinSkillContent(name string) ([]byte, bool) {
	switch name {
	case "mcp-cli":
		return kankuro.EmbeddedMCPCLISkill, true
	default:
		return nil, false
	}
}

// MergeBuiltinSkills merges built-in skills into the provided skill list.
// If a built-in skill is already present (by ID), it is not duplicated.
func MergeBuiltinSkills(skills []Skill) []Skill {
	builtins := BuiltinSkills()
	result := make([]Skill, 0, len(skills)+len(builtins))
	result = append(result, skills...)

	existingIDs := make(map[string]struct{}, len(result))
	for _, s := range result {
		existingIDs[s.ID] = struct{}{}
	}

	for _, builtin := range builtins {
		if _, ok := existingIDs[builtin.ID]; !ok {
			result = append(result, builtin)
		}
	}
	return result
}
