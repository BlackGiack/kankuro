package config

import (
	"errors"
	"fmt"
)

// ValidateCard enforces required fields and checks skill/MCP references.
func (r *FileRepository) ValidateCard(card IDCard) error {
	if card.Name == "" {
		return errors.New("id card name is required")
	}
	if card.Driver == "" {
		return errors.New("driver is required")
	}
	if card.WorkingDir == "" {
		return errors.New("working_dir is required")
	}
	skills, err := r.ListSkills()
	if err != nil {
		return err
	}
	mcpList, err := r.ListMCPs()
	if err != nil {
		return err
	}
	personas, err := r.ListPersonas()
	if err != nil {
		return err
	}
	skillIDs := map[string]struct{}{}
	for _, s := range skills {
		skillIDs[s.ID] = struct{}{}
	}
	mcpIDs := map[string]struct{}{}
	for _, m := range mcpList {
		mcpIDs[m.ID] = struct{}{}
	}
	personaIDs := map[string]struct{}{}
	for _, p := range personas {
		personaIDs[p.ID] = struct{}{}
	}
	for _, id := range card.SkillIDs {
		if _, ok := skillIDs[id]; !ok {
			return fmt.Errorf("skill reference not found: %s", id)
		}
	}
	for _, id := range card.PersonaIDs {
		if _, ok := personaIDs[id]; !ok {
			return fmt.Errorf("persona reference not found: %s", id)
		}
	}
	for _, id := range card.MCPIDs {
		if _, ok := mcpIDs[id]; !ok {
			return fmt.Errorf("mcp reference not found: %s", id)
		}
	}
	return nil
}
