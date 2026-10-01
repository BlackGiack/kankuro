package session

import (
	"time"

	"kankuro/internal/driver"
)

// startWatchLocked runs watch for one activation, tracked by m.watchers. It
// must be called with m.mu held, after checking !m.closed: the WaitGroup's Add
// happens under m.mu after that check, and Shutdown sets closed under m.mu
// before it calls Wait, so no Add can race a Wait that started at zero.
func (m *Manager) startWatchLocked(id string, gen uint64, d driver.AgentDriver) {
	m.watchers.Go(func() { m.watch(id, gen, d) })
}

// watch polls one activation's driver until it terminates, its generation goes
// stale, or Shutdown closes m.stop. It is the only goroutine that writes
// session records in the background, which is why it must be stoppable.
func (m *Manager) watch(sessionID string, gen uint64, d driver.AgentDriver) {
	ticker := time.NewTicker(watchInterval)
	defer ticker.Stop()
	for {
		select {
		case <-m.stop:
			return
		case <-ticker.C:
		}
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
