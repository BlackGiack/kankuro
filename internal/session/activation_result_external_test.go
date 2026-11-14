package session_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/driver"
	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/google/uuid"
)

//nolint:govet // Independent fixture assertions use scoped error variables.
func TestActivationResultCapturesPublishedGenerations(t *testing.T) {
	h := newGateHarness(t, false, nil)
	id := uuid.NewString()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	release, err := h.sm.ReserveCardReferenceForDashboardWithCancel(
		h.card.ID,
		"default",
		id,
		cancel,
		func([]session.Snapshot) (bool, error) { return true, nil },
	)
	if err != nil {
		t.Fatal(err)
	}
	first := h.sm.LaunchInDashboardWithIDResult(ctx, id, "default", h.card, nil, nil, nil, func([]session.Snapshot) error { return nil })
	release()
	if first.Err != nil || first.Effect != session.ActivationPublished || first.SessionID != id || first.ProcessActivationID == "" {
		t.Fatalf("launch result: %+v", first)
	}
	if err := h.sm.Kill(id); err != nil {
		t.Fatal(err)
	}
	second := h.sm.RelaunchWithChoiceCheckedResult(
		ctx,
		id,
		h.card,
		session.ConversationChoice{},
		func([]session.Snapshot) error { return nil },
	)
	if second.Err != nil || second.Effect != session.ActivationPublished || second.ProcessActivationID == "" ||
		first.ProcessActivationID == second.ProcessActivationID {
		t.Fatalf("relaunch: first=%+v second=%+v", first, second)
	}
	if err := h.sm.Kill(id); err != nil {
		t.Fatal(err)
	}
	third := h.sm.SwitchAgentWithChoiceCheckedResult(
		ctx,
		id,
		h.card,
		nil,
		nil,
		nil,
		session.ConversationChoice{},
		func([]session.Snapshot) error { return nil },
	)
	if third.Err != nil || third.Effect != session.ActivationPublished || third.ProcessActivationID == "" ||
		second.ProcessActivationID == third.ProcessActivationID {
		t.Fatalf("switch: second=%+v third=%+v", second, third)
	}
	if first.ProcessActivationID == third.ProcessActivationID {
		t.Fatal("earlier result changed or identity reused")
	}
	check := func([]session.Snapshot) error { return nil }
	before, _ := h.sm.GetSession(id)
	rename, err := h.sm.SetNameExpectedProcessChecked(id, "stale rename", first.ProcessActivationID, check)
	if err == nil || rename.Applied {
		t.Fatalf("stale rename applied: %+v %v", rename, err)
	}
	after, _ := h.sm.GetSession(id)
	if after.Name != before.Name {
		t.Fatalf("stale rename changed name: %q", after.Name)
	}
	if handle, err := h.sm.AttachExpectedProcessChecked(id, first.ProcessActivationID, check); err == nil {
		handle.Close()
		t.Fatal("stale process attachment succeeded")
	}
	handle, err := h.sm.AttachExpectedProcessChecked(id, third.ProcessActivationID, check)
	if err != nil {
		t.Fatalf("current process attachment: %v", err)
	}
	handle.Close()
}

func TestActivationResultPrePublicationRejection(t *testing.T) {
	h := newGateHarness(t, false, nil)
	id := h.launchTerminated(t)
	rejected := errors.New("publication rejected")
	result := h.sm.RelaunchWithChoiceCheckedResult(
		context.Background(),
		id,
		h.card,
		session.ConversationChoice{},
		func([]session.Snapshot) error { return rejected },
	)
	if result.Effect != session.ActivationNotPublished || !result.Spawned || !errors.Is(result.Err, rejected) ||
		result.ProcessActivationID != "" {
		t.Fatalf("result: %+v", result)
	}
}

func TestActivationResultRetainsPublicationOnPersistenceFailure(t *testing.T) {
	h := newGateHarness(t, false, nil)
	id := h.launchTerminated(t)
	sessionsPath := filepath.Join(h.repoDir, "sessions")
	if err := os.RemoveAll(sessionsPath); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sessionsPath, []byte("not a directory"), 0600); err != nil {
		t.Fatal(err)
	}
	result := h.sm.RelaunchWithChoiceCheckedResult(
		context.Background(),
		id,
		h.card,
		session.ConversationChoice{},
		func([]session.Snapshot) error { return nil },
	)
	if result.Effect != session.ActivationPublished || result.SessionID != id || result.ProcessActivationID == "" ||
		!errors.Is(result.Err, session.ErrPersistFailed) {
		t.Fatalf("publication evidence lost: %+v", result)
	}
	if err := os.Remove(sessionsPath); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(sessionsPath, 0700); err != nil {
		t.Fatal(err)
	}
	if err := h.sm.Kill(id); err != nil {
		t.Fatal(err)
	}
}

func TestWorkspaceStartSerializesWithRemoval(t *testing.T) {
	h := newGateHarness(t, false, nil)
	id := h.launchTerminated(t)
	entered, resume := make(chan struct{}), make(chan struct{})
	opened := make(chan error, 1)
	go func() {
		opened <- h.sm.OpenWorkspaceChecked(context.Background(), id, "/bin/true", func([]session.Snapshot) error { close(entered); <-resume; return nil })
	}()
	<-entered
	removed := make(chan error, 1)
	go func() { _, err := h.sm.RemoveDetailed(id); removed <- err }()
	select {
	case err := <-removed:
		t.Fatalf("removal completed before start: %v", err)
	default:
	}
	close(resume)
	if err := <-opened; err != nil {
		t.Fatalf("workspace start: %v", err)
	}
	if err := <-removed; err != nil {
		t.Fatalf("removal: %v", err)
	}
}

//nolint:govet // Independent fixture assertions use scoped error variables.
func TestLaunchAndSwitchResultsRetainPublicationOnPersistenceFailure(t *testing.T) {
	h := newGateHarness(t, false, nil)
	sessionsPath := filepath.Join(h.repoDir, "sessions")
	breakPersistence := func() {
		t.Helper()
		if err := os.RemoveAll(sessionsPath); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(sessionsPath, []byte("not a directory"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	restorePersistence := func() {
		t.Helper()
		if err := os.Remove(sessionsPath); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(sessionsPath, 0700); err != nil {
			t.Fatal(err)
		}
	}
	breakPersistence()
	id := uuid.NewString()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	release, err := h.sm.ReserveCardReferenceForDashboardWithCancel(
		h.card.ID,
		"default",
		id,
		cancel,
		func([]session.Snapshot) (bool, error) { return true, nil },
	)
	if err != nil {
		t.Fatal(err)
	}
	launch := h.sm.LaunchInDashboardWithIDResult(ctx, id, "default", h.card, nil, nil, nil, func([]session.Snapshot) error { return nil })
	release()
	if launch.Effect != session.ActivationPublished || launch.ProcessActivationID == "" ||
		!errors.Is(launch.Err, session.ErrPersistFailed) {
		t.Fatalf("launch: %+v", launch)
	}
	restorePersistence()
	if err := h.sm.Kill(id); err != nil {
		t.Fatal(err)
	}
	breakPersistence()
	switched := h.sm.SwitchAgentWithChoiceCheckedResult(
		ctx,
		id,
		h.card,
		nil,
		nil,
		nil,
		session.ConversationChoice{},
		func([]session.Snapshot) error { return nil },
	)
	if switched.Effect != session.ActivationPublished || switched.ProcessActivationID == "" ||
		switched.ProcessActivationID == launch.ProcessActivationID ||
		!errors.Is(switched.Err, session.ErrPersistFailed) {
		t.Fatalf("switch: %+v", switched)
	}
	restorePersistence()
	if err := h.sm.Kill(id); err != nil {
		t.Fatal(err)
	}
}

//nolint:govet // Independent fixture assertions use scoped error variables.
func TestPublishedProcessIdentityIsIndependentOfMCPAuthority(t *testing.T) {
	h := newGateHarness(t, false, nil)
	h.sm.SetMCPEndpoint("http://127.0.0.1:12345/mcp")
	id := uuid.NewString()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	release, err := h.sm.ReserveCardReferenceForDashboardWithCancel(
		h.card.ID,
		"default",
		id,
		cancel,
		func([]session.Snapshot) (bool, error) { return true, nil },
	)
	if err != nil {
		t.Fatal(err)
	}
	result := h.sm.LaunchInDashboardWithIDResult(
		ctx,
		id,
		"default",
		h.card,
		nil,
		nil,
		config.BuiltinMCPs(),
		func([]session.Snapshot) error { return nil },
	)
	release()
	if result.Err != nil || result.Effect != session.ActivationPublished || result.ProcessActivationID == "" {
		t.Fatalf("launch: %+v", result)
	}
	snapshot, ok := h.sm.GetSession(id)
	if !ok || snapshot.ActivationID == "" || snapshot.ActivationID == result.ProcessActivationID {
		t.Fatalf("authority and process identity: %+v, %+v", snapshot, result)
	}
	if err := h.sm.Kill(id); err != nil {
		t.Fatal(err)
	}
}

type uncertainTeardownDriver struct{ *gateDriver }

func (*uncertainTeardownDriver) Kill() error { return errors.New("injected unconfirmed kill") }

//nolint:govet // Independent fixture assertions use scoped error variables.
func TestUncertainUnpublishedSpawnRetainsTemporaryWorkspace(t *testing.T) {
	root := t.TempDir()
	t.Setenv("KANKURO_TMP_DIR", filepath.Join(root, "temporary"))
	repo, err := config.NewFileRepository(filepath.Join(root, "config"))
	if err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(root, "workspace")
	if err := os.MkdirAll(source, 0700); err != nil {
		t.Fatal(err)
	}
	if err := repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: source}}); err != nil {
		t.Fatal(err)
	}
	workspaces, err := repo.ListWorkspaces()
	if err != nil || len(workspaces) == 0 {
		t.Fatalf("workspaces: %v", err)
	}
	var made *uncertainTeardownDriver
	registry := driver.NewRegistry()
	registry.Register("uncertain", func() driver.AgentDriver {
		made = &uncertainTeardownDriver{gateDriver: &gateDriver{spawns: &atomic.Int32{}}}
		return made
	})
	manager := newTestManager(t, repo.Dir(), repo, session.WithDriverRegistry(registry))
	card := config.IDCard{ID: "card", Name: "uncertain", Driver: "uncertain", WorkspaceID: workspaces[0].ID, UseTmpWorkspace: true}
	id := uuid.NewString()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	release, err := manager.ReserveCardReferenceForDashboardWithCancel(
		card.ID,
		"default",
		id,
		cancel,
		func([]session.Snapshot) (bool, error) { return true, nil },
	)
	if err != nil {
		t.Fatal(err)
	}
	rejected := errors.New("publication denied")
	result := manager.LaunchInDashboardWithIDResult(ctx, id, "default", card, nil, nil, nil, func([]session.Snapshot) error {
		if made != nil && made.spawns.Load() > 0 {
			return rejected
		}
		return nil
	})
	release()
	if result.Effect != session.ActivationUncertain || !result.Spawned || result.ProcessActivationID != "" ||
		!errors.Is(result.Err, rejected) {
		t.Fatalf("result: %+v", result)
	}
	if _, err := os.Stat(result.WorkingDir); err != nil {
		t.Fatalf("uncertain workspace removed: %v", err)
	}
	if made == nil {
		t.Fatal("driver not created")
	}
	if err := made.gateDriver.Kill(); err != nil {
		t.Fatal(err)
	}
}

func TestActivationCancellationAfterSpawnConfirmsTeardown(t *testing.T) {
	h := newGateHarness(t, false, nil)
	id := h.launchTerminated(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result := h.sm.RelaunchWithChoiceCheckedResult(
		ctx,
		id,
		h.card,
		session.ConversationChoice{New: true},
		func([]session.Snapshot) error { cancel(); return nil },
	)
	if result.Effect != session.ActivationNotPublished || !result.Spawned || result.ProcessActivationID != "" ||
		!errors.Is(result.Err, context.Canceled) {
		t.Fatalf("cancel result: %+v", result)
	}
	snapshot, ok := h.sm.GetSession(id)
	if !ok || snapshot.Status != session.StatusTerminated {
		t.Fatalf("cancelled process published: %+v", snapshot)
	}
}
