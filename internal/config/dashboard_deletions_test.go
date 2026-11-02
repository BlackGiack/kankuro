package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BlackGiack/kankuro/internal/config"
)

func TestDashboardDeletionManifestRequiresCompletedEffects(t *testing.T) {
	dir := t.TempDir()
	repo, err := config.NewFileRepository(dir)
	if err != nil {
		t.Fatal(err)
	}
	if errAt17 := os.WriteFile(
		filepath.Join(dir, config.DashboardsFilename),
		[]byte("schema_version: 1\ndashboards:\n  - {id: a, name: A, state: active}\n"),
		0o600,
	); errAt17 != nil {
		t.Fatal(errAt17)
	}
	manifest := config.DashboardDeletion{SchemaVersion: 1, DashboardID: "a", Name: "A",
		Sessions: []config.DashboardDeletionSession{{ID: "session"}}, Cards: []config.DashboardDeletionCard{{ID: "card"}}}
	if errAt26 := repo.BeginDashboardDeletion(manifest); errAt26 != nil {
		t.Fatal(errAt26)
	}
	if errAt29 := repo.FinalizeDashboardDeletion("a"); errAt29 == nil {
		t.Fatal("incomplete deletion finalized")
	}
	manifests, err := repo.ListDashboardDeletions()
	if err != nil || len(manifests) != 1 || manifests[0].DashboardID != "a" {
		t.Fatalf("manifests: %+v, %v", manifests, err)
	}
	manifest.AdmissionsDrained, manifest.HistoryRemoved = true, true
	manifest.Sessions[0].Stopped, manifest.Sessions[0].Removed = true, true
	manifest.Sessions[0].RecordDeleted, manifest.Sessions[0].TemporaryRemoved, manifest.Cards[0].Deleted = true, true, true
	if errAt39 := repo.SaveDashboardDeletion(manifest); errAt39 != nil {
		t.Fatal(errAt39)
	}
	if errAt42 := repo.FinalizeDashboardDeletion("a"); errAt42 != nil {
		t.Fatal(errAt42)
	}
	dashboards, err := repo.ListDashboards()
	if err != nil || len(dashboards) != 0 {
		t.Fatalf("dashboards: %+v, %v", dashboards, err)
	}
	manifests, err = repo.ListDashboardDeletions()
	if err != nil || len(manifests) != 0 {
		t.Fatalf("manifests: %+v, %v", manifests, err)
	}
	if _, errAt53 := repo.CreateDashboard(config.Dashboard{ID: "a", Name: "replacement"}); errAt53 == nil {
		t.Fatal("deleted dashboard ID reused")
	}
}

func TestDashboardDeletionRejectsUnsafeTemporaryPath(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	manifest := config.DashboardDeletion{SchemaVersion: 1, DashboardID: "a", Name: "A", Sessions: []config.DashboardDeletionSession{
		{ID: "session", UseTmpWorkspace: true, TemporaryRoot: "/tmp/other", WorkingDir: "/tmp/other/work"},
	}}
	if errAt66 := repo.BeginDashboardDeletion(manifest); errAt66 == nil {
		t.Fatal("unsafe temporary cleanup path accepted")
	}
}
