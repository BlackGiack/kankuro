package session

// SetStatusForTest forces a session's status without driving a watcher tick
// or a real process exit, so tests can build deterministic interleavings.
func (m *Manager) SetStatusForTest(id string, st Status) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s, ok := m.sessions[id]; ok {
		s.Status = st
	}
}
