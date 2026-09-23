package session

import (
	"bytes"
	"fmt"

	"kankuro/internal/driver"
)

// driverInputWriter adapts a driver.AgentDriver to [io.Writer] by routing
// writes through its locked SendInputBytes, instead of a raw [os.File] PTY
// handle. Writing directly to the PTY (the previous behavior) raced Kill,
// which closes and nils that same fd under the driver's own lock - a
// keystroke landing during a kill could hit a closed fd and be silently
// dropped. See debt_impl_plan_4.md step 4.
type driverInputWriter struct {
	driver driver.AgentDriver
}

func (w driverInputWriter) Write(p []byte) (int, error) {
	return w.driver.SendInputBytes(p)
}

// Attach returns an attach handle with scrollback replay and live output stream.
func (m *Manager) Attach(id string) (*AttachHandle, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
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
