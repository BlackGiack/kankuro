package config

import (
	"fmt"
)

// ValidateCard enforces required fields and checks skill/MCP references.
func (r *FileRepository) ValidateCard(card IDCard) error {
	if err := validateCardFields(card); err != nil {
		return err
	}

	if err := r.validateWorkspaceRef(card.WorkspaceID); err != nil {
		return err
	}

	if err := r.validateSkillRefs(card.SkillIDs); err != nil {
		return err
	}

	if err := r.validatePersonaRefs(card.PersonaIDs); err != nil {
		return err
	}

	return r.validateMCPRefs(card.MCPIDs)
}

func (r *FileRepository) validateWorkspaceRef(workspaceID string) error {
	workspaces, err := r.ListWorkspaces()
	if err != nil {
		return err
	}
	set := idSet(workspaces, func(w Workspace) string { return w.ID })
	if _, ok := set[workspaceID]; !ok {
		return fmt.Errorf("workspace reference not found: %s", workspaceID)
	}
	return nil
}

func (r *FileRepository) validateSkillRefs(skillIDs []string) error {
	skills, err := r.ListSkills()
	if err != nil {
		return err
	}
	if len(skillIDs) == 0 {
		return nil
	}
	return validateRefs(skillIDs, idSet(skills, func(s Skill) string { return s.ID }), "skill")
}

func (r *FileRepository) validatePersonaRefs(personaIDs []string) error {
	personas, err := r.ListPersonas()
	if err != nil {
		return err
	}
	if len(personaIDs) == 0 {
		return nil
	}
	return validateRefs(personaIDs, idSet(personas, func(p Persona) string { return p.ID }), "persona")
}

func (r *FileRepository) validateMCPRefs(mcpIDs []string) error {
	mcps, err := r.ListMCPs()
	if err != nil {
		return err
	}
	if len(mcpIDs) == 0 {
		return nil
	}
	return validateRefs(mcpIDs, idSet(mcps, func(m MCPServer) string { return m.ID }), "mcp")
}
