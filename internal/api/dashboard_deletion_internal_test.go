package api

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/session"
)

//nolint:govet // This fixture exercises both sides of the pending-activation boundary.
func TestDashboardCascadeCancelsOwnedAndForeignPendingActivations(t *testing.T) {
	entered := make(chan struct{}, 2)
	copier := session.WorkspaceCopier(func(ctx context.Context, _, _ string) error {
		entered <- struct{}{}
		<-ctx.Done()
		return ctx.Err()
	})
	service, repo, manager, _, workspaceID := newServiceFixture(t, session.WithWorkspaceCopier(copier))
	dashboards := "schema_version: 1\ndashboards:\n  - {id: a, name: A, state: active}\n  - {id: b, name: B, state: active}\n"
	if err := os.WriteFile(filepath.Join(repo.Dir(), config.DashboardsFilename), []byte(dashboards), 0o600); err != nil {
		t.Fatal(err)
	}
	owned := saveCard(
		t,
		repo.FileRepository,
		config.IDCard{Name: "owned", Driver: "cat", WorkspaceID: workspaceID, DashboardID: "a", Shared: true, UseTmpWorkspace: true},
	)
	foreign := saveCard(
		t,
		repo.FileRepository,
		config.IDCard{Name: "foreign", Driver: "cat", WorkspaceID: workspaceID, DashboardID: "b", Shared: true, UseTmpWorkspace: true},
	)
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	results := make(chan error, 2)
	go func() { _, err := service.LaunchInDashboard(ctx, "a", foreign.ID); results <- err }()
	go func() { _, err := service.LaunchInDashboard(ctx, "b", owned.ID); results <- err }()
	for range 2 {
		select {
		case <-entered:
		case <-ctx.Done():
			t.Fatal("pending activations did not start")
		}
	}
	preflight, err := service.PreflightDashboardDeletion("a")
	if err != nil || !preflight.RequiresForeignConfirmation || len(preflight.Manifest.Pending) != 2 {
		t.Fatalf("preflight: %+v, %v", preflight, err)
	}
	result, err := service.ConfirmDashboardDeletion(ctx, "a", preflight.Token, true)
	if err != nil || result.State != "deleted" {
		t.Fatalf("cascade: %+v, %v", result, err)
	}
	for range 2 {
		if err := <-results; err == nil {
			t.Fatal("pending activation was not cancelled")
		}
	}
	if len(manager.List()) != 0 {
		t.Fatal("pending activation was published")
	}
	cards, err := repo.ListIDCards()
	if err != nil || len(cards) != 1 || cards[0].ID != foreign.ID {
		t.Fatalf("foreign card: %+v, %v", cards, err)
	}
}

type failingDashboardSaveRepository struct {
	*testRepository

	failRemoved bool
}

func (r *failingDashboardSaveRepository) SaveDashboardDeletion(manifest config.DashboardDeletion) error {
	if r.failRemoved {
		for _, item := range manifest.Sessions {
			if item.Removed {
				return errors.New("injected manifest write failure")
			}
		}
	}
	return r.FileRepository.SaveDashboardDeletion(manifest)
}

func TestDashboardCascadeRetriesAfterSessionRowWasRemoved(t *testing.T) {
	_, repo, manager, _, workspaceID := newServiceFixture(t)
	if err := os.WriteFile(
		filepath.Join(repo.Dir(), config.DashboardsFilename),
		[]byte("schema_version: 1\ndashboards:\n  - {id: a, name: A, state: active}\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	card := saveCard(
		t,
		repo.FileRepository,
		config.IDCard{Name: "card", Driver: "cat", WorkspaceID: workspaceID, DashboardID: "a", UseTmpWorkspace: true},
	)
	failing := &failingDashboardSaveRepository{testRepository: repo, failRemoved: true}
	service := New(failing, manager)
	id, err := service.LaunchInDashboard(context.Background(), "a", card.ID)
	if err != nil {
		t.Fatal(err)
	}
	preflight, err := service.PreflightDashboardDeletion("a")
	if err != nil {
		t.Fatal(err)
	}
	result, err := service.ConfirmDashboardDeletion(context.Background(), "a", preflight.Token, false)
	if err == nil || result.State != "deleting" {
		t.Fatalf("injected failure: %+v, %v", result, err)
	}
	if _, exists := manager.GetSession(id); exists {
		t.Fatal("session row should already have been removed")
	}
	manifests, err := repo.ListDashboardDeletions()
	if err != nil || len(manifests) != 1 || manifests[0].Sessions[0].Removed {
		t.Fatalf("durable manifest: %+v, %v", manifests, err)
	}
	failing.failRemoved = false
	result, err = service.RetryDashboardDeletion(context.Background(), "a")
	if err != nil || result.State != "deleted" {
		t.Fatalf("retry: %+v, %v", result, err)
	}
}

//nolint:govet // The fixture exercises a restart after the durable intent boundary.
func TestDashboardCascadeResumesAfterRestart(t *testing.T) {
	service, repo, manager, _, workspaceID := newServiceFixture(t)
	dashboards := "schema_version: 1\ndashboards:\n  - {id: a, name: A, state: active}\n  - {id: b, name: B, state: active}\n"
	if err := os.WriteFile(filepath.Join(repo.Dir(), config.DashboardsFilename), []byte(dashboards), 0o600); err != nil {
		t.Fatal(err)
	}
	card := saveCard(
		t,
		repo.FileRepository,
		config.IDCard{Name: "shared", Driver: "cat", WorkspaceID: workspaceID, DashboardID: "a", Shared: true, UseTmpWorkspace: true},
	)
	idA, err := service.LaunchInDashboard(context.Background(), "a", card.ID)
	if err != nil {
		t.Fatal(err)
	}
	idB, err := service.LaunchInDashboard(context.Background(), "b", card.ID)
	if err != nil {
		t.Fatal(err)
	}
	preflight, err := service.PreflightDashboardDeletion("a")
	if err != nil {
		t.Fatal(err)
	}
	block, err := manager.ReserveDashboardDeletion("a")
	if err != nil {
		t.Fatal(err)
	}
	result, err := service.ConfirmDashboardDeletion(context.Background(), "a", preflight.Token, true)
	block()
	if err == nil || result.State != "deleting" {
		t.Fatalf("intent boundary: %+v, %v", result, err)
	}
	if _, err := service.LaunchInDashboard(context.Background(), "b", card.ID); err == nil {
		t.Fatal("deleting owner's card remained launchable")
	}
	manager.Shutdown()
	reloaded := session.NewManager(repo.Dir(), repo, session.WithStartupDelay(0))
	t.Cleanup(func() { reloaded.Shutdown() })
	resumed := New(repo, reloaded)
	result, err = resumed.RetryDashboardDeletion(context.Background(), "a")
	if err != nil || result.State != "deleted" {
		t.Fatalf("resume: %+v, %v", result, err)
	}
	if _, exists := reloaded.GetSession(idA); exists {
		t.Fatal("owned session survived restart cleanup")
	}
	if item, exists := reloaded.GetSession(idB); !exists || item.DashboardID != "b" || item.CardID != card.ID {
		t.Fatalf("foreign session: %+v, %v", item, exists)
	}
	if _, err := resumed.RetryDashboardDeletion(context.Background(), "a"); err != nil {
		t.Fatalf("completed retry: %v", err)
	}
}

//nolint:govet // One scenario checks the cross-dashboard cascade and restart boundary.
func TestDashboardCascadeDeletesOwnedSharedCardsAndPreservesForeignSession(t *testing.T) {
	service, repo, manager, _, workspaceID := newServiceFixture(t)
	dashboards := "schema_version: 1\ndashboards:\n  - {id: a, name: A, state: active}\n  - {id: b, name: B, state: active}\n"
	if err := os.WriteFile(filepath.Join(repo.Dir(), config.DashboardsFilename), []byte(dashboards), 0o600); err != nil {
		t.Fatal(err)
	}
	shared := saveCard(
		t,
		repo.FileRepository,
		config.IDCard{Name: "shared", Driver: "cat", WorkspaceID: workspaceID, DashboardID: "a", Shared: true, UseTmpWorkspace: true},
	)
	owned := saveCard(t, repo.FileRepository, config.IDCard{Name: "owned", Driver: "cat", WorkspaceID: workspaceID, DashboardID: "a"})
	foreign := saveCard(t, repo.FileRepository, config.IDCard{Name: "foreign", Driver: "cat", WorkspaceID: workspaceID, DashboardID: "b"})
	idA, err := service.LaunchInDashboard(context.Background(), "a", shared.ID)
	if err != nil {
		t.Fatal(err)
	}
	idB, err := service.LaunchInDashboard(context.Background(), "b", shared.ID)
	if err != nil {
		t.Fatal(err)
	}
	beforeA, _ := manager.GetSession(idA)
	beforeB, _ := manager.GetSession(idB)
	preflight, err := service.PreflightDashboardDeletion("a")
	if err != nil || !preflight.RequiresForeignConfirmation {
		t.Fatalf("preflight: %+v, %v", preflight, err)
	}
	if _, err := service.ConfirmDashboardDeletion(context.Background(), "a", preflight.Token, false); err == nil {
		t.Fatal("foreign references accepted without confirmation")
	}
	if _, exists := manager.GetSession(idA); !exists {
		t.Fatal("unconfirmed deletion removed a session")
	}
	result, err := service.ConfirmDashboardDeletion(context.Background(), "a", preflight.Token, true)
	if err != nil || result.State != "deleted" {
		t.Fatalf("cascade: %+v, %v", result, err)
	}
	if _, exists := manager.GetSession(idA); exists {
		t.Fatal("owned session survived")
	}
	remaining, exists := manager.GetSession(idB)
	if !exists || remaining.Status == session.StatusTerminated || remaining.CardID != shared.ID {
		t.Fatalf("foreign session: %+v, %v", remaining, exists)
	}
	if _, err := os.Stat(filepath.Dir(beforeA.WorkingDir)); !os.IsNotExist(err) {
		t.Fatalf("owned workspace remains: %v", err)
	}
	if _, err := os.Stat(beforeB.WorkingDir); err != nil {
		t.Fatalf("foreign workspace lost: %v", err)
	}
	cards, err := repo.ListIDCards()
	if err != nil || len(cards) != 1 || cards[0].ID != foreign.ID {
		t.Fatalf("remaining cards: %+v, %v (removed %s)", cards, err, owned.ID)
	}
	manager.Shutdown()
	reloaded := session.NewManager(repo.Dir(), repo, session.WithStartupDelay(0))
	t.Cleanup(func() { reloaded.Shutdown() })
	reloadedService := New(repo, reloaded)
	restored, exists := reloaded.GetSession(idB)
	if !exists || restored.DashboardID != "b" || restored.CardID != shared.ID || restored.WorkingDir != beforeB.WorkingDir {
		t.Fatalf("restored foreign session: %+v, %v", restored, exists)
	}
	if err := reloadedService.RelaunchInDashboard(context.Background(), "b", idB, session.ConversationChoice{New: true}); err == nil {
		t.Fatal("deleted card relaunched")
	}
	if err := reloadedService.SwitchInDashboard(
		context.Background(),
		"b",
		idB,
		foreign.ID,
		session.ConversationChoice{New: true},
	); err != nil {
		t.Fatal(err)
	}
}
