package api

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/session"
)

//nolint:gocognit,gocyclo,cyclop // The scenario follows the failed cleanup across a restart.
func TestCardDeletionRetriesPendingWorkspaceCleanupAfterRestart(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("permission-based cleanup failure requires a non-root user")
	}
	t.Setenv("KANKURO_TMP_DIR", t.TempDir())
	entered := make(chan string, 1)
	copier := session.WorkspaceCopier(func(ctx context.Context, _, destination string) error {
		if err := os.MkdirAll(destination, 0o700); err != nil {
			return err
		}
		root := filepath.Dir(destination)
		if err := os.Chmod(root, 0o500); err != nil {
			return err
		}
		entered <- destination
		<-ctx.Done()
		return ctx.Err()
	})
	service, repo, manager, originalWorkspace, workspaceID := newServiceFixture(t, session.WithWorkspaceCopier(copier))
	dashboards := "schema_version: 1\ndashboards:\n  - {id: a, name: A, state: active}\n  - {id: b, name: B, state: active}\n"
	if err := os.WriteFile(filepath.Join(repo.Dir(), config.DashboardsFilename), []byte(dashboards), 0o600); err != nil {
		t.Fatal(err)
	}
	shared := saveCard(
		t,
		repo.FileRepository,
		config.IDCard{Name: "shared", Driver: "cat", WorkspaceID: workspaceID, DashboardID: "a", Shared: true, UseTmpWorkspace: true},
	)
	foreign := saveCard(t, repo.FileRepository, config.IDCard{Name: "foreign", Driver: "cat", WorkspaceID: workspaceID, DashboardID: "b"})
	survivingID, err := service.LaunchInDashboard(context.Background(), "b", foreign.ID)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	launch := make(chan error, 1)
	go func() { _, launchErr := service.LaunchInDashboard(ctx, "b", shared.ID); launch <- launchErr }()
	var destination string
	select {
	case destination = <-entered:
	case <-ctx.Done():
		t.Fatal("pending copy did not start")
	}
	root := filepath.Dir(destination)
	t.Cleanup(func() {
		if _, statErr := os.Stat(root); statErr == nil {
			if chmodErr := os.Chmod(root, 0o700); chmodErr != nil {
				t.Errorf("restore fixture permissions: %v", chmodErr)
			}
		} else if !os.IsNotExist(statErr) {
			t.Errorf("inspect fixture root: %v", statErr)
		}
	})
	// Editing the card while the copy is blocked must not change the cleanup
	// identity captured by the already-admitted activation.
	newWorkspace := t.TempDir()
	if errAt71 := repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: originalWorkspace}, {Path: newWorkspace}}); errAt71 != nil {
		t.Fatal(errAt71)
	}
	workspaces, err := repo.ListWorkspaces()
	if err != nil {
		t.Fatal(err)
	}
	cards, err := repo.ListIDCards()
	if err != nil {
		t.Fatal(err)
	}
	for _, card := range cards {
		if card.ID == shared.ID {
			shared = card
			break
		}
	}
	updated := shared
	for _, workspace := range workspaces {
		if workspace.Path == newWorkspace {
			updated.WorkspaceID = workspace.ID
		}
	}
	if updated.WorkspaceID == shared.WorkspaceID {
		t.Fatal("new workspace was not discovered")
	}
	if _, errAt97 := service.UpdateCardInDashboard("a", shared, updated); errAt97 != nil {
		t.Fatal(errAt97)
	}
	preflight, err := service.PreflightCardDeletion("a", shared.ID)
	if err != nil || !preflight.RequiresForeignConfirmation {
		t.Fatalf("preflight: %+v %v", preflight, err)
	}
	result, err := service.ConfirmCardDeletion(ctx, "a", shared.ID, preflight.Token, true)
	if err == nil || result.State != config.CardDeleting {
		t.Fatalf("cleanup failure must remain pending: %+v %v", result, err)
	}
	if launchErr := <-launch; launchErr == nil {
		t.Fatal("cancelled launch succeeded")
	}
	if _, errAt111 := os.Stat(destination); errAt111 != nil {
		t.Fatalf("expected workspace after failed cleanup: %v", errAt111)
	}
	tombstones, err := repo.ListCardTombstones()
	if err != nil || len(tombstones) != 1 {
		t.Fatalf("tombstones: %+v %v", tombstones, err)
	}
	pending := tombstones[0].Pending
	if len(pending) != 1 || pending[0].TemporaryRoot != root || pending[0].WorkingDir != destination || pending[0].Cleaned ||
		pending[0].Error == "" {
		t.Fatalf("durable pending cleanup: %+v", pending)
	}
	if _, ok := manager.GetSession(survivingID); !ok {
		t.Fatal("foreign session disappeared")
	}
	manager.Shutdown()
	if errAt127 := os.Chmod(root, 0o700); errAt127 != nil {
		t.Fatal(errAt127)
	}
	restarted := session.NewManager(repo.Dir(), repo, session.WithStartupDelay(0))
	t.Cleanup(func() { restarted.Shutdown() })
	if errAt132 := restarted.LoadErrors(); errAt132 != nil {
		t.Fatal(errAt132)
	}
	retry := New(repo, restarted)
	result, err = retry.RetryCardDeletion(context.Background(), "a", shared.ID)
	if err != nil || result.State != config.CardDeleted {
		t.Fatalf("retry: %+v %v", result, err)
	}
	if _, errAt140 := os.Stat(root); !os.IsNotExist(errAt140) {
		t.Fatalf("exact temporary root still exists: %v", errAt140)
	}
	tombstones, err = repo.ListCardTombstones()
	if err != nil || len(tombstones) != 1 || len(tombstones[0].Pending) != 1 || !tombstones[0].Pending[0].Cleaned ||
		tombstones[0].Pending[0].Error != "" {
		t.Fatalf("completed cleanup: %+v %v", tombstones, err)
	}
	if _, ok := restarted.GetSession(survivingID); !ok {
		t.Fatal("foreign session was not preserved after restart")
	}
}
