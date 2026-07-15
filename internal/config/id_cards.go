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
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.listIDCardsLocked()
}

// listIDCardsLocked is ListIDCards without acquiring r.mu, for use by callers
// that already hold it.
func (r *FileRepository) listIDCardsLocked() ([]IDCard, error) {
	f, err := r.LoadIDCardsFile()
	if err != nil {
		return nil, err
	}
	sort.Slice(f.IDCards, func(i, j int) bool { return f.IDCards[i].Name < f.IDCards[j].Name })
	return f.IDCards, nil
}

// CreateIDCard validates and persists a new ID card.
func (r *FileRepository) CreateIDCard(card IDCard) (IDCard, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if card.ID == "" {
		card.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	card.CreatedAt = now
	card.UpdatedAt = now

	f, err := r.LoadIDCardsFile()
	if err != nil {
		return IDCard{}, err
	}
	if err = r.ValidateCard(card); err != nil {
		return IDCard{}, err
	}
	for _, existing := range f.IDCards {
		if existing.ID == card.ID {
			return IDCard{}, fmt.Errorf("id card id already exists: %s", card.ID)
		}
	}
	f.IDCards = append(f.IDCards, card)
	if err = r.WriteIDCardsFile(f); err != nil {
		return IDCard{}, err
	}
	return card, nil
}

// UpdateIDCard validates and updates an existing ID card.
func (r *FileRepository) UpdateIDCard(card IDCard) (IDCard, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if card.ID == "" {
		return IDCard{}, errors.New("id is required")
	}
	f, err := r.LoadIDCardsFile()
	if err != nil {
		return IDCard{}, err
	}
	if err = r.ValidateCard(card); err != nil {
		return IDCard{}, err
	}
	found := false
	for _, existing := range f.IDCards {
		if existing.ID == card.ID {
			card.CreatedAt = existing.CreatedAt
			found = true
			break
		}
	}
	if !found {
		return IDCard{}, fmt.Errorf("id card not found: %s", card.ID)
	}
	card.UpdatedAt = time.Now().UTC()

	f.IDCards, card = UpsertInSlice(
		f.IDCards,
		card,
		func(c IDCard) string { return c.ID },
		func(c *IDCard, id string) { c.ID = id },
	)

	if err = r.WriteIDCardsFile(f); err != nil {
		return IDCard{}, err
	}
	return card, nil
}

// DeleteIDCard removes one ID card by identifier.
//
// It does not check whether the card is referenced by a live session:
// internal/config cannot see internal/session (that package imports this one,
// so the reverse would be an import cycle). Callers that can see sessions
// (internal/app's OnDeleteCard) must apply that guard before calling this.
func (r *FileRepository) DeleteIDCard(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	f, err := r.LoadIDCardsFile()
	if err != nil {
		return err
	}
	orig := len(f.IDCards)
	f.IDCards = slices.DeleteFunc(f.IDCards, func(c IDCard) bool { return c.ID == id })
	if len(f.IDCards) == orig {
		return fmt.Errorf("id card not found: %s", id)
	}
	return r.WriteIDCardsFile(f)
}

// LoadIDCardsFile reads and normalizes the ID-card YAML payload.
func (r *FileRepository) LoadIDCardsFile() (IDCardsFile, error) {
	path := filepath.Join(r.dir, IDCardsFilename)
	f := IDCardsFile{}
	if err := ReadYAML(path, &f); err != nil {
		return IDCardsFile{}, err
	}

	if f.SchemaVersion == 0 {
		f.SchemaVersion = SchemaVersion
	}
	if f.IDCards == nil {
		f.IDCards = []IDCard{}
	}
	return f, nil
}

// WriteIDCardsFile writes ID-card data using the current schema version.
func (r *FileRepository) WriteIDCardsFile(f IDCardsFile) error {
	f.SchemaVersion = SchemaVersion
	return r.WriteAtomic(filepath.Join(r.dir, IDCardsFilename), f)
}
