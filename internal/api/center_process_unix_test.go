//go:build unix

//nolint:testpackage // Exercise the process-group lifecycle directly.
package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestCenterWorkspaceCancellationKillsGitHelpers(t *testing.T) {
	if _, err := os.Stat("/bin/sleep"); err != nil {
		t.Skip("test requires /bin/sleep")
	}
	unrelated := exec.Command("/bin/sleep", "60")
	if err := unrelated.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = unrelated.Process.Kill(); _ = unrelated.Wait() })
	dir := t.TempDir()
	pidFile := filepath.Join(t.TempDir(), "helper.pid")
	program := filepath.Join(dir, "git")
	script := fmt.Sprintf("#!/bin/sh\n/bin/sleep 60 &\necho $! > '%s'\nwait\n", pidFile)
	if err := os.WriteFile(program, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	done := make(chan error, 1)
	go func() { _, err := runGit(ctx, dir, "status", "--porcelain=v2", "-z"); done <- err }()
	deadline := time.Now().Add(5 * time.Second)
	pid := 0
	for time.Now().Before(deadline) {
		data, err := os.ReadFile(pidFile)
		if err == nil {
			pid, _ = strconv.Atoi(strings.TrimSpace(string(data)))
			if pid > 0 {
				break
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	if pid <= 0 {
		cancel()
		<-done
		t.Fatal("test child process did not start")
	}
	started := time.Now()
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("cancel result: %v", err)
		}
	case <-time.After(2 * time.Second):
		_ = syscall.Kill(pid, syscall.SIGKILL)
		t.Fatal("Git did not exit promptly")
	}
	if err := syscall.Kill(unrelated.Process.Pid, 0); err != nil {
		t.Errorf("unrelated process was affected: %v", err)
	}
	if time.Since(started) > 1500*time.Millisecond {
		t.Errorf("cancellation was slow: %v", time.Since(started))
	}
	waitForHelperExit(t, pid)
}

func TestCenterWorkspaceOutputLimitKillsGitHelpers(t *testing.T) {
	if _, err := os.Stat("/bin/sleep"); err != nil {
		t.Skip("test requires /bin/sleep")
	}
	bin := t.TempDir()
	pidFile := filepath.Join(t.TempDir(), "helper.pid")
	script := fmt.Sprintf(
		"#!/bin/sh\n/bin/sleep 60 &\necho $! > '%s'\nwhile :; do printf '%s'; done\n",
		pidFile,
		strings.Repeat("0123456789", 800),
	)
	if err := os.WriteFile(filepath.Join(bin, "git"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin)
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	_, err := runGit(ctx, t.TempDir(), "status")
	if err == nil || !strings.Contains(err.Error(), "output exceeded") {
		t.Fatalf("output limit: %v", err)
	}
	data, err := os.ReadFile(pidFile)
	if err != nil {
		t.Fatal(err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || pid <= 0 {
		t.Fatalf("invalid helper pid %q", data)
	}
	waitForHelperExit(t, pid)
}

func waitForHelperExit(t *testing.T, pid int) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if errors.Is(syscall.Kill(pid, 0), syscall.ESRCH) {
			return
		}
		// A killed orphan can remain a zombie until its new parent reaps it.
		if stat, readErr := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid)); readErr == nil {
			end := bytes.LastIndexByte(stat, ')')
			if end >= 0 && bytes.HasPrefix(stat[end+1:], []byte(" Z ")) {
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	_ = syscall.Kill(pid, syscall.SIGKILL)
	t.Fatal("Git helper survived termination")
}
