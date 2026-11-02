package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/google/uuid"
)

type failingFinishRepository struct {
	*testRepository

	fail bool
}

func (r *failingFinishRepository) FinishCardDeletion(id string) error {
	if r.fail {
		return errors.New("injected deletion write failure")
	}
	return r.FileRepository.FinishCardDeletion(id)
}

func TestCardDeletionFailureRemainsRetryable(t *testing.T) {
	_, repo, manager, _, workspaceID := newServiceFixture(t)
	if err := os.WriteFile(
		filepath.Join(repo.Dir(), config.DashboardsFilename),
		[]byte("schema_version: 1\ndashboards:\n  - {id: a, name: A, state: active}\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	card := saveCard(t, repo.FileRepository, config.IDCard{Name: "card", Driver: "cat", WorkspaceID: workspaceID, DashboardID: "a"})
	failing := &failingFinishRepository{testRepository: repo, fail: true}
	service := New(failing, manager)
	preflight, err := service.PreflightCardDeletion("a", card.ID)
	if err != nil {
		t.Fatal(err)
	}
	result, err := service.ConfirmCardDeletion(context.Background(), "a", card.ID, preflight.Token, false)
	if err == nil || result.State != "deleting" {
		t.Fatalf("failed deletion: %+v, %v", result, err)
	}
	if _, errAt51 := service.LaunchInDashboard(context.Background(), "a", card.ID); errAt51 == nil {
		t.Fatal("deleting card could launch")
	}
	failing.fail = false
	result, err = service.RetryCardDeletion(context.Background(), "a", card.ID)
	if err != nil || result.State != "deleted" {
		t.Fatalf("retry: %+v, %v", result, err)
	}
}

//nolint:govet // Each setup and operation error is checked locally.
func TestPersistedScopeMCPCreateUsesCallerDashboard(t *testing.T) {
	setBuiltinMCPCLISkillForTest(t)
	service, repo, manager, _, workspaceID := newServiceFixture(t)
	manager.SetMCPEndpoint("http://127.0.0.1:12345/mcp")
	dashboards := "schema_version: 1\ndashboards:\n  - {id: a, name: A, state: active}\n  - {id: b, name: B, state: active}\n"
	if err := os.WriteFile(filepath.Join(repo.Dir(), config.DashboardsFilename), []byte(dashboards), 0o600); err != nil {
		t.Fatal(err)
	}
	shared := saveCard(t, repo.FileRepository, config.IDCard{Name: "shared", Driver: "cat", WorkspaceID: workspaceID,
		DashboardID: "a", Shared: true, UseTmpWorkspace: true, DashboardControl: true, MCPIDs: []string{config.BuiltinMCPKankuro}})
	ids := make(map[string]string)
	for _, dashboard := range []string{"a", "b"} {
		id, err := service.LaunchInDashboard(context.Background(), dashboard, shared.ID)
		if err != nil {
			t.Fatal(err)
		}
		ids[dashboard] = id
	}
	runner := NewOperationRunner(repo.Dir(), service)
	service.SetOperationRunner(runner)
	t.Cleanup(runner.Shutdown)
	tokens := make(map[string]string)
	for dashboard, id := range ids {
		item, ok := manager.GetSession(id)
		if !ok {
			t.Fatal("session disappeared")
		}
		tokens[dashboard] = workspaceMCPToken(t, item.WorkingDir)
	}
	if peers, err := service.ListSessions(tokens["a"]); err != nil || len(peers) != 0 {
		t.Fatalf("cross-dashboard discovery: %+v, %v", peers, err)
	}
	result, err := service.CreateSession(context.Background(), tokens["a"], uuid.NewString(), shared.ID)
	if err != nil || result.State != OperationSucceeded || result.DashboardID != "a" || result.Result == nil {
		t.Fatalf("create: %+v, %v", result, err)
	}
	created, ok := manager.GetSession(result.Result.SessionID)
	if !ok || created.DashboardID != "a" {
		t.Fatalf("created session scope: %+v, %v", created, ok)
	}
	if peers, err := service.ListSessions(tokens["b"]); err != nil || len(peers) != 0 {
		t.Fatalf("foreign sessions visible: %+v, %v", peers, err)
	}
}

func TestMembershipInventoryScope(t *testing.T) {
	dashboards := []config.Dashboard{
		{ID: "a", State: config.DashboardActive},
		{ID: "b", State: config.DashboardActive},
		{ID: "c", State: config.DashboardDeleting},
	}
	cards := []config.IDCard{
		{ID: "exclusive", DashboardID: "a"},
		{ID: "shared", DashboardID: "a", Shared: true},
		{ID: "deleting-owner", DashboardID: "c", Shared: true},
	}
	sessions := []session.Snapshot{{ID: "one", DashboardID: "a", CardID: "exclusive"}, {ID: "two", DashboardID: "b", CardID: "shared"}}
	i, err := newMembershipInventory(dashboards, cards, sessions, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		dashboard, card string
		want            bool
	}{
		{"a", "exclusive", true}, {"b", "exclusive", false}, {"b", "shared", true}, {"unknown", "shared", false}, {"c", "shared", false}, {"b", "deleting-owner", false}, {"a", "missing", false},
	} {
		if got := i.allowsCard(test.dashboard, test.card); got != test.want {
			t.Errorf("card %s in %s = %v", test.card, test.dashboard, got)
		}
	}
	if !i.allowsSession("a", "one") || i.allowsSession("a", "two") || i.allowsSession("unknown", "one") {
		t.Fatal("incorrect session scope")
	}
}

//nolint:govet,gocognit,gocyclo,cyclop // Setup failures use short local error scopes.
func TestReadDashboardDataScopesCardsAndSessions(t *testing.T) {
	service, repo, manager, _, workspaceID := newServiceFixture(t)
	dashboards := "schema_version: 1\ndashboards:\n  - {id: a, name: A, state: active}\n  - {id: b, name: B, state: active}\n"
	if err := os.WriteFile(filepath.Join(repo.Dir(), config.DashboardsFilename), []byte(dashboards), 0o600); err != nil {
		t.Fatal(err)
	}
	exclusive := saveCard(t, repo.FileRepository, config.IDCard{Name: "private", Driver: "cat", WorkspaceID: workspaceID, DashboardID: "a"})
	shared := saveCard(
		t,
		repo.FileRepository,
		config.IDCard{Name: "shared", Driver: "cat", WorkspaceID: workspaceID, DashboardID: "a", Shared: true},
	)
	foreign, err := service.CreateCardInDashboard("b", config.IDCard{Name: "foreign", Driver: "cat", WorkspaceID: workspaceID})
	if err != nil || foreign.Shared || foreign.DashboardID != "b" {
		t.Fatalf("new exclusive card: %+v, %v", foreign, err)
	}
	for _, dashboard := range []string{"a", "b"} {
		if _, err := service.LaunchInDashboard(context.Background(), dashboard, shared.ID); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := service.LaunchInDashboard(context.Background(), "b", exclusive.ID); err == nil {
		t.Fatal("foreign exclusive card accepted")
	}
	cards, err := repo.ListIDCards()
	if err != nil {
		t.Fatal(err)
	}
	for _, card := range cards {
		if card.ID == shared.ID {
			shared = card
		}
	}
	updated := shared
	updated.Shared = false
	if _, err := service.UpdateCardInDashboard("a", shared, updated); err == nil {
		t.Fatal("unshared a card referenced by another dashboard")
	}
	if _, err := service.UpdateCardInDashboard("b", shared, updated); err == nil {
		t.Fatal("consumer edited foreign card")
	}
	updated = shared
	updated.Name = "updated shared"
	shared, err = service.UpdateCardInDashboard("a", shared, updated)
	if err != nil {
		t.Fatal(err)
	}
	impact := manager.CardReferences(shared.ID)
	if len(impact.Sessions) != 2 || len(impact.PendingByDashboard) != 0 {
		t.Fatalf("published card impact: %+v", impact)
	}
	var callerA, callerB string
	for _, item := range manager.List() {
		if item.DashboardID == "a" {
			callerA = item.ID
		}
		if item.DashboardID == "b" {
			callerB = item.ID
		}
	}
	scope := persistedDashboardScope{service: service}
	if callerA == "" || callerB == "" || scope.DashboardForSession(callerB) != "b" || !scope.AllowsCard(callerB, shared.ID) ||
		scope.AllowsCard(callerB, exclusive.ID) ||
		scope.AllowsSession(callerB, callerA) {
		t.Fatal("persisted membership scope is incorrect")
	}
	dashboardPreflight, err := service.PreflightDashboardDeletion("a")
	if err != nil || !dashboardPreflight.RequiresForeignConfirmation || len(dashboardPreflight.Manifest.Cards) != 2 ||
		len(dashboardPreflight.Manifest.Sessions) != 1 {
		t.Fatalf("dashboard deletion preflight: %+v, %v", dashboardPreflight, err)
	}
	preflight, err := service.PreflightCardDeletion("a", shared.ID)
	if err != nil || !preflight.RequiresForeignConfirmation || preflight.Token == "" {
		t.Fatalf("preflight: %+v, %v", preflight, err)
	}
	if _, err := service.PreflightCardDeletion("b", shared.ID); err == nil {
		t.Fatal("consumer could preflight deletion")
	}
	verify := func() error {
		return manager.WithCardReferencesAndSnapshot(
			shared.ID,
			func(current session.CardReferenceImpact, sessions []session.Snapshot) error {
				card, err := service.cardOwnedBy("a", shared.ID, current, sessions)
				if err != nil {
					return err
				}
				return service.validateCardConfirmationLocked(preflight.Token, "a", card, current)
			},
		)
	}
	if err := verify(); err != nil {
		t.Fatalf("unchanged impact rejected: %v", err)
	}
	release, err := manager.ReserveCardReferenceForDashboardWithCancel(
		shared.ID,
		"b",
		"pending-test",
		func() {},
		func([]session.Snapshot) (bool, error) { return true, nil },
	)
	if err != nil {
		t.Fatal(err)
	}
	release()
	if err := verify(); err == nil {
		t.Fatal("changed and restored impact accepted stale confirmation")
	}
	data, err := service.ReadDashboardData("b")
	if err != nil {
		t.Fatal(err)
	}
	if len(data.Cards) != 2 || len(data.Sessions) != 1 || data.Sessions[0].DashboardID != "b" {
		t.Fatalf("dashboard b data: %+v", data)
	}
	data, err = service.ReadDashboardData("a")
	if err != nil {
		t.Fatal(err)
	}
	if len(data.Cards) != 2 || len(data.Sessions) != 1 || data.Sessions[0].DashboardID != "a" {
		t.Fatalf("dashboard a data: %+v", data)
	}
	if _, err := service.ReadDashboardData("unknown"); err == nil {
		t.Fatal("unknown dashboard accepted")
	}
	if _, err := service.ConfirmCardDeletion(context.Background(), "a", shared.ID, preflight.Token, true); err == nil {
		t.Fatal("stale confirmation accepted")
	}
	preflight, err = service.PreflightCardDeletion("a", shared.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.ConfirmCardDeletion(context.Background(), "a", shared.ID, preflight.Token, false); err == nil {
		t.Fatal("foreign references accepted without explicit confirmation")
	}
	result, err := service.ConfirmCardDeletion(context.Background(), "a", shared.ID, preflight.Token, true)
	if err != nil || result.State != "deleted" {
		t.Fatalf("confirmed deletion: %+v, %v", result, err)
	}
	result, err = service.RetryCardDeletion(context.Background(), "a", shared.ID)
	if err != nil || result.State != "deleted" {
		t.Fatalf("repeat deletion: %+v, %v", result, err)
	}
	data, err = service.ReadDashboardData("b")
	if err != nil || len(data.Cards) != 1 || data.Cards[0].ID != foreign.ID || len(data.Sessions) != 1 ||
		data.Sessions[0].CardID != shared.ID ||
		data.Sessions[0].Status == session.StatusTerminated {
		t.Fatalf("deleted card continuity: %+v, %v", data, err)
	}
	if scope.DashboardForSession(callerB) != "b" || scope.AllowsCard(callerB, shared.ID) {
		t.Fatal("deleted card changed session scope or remained launchable")
	}
	operation := Operation{DashboardID: "b", Kind: createSessionOperation, TargetSession: callerB,
		Arguments: json.RawMessage(fmt.Sprintf(`{"card_id":%q}`, shared.ID))}
	if err := service.authorizeExistingOperation(CallerContext{SessionID: callerB, DashboardID: "b"}, operation); err != nil {
		t.Fatalf("deleted-card operation history unavailable: %v", err)
	}
	if err := manager.Kill(callerB); err != nil {
		t.Fatal(err)
	}
	if err := service.RelaunchInDashboard(context.Background(), "b", callerB, session.ConversationChoice{New: true}); err == nil {
		t.Fatal("deleted card relaunched")
	}
	if err := service.SwitchInDashboard(context.Background(), "b", callerB, foreign.ID, session.ConversationChoice{New: true}); err != nil {
		t.Fatal(err)
	}
	item, ok := manager.GetSession(callerB)
	if !ok || item.DashboardID != "b" || item.CardID != foreign.ID {
		t.Fatalf("switch away: %+v, %v", item, ok)
	}
}

func TestMembershipInventoryRejectsUnknownMembership(t *testing.T) {
	dashboards := []config.Dashboard{{ID: "a", State: config.DashboardActive}}
	if _, err := newMembershipInventory(dashboards, []config.IDCard{{ID: "card", DashboardID: "missing"}}, nil, nil); err == nil {
		t.Fatal("unknown card owner accepted")
	}
	if _, err := newMembershipInventory(dashboards, nil, []session.Snapshot{{ID: "s"}}, nil); err == nil {
		t.Fatal("missing session owner accepted")
	}
	if _, err := newMembershipInventory(
		dashboards,
		nil,
		[]session.Snapshot{{ID: "s", DashboardID: "a", CardID: "missing"}},
		nil,
	); err == nil {
		t.Fatal("missing current card accepted")
	}
}

func TestMembershipInventoryRetainsDeletedCardReferences(t *testing.T) {
	dashboards := []config.Dashboard{{ID: "b", State: config.DashboardActive}}
	sessions := []session.Snapshot{
		{
			ID:            "s",
			DashboardID:   "b",
			CardID:        "deleted",
			Conversations: map[string][]session.ConversationBinding{"deleted": {{ID: "native"}}},
		},
	}
	tombstones := []config.CardTombstone{{ID: "deleted", DashboardID: "removed-owner", State: "deleted"}}
	inventory, err := newMembershipInventory(dashboards, nil, sessions, tombstones)
	if err != nil {
		t.Fatal(err)
	}
	if !inventory.allowsSession("b", "s") || inventory.allowsCard("b", "deleted") {
		t.Fatal("deleted-card reference affected session scope or became launchable")
	}
}

func TestDashboardDetailsThroughService(t *testing.T) {
	service, repo, _, _, _ := newServiceFixture(t)
	if err := os.WriteFile(
		filepath.Join(repo.Dir(), config.DashboardsFilename),
		[]byte("schema_version: 1\ndashboards: []\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	created, err := service.CreateDashboard("One", "First")
	if err != nil {
		t.Fatal(err)
	}
	updated, err := service.UpdateDashboardDetails(created.ID, "Two", "Second")
	if err != nil {
		t.Fatal(err)
	}
	dashboards, err := service.ListDashboards()
	if err != nil || len(dashboards) != 1 || dashboards[0] != updated || updated.ID != created.ID {
		t.Fatalf("dashboards: %+v, %v", dashboards, err)
	}
}

func TestDashboardDeletionManifestClosesAdmissions(t *testing.T) {
	service, repo, _, _, workspaceID := newServiceFixture(t)
	dashboards := "schema_version: 1\ndashboards:\n  - {id: a, name: A, state: active}\n  - {id: b, name: B, state: active}\n"
	if err := os.WriteFile(filepath.Join(repo.Dir(), config.DashboardsFilename), []byte(dashboards), 0o600); err != nil {
		t.Fatal(err)
	}
	shared := saveCard(
		t,
		repo.FileRepository,
		config.IDCard{Name: "shared", Driver: "cat", WorkspaceID: workspaceID, DashboardID: "a", Shared: true},
	)
	manifest := config.DashboardDeletion{
		SchemaVersion: 1,
		DashboardID:   "a",
		Name:          "A",
		Cards:         []config.DashboardDeletionCard{{ID: shared.ID}},
	}
	if err := repo.BeginDashboardDeletion(manifest); err != nil {
		t.Fatal(err)
	}
	if _, err := service.LaunchInDashboard(context.Background(), "b", shared.ID); err == nil {
		t.Fatal("foreign use of deleting owner's card accepted")
	}
	if _, err := service.CreateCardInDashboard("a", config.IDCard{Name: "new", Driver: "cat", WorkspaceID: workspaceID}); err == nil {
		t.Fatal("card created in deleting dashboard")
	}
	if _, err := service.UpdateDashboardDetails("a", "changed", ""); err == nil {
		t.Fatal("deleting dashboard edited")
	}
}

//nolint:govet // Local setup errors are checked at their point of use.
func TestScopedLaunchRevalidatesChangedCardBeforePublication(t *testing.T) {
	entered := make(chan struct{})
	continueCopy := make(chan struct{})
	copier := session.WorkspaceCopier(func(ctx context.Context, _, destination string) error {
		close(entered)
		select {
		case <-continueCopy:
		case <-ctx.Done():
			return ctx.Err()
		}
		return os.MkdirAll(destination, 0o700)
	})
	service, repo, manager, _, workspaceID := newServiceFixture(t, session.WithWorkspaceCopier(copier))
	dashboards := "schema_version: 1\ndashboards:\n  - {id: a, name: A, state: active}\n  - {id: b, name: B, state: active}\n"
	if err := os.WriteFile(filepath.Join(repo.Dir(), config.DashboardsFilename), []byte(dashboards), 0o600); err != nil {
		t.Fatal(err)
	}
	card := saveCard(
		t,
		repo.FileRepository,
		config.IDCard{Name: "shared", Driver: "cat", WorkspaceID: workspaceID, DashboardID: "a", Shared: true, UseTmpWorkspace: true},
	)
	result := make(chan error, 1)
	ctx := t.Context()
	cardID := card.ID
	go func() { _, err := service.LaunchInDashboard(ctx, "b", cardID); result <- err }()
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		t.Fatal("copy did not start")
	}
	impact := manager.CardReferences(card.ID)
	if impact.PendingByDashboard["b"] != 1 {
		t.Fatalf("pending impact: %+v", impact)
	}
	card.Shared = false
	if _, err := repo.UpdateIDCard(card); err != nil {
		close(continueCopy)
		t.Fatal(err)
	}
	close(continueCopy)
	select {
	case err := <-result:
		if err == nil {
			t.Fatal("stale activation was published")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("launch did not finish")
	}
	if len(manager.List()) != 0 {
		t.Fatal("rejected activation left a session")
	}
	if impact := manager.CardReferences(card.ID); len(impact.PendingByDashboard) != 0 {
		t.Fatalf("reservation not released: %+v", impact)
	}
}

//nolint:govet // The setup checks each failure at the point of use.
func TestConfirmedCardDeletionCancelsForeignPendingLaunch(t *testing.T) {
	entered := make(chan struct{})
	copier := session.WorkspaceCopier(func(ctx context.Context, _, _ string) error {
		close(entered)
		<-ctx.Done()
		return ctx.Err()
	})
	service, repo, manager, _, workspaceID := newServiceFixture(t, session.WithWorkspaceCopier(copier))
	dashboards := "schema_version: 1\ndashboards:\n  - {id: a, name: A, state: active}\n  - {id: b, name: B, state: active}\n"
	if err := os.WriteFile(filepath.Join(repo.Dir(), config.DashboardsFilename), []byte(dashboards), 0o600); err != nil {
		t.Fatal(err)
	}
	card := saveCard(
		t,
		repo.FileRepository,
		config.IDCard{Name: "shared", Driver: "cat", WorkspaceID: workspaceID, DashboardID: "a", Shared: true, UseTmpWorkspace: true},
	)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	launchResult := make(chan error, 1)
	go func() { _, err := service.LaunchInDashboard(ctx, "b", card.ID); launchResult <- err }()
	select {
	case <-entered:
	case <-ctx.Done():
		t.Fatal("copy did not start")
	}
	preflight, err := service.PreflightCardDeletion("a", card.ID)
	if err != nil || !preflight.RequiresForeignConfirmation || preflight.Impact.PendingByDashboard["b"] != 1 {
		t.Fatalf("pending preflight: %+v, %v", preflight, err)
	}
	result, err := service.ConfirmCardDeletion(ctx, "a", card.ID, preflight.Token, true)
	if err != nil || result.State != "deleted" {
		t.Fatalf("delete: %+v, %v", result, err)
	}
	if err := <-launchResult; err == nil {
		t.Fatal("pending activation was not cancelled")
	}
	if len(manager.List()) != 0 {
		t.Fatal("pending activation was published")
	}
}
