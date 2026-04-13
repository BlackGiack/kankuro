package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"sort"
)

// ListPersonas returns the persona registry sorted by name.
func (r *FileRepository) ListPersonas() ([]Persona, error) {
	f, err := r.LoadPersonasFile()
	if err != nil {
		return nil, err
	}
	sort.Slice(f.Personas, func(i, j int) bool { return f.Personas[i].Name < f.Personas[j].Name })
	return f.Personas, nil
}

// UpsertPersona creates or updates one persona entry.
func (r *FileRepository) UpsertPersona(persona Persona) (Persona, error) {
	if persona.Name == "" {
		return Persona{}, errors.New("persona name is required")
	}

	f, err := r.LoadPersonasFile()
	if err != nil {
		return Persona{}, err
	}

	f.Personas, persona = UpsertInSlice(
		f.Personas,
		persona,
		func(p Persona) string { return p.ID },
		func(p *Persona, id string) { p.ID = id },
	)

	if err = r.WritePersonasFile(f); err != nil {
		return Persona{}, err
	}
	return persona, nil
}

// DeletePersona removes a persona unless it is referenced by existing ID cards.
func (r *FileRepository) DeletePersona(id string) error {
	cards, err := r.ListIDCards()
	if err != nil {
		return err
	}
	for _, card := range cards {
		if slices.Contains(card.PersonaIDs, id) {
			return fmt.Errorf("persona %s is referenced by id card %s", id, card.ID)
		}
	}
	f, err := r.LoadPersonasFile()
	if err != nil {
		return err
	}
	orig := len(f.Personas)
	f.Personas = slices.DeleteFunc(f.Personas, func(p Persona) bool { return p.ID == id })
	if len(f.Personas) == orig {
		return fmt.Errorf("persona not found: %s", id)
	}
	return r.WritePersonasFile(f)
}

// LoadPersonasFile reads and normalizes the persona-registry YAML payload.
func (r *FileRepository) LoadPersonasFile() (PersonasFile, error) {
	path := filepath.Join(r.dir, PersonasFilename)
	f := PersonasFile{}
	if err := ReadYAML(path, &f); err != nil {
		return PersonasFile{}, err
	}
	if f.SchemaVersion == 0 {
		f.SchemaVersion = SchemaVersion
	}
	if f.Personas == nil {
		f.Personas = []Persona{}
	}
	return f, nil
}

// WritePersonasFile writes personas data using the current schema version.
func (r *FileRepository) WritePersonasFile(f PersonasFile) error {
	f.SchemaVersion = SchemaVersion
	return r.WriteAtomic(filepath.Join(r.dir, PersonasFilename), f)
}
