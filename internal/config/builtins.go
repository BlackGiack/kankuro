package config

import (
	"slices"
	"sync"
)

const (
	// BuiltinSkillMCPCLI is the stable ID for the built-in mcp-cli skill.
	BuiltinSkillMCPCLI = "builtin:mcp-cli"
	// BuiltinMCPKankuro is the stable selectable ID for Kankuro's MCP endpoint.
	BuiltinMCPKankuro = "builtin:kankuro"
)

// BuiltinMCPs returns the built-in servers offered by Kankuro.
func BuiltinMCPs() []MCPServer {
	return []MCPServer{
		{ID: BuiltinMCPKankuro, Name: "kankuro", Description: "Register this agent's native conversation with Kankuro", Builtin: true},
	}
}

// MergeBuiltinMCPs adds built-ins without duplicating persisted entries.
func MergeBuiltinMCPs(mcps []MCPServer) []MCPServer {
	result := append([]MCPServer(nil), mcps...)
	for _, builtin := range BuiltinMCPs() {
		found := false
		for _, existing := range result {
			if existing.ID == builtin.ID {
				found = true
				break
			}
		}
		if !found {
			result = append(result, builtin)
		}
	}
	return result
}

//nolint:gochecknoglobals // startup-injected skill content is shared process configuration.
var builtinMCPCLISkill struct {
	sync.RWMutex

	content []byte
}

// SetBuiltinMCPCLISkillContent sets process-wide content supplied by main at startup.
// It clones the input so the embedding owner cannot mutate config's shared bytes.
func SetBuiltinMCPCLISkillContent(content []byte) {
	builtinMCPCLISkill.Lock()
	defer builtinMCPCLISkill.Unlock()
	if len(content) == 0 {
		builtinMCPCLISkill.content = nil
		return
	}
	builtinMCPCLISkill.content = slices.Clone(content)
}

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

// GetBuiltinSkillContent returns the injected content for a built-in skill by name.
func GetBuiltinSkillContent(name string) ([]byte, bool) {
	if name != "mcp-cli" {
		return nil, false
	}
	builtinMCPCLISkill.RLock()
	defer builtinMCPCLISkill.RUnlock()
	if len(builtinMCPCLISkill.content) == 0 {
		return nil, false
	}
	return slices.Clone(builtinMCPCLISkill.content), true
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
