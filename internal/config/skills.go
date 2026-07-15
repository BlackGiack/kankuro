package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"sort"
)

// ListSkills returns the skill registry sorted by name, always including built-in skills.
func (r *FileRepository) ListSkills() ([]Skill, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.listSkillsLocked()
}

func (r *FileRepository) listSkillsLocked() ([]Skill, error) {
	f, err := r.LoadSkillsFile()
	if err != nil {
		return nil, err
	}
	skills := MergeBuiltinSkills(f.Skills)
	sort.Slice(skills, func(i, j int) bool { return skills[i].Name < skills[j].Name })
	return skills, nil
}

// UpsertSkill creates or updates one skill entry.
func (r *FileRepository) UpsertSkill(skill Skill) (Skill, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if skill.Name == "" {
		return Skill{}, errors.New("skill name is required")
	}

	f, err := r.LoadSkillsFile()
	if err != nil {
		return Skill{}, err
	}

	f.Skills, skill = UpsertInSlice(
		f.Skills,
		skill,
		func(s Skill) string { return s.ID },
		func(s *Skill, id string) { s.ID = id },
	)

	if err = r.WriteSkillsFile(f); err != nil {
		return Skill{}, err
	}
	return skill, nil
}

// DeleteSkill removes a skill unless it is referenced by existing ID cards.
func (r *FileRepository) DeleteSkill(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cards, err := r.listIDCardsLocked()
	if err != nil {
		return err
	}
	for _, card := range cards {
		if slices.Contains(card.SkillIDs, id) {
			return fmt.Errorf("skill %s is referenced by id card %s", id, card.ID)
		}
	}
	f, err := r.LoadSkillsFile()
	if err != nil {
		return err
	}
	orig := len(f.Skills)
	f.Skills = slices.DeleteFunc(f.Skills, func(s Skill) bool { return s.ID == id })
	if len(f.Skills) == orig {
		return fmt.Errorf("skill not found: %s", id)
	}
	return r.WriteSkillsFile(f)
}

// LoadSkillsFile reads and normalizes the skill-registry YAML payload.
func (r *FileRepository) LoadSkillsFile() (SkillsFile, error) {
	path := filepath.Join(r.dir, SkillsFilename)
	f := SkillsFile{}
	if err := ReadYAML(path, &f); err != nil {
		return SkillsFile{}, err
	}
	if f.SchemaVersion == 0 {
		f.SchemaVersion = SchemaVersion
	}
	if f.Skills == nil {
		f.Skills = []Skill{}
	}
	return f, nil
}

// WriteSkillsFile writes skills data using the current schema version.
func (r *FileRepository) WriteSkillsFile(f SkillsFile) error {
	f.SchemaVersion = SchemaVersion
	return r.WriteAtomic(filepath.Join(r.dir, SkillsFilename), f)
}
