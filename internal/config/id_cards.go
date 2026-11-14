package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/BlackGiack/kankuro/internal/fsutil"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

// CardSchemaVersion is the membership-bearing ID-card format version.
const CardSchemaVersion = 2

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
	if err := r.checkCardNotTombstonedLocked(card.ID); err != nil {
		return IDCard{}, err
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
	persisted, err := r.writeIDCardsFile(f)
	if err != nil {
		return IDCard{}, err
	}
	for _, item := range persisted.IDCards {
		if item.ID == card.ID {
			return item, nil
		}
	}
	return IDCard{}, errors.New("saved card absent from encoded inventory")
}

// UpdateIDCard validates and updates an existing ID card.
func (r *FileRepository) UpdateIDCard(card IDCard) (IDCard, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if card.ID == "" {
		return IDCard{}, errors.New("id is required")
	}
	if err := r.checkCardNotTombstonedLocked(card.ID); err != nil {
		return IDCard{}, err
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
			if existing.DashboardID != card.DashboardID {
				return IDCard{}, errors.New("card dashboard ownership cannot be changed")
			}
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

	persisted, err := r.writeIDCardsFile(f)
	if err != nil {
		return IDCard{}, err
	}
	for _, item := range persisted.IDCards {
		if item.ID == card.ID {
			return item, nil
		}
	}
	return IDCard{}, errors.New("saved card absent from encoded inventory")
}

// LoadIDCardsFile reads and validates the current ID-card YAML payload.
func (r *FileRepository) LoadIDCardsFile() (IDCardsFile, error) {
	path := filepath.Join(r.dir, IDCardsFilename)
	f := IDCardsFile{}
	if err := ReadYAML(path, &f); err != nil {
		return IDCardsFile{}, err
	}

	if f.SchemaVersion != CardSchemaVersion {
		return IDCardsFile{}, fmt.Errorf("unsupported card schema version: %d", f.SchemaVersion)
	}
	if f.IDCards == nil {
		f.IDCards = []IDCard{}
	}
	if err := validateCardMembershipFormat(f); err != nil {
		return IDCardsFile{}, err
	}
	return f, nil
}

// WriteIDCardsFile validates and writes the current card schema.
func (r *FileRepository) WriteIDCardsFile(f IDCardsFile) error {
	_, err := r.writeIDCardsFile(f)
	return err
}

// writeIDCardsFile returns the representation decoded from the exact bytes
// written. Encoding and decoding occur before the atomic write, so a returned
// card never comes from a later writer and no post-write read can fail.
func (r *FileRepository) writeIDCardsFile(f IDCardsFile) (IDCardsFile, error) {
	if f.SchemaVersion != CardSchemaVersion {
		return IDCardsFile{}, fmt.Errorf("unsupported card schema version: %d", f.SchemaVersion)
	}
	if err := validateCardMembershipFormat(f); err != nil {
		return IDCardsFile{}, err
	}
	raw, err := yaml.Marshal(f)
	if err != nil {
		return IDCardsFile{}, fmt.Errorf("marshal %s: %w", IDCardsFilename, err)
	}
	var persisted IDCardsFile
	if err = yaml.Unmarshal(raw, &persisted); err != nil {
		return IDCardsFile{}, fmt.Errorf("decode encoded cards: %w", err)
	}
	if err = validateCardMembershipFormat(persisted); err != nil {
		return IDCardsFile{}, err
	}
	if err = fsutil.WriteFileAtomic(filepath.Join(r.dir, IDCardsFilename), raw, configFileMode); err != nil {
		return IDCardsFile{}, err
	}
	return persisted, nil
}

func validateCardMembershipFormat(f IDCardsFile) error {
	seen := make(map[string]bool, len(f.IDCards))
	for _, card := range f.IDCards {
		if card.ID == "" || strings.TrimSpace(card.DashboardID) == "" || seen[card.ID] {
			return errors.New("card inventory has missing or duplicate identity or membership")
		}
		seen[card.ID] = true
	}
	return nil
}
