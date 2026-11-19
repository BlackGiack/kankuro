package session

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/BlackGiack/kankuro/internal/driver"
	"github.com/creack/pty"
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
	return m.attachChecked(id, "", check)
}

// AttachExpectedProcessChecked rejects a replaced process before subscribing or consuming the header.
func (m *Manager) AttachExpectedProcessChecked(id, processID string, check func([]Snapshot) error) (*AttachHandle, error) {
	if processID == "" {
		return nil, errors.New("expected process identity is required")
	}
	return m.attachChecked(id, processID, check)
}

func (m *Manager) attachChecked(id, processID string, check func([]Snapshot) error) (*AttachHandle, error) {
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
	if processID != "" && s.ProcessActivationID != processID {
		return nil, errors.New("session process changed before attachment")
	}
	if s.Status == StatusTerminated {
		return nil, fmt.Errorf("session terminated: %s", id)
	}
	ch := make(chan []byte, outputChanSize)
	if s.outputComplete {
		close(ch)
	} else {
		s.Subscribers[ch] = struct{}{}
	}
	subscribers := s.Subscribers
	header := s.PendingAttachHeader
	s.PendingAttachHeader = nil
	generation := s.generation
	d := s.Driver
	terminal := s.PTY
	return &AttachHandle{
		Input:      driverInputWriter{driver: s.Driver},
		PTY:        s.PTY,
		Output:     ch,
		Scrollback: bytes.Clone(s.Scrollback),
		Header:     header,
		Done:       s.Done,
		Close: func() {
			m.removeSubscriber(subscribers, ch)
		},
		WriteChecked: func(data []byte) (int, error) {
			m.mu.RLock()
			current, exists := m.sessions[id]
			active := exists && current.generation == generation && current.Status != StatusTerminated
			m.mu.RUnlock()
			if !active {
				return 0, errors.New("attachment process is no longer active")
			}
			return d.SendInputBytes(data)
		},
		ResizeChecked: func(rows, columns int) error {
			if rows <= 0 || columns <= 0 || rows > 65535 || columns > 65535 {
				return errors.New("invalid terminal size")
			}
			m.mu.RLock()
			defer m.mu.RUnlock()
			current, exists := m.sessions[id]
			if !exists || current.generation != generation || current.Status == StatusTerminated || terminal == nil {
				return errors.New("attachment process is no longer active")
			}
			return pty.Setsize(terminal, &pty.Winsize{Rows: uint16(rows), Cols: uint16(columns)})
		},
	}, nil
}

func (m *Manager) removeSubscriber(subscribers map[chan []byte]struct{}, ch chan []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, okSub := subscribers[ch]; okSub {
		delete(subscribers, ch)
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
