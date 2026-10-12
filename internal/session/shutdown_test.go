package session_test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/BlackGiack/kankuro/internal/session"
)

// persistedStatus reads a session's status straight from its JSON record.
func persistedStatus(t *testing.T, configDir, sid string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(configDir, "sessions", sid+".json"))
	if err != nil {
		t.Fatalf("read record: %v", err)
	}
	var rec struct {
		Status string `json:"status"`
	}
	if err = json.Unmarshal(raw, &rec); err != nil {
		t.Fatalf("parse record: %v", err)
	}
	return rec.Status
}

// TestShutdownStopsWatchersBeforeTheyPersist pins two things about the stop
// channel. (1) Shutdown must not wait for a watcher's next tick: the WaitGroup
// alone would still make it return only after the watcher persisted, so the
// duration is what proves stop works, and what keeps quitting the app quick.
// (2) Nothing writes after Shutdown returns. That is the one place a sleep is
// the point: a leftover watcher only acts on its next tick (watchInterval, 2s,
// unexported), so the test waits past it and checks that the sessions
// directory it deleted does not come back.
func TestShutdownStopsWatchersBeforeTheyPersist(t *testing.T) {
	h := newGateHarness(t, false, nil)
	if _, err := h.sm.Launch(h.card, nil, nil, nil); err != nil {
		t.Fatalf("launch: %v", err)
	}

	// Launched right before: the watcher's first tick is about 2s away.
	start := time.Now()
	h.sm.Shutdown()
	if d := time.Since(start); d > time.Second {
		t.Fatalf("Shutdown took %v: watchers are not stopped, it waited for their tick", d)
	}

	sessionsDir := filepath.Join(h.repoDir, "sessions")
	if err := os.RemoveAll(sessionsDir); err != nil {
		t.Fatal(err)
	}
	time.Sleep(2*time.Second + 500*time.Millisecond) // watchInterval + margin

	if _, err := os.Stat(sessionsDir); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("a watcher outlived Shutdown and wrote to disk (stat err: %v)", err)
	}
}

// TestActivationAfterShutdownIsRefused: once Shutdown has run, nothing may
// start a process the manager could no longer stop or watch.
func TestActivationAfterShutdownIsRefused(t *testing.T) {
	h := newGateHarness(t, false, nil)
	sid := h.launchTerminated(t)
	spawnsBefore := h.spawns.Load()

	h.sm.Shutdown()

	if _, err := h.sm.Launch(h.card, nil, nil, nil); !errors.Is(err, session.ErrManagerClosed) {
		t.Fatalf("Launch: want ErrManagerClosed, got %v", err)
	}
	if err := h.sm.Relaunch(sid, h.card, nil, nil, nil); !errors.Is(err, session.ErrManagerClosed) {
		t.Fatalf("Relaunch: want ErrManagerClosed, got %v", err)
	}
	if err := h.sm.SwitchAgent(sid, h.card, nil, nil, nil); !errors.Is(err, session.ErrManagerClosed) {
		t.Fatalf("SwitchAgent: want ErrManagerClosed, got %v", err)
	}
	if got := h.spawns.Load(); got != spawnsBefore {
		t.Fatalf("Spawn was called %d time(s) after Shutdown", got-spawnsBefore)
	}
}

// onceCloser returns a func that closes ch at most once. Registered as a
// cleanup, it opens a gate that a failed assertion left shut, so the cleanup
// Shutdown of the test's manager can finish instead of waiting forever.
func onceCloser(t *testing.T, ch chan struct{}) func() {
	t.Helper()
	var once sync.Once
	closeOnce := func() { once.Do(func() { close(ch) }) }
	t.Cleanup(closeOnce)
	return closeOnce
}

// shutdownAsync runs Shutdown in a goroutine and returns a channel that
// receives once it has returned.
func shutdownAsync(sm *session.Manager) <-chan struct{} {
	returned := make(chan struct{})
	go func() {
		sm.Shutdown()
		close(returned)
	}()
	return returned
}

// assertStillBlocked fails if ch is closed within a short window. The window
// is the point of these tests: "has not returned yet" can only be observed by
// waiting, and a Shutdown that wrongly returns does so at once.
func assertStillBlocked(t *testing.T, ch <-chan struct{}, why string) {
	t.Helper()
	select {
	case <-ch:
		t.Fatalf("Shutdown returned early: %s", why)
	case <-time.After(100 * time.Millisecond):
	}
}

// TestShutdownDuringSpawnKillsTheNewProcess: a Relaunch whose Spawn is still
// in flight when Shutdown runs must not leave the process it spawns behind.
// Shutdown has to wait for the activation to reach its final check and kill
// that process: main returns right after Shutdown, so anything left for the
// activation goroutine to do "later" would never happen.
func TestShutdownDuringSpawnKillsTheNewProcess(t *testing.T) {
	entered, gate := make(chan struct{}, 1), make(chan struct{})
	var second *gateDriver
	h := newGateHarness(t, false, func(n int, d *gateDriver) {
		if n == 1 {
			d.spawnEntered, d.spawnGate = entered, gate
			second = d
		}
	})
	openGate := onceCloser(t, gate) // after the harness: cleanups run last-in-first-out
	sid := h.launchTerminated(t)

	result := make(chan error, 1)
	go func() { result <- h.sm.Relaunch(sid, h.card, nil, nil, nil) }()
	waitSignal(t, entered, "Relaunch to enter Spawn")

	returned := shutdownAsync(h.sm)
	assertStillBlocked(t, returned, "a Relaunch is still spawning")

	openGate()
	waitSignal(t, returned, "Shutdown to return once the Relaunch finished")
	if err := waitErr(t, result, "Relaunch"); !errors.Is(err, session.ErrManagerClosed) {
		t.Fatalf("Relaunch: want ErrManagerClosed, got %v", err)
	}
	if !second.terminated.Load() {
		t.Fatal("the process spawned during Shutdown was left running")
	}
	snap, _ := h.sm.GetSession(sid)
	if snap.Status != session.StatusTerminated {
		t.Fatalf("session must stay Terminated, got %s", snap.Status)
	}
}

// TestShutdownDuringTmpLaunchRemovesTheCopy: same for a new temporary-workspace
// Launch, whose workspace copy would otherwise stay behind with no record
// pointing at it.
func TestShutdownDuringTmpLaunchRemovesTheCopy(t *testing.T) {
	entered, gate := make(chan struct{}, 1), make(chan struct{})
	var first *gateDriver
	h := newGateHarness(t, true, func(n int, d *gateDriver) {
		if n == 0 {
			d.spawnEntered, d.spawnGate = entered, gate
			first = d
		}
	})
	openGate := onceCloser(t, gate)         // after the harness: cleanups run last-in-first-out
	tmpRoot := os.Getenv("KANKURO_TMP_DIR") // set by newGateHarness for useTmp

	result := make(chan error, 1)
	go func() {
		_, err := h.sm.Launch(h.card, nil, nil, nil)
		result <- err
	}()
	waitSignal(t, entered, "Launch to enter Spawn")
	if entries, _ := os.ReadDir(tmpRoot); len(entries) != 1 {
		t.Fatalf("expected one workspace copy while the Launch is in flight, got %d", len(entries))
	}

	returned := shutdownAsync(h.sm)
	assertStillBlocked(t, returned, "a Launch is still spawning")

	openGate()
	waitSignal(t, returned, "Shutdown to return once the Launch finished")
	if err := waitErr(t, result, "Launch"); !errors.Is(err, session.ErrManagerClosed) {
		t.Fatalf("Launch: want ErrManagerClosed, got %v", err)
	}
	if !first.terminated.Load() {
		t.Fatal("the process spawned during Shutdown was left running")
	}
	if entries, _ := os.ReadDir(tmpRoot); len(entries) != 0 {
		t.Fatalf("workspace copy left behind after Shutdown: %d entries in %s", len(entries), tmpRoot)
	}
}

func TestShutdownIsIdempotent(t *testing.T) {
	h := newGateHarness(t, false, nil)
	if _, err := h.sm.Launch(h.card, nil, nil, nil); err != nil {
		t.Fatalf("launch: %v", err)
	}

	if got := h.sm.Shutdown(); got != 1 {
		t.Fatalf("first Shutdown: want 1 running session, got %d", got)
	}

	done := make(chan int, 1)
	go func() { done <- h.sm.Shutdown() }()
	select {
	case got := <-done:
		if got != 0 {
			t.Fatalf("second Shutdown: nothing was running, got %d", got)
		}
	case <-time.After(time.Second):
		t.Fatal("second Shutdown did not return promptly")
	}
}

// TestShutdownKillsSessionsInParallel: with a graceful stop per agent, killing
// one after another would take N grace periods. Two sessions whose Kill takes
// 300ms must be gone in about one.
func TestShutdownKillsSessionsInParallel(t *testing.T) {
	const killTime = 300 * time.Millisecond
	h := newGateHarness(t, false, func(_ int, d *gateDriver) { d.killDelay = killTime })

	var sids []string
	for range 2 {
		sid, err := h.sm.Launch(h.card, nil, nil, nil)
		if err != nil {
			t.Fatalf("launch: %v", err)
		}
		sids = append(sids, sid)
	}

	start := time.Now()
	got := h.sm.Shutdown()
	elapsed := time.Since(start)

	if got != 2 {
		t.Fatalf("Shutdown reported %d running sessions, want 2", got)
	}
	if elapsed >= 500*time.Millisecond {
		t.Fatalf("Shutdown took %v; sessions are being killed one after another", elapsed)
	}
	for _, sid := range sids {
		snap, _ := h.sm.GetSession(sid)
		if snap.Status != session.StatusTerminated {
			t.Fatalf("session %s: want Terminated, got %s", sid, snap.Status)
		}
		if st := persistedStatus(t, h.repoDir, sid); st != "terminated" {
			t.Fatalf("session %s: persisted status %q, want terminated", sid, st)
		}
	}
}
