package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"sort"
	"time"

	"github.com/google/uuid"
)

// ListIDCards returns all ID cards sorted by name.
func (r *FileRepository) ListIDCards() ([]IDCard, error) {
	f, err := r.loadIDCardsFile()
	if err != nil {
		return nil, err
	}
	sort.Slice(f.IDCards, func(i, j int) bool { return f.IDCards[i].Name < f.IDCards[j].Name })
	return f.IDCards, nil
}

// CreateIDCard validates and persists a new ID card.
func (r *FileRepository) CreateIDCard(card IDCard) (IDCard, error) {
	if card.ID == "" {
		card.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	card.CreatedAt = now
	card.UpdatedAt = now

	f, err := r.loadIDCardsFile()
	if err != nil {
		return IDCard{}, err
	}
	if err := r.validateCard(card); err != nil {
		return IDCard{}, err
	}
	for _, existing := range f.IDCards {
		if existing.ID == card.ID {
			return IDCard{}, fmt.Errorf("id card id already exists: %s", card.ID)
		}
	}
	f.IDCards = append(f.IDCards, card)
	if err := r.writeIDCardsFile(f); err != nil {
		return IDCard{}, err
	}
	return card, nil
}

// UpdateIDCard validates and updates an existing ID card.
func (r *FileRepository) UpdateIDCard(card IDCard) (IDCard, error) {
	if card.ID == "" {
		return IDCard{}, errors.New("id is required")
	}
	f, err := r.loadIDCardsFile()
	if err != nil {
		return IDCard{}, err
	}
	if err := r.validateCard(card); err != nil {
		return IDCard{}, err
	}
	idx := -1
	for i, existing := range f.IDCards {
		if existing.ID == card.ID {
			idx = i
			card.CreatedAt = existing.CreatedAt
			break
		}
	}
	if idx == -1 {
		return IDCard{}, fmt.Errorf("id card not found: %s", card.ID)
	}
	card.UpdatedAt = time.Now().UTC()
	f.IDCards[idx] = card
	if err := r.writeIDCardsFile(f); err != nil {
		return IDCard{}, err
	}
	return card, nil
}

// DeleteIDCard removes one ID card by identifier.
func (r *FileRepository) DeleteIDCard(id string) error {
	f, err := r.loadIDCardsFile()
	if err != nil {
		return err
	}
	orig := len(f.IDCards)
	f.IDCards = slices.DeleteFunc(f.IDCards, func(c IDCard) bool { return c.ID == id })
	if len(f.IDCards) == orig {
		return fmt.Errorf("id card not found: %s", id)
	}
	return r.writeIDCardsFile(f)
}

// loadIDCardsFile reads and normalizes the ID-card YAML payload.
func (r *FileRepository) loadIDCardsFile() (idCardsFile, error) {
	path := filepath.Join(r.dir, idCardsFilename)
	f := idCardsFile{}
	if err := readYAML(path, &f); err != nil {
		return idCardsFile{}, err
	}
	if f.SchemaVersion == 0 {
		f.SchemaVersion = SchemaVersion
	}
	if f.IDCards == nil {
		f.IDCards = []IDCard{}
	}
	return f, nil
}

// writeIDCardsFile writes ID-card data using the current schema version.
func (r *FileRepository) writeIDCardsFile(f idCardsFile) error {
	f.SchemaVersion = SchemaVersion
	return r.writeAtomic(filepath.Join(r.dir, idCardsFilename), f)
}
