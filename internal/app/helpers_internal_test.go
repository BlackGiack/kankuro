package app

import (
	"testing"

	"kankuro/internal/session"
)

// newTestManager is session.NewManager plus a Cleanup that shuts it down. It
// is a copy of the one in main_test.go (package app_test), which this package
// cannot see; see there for why the cleanup order is safe.
func newTestManager(t testing.TB, configDir string, repo session.WorkspaceReader) *session.Manager {
	t.Helper()
	sm := session.NewManager(configDir, repo)
	t.Cleanup(func() { sm.Shutdown() })
	return sm
}
