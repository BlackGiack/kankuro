package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BlackGiack/kankuro/internal/config"
)

func TestCardDeletionIntentClosesAndRetainsIdentity(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if errAt16 := repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: t.TempDir()}}); errAt16 != nil {
		t.Fatal(errAt16)
	}
	workspaces, err := repo.ListWorkspaces()
	if err != nil || len(workspaces) != 1 {
		t.Fatalf("workspaces: %v, %v", workspaces, err)
	}
	card, err := repo.CreateIDCard(
		config.IDCard{Name: "shared", Driver: "cat", WorkspaceID: workspaces[0].ID, DashboardID: "a", Shared: true},
	)
	if err != nil {
		t.Fatal(err)
	}
	if errAt29 := repo.BeginCardDeletion(card); errAt29 != nil {
		t.Fatal(errAt29)
	}
	if _, errAt32 := repo.UpdateIDCard(card); errAt32 == nil {
		t.Fatal("deleting card could be saved")
	}
	if errAt35 := repo.FinishCardDeletion(card.ID); errAt35 != nil {
		t.Fatal(errAt35)
	}
	if errAt38 := repo.FinishCardDeletion(card.ID); errAt38 != nil {
		t.Fatalf("repeat finish: %v", errAt38)
	}
	if _, errAt41 := repo.CreateIDCard(card); errAt41 == nil {
		t.Fatal("deleted card ID was reused")
	}
	cards, err := repo.ListIDCards()
	if err != nil || len(cards) != 0 {
		t.Fatalf("cards: %+v, %v", cards, err)
	}
	tombstones, err := repo.ListCardTombstones()
	if err != nil || len(tombstones) != 1 || tombstones[0].ID != card.ID || tombstones[0].State != "deleted" ||
		tombstones[0].DashboardID != "a" {
		t.Fatalf("tombstones: %+v, %v", tombstones, err)
	}
	reopened, err := config.NewFileRepository(repo.Dir())
	if err != nil {
		t.Fatal(err)
	}
	if _, errAt57 := reopened.CreateIDCard(card); errAt57 == nil {
		t.Fatal("deleted identity was reusable after restart")
	}
}

func TestCorruptTombstoneInventoryBlocksCardWrites(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if errAt67 := os.WriteFile(
		filepath.Join(repo.Dir(), config.CardTombstonesFilename),
		[]byte("schema_version: 99\ntombstones: []\n"),
		0o600,
	); errAt67 != nil {
		t.Fatal(errAt67)
	}
	if _, errAt74 := repo.CreateIDCard(config.IDCard{Name: "card"}); errAt74 == nil {
		t.Fatal("corrupt tombstones allowed a card write")
	}
}

func TestCardDeletionCannotFinishBeforePendingCleanup(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if errAt84 := repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: t.TempDir()}}); errAt84 != nil {
		t.Fatal(errAt84)
	}
	workspaces, err := repo.ListWorkspaces()
	if err != nil || len(workspaces) != 1 {
		t.Fatalf("workspaces: %+v %v", workspaces, err)
	}
	card, err := repo.CreateIDCard(config.IDCard{Name: "card", Driver: "cat", WorkspaceID: workspaces[0].ID, DashboardID: "a"})
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(t.TempDir(), "pending-session")
	pending := config.CardDeletionPending{
		DashboardID:   "b",
		SessionID:     "pending-session",
		TemporaryRoot: root,
		WorkingDir:    filepath.Join(root, "workspace"),
	}
	if errAt102 := repo.BeginCardDeletionWithPending(card, []config.CardDeletionPending{pending}); errAt102 != nil {
		t.Fatal(errAt102)
	}
	if errAt105 := repo.FinishCardDeletion(card.ID); errAt105 == nil {
		t.Fatal("unfinished cleanup allowed deletion to complete")
	}
	changed := pending
	changed.WorkingDir = filepath.Join(root, "other")
	if errAt110 := repo.SaveCardDeletionPending(card.ID, changed); errAt110 == nil {
		t.Fatal("cleanup path was allowed to change")
	}
	pending.Cleaned = true
	if errAt114 := repo.SaveCardDeletionPending(card.ID, pending); errAt114 != nil {
		t.Fatal(errAt114)
	}
	if errAt117 := repo.FinishCardDeletion(card.ID); errAt117 != nil {
		t.Fatal(errAt117)
	}
}
