package api

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/google/uuid"
)

const maxPendingConfirmations = 1024

type tombstoneWriter interface {
	BeginCardDeletion(config.IDCard) error
	BeginCardDeletionWithPending(config.IDCard, []config.CardDeletionPending) error
	SaveCardDeletionPending(string, config.CardDeletionPending) error
	FinishCardDeletion(string) error
}

// CardDeletionResult reports the durable state of a confirmed global deletion.
type CardDeletionResult struct{ State string }

const confirmationLifetime = 2 * time.Minute

type cardConfirmation struct {
	ownerID string
	cardID  string
	digest  [sha256.Size]byte
	expires time.Time
}

// CardDeletionPreflight describes all references to an owned card. Its token is
// a short-lived confirmation cue, not authorization for MCP or a later action.
type CardDeletionPreflight struct {
	Card                        config.IDCard
	Impact                      session.CardReferenceImpact
	RequiresForeignConfirmation bool
	Token                       string
}

// PreflightCardDeletion captures ownership, card revision and reference impact
// under admission. A deletion must recheck all three under that same lock.
//
//nolint:gocognit // Capture confirmation and references under one admission lock.
func (s *Service) PreflightCardDeletion(ownerID, cardID string) (CardDeletionPreflight, error) {
	var result CardDeletionPreflight
	err := s.sessions.WithCardReferencesAndSnapshot(cardID, func(impact session.CardReferenceImpact, sessions []session.Snapshot) error {
		card, err := s.cardOwnedBy(ownerID, cardID, impact, sessions)
		if err != nil {
			return err
		}
		digest, err := cardImpactDigest(card, impact)
		if err != nil {
			return err
		}
		token := uuid.NewString()
		s.confirmationMu.Lock()
		defer s.confirmationMu.Unlock()
		if s.confirmations == nil {
			s.confirmations = make(map[string]cardConfirmation)
		}
		now := time.Now()
		for key, confirmation := range s.confirmations {
			if !now.Before(confirmation.expires) {
				delete(s.confirmations, key)
			}
		}
		if len(s.confirmations) >= maxPendingConfirmations {
			return errors.New("too many pending card confirmations")
		}
		s.confirmations[token] = cardConfirmation{ownerID: ownerID, cardID: cardID, digest: digest, expires: now.Add(confirmationLifetime)}
		result = CardDeletionPreflight{Card: card, Impact: impact, Token: token}
		for _, ref := range impact.Sessions {
			if ref.DashboardID != ownerID {
				result.RequiresForeignConfirmation = true
			}
		}
		for dashboardID := range impact.PendingByDashboard {
			if dashboardID != ownerID {
				result.RequiresForeignConfirmation = true
			}
		}
		return nil
	})
	return result, err
}

// ConfirmCardDeletion persists a deletion intent after rechecking the token,
// then cancels and drains pending activations before removing configuration.
// A non-nil error after intent persistence leaves a retryable deleting state.
func (s *Service) ConfirmCardDeletion(
	ctx context.Context,
	ownerID, cardID, token string,
	foreignConfirmed bool,
) (CardDeletionResult, error) {
	if err := ctx.Err(); err != nil {
		return CardDeletionResult{}, err
	}
	writer, ok := s.repo.(tombstoneWriter)
	if !ok {
		return CardDeletionResult{}, errors.New("card deletion storage unavailable")
	}
	err := s.sessions.WithCardReferencesAndSnapshot(cardID, func(impact session.CardReferenceImpact, sessions []session.Snapshot) error {
		card, err := s.cardOwnedBy(ownerID, cardID, impact, sessions)
		if err != nil {
			return err
		}
		if errAt107 := s.validateCardConfirmationLocked(token, ownerID, card, impact); errAt107 != nil {
			return errAt107
		}
		if hasForeignCardReferences(impact, ownerID) && !foreignConfirmed {
			return errors.New("foreign card references require explicit confirmation")
		}
		if errAt113 := ctx.Err(); errAt113 != nil {
			return errAt113
		}
		pending, err := cardPendingCleanup(impact, sessions)
		if err != nil {
			return err
		}
		return writer.BeginCardDeletionWithPending(card, pending)
	})
	if err != nil {
		return CardDeletionResult{}, err
	}
	return s.RetryCardDeletion(ctx, ownerID, cardID)
}

// RetryCardDeletion resumes an existing durable intent; it never creates one.
func (s *Service) RetryCardDeletion(ctx context.Context, ownerID, cardID string) (CardDeletionResult, error) {
	writer, ok := s.repo.(tombstoneWriter)
	if !ok {
		return CardDeletionResult{}, errors.New("card deletion storage unavailable")
	}
	release, err := s.sessions.ReserveCardDeletion(cardID)
	if err != nil {
		return CardDeletionResult{}, err
	}
	defer release()
	workerCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		select {
		case <-s.sessions.ShutdownStarted():
			cancel()
		case <-workerCtx.Done():
		}
	}()
	result := CardDeletionResult{State: config.CardDeleting}
	if errAt155 := s.validateDeletionOwner(ownerID, cardID); errAt155 != nil {
		return CardDeletionResult{}, errAt155
	}
	s.sessions.CancelPendingCardReferences(cardID)
	if errAt159 := s.sessions.WaitCardReferences(workerCtx, cardID); errAt159 != nil {
		return result, fmt.Errorf("card deletion is pending: %w", errAt159)
	}
	if errAt162 := s.cleanupCardPending(workerCtx, writer, ownerID, cardID); errAt162 != nil {
		return result, errAt162
	}
	err = s.sessions.WithCardReferencesAndSnapshot(cardID, func(impact session.CardReferenceImpact, sessions []session.Snapshot) error {
		if len(impact.PendingByDashboard) != 0 {
			return errors.New("card activations are still pending")
		}
		if errAt169 := s.validateDeletionInventory(ownerID, cardID, sessions); errAt169 != nil {
			return errAt169
		}
		return writer.FinishCardDeletion(cardID)
	})
	if err != nil {
		return result, err
	}
	result.State = config.CardDeleted
	return result, nil
}

// cardPendingCleanup records only new temporary workspaces; pending switches or
// relaunches belong to existing sessions and must never remove their workspace.
func cardPendingCleanup(impact session.CardReferenceImpact, sessions []session.Snapshot) ([]config.CardDeletionPending, error) {
	published := make(map[string]bool, len(sessions))
	for _, item := range sessions {
		published[item.ID] = true
	}
	var pending []config.CardDeletionPending
	for _, item := range impact.Pending {
		if published[item.SessionID] || (item.TemporaryRoot == "" && item.WorkingDir == "") {
			continue
		}
		if item.TemporaryRoot == "" || item.WorkingDir == "" {
			return nil, errors.New("pending card activation cleanup path is unavailable")
		}
		pending = append(pending, config.CardDeletionPending{DashboardID: item.DashboardID, SessionID: item.SessionID,
			TemporaryRoot: item.TemporaryRoot, WorkingDir: item.WorkingDir})
	}
	return pending, nil
}

func (s *Service) cleanupCardPending(ctx context.Context, writer tombstoneWriter, ownerID, cardID string) error {
	tombstones, err := s.repo.ListCardTombstones()
	if err != nil {
		return err
	}
	for _, item := range tombstones {
		if item.ID != cardID || item.DashboardID != ownerID {
			continue
		}
		for _, pending := range item.Pending {
			if pending.Cleaned {
				continue
			}
			_, cleanupErr := s.sessions.RemoveResidual(ctx, pending.SessionID, pending.WorkingDir, pending.TemporaryRoot, true)
			if cleanupErr != nil {
				pending.Error = cleanupErr.Error()
				return errors.Join(
					fmt.Errorf("card activation cleanup is pending: %w", cleanupErr),
					writer.SaveCardDeletionPending(cardID, pending),
				)
			}
			pending.Cleaned, pending.Error = true, ""
			if errAt228 := writer.SaveCardDeletionPending(cardID, pending); errAt228 != nil {
				return errAt228
			}
		}
		return nil
	}
	return errors.New("card deletion intent not found")
}

func (s *Service) validateDeletionOwner(ownerID, cardID string) error {
	return s.sessions.WithCardReferencesAndSnapshot(cardID, func(_ session.CardReferenceImpact, sessions []session.Snapshot) error {
		return s.validateDeletionInventory(ownerID, cardID, sessions)
	})
}

func (s *Service) validateDeletionInventory(ownerID, cardID string, sessions []session.Snapshot) error {
	inventory, err := s.loadMembershipInventory(sessions)
	if err != nil {
		return err
	}
	_, ownerExists := inventory.dashboards[ownerID]
	tombstone, exists := inventory.tombstones[cardID]
	if !ownerExists || !exists || tombstone.DashboardID != ownerID {
		return errors.New("card deletion is not owned by this dashboard")
	}
	return nil
}

// validateCardConfirmationLocked is called only under the manager admission
// lock immediately before a durable deletion intent is written.
func (s *Service) validateCardConfirmationLocked(token, ownerID string, card config.IDCard, impact session.CardReferenceImpact) error {
	digest, err := cardImpactDigest(card, impact)
	if err != nil {
		return err
	}
	s.confirmationMu.Lock()
	defer s.confirmationMu.Unlock()
	confirmation, exists := s.confirmations[token]
	if !exists || !time.Now().Before(confirmation.expires) || confirmation.ownerID != ownerID || confirmation.cardID != card.ID ||
		confirmation.digest != digest {
		return errors.New("card deletion impact changed or confirmation expired; review it again")
	}
	return nil
}

func (s *Service) cardOwnedBy(
	ownerID, cardID string,
	impact session.CardReferenceImpact,
	sessions []session.Snapshot,
) (config.IDCard, error) {
	inventory, err := s.loadMembershipInventory(sessions)
	if err != nil {
		return config.IDCard{}, err
	}
	card, exists := inventory.cards[cardID]
	if !exists || inventory.hasTombstone(cardID) || card.DashboardID != ownerID || !inventory.activeDashboard(ownerID) {
		return config.IDCard{}, errors.New("card is not owned by an active dashboard")
	}
	for _, ref := range impact.Sessions {
		if _, existsAt204 := inventory.dashboards[ref.DashboardID]; !existsAt204 {
			return config.IDCard{}, fmt.Errorf("session %s has invalid dashboard membership", ref.SessionID)
		}
	}
	for dashboardID, count := range impact.PendingByDashboard {
		if _, existsAt209 := inventory.dashboards[dashboardID]; !existsAt209 || count <= 0 {
			return config.IDCard{}, errors.New("pending card reference has invalid dashboard membership")
		}
	}
	for _, pending := range impact.Pending {
		if pending.SessionID == "" || !pending.Cancellable {
			return config.IDCard{}, errors.New("pending card reference has unknown session identity")
		}
	}
	return card, nil
}

func cardImpactDigest(card config.IDCard, impact session.CardReferenceImpact) ([sha256.Size]byte, error) {
	raw, err := json.Marshal(struct { //nolint:musttag // Internal digest includes existing YAML-backed card fields.
		Card   config.IDCard               `json:"card"`
		Impact session.CardReferenceImpact `json:"impact"`
	}{card, impact})
	if err != nil {
		return [sha256.Size]byte{}, err
	}
	return sha256.Sum256(raw), nil
}
