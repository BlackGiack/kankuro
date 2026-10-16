package session

import (
	"fmt"
	"strings"
	"sync"
)

// ReserveCardReference validates card existence and records an in-flight
// activation under the same lock used by DeleteCardIfUnreferenced. The caller
// must release the reservation after activation publishes or rolls back.
func (m *Manager) ReserveCardReference(cardID string, exists func() (bool, error)) (func(), error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return nil, ErrManagerClosed
	}

	cardExists, err := exists()
	if err != nil {
		return nil, fmt.Errorf("check card before activation: %w", err)
	}
	if !cardExists {
		return nil, fmt.Errorf("card not found: %s", cardID)
	}
	m.pendingCardRefs[cardID]++

	var once sync.Once
	return func() {
		once.Do(func() {
			m.mu.Lock()
			defer m.mu.Unlock()
			if m.pendingCardRefs[cardID] <= 1 {
				delete(m.pendingCardRefs, cardID)
				return
			}
			m.pendingCardRefs[cardID]--
		})
	}, nil
}

// DeleteCardIfUnreferenced serializes the durable-reference check and card
// deletion against activation admission. It includes terminated sessions and
// their saved conversation bindings.
func (m *Manager) DeleteCardIfUnreferenced(cardID string, deleteCard func() error) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if pending := m.pendingCardRefs[cardID]; pending > 0 {
		return fmt.Errorf("cannot delete card: referenced by %d activation(s) in progress", pending)
	}
	var referencedBy []string
	for _, current := range m.sessions {
		_, hasBinding := current.Conversations[cardID]
		if current.CardID == cardID || hasBinding {
			referencedBy = append(referencedBy, current.Name)
		}
	}
	if len(referencedBy) > 0 {
		return fmt.Errorf("cannot delete card: referenced by session(s) %s", strings.Join(referencedBy, ", "))
	}
	return deleteCard()
}
