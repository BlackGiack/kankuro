package session_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/google/uuid"
)

//nolint:gocognit // One scenario checks reservation, rejection, cancellation and persistence.
func TestDashboardLaunchPersistsMembershipBeforeReturn(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if errAt20 := repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: t.TempDir()}}); errAt20 != nil {
		t.Fatal(errAt20)
	}
	workspaces, err := repo.ListWorkspaces()
	if err != nil || len(workspaces) != 1 {
		t.Fatalf("workspaces: %v, %v", workspaces, err)
	}
	m := newTestManager(t, repo.Dir(), repo)
	card := config.IDCard{ID: "card", Name: "agent", Driver: "cat", WorkspaceID: workspaces[0].ID}
	launch := func(ctx context.Context, id string, check func([]session.Snapshot) error) error {
		release, reserveErr := m.ReserveCardReferenceForDashboardWithCancel(
			card.ID,
			"dashboard-b",
			id,
			func() {},
			func([]session.Snapshot) (bool, error) { return true, nil },
		)
		if reserveErr != nil {
			return reserveErr
		}
		defer release()
		_, launchErr := m.LaunchInDashboardWithIDContext(ctx, id, "dashboard-b", card, nil, nil, nil, check)
		return launchErr
	}
	id := uuid.NewString()
	if errAt30 := launch(
		context.Background(),
		id,
		func([]session.Snapshot) error { return nil },
	); errAt30 != nil {
		t.Fatal(errAt30)
	}
	snapshot, ok := m.GetSession(id)
	if !ok || snapshot.DashboardID != "dashboard-b" {
		t.Fatalf("snapshot: %+v, %v", snapshot, ok)
	}
	raw, err := os.ReadFile(filepath.Join(repo.Dir(), "sessions", id+".json"))
	if err != nil {
		t.Fatal(err)
	}
	var data session.Data
	if errAt42 := json.Unmarshal(raw, &data); errAt42 != nil {
		t.Fatal(errAt42)
	}
	if data.DashboardID != "dashboard-b" {
		t.Fatalf("persisted dashboard: %q", data.DashboardID)
	}
	if errAt48 := m.Kill(id); errAt48 != nil {
		t.Fatal(errAt48)
	}
	if errAt51 := testRelaunch(m, id, card); errAt51 != nil {
		t.Fatal(errAt51)
	}
	snapshot, ok = m.GetSession(id)
	if !ok || snapshot.DashboardID != "dashboard-b" {
		t.Fatalf("membership changed on relaunch: %+v, %v", snapshot, ok)
	}
	rejectedID := uuid.NewString()
	revoked := errors.New("membership revoked")
	rejectErr := launch(context.Background(), rejectedID,
		func([]session.Snapshot) error { return revoked })
	if !errors.Is(rejectErr, revoked) {
		t.Fatalf("publication error: %v", rejectErr)
	}
	if _, found := m.GetSession(rejectedID); found {
		t.Fatal("rejected activation was published")
	}
	if _, statErr := os.Stat(filepath.Join(repo.Dir(), "sessions", rejectedID+".json")); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("rejected record exists: %v", statErr)
	}
	cancelledID := uuid.NewString()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cancelErr := launch(ctx, cancelledID,
		func([]session.Snapshot) error { cancel(); return nil })
	if !errors.Is(cancelErr, context.Canceled) {
		t.Fatalf("cancellation error: %v", cancelErr)
	}
	if _, found := m.GetSession(cancelledID); found {
		t.Fatal("cancelled activation was published")
	}
}
