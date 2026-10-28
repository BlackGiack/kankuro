package app_test

import (
	"os"
	"testing"

	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/BlackGiack/kankuro/internal/testenv"
)

// TestMain runs the package inside testenv's private temp root: TMPDIR and
// KANKURO_TMP_DIR point into it, so nothing reaches the developer's real
// directories, and the package fails if a test leaves anything behind - which
// is what a manager without a Shutdown cleanup does (see newTestManager). A
// test may still override KANKURO_TMP_DIR with t.Setenv.
func TestMain(m *testing.M) {
	os.Exit(testenv.Run(m, "app"))
}

// newTestManager is session.NewManager plus a Cleanup that shuts it down.
// t.Cleanup runs last-in-first-out, so as long as the config dir's
// t.TempDir() was created *before* this call - always the case when the
// manager is built from it - Shutdown runs before that dir is removed, and no
// watcher can write into it afterwards.
func newTestManager(t testing.TB, configDir string, repo session.WorkspaceReader) *session.Manager {
	t.Helper()
	sm := session.NewManager(configDir, repo, session.WithStartupDelay(0))
	t.Cleanup(func() { sm.Shutdown() })
	return sm
}
