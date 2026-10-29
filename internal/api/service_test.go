package api //nolint:testpackage // private resolver is tested alongside its public callers.

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/driver"
	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/BlackGiack/kankuro/internal/testenv"
	"github.com/google/uuid"
)

func TestMain(m *testing.M) {
	puppet, err := os.ReadFile("../../kankuro_skills/kankuro-puppet/SKILL.md")
	if err != nil {
		os.Exit(2)
	}
	config.SetBuiltinKankuroPuppetSkillContent(puppet)
	os.Exit(testenv.Run(m, "api"))
}

type testRepository struct {
	*config.FileRepository

	mu sync.Mutex

	cardsErr, skillsErr, personasErr, mcpsErr, deleteErr error
	cardReads, skillReads, personaReads, mcpReads        int
}

type testDashboardScope struct {
	mu       sync.RWMutex
	sessions map[string]string
	cards    map[string]string
}

func (scope *testDashboardScope) DashboardForSession(id string) string {
	scope.mu.RLock()
	defer scope.mu.RUnlock()
	return scope.sessions[id]
}
func (scope *testDashboardScope) AllowsSession(callerID, targetID string) bool {
	scope.mu.RLock()
	defer scope.mu.RUnlock()
	return scope.sessions[callerID] != "" && scope.sessions[callerID] == scope.sessions[targetID]
}
func (scope *testDashboardScope) AllowsCard(callerID, cardID string) bool {
	scope.mu.RLock()
	defer scope.mu.RUnlock()
	return scope.sessions[callerID] != "" && scope.sessions[callerID] == scope.cards[cardID]
}
func (scope *testDashboardScope) AssignSession(callerID, targetID, _ string) {
	scope.mu.Lock()
	defer scope.mu.Unlock()
	scope.sessions[targetID] = scope.sessions[callerID]
}
func (scope *testDashboardScope) setSession(id, dashboard string) {
	scope.mu.Lock()
	defer scope.mu.Unlock()
	scope.sessions[id] = dashboard
}
func (scope *testDashboardScope) setCard(id, dashboard string) {
	scope.mu.Lock()
	defer scope.mu.Unlock()
	scope.cards[id] = dashboard
}
func (scope *testDashboardScope) SessionDashboard(id string) (string, bool) {
	scope.mu.RLock()
	defer scope.mu.RUnlock()
	dashboard, exists := scope.sessions[id]
	return dashboard, exists
}
func (scope *testDashboardScope) CardDashboard(id string) (string, bool) {
	scope.mu.RLock()
	defer scope.mu.RUnlock()
	dashboard, exists := scope.cards[id]
	return dashboard, exists
}

func (r *testRepository) ListIDCards() ([]config.IDCard, error) {
	r.mu.Lock()
	r.cardReads++
	r.mu.Unlock()
	if r.cardsErr != nil {
		return nil, r.cardsErr
	}
	return r.FileRepository.ListIDCards()
}
func (r *testRepository) DeleteIDCard(id string) error {
	if r.deleteErr != nil {
		return r.deleteErr
	}
	return r.FileRepository.DeleteIDCard(id)
}
func (r *testRepository) ListSkills() ([]config.Skill, error) {
	r.mu.Lock()
	r.skillReads++
	r.mu.Unlock()
	if r.skillsErr != nil {
		return nil, r.skillsErr
	}
	return r.FileRepository.ListSkills()
}
func (r *testRepository) ListPersonas() ([]config.Persona, error) {
	r.mu.Lock()
	r.personaReads++
	r.mu.Unlock()
	if r.personasErr != nil {
		return nil, r.personasErr
	}
	return r.FileRepository.ListPersonas()
}
func (r *testRepository) ListMCPs() ([]config.MCPServer, error) {
	r.mu.Lock()
	r.mcpReads++
	r.mu.Unlock()
	if r.mcpsErr != nil {
		return nil, r.mcpsErr
	}
	return r.FileRepository.ListMCPs()
}

//nolint:govet // short declarations localize setup failures in this test helper.
func newServiceFixture(t *testing.T, managerOptions ...session.Option) (*Service, *testRepository, *session.Manager, string, string) {
	t.Helper()
	repoDir := t.TempDir()
	base, err := config.NewFileRepository(repoDir)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}
	workspace := t.TempDir()
	if err := base.SaveWorkspaceSources([]config.WorkspaceSource{{Path: workspace}}); err != nil {
		t.Fatalf("save workspace source: %v", err)
	}
	workspaces, err := base.ListWorkspaces()
	if err != nil || len(workspaces) != 1 {
		t.Fatalf("list workspaces: %v (%d)", err, len(workspaces))
	}
	repo := &testRepository{FileRepository: base}
	sessions := session.NewManager(repoDir, repo, append(managerOptions, session.WithStartupDelay(0))...)
	t.Cleanup(func() { sessions.Shutdown() })
	return New(repo, sessions), repo, sessions, workspace, workspaces[0].ID
}

func saveCard(t *testing.T, repo *config.FileRepository, card config.IDCard) config.IDCard {
	t.Helper()
	saved, err := repo.CreateIDCard(card)
	if err != nil {
		t.Fatalf("save card %q: %v", card.Name, err)
	}
	return saved
}

func TestListCardsReturnsResolvedCopiedWorkspaceData(t *testing.T) {
	service, _, _, _, _ := newServiceFixture(t)
	if _, err := service.ListCards(""); err == nil {
		t.Fatal("unauthenticated card discovery succeeded")
	}
	if _, err := service.ListSessions(""); err == nil {
		t.Fatal("unauthenticated session discovery succeeded")
	}
	if _, err := service.GetSession("", "target"); err == nil {
		t.Fatal("unauthenticated inspection succeeded")
	}
}

func TestDashboardScopeResolverFiltersCardsAndSessions(t *testing.T) {
	service, _, _, _, _ := newServiceFixture(t)
	service.SetDashboardScopeResolver(&testDashboardScope{
		sessions: map[string]string{"caller": "factory-a"}, cards: map[string]string{},
	})
	if _, scopeErr := service.ListCards("invalid"); scopeErr == nil {
		t.Fatal("fabricated caller metadata authorized discovery")
	}
}

//nolint:gocognit,gocyclo,cyclop // One live-session fixture verifies lifecycle, replay, restart recovery, and cancellation.
func TestOperationRunnerCreatesOnceAndReplaysRequest(t *testing.T) {
	skillContent, err := os.ReadFile("../../kankuro_skills/mcp-cli/SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	config.SetBuiltinMCPCLISkillContent(skillContent)
	t.Cleanup(func() { config.SetBuiltinMCPCLISkillContent(nil) })
	service, repo, manager, workspace, workspaceID := newServiceFixture(t)
	manager.SetMCPEndpoint("http://127.0.0.1:12345/mcp")
	callerCard := saveCard(t, repo.FileRepository, config.IDCard{
		ID: "caller", Name: "orchestrator", Driver: "cat", WorkspaceID: workspaceID,
		DashboardControl: true, MCPIDs: []string{config.BuiltinMCPKankuro},
	})
	targetCard := saveCard(t, repo.FileRepository, config.IDCard{
		ID: "target", Name: "other project", Driver: "cat", WorkspaceID: workspaceID,
	})
	secondCard := saveCard(t, repo.FileRepository, config.IDCard{
		ID: "second", Name: "reviewer", Driver: "cat", WorkspaceID: workspaceID,
	})
	callerID, err := service.Launch(callerCard)
	if err != nil {
		t.Fatalf("launch caller: %v", err)
	}

	var workspaceMCP struct {
		Servers map[string]struct {
			Headers map[string]string `json:"headers"`
		} `json:"mcpServers"`
	}
	configData, err := os.ReadFile(filepath.Join(workspace, ".kankuro", "mcp-servers.json"))
	if err != nil {
		t.Fatal(err)
	}
	if decodeErr := json.Unmarshal(configData, &workspaceMCP); decodeErr != nil {
		t.Fatal(decodeErr)
	}
	token := strings.TrimPrefix(workspaceMCP.Servers["kankuro"].Headers["Authorization"], "Bearer ")
	if token == "" {
		t.Fatal("activation token missing from isolated workspace MCP config")
	}
	if cards, discoveryErr := service.ListCards(token); discoveryErr != nil || len(cards) == 0 {
		t.Fatalf("authenticated discovery failed: cards=%#v err=%v", cards, discoveryErr)
	}
	if sessions, discoveryErr := service.ListSessions(token); discoveryErr != nil || len(sessions) != 0 {
		t.Fatalf("authenticated session discovery failed: sessions=%#v err=%v", sessions, discoveryErr)
	}
	registrationOnly := saveCard(t, repo.FileRepository, config.IDCard{
		ID: "registration-only", Name: "registration only", Driver: "cat", WorkspaceID: workspaceID,
		MCPIDs: []string{config.BuiltinMCPKankuro},
	})
	if _, launchErr := service.Launch(registrationOnly); launchErr != nil {
		t.Fatalf("launch registration-only card: %v", launchErr)
	}
	configData, err = os.ReadFile(filepath.Join(workspace, ".kankuro", "mcp-servers.json"))
	if err != nil {
		t.Fatal(err)
	}
	if decodeErr := json.Unmarshal(configData, &workspaceMCP); decodeErr != nil {
		t.Fatal(decodeErr)
	}
	registrationToken := strings.TrimPrefix(workspaceMCP.Servers["kankuro"].Headers["Authorization"], "Bearer ")
	if _, authErr := manager.ResolveDashboardCaller(registrationToken); authErr == nil {
		t.Fatal("registration-only activation received dashboard control")
	}
	if _, authErr := service.ListCards(registrationToken); authErr == nil {
		t.Fatal("registration-only caller discovered cards")
	}
	if _, authErr := service.ListSessions(registrationToken); authErr == nil {
		t.Fatal("registration-only caller discovered sessions")
	}
	if _, authErr := service.GetSession(registrationToken, callerID); authErr == nil {
		t.Fatal("registration-only caller inspected a session")
	}

	scope := &testDashboardScope{
		sessions: map[string]string{callerID: "factory-a"},
		cards:    map[string]string{callerCard.ID: "factory-a", targetCard.ID: "factory-a", secondCard.ID: "factory-a"},
	}
	service.SetDashboardScopeResolver(scope)
	runner := NewOperationRunner(repo.Dir(), service)
	service.SetOperationRunner(runner)
	t.Cleanup(runner.Shutdown)
	requestID := uuid.NewString()
	type createResponse struct {
		operation OperationResult
		err       error
	}
	start := make(chan struct{})
	responses := make(chan createResponse, 2)
	for range 2 {
		go func() {
			<-start
			operation, createErr := service.CreateSession(context.Background(), token, requestID, targetCard.ID)
			responses <- createResponse{operation: operation, err: createErr}
		}()
	}
	close(start)
	firstResponse, secondResponse := <-responses, <-responses
	if firstResponse.err != nil || secondResponse.err != nil {
		t.Fatalf("concurrent create errors: %v, %v", firstResponse.err, secondResponse.err)
	}
	first, concurrentReplay := firstResponse.operation, secondResponse.operation
	if first.State != OperationSucceeded || first.TargetSession == "" {
		t.Fatalf("create outcome = %#v", first)
	}
	if concurrentReplay.ID != first.ID || concurrentReplay.TargetSession != first.TargetSession {
		t.Fatalf("concurrent duplicate created a second operation: %#v vs %#v", first, concurrentReplay)
	}
	name := "  仕事  "
	namedRequest := uuid.NewString()
	named, namedErr := service.CreateSession(context.Background(), token, namedRequest, targetCard.ID, &name)
	if namedErr != nil || named.State != OperationSucceeded || named.Result == nil ||
		named.Result.Inspection == nil || named.Result.Inspection.Name != "仕事" {
		t.Fatalf("named creation = %#v, %v", named, namedErr)
	}
	namedReplay, namedErr := service.CreateSession(context.Background(), token, namedRequest, targetCard.ID, &name)
	if namedErr != nil || namedReplay.TargetSession != named.TargetSession {
		t.Fatalf("named replay = %#v, %v", namedReplay, namedErr)
	}
	changedName := "仕事"
	if _, namedErr = service.CreateSession(context.Background(), token, namedRequest, targetCard.ID, &changedName); namedErr == nil {
		t.Fatal("changed raw name accepted")
	}
	if _, namedErr = service.CreateSession(context.Background(), token, namedRequest, targetCard.ID); namedErr == nil {
		t.Fatal("omitted name accepted for named retry")
	}
	scope.setSession(named.TargetSession, "other-dashboard")

	second, err := service.CreateSession(context.Background(), token, requestID, targetCard.ID)
	if err != nil {
		t.Fatalf("replay create: %v", err)
	}
	if second.ID != first.ID || second.TargetSession != first.TargetSession {
		t.Fatalf("duplicate request changed outcome: first=%#v second=%#v", first, second)
	}
	repo.cardsErr = errors.New("temporary card registry outage")
	registryFailureReplay, replayErr := service.CreateSession(context.Background(), token, requestID, targetCard.ID)
	repo.cardsErr = nil
	if replayErr != nil || registryFailureReplay.ID != first.ID || registryFailureReplay.State != OperationSucceeded {
		t.Fatalf("replay depended on mutable card registry: %#v, %v", registryFailureReplay, replayErr)
	}
	if _, conflictErr := service.CreateSession(context.Background(), token, requestID, callerCard.ID); conflictErr == nil {
		t.Fatal("same request ID with different arguments was accepted")
	}
	assertRejectedCode := func(name, wantCode string, request func() error) {
		t.Helper()
		requestErr := request()
		if requestErr == nil || SafeRejection(requestErr).Code != wantCode {
			t.Errorf("%s rejection = %v, want code %q", name, requestErr, wantCode)
		}
	}
	assertRejectedCode("malformed relaunch conversation", "invalid_request", func() error {
		_, requestErr := service.RelaunchSession(context.Background(), token, uuid.NewString(), "missing-target",
			&ConversationRequest{Mode: "resume"})
		return requestErr
	})
	assertRejectedCode("blank switch card", "invalid_request", func() error {
		_, requestErr := service.SwitchSessionCard(context.Background(), token, uuid.NewString(), "missing-target", " ", nil)
		return requestErr
	})
	assertRejectedCode("missing operation", "not_found", func() error {
		_, requestErr := service.GetOperation(token, uuid.NewString())
		return requestErr
	})
	assertRejectedCode("card registry unavailable", "registry_unavailable", func() error {
		repo.cardsErr = errors.New("registry offline")
		defer func() { repo.cardsErr = nil }()
		_, requestErr := service.CreateSession(context.Background(), token, uuid.NewString(), targetCard.ID)
		return requestErr
	})

	// A file in place of the journal directory deterministically rejects the
	// initial intent write without relying on filesystem permission behavior.
	blockedJournal := filepath.Join(t.TempDir(), "not-a-directory")
	if writeErr := os.WriteFile(blockedJournal, []byte("block"), 0o600); writeErr != nil {
		t.Fatal(writeErr)
	}
	originalJournalDir := runner.dir
	runner.dir = blockedJournal
	assertRejectedCode("journal intent failure", "journal_unavailable", func() error {
		_, requestErr := service.CreateSession(context.Background(), token, uuid.NewString(), targetCard.ID)
		return requestErr
	})
	runner.dir = originalJournalDir
	if _, ok := manager.GetSession(first.TargetSession); !ok {
		t.Fatal("created session was not published")
	}
	created, ok := manager.GetSession(first.TargetSession)
	if !ok {
		t.Fatal("created session disappeared")
	}
	if cards, discoveryErr := service.ListCards(token); discoveryErr != nil || len(cards) != 3 {
		t.Fatalf("scoped cards = %#v, %v", cards, discoveryErr)
	}
	if sessions, discoveryErr := service.ListSessions(token); discoveryErr != nil || len(sessions) != 1 {
		t.Fatalf("scoped sessions = %#v, %v", sessions, discoveryErr)
	}
	if _, inspectErr := service.GetSession(token, first.TargetSession); inspectErr != nil {
		t.Fatalf("scoped inspection: %v", inspectErr)
	}
	scope.setSession(first.TargetSession, "factory-b")
	if _, crossScopeErr := service.StopSession(context.Background(), token, uuid.NewString(), first.TargetSession); crossScopeErr == nil {
		t.Fatal("cross-dashboard stop was accepted")
	}
	if _, movedTargetReplayErr := service.CreateSession(
		context.Background(), token, requestID, targetCard.ID,
	); movedTargetReplayErr == nil {
		t.Fatal("create replay disclosed a target moved outside the caller's dashboard")
	}
	if _, lookupErr := service.GetOperation(token, first.ID); lookupErr == nil {
		t.Fatal("operation lookup disclosed a target moved outside the caller's dashboard")
	}
	if _, cancelErr := service.CancelOperation(token, first.ID); cancelErr == nil {
		t.Fatal("operation cancellation was allowed after the target moved")
	}
	scope.setSession(first.TargetSession, "factory-a")
	scope.setCard(targetCard.ID, "factory-b")
	if _, movedCardReplayErr := service.CreateSession(context.Background(), token, requestID, targetCard.ID); movedCardReplayErr == nil {
		t.Fatal("create replay disclosed a destination card moved outside the caller's dashboard")
	}
	if _, lookupErr := service.GetOperation(token, first.ID); lookupErr == nil {
		t.Fatal("operation lookup disclosed a destination card moved outside the caller's dashboard")
	}
	scope.setCard(targetCard.ID, "factory-a")
	handoff := filepath.Join(created.WorkingDir, "handoff.txt")
	if writeErr := os.WriteFile(handoff, []byte("caller prepared this file"), 0o600); writeErr != nil {
		t.Fatal(writeErr)
	}
	stopRequest := uuid.NewString()
	stopped, err := service.StopSession(context.Background(), token, stopRequest, first.TargetSession)
	if err != nil || stopped.State != OperationSucceeded {
		t.Fatalf("stop operation = %#v, %v", stopped, err)
	}
	switchRequest := uuid.NewString()
	switched, err := service.SwitchSessionCard(context.Background(), token, switchRequest, first.TargetSession, secondCard.ID, nil)
	if err != nil || switched.State != OperationSucceeded {
		t.Fatalf("switch operation = %#v, %v", switched, err)
	}
	current, ok := manager.GetSession(first.TargetSession)
	if !ok || current.CardID != secondCard.ID || current.WorkingDir != created.WorkingDir {
		t.Fatalf("switch did not preserve target workspace: %#v", current)
	}
	if content, readErr := os.ReadFile(handoff); readErr != nil || string(content) != "caller prepared this file" {
		t.Fatalf("handoff after switch = %q, %v", content, readErr)
	}
	stopAgain, err := service.StopSession(context.Background(), token, uuid.NewString(), first.TargetSession)
	if err != nil || stopAgain.State != OperationSucceeded {
		t.Fatalf("stop after switch = %#v, %v", stopAgain, err)
	}
	relaunched, err := service.RelaunchSession(context.Background(), token, uuid.NewString(), first.TargetSession, nil)
	if err != nil || relaunched.State != OperationSucceeded {
		t.Fatalf("relaunch operation = %#v, %v", relaunched, err)
	}
	stopFinal, err := service.StopSession(context.Background(), token, uuid.NewString(), first.TargetSession)
	if err != nil || stopFinal.State != OperationSucceeded {
		t.Fatalf("final stop = %#v, %v", stopFinal, err)
	}
	deleteRequest := uuid.NewString()
	deleted, err := service.DeleteSession(context.Background(), token, deleteRequest, first.TargetSession)
	if err != nil || deleted.State != OperationSucceeded {
		t.Fatalf("delete operation = %#v, %v", deleted, err)
	}
	deleteReplay, err := service.DeleteSession(context.Background(), token, deleteRequest, first.TargetSession)
	if err != nil || deleteReplay.ID != deleted.ID {
		t.Fatalf("delete replay = %#v, %v", deleteReplay, err)
	}
	if _, missingErr := service.DeleteSession(context.Background(), token, uuid.NewString(), first.TargetSession); missingErr == nil {
		t.Fatal("a new delete request for an absent session was accepted")
	}
	runner.Shutdown()
	recoveredRunner := NewOperationRunner(repo.Dir(), service)
	service.SetOperationRunner(recoveredRunner)
	t.Cleanup(recoveredRunner.Shutdown)
	recovered, err := service.CreateSession(context.Background(), token, requestID, targetCard.ID)
	if err != nil || recovered.ID != first.ID || recovered.TargetSession != first.TargetSession {
		t.Fatalf("restart replay = %#v, %v; want original outcome %#v", recovered, err, first)
	}

	started := make(chan struct{})
	completed := make(chan Operation, 1)
	cancelRequestID := uuid.NewString()
	go func() {
		operation, submitErr := recoveredRunner.Submit(context.Background(), token, cancelRequestID,
			"cancel_test", "", struct{}{}, func(operationCtx context.Context, _ Operation) (json.RawMessage, error) {
				close(started)
				<-operationCtx.Done()
				return nil, operationCtx.Err()
			})
		if submitErr != nil {
			t.Errorf("submit cancellable operation: %v", submitErr)
		}
		completed <- operation
	}()
	<-started
	cancelKey := operationKey{caller: callerID, request: cancelRequestID}
	recoveredRunner.mu.Lock()
	operationID := recoveredRunner.records[cancelKey].ID
	recoveredRunner.mu.Unlock()
	if _, cancelErr := recoveredRunner.Cancel(token, operationID); cancelErr != nil {
		t.Fatalf("cancel operation: %v", cancelErr)
	}
	select {
	case outcome := <-completed:
		if outcome.State != OperationCancelled {
			t.Fatalf("cancel outcome = %s, want cancelled", outcome.State)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("cancelled operation did not finish")
	}
}

//nolint:govet // The test follows a single journaled effect through failure, replay, and restart.
func TestFinalJournalWriteFailureReturnsOutcomeAndRecoversInterrupted(t *testing.T) {
	skillContent, err := os.ReadFile("../../kankuro_skills/mcp-cli/SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	config.SetBuiltinMCPCLISkillContent(skillContent)
	t.Cleanup(func() { config.SetBuiltinMCPCLISkillContent(nil) })
	service, repo, manager, workspace, workspaceID := newServiceFixture(t)
	manager.SetMCPEndpoint("http://127.0.0.1:12345/mcp")
	callerCard := saveCard(t, repo.FileRepository, config.IDCard{
		ID: "controller", Name: "controller", Driver: "cat", WorkspaceID: workspaceID,
		DashboardControl: true, MCPIDs: []string{config.BuiltinMCPKankuro},
	})
	targetCard := saveCard(t, repo.FileRepository, config.IDCard{
		ID: "target", Name: "target", Driver: "cat", WorkspaceID: workspaceID,
	})
	if _, err := service.Launch(callerCard); err != nil {
		t.Fatalf("launch controller: %v", err)
	}
	var workspaceMCP struct {
		Servers map[string]struct {
			Headers map[string]string `json:"headers"`
		} `json:"mcpServers"`
	}
	configData, err := os.ReadFile(filepath.Join(workspace, ".kankuro", "mcp-servers.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(configData, &workspaceMCP); err != nil {
		t.Fatal(err)
	}
	token := strings.TrimPrefix(workspaceMCP.Servers["kankuro"].Headers["Authorization"], "Bearer ")
	runner := NewOperationRunner(repo.Dir(), service)
	service.SetOperationRunner(runner)
	var finalWriteFailure sync.Once
	runner.writeFileAtomic = func(path string, data []byte) error {
		var stored Operation
		if err := json.Unmarshal(data, &stored); err != nil {
			return err
		}
		if stored.State == OperationSucceeded {
			var writeErr error
			finalWriteFailure.Do(func() { writeErr = errors.New("injected final journal write failure") })
			if writeErr != nil {
				return writeErr
			}
		}
		return os.WriteFile(path, data, 0o600)
	}
	requestID := uuid.NewString()
	first, err := service.CreateSession(context.Background(), token, requestID, targetCard.ID)
	if err != nil {
		t.Fatalf("create with final journal failure: %v", err)
	}
	if first.State != OperationSucceededWithWarning || first.Warning == "" || first.TargetSession == "" ||
		first.Result == nil || first.Result.SessionID != first.TargetSession {
		t.Fatalf("observed create outcome was lost on journal failure: %#v", first)
	}
	replayed, err := service.CreateSession(context.Background(), token, requestID, targetCard.ID)
	if err != nil || replayed.ID != first.ID || replayed.TargetSession != first.TargetSession ||
		replayed.Result == nil || replayed.Result.SessionID != first.TargetSession {
		t.Fatalf("in-memory replay = %#v, %v", replayed, err)
	}
	if len(manager.List()) != 2 {
		t.Fatalf("replay repeated lifecycle effect: sessions=%d, want controller+target", len(manager.List()))
	}
	runner.Shutdown()
	recoveredRunner := NewOperationRunner(repo.Dir(), service)
	service.SetOperationRunner(recoveredRunner)
	t.Cleanup(recoveredRunner.Shutdown)
	recovered, err := service.GetOperation(token, first.ID)
	if err != nil || recovered.State != OperationInterrupted || recovered.TargetSession != first.TargetSession {
		t.Fatalf("restart recovery = %#v, %v; want interrupted original target", recovered, err)
	}
	if len(manager.List()) != 2 {
		t.Fatalf("interrupted journal entry was auto-executed: sessions=%d", len(manager.List()))
	}
}

//nolint:gocognit,gocyclo,cyclop,govet // The test checks cancellation, caller revocation, cleanup, and reservation reuse in one fixture.
func TestRunnerCancellationDuringTemporaryWorkspaceCopyReleasesActivation(t *testing.T) {
	skillContent, skillErr := os.ReadFile("../../kankuro_skills/mcp-cli/SKILL.md")
	if skillErr != nil {
		t.Fatal(skillErr)
	}
	config.SetBuiltinMCPCLISkillContent(skillContent)
	t.Cleanup(func() { config.SetBuiltinMCPCLISkillContent(nil) })
	t.Setenv("KANKURO_TMP_DIR", t.TempDir())
	copyEntered := make(chan struct{}, 2)
	var copyCalls atomic.Int32
	copier := session.WorkspaceCopier(func(ctx context.Context, _, destination string) error {
		if copyCalls.Add(1) <= 2 {
			copyEntered <- struct{}{}
			<-ctx.Done()
			return ctx.Err()
		}
		return os.MkdirAll(destination, 0o700)
	})
	service, repo, manager, workspace, workspaceID := newServiceFixture(t, session.WithWorkspaceCopier(copier))
	uiService := New(repo, manager)
	manager.SetMCPEndpoint("http://127.0.0.1:12345/mcp")
	callerCard := saveCard(t, repo.FileRepository, config.IDCard{
		ID: "controller", Name: "controller", Driver: "cat", WorkspaceID: workspaceID,
		DashboardControl: true, MCPIDs: []string{config.BuiltinMCPKankuro},
	})
	targetCard := saveCard(t, repo.FileRepository, config.IDCard{
		ID: "target", Name: "target", Driver: "cat", WorkspaceID: workspaceID, UseTmpWorkspace: true,
	})
	if _, err := service.Launch(callerCard); err != nil {
		t.Fatalf("launch controller: %v", err)
	}
	var mcpConfig struct {
		Servers map[string]struct {
			Headers map[string]string `json:"headers"`
		} `json:"mcpServers"`
	}
	data, err := os.ReadFile(filepath.Join(workspace, ".kankuro", "mcp-servers.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &mcpConfig); err != nil {
		t.Fatal(err)
	}
	token := strings.TrimPrefix(mcpConfig.Servers["kankuro"].Headers["Authorization"], "Bearer ")
	runner := NewOperationRunner(repo.Dir(), service)
	service.SetOperationRunner(runner)
	t.Cleanup(runner.Shutdown)
	requestID := uuid.NewString()
	type createReply struct {
		result OperationResult
		err    error
	}
	replies := make(chan createReply, 1)
	go func() {
		result, createErr := service.CreateSession(context.Background(), token, requestID, targetCard.ID)
		replies <- createReply{result: result, err: createErr}
	}()
	select {
	case <-copyEntered:
	case <-time.After(time.Second):
		t.Fatal("workspace copy did not reach the cancellation gate")
	}
	if deleteErr := uiService.DeleteCard(targetCard.ID); deleteErr == nil {
		t.Fatal("delete card succeeded while MCP create was preparing")
	}
	// Resolve the accepted operation through its request identity without racing
	// on runner internals after the copy gate has signalled admission.
	caller, err := service.ResolveCaller(token)
	if err != nil {
		t.Fatal(err)
	}
	key := operationKey{caller: caller.SessionID, request: requestID}
	runner.mu.Lock()
	operation := cloneOperation(runner.records[key])
	runner.mu.Unlock()
	if operation.ID == "" {
		t.Fatal("copy gate was reached without a journaled operation")
	}
	if _, err := service.CancelOperation(token, operation.ID); err != nil {
		t.Fatalf("cancel during workspace copy: %v", err)
	}
	var cancelled createReply
	select {
	case cancelled = <-replies:
	case <-time.After(time.Second):
		t.Fatal("cancelled workspace-copy operation did not return promptly")
	}
	if cancelled.err != nil || cancelled.result.State != OperationCancelled ||
		cancelled.result.TargetSession != operation.TargetSession {
		t.Fatalf("cancelled create = %#v, %v", cancelled.result, cancelled.err)
	}
	if _, exists := manager.GetSession(operation.TargetSession); exists {
		t.Fatal("cancelled activation was published")
	}
	if _, statErr := os.Stat(filepath.Join(os.Getenv("KANKURO_TMP_DIR"), operation.TargetSession)); !os.IsNotExist(statErr) {
		t.Fatalf("cancelled activation left a temporary copy: %v", statErr)
	}
	secondRequestID := uuid.NewString()
	secondReplies := make(chan createReply, 1)
	go func() {
		result, createErr := service.CreateSession(context.Background(), token, secondRequestID, targetCard.ID)
		secondReplies <- createReply{result: result, err: createErr}
	}()
	select {
	case <-copyEntered:
	case <-time.After(time.Second):
		t.Fatal("second workspace copy did not reach caller-revocation gate")
	}
	secondKey := operationKey{caller: caller.SessionID, request: secondRequestID}
	runner.mu.Lock()
	secondOperation := cloneOperation(runner.records[secondKey])
	runner.mu.Unlock()
	if err := manager.Kill(caller.SessionID); err != nil {
		t.Fatalf("revoke controller activation during preparation: %v", err)
	}
	var revokedReply createReply
	select {
	case revokedReply = <-secondReplies:
	case <-time.After(time.Second):
		t.Fatal("caller-revoked preparation did not return promptly")
	}
	if revokedReply.err != nil || revokedReply.result.State != OperationCancelled || secondOperation.ID == "" ||
		revokedReply.result.TargetSession != secondOperation.TargetSession {
		t.Fatalf("caller revocation outcome = %#v, %v", revokedReply.result, revokedReply.err)
	}
	if _, exists := manager.GetSession(secondOperation.TargetSession); exists {
		t.Fatal("caller-revoked activation was published")
	}
	if _, statErr := os.Stat(filepath.Join(os.Getenv("KANKURO_TMP_DIR"), secondOperation.TargetSession)); !os.IsNotExist(statErr) {
		t.Fatalf("caller-revoked activation left a temporary copy: %v", statErr)
	}
	if _, retryErr := manager.LaunchWithIDContext(
		context.Background(), operation.TargetSession, targetCard, nil, nil, nil,
	); retryErr != nil {
		t.Fatalf("activation reservation was not released after cancellation: %v", retryErr)
	}
	if err := manager.Kill(operation.TargetSession); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.RemoveDetailed(operation.TargetSession); err != nil {
		t.Fatalf("remove retry session after reservation check: %v", err)
	}
	if err := uiService.DeleteCard(targetCard.ID); err != nil {
		t.Fatalf("delete card after cancelled preparations and cleanup: %v", err)
	}
}

func TestPendingCardReferenceBlocksCrossServiceDeleteDuringMCPCreate(t *testing.T) {
	setBuiltinMCPCLISkillForTest(t)
	allowCopy := make(chan struct{})
	copyEntered := make(chan struct{})
	var gateOnce sync.Once
	copier := session.WorkspaceCopier(func(ctx context.Context, _, destination string) error {
		gateOnce.Do(func() {
			close(copyEntered)
			select {
			case <-allowCopy:
			case <-ctx.Done():
			}
		})
		if err := ctx.Err(); err != nil {
			return err
		}
		return os.MkdirAll(destination, 0o700)
	})
	service, repo, manager, workspace, workspaceID := newServiceFixture(t, session.WithWorkspaceCopier(copier))
	uiService := New(repo, manager) // A second frontend service shares the manager and repository.
	manager.SetMCPEndpoint("http://127.0.0.1:12345/mcp")
	callerCard := saveCard(t, repo.FileRepository, config.IDCard{
		ID: "create-controller", Name: "controller", Driver: "cat", WorkspaceID: workspaceID,
		DashboardControl: true, MCPIDs: []string{config.BuiltinMCPKankuro},
	})
	targetCard := saveCard(t, repo.FileRepository, config.IDCard{
		ID: "copy-target", Name: "temporary target", Driver: "cat", WorkspaceID: workspaceID, UseTmpWorkspace: true,
	})
	if _, err := service.Launch(callerCard); err != nil {
		t.Fatalf("launch controller: %v", err)
	}
	token := workspaceMCPToken(t, workspace)
	runner := NewOperationRunner(repo.Dir(), service)
	service.SetOperationRunner(runner)
	t.Cleanup(runner.Shutdown)
	type reply struct {
		result OperationResult
		err    error
	}
	replies := make(chan reply, 1)
	requestID := uuid.NewString()
	go func() {
		result, err := service.CreateSession(context.Background(), token, requestID, targetCard.ID)
		replies <- reply{result: result, err: err}
	}()
	select {
	case <-copyEntered:
	case <-time.After(2 * time.Second):
		t.Fatal("MCP create did not reach temporary-copy gate")
	}
	if err := uiService.DeleteCard(targetCard.ID); err == nil || !strings.Contains(err.Error(), "activation(s) in progress") {
		t.Fatalf("cross-service deletion during create = %v, want pending-reference refusal", err)
	}
	close(allowCopy)
	var created reply
	select {
	case created = <-replies:
	case <-time.After(5 * time.Second):
		t.Fatal("MCP create did not complete after copy gate release")
	}
	if created.err != nil || created.result.State != OperationSucceeded || created.result.Result == nil {
		t.Fatalf("MCP create result = %#v, %v", created.result, created.err)
	}
	sessionID := created.result.Result.SessionID
	if err := manager.Kill(sessionID); err != nil {
		t.Fatalf("stop created session: %v", err)
	}
	if err := uiService.Relaunch(sessionID); err != nil {
		t.Fatalf("relaunch with retained card: %v", err)
	}
	if err := manager.Kill(sessionID); err != nil {
		t.Fatalf("stop relaunched session: %v", err)
	}
	if err := uiService.DeleteCard(targetCard.ID); err == nil {
		t.Fatal("delete succeeded after session publication and termination")
	}
}

type instructionGate struct {
	entered chan struct{}
	release chan struct{}
	once    sync.Once
}

type gatedInstructionDriver struct {
	driver.AgentDriver

	destination string
	gate        *instructionGate
}

func (d *gatedInstructionDriver) SetAgentsMD(
	card config.IDCard,
	personas []config.Persona,
	mcps []config.MCPServer,
	workspace string,
	mode driver.InstructionSyncMode,
) (bool, error) {
	managed, err := d.AgentDriver.SetAgentsMD(card, personas, mcps, workspace, mode)
	if err != nil || card.ID != d.destination {
		return managed, err
	}
	d.gate.once.Do(func() { close(d.gate.entered) })
	<-d.gate.release
	return managed, nil
}

func TestPendingCardReferenceBlocksCrossServiceDeleteDuringMCPSwitch(t *testing.T) {
	setBuiltinMCPCLISkillForTest(t)
	gate := &instructionGate{entered: make(chan struct{}), release: make(chan struct{})}
	registry := driver.NewBuiltinRegistry()
	registry.Register("cat", func() driver.AgentDriver {
		return &gatedInstructionDriver{
			AgentDriver: driver.NewGenericCLIDriver(), destination: "switch-destination", gate: gate,
		}
	})
	service, repo, manager, workspace, workspaceID := newServiceFixture(t, session.WithDriverRegistry(registry))
	uiService := New(repo, manager)
	manager.SetMCPEndpoint("http://127.0.0.1:12345/mcp")
	callerCard := saveCard(t, repo.FileRepository, config.IDCard{
		ID: "switch-controller", Name: "controller", Driver: "cat", WorkspaceID: workspaceID,
		DashboardControl: true, MCPIDs: []string{config.BuiltinMCPKankuro},
	})
	sourceCard := saveCard(t, repo.FileRepository, config.IDCard{
		ID: "switch-source", Name: "source", Driver: "cat", WorkspaceID: workspaceID, UseTmpWorkspace: true,
	})
	destination := saveCard(t, repo.FileRepository, config.IDCard{
		ID: "switch-destination", Name: "destination", Driver: "cat", WorkspaceID: workspaceID,
	})
	if _, err := service.Launch(callerCard); err != nil {
		t.Fatalf("launch controller: %v", err)
	}
	token := workspaceMCPToken(t, workspace)
	sourceID, launchErr := service.Launch(sourceCard)
	if launchErr != nil {
		t.Fatalf("launch source: %v", launchErr)
	}
	if err := manager.Kill(sourceID); err != nil {
		t.Fatalf("stop source: %v", err)
	}
	if bindings, err := service.Conversations(sourceID, destination.ID); err != nil || len(bindings) != 0 {
		t.Fatalf("destination should have no saved conversation binding before switch: %v, %v", bindings, err)
	}
	runner := NewOperationRunner(repo.Dir(), service)
	service.SetOperationRunner(runner)
	t.Cleanup(runner.Shutdown)
	type reply struct {
		result OperationResult
		err    error
	}
	replies := make(chan reply, 1)
	go func() {
		result, switchErr := service.SwitchSessionCard(
			context.Background(), token, uuid.NewString(), sourceID, destination.ID, nil,
		)
		replies <- reply{result: result, err: switchErr}
	}()
	select {
	case <-gate.entered:
	case <-time.After(2 * time.Second):
		t.Fatal("MCP switch did not reach destination preparation gate")
	}
	if err := uiService.DeleteCard(destination.ID); err == nil || !strings.Contains(err.Error(), "activation(s) in progress") {
		t.Fatalf("cross-service deletion during switch = %v, want pending-reference refusal", err)
	}
	close(gate.release)
	var switched reply
	select {
	case switched = <-replies:
	case <-time.After(5 * time.Second):
		t.Fatal("MCP switch did not complete after gate release")
	}
	if switched.err != nil || switched.result.State != OperationSucceeded {
		t.Fatalf("MCP switch result = %#v, %v", switched.result, switched.err)
	}
	snapshot, found := manager.GetSession(sourceID)
	if !found || snapshot.CardID != destination.ID {
		t.Fatalf("switch did not publish destination reference: %#v, found=%v", snapshot, found)
	}
	if err := manager.Kill(sourceID); err != nil {
		t.Fatalf("stop switched session: %v", err)
	}
	if err := uiService.DeleteCard(destination.ID); err == nil {
		t.Fatal("delete succeeded after switch publication and termination")
	}
}

func TestDeletionWinningBeforeAdmissionRejectsLoadedCardSnapshot(t *testing.T) {
	service, repo, manager, _, workspaceID := newServiceFixture(t)
	staleSnapshot := saveCard(t, repo.FileRepository, config.IDCard{
		ID: "deleted-before-launch", Name: "stale", Driver: "cat", WorkspaceID: workspaceID,
	})
	if err := service.DeleteCard(staleSnapshot.ID); err != nil {
		t.Fatalf("delete unreferenced card: %v", err)
	}
	if _, err := service.Launch(staleSnapshot); err == nil || !strings.Contains(err.Error(), "card not found") {
		t.Fatalf("launch with stale loaded snapshot = %v, want missing-card refusal", err)
	}
	if sessions := manager.List(); len(sessions) != 0 {
		t.Fatalf("stale snapshot published orphan session: %#v", sessions)
	}
}

func workspaceMCPToken(t *testing.T, workspace string) string {
	t.Helper()
	data, readErr := os.ReadFile(filepath.Join(workspace, ".kankuro", "mcp-servers.json"))
	if readErr != nil {
		t.Fatal(readErr)
	}
	var mcpConfig struct {
		Servers map[string]struct {
			Headers map[string]string `json:"headers"`
		} `json:"mcpServers"`
	}
	if err := json.Unmarshal(data, &mcpConfig); err != nil {
		t.Fatal(err)
	}
	token := strings.TrimPrefix(mcpConfig.Servers["kankuro"].Headers["Authorization"], "Bearer ")
	if token == "" {
		t.Fatal("controller activation token is missing")
	}
	return token
}

func setBuiltinMCPCLISkillForTest(t *testing.T) {
	t.Helper()
	content, err := os.ReadFile("../../kankuro_skills/mcp-cli/SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	config.SetBuiltinMCPCLISkillContent(content)
	t.Cleanup(func() { config.SetBuiltinMCPCLISkillContent(nil) })
}

//nolint:gocognit,gocyclo,cyclop,govet // The test exercises queue, disconnect, completion, and late-cancel boundaries together.
func TestOperationQueueBoundAndDisconnectedWaitReturnPromptly(t *testing.T) {
	skillContent, skillErr := os.ReadFile("../../kankuro_skills/mcp-cli/SKILL.md")
	if skillErr != nil {
		t.Fatal(skillErr)
	}
	config.SetBuiltinMCPCLISkillContent(skillContent)
	t.Cleanup(func() { config.SetBuiltinMCPCLISkillContent(nil) })
	service, repo, manager, workspace, workspaceID := newServiceFixture(t)
	manager.SetMCPEndpoint("http://127.0.0.1:12345/mcp")
	card := saveCard(t, repo.FileRepository, config.IDCard{
		ID: "controller", Name: "controller", Driver: "cat", WorkspaceID: workspaceID,
		DashboardControl: true, MCPIDs: []string{config.BuiltinMCPKankuro},
	})
	if _, err := service.Launch(card); err != nil {
		t.Fatalf("launch controller: %v", err)
	}
	var mcpConfig struct {
		Servers map[string]struct {
			Headers map[string]string `json:"headers"`
		} `json:"mcpServers"`
	}
	data, err := os.ReadFile(filepath.Join(workspace, ".kankuro", "mcp-servers.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &mcpConfig); err != nil {
		t.Fatal(err)
	}
	token := strings.TrimPrefix(mcpConfig.Servers["kankuro"].Headers["Authorization"], "Bearer ")
	runner := NewOperationRunner(repo.Dir(), service)
	t.Cleanup(runner.Shutdown)
	releaseWorkers := make(chan struct{})
	var releaseWorkersOnce sync.Once
	releaseAllWorkers := func() { releaseWorkersOnce.Do(func() { close(releaseWorkers) }) }
	t.Cleanup(releaseAllWorkers)
	workerEntered := make(chan struct{}, operationWorkers)
	workerDone := make(chan struct{}, operationWorkers)
	disconnectWaits := make([]context.CancelFunc, 0, operationWorkers)
	type submitResult struct {
		operation Operation
		err       error
	}
	results := make(chan submitResult, operationWorkers)
	for range operationWorkers {
		ctx, cancel := context.WithCancel(context.Background())
		disconnectWaits = append(disconnectWaits, cancel)
		requestID := uuid.NewString()
		go func(ctx context.Context, cancel context.CancelFunc, requestID string) {
			operation, submitErr := runner.Submit(ctx, token, requestID, "gated_operation", "", struct{}{},
				func(context.Context, Operation) (json.RawMessage, error) {
					workerEntered <- struct{}{}
					<-releaseWorkers
					workerDone <- struct{}{}
					return json.RawMessage(`{}`), nil
				})
			results <- submitResult{operation: operation, err: submitErr}
			cancel()
		}(ctx, cancel, requestID)
	}
	for range operationWorkers {
		select {
		case <-workerEntered:
		case <-time.After(time.Second):
			t.Fatal("all operation workers did not enter the controlled gate")
		}
	}
	for _, cancel := range disconnectWaits {
		cancel()
	}
	var disconnected []Operation
	for range operationWorkers {
		select {
		case result := <-results:
			if result.err != nil || (result.operation.State != OperationPending && result.operation.State != OperationRunning) {
				t.Fatalf("disconnected wait result = %#v, %v", result.operation, result.err)
			}
			disconnected = append(disconnected, result.operation)
		case <-time.After(time.Second):
			t.Fatal("disconnected request did not return promptly")
		}
	}
	for range operationQueueCapacity {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		operation, submitErr := runner.Submit(ctx, token, uuid.NewString(), "queued_operation", "", struct{}{},
			func(context.Context, Operation) (json.RawMessage, error) {
				return json.RawMessage(`{}`), nil
			})
		if submitErr != nil || operation.State != OperationPending {
			t.Fatalf("queued admission = %#v, %v", operation, submitErr)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	submitOverCapacity := func() error {
		_, submitErr := runner.Submit(ctx, token, uuid.NewString(), "over_capacity", "", struct{}{},
			func(context.Context, Operation) (json.RawMessage, error) { return nil, nil })
		return submitErr
	}
	if submitErr := submitOverCapacity(); submitErr == nil || SafeRejection(submitErr).Code != "capacity_exhausted" {
		t.Fatalf("full operation queue rejection = %v, want capacity_exhausted", submitErr)
	}
	releaseAllWorkers()
	for range operationWorkers {
		select {
		case <-workerDone:
		case <-time.After(time.Second):
			t.Fatal("released operation worker did not finish")
		}
	}
	for _, operation := range disconnected {
		completed, waitErr := runner.wait(context.Background(), operation.ID)
		if waitErr != nil || completed.State != OperationSucceeded {
			t.Fatalf("operation after disconnected wait = %#v, %v", completed, waitErr)
		}
	}
	// Terminal publication can precede a worker's return to its queue loop.
	// Drain and join this runner before starting the independently gated phase,
	// so leftover capacity-phase queue entries cannot consume its admission slot.
	runner.Shutdown()
	runner = NewOperationRunner(repo.Dir(), service)
	t.Cleanup(runner.Shutdown)
	publicationEntered, releasePublication := make(chan struct{}), make(chan struct{})
	var releasePublicationOnce sync.Once
	releasePublishedEffect := func() { releasePublicationOnce.Do(func() { close(releasePublication) }) }
	t.Cleanup(releasePublishedEffect)
	requestID := uuid.NewString()
	completedOperation := make(chan Operation, 1)
	go func() {
		operation, submitErr := runner.Submit(context.Background(), token, requestID, "publication_test", "", struct{}{},
			func(context.Context, Operation) (json.RawMessage, error) {
				close(publicationEntered)
				<-releasePublication
				return json.RawMessage(`{"published":true}`), nil
			})
		if submitErr != nil {
			t.Errorf("submit publication-gated operation: %v", submitErr)
		}
		completedOperation <- operation
	}()
	select {
	case <-publicationEntered:
	case <-time.After(time.Second):
		t.Fatal("operation did not reach the publication gate")
	}
	caller, err := service.ResolveCaller(token)
	if err != nil {
		t.Fatal(err)
	}
	publicationKey := operationKey{caller: caller.SessionID, request: requestID}
	runner.mu.Lock()
	publicationID := runner.records[publicationKey].ID
	runner.mu.Unlock()
	if _, err := runner.Cancel(token, publicationID); err != nil {
		t.Fatalf("cancel at publication boundary: %v", err)
	}
	releasePublishedEffect()
	select {
	case published := <-completedOperation:
		if published.State != OperationSucceededWithWarning ||
			!strings.Contains(published.Warning, "cancellation arrived too late") ||
			string(published.Result) != `{"published":true}` {
			t.Fatalf("published success after cancellation = %#v", published)
		}
	case <-time.After(time.Second):
		t.Fatal("publication-gated operation did not finish")
	}
	runner.Shutdown()
}

func TestOperationJournalRecoversInterruptedAndFailsClosedOnCorruption(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "operations")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	requestID, operationID, callerID := uuid.NewString(), uuid.NewString(), uuid.NewString()
	interrupted := Operation{ID: operationID, CallerSession: callerID, DashboardID: DefaultDashboardID, RequestID: requestID,
		Kind: "stop_session", State: OperationRunning, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	data, err := json.Marshal(interrupted)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, callerID+"-"+requestID+".json")
	if writeErr := os.WriteFile(path, data, 0o600); writeErr != nil {
		t.Fatal(writeErr)
	}
	runner := NewOperationRunner(filepath.Dir(dir), nil)
	key := operationKey{caller: callerID, request: requestID}
	if got := runner.records[key]; got == nil || got.State != OperationInterrupted || got.Warning == "" {
		t.Fatalf("recovered operation = %#v, want interrupted with reconciliation warning", got)
	}
	runner.Shutdown()

	if writeErr := os.WriteFile(filepath.Join(dir, "corrupt.json"), []byte("{"), 0o600); writeErr != nil {
		t.Fatal(writeErr)
	}
	corruptRunner := NewOperationRunner(filepath.Dir(dir), nil)
	defer corruptRunner.Shutdown()
	if corruptRunner.loadErr == nil {
		t.Fatal("corrupt journal did not disable mutation admission")
	}
}

//nolint:govet // short declarations localize setup failures within the registry fixture.
func TestResolveCapabilitiesOrdersAndFiltersEveryRegistry(t *testing.T) {
	base, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	repo := &testRepository{FileRepository: base}
	for _, name := range []string{"skill-a", "skill-b", "skill-c", "skill-unused"} {
		if _, err := base.UpsertSkill(config.Skill{ID: name, Name: name}); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"persona-a", "persona-b", "persona-unused"} {
		if _, err := base.UpsertPersona(config.Persona{ID: name, Name: name}); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"mcp-a", "mcp-b", "mcp-unused"} {
		if _, err := base.UpsertMCP(config.MCPServer{ID: name, Name: name, Command: "mcp"}); err != nil {
			t.Fatal(err)
		}
	}
	bundle, err := New(repo, nil).resolveCapabilities(config.IDCard{
		SkillIDs:   []string{"skill-c", "skill-a", "skill-c", "missing"},
		PersonaIDs: []string{"persona-b", "persona-a", "persona-b", "missing"},
		MCPIDs:     []string{"mcp-b", "mcp-a", "mcp-b", "missing"},
	})
	if err != nil {
		t.Fatal(err)
	}
	assertIDs(t, idsSkills(bundle.Skills), []string{"skill-a", "skill-c"})
	assertIDs(t, idsPersonas(bundle.Personas), []string{"persona-a", "persona-b"})
	assertIDs(t, idsMCPs(bundle.MCPs), []string{"mcp-a", "mcp-b"})
	if repo.skillReads != 1 || repo.personaReads != 1 || repo.mcpReads != 1 {
		t.Fatalf("reads skills/personas/mcps = %d/%d/%d; want one each", repo.skillReads, repo.personaReads, repo.mcpReads)
	}
}

func TestResolveCapabilitiesAddsMCPAccessSkillOnlyWhenKankuroMCPSelected(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	service := New(repo, nil)
	withoutMCP, err := service.resolveCapabilities(config.IDCard{})
	if err != nil {
		t.Fatal(err)
	}
	if countMCPCLI(withoutMCP.Skills) > 0 {
		t.Fatal("mcp-cli access skill was added without selecting Kankuro MCP")
	}
	withMCP, err := service.resolveCapabilities(config.IDCard{MCPIDs: []string{config.BuiltinMCPKankuro}})
	if err != nil {
		t.Fatal(err)
	}
	if countMCPCLI(withMCP.Skills) != 1 || countKankuroPuppet(withMCP.Skills) != 1 {
		t.Fatalf("Kankuro protocol skills = %#v, want one mcp-cli and one kankuro-puppet", withMCP.Skills)
	}
	withMCPAndSkill, err := service.resolveCapabilities(config.IDCard{
		MCPIDs: []string{config.BuiltinMCPKankuro}, SkillIDs: []string{config.BuiltinSkillMCPCLI},
	})
	if err != nil {
		t.Fatal(err)
	}
	if countMCPCLI(withMCPAndSkill.Skills) != 1 || countKankuroPuppet(withMCPAndSkill.Skills) != 1 {
		t.Fatalf("explicit mcp-cli selection duplicated or puppet skill missing: %#v", withMCPAndSkill.Skills)
	}
}

func countKankuroPuppet(skills []config.Skill) int {
	count := 0
	for _, skill := range skills {
		if skill.ID == config.BuiltinSkillKankuroPuppet {
			count++
		}
	}
	return count
}

func countMCPCLI(skills []config.Skill) int {
	count := 0
	for _, skill := range skills {
		if skill.ID == config.BuiltinSkillMCPCLI {
			count++
		}
	}
	return count
}

//nolint:gocognit // one table-driven scenario validates each abort point and read count.
func TestResolveCapabilitiesPropagatesEachRegistryFailure(t *testing.T) {
	for _, tc := range []struct {
		name, context string
		registry      string
	}{
		{"skills", "list skills:", "skills"}, {"personas", "list personas:", "personas"}, {"mcps", "list mcps:", "mcps"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			api, repo, sessions, _, workspaceID := newServiceFixture(t)
			saved := saveCard(t, repo.FileRepository, config.IDCard{
				ID: "card", Name: "card", Driver: "cat", WorkspaceID: workspaceID,
			})
			want := errors.New("injected " + tc.name + " error")
			switch tc.registry {
			case "skills":
				repo.skillsErr = want
			case "personas":
				repo.personasErr = want
			case "mcps":
				repo.mcpsErr = want
			}
			_, err := api.Launch(saved) // empty selections still read each registry.
			if !errors.Is(err, want) || err == nil || !strings.HasPrefix(err.Error(), tc.context) {
				t.Fatalf("resolve error = %v, want wrapped %q failure", err, tc.context)
			}
			wantReads := map[string]int{"skills": 1, "personas": 1, "mcps": 1}
			if tc.registry == "skills" {
				wantReads["personas"], wantReads["mcps"] = 0, 0
			}
			if tc.registry == "personas" {
				wantReads["mcps"] = 0
			}
			if repo.skillReads != wantReads["skills"] || repo.personaReads != wantReads["personas"] || repo.mcpReads != wantReads["mcps"] {
				t.Fatalf("unexpected read counts: skills=%d personas=%d mcps=%d", repo.skillReads, repo.personaReads, repo.mcpReads)
			}
			if got := sessions.List(); len(got) != 0 {
				t.Fatalf("activation spawned despite %s registry failure: %#v", tc.name, got)
			}
		})
	}
}

func TestResolveCardDistinguishesRegistryFailureAndMissingCard(t *testing.T) {
	base, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	repo := &testRepository{FileRepository: base}
	_, err = New(repo, nil).resolveCard("absent")
	if err == nil || err.Error() != "card not found: absent" {
		t.Fatalf("missing-card error = %v", err)
	}
	want := errors.New("unreadable cards")
	repo.cardsErr = want
	_, err = New(repo, nil).resolveCard("absent")
	if !errors.Is(err, want) || err == nil || !strings.HasPrefix(err.Error(), "list cards:") {
		t.Fatalf("registry error = %v", err)
	}
}

func TestServiceLaunchUsesSuppliedCardSnapshot(t *testing.T) {
	api, repo, sessions, workspace, workspaceID := newServiceFixture(t)
	saved := saveCard(t, repo.FileRepository, config.IDCard{ID: "stored", Name: "stored-card", Driver: "cat", WorkspaceID: workspaceID})
	selected := saved
	selected.Name = "selected-snapshot"
	sid, err := api.Launch(selected)
	if err != nil {
		t.Fatalf("Launch: %v", err)
	}
	snapshot, ok := sessions.GetSession(sid)
	if !ok {
		t.Fatal("launched session missing from manager")
	}
	if snapshot.CardID != selected.ID || snapshot.Name != selected.Name || snapshot.WorkingDir != workspace {
		t.Fatalf("session did not use supplied snapshot: %#v", snapshot)
	}
}

//nolint:govet // short-lived error declarations keep failure assertions local to each operation.
func TestServiceRelaunchAndSwitchResolveLatestSavedCards(t *testing.T) {
	api, repo, sessions, workspace, workspaceID := newServiceFixture(t)
	first := saveCard(t, repo.FileRepository, config.IDCard{
		ID: "first", Name: "old", Driver: "cat", WorkspaceID: workspaceID, AgentsMD: "before",
	})
	second := saveCard(t, repo.FileRepository, config.IDCard{ID: "second", Name: "destination", Driver: "cat", WorkspaceID: workspaceID})
	sid, err := api.Launch(first)
	if err != nil {
		t.Fatal(err)
	}
	if err := sessions.Kill(sid); err != nil {
		t.Fatal(err)
	}
	first.AgentsMD = "updated-before-relaunch"
	if _, err := repo.UpdateIDCard(first); err != nil {
		t.Fatal(err)
	}
	if err := api.Relaunch(sid); err != nil {
		t.Fatalf("Relaunch: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(workspace, "AGENTS.md"))
	if err != nil || string(content) != "before" {
		t.Fatalf("relaunch changed deployed instruction: %q, %v", content, err)
	}
	if err := sessions.Kill(sid); err != nil {
		t.Fatal(err)
	}
	second.AgentsMD = "updated-before-switch"
	if _, err := repo.UpdateIDCard(second); err != nil {
		t.Fatal(err)
	}
	if err := api.SwitchAgent(sid, second.ID); err != nil {
		t.Fatalf("SwitchAgent: %v", err)
	}
	got, _ := sessions.GetSession(sid)
	content, err = os.ReadFile(filepath.Join(workspace, "AGENTS.md"))
	if err != nil || string(content) != second.AgentsMD {
		t.Fatalf("switch did not sync latest saved card configuration: %q, %v", content, err)
	}
	if got.CardID != second.ID || got.WorkingDir != workspace {
		t.Fatalf("switch did not use latest saved card and existing workspace: %#v", got)
	}
}

//nolint:gocognit,govet // exercises manager and filesystem effects across a full public switch operation.
func TestServiceSwitchPreservesWorkspaceAcrossDifferentCardSettings(t *testing.T) {
	api, repo, sessions, sourceWorkspace, sourceWorkspaceID := newServiceFixture(t)
	spawnLog := filepath.Join(t.TempDir(), "spawn-dirs")
	driverPath := filepath.Join(t.TempDir(), "record-pwd")
	driverScript := "#!/bin/sh\nprintf '%s\\n' \"$PWD\" >> " + shellQuote(spawnLog) + "\nexec cat\n"
	if err := os.WriteFile(driverPath, []byte(driverScript), 0o700); err != nil {
		t.Fatal(err)
	}
	destinationWorkspace := t.TempDir()
	if err := repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: sourceWorkspace}, {Path: destinationWorkspace}}); err != nil {
		t.Fatal(err)
	}
	workspaces, err := repo.ListWorkspaces()
	if err != nil || len(workspaces) != 2 {
		t.Fatalf("list workspaces: %v (%d)", err, len(workspaces))
	}
	var destinationID string
	for _, workspace := range workspaces {
		if workspace.Path == destinationWorkspace {
			destinationID = workspace.ID
		}
	}
	start := saveCard(t, repo.FileRepository, config.IDCard{ID: "source", Name: "source", Driver: "cat", WorkspaceID: sourceWorkspaceID})
	destination := saveCard(t, repo.FileRepository, config.IDCard{
		ID: "dest", Name: "dest", Driver: driverPath, WorkspaceID: destinationID, UseTmpWorkspace: true,
	})
	workFile := filepath.Join(sourceWorkspace, "preserve.txt")
	if err := os.WriteFile(workFile, []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}
	sid, err := api.Launch(start)
	if err != nil {
		t.Fatal(err)
	}
	if err := sessions.Kill(sid); err != nil {
		t.Fatal(err)
	}
	if err := api.SwitchAgent(sid, destination.ID); err != nil {
		t.Fatalf("SwitchAgent: %v", err)
	}
	var spawnDirs []byte
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		spawnDirs, err = os.ReadFile(spawnLog)
		if err == nil && len(strings.TrimSpace(string(spawnDirs))) > 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("read actual spawn directories: %v", err)
	}
	if got := strings.Split(strings.TrimSpace(string(spawnDirs)), "\n"); len(got) != 1 || got[0] != sourceWorkspace {
		t.Fatalf("actual switch spawn directory = %q, want retained source workspace %q", got, sourceWorkspace)
	}
	got, _ := sessions.GetSession(sid)
	content, err := os.ReadFile(workFile)
	if err != nil || string(content) != "keep" {
		t.Fatalf("existing work file: %q, %v", content, err)
	}
	if got.WorkingDir != sourceWorkspace || got.UseTmpWorkspace || got.CardID != destination.ID {
		t.Fatalf("switch changed session workspace identity: %#v", got)
	}
	if _, err := os.Stat(filepath.Join(destinationWorkspace, "preserve.txt")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("destination workspace unexpectedly used; stat error=%v", err)
	}
}

//nolint:gocognit,govet // Table-driven lifecycle test verifies path, metadata, spawn directory, and user work.
func TestServiceRelaunchPreservesSavedWorkspaceIdentity(t *testing.T) {
	for _, sourceTmp := range []bool{false, true} {
		name := "source-original"
		if sourceTmp {
			name = "source-temporary"
		}
		t.Run(name, func(t *testing.T) {
			t.Setenv("KANKURO_TMP_DIR", t.TempDir())
			api, repo, sessions, sourceWorkspace, sourceWorkspaceID := newServiceFixture(t)
			destinationWorkspace := t.TempDir()
			if err := repo.SaveWorkspaceSources(
				[]config.WorkspaceSource{{Path: sourceWorkspace}, {Path: destinationWorkspace}},
			); err != nil {
				t.Fatal(err)
			}
			workspaces, err := repo.ListWorkspaces()
			if err != nil {
				t.Fatal(err)
			}
			var destinationID string
			for _, workspace := range workspaces {
				if workspace.Path == destinationWorkspace {
					destinationID = workspace.ID
				}
			}
			spawnLog := filepath.Join(t.TempDir(), "spawn-dirs")
			driverPath := filepath.Join(t.TempDir(), "record-pwd")
			script := "#!/bin/sh\nprintf '%s\\n' \"$PWD\" >> " + shellQuote(spawnLog) + "\nexec cat\n"
			if err := os.WriteFile(driverPath, []byte(script), 0o700); err != nil {
				t.Fatal(err)
			}
			card := saveCard(t, repo.FileRepository, config.IDCard{
				ID: "card", Name: "card", Driver: driverPath, WorkspaceID: sourceWorkspaceID, UseTmpWorkspace: sourceTmp,
			})
			sessionID, err := api.Launch(card)
			if err != nil {
				t.Fatalf("launch: %v", err)
			}
			waitForSpawnLog(t, spawnLog, 1)
			before, _ := sessions.GetSession(sessionID)
			workFile := filepath.Join(before.WorkingDir, "keep.txt")
			if err := os.WriteFile(workFile, []byte("keep"), 0o600); err != nil {
				t.Fatal(err)
			}
			if err := sessions.Kill(sessionID); err != nil {
				t.Fatal(err)
			}
			card.WorkspaceID = destinationID
			card.UseTmpWorkspace = !sourceTmp
			if _, err := repo.UpdateIDCard(card); err != nil {
				t.Fatal(err)
			}
			if err := api.Relaunch(sessionID); err != nil {
				t.Fatalf("relaunch: %v", err)
			}
			waitForSpawnLog(t, spawnLog, 2)
			after, _ := sessions.GetSession(sessionID)
			if after.WorkingDir != before.WorkingDir || after.OriginalDir != before.OriginalDir || after.UseTmpWorkspace != sourceTmp {
				t.Fatalf("relaunch changed saved workspace identity: before=%#v after=%#v", before, after)
			}
			content, err := os.ReadFile(workFile)
			if err != nil || string(content) != "keep" {
				t.Fatalf("work file after relaunch: %q, %v", content, err)
			}
			if err := sessions.Kill(sessionID); err != nil {
				t.Fatal(err)
			}
			spawnDirs, err := os.ReadFile(spawnLog)
			if err != nil {
				t.Fatal(err)
			}
			for got := range strings.SplitSeq(strings.TrimSpace(string(spawnDirs)), "\n") {
				if got != before.WorkingDir {
					t.Fatalf("spawn directory %q, want retained %q", got, before.WorkingDir)
				}
			}
		})
	}
}

func waitForSpawnLog(t *testing.T, path string, count int) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		data, err := os.ReadFile(path)
		if err == nil {
			lines := strings.Split(strings.TrimSpace(string(data)), "\n")
			if len(lines) >= count && lines[0] != "" {
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("spawn log %s did not reach %d entries", path, count)
}

func shellQuote(value string) string { return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'" }

//nolint:gocognit,govet,nestif // a table exercises three distinct activation entry points and their cleanup.
func TestServiceActivationPersistenceFailuresKeepSessionsUsable(t *testing.T) {
	for _, operation := range []string{"launch", "relaunch", "switch"} {
		t.Run(operation, func(t *testing.T) {
			api, repo, sessions, _, workspaceID := newServiceFixture(t)
			card := saveCard(t, repo.FileRepository, config.IDCard{ID: "card", Name: "card", Driver: "cat", WorkspaceID: workspaceID})
			var sid string
			if operation != "launch" {
				var err error
				sid, err = api.Launch(card)
				if err != nil {
					t.Fatal(err)
				}
				if err := sessions.Kill(sid); err != nil {
					t.Fatal(err)
				}
			}
			if operation == "launch" {
				if err := os.WriteFile(filepath.Join(repo.Dir(), "sessions"), []byte("blocked"), 0o600); err != nil {
					t.Fatal(err)
				}
			} else {
				sessionsPath := filepath.Join(repo.Dir(), "sessions")
				if err := os.Remove(filepath.Join(sessionsPath, sid+".json")); err != nil {
					t.Fatal(err)
				}
				if err := os.Remove(sessionsPath); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(sessionsPath, []byte("blocked"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			var err error
			if operation == "launch" {
				sid, err = api.Launch(card)
			}
			if operation == "relaunch" {
				err = api.Relaunch(sid)
			}
			if operation == "switch" {
				err = api.SwitchAgent(sid, card.ID)
			}
			if !errors.Is(err, session.ErrPersistFailed) {
				t.Fatalf("%s error = %v, want ErrPersistFailed", operation, err)
			}
			if operation == "launch" && sid == "" {
				t.Fatal("Launch discarded the usable session ID on persistence error")
			}
			active, ok := sessions.GetSession(sid)
			if !ok || active.Status == session.StatusTerminated {
				t.Fatalf("activation not usable after persistence failure: %#v, exists=%v", active, ok)
			}
			handle, err := sessions.Attach(sid)
			if err != nil {
				t.Fatalf("Attach after persistence failure: %v", err)
			}
			handle.Close()
		})
	}
}

//nolint:govet // local error declarations keep each expected failure assertion readable.
func TestServiceDeleteCardGuardsAndRepositoryErrors(t *testing.T) {
	api, repo, sessions, _, workspaceID := newServiceFixture(t)
	card := saveCard(t, repo.FileRepository, config.IDCard{ID: "referenced", Name: "referenced", Driver: "cat", WorkspaceID: workspaceID})
	sid, err := api.Launch(card)
	if err != nil {
		t.Fatal(err)
	}
	if err := api.DeleteCard(card.ID); err == nil || !strings.Contains(err.Error(), "referenced by session(s) referenced") {
		t.Fatalf("running reference was not refused: %v", err)
	}
	if err := sessions.Kill(sid); err != nil {
		t.Fatal(err)
	}
	if err := api.DeleteCard(card.ID); err == nil {
		t.Fatal("terminated-session reference was not refused")
	}
	unreferenced := saveCard(t, repo.FileRepository, config.IDCard{ID: "free", Name: "free", Driver: "cat", WorkspaceID: workspaceID})
	if err := api.DeleteCard(unreferenced.ID); err != nil {
		t.Fatalf("delete unreferenced card: %v", err)
	}
	if repo.deleteErr != nil {
		t.Fatal("test setup unexpectedly had deletion error")
	}
	repo.deleteErr = errors.New("write failed")
	if err := api.DeleteCard("other"); !errors.Is(err, repo.deleteErr) {
		t.Fatalf("repository deletion error not preserved: %v", err)
	}
}

//nolint:govet // short declarations keep each lifecycle failure assertion local.
func TestServiceRemoveSessionDelegatesGuardsCleanupAndManagerErrors(t *testing.T) {
	api, repo, sessions, _, workspaceID := newServiceFixture(t)
	card := saveCard(t, repo.FileRepository, config.IDCard{
		ID: "tmp", Name: "tmp", Driver: "cat", WorkspaceID: workspaceID, UseTmpWorkspace: true,
	})
	sid, err := api.Launch(card)
	if err != nil {
		t.Fatal(err)
	}
	if err := api.RemoveSession(sid); err == nil || !strings.Contains(err.Error(), "session still running") {
		t.Fatalf("running removal not rejected: %v", err)
	}
	if err := sessions.Kill(sid); err != nil {
		t.Fatal(err)
	}
	snapshot, _ := sessions.GetSession(sid)
	if _, err := os.Stat(snapshot.WorkingDir); err != nil {
		t.Fatalf("temporary workspace missing before remove: %v", err)
	}
	if err := api.RemoveSession(sid); err != nil {
		t.Fatalf("remove terminated session: %v", err)
	}
	if _, err := os.Stat(snapshot.WorkingDir); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("temporary workspace remains: %v", err)
	}

	failedCard := saveCard(t, repo.FileRepository, config.IDCard{ID: "cleanup", Name: "cleanup", Driver: "cat", WorkspaceID: workspaceID})
	failedID, err := api.Launch(failedCard)
	if err != nil {
		t.Fatal(err)
	}
	if err := sessions.Kill(failedID); err != nil {
		t.Fatal(err)
	}
	recordPath := filepath.Join(repo.Dir(), "sessions", failedID+".json")
	if err := os.Remove(recordPath); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(recordPath, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(recordPath, "blocker"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := api.RemoveSession(failedID); err == nil {
		t.Fatal("expected session-record cleanup failure")
	}
	if _, ok := sessions.GetSession(failedID); ok {
		t.Fatal("manager retained session after cleanup failure")
	}

	closedCard := saveCard(t, repo.FileRepository, config.IDCard{ID: "closed", Name: "closed", Driver: "cat", WorkspaceID: workspaceID})
	closedID, err := api.Launch(closedCard)
	if err != nil {
		t.Fatal(err)
	}
	sessions.Shutdown()
	if err := api.Relaunch(closedID); !errors.Is(err, session.ErrManagerClosed) {
		t.Fatalf("shutdown error not propagated through service: %v", err)
	}
}

func idsSkills(items []config.Skill) []string {
	out := make([]string, len(items))
	for i, item := range items {
		out[i] = item.ID
	}
	return out
}
func idsPersonas(items []config.Persona) []string {
	out := make([]string, len(items))
	for i, item := range items {
		out[i] = item.ID
	}
	return out
}
func idsMCPs(items []config.MCPServer) []string {
	out := make([]string, len(items))
	for i, item := range items {
		out[i] = item.ID
	}
	return out
}
func assertIDs(t *testing.T, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("IDs = %v, want %v", got, want)
	}
}

func TestServiceMissingSessionAndActivationErrorsRemainDistinct(t *testing.T) {
	api, repo, sessions, _, workspaceID := newServiceFixture(t)
	card := saveCard(t, repo.FileRepository, config.IDCard{ID: "card", Name: "card", Driver: "cat", WorkspaceID: workspaceID})
	if err := api.Relaunch("missing"); err == nil || !strings.Contains(err.Error(), "session not found") {
		t.Fatalf("missing session error = %v", err)
	}
	if err := api.SwitchAgent("missing", card.ID); err == nil || strings.Contains(err.Error(), "card not found") {
		t.Fatalf("switch lifecycle error = %v", err)
	}
	repo.personasErr = errors.New("personas unavailable")
	_, err := api.Launch(card)
	if err == nil || !strings.HasPrefix(err.Error(), "list personas:") {
		t.Fatalf("launch should abort on capability failure: %v", err)
	}
	if len(sessions.List()) != 0 {
		t.Fatalf("activation spawned despite capability resolution failure: %#v", sessions.List())
	}
}

//nolint:gocognit // One end-to-end scenario verifies the messaging lifecycle and replay behavior.
func TestMessagingOperationsPinActivationsAndPersistExplicitReports(t *testing.T) {
	mcpSkill, err := os.ReadFile("../../kankuro_skills/mcp-cli/SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	puppetSkill, err := os.ReadFile("../../kankuro_skills/kankuro-puppet/SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	previousMCPSkill, _ := config.GetBuiltinSkillContent("mcp-cli")
	previousPuppetSkill, _ := config.GetBuiltinSkillContent("kankuro-puppet")
	config.SetBuiltinMCPCLISkillContent(mcpSkill)
	config.SetBuiltinKankuroPuppetSkillContent(puppetSkill)
	t.Cleanup(func() {
		config.SetBuiltinMCPCLISkillContent(previousMCPSkill)
		config.SetBuiltinKankuroPuppetSkillContent(previousPuppetSkill)
	})
	service, repo, manager, workspace, workspaceID := newServiceFixture(t)
	manager.SetMCPEndpoint("http://127.0.0.1:12345/mcp")
	callerCard := saveCard(
		t,
		repo.FileRepository,
		config.IDCard{ID: "orchestrator", Name: "orchestrator", Driver: "cat", WorkspaceID: workspaceID,
			DashboardControl: true, MCPIDs: []string{config.BuiltinMCPKankuro}},
	)
	recipientCard := saveCard(
		t,
		repo.FileRepository,
		config.IDCard{ID: "recipient", Name: "recipient", Driver: "cat", WorkspaceID: workspaceID,
			MCPIDs: []string{config.BuiltinMCPKankuro}},
	)
	_, err = service.Launch(callerCard)
	if err != nil {
		t.Fatal(err)
	}
	tokenFromWorkspace := func() string {
		var configFile struct {
			Servers map[string]struct {
				Headers map[string]string `json:"headers"`
			} `json:"mcpServers"`
		}
		data, readErr := os.ReadFile(filepath.Join(workspace, ".kankuro", "mcp-servers.json"))
		if readErr != nil {
			t.Fatal(readErr)
		}
		if readErr = json.Unmarshal(data, &configFile); readErr != nil {
			t.Fatal(readErr)
		}
		return strings.TrimPrefix(configFile.Servers["kankuro"].Headers["Authorization"], "Bearer ")
	}
	callerToken := tokenFromWorkspace()
	recipientID, err := service.Launch(recipientCard)
	if err != nil {
		t.Fatal(err)
	}
	recipientToken := tokenFromWorkspace()
	runner := NewOperationRunner(repo.Dir(), service)
	service.SetOperationRunner(runner)
	t.Cleanup(runner.Shutdown)
	inspection, err := service.GetSession(callerToken, recipientID)
	if err != nil || inspection.ActivationID == "" || !inspection.KankuroMCPEnabled {
		t.Fatalf("recipient inspection = %#v, %v", inspection, err)
	}
	requestID := uuid.NewString()
	result, err := service.SendMessage(
		context.Background(),
		callerToken,
		requestID,
		recipientID,
		inspection.ActivationID,
		"check the workspace",
	)
	if err != nil || result.State != OperationSucceeded || result.Result == nil || result.Result.MessageID == "" ||
		result.Result.Delivery != string(session.MessageSubmitted) {
		t.Fatalf("send result = %#v, %v", result, err)
	}
	retry, err := service.SendMessage(
		context.Background(),
		callerToken,
		requestID,
		recipientID,
		inspection.ActivationID,
		"check the workspace",
	)
	if err != nil || retry.ID != result.ID || retry.Result.MessageID != result.Result.MessageID {
		t.Fatalf("send retry = %#v, %v", retry, err)
	}
	if _, err = service.SendMessage(
		context.Background(),
		callerToken,
		requestID,
		recipientID,
		inspection.ActivationID,
		"different text",
	); err == nil {
		t.Fatal("changed send retry was accepted")
	}
	stored, err := service.GetMessage(callerToken, result.Result.MessageID)
	if err != nil || stored.Text != "check the workspace" {
		t.Fatalf("stored message = %#v, %v", stored, err)
	}
	if _, err = service.ReportReply(recipientToken, uuid.NewString(), result.Result.MessageID, "completed", "done"); err != nil {
		t.Fatalf("recipient without Dashboard Control could not report: %v", err)
	}
	page, err := service.GetReplies(context.Background(), callerToken, result.Result.MessageID, "", 50, 0)
	if err != nil || len(page.Items) != 1 || page.Items[0].Kind != "completed" {
		t.Fatalf("reply page = %#v, %v", page, err)
	}
	if err = manager.Kill(recipientID); err != nil {
		t.Fatal(err)
	}
	if err = service.RemoveSession(recipientID); err != nil {
		t.Fatal(err)
	}
	if _, err = service.GetMessage(callerToken, result.Result.MessageID); err == nil {
		t.Fatal("recipient deletion kept message content")
	}
}
