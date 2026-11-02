package session

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/BlackGiack/kankuro/internal/driver"
)

// driverInputWriter routes attach input through the driver lock shared with Kill.
type driverInputWriter struct {
	driver driver.AgentDriver
}

func (w driverInputWriter) Write(p []byte) (int, error) {
	return w.driver.SendInputBytes(p)
}

// AttachChecked validates scope under the manager lock before subscribing.
// The callback must not call the manager.
func (m *Manager) AttachChecked(id string, check func([]Snapshot) error) (*AttachHandle, error) {
	if check == nil {
		return nil, errors.New("attach scope check is required")
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return nil, ErrManagerClosed
	}
	if err := check(m.listLocked()); err != nil {
		return nil, err
	}
	s, ok := m.sessions[id]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", id)
	}
	if s.Status == StatusTerminated {
		return nil, fmt.Errorf("session terminated: %s", id)
	}
	ch := make(chan []byte, outputChanSize)
	s.Subscribers[ch] = struct{}{}
	header := s.PendingAttachHeader
	s.PendingAttachHeader = nil
	return &AttachHandle{
		Input:      driverInputWriter{driver: s.Driver},
		PTY:        s.PTY,
		Output:     ch,
		Scrollback: bytes.Clone(s.Scrollback),
		Header:     header,
		Done:       s.Done,
		Close: func() {
			m.removeSubscriber(id, ch)
		},
	}, nil
}

// Detach validates that a session exists while leaving it running in background.
func (m *Manager) Detach(id string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if _, ok := m.sessions[id]; !ok {
		return fmt.Errorf("session not found: %s", id)
	}
	return nil
}

func (m *Manager) removeSubscriber(sessionID string, ch chan []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[sessionID]
	if !ok {
		return
	}
	if _, okSub := s.Subscribers[ch]; okSub {
		delete(s.Subscribers, ch)
		close(ch)
	}
}

func (m *Manager) closeSubscribersLocked(sessionID string) {
	s, ok := m.sessions[sessionID]
	if !ok {
		return
	}
	for ch := range s.Subscribers {
		close(ch)
		delete(s.Subscribers, ch)
	}
}
