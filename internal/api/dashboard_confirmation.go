package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"time"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/google/uuid"
)

type dashboardConfirmation struct {
	dashboardID string
	digest      [sha256.Size]byte
	expires     time.Time
}

// DashboardDeletionPreflight lists owned sessions and cards, including foreign
// references to those cards. The token must be rechecked before writing intent.
type DashboardDeletionPreflight struct {
	Manifest                    config.DashboardDeletion
	RequiresForeignConfirmation bool
	Token                       string
}

// PreflightDashboardDeletion inventories a cascade under the admission lock.
func (s *Service) PreflightDashboardDeletion(dashboardID string) (DashboardDeletionPreflight, error) {
	var result DashboardDeletionPreflight
	err := s.sessions.WithAdmissionSnapshot(func(snapshot session.AdmissionSnapshot) error {
		manifest, foreign, errAt40 := s.prepareDashboardDeletion(dashboardID, snapshot)
		if errAt40 != nil {
			return errAt40
		}
		digest, errAt40 := dashboardImpactDigest(manifest, snapshot.Revision)
		if errAt40 != nil {
			return errAt40
		}
		token := uuid.NewString()
		s.confirmationMu.Lock()
		defer s.confirmationMu.Unlock()
		if s.dashboardConfirmations == nil {
			s.dashboardConfirmations = make(map[string]dashboardConfirmation)
		}
		now := time.Now()
		for key, confirmation := range s.dashboardConfirmations {
			if !now.Before(confirmation.expires) {
				delete(s.dashboardConfirmations, key)
			}
		}
		if len(s.dashboardConfirmations) >= maxPendingConfirmations {
			return errors.New("too many pending dashboard confirmations")
		}
		s.dashboardConfirmations[token] = dashboardConfirmation{
			dashboardID: dashboardID,
			digest:      digest,
			expires:     now.Add(confirmationLifetime),
		}
		result = DashboardDeletionPreflight{Manifest: manifest, RequiresForeignConfirmation: foreign, Token: token}
		return nil
	})
	return result, err
}

//nolint:gocognit,funlen // The complete cascade inventory is captured in one locked snapshot.
func (s *Service) prepareDashboardDeletion(
	dashboardID string,
	snapshot session.AdmissionSnapshot,
) (config.DashboardDeletion, bool, error) {
	inventory, err := s.loadMembershipInventory(snapshot.Sessions)
	if err != nil {
		return config.DashboardDeletion{}, false, err
	}
	dashboard, exists := inventory.dashboards[dashboardID]
	if !exists || !inventory.activeDashboard(dashboardID) {
		return config.DashboardDeletion{}, false, errors.New("dashboard unavailable")
	}
	manifest := config.DashboardDeletion{
		SchemaVersion: 1,
		DashboardID:   dashboardID,
		Name:          dashboard.Name,
		Description:   dashboard.Description,
	}
	published := make(map[string]bool)
	for _, item := range snapshot.Sessions {
		published[item.ID] = true
		if item.DashboardID != dashboardID {
			continue
		}
		entry := config.DashboardDeletionSession{ID: item.ID, Name: item.Name, Status: string(item.Status), WorkingDir: item.WorkingDir,
			UseTmpWorkspace: item.UseTmpWorkspace, Stopped: item.Status == session.StatusTerminated}
		if item.UseTmpWorkspace {
			entry.WorkingDir, err = filepath.Abs(item.WorkingDir)
			if err != nil {
				return config.DashboardDeletion{}, false, err
			}
			entry.TemporaryRoot = filepath.Dir(entry.WorkingDir)
			if filepath.Base(entry.TemporaryRoot) != item.ID {
				return config.DashboardDeletion{}, false, fmt.Errorf("session %s has invalid temporary path", item.ID)
			}
		}
		manifest.Sessions = append(manifest.Sessions, entry)
	}
	ownedCards := make(map[string]bool)
	foreign := false
	for _, card := range inventory.cards {
		if card.DashboardID != dashboardID {
			continue
		}
		if inventory.hasTombstone(card.ID) {
			return config.DashboardDeletion{}, false, errors.New("an owned card deletion is already in progress")
		}
		ownedCards[card.ID] = true
		raw, marshalErr := json.Marshal(card) //nolint:musttag // Internal digest uses the existing YAML-backed card fields.
		if marshalErr != nil {
			return config.DashboardDeletion{}, false, marshalErr
		}
		hash := sha256.Sum256(raw)
		entry := config.DashboardDeletionCard{ID: card.ID, Name: card.Name, Shared: card.Shared, Revision: hex.EncodeToString(hash[:])}
		impact := snapshot.CardImpacts[card.ID]
		for _, ref := range impact.Sessions {
			entry.References = append(
				entry.References,
				config.DashboardDeletionReference{
					DashboardID: ref.DashboardID,
					SessionID:   ref.SessionID,
					Current:     ref.Current,
					Binding:     ref.Binding,
				},
			)
			if ref.DashboardID != dashboardID {
				foreign = true
			}
		}
		for _, pending := range impact.Pending {
			entry.References = append(
				entry.References,
				config.DashboardDeletionReference{DashboardID: pending.DashboardID, SessionID: pending.SessionID, Pending: true},
			)
			if pending.DashboardID != dashboardID {
				foreign = true
			}
		}
		manifest.Cards = append(manifest.Cards, entry)
	}
	sort.Slice(manifest.Cards, func(i, j int) bool { return manifest.Cards[i].ID < manifest.Cards[j].ID })
	for cardID, impact := range snapshot.CardImpacts {
		for _, pending := range impact.Pending {
			if _, existsAt170 := inventory.dashboards[pending.DashboardID]; !existsAt170 || pending.SessionID == "" ||
				!pending.Cancellable {
				return config.DashboardDeletion{}, false, errors.New("pending activation has invalid membership, identity or cancellation")
			}
			if pending.DashboardID != dashboardID && !ownedCards[cardID] {
				continue
			}
			_, existsAt176 := inventory.cards[cardID]
			if !existsAt176 {
				return config.DashboardDeletion{}, false, errors.New("pending activation references unavailable card")
			}
			entry := config.DashboardDeletionPending{CardID: cardID, DashboardID: pending.DashboardID, SessionID: pending.SessionID}
			if !published[pending.SessionID] {
				if (pending.WorkingDir == "") != (pending.TemporaryRoot == "") {
					return config.DashboardDeletion{}, false, errors.New("pending activation cleanup path is unavailable")
				}
				entry.TemporaryRoot = pending.TemporaryRoot
				entry.WorkingDir = pending.WorkingDir
			}
			manifest.Pending = append(manifest.Pending, entry)
		}
	}
	sort.Slice(manifest.Pending, func(i, j int) bool {
		a, b := manifest.Pending[i], manifest.Pending[j]
		if a.CardID != b.CardID {
			return a.CardID < b.CardID
		}
		if a.DashboardID != b.DashboardID {
			return a.DashboardID < b.DashboardID
		}
		return a.SessionID < b.SessionID
	})
	return manifest, foreign, nil
}

func dashboardImpactDigest(manifest config.DashboardDeletion, revision uint64) ([sha256.Size]byte, error) {
	raw, err := json.Marshal(struct {
		Manifest config.DashboardDeletion `json:"manifest"`
		Revision uint64                   `json:"revision"`
	}{manifest, revision})
	if err != nil {
		return [sha256.Size]byte{}, err
	}
	return sha256.Sum256(raw), nil
}
