package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BlackGiack/kankuro/internal/config"
)

func TestDashboardInventory(t *testing.T) {
	dir := t.TempDir()
	repo, err := config.NewFileRepository(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, errAt17 := repo.ListDashboards(); errAt17 == nil {
		t.Fatal("missing inventory must not imply a default dashboard")
	}
	path := filepath.Join(dir, config.DashboardsFilename)
	if errAt21 := os.WriteFile(path, []byte("schema_version: 1\ndashboards: []\n"), 0o600); errAt21 != nil {
		t.Fatal(errAt21)
	}
	d, err := repo.CreateDashboard(config.Dashboard{Name: "One", Description: "first"})
	if err != nil {
		t.Fatal(err)
	}
	if d.ID == "" || d.State != config.DashboardActive {
		t.Fatalf("invalid created dashboard: %+v", d)
	}
	if _, errAt31 := repo.CreateDashboard(config.Dashboard{ID: d.ID, Name: "duplicate"}); errAt31 == nil {
		t.Fatal("duplicate ID accepted")
	}
	if _, errAt34 := repo.UpdateDashboardDetails(d.ID, " ", ""); errAt34 == nil {
		t.Fatal("blank name accepted")
	}
	updated, err := repo.UpdateDashboardDetails(d.ID, "Two", "changed")
	if err != nil {
		t.Fatal(err)
	}
	if updated.ID != d.ID || updated.State != config.DashboardActive {
		t.Fatalf("identity or state changed: %+v", updated)
	}
	list, err := repo.ListDashboards()
	if err != nil || len(list) != 1 || list[0].Name != "Two" || list[0].Description != "changed" {
		t.Fatalf("inventory: %+v, %v", list, err)
	}
}

func TestDashboardInventoryRejectsInvalidRecords(t *testing.T) {
	for _, raw := range []string{
		"schema_version: 2\ndashboards: []\n",
		"schema_version: 1\ndashboards:\n  - {id: a, name: A, state: active}\n  - {id: a, name: B, state: active}\n",
		"schema_version: 1\ndashboards:\n  - {id: a, name: A, state: unknown}\n",
	} {
		dir := t.TempDir()
		repo, err := config.NewFileRepository(dir)
		if err != nil {
			t.Fatal(err)
		}
		if errAt61 := os.WriteFile(filepath.Join(dir, config.DashboardsFilename), []byte(raw), 0o600); errAt61 != nil {
			t.Fatal(errAt61)
		}
		if _, errAt64 := repo.ListDashboards(); errAt64 == nil {
			t.Fatalf("accepted invalid inventory: %s", raw)
		}
		if _, errAt67 := repo.CreateDashboard(config.Dashboard{Name: "new"}); errAt67 == nil {
			t.Fatalf("mutated invalid inventory: %s", raw)
		}
	}
}

func TestCardDashboardOwnershipCannotBeChanged(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if errAt78 := repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: t.TempDir()}}); errAt78 != nil {
		t.Fatal(errAt78)
	}
	workspaces, err := repo.ListWorkspaces()
	if err != nil || len(workspaces) != 1 {
		t.Fatalf("workspaces: %v, %v", workspaces, err)
	}
	card, err := repo.CreateIDCard(
		config.IDCard{Name: "card", Driver: "cat", WorkspaceID: workspaces[0].ID, DashboardID: "a", Shared: true},
	)
	if err != nil {
		t.Fatal(err)
	}
	card.DashboardID = "b"
	if _, errAt92 := repo.UpdateIDCard(card); errAt92 == nil {
		t.Fatal("ownership change accepted")
	}
	cards, err := repo.ListIDCards()
	if err != nil || len(cards) != 1 || cards[0].DashboardID != "a" {
		t.Fatalf("ownership changed: %+v, %v", cards, err)
	}
}

func TestMembershipCardSchemaIsPreservedAndValidated(t *testing.T) {
	dir := t.TempDir()
	repo, err := config.NewFileRepository(dir)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, config.IDCardsFilename)
	if errAt108 := os.WriteFile(
		path,
		[]byte("schema_version: 2\nid_cards:\n  - {id: a, name: A, dashboard_id: d}\n"),
		0o600,
	); errAt108 != nil {
		t.Fatal(errAt108)
	}
	file, err := repo.LoadIDCardsFile()
	if err != nil || file.SchemaVersion != config.CardSchemaVersion {
		t.Fatalf("load: %+v, %v", file, err)
	}
	if errAt115 := repo.WriteIDCardsFile(file); errAt115 != nil {
		t.Fatal(errAt115)
	}
	file, err = repo.LoadIDCardsFile()
	if err != nil || file.SchemaVersion != config.CardSchemaVersion {
		t.Fatalf("round trip: %+v, %v", file, err)
	}
	file.IDCards[0].DashboardID = ""
	if errAt123 := repo.WriteIDCardsFile(file); errAt123 == nil {
		t.Fatal("missing membership accepted")
	}
}
