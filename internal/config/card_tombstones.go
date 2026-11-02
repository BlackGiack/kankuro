package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	CardTombstonesFilename = "card_tombstones.yaml"
	CardDeleting           = "deleting"
	CardDeleted            = "deleted"
)

// CardTombstone retains non-executable identity after global card deletion.
type CardTombstone struct {
	ID          string                `yaml:"id"`
	DashboardID string                `yaml:"dashboard_id"`
	Name        string                `yaml:"name"`
	State       string                `yaml:"state"`
	DeletedAt   time.Time             `yaml:"deleted_at"`
	Pending     []CardDeletionPending `yaml:"pending,omitempty"`
}

// CardDeletionPending retains the cleanup identity of an unpublished launch.
type CardDeletionPending struct {
	DashboardID   string `yaml:"dashboard_id"`
	SessionID     string `yaml:"session_id"`
	WorkingDir    string `yaml:"working_dir"`
	TemporaryRoot string `yaml:"temporary_root"`
	Cleaned       bool   `yaml:"cleaned"`
	Error         string `yaml:"error,omitempty"`
}

type cardTombstonesFile struct {
	SchemaVersion int             `yaml:"schema_version"`
	Tombstones    []CardTombstone `yaml:"tombstones"`
}

// ListCardTombstones returns retained card identities and deletion states.
func (r *FileRepository) ListCardTombstones() ([]CardTombstone, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	file, err := r.loadCardTombstonesLocked()
	if err != nil {
		return nil, err
	}
	return file.Tombstones, nil
}

// BeginCardDeletion durably closes a card identity before configuration removal.
// The caller must coordinate references and authorization under admission.
func (r *FileRepository) BeginCardDeletion(expected IDCard) error {
	return r.BeginCardDeletionWithPending(expected, nil)
}

// BeginCardDeletionWithPending stores cleanup identities in the same atomic
// write as the deletion intent, before pending activations are cancelled.
func (r *FileRepository) BeginCardDeletionWithPending(expected IDCard, pending []CardDeletionPending) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := validateNewCardPending(pending); err != nil {
		return err
	}
	tombstones, err := r.loadCardTombstonesLocked()
	if err != nil {
		return err
	}
	for _, item := range tombstones.Tombstones {
		if item.ID == expected.ID {
			if item.DashboardID != expected.DashboardID {
				return errors.New("card tombstone owner mismatch")
			}
			return nil
		}
	}
	cards, err := r.LoadIDCardsFile()
	if err != nil {
		return err
	}
	var found bool
	expectedBytes, err := yaml.Marshal(expected)
	if err != nil {
		return err
	}
	for _, card := range cards.IDCards {
		if card.ID == expected.ID {
			actualBytes, marshalErr := yaml.Marshal(card)
			if marshalErr != nil {
				return marshalErr
			}
			found = bytes.Equal(actualBytes, expectedBytes)
			break
		}
	}
	if !found || expected.ID == "" || expected.DashboardID == "" {
		return errors.New("card changed or is unavailable")
	}
	tombstones.Tombstones = append(tombstones.Tombstones, CardTombstone{
		ID:          expected.ID,
		DashboardID: expected.DashboardID,
		Name:        expected.Name,
		State:       CardDeleting,
		DeletedAt:   time.Now().UTC(),
		Pending:     pending,
	})
	return r.WriteAtomic(filepath.Join(r.dir, CardTombstonesFilename), tombstones)
}

func validateNewCardPending(pending []CardDeletionPending) error {
	seen := make(map[string]bool)
	for _, item := range pending {
		if item.DashboardID == "" || seen[item.SessionID] || !validTemporaryIdentity(item.SessionID, item.TemporaryRoot, item.WorkingDir) ||
			item.Cleaned ||
			item.Error != "" {
			return errors.New("invalid pending card activation cleanup")
		}
		seen[item.SessionID] = true
	}
	return nil
}

// SaveCardDeletionPending checkpoints a cleanup result without changing identity.
func (r *FileRepository) SaveCardDeletionPending(cardID string, pending CardDeletionPending) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if pending.Cleaned && pending.Error != "" {
		return errors.New("completed card cleanup cannot retain an error")
	}
	file, err := r.loadCardTombstonesLocked()
	if err != nil {
		return err
	}
	for index, item := range file.Tombstones {
		if item.ID != cardID || item.State != CardDeleting {
			continue
		}
		for pendingIndex, previous := range item.Pending {
			if previous.SessionID != pending.SessionID {
				continue
			}
			if previous.DashboardID != pending.DashboardID || previous.WorkingDir != pending.WorkingDir ||
				previous.TemporaryRoot != pending.TemporaryRoot {
				return errors.New("pending card cleanup identity changed")
			}
			file.Tombstones[index].Pending[pendingIndex] = pending
			return r.WriteAtomic(filepath.Join(r.dir, CardTombstonesFilename), file)
		}
	}
	return errors.New("pending card cleanup not found")
}

// FinishCardDeletion removes configuration and marks a durable intent complete.
// Repeating it is safe after either write or a process interruption.
func (r *FileRepository) FinishCardDeletion(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	tombstones, err := r.loadCardTombstonesLocked()
	if err != nil {
		return err
	}
	index := slices.IndexFunc(tombstones.Tombstones, func(item CardTombstone) bool { return item.ID == id })
	if index < 0 {
		return errors.New("card deletion intent not found")
	}
	item := tombstones.Tombstones[index]
	if item.State == CardDeleted {
		return nil
	}
	for _, pending := range item.Pending {
		if !pending.Cleaned {
			return errors.New("card activation cleanup is incomplete")
		}
	}
	cards, err := r.LoadIDCardsFile()
	if err != nil {
		return err
	}
	for _, card := range cards.IDCards {
		if card.ID == id && card.DashboardID != item.DashboardID {
			return errors.New("card deletion owner mismatch")
		}
	}
	originalCount := len(cards.IDCards)
	cards.IDCards = slices.DeleteFunc(cards.IDCards, func(card IDCard) bool { return card.ID == id })
	if len(cards.IDCards) != originalCount {
		if errAt116 := r.WriteIDCardsFile(cards); errAt116 != nil {
			return errAt116
		}
	}
	tombstones.Tombstones[index].State = CardDeleted
	return r.WriteAtomic(filepath.Join(r.dir, CardTombstonesFilename), tombstones)
}

func (r *FileRepository) loadCardTombstonesLocked() (cardTombstonesFile, error) {
	path := filepath.Join(r.dir, CardTombstonesFilename)
	var file cardTombstonesFile
	if err := ReadYAML(path, &file); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cardTombstonesFile{SchemaVersion: 1, Tombstones: []CardTombstone{}}, nil
		}
		return cardTombstonesFile{}, err
	}
	if file.SchemaVersion != 1 {
		return cardTombstonesFile{}, errors.New("unsupported card tombstone schema")
	}
	seen := make(map[string]bool)
	for _, item := range file.Tombstones {
		if item.ID == "" || item.DashboardID == "" || (item.State != CardDeleting && item.State != CardDeleted) || seen[item.ID] {
			return cardTombstonesFile{}, fmt.Errorf("invalid card tombstone %q", item.ID)
		}
		seen[item.ID] = true
		pendingIDs := make(map[string]bool)
		for _, pending := range item.Pending {
			if pending.DashboardID == "" || pendingIDs[pending.SessionID] ||
				!validTemporaryIdentity(pending.SessionID, pending.TemporaryRoot, pending.WorkingDir) ||
				(pending.Cleaned && pending.Error != "") ||
				(item.State == CardDeleted && !pending.Cleaned) {
				return cardTombstonesFile{}, errors.New("invalid pending card cleanup")
			}
			pendingIDs[pending.SessionID] = true
		}
	}
	return file, nil
}

func (r *FileRepository) checkCardNotTombstonedLocked(id string) error {
	tombstones, err := r.loadCardTombstonesLocked()
	if err != nil {
		return err
	}
	for _, item := range tombstones.Tombstones {
		if item.ID == id {
			return errors.New("card is deleting or deleted")
		}
	}
	return nil
}
