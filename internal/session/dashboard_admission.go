package session

import (
	"context"
	"errors"
	"sort"
	"sync"
)

// ReserveCardDeletion serializes cleanup retries across service instances.
func (m *Manager) ReserveCardDeletion(id string) (func(), error) {
	if id == "" {
		return nil, errors.New("card ID is required")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return nil, ErrManagerClosed
	}
	if m.cardDeletionWorkers == nil {
		m.cardDeletionWorkers = make(map[string]bool)
	}
	if m.cardDeletionWorkers[id] {
		return nil, errors.New("card deletion is already running")
	}
	m.cardDeletionWorkers[id] = true
	m.activations.Add(1)
	var once sync.Once
	return func() {
		once.Do(func() { m.mu.Lock(); delete(m.cardDeletionWorkers, id); m.mu.Unlock(); m.activations.Done() })
	}, nil
}

// ReserveDashboardDeletion prevents concurrent cascade workers across service
// instances. A durable manifest remains authoritative after process restart.
func (m *Manager) ReserveDashboardDeletion(id string) (func(), error) {
	if id == "" {
		return nil, errors.New("dashboard ID is required")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return nil, ErrManagerClosed
	}
	if m.dashboardDeletionWorkers == nil {
		m.dashboardDeletionWorkers = make(map[string]bool)
	}
	if m.dashboardDeletionWorkers[id] {
		return nil, errors.New("dashboard deletion is already running")
	}
	m.dashboardDeletionWorkers[id] = true
	m.activations.Add(1)
	var once sync.Once
	return func() {
		once.Do(func() { m.mu.Lock(); delete(m.dashboardDeletionWorkers, id); m.mu.Unlock(); m.activations.Done() })
	}, nil
}

// AdmissionSnapshot is a point-in-time view captured under the manager lock.
type AdmissionSnapshot struct {
	Sessions    []Snapshot
	CardImpacts map[string]CardReferenceImpact
	Revision    uint64
}

// WithAdmissionSnapshot serializes short validation and durable intent writes
// with all card reservations. The callback must not call the manager or wait.
func (m *Manager) WithAdmissionSnapshot(check func(AdmissionSnapshot) error) error {
	if check == nil {
		return errors.New("admission check is required")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return ErrManagerClosed
	}
	snapshot := AdmissionSnapshot{
		Sessions:    m.listLocked(),
		CardImpacts: make(map[string]CardReferenceImpact),
		Revision:    m.cardReferenceRevision,
	}
	sort.Slice(snapshot.Sessions, func(i, j int) bool { return snapshot.Sessions[i].ID < snapshot.Sessions[j].ID })
	ids := make(map[string]bool)
	for id := range m.pendingCardRefs {
		ids[id] = true
	}
	for _, item := range m.sessions {
		ids[item.CardID] = true
		for id := range item.Conversations {
			ids[id] = true
		}
	}
	for id := range ids {
		snapshot.CardImpacts[id] = m.cardReferencesLocked(id)
	}
	return check(snapshot)
}

// CancelDashboardAdmissions requests cancellation for activations owned by the
// dashboard and activations in other dashboards using its owned cards.
func (m *Manager) CancelDashboardAdmissions(dashboardID string, ownedCards map[string]bool) {
	m.mu.RLock()
	var cancels []context.CancelFunc
	for cardID, entries := range m.pendingCardDetails {
		for reservation, reference := range entries {
			if reference.DashboardID == dashboardID || ownedCards[cardID] {
				if cancel := m.pendingCardCancels[cardID][reservation]; cancel != nil {
					cancels = append(cancels, cancel)
				}
			}
		}
	}
	m.mu.RUnlock()
	for _, cancel := range cancels {
		cancel()
	}
}

// WaitDashboardAdmissions drains the same reservation set without holding locks.
func (m *Manager) WaitDashboardAdmissions(ctx context.Context, dashboardID string, ownedCards map[string]bool) error {
	for {
		m.mu.Lock()
		pending := false
		for cardID, entries := range m.pendingCardDetails {
			for _, reference := range entries {
				if reference.DashboardID == dashboardID || ownedCards[cardID] {
					pending = true
					break
				}
			}
			if pending {
				break
			}
		}
		if !pending {
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
