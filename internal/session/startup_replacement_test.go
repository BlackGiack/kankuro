//nolint:testpackage // The fixture reads the activation credential and message store under their locks.
package session

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/google/uuid"
)

type replacementFixture struct {
	manager     *Manager
	clock       *controlledStartupClock
	senderToken string
	id          string
	original    config.IDCard
	replacement config.IDCard
}

func newReplacementFixture(t *testing.T) replacementFixture {
	t.Helper()
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	workspace := t.TempDir()
	if err = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: workspace}}); err != nil {
		t.Fatal(err)
	}
	workspaces, err := repo.ListWorkspaces()
	if err != nil || len(workspaces) != 1 {
		t.Fatalf("workspaces: %v, %v", workspaces, err)
	}
	manager := NewManager(repo.Dir(), repo, WithStartupDelay(0))
	manager.SetMCPEndpoint("http://127.0.0.1:12345/mcp")
	t.Cleanup(func() { manager.Shutdown() })
	sender := config.IDCard{ID: "sender", Name: "sender", Driver: "cat", WorkspaceID: workspaces[0].ID, DashboardControl: true}
	senderID, err := manager.Launch(sender, nil, nil, config.BuiltinMCPs())
	if err != nil {
		t.Fatal(err)
	}
	manager.mu.RLock()
	senderToken := manager.sessions[senderID].ActivationToken
	manager.mu.RUnlock()
	clock := newControlledStartupClock()
	manager.mu.Lock()
	manager.startupClock, manager.startupDelay = clock, 10*time.Second
	manager.mu.Unlock()
	return replacementFixture{manager: manager, clock: clock, senderToken: senderToken, id: uuid.NewString(),
		original:    config.IDCard{ID: "original", Name: "original card", Driver: "cat", WorkspaceID: workspaces[0].ID},
		replacement: config.IDCard{ID: "other-card", Name: "other card", Driver: "cat", WorkspaceID: workspaces[0].ID}}
}

func nextReplacementTimer(t *testing.T, clock *controlledStartupClock) *controlledStartupTimer {
	t.Helper()
	select {
	case timer := <-clock.created:
		if delay := <-clock.durations; delay != 10*time.Second {
			t.Fatalf("startup timer delay = %v", delay)
		}
		return timer
	case <-time.After(5 * time.Second):
		t.Fatal("startup timer not created")
		return nil
	}
}

func runReplacement(t *testing.T, operation func() error) {
	t.Helper()
	finished := make(chan error, 1)
	go func() { finished <- operation() }()
	select {
	case err := <-finished:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("lifecycle operation did not return while startup timer was held")
	}
}

func requireStartingAdmissionRejected(t *testing.T, f replacementFixture, activation string) {
	t.Helper()
	f.manager.messages.mu.Lock()
	before := len(f.manager.messages.data.Messages)
	f.manager.messages.mu.Unlock()
	admission, err := f.manager.ReserveMessageAdmission(f.senderToken, f.id, activation)
	if err == nil {
		admission.Release()
		t.Fatal("starting activation admitted a prompt")
	}
	var failure *MessagingError
	if !errors.As(err, &failure) || failure.Code != "starting" {
		t.Fatalf("starting rejection = %v", err)
	}
	f.manager.messages.mu.Lock()
	after := len(f.manager.messages.data.Messages)
	f.manager.messages.mu.Unlock()
	if after != before {
		t.Fatalf("starting rejection changed message history: %d -> %d", before, after)
	}
}

func requireAttachmentOutput(t *testing.T, f replacementFixture, marker string) {
	t.Helper()
	handle, err := f.manager.Attach(f.id)
	if err != nil {
		t.Fatalf("attach while starting: %v", err)
	}
	defer handle.Close()
	if _, err = handle.Input.Write([]byte(marker + "\n")); err != nil {
		t.Fatalf("write attached input: %v", err)
	}
	var received []byte
	deadline := time.After(5 * time.Second)
	for !bytes.Contains(received, []byte(marker)) {
		select {
		case chunk, ok := <-handle.Output:
			if !ok {
				t.Fatal("attachment closed before startup output")
			}
			received = append(received, chunk...)
		case <-deadline:
			t.Fatalf("attachment never received %q; output %q", marker, received)
		}
	}
	if !bytes.Contains(f.manager.GetScrollback(f.id), []byte(marker)) {
		t.Fatalf("scrollback missing %q", marker)
	}
}

func assertReplacement(
	t *testing.T, f replacementFixture, old Snapshot, oldTimer, currentTimer *controlledStartupTimer,
	cardID, marker string,
) Snapshot {
	t.Helper()
	current, ok := f.manager.GetSession(f.id)
	if !ok || current.ActivationID == "" || current.ActivationID == old.ActivationID || current.CardID != cardID ||
		current.Name != "custom session" || current.Phase != "starting" || !current.ProcessStarted || current.MessageDeliveryAvailable {
		t.Fatalf("replacement snapshot = %+v; previous = %+v", current, old)
	}
	requireStartingAdmissionRejected(t, f, current.ActivationID)
	requireAttachmentOutput(t, f, marker)
	oldTimer.ready <- f.clock.Now()
	current, _ = f.manager.GetSession(f.id)
	if current.Phase != "starting" || current.MessageDeliveryAvailable {
		t.Fatalf("stale timer changed replacement: %+v", current)
	}
	currentTimer.ready <- f.clock.Now().Add(10 * time.Second)
	waitStartupSignal(t, currentTimer.stopped)
	current, _ = f.manager.GetSession(f.id)
	if current.Phase != "active" || !current.ProcessStarted || !current.MessageDeliveryAvailable {
		t.Fatalf("timer did not enable replacement: %+v", current)
	}
	admission, err := f.manager.ReserveMessageAdmission(f.senderToken, f.id, current.ActivationID)
	if err != nil {
		t.Fatalf("active replacement admission: %v", err)
	}
	admission.Release()
	if stale, staleErr := f.manager.ReserveMessageAdmission(f.senderToken, f.id, old.ActivationID); staleErr == nil {
		stale.Release()
		t.Fatal("old activation accepted after replacement")
	}
	return current
}

func TestRelaunchAndDistinctCardSwitchGateMessageDelivery(t *testing.T) {
	f := newReplacementFixture(t)
	name := "custom session"
	runReplacement(t, func() error {
		_, err := f.manager.LaunchWithIDContext(context.Background(), f.id, f.original, nil, nil, config.BuiltinMCPs(), &name)
		return err
	})
	firstTimer := nextReplacementTimer(t, f.clock)
	first, _ := f.manager.GetSession(f.id)
	requireStartingAdmissionRejected(t, f, first.ActivationID)
	f.clock.advance(9 * time.Second)
	if err := f.manager.Kill(f.id); err != nil {
		t.Fatal(err)
	}
	waitStartupSignal(t, firstTimer.stopped)
	runReplacement(t, func() error { return f.manager.Relaunch(f.id, f.original) })
	secondTimer := nextReplacementTimer(t, f.clock)
	assertReplacement(t, f, first, firstTimer, secondTimer, f.original.ID, "relaunch output marker")

	if err := f.manager.Kill(f.id); err != nil {
		t.Fatal(err)
	}
	runReplacement(t, func() error { return f.manager.Relaunch(f.id, f.original) })
	thirdTimer := nextReplacementTimer(t, f.clock)
	third, _ := f.manager.GetSession(f.id)
	requireStartingAdmissionRejected(t, f, third.ActivationID)
	f.clock.advance(9 * time.Second)
	if err := f.manager.Kill(f.id); err != nil {
		t.Fatal(err)
	}
	waitStartupSignal(t, thirdTimer.stopped)
	runReplacement(t, func() error { return f.manager.SwitchAgent(f.id, f.replacement, nil, nil, config.BuiltinMCPs()) })
	fourthTimer := nextReplacementTimer(t, f.clock)
	fourth := assertReplacement(t, f, third, thirdTimer, fourthTimer, f.replacement.ID, "switch output marker")
	if fourth.ActivationID == first.ActivationID {
		t.Fatal(fmt.Errorf("switch reused original activation %s", first.ActivationID))
	}
}
