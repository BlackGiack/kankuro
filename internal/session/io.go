package session

import (
	"bytes"
	"time"

	"kankuro/internal/driver"
)

func (m *Manager) trackOutput(sessionID string, gen uint64, output <-chan []byte) {
	for chunk := range output {
		m.mu.Lock()
		s, ok := m.sessions[sessionID]
		if !ok || s.generation != gen {
			m.mu.Unlock()
			return
		}
		now := time.Now().UTC()
		s.Scrollback = appendScrollback(s.Scrollback, chunk)
		broadcast(chunk, s.Subscribers)
		updateSessionOutputState(s, chunk, now)
		if s.Status != StatusTerminated {
			if s.Pinged {
				s.Status = StatusWaiting
			} else {
				s.Status = StatusRunning
			}
		}
		m.mu.Unlock()
	}
	m.mu.Lock()
	if s, ok := m.sessions[sessionID]; ok && s.generation == gen {
		m.closeSubscribersLocked(sessionID)
	}
	m.mu.Unlock()
}

func appendScrollback(existing, chunk []byte) []byte {
	if len(chunk) >= ScrollbackLimit {
		return bytes.Clone(chunk[len(chunk)-ScrollbackLimit:])
	}
	if len(existing)+len(chunk) <= ScrollbackLimit {
		return append(existing, chunk...)
	}
	trim := len(existing) + len(chunk) - ScrollbackLimit
	next := make([]byte, len(existing)-trim, ScrollbackLimit)
	copy(next, existing[trim:])
	return append(next, chunk...)
}

func broadcast(chunk []byte, subscribers map[chan []byte]struct{}) {
	if len(subscribers) == 0 {
		return
	}
	for ch := range subscribers {
		data := bytes.Clone(chunk)
		select {
		case ch <- data:
		default:
		}
	}
}

func updateSessionOutputState(s *Session, chunk []byte, now time.Time) {
	s.PendingLine = append(s.PendingLine, chunk...)
	if len(s.PendingLine) > pendingLineLimit {
		s.PendingLine = append([]byte(nil), s.PendingLine[len(s.PendingLine)-pendingLineLimit:]...)
	}

	if wordRE.Match(chunk) {
		s.LastWordAt = now
	}

	for {
		idx := bytes.IndexByte(s.PendingLine, '\n')
		if idx == -1 {
			break
		}
		line := sanitizeLine(s.PendingLine[:idx])
		s.PendingLine = append([]byte(nil), s.PendingLine[idx+1:]...)
		if line == "" {
			continue
		}
		s.LastOutput = line
		s.Pinged = driver.IsPromptLike(line, s.LastWordAt, now)
	}

	sanitized := sanitizeLine(s.PendingLine)
	if sanitized != "" {
		s.LastOutput = sanitized
		s.Pinged = driver.IsPromptLike(sanitized, s.LastWordAt, now)
	}
}

func sanitizeLine(line []byte) string {
	clean := ansiRE.ReplaceAll(bytes.Clone(line), nil)
	clean = bytes.ReplaceAll(clean, []byte("\r"), nil)
	return string(bytes.TrimSpace(clean))
}
