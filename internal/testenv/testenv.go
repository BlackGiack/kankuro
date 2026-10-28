// Package testenv isolates a test binary from the developer's temp
// directories and fails it if a test leaves anything behind. It is imported
// only from _test files (the TestMain of internal/session and internal/app).
package testenv

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// settle is how long Run waits after the tests before looking for leftovers.
// It must exceed session's watchInterval (2s, unexported), the period of the
// background watcher: a watcher that outlived its test only writes on its next
// tick, so a shorter wait would let a forgotten Shutdown go unnoticed. It
// costs this long once per package; that is the price of catching a missing
// newTestManager deterministically instead of by luck.
const settle = 3 * time.Second

// Run runs the tests of one package inside a private root directory and
// returns the exit code for [os.Exit].
//
// TMPDIR points into the root, so every t.TempDir() of the run lives there
// (os.TempDir reads $TMPDIR on each call on Unix), and KANKURO_TMP_DIR points
// at a sibling directory, so temporary workspace copies never land in the
// developer's real tmp root. After the tests and the settle wait, anything
// still under the TempDir area is a leak - typically a goroutine writing into
// a directory the testing package had already removed - and fails the
// package. The root is always removed, so even a leak never reaches the real
// /tmp.
func Run(m *testing.M, name string) int {
	root, err := os.MkdirTemp("", "kankuro-"+name+"-test-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(root)

	tempDir := filepath.Join(root, "t")
	wsDir := filepath.Join(root, "ws")
	homeDir := filepath.Join(root, "home")
	for _, dir := range []string{tempDir, wsDir, homeDir} {
		if err = os.Mkdir(dir, 0o700); err != nil {
			panic(err)
		}
	}
	if err = os.Setenv("HOME", homeDir); err != nil {
		panic(err)
	}
	if err = os.Setenv("TMPDIR", tempDir); err != nil {
		panic(err)
	}
	if err = os.Setenv("KANKURO_TMP_DIR", wsDir); err != nil {
		panic(err)
	}

	code := m.Run()

	time.Sleep(settle)
	leaked := leftovers(tempDir)
	if len(leaked) > 0 {
		fmt.Fprintf(os.Stderr, "FAIL: tests left %d path(s) behind in their temp directory "+
			"(a goroutine outlived its test - is a manager missing its Shutdown cleanup?):\n", len(leaked))
		for _, p := range leaked {
			fmt.Fprintf(os.Stderr, "  %s\n", p)
		}
		if code == 0 {
			code = 1
		}
	}
	return code
}

// leftovers lists every path under dir, relative to it.
func leftovers(dir string) []string {
	var out []string
	_ = filepath.WalkDir(dir, func(path string, _ fs.DirEntry, err error) error {
		if err != nil || path == dir {
			return nil //nolint:nilerr // best effort: an unreadable entry is reported by its parent
		}
		rel, relErr := filepath.Rel(dir, path)
		if relErr != nil {
			rel = path
		}
		out = append(out, rel)
		return nil
	})
	return out
}
