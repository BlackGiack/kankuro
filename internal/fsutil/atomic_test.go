package fsutil_test

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/BlackGiack/kankuro/internal/fsutil"
)

func TestWriteFileAtomicLeavesNoResidue(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "probe.json")

	if err := fsutil.WriteFileAtomic(path, []byte("hello"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	if len(entries) != 1 || entries[0].Name() != "probe.json" {
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Fatalf("expected only probe.json in the directory, found %v", names)
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != "hello" {
		t.Fatalf("content = %q, err = %v", got, err)
	}
}

func TestWriteFileAtomicSetsExactMode(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "probe.json")

	// os.CreateTemp makes 0600 files, and an existing target has its own mode;
	// the result must be exactly the requested one (0640 is neither), so the
	// explicit Chmod is what this test pins.
	if err := os.WriteFile(path, []byte("old"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatalf("chmod seed: %v", err)
	}

	if err := fsutil.WriteFileAtomic(path, []byte("new"), 0o640); err != nil {
		t.Fatalf("write: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode().Perm() != 0o640 {
		t.Fatalf("expected mode 0640, got %o", info.Mode().Perm())
	}
}

func TestWriteFileAtomicErrorCleansUp(t *testing.T) {
	dir := t.TempDir()
	// A directory at the target path makes the final rename fail.
	target := filepath.Join(dir, "probe.json")
	if err := os.Mkdir(target, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	if err := fsutil.WriteFileAtomic(target, []byte("x"), 0o600); err == nil {
		t.Fatal("expected an error when the target is a directory")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tmp") {
			t.Fatalf("temp file %q left behind after a failed write", e.Name())
		}
	}
}

// pollUntilStopped reads path in a loop until stop is closed and reports the
// first read that is not one complete payload ("" if every read was clean).
func pollUntilStopped(path string, size int, stop <-chan struct{}) string {
	for {
		select {
		case <-stop:
			return ""
		default:
		}
		got, err := os.ReadFile(path)
		if err != nil {
			return "read during writes: " + err.Error()
		}
		if len(got) != size || len(bytes.Trim(got, string(got[:1]))) != 0 {
			return "reader observed a partial or spliced file"
		}
	}
}

func TestWriteFileAtomicConcurrentWritesNeverSplice(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "probe.json")

	const n = 20
	payloads := make([][]byte, n)
	for i := range n {
		payloads[i] = bytes.Repeat([]byte{byte('a' + i)}, 64*1024)
	}

	if err := fsutil.WriteFileAtomic(path, payloads[0], 0o600); err != nil {
		t.Fatalf("seed write: %v", err)
	}

	// A reader polling during the writes: with a non-atomic write it would
	// observe the truncated/partial file; with rename it only ever sees a
	// complete payload.
	stop := make(chan struct{})
	readerDone := make(chan string, 1)
	go func() { readerDone <- pollUntilStopped(path, len(payloads[0]), stop) }()

	var wg sync.WaitGroup
	errs := make(chan error, n)
	for i := range n {
		wg.Go(func() {
			if err := fsutil.WriteFileAtomic(path, payloads[i], 0o600); err != nil {
				errs <- err
			}
		})
	}
	wg.Wait()
	close(stop)
	if msg := <-readerDone; msg != "" {
		t.Fatal(msg)
	}
	close(errs)
	for err := range errs {
		t.Fatalf("concurrent write failed: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read final file: %v", err)
	}
	if !slices.ContainsFunc(payloads, func(p []byte) bool { return bytes.Equal(got, p) }) {
		t.Fatalf("final content (len %d) matches none of the payloads: spliced write", len(got))
	}

	matches, err := filepath.Glob(filepath.Join(dir, "*.tmp"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(matches) != 0 {
		t.Fatalf("leftover temp files: %v", matches)
	}
}
