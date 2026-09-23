package config

import (
	"fmt"
)

// ValidateCard enforces required fields and checks skill/MCP references.
//
// Its sub-checks read via the unexported *Locked helpers rather than the
// public List* methods: CreateIDCard/UpdateIDCard call ValidateCard while
// already holding FileRepository.mu, and re-locking a non-reentrant mutex
// from the same goroutine would deadlock. ValidateCard itself does not take
// the lock, including when called standalone (as tests do) without any lock
// held. That is still safe, but not because there is no concurrent writer to
// race: there can be. It is safe because each *Locked reader reads one whole
// YAML file that was published by WriteAtomic's temp-file-then-rename, so it
// always sees a complete, valid document - a concurrent writer can only make
// the read stale (it may miss a write that lands a moment later), never
// torn or partial.

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
	workspaces, err := r.listWorkspacesLocked()
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
	skills, err := r.listSkillsLocked()
	if err != nil {
		return err
	}
	if len(skillIDs) == 0 {
		return nil
	}
	return validateRefs(skillIDs, idSet(skills, func(s Skill) string { return s.ID }), "skill")
}

func (r *FileRepository) validatePersonaRefs(personaIDs []string) error {
	personas, err := r.listPersonasLocked()
	if err != nil {
		return err
	}
	if len(personaIDs) == 0 {
		return nil
	}
	return validateRefs(personaIDs, idSet(personas, func(p Persona) string { return p.ID }), "persona")
}

func (r *FileRepository) validateMCPRefs(mcpIDs []string) error {
	mcps, err := r.listMCPsLocked()
	if err != nil {
		return err
	}
	if len(mcpIDs) == 0 {
		return nil
	}
	return validateRefs(mcpIDs, idSet(mcps, func(m MCPServer) string { return m.ID }), "mcp")
}
