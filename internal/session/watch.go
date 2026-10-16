package session

import (
	"time"

	"github.com/BlackGiack/kankuro/internal/driver"
)

// startWatchLocked runs watch for one activation, tracked by m.watchers. It
// must be called with m.mu held, after checking !m.closed: the WaitGroup's Add
// happens under m.mu after that check, and Shutdown sets closed under m.mu
// before it calls Wait, so no Add can race a Wait that started at zero.
func (m *Manager) startWatchLocked(id string, gen uint64, d driver.AgentDriver) {
	done := m.sessions[id].Done
	m.watchers.Go(func() { m.watch(id, gen, d, done) })
}

// watch polls one activation's driver until it terminates, its generation goes
// stale, or Shutdown closes m.stop. It is the only goroutine that writes
// session records in the background, which is why it must be stoppable.
//
//nolint:gocognit // One poll loop owns status, registration timeout, persistence, and subscriber teardown.
func (m *Manager) watch(sessionID string, gen uint64, d driver.AgentDriver, done <-chan struct{}) {
	ticker := time.NewTicker(watchInterval)
	defer ticker.Stop()
	for {
		select {
		case <-m.stop:
			return
		case <-done:
		case <-ticker.C:
		}
		status := d.GetStatus()
		select {
		case <-done:
			status = driver.StatusTerminated
		default:
		}
		m.mu.Lock()
		s, ok := m.sessions[sessionID]
		if !ok || s.generation != gen {
			m.mu.Unlock()
			return
		}
		if s.deleting || s.stopping {
			m.mu.Unlock()
			select {
			case <-done:
				return
			default:
			}
			continue
		}
		if s.Status == StatusTerminated {
			m.mu.Unlock()
			return
		}
		if status == driver.StatusTerminated {
			m.cancelStartupLocked(s)
			s.Status = StatusTerminated
			m.cancelPendingMessageAdmissionsLocked(s)
			s.Pinged = false
			if s.Registration == RegistrationPending {
				s.Registration = RegistrationUnconfirmed
			}
			m.revokeSessionActivationLocked(s)
			if exitCoder, hasExitCode := d.(interface{ ExitCode() int }); hasExitCode {
				s.ExitCode = exitCoder.ExitCode()
			}
			m.persistSessionLocked(s)
			m.mu.Unlock()
			if m.messages != nil {
				m.messages.Wake()
			}
			return
		}
		if s.Registration == RegistrationPending && time.Since(s.RegistrationAt) >= 60*time.Second {
			s.Registration = RegistrationUnconfirmed
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
