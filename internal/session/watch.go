package session

import (
	"time"

	"kankuro/internal/driver"
)

func (m *Manager) watch(sessionID string, gen uint64, d driver.AgentDriver) {
	ticker := time.NewTicker(watchInterval)
	defer ticker.Stop()
	for {
		<-ticker.C
		status := d.GetStatus()
		m.mu.Lock()
		s, ok := m.sessions[sessionID]
		if !ok || s.generation != gen {
			m.mu.Unlock()
			return
		}
		if status == driver.StatusTerminated {
			s.Status = StatusTerminated
			s.Pinged = false
			if exitCoder, hasExitCode := d.(interface{ ExitCode() int }); hasExitCode {
				s.ExitCode = exitCoder.ExitCode()
			}
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
