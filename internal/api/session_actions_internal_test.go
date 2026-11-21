package api

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"testing"

	"github.com/BlackGiack/kankuro/internal/config"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/session"
)

func TestLifecycleResultMapsPublicationEvidence(t *testing.T) {
	warning := errors.New("metadata warning")
	for _, tc := range []struct {
		name     string
		evidence session.ActivationResult
		err      error
		want     contracts.LifecycleEffect
	}{
		{"failure", session.ActivationResult{}, errors.New("failed"), contracts.NoEffect},
		{"published warning", session.ActivationResult{Effect: session.ActivationPublished, SessionID: "s", ProcessActivationID: "p"}, warning, contracts.UsableActivation},
		{"uncertain", session.ActivationResult{Effect: session.ActivationUncertain, SessionID: "s", PID: 42}, errors.New("teardown uncertain"), contracts.UncertainEffect},
		{"invalid published", session.ActivationResult{Effect: session.ActivationPublished, SessionID: "s"}, nil, contracts.UncertainEffect},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := lifecycleResult(tc.evidence, tc.err)
			if got.Effect != tc.want {
				t.Fatalf("got %+v, want %s", got, tc.want)
			}
			if err := got.Validate(); err != nil {
				t.Fatalf("invalid mapped result %+v: %v", got, err)
			}
			if tc.err != nil && got.Diagnostic != tc.err.Error() {
				t.Fatalf("diagnostic %q", got.Diagnostic)
			}
		})
	}
}

func TestSessionContractsRevalidateScopeAndChoice(t *testing.T) {
	service, repo, manager, _, workspaceID := newServiceFixture(t)
	card := saveCard(
		t,
		repo.FileRepository,
		config.IDCard{Name: "agent", Driver: "cat", WorkspaceID: workspaceID, DashboardID: DefaultDashboardID},
	)
	launch := service.LaunchSession(context.Background(), contracts.LaunchSessionRequest{DashboardID: DefaultDashboardID, CardID: card.ID})
	if launch.Effect != contracts.UsableActivation || launch.Validate() != nil {
		t.Fatalf("launch: %+v", launch)
	}
	if _, err := service.ReadConversationPreparation(
		context.Background(),
		contracts.ConversationPreparationRequest{DashboardID: "foreign", SessionID: launch.SessionID},
	); err == nil {
		t.Fatal("foreign dashboard read succeeded")
	}
	if err := manager.Kill(launch.SessionID); err != nil {
		t.Fatal(err)
	}
	read, err := service.ReadConversationPreparation(
		context.Background(),
		contracts.ConversationPreparationRequest{DashboardID: DefaultDashboardID, SessionID: launch.SessionID},
	)
	if err != nil || !read.Availability.Enabled || read.CardID != card.ID || read.HasSaved {
		t.Fatalf("preparation: %+v, %v", read, err)
	}
	invalid := service.RelaunchDashboardSession(
		context.Background(),
		contracts.RelaunchSessionRequest{DashboardID: DefaultDashboardID, SessionID: launch.SessionID},
	)
	if invalid.Effect != contracts.NoEffect {
		t.Fatalf("empty choice: %+v", invalid)
	}
	foreign := service.RelaunchDashboardSession(
		context.Background(),
		contracts.RelaunchSessionRequest{
			DashboardID: "foreign",
			SessionID:   launch.SessionID,
			CardID:      card.ID,
			Choice:      contracts.ConversationChoice{Kind: contracts.ConversationNew},
		},
	)
	if foreign.Effect != contracts.NoEffect {
		t.Fatalf("foreign scope: %+v", foreign)
	}
	resumed := service.RelaunchDashboardSession(
		context.Background(),
		contracts.RelaunchSessionRequest{
			DashboardID: DefaultDashboardID,
			SessionID:   launch.SessionID,
			CardID:      card.ID,
			Choice:      contracts.ConversationChoice{Kind: contracts.ConversationNew},
		},
	)
	if resumed.Effect != contracts.UsableActivation || resumed.ProcessActivationID == launch.ProcessActivationID {
		t.Fatalf("relaunch: %+v", resumed)
	}
}

func TestOpenSessionWorkspaceChecksScopeAndConfiguration(t *testing.T) {
	service, repo, manager, _, workspaceID := newServiceFixture(t)
	card := saveCard(
		t,
		repo.FileRepository,
		config.IDCard{Name: "agent", Driver: "cat", WorkspaceID: workspaceID, DashboardID: DefaultDashboardID},
	)
	launch := service.LaunchSession(context.Background(), contracts.LaunchSessionRequest{DashboardID: DefaultDashboardID, CardID: card.ID})
	if launch.Effect != contracts.UsableActivation {
		t.Fatalf("launch: %+v", launch)
	}
	request := contracts.OpenSessionWorkspaceRequest{DashboardID: "foreign", SessionID: launch.SessionID, Kind: contracts.WorkspaceIDE}
	if result := service.OpenSessionWorkspace(context.Background(), request); result.Diagnostic == "" {
		t.Fatal("foreign scope accepted")
	}
	request.DashboardID = DefaultDashboardID
	request.Kind = "unknown"
	if result := service.OpenSessionWorkspace(context.Background(), request); result.Diagnostic == "" {
		t.Fatal("unknown kind accepted")
	}
	request.Kind = contracts.WorkspaceIDE
	if err := repo.SaveAppSettings(config.AppSettings{IDECommand: "/bin/true"}); err != nil {
		t.Fatal(err)
	}
	if result := service.OpenSessionWorkspace(context.Background(), request); !result.Started {
		t.Fatalf("configured command: %+v", result)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if result := service.OpenSessionWorkspace(ctx, request); result.Diagnostic != context.Canceled.Error() {
		t.Fatalf("cancellation: %+v", result)
	}
	if err := manager.Kill(launch.SessionID); err != nil {
		t.Fatal(err)
	}
}

func TestScopedMutationsRetainPartialRenameAndRejectForeignScope(t *testing.T) {
	service, repo, manager, _, workspaceID := newServiceFixture(t)
	card := saveCard(
		t,
		repo.FileRepository,
		config.IDCard{Name: "agent", Driver: "cat", WorkspaceID: workspaceID, DashboardID: DefaultDashboardID},
	)
	launched := service.LaunchSession(
		context.Background(),
		contracts.LaunchSessionRequest{DashboardID: DefaultDashboardID, CardID: card.ID},
	)
	if launched.Effect != contracts.UsableActivation {
		t.Fatalf("launch: %+v", launched)
	}
	foreign := contracts.ScopedSessionRequest{DashboardID: "foreign", SessionID: launched.SessionID}
	if result := service.StopDashboardSession(context.Background(), foreign); result.Diagnostic == "" || result.ProcessTerminated {
		t.Fatalf("foreign stop: %+v", result)
	}
	if result := service.RemoveDashboardSession(context.Background(), foreign); result.Diagnostic == "" || result.InMemoryRemoved {
		t.Fatalf("foreign remove: %+v", result)
	}
	sessionsPath := filepath.Join(repo.Dir(), "sessions")
	if err := os.RemoveAll(sessionsPath); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sessionsPath, []byte("not a directory"), 0600); err != nil {
		t.Fatal(err)
	}
	renamed := service.RenameDashboardSession(
		context.Background(),
		contracts.RenameSessionRequest{DashboardID: DefaultDashboardID, SessionID: launched.SessionID, Name: "changed"},
	)
	if !renamed.Applied || renamed.Persisted || renamed.Diagnostic == "" {
		t.Fatalf("rename: %+v", renamed)
	}
	snapshot, ok := manager.GetSession(launched.SessionID)
	if !ok || snapshot.Name != "changed" {
		t.Fatalf("in-memory rename lost: %+v", snapshot)
	}
	if err := os.Remove(sessionsPath); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(sessionsPath, 0700); err != nil {
		t.Fatal(err)
	}
	if result := service.StopDashboardSession(
		context.Background(),
		contracts.ScopedSessionRequest{DashboardID: DefaultDashboardID, SessionID: launched.SessionID},
	); !result.ProcessTerminated ||
		result.Diagnostic != "" {
		t.Fatalf("stop: %+v", result)
	}
}

func TestScopedStopCancellationWhileWaitingForAdmissionHasNoEffect(t *testing.T) {
	service, repo, manager, _, workspaceID := newServiceFixture(t)
	card := saveCard(
		t,
		repo.FileRepository,
		config.IDCard{Name: "agent", Driver: "cat", WorkspaceID: workspaceID, DashboardID: DefaultDashboardID},
	)
	launch := service.LaunchSession(context.Background(), contracts.LaunchSessionRequest{DashboardID: DefaultDashboardID, CardID: card.ID})
	if launch.Effect != contracts.UsableActivation {
		t.Fatalf("launch: %+v", launch)
	}
	entered, release := make(chan struct{}), make(chan struct{})
	admissionDone := make(chan error, 1)
	go func() {
		admissionDone <- manager.WithAdmissionSnapshot(func(session.AdmissionSnapshot) error { close(entered); <-release; return nil })
	}()
	<-entered
	ctx, cancel := context.WithCancel(context.Background())
	stopped := make(chan contracts.StopSessionResult, 1)
	go func() {
		stopped <- service.StopDashboardSession(ctx, contracts.ScopedSessionRequest{DashboardID: DefaultDashboardID, SessionID: launch.SessionID})
	}()
	cancel()
	close(release)
	if err := <-admissionDone; err != nil {
		t.Fatal(err)
	}
	result := <-stopped
	if result.ProcessTerminated || result.DiagnosticCode != contracts.DiagnosticCancelled {
		t.Fatalf("cancelled stop: %+v", result)
	}
	snapshot, ok := manager.GetSession(launch.SessionID)
	if !ok || snapshot.Status == session.StatusTerminated {
		t.Fatalf("cancelled stop affected process: %+v", snapshot)
	}
}

func TestLaunchWorkspaceOverridesDoNotChangeCard(t *testing.T) {
	service, repo, manager, original, originalID := newServiceFixture(t)
	other := t.TempDir()
	if err := repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: original}, {Path: other}}); err != nil {
		t.Fatal(err)
	}
	workspaces, err := repo.ListWorkspaces()
	if err != nil {
		t.Fatal(err)
	}
	var otherID string
	for _, ws := range workspaces {
		if ws.Path == other {
			otherID = ws.ID
		}
	}
	card := saveCard(
		t,
		repo.FileRepository,
		config.IDCard{Name: "override", Driver: "cat", WorkspaceID: originalID, DashboardID: DefaultDashboardID},
	)
	for _, ids := range [][]string{{}, {""}, {"unknown"}, {otherID, otherID}} {
		result := service.LaunchSession(
			context.Background(),
			contracts.LaunchSessionRequest{DashboardID: DefaultDashboardID, CardID: card.ID, WorkspaceIDs: ids},
		)
		if result.Effect != contracts.NoEffect {
			t.Fatalf("invalid IDs %v: %+v", ids, result)
		}
	}
	result := service.LaunchSession(
		context.Background(),
		contracts.LaunchSessionRequest{DashboardID: DefaultDashboardID, CardID: card.ID, WorkspaceIDs: []string{otherID}},
	)
	if result.Effect != contracts.UsableActivation {
		t.Fatalf("override launch: %+v", result)
	}
	snapshot, ok := manager.GetSession(result.SessionID)
	if !ok || snapshot.WorkingDir != other || snapshot.OriginalDir != other {
		t.Fatalf("override workspace: %+v", snapshot)
	}
	if killErr := manager.Kill(result.SessionID); killErr != nil {
		t.Fatal(killErr)
	}
	if relaunchErr := service.RelaunchInDashboard(
		context.Background(),
		DefaultDashboardID,
		result.SessionID,
		session.ConversationChoice{New: true},
	); relaunchErr != nil {
		t.Fatal(relaunchErr)
	}
	relaunched, ok := manager.GetSession(result.SessionID)
	if !ok || relaunched.WorkingDir != other || relaunched.OriginalDir != other {
		t.Fatalf("relaunch changed workspace: %+v", relaunched)
	}
	saved, err := service.loadCard(card.ID)
	if err != nil || saved.WorkspaceID != originalID {
		t.Fatalf("card mutated: %+v %v", saved, err)
	}
}
