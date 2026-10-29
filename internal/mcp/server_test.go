package mcp_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/BlackGiack/kankuro/internal/api"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/mcp"
	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/google/uuid"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

const testShutdownTimeout = 5 * time.Second

type testRegistrar struct {
	registered bool
}

type discoveryRegistrar struct {
	testRegistrar

	mutationCalls int
	lastRequestID string
}

func (r *discoveryRegistrar) SendMessage(context.Context, string, string, string, string, string) (api.OperationResult, error) {
	return api.OperationResult{}, nil
}
func (r *discoveryRegistrar) GetMessage(string, string) (api.MessageRecord, error) {
	return api.MessageRecord{}, nil
}
func (r *discoveryRegistrar) ReportReply(string, string, string, string, string) (api.MessageReply, error) {
	return api.MessageReply{}, nil
}
func (r *discoveryRegistrar) GetReplies(context.Context, string, string, string, int, time.Duration) (api.ReplyPage, error) {
	return api.ReplyPage{}, nil
}

func (r *discoveryRegistrar) ResolveCaller(string) (api.CallerContext, error) {
	return api.CallerContext{SessionID: "caller", DashboardID: api.DefaultDashboardID, DashboardControl: true}, nil
}
func (*discoveryRegistrar) ListCards(string) ([]api.CardDiscovery, error) {
	return nil, nil
}
func (*discoveryRegistrar) ListSessions(string) ([]api.SessionDiscovery, error) { return nil, nil }
func (*discoveryRegistrar) GetSession(string, string) (api.SessionInspection, error) {
	return api.SessionInspection{}, nil
}
func (r *discoveryRegistrar) CreateSession(_ context.Context, _ string, requestID, _ string, _ ...*string) (api.OperationResult, error) {
	r.mutationCalls++
	r.lastRequestID = requestID
	if requestID == "bad-request" {
		return api.OperationResult{}, &api.RejectedError{Code: "invalid_request", Message: "request_id must be a UUID"}
	}
	return api.OperationResult{
		ID: "operation-id", RequestID: requestID, Kind: "create_session", State: api.OperationSucceeded,
		TargetSession: "target-session", Result: &api.OperationEffects{SessionID: "target-session"},
	}, nil
}
func (*discoveryRegistrar) StopSession(context.Context, string, string, string) (api.OperationResult, error) {
	return api.OperationResult{}, nil
}
func (*discoveryRegistrar) RelaunchSession(context.Context, string, string, string, *api.ConversationRequest) (api.OperationResult, error) {
	return api.OperationResult{}, nil
}
func (*discoveryRegistrar) SwitchSessionCard(
	context.Context, string, string, string, string, *api.ConversationRequest,
) (api.OperationResult, error) {
	return api.OperationResult{}, nil
}
func (*discoveryRegistrar) DeleteSession(context.Context, string, string, string) (api.OperationResult, error) {
	return api.OperationResult{}, nil
}
func (*discoveryRegistrar) GetOperation(string, string) (api.OperationResult, error) {
	return api.OperationResult{}, nil
}
func (*discoveryRegistrar) CancelOperation(string, string) (api.OperationResult, error) {
	return api.OperationResult{}, nil
}

func (r *testRegistrar) HasActivationCredential(token string) bool {
	return token == "valid-test-token"
}
func (r *testRegistrar) RegisterConversation(_ string, id, name string) (api.RegistrationResult, error) {
	r.registered = id == "native-id" && name == "Work"
	return api.RegistrationResult{Status: session.RegistrationRegistered, Persistence: session.RegistrationSaved}, nil
}

type authRoundTripper struct {
	next  http.RoundTripper
	token string
}

func (t authRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	token := t.token
	if token == "" {
		token = "valid-test-token"
	}
	request.Header.Set("Authorization", "Bearer "+token)
	return t.next.RoundTrip(request)
}

func readMCPBearer(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var configuration struct {
		Servers map[string]struct {
			Headers map[string]string `json:"headers"`
		} `json:"mcpServers"`
	}
	if decodeErr := json.Unmarshal(data, &configuration); decodeErr != nil {
		t.Fatal(decodeErr)
	}
	token := strings.TrimPrefix(configuration.Servers["kankuro"].Headers["Authorization"], "Bearer ")
	if token == "" {
		t.Fatal("Kankuro bearer credential is missing")
	}
	return token
}

//nolint:govet // Each request owns short-lived independent errors in the HTTP integration test.
func TestStreamableMCPRegistrationAndAuthentication(t *testing.T) {
	registrar := &testRegistrar{}
	runtime, err := mcp.Start(registrar)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), testShutdownTimeout)
		defer cancel()
		if err := runtime.Shutdown(ctx); err != nil {
			t.Errorf("shutdown MCP: %v", err)
		}
	})

	for _, authorization := range []string{"", "Bearer invalid-test-token"} {
		request, err := http.NewRequest(http.MethodPost, runtime.Endpoint(), strings.NewReader(`{}`))
		if err != nil {
			t.Fatal(err)
		}
		request.Header.Set("Content-Type", "application/json")
		if authorization != "" {
			request.Header.Set("Authorization", authorization)
		}
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = io.Copy(io.Discard, response.Body)
		_ = response.Body.Close()
		if response.StatusCode != http.StatusUnauthorized {
			t.Fatalf("unauthorized HTTP request status = %d, want %d", response.StatusCode, http.StatusUnauthorized)
		}
	}

	client := sdk.NewClient(&sdk.Implementation{Name: "test-client", Version: "test"}, nil)
	// The test owns this transport so closing the client tears down every
	// connection before the server's graceful-shutdown deadline begins.
	clientHTTPTransport := &http.Transport{DisableKeepAlives: true}
	defer clientHTTPTransport.CloseIdleConnections()
	transport := &sdk.StreamableClientTransport{
		Endpoint:   runtime.Endpoint(),
		HTTPClient: &http.Client{Transport: authRoundTripper{next: clientHTTPTransport}},
	}
	ctx, cancel := context.WithTimeout(context.Background(), testShutdownTimeout)
	defer cancel()
	clientSession, err := client.Connect(ctx, transport, nil)
	if err != nil {
		t.Fatalf("connect MCP client: %v", err)
	}
	defer closeTestClientSession(t, clientSession, clientHTTPTransport)
	tools, err := clientSession.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("list MCP tools: %v", err)
	}
	if len(tools.Tools) != 1 || tools.Tools[0].Name != "register_conversation" {
		t.Fatalf("MCP tools = %#v, want only register_conversation", tools.Tools)
	}
	result, err := clientSession.CallTool(ctx, &sdk.CallToolParams{
		Name:      "register_conversation",
		Arguments: map[string]any{"conversation_id": "native-id", "name": "Work"},
	})
	if err != nil {
		t.Fatalf("call registration tool: %v", err)
	}
	if result.IsError || !registrar.registered {
		t.Fatalf("registration result = %#v, registered=%v", result, registrar.registered)
	}
	closeTestClientSession(t, clientSession, clientHTTPTransport)
}

func closeTestClientSession(t *testing.T, clientSession *sdk.ClientSession, transport *http.Transport) {
	t.Helper()
	if err := clientSession.Close(); err != nil {
		t.Errorf("close MCP client session: %v", err)
	}
	transport.CloseIdleConnections()
}

//nolint:gocognit // Handshake, tool inventory, successful dispatch, and typed rejection share one client fixture.
func TestMCPRegistersLifecycleAndDiscoveryTools(t *testing.T) {
	registrar := &discoveryRegistrar{}
	runtime, err := mcp.Start(registrar)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), testShutdownTimeout)
		defer cancel()
		if shutdownErr := runtime.Shutdown(ctx); shutdownErr != nil {
			t.Errorf("shutdown MCP: %v", shutdownErr)
		}
	})
	client := sdk.NewClient(&sdk.Implementation{Name: "tool-list-test", Version: "test"}, nil)
	clientHTTPTransport := &http.Transport{DisableKeepAlives: true}
	defer clientHTTPTransport.CloseIdleConnections()
	transport := &sdk.StreamableClientTransport{
		Endpoint:   runtime.Endpoint(),
		HTTPClient: &http.Client{Transport: authRoundTripper{next: clientHTTPTransport}},
	}
	ctx, cancel := context.WithTimeout(context.Background(), testShutdownTimeout)
	defer cancel()
	clientSession, err := client.Connect(ctx, transport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer closeTestClientSession(t, clientSession, clientHTTPTransport)
	listed, err := clientSession.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{"register_conversation": false, "list_cards": false, "list_sessions": false, "get_session": false,
		"create_session": false, "stop_session": false, "relaunch_session": false, "switch_session_card": false,
		"delete_session": false, "get_operation": false, "cancel_operation": false}
	want["send_message"], want["get_message"], want["report_reply"], want["get_replies"] = false, false, false, false
	for _, tool := range listed.Tools {
		if _, ok := want[tool.Name]; ok {
			want[tool.Name] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("MCP tool %q was not registered", name)
		}
	}
	requestID := uuid.NewString()
	called, err := clientSession.CallTool(ctx, &sdk.CallToolParams{
		Name: "create_session", Arguments: map[string]any{"request_id": requestID, "card_id": "card-id"},
	})
	if err != nil || called.IsError {
		t.Fatalf("call lifecycle tool: result=%#v error=%v", called, err)
	}
	var completed struct {
		Operation struct {
			ID            string `json:"operation_id"`
			TargetSession string `json:"target_session_id"`
			Result        struct {
				SessionID string `json:"session_id"`
			} `json:"result"`
			ErrorCode string `json:"error_code"`
		} `json:"operation"`
	}
	calledPayload, marshalErr := json.Marshal(called.StructuredContent)
	if marshalErr != nil {
		t.Fatal(marshalErr)
	}
	if decodeErr := json.Unmarshal(calledPayload, &completed); decodeErr != nil {
		t.Fatal(decodeErr)
	}
	if completed.Operation.ID != "operation-id" || completed.Operation.TargetSession != "target-session" ||
		completed.Operation.Result.SessionID != "target-session" {
		t.Fatalf("structured lifecycle response = %#v", completed)
	}
	if registrar.mutationCalls != 1 || registrar.lastRequestID != requestID {
		t.Fatalf("lifecycle dispatch calls=%d request_id=%q", registrar.mutationCalls, registrar.lastRequestID)
	}
	rejected, err := clientSession.CallTool(ctx, &sdk.CallToolParams{
		Name: "create_session", Arguments: map[string]any{"request_id": "bad-request", "card_id": "card-id"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !rejected.IsError {
		t.Fatalf("rejected admission was not a tool error: %#v", rejected)
	}
	var rejection struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	payload, err := json.Marshal(rejected.StructuredContent)
	if err != nil {
		t.Fatalf("marshal rejection: %v", err)
	}
	if decodeErr := json.Unmarshal(payload, &rejection); decodeErr != nil {
		t.Fatalf("decode rejection: %v", decodeErr)
	}
	if rejection.Error.Code != "invalid_request" || rejection.Error.Message != "request_id must be a UUID" {
		t.Fatalf("rejection payload = %#v raw=%s", rejection, payload)
	}
}

//nolint:gocognit,gocyclo,cyclop // One end-to-end scenario validates each lifecycle tool over the real HTTP boundary.
func TestRealServiceLifecycleResultsOverMCP(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	skillContent, skillReadErr := os.ReadFile("../../kankuro_skills/mcp-cli/SKILL.md")
	if skillReadErr != nil {
		t.Fatal(skillReadErr)
	}
	config.SetBuiltinMCPCLISkillContent(skillContent)
	puppetContent, puppetReadErr := os.ReadFile("../../kankuro_skills/kankuro-puppet/SKILL.md")
	if puppetReadErr != nil {
		t.Fatal(puppetReadErr)
	}
	config.SetBuiltinKankuroPuppetSkillContent(puppetContent)
	t.Cleanup(func() { config.SetBuiltinMCPCLISkillContent(nil); config.SetBuiltinKankuroPuppetSkillContent(nil) })
	configDir := t.TempDir()
	t.Setenv("KANKURO_TMP_DIR", filepath.Join(configDir, "tmp"))
	repository, setupErr := config.NewFileRepository(configDir)
	if setupErr != nil {
		t.Fatal(setupErr)
	}
	callerWorkspace, projectWorkspace, subWorkspace := t.TempDir(), t.TempDir(), t.TempDir()
	if writeErr := os.WriteFile(filepath.Join(projectWorkspace, "project.txt"), []byte("project source"), 0o600); writeErr != nil {
		t.Fatal(writeErr)
	}
	if writeErr := os.WriteFile(filepath.Join(subWorkspace, "subproject.txt"), []byte("subproject source"), 0o600); writeErr != nil {
		t.Fatal(writeErr)
	}
	if saveWorkspaceErr := repository.SaveWorkspaceSources([]config.WorkspaceSource{
		{Path: callerWorkspace}, {Path: projectWorkspace}, {Path: subWorkspace},
	}); saveWorkspaceErr != nil {
		t.Fatal(saveWorkspaceErr)
	}
	workspaces, listWorkspacesErr := repository.ListWorkspaces()
	if listWorkspacesErr != nil || len(workspaces) != 3 {
		t.Fatalf("list fixture workspace: %v (%d workspaces)", listWorkspacesErr, len(workspaces))
	}
	workspaceID := func(path string) string {
		for _, workspace := range workspaces {
			if workspace.Path == path {
				return workspace.ID
			}
		}
		t.Fatalf("workspace %q not found in %#v", path, workspaces)
		return ""
	}
	manager := session.NewManager(configDir, repository, session.WithStartupDelay(0))
	t.Cleanup(func() { manager.Shutdown() })
	service := api.New(repository, manager)
	runner := api.NewOperationRunner(configDir, service)
	service.SetOperationRunner(runner)
	t.Cleanup(runner.Shutdown)
	runtime, startMCPError := mcp.Start(service)
	if startMCPError != nil {
		t.Fatal(startMCPError)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), testShutdownTimeout)
		defer cancel()
		if shutdownErr := runtime.Shutdown(ctx); shutdownErr != nil {
			t.Errorf("shutdown MCP: %v", shutdownErr)
		}
	})
	manager.SetMCPEndpoint(runtime.Endpoint())
	callerCard, createCallerCardErr := repository.CreateIDCard(config.IDCard{
		ID: "controller", Name: "controller", Driver: "cat", WorkspaceID: workspaceID(callerWorkspace),
		DashboardControl: true, MCPIDs: []string{config.BuiltinMCPKankuro},
	})
	if createCallerCardErr != nil {
		t.Fatal(createCallerCardErr)
	}
	firstCard, createFirstCardErr := repository.CreateIDCard(config.IDCard{
		ID: "worker", Name: "worker", Driver: "cat", WorkspaceID: workspaceID(projectWorkspace),
		UseTmpWorkspace: true, SubWorkspaceIDs: []string{workspaceID(subWorkspace)},
		MCPIDs: []string{config.BuiltinMCPKankuro},
	})
	if createFirstCardErr != nil {
		t.Fatal(createFirstCardErr)
	}
	secondCard, createSecondCardErr := repository.CreateIDCard(config.IDCard{
		ID: "reviewer", Name: "reviewer", Driver: "cat", WorkspaceID: workspaceID(projectWorkspace),
		MCPIDs: []string{config.BuiltinMCPKankuro},
	})
	if createSecondCardErr != nil {
		t.Fatal(createSecondCardErr)
	}
	if _, launchErr := service.Launch(callerCard); launchErr != nil {
		t.Fatalf("launch controller: %v", launchErr)
	}
	mcpConfig, readMCPConfigErr := os.ReadFile(filepath.Join(callerWorkspace, ".kankuro", "mcp-servers.json"))
	if readMCPConfigErr != nil {
		t.Fatal(readMCPConfigErr)
	}
	var activation struct {
		Servers map[string]struct {
			Headers map[string]string `json:"headers"`
		} `json:"mcpServers"`
	}
	if decodeMCPConfigErr := json.Unmarshal(mcpConfig, &activation); decodeMCPConfigErr != nil {
		t.Fatal(decodeMCPConfigErr)
	}
	token := strings.TrimPrefix(activation.Servers["kankuro"].Headers["Authorization"], "Bearer ")
	if token == "" {
		t.Fatal("controller activation token was not written")
	}
	client := sdk.NewClient(&sdk.Implementation{Name: "real-service-test", Version: "test"}, nil)
	clientHTTPTransport := &http.Transport{DisableKeepAlives: true}
	transport := &sdk.StreamableClientTransport{
		Endpoint: runtime.Endpoint(), HTTPClient: &http.Client{Transport: authRoundTripper{
			next: clientHTTPTransport, token: token,
		}},
	}
	ctx, cancel := context.WithTimeout(context.Background(), testShutdownTimeout)
	defer cancel()
	clientSession, connectErr := client.Connect(ctx, transport, nil)
	if connectErr != nil {
		t.Fatal(connectErr)
	}
	t.Cleanup(func() { closeTestClientSession(t, clientSession, clientHTTPTransport) })

	type operationResponse struct {
		Operation struct {
			ID            string `json:"operation_id"`
			State         string `json:"state"`
			TargetSession string `json:"target_session_id"`
			ErrorCode     string `json:"error_code"`
			ErrorMessage  string `json:"error_message"`
			Warning       string `json:"warning"`
			Result        struct {
				SessionID                 string `json:"session_id"`
				ProcessTerminated         bool   `json:"process_terminated"`
				SessionRecordPersisted    bool   `json:"session_record_persisted"`
				SessionRecordDeleted      bool   `json:"session_record_deleted"`
				TemporaryWorkspaceRemoved bool   `json:"temporary_workspace_removed"`
				InMemoryRemoved           bool   `json:"in_memory_removed"`
				Inspection                struct {
					WorkingDir           string `json:"working_dir"`
					ActiveConversationID string `json:"active_conversation_id"`
					Conversations        map[string][]struct {
						ID string `json:"id"`
					} `json:"conversations"`
					SubWorkspaces []struct {
						SourcePath      string `json:"source_path"`
						DestinationPath string `json:"destination_path"`
					} `json:"sub_workspaces"`
				} `json:"inspection"`
			} `json:"result"`
		} `json:"operation"`
	}
	call := func(name string, arguments map[string]any) operationResponse {
		t.Helper()
		result, callErr := clientSession.CallTool(ctx, &sdk.CallToolParams{Name: name, Arguments: arguments})
		if callErr != nil || result == nil || result.IsError {
			t.Fatalf("MCP %s failed: result=%#v error=%v", name, result, callErr)
		}
		encoded, encodeErr := json.Marshal(result.StructuredContent)
		if encodeErr != nil {
			t.Fatal(encodeErr)
		}
		var response operationResponse
		if decodeErr := json.Unmarshal(encoded, &response); decodeErr != nil {
			t.Fatal(decodeErr)
		}
		if response.Operation.ID == "" || response.Operation.State == "" {
			t.Fatalf("MCP %s returned no operation envelope: %s", name, encoded)
		}
		return response
	}
	connectAgent := func(agentToken string) *sdk.ClientSession {
		t.Helper()
		agentClient := sdk.NewClient(&sdk.Implementation{Name: "target-agent-test", Version: "test"}, nil)
		agentHTTPTransport := &http.Transport{DisableKeepAlives: true}
		agentTransport := &sdk.StreamableClientTransport{
			Endpoint: runtime.Endpoint(), HTTPClient: &http.Client{Transport: authRoundTripper{
				next: agentHTTPTransport, token: agentToken,
			}},
		}
		agentSession, agentConnectErr := agentClient.Connect(ctx, agentTransport, nil)
		if agentConnectErr != nil {
			t.Fatal(agentConnectErr)
		}
		t.Cleanup(func() { closeTestClientSession(t, agentSession, agentHTTPTransport) })
		return agentSession
	}
	requestID := uuid.NewString()
	created := call("create_session", map[string]any{"request_id": requestID, "card_id": firstCard.ID})
	targetID := created.Operation.TargetSession
	if targetID == "" || created.Operation.Result.SessionID != targetID {
		t.Fatalf("create result did not return structured target effects: %#v", created.Operation)
	}
	workingDir := created.Operation.Result.Inspection.WorkingDir
	if workingDir == "" || workingDir == callerWorkspace || workingDir == projectWorkspace ||
		len(created.Operation.Result.Inspection.SubWorkspaces) != 1 {
		t.Fatalf("cross-project workspace details = %#v", created.Operation.Result.Inspection)
	}
	subWorkspaceCopy := created.Operation.Result.Inspection.SubWorkspaces[0].DestinationPath
	if subWorkspaceCopy == subWorkspace || subWorkspaceCopy == "" {
		t.Fatalf("sub-workspace was not copied: %#v", created.Operation.Result.Inspection.SubWorkspaces[0])
	}
	if _, statErr := os.Stat(subWorkspaceCopy); statErr != nil {
		t.Fatalf("sub-workspace copy missing: %v", statErr)
	}
	handoffFile := filepath.Join(workingDir, "caller-handoff.txt")
	if writeErr := os.WriteFile(handoffFile, []byte("caller-created handoff"), 0o600); writeErr != nil {
		t.Fatal(writeErr)
	}
	targetToken := readMCPBearer(t, filepath.Join(workingDir, ".kankuro", "mcp-servers.json"))
	targetClientSession := connectAgent(targetToken)
	targetInspection, inspectTargetErr := service.GetSession(token, targetID)
	if inspectTargetErr != nil {
		t.Fatal(inspectTargetErr)
	}
	messageID := "mcp-error-contract-message"
	if _, createMessageErr := manager.Messages().CreateIntent(session.MessageRecord{
		ID: messageID, DashboardID: api.DefaultDashboardID, SenderSessionID: "controller",
		RecipientSessionID: targetID, RecipientActivationID: targetInspection.ActivationID, Text: "report through MCP",
	}, uuid.NewString(), session.SendFingerprint(targetID, targetInspection.ActivationID, "report through MCP")); createMessageErr != nil {
		t.Fatal(createMessageErr)
	}
	malformedCursor, malformedCursorErr := clientSession.CallTool(ctx, &sdk.CallToolParams{
		Name: "get_replies", Arguments: map[string]any{"message_id": messageID, "cursor": "%%%", "limit": 10},
	})
	if malformedCursorErr != nil {
		t.Fatal(malformedCursorErr)
	}
	assertMCPErrorCode(t, malformedCursor, "invalid_cursor")
	missingMessage, missingMessageErr := clientSession.CallTool(ctx, &sdk.CallToolParams{
		Name: "get_message", Arguments: map[string]any{"message_id": "missing-message"},
	})
	if missingMessageErr != nil {
		t.Fatal(missingMessageErr)
	}
	assertMCPErrorCode(t, missingMessage, "history_unavailable")
	invalidWait, invalidWaitErr := clientSession.CallTool(ctx, &sdk.CallToolParams{
		Name: "get_replies", Arguments: map[string]any{"message_id": messageID, "wait_seconds": 11},
	})
	if invalidWaitErr != nil {
		t.Fatal(invalidWaitErr)
	}
	assertMCPErrorCode(t, invalidWait, "invalid_request")
	progress, progressErr := targetClientSession.CallTool(ctx, &sdk.CallToolParams{
		Name: "report_reply", Arguments: map[string]any{"request_id": uuid.NewString(), "message_id": messageID,
			"kind": "progress", "text": "tests running"},
	})
	if progressErr != nil || progress == nil || progress.IsError {
		t.Fatalf("progress report failed: %#v, %v", progress, progressErr)
	}
	reportRequestID := uuid.NewString()
	firstReport, firstReportErr := targetClientSession.CallTool(ctx, &sdk.CallToolParams{
		Name: "report_reply", Arguments: map[string]any{"request_id": reportRequestID, "message_id": messageID,
			"kind": "completed", "text": "done"},
	})
	if firstReportErr != nil || firstReport == nil || firstReport.IsError {
		t.Fatalf("first correlated report failed: result=%#v err=%v", firstReport, firstReportErr)
	}
	changedReport, changedReportErr := targetClientSession.CallTool(ctx, &sdk.CallToolParams{
		Name: "report_reply", Arguments: map[string]any{"request_id": reportRequestID, "message_id": messageID,
			"kind": "completed", "text": "different report"},
	})
	if changedReportErr != nil {
		t.Fatal(changedReportErr)
	}
	assertMCPErrorCode(t, changedReport, "request_id_conflict")
	page, pageErr := manager.Messages().GetReplies(api.DefaultDashboardID, messageID, "", 0)
	if pageErr != nil || len(page.Items) != 2 || page.Items[0].Kind != "progress" || page.Items[1].Kind != "completed" {
		t.Fatalf("correlated reports = %#v, %v", page, pageErr)
	}
	deniedLifecycle, deniedLifecycleErr := targetClientSession.CallTool(ctx, &sdk.CallToolParams{
		Name: "create_session", Arguments: map[string]any{"request_id": uuid.NewString(), "card_id": secondCard.ID},
	})
	if deniedLifecycleErr != nil || deniedLifecycle == nil || !deniedLifecycle.IsError {
		t.Fatalf("registration-only lifecycle call was not rejected: result=%#v err=%v", deniedLifecycle, deniedLifecycleErr)
	}
	var deniedPayload struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	deniedEncoded, encodeDeniedErr := json.Marshal(deniedLifecycle.StructuredContent)
	if encodeDeniedErr != nil {
		t.Fatal(encodeDeniedErr)
	}
	if decodeDeniedErr := json.Unmarshal(deniedEncoded, &deniedPayload); decodeDeniedErr != nil {
		t.Fatal(decodeDeniedErr)
	}
	if deniedPayload.Error.Code != "forbidden" {
		t.Fatalf("registration-only lifecycle rejection = %s", deniedEncoded)
	}
	registeredA, registerAErr := targetClientSession.CallTool(ctx, &sdk.CallToolParams{
		Name: "register_conversation", Arguments: map[string]any{"conversation_id": "native-A", "name": "A"},
	})
	if registerAErr != nil || registeredA == nil || registeredA.IsError {
		t.Fatalf("register first-card conversation: result=%#v err=%v", registeredA, registerAErr)
	}
	deletionWhileRunning := call("delete_session", map[string]any{
		"request_id": uuid.NewString(), "session_id": targetID,
	})
	if deletionWhileRunning.Operation.State != string(api.OperationFailed) ||
		deletionWhileRunning.Operation.ErrorCode != "lifecycle_conflict" {
		t.Fatalf("delete precondition response = %#v", deletionWhileRunning.Operation)
	}
	replay := call("create_session", map[string]any{"request_id": requestID, "card_id": firstCard.ID})
	if replay.Operation.ID != created.Operation.ID || replay.Operation.Result.SessionID != targetID {
		t.Fatalf("create replay changed saved operation result: %#v", replay.Operation)
	}
	lookedUp := call("get_operation", map[string]any{"operation_id": created.Operation.ID})
	if lookedUp.Operation.ID != created.Operation.ID || lookedUp.Operation.Result.SessionID != targetID {
		t.Fatalf("operation lookup lost structured result: %#v", lookedUp.Operation)
	}
	stopped := call("stop_session", map[string]any{"request_id": uuid.NewString(), "session_id": targetID})
	if !stopped.Operation.Result.ProcessTerminated || !stopped.Operation.Result.SessionRecordPersisted {
		t.Fatalf("stop effects missing from MCP result: %#v", stopped.Operation.Result)
	}
	switched := call("switch_session_card", map[string]any{
		"request_id": uuid.NewString(), "session_id": targetID, "card_id": secondCard.ID,
	})
	if switched.Operation.Result.SessionID != targetID {
		t.Fatalf("switch result missing target: %#v", switched.Operation.Result)
	}
	if _, staleErr := targetClientSession.CallTool(ctx, &sdk.CallToolParams{
		Name: "create_session", Arguments: map[string]any{"request_id": uuid.NewString(), "card_id": firstCard.ID},
	}); staleErr == nil {
		t.Fatal("stale pre-switch credential remained usable at the MCP boundary")
	}
	secondToken := readMCPBearer(t, filepath.Join(workingDir, ".kankuro", "mcp-servers.json"))
	secondClientSession := connectAgent(secondToken)
	registeredB, registerBErr := secondClientSession.CallTool(ctx, &sdk.CallToolParams{
		Name: "register_conversation", Arguments: map[string]any{"conversation_id": "native-B", "name": "B"},
	})
	if registerBErr != nil || registeredB == nil || registeredB.IsError {
		t.Fatalf("register second-card conversation: result=%#v err=%v", registeredB, registerBErr)
	}
	call("stop_session", map[string]any{"request_id": uuid.NewString(), "session_id": targetID})
	choiceRequired := call("switch_session_card", map[string]any{
		"request_id": uuid.NewString(), "session_id": targetID, "card_id": firstCard.ID,
	})
	if choiceRequired.Operation.State != string(api.OperationFailed) ||
		choiceRequired.Operation.ErrorCode != "conversation_choice_required" {
		t.Fatalf("switch without conversation choice = %#v", choiceRequired.Operation)
	}
	selectedA := call("switch_session_card", map[string]any{
		"request_id": uuid.NewString(), "session_id": targetID, "card_id": firstCard.ID,
		"conversation": map[string]any{"mode": "resume", "conversation_id": "native-A"},
	})
	if selectedA.Operation.Result.Inspection.ActiveConversationID != "native-A" ||
		len(selectedA.Operation.Result.Inspection.Conversations[firstCard.ID]) != 1 ||
		len(selectedA.Operation.Result.Inspection.Conversations[secondCard.ID]) != 1 {
		t.Fatalf("A/B/A binding selection result = %#v", selectedA.Operation.Result.Inspection)
	}
	if contents, readErr := os.ReadFile(handoffFile); readErr != nil || string(contents) != "caller-created handoff" {
		t.Fatalf("workspace handoff after A/B/A = %q, %v", contents, readErr)
	}
	call("stop_session", map[string]any{"request_id": uuid.NewString(), "session_id": targetID})
	restoredA := call("relaunch_session", map[string]any{
		"request_id": uuid.NewString(), "session_id": targetID,
		"conversation": map[string]any{"mode": "resume", "conversation_id": "native-A"},
	})
	if restoredA.Operation.Result.Inspection.ActiveConversationID != "native-A" ||
		restoredA.Operation.Result.Inspection.WorkingDir != workingDir {
		t.Fatalf("relaunch did not retain A binding/workspace: %#v", restoredA.Operation.Result.Inspection)
	}
	call("stop_session", map[string]any{"request_id": uuid.NewString(), "session_id": targetID})
	sessionsDirectory := filepath.Join(configDir, "sessions")
	sessionsBackup := filepath.Join(configDir, "sessions-backup")
	if renameSessionsErr := os.Rename(sessionsDirectory, sessionsBackup); renameSessionsErr != nil {
		t.Fatal(renameSessionsErr)
	}
	if writeBlockerErr := os.WriteFile(sessionsDirectory, []byte("block persistence"), 0o600); writeBlockerErr != nil {
		_ = os.Rename(sessionsBackup, sessionsDirectory)
		t.Fatal(writeBlockerErr)
	}
	restoredSessionsDirectory := false
	restoreSessionsDirectory := func() {
		if restoredSessionsDirectory {
			return
		}
		if removeErr := os.Remove(sessionsDirectory); removeErr != nil && !os.IsNotExist(removeErr) {
			t.Errorf("remove persistence blocker: %v", removeErr)
			return
		}
		if renameErr := os.Rename(sessionsBackup, sessionsDirectory); renameErr != nil {
			t.Errorf("restore session records: %v", renameErr)
			return
		}
		restoredSessionsDirectory = true
	}
	t.Cleanup(restoreSessionsDirectory)
	relaunched := call("relaunch_session", map[string]any{
		"request_id": uuid.NewString(), "session_id": targetID,
		"conversation": map[string]any{"mode": "resume", "conversation_id": "native-A"},
	})
	restoreSessionsDirectory()
	if relaunched.Operation.State != string(api.OperationSucceededWithWarning) ||
		relaunched.Operation.Result.SessionID != targetID || relaunched.Operation.Warning == "" {
		t.Fatalf("relaunch persistence warning missing structured outcome: operation=%#v", relaunched.Operation)
	}
	call("stop_session", map[string]any{"request_id": uuid.NewString(), "session_id": targetID})
	partialSessionsDirectory := filepath.Join(configDir, "sessions")
	partialSessionsBackup := filepath.Join(configDir, "sessions-backup-delete")
	if renameErr := os.Rename(partialSessionsDirectory, partialSessionsBackup); renameErr != nil {
		t.Fatal(renameErr)
	}
	if writeErr := os.WriteFile(partialSessionsDirectory, []byte("block record deletion"), 0o600); writeErr != nil {
		_ = os.Rename(partialSessionsBackup, partialSessionsDirectory)
		t.Fatal(writeErr)
	}
	restoredPartialSessions := false
	restorePartialSessions := func() {
		if restoredPartialSessions {
			return
		}
		if removeErr := os.RemoveAll(partialSessionsDirectory); removeErr != nil {
			t.Errorf("remove deletion blocker: %v", removeErr)
			return
		}
		if renameErr := os.Rename(partialSessionsBackup, partialSessionsDirectory); renameErr != nil {
			t.Errorf("restore session record after partial deletion: %v", renameErr)
			return
		}
		restoredPartialSessions = true
	}
	t.Cleanup(restorePartialSessions)
	deleteRequestID := uuid.NewString()
	deleted := call("delete_session", map[string]any{"request_id": deleteRequestID, "session_id": targetID})
	restorePartialSessions()
	if deleted.Operation.State != string(api.OperationSucceededWithWarning) ||
		!deleted.Operation.Result.InMemoryRemoved || deleted.Operation.Result.SessionRecordDeleted ||
		!deleted.Operation.Result.TemporaryWorkspaceRemoved || deleted.Operation.Result.SessionID != targetID {
		t.Fatalf("delete partial-success effects missing from MCP result: %#v", deleted.Operation)
	}
	if _, statErr := os.Stat(filepath.Join(configDir, "tmp", targetID)); !os.IsNotExist(statErr) {
		t.Fatalf("temporary workspace was not removed on partial delete: %v", statErr)
	}
	deleteReplay := call("delete_session", map[string]any{"request_id": deleteRequestID, "session_id": targetID})
	if deleteReplay.Operation.ID != deleted.Operation.ID || !deleteReplay.Operation.Result.TemporaryWorkspaceRemoved {
		t.Fatalf("partial delete replay changed its saved effects: %#v", deleteReplay.Operation)
	}
}

func assertMCPErrorCode(t *testing.T, result *sdk.CallToolResult, want string) {
	t.Helper()
	if result == nil || !result.IsError {
		t.Fatalf("MCP result = %#v, want error code %q", result, want)
	}
	encoded, err := json.Marshal(result.StructuredContent)
	if err != nil {
		t.Fatal(err)
	}
	var rejection struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err = json.Unmarshal(encoded, &rejection); err != nil {
		t.Fatalf("decode MCP rejection %s: %v", encoded, err)
	}
	if rejection.Error.Code != want {
		t.Fatalf("MCP error code = %q, want %q (payload %s)", rejection.Error.Code, want, encoded)
	}
}

//nolint:govet // Each subtest builds an independent request and error value.
func TestRejectsInvalidHostAndOrigin(t *testing.T) {
	registrar := &testRegistrar{}
	runtime, err := mcp.Start(registrar)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = runtime.Shutdown(ctx)
	})

	for _, test := range []struct {
		name   string
		host   string
		origin string
	}{
		{name: "host", host: "example.com"},
		{name: "origin", origin: "https://attacker.example"},
	} {
		t.Run(test.name, func(t *testing.T) {
			request, err := http.NewRequest(http.MethodPost, runtime.Endpoint(), strings.NewReader(`{}`))
			if err != nil {
				t.Fatal(err)
			}
			request.Header.Set("Authorization", "Bearer valid-test-token")
			if test.host != "" {
				request.Host = test.host
			}
			if test.origin != "" {
				request.Header.Set("Origin", test.origin)
			}
			response, err := http.DefaultClient.Do(request)
			if err != nil {
				t.Fatal(err)
			}
			_, _ = io.Copy(io.Discard, response.Body)
			_ = response.Body.Close()
			if response.StatusCode != http.StatusForbidden {
				t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusForbidden)
			}
		})
	}
}
