package session

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

type cardReservation struct{ _ byte }

// ReserveCardReferenceForDashboardWithCancel registers a cancellable pending
// activation for a future deletion drain.
func (m *Manager) ReserveCardReferenceForDashboardWithCancel(
	cardID, dashboardID, sessionID string,
	cancel context.CancelFunc,
	check func([]Snapshot) (bool, error),
) (func(), error) {
	if strings.TrimSpace(dashboardID) == "" {
		return nil, errors.New("dashboard ID is required")
	}
	if check == nil {
		return nil, errors.New("card reference check is required")
	}
	return m.reserveCardReference(cardID, dashboardID, sessionID, cancel, func() (bool, error) { return check(m.listLocked()) })
}

//nolint:gocognit // Keep reservation and cancellation checks in the same critical section.
func (m *Manager) reserveCardReference(
	cardID, dashboardID, sessionID string,
	cancel context.CancelFunc,
	check func() (bool, error),
) (func(), error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return nil, ErrManagerClosed
	}

	cardExists, err := check()
	if err != nil {
		return nil, fmt.Errorf("check card before activation: %w", err)
	}
	if !cardExists {
		return nil, fmt.Errorf("card not found: %s", cardID)
	}
	if m.pendingCardRefs == nil {
		m.pendingCardRefs = make(map[string]int)
	}
	if m.pendingCardRefsByDashboard == nil {
		m.pendingCardRefsByDashboard = make(map[string]map[string]int)
	}
	m.pendingCardRefs[cardID]++
	m.cardReferenceRevision++
	if m.pendingCardRefsByDashboard[cardID] == nil {
		m.pendingCardRefsByDashboard[cardID] = make(map[string]int)
	}
	m.pendingCardRefsByDashboard[cardID][dashboardID]++
	reservation := &cardReservation{}
	if m.pendingCardDetails == nil {
		m.pendingCardDetails = make(map[string]map[*cardReservation]CardPendingReference)
	}
	if m.pendingCardDetails[cardID] == nil {
		m.pendingCardDetails[cardID] = make(map[*cardReservation]CardPendingReference)
	}
	m.pendingCardDetails[cardID][reservation] = CardPendingReference{
		DashboardID: dashboardID,
		SessionID:   sessionID,
		Cancellable: cancel != nil,
	}
	if cancel != nil {
		if m.pendingCardCancels == nil {
			m.pendingCardCancels = make(map[string]map[*cardReservation]context.CancelFunc)
		}
		if m.pendingCardCancels[cardID] == nil {
			m.pendingCardCancels[cardID] = make(map[*cardReservation]context.CancelFunc)
		}
		m.pendingCardCancels[cardID][reservation] = cancel
	}
	m.notifyCardReferencesLocked()

	var once sync.Once
	return func() {
		once.Do(func() {
			m.mu.Lock()
			defer m.mu.Unlock()
			m.cardReferenceRevision++
			delete(m.pendingCardDetails[cardID], reservation)
			if len(m.pendingCardDetails[cardID]) == 0 {
				delete(m.pendingCardDetails, cardID)
			}
			if cancel != nil {
				delete(m.pendingCardCancels[cardID], reservation)
				if len(m.pendingCardCancels[cardID]) == 0 {
					delete(m.pendingCardCancels, cardID)
				}
			}
			m.notifyCardReferencesLocked()
			byDashboard := m.pendingCardRefsByDashboard[cardID]
			if byDashboard[dashboardID] <= 1 {
				delete(byDashboard, dashboardID)
			} else {
				byDashboard[dashboardID]--
			}
			if len(byDashboard) == 0 {
				delete(m.pendingCardRefsByDashboard, cardID)
			}
			if m.pendingCardRefs[cardID] <= 1 {
				delete(m.pendingCardRefs, cardID)
				return
			}
			m.pendingCardRefs[cardID]--
		})
	}, nil
}

func (m *Manager) notifyCardReferencesLocked() {
	if m.cardRefsChanged != nil {
		close(m.cardRefsChanged)
	}
	m.cardRefsChanged = make(chan struct{})
}

// CancelPendingCardReferences asks in-flight scoped activations to stop. It
// invokes cancellation after releasing the manager lock.
func (m *Manager) CancelPendingCardReferences(cardID string) {
	m.mu.RLock()
	cancels := make([]context.CancelFunc, 0, len(m.pendingCardCancels[cardID]))
	for _, cancel := range m.pendingCardCancels[cardID] {
		cancels = append(cancels, cancel)
	}
	m.mu.RUnlock()
	for _, cancel := range cancels {
		cancel()
	}
}

// WaitCardReferences waits without holding the manager lock until all pending
// references are released or the caller cancels its wait.
func (m *Manager) WaitCardReferences(ctx context.Context, cardID string) error {
	for {
		m.mu.Lock()
		if m.pendingCardRefs[cardID] == 0 {
			m.mu.Unlock()
			return nil
		}
		if m.cardRefsChanged == nil {
			m.cardRefsChanged = make(chan struct{})
		}
		changed := m.cardRefsChanged
		m.mu.Unlock()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-changed:
		}
	}
}

// CardSessionReference identifies current-card and saved-binding references.
type CardSessionReference struct {
	SessionID   string
	DashboardID string
	Current     bool
	Binding     bool
}

// CardPendingReference identifies a reserved activation before publication.
type CardPendingReference struct {
	DashboardID   string
	SessionID     string
	Cancellable   bool
	WorkingDir    string
	TemporaryRoot string
}

// bindPendingWorkspace records the exact path before a scoped new launch can
// create it. The availability check shares the deletion admission lock.
func (m *Manager) bindPendingWorkspace(
	ctx context.Context,
	cardID, sessionID, workspace string,
	useTmp bool,
	check func([]Snapshot) error,
) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return ErrManagerClosed
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := check(m.listLocked()); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	for reservation, pending := range m.pendingCardDetails[cardID] {
		if pending.SessionID != sessionID {
			continue
		}
		if useTmp {
			root, err := filepath.Abs(filepath.Join(m.tmpRoot(), sessionID))
			if err != nil {
				return err
			}
			pending.TemporaryRoot = root
			pending.WorkingDir = filepath.Join(root, filepath.Base(workspace))
		}
		m.pendingCardDetails[cardID][reservation] = pending
		m.cardReferenceRevision++
		m.notifyCardReferencesLocked()
		return nil
	}
	return errors.New("pending card activation reservation not found")
}

// CardReferenceImpact is a snapshot; confirmations must recheck it under the
// admission lock before any deletion.
type CardReferenceImpact struct {
	Sessions           []CardSessionReference
	PendingByDashboard map[string]int
	Pending            []CardPendingReference
	Revision           uint64
}

// CardReferences returns all published and pending references to a card.
func (m *Manager) CardReferences(cardID string) CardReferenceImpact {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cardReferencesLocked(cardID)
}

// WithCardReferencesAndSnapshot supplies the complete session inventory under
// the admission lock for fail-closed membership validation.
func (m *Manager) WithCardReferencesAndSnapshot(cardID string, check func(CardReferenceImpact, []Snapshot) error) error {
	if check == nil {
		return errors.New("card reference check is required")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return ErrManagerClosed
	}
	return check(m.cardReferencesLocked(cardID), m.listLocked())
}

func (m *Manager) cardReferencesLocked(cardID string) CardReferenceImpact {
	impact := CardReferenceImpact{PendingByDashboard: maps.Clone(m.pendingCardRefsByDashboard[cardID]), Revision: m.cardReferenceRevision}
	for _, pending := range m.pendingCardDetails[cardID] {
		impact.Pending = append(impact.Pending, pending)
	}
	sort.Slice(impact.Pending, func(i, j int) bool {
		if impact.Pending[i].DashboardID == impact.Pending[j].DashboardID {
			return impact.Pending[i].SessionID < impact.Pending[j].SessionID
		}
		return impact.Pending[i].DashboardID < impact.Pending[j].DashboardID
	})
	for _, current := range m.sessions {
		_, binding := current.Conversations[cardID]
		if current.CardID == cardID || binding {
			impact.Sessions = append(impact.Sessions, CardSessionReference{
				SessionID: current.ID, DashboardID: current.DashboardID, Current: current.CardID == cardID, Binding: binding,
			})
		}
	}
	sort.Slice(impact.Sessions, func(i, j int) bool { return impact.Sessions[i].SessionID < impact.Sessions[j].SessionID })
	return impact
}
