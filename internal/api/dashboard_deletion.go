package api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/session"
)

type dashboardDeletionWriter interface {
	BeginDashboardDeletion(config.DashboardDeletion) error
	SaveDashboardDeletion(config.DashboardDeletion) error
	FinalizeDashboardDeletion(string) error
}

var errDashboardDeletionNotFound = errors.New("dashboard deletion manifest not found")

// DashboardDeletionResult describes durable cascade progress.
type DashboardDeletionResult struct {
	State    string
	Manifest config.DashboardDeletion
}

// ConfirmDashboardDeletion writes the complete manifest under admission before
// any teardown, after rechecking the token and explicit foreign-impact consent.
func (s *Service) ConfirmDashboardDeletion(
	ctx context.Context,
	dashboardID, token string,
	foreignConfirmed bool,
) (DashboardDeletionResult, error) {
	if err := ctx.Err(); err != nil {
		return DashboardDeletionResult{}, err
	}
	writer, ok := s.repo.(dashboardDeletionWriter)
	if !ok {
		return DashboardDeletionResult{}, errors.New("dashboard deletion storage unavailable")
	}
	var manifest config.DashboardDeletion
	err := s.sessions.WithAdmissionSnapshot(func(snapshot session.AdmissionSnapshot) error {
		current, foreign, errAt55 := s.prepareDashboardDeletion(dashboardID, snapshot)
		if errAt55 != nil {
			return errAt55
		}
		digest, errAt55 := dashboardImpactDigest(current, snapshot.Revision)
		if errAt55 != nil {
			return errAt55
		}
		s.confirmationMu.Lock()
		confirmation, exists := s.dashboardConfirmations[token]
		s.confirmationMu.Unlock()
		if !exists || confirmation.dashboardID != dashboardID || !time.Now().Before(confirmation.expires) || confirmation.digest != digest {
			return errors.New("dashboard deletion impact changed or confirmation expired; review it again")
		}
		if foreign && !foreignConfirmed {
			return errors.New("foreign card references require explicit confirmation")
		}
		if errAt72 := ctx.Err(); errAt72 != nil {
			return errAt72
		}
		if errAt75 := writer.BeginDashboardDeletion(current); errAt75 != nil {
			return errAt75
		}
		manifest = current
		return nil
	})
	if err != nil {
		return DashboardDeletionResult{}, err
	}
	result, err := s.RetryDashboardDeletion(ctx, dashboardID)
	if result.State == "" {
		result = DashboardDeletionResult{State: config.CardDeleting, Manifest: manifest}
	}
	return result, err
}

//nolint:gocognit // Each cleanup effect is checked and checkpointed.
func (s *Service) removeDashboardSession(
	ctx context.Context,
	dashboardID string,
	item *config.DashboardDeletionSession,
	writer dashboardDeletionWriter,
	manifest *config.DashboardDeletion,
) error {
	snapshot, exists := s.sessions.GetSession(item.ID)
	//nolint:nestif // Check and persist each cleanup effect before proceeding.
	if exists {
		working := snapshot.WorkingDir
		if snapshot.UseTmpWorkspace {
			var err error
			working, err = filepath.Abs(working)
			if err != nil {
				return err
			}
		}
		if snapshot.DashboardID != dashboardID || snapshot.UseTmpWorkspace != item.UseTmpWorkspace || working != item.WorkingDir {
			return errors.New("session identity or workspace changed after confirmation")
		}
		if snapshot.Status != session.StatusTerminated {
			stopped, err := s.sessions.KillDetailed(item.ID)
			item.Stopped = stopped.ProcessTerminated
			if err != nil {
				item.Error = err.Error()
				return errors.Join(err, writer.SaveDashboardDeletion(*manifest))
			}
			if !stopped.ProcessTerminated {
				return errors.New("session process remains active")
			}
		} else {
			item.Stopped = true
		}
		item.Error = ""
		if err := writer.SaveDashboardDeletion(*manifest); err != nil {
			return err
		}
		removed, err := s.sessions.RemoveDetailed(item.ID)
		item.RecordDeleted = removed.SessionRecordDeleted
		item.TemporaryRemoved = !item.UseTmpWorkspace || removed.TemporaryWorkspaceRemoved
		if err != nil {
			item.Error = err.Error()
			return errors.Join(err, writer.SaveDashboardDeletion(*manifest))
		}
	} else {
		item.Stopped = true
		removed, err := s.sessions.RemoveResidual(ctx, item.ID, item.WorkingDir, item.TemporaryRoot, item.UseTmpWorkspace)
		item.RecordDeleted = removed.SessionRecordDeleted
		item.TemporaryRemoved = !item.UseTmpWorkspace || removed.TemporaryWorkspaceRemoved
		if err != nil {
			item.Error = err.Error()
			return errors.Join(err, writer.SaveDashboardDeletion(*manifest))
		}
	}
	if !item.RecordDeleted || !item.TemporaryRemoved {
		return errors.New("session cleanup effects are incomplete")
	}
	item.Removed, item.Error = true, ""
	return writer.SaveDashboardDeletion(*manifest)
}

// RetryDashboardDeletion resumes a persisted cascade without replaying launches.
func (s *Service) RetryDashboardDeletion(ctx context.Context, dashboardID string) (DashboardDeletionResult, error) {
	writer, ok := s.repo.(dashboardDeletionWriter)
	if !ok {
		return DashboardDeletionResult{}, errors.New("dashboard deletion storage unavailable")
	}
	release, err := s.sessions.ReserveDashboardDeletion(dashboardID)
	if err != nil {
		return DashboardDeletionResult{}, err
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
	manifest, err := s.dashboardDeletion(dashboardID)
	if err != nil {
		if !errors.Is(err, errDashboardDeletionNotFound) {
			return DashboardDeletionResult{}, err
		}
		deleted, checkErr := s.dashboardAlreadyDeleted(dashboardID)
		if checkErr != nil {
			return DashboardDeletionResult{}, checkErr
		}
		if deleted {
			return DashboardDeletionResult{State: config.CardDeleted}, nil
		}
		return DashboardDeletionResult{}, err
	}
	result := DashboardDeletionResult{State: config.CardDeleting, Manifest: manifest}
	err = s.runDashboardDeletion(workerCtx, writer, &result.Manifest)
	if err != nil {
		return result, err
	}
	result.State = config.CardDeleted
	return result, nil
}

func (s *Service) dashboardAlreadyDeleted(id string) (bool, error) {
	items, err := s.repo.ListDashboardTombstones()
	if err != nil {
		return false, err
	}
	found := false
	for _, item := range items {
		if item.ID == id {
			found = true
			break
		}
	}
	if !found {
		return false, nil
	}
	dashboards, err := s.ListDashboards()
	if err != nil {
		return false, err
	}
	for _, dashboard := range dashboards {
		if dashboard.ID == id {
			return false, nil
		}
	}
	return true, nil
}

func (s *Service) dashboardDeletion(id string) (config.DashboardDeletion, error) {
	manifests, err := s.repo.ListDashboardDeletions()
	if err != nil {
		return config.DashboardDeletion{}, err
	}
	for _, manifest := range manifests {
		if manifest.DashboardID == id {
			return manifest, nil
		}
	}
	return config.DashboardDeletion{}, errDashboardDeletionNotFound
}

// GetDashboardDeletion returns the durable per-item recovery state for humans.
func (s *Service) GetDashboardDeletion(id string) (config.DashboardDeletion, error) {
	return s.dashboardDeletion(id)
}

//nolint:gocognit // Each irreversible effect is checkpointed before proceeding.
func (s *Service) runDashboardDeletion(ctx context.Context, writer dashboardDeletionWriter, manifest *config.DashboardDeletion) error {
	ownedCards := make(map[string]bool, len(manifest.Cards))
	for _, card := range manifest.Cards {
		ownedCards[card.ID] = true
	}
	//nolint:nestif // Drain all reservations before irreversible effects.
	if !manifest.AdmissionsDrained {
		s.sessions.CancelDashboardAdmissions(manifest.DashboardID, ownedCards)
		if err := s.sessions.WaitDashboardAdmissions(ctx, manifest.DashboardID, ownedCards); err != nil {
			return fmt.Errorf("dashboard deletion is waiting for activations: %w", err)
		}
		for index := range manifest.Pending {
			pending := &manifest.Pending[index]
			if pending.Cleaned {
				continue
			}
			if pending.TemporaryRoot != "" {
				if _, err := s.sessions.RemoveResidual(
					ctx,
					pending.SessionID,
					pending.WorkingDir,
					pending.TemporaryRoot,
					true,
				); err != nil {
					pending.Error = err.Error()
					return errors.Join(err, writer.SaveDashboardDeletion(*manifest))
				}
			}
			pending.Cleaned, pending.Error = true, ""
			if err := writer.SaveDashboardDeletion(*manifest); err != nil {
				return err
			}
		}
		manifest.AdmissionsDrained = true
		if err := writer.SaveDashboardDeletion(*manifest); err != nil {
			return err
		}
	}
	for index := range manifest.Sessions {
		item := &manifest.Sessions[index]
		if item.Removed {
			if _, exists := s.sessions.GetSession(item.ID); exists {
				return errors.New("removed dashboard session unexpectedly reappeared")
			}
			if _, err := s.sessions.RemoveResidual(ctx, item.ID, item.WorkingDir, item.TemporaryRoot, item.UseTmpWorkspace); err != nil {
				item.Error = err.Error()
				return errors.Join(err, writer.SaveDashboardDeletion(*manifest))
			}
			continue
		}
		if err := s.removeDashboardSession(ctx, manifest.DashboardID, item, writer, manifest); err != nil {
			return err
		}
	}
	for index := range manifest.Cards {
		item := &manifest.Cards[index]
		if err := s.deleteDashboardCard(manifest.DashboardID, item); err != nil {
			item.Error = err.Error()
			return errors.Join(err, writer.SaveDashboardDeletion(*manifest))
		}
		item.Deleted, item.Error = true, ""
		if err := writer.SaveDashboardDeletion(*manifest); err != nil {
			return err
		}
	}
	if !manifest.HistoryRemoved {
		store := s.sessions.Messages()
		if store == nil {
			return errors.New("dashboard message history unavailable")
		}
		if err := store.DeleteDashboard(manifest.DashboardID); err != nil {
			return err
		}
		manifest.HistoryRemoved = true
		if err := writer.SaveDashboardDeletion(*manifest); err != nil {
			return err
		}
	}
	return writer.FinalizeDashboardDeletion(manifest.DashboardID)
}

func (s *Service) deleteDashboardCard(ownerID string, item *config.DashboardDeletionCard) error {
	writer, ok := s.repo.(tombstoneWriter)
	if !ok {
		return errors.New("card deletion storage unavailable")
	}
	return s.sessions.WithCardReferencesAndSnapshot(item.ID, func(impact session.CardReferenceImpact, sessions []session.Snapshot) error {
		if len(impact.PendingByDashboard) != 0 {
			return errors.New("card activations are still pending")
		}
		inventory, err := s.loadMembershipInventory(sessions)
		if err != nil {
			return err
		}
		if tombstone, exists := inventory.tombstones[item.ID]; exists {
			if tombstone.DashboardID != ownerID {
				return errors.New("card tombstone owner mismatch")
			}
			return writer.FinishCardDeletion(item.ID)
		}
		card, exists := inventory.cards[item.ID]
		if !exists || card.DashboardID != ownerID {
			return errors.New("owned card is unavailable")
		}
		raw, err := json.Marshal(card) //nolint:musttag // Compare the same internal digest as preflight.
		if err != nil {
			return err
		}
		hash := sha256.Sum256(raw)
		if hex.EncodeToString(hash[:]) != item.Revision {
			return errors.New("owned card changed after confirmation")
		}
		if errAt360 := writer.BeginCardDeletion(card); errAt360 != nil {
			return errAt360
		}
		return writer.FinishCardDeletion(item.ID)
	})
}
