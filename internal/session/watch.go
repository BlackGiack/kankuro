package session

import (
	"time"

	"kankuro/internal/driver"
)

func (m *Manager) watch(sessionID string, d driver.AgentDriver) {
	ticker := time.NewTicker(watchInterval)
	defer ticker.Stop()
	for {
		<-ticker.C
		status := d.GetStatus()
		m.mu.Lock()
		s, ok := m.sessions[sessionID]
		if !ok {
			m.mu.Unlock()
			return
		}
		if status == driver.StatusTerminated {
			s.Status = StatusTerminated
			s.Pinged = false
			m.persistSessionLocked(s)
			m.closeSubscribersLocked(sessionID)
			m.mu.Unlock()
			return
		}
		if driver.IsPromptLike(s.LastOutput, s.LastWordAt, time.Now().UTC()) {
			s.Status = StatusWaiting
			s.Pinged = true
		} else {
			s.Status = StatusRunning
			s.Pinged = false
		}
		m.mu.Unlock()
	}
}
