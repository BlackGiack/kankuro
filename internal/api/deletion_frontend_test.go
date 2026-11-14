//nolint:testpackage // Exercises package-private resource ownership.
package api

import (
	"context"

	"testing"

	"github.com/BlackGiack/kankuro/internal/api/contracts"

	"github.com/BlackGiack/kankuro/internal/config"
)

func TestDashboardDeletionProjectionDetachesReferences(t *testing.T) {
	manifest := config.DashboardDeletion{
		DashboardID: "d",
		Cards: []config.DashboardDeletionCard{
			{ID: "c", References: []config.DashboardDeletionReference{{DashboardID: "other", SessionID: "s", Binding: true}}},
		},
		Pending: []config.DashboardDeletionPending{{DashboardID: "other", SessionID: "s", WorkingDir: "private"}},
	}
	got := projectDashboardDeletion(manifest)
	if len(got.Cards) != 1 || len(got.Cards[0].References) != 1 || !got.Cards[0].References[0].Saved ||
		got.Cards[0].References[0].DashboardID != "other" {
		t.Fatalf("projection: %+v", got)
	}
	got.Cards[0].References[0].SessionID = "changed"
	if manifest.Cards[0].References[0].SessionID != "s" {
		t.Fatal("projection shares references")
	}
}

func TestCardDeletionReportsPendingWhenCleanupAdmissionFails(t *testing.T) {
	service, repo, manager, _, workspaceID := newServiceFixture(t)
	card := saveCard(
		t,
		repo.FileRepository,
		config.IDCard{Name: "pending", Driver: "cat", WorkspaceID: workspaceID, DashboardID: "default"},
	)
	prepared, err := service.PrepareHumanCardDeletion("default", card.ID)
	if err != nil {
		t.Fatal(err)
	}
	release, err := manager.ReserveCardDeletion(card.ID)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	progress, err := service.ConfirmHumanCardDeletion(context.Background(), "default", card.ID, prepared.Token, false)
	if err == nil || progress.Outcome != contracts.DeletionPendingOutcome {
		t.Fatalf("progress %+v error %v", progress, err)
	}
}

func TestDeletionPreparationFailuresReportNoIntent(t *testing.T) {
	service, _, _, _, _ := newServiceFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for _, run := range []func() (contracts.DeletionProgress, error){
		func() (contracts.DeletionProgress, error) {
			return service.ConfirmHumanCardDeletion(ctx, "default", "missing", "token", false)
		},
		func() (contracts.DeletionProgress, error) {
			return service.ConfirmHumanDashboardDeletion(ctx, "default", "token", false)
		},
	} {
		progress, err := run()
		if err == nil || progress.Outcome != contracts.DeletionNoIntent {
			t.Fatalf("progress %+v error %v", progress, err)
		}
	}
}
