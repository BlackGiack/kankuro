package session_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/BlackGiack/kankuro/internal/api"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/driver"
	"github.com/BlackGiack/kankuro/internal/mcp"
	"github.com/BlackGiack/kankuro/internal/session"
)

const (
	testChildFileTimeout  = 2 * time.Second
	testChildPollInterval = 10 * time.Millisecond
)

type earlyReportingDriver struct {
	*driver.GenericCLIDriver

	manager     *session.Manager
	result      session.RegistrationResult
	registerErr error
	spawnErr    error
	token       string
}

func (d *earlyReportingDriver) Spawn(agentConfig driver.AgentConfig) (*driver.SessionHandle, error) {
	data, err := os.ReadFile(driver.MCPConfigPath(agentConfig.WorkingDir))
	if err != nil {
		return nil, err
	}
	var privateConfig struct {
		MCPServers map[string]struct {
			Headers map[string]string `json:"headers"`
		} `json:"mcpServers"`
	}
	if unmarshalErr := json.Unmarshal(data, &privateConfig); unmarshalErr != nil {
		return nil, unmarshalErr
	}
	token := strings.TrimPrefix(privateConfig.MCPServers["kankuro"].Headers["Authorization"], "Bearer ")
	d.token = token
	d.result, d.registerErr = d.manager.RegisterConversation(token, "early-native-id", "Early")
	if d.registerErr != nil {
		return nil, d.registerErr
	}
	if d.spawnErr != nil {
		return nil, d.spawnErr
	}
	return d.GenericCLIDriver.Spawn(agentConfig)
}

//nolint:cyclop,gocognit,gocyclo,govet // Exercises child-shell discovery, registration, persistence retry, revocation, and restart.
func TestMCPCLIRegistersConversationForLiveActivation(t *testing.T) {
	mcpCLI, err := exec.LookPath("mcp-cli")
	if err != nil {
		t.Skip("mcp-cli is unavailable; install mcp-cli v0.3.0 to run interoperability coverage")
	}
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	workspace := t.TempDir()
	if err := repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: workspace}}); err != nil {
		t.Fatal(err)
	}
	workspaces, err := repo.ListWorkspaces()
	if err != nil || len(workspaces) == 0 {
		t.Fatalf("workspaces: %v", err)
	}

	t.Setenv(driver.EnvKankuroMCPConfigPath, "/parent/activation.json")
	t.Setenv(driver.EnvKankuroMCPToken, "parent-token")
	manager := newTestManager(t, repo.Dir(), repo)
	runtime, err := mcp.Start(api.New(repo, manager))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := runtime.Shutdown(ctx); err != nil {
			t.Errorf("shutdown MCP runtime: %v", err)
		}
	})
	t.Cleanup(func() { manager.Shutdown() })
	manager.SetMCPEndpoint(runtime.Endpoint())
	childEnvPath := filepath.Join(t.TempDir(), "child-env")
	childResultPath := filepath.Join(t.TempDir(), "child-mcp-result")
	driverScript := filepath.Join(t.TempDir(), "agent-shell")
	script := "#!/bin/sh\nprintf '%s|%s\\n' \"${KANKURO_MCP_CONFIG_PATH-unset}\" \"${KANKURO_MCP_TOKEN-unset}\" > " +
		testShellQuote(childEnvPath) + "\nif [ \"${TEST_REGISTER_ONCE-unset}\" = \"1\" ]; then\nMCP_NO_DAEMON=1 " +
		testShellQuote(mcpCLI) + " -c .kankuro/mcp-servers.json call kankuro register_conversation '" +
		`{"conversation_id":"native-session-123","name":"Work"}` + "' > " +
		testShellQuote(childResultPath) + "\nfi\nexec cat\n"
	if err := os.WriteFile(driverScript, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	card := config.IDCard{
		ID: "card", Name: "card", Driver: driverScript, WorkspaceID: workspaces[0].ID,
		Env: map[string]string{
			driver.EnvKankuroMCPConfigPath: "/card/activation.json",
			driver.EnvKankuroMCPToken:      "card-token",
			"TEST_REGISTER_ONCE":           "1",
		},
	}
	card, err = repo.CreateIDCard(card)
	if err != nil {
		t.Fatal(err)
	}
	selectedMCPs := append(
		config.BuiltinMCPs(),
		config.MCPServer{
			ID:      "external",
			Name:    "external",
			URL:     "http://127.0.0.1:9000/mcp",
			Headers: map[string]string{"X-Token": "private"},
		},
	)
	sessionID, err := manager.Launch(card, nil, nil, selectedMCPs)
	if err != nil {
		t.Fatalf("launch: %v", err)
	}
	snapshot, found := manager.GetSession(sessionID)
	if !found {
		t.Fatal("launched session was not published")
	}
	mcpConfigPath := driver.MCPConfigPath(snapshot.WorkingDir)
	configInfo, err := os.Stat(mcpConfigPath)
	if err != nil || configInfo.Mode().Perm() != 0o600 {
		t.Fatalf("workspace MCP config mode = %v, %v; want 0600", configInfo, err)
	}
	metadataDirInfo, err := os.Stat(filepath.Dir(mcpConfigPath))
	if err != nil || metadataDirInfo.Mode().Perm() != 0o700 {
		t.Fatalf("workspace MCP metadata directory mode = %v, %v; want 0700", metadataDirInfo, err)
	}
	data, err := os.ReadFile(mcpConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	var workspaceConfig struct {
		MCPServers map[string]struct {
			URL     string            `json:"url"`
			Headers map[string]string `json:"headers"`
		} `json:"mcpServers"`
	}
	if err := json.Unmarshal(data, &workspaceConfig); err != nil {
		t.Fatal(err)
	}
	entry, ok := workspaceConfig.MCPServers["kankuro"]
	if !ok || entry.URL != runtime.Endpoint() || !strings.HasPrefix(entry.Headers["Authorization"], "Bearer ") {
		t.Fatalf("workspace MCP config invalid: %#v", workspaceConfig)
	}
	if external := workspaceConfig.MCPServers["external"]; external.URL != "http://127.0.0.1:9000/mcp" ||
		external.Headers["X-Token"] != "private" {
		t.Fatalf("workspace config omitted selected MCP server: %#v", external)
	}
	if childEnv := strings.TrimSpace(string(waitForTestFile(t, childEnvPath))); childEnv != "unset|unset" {
		t.Fatalf("agent child received reserved MCP variables: %q", childEnv)
	}
	childResult := string(waitForTestFile(t, childResultPath))
	if !strings.Contains(childResult, "pending_activation") && !strings.Contains(childResult, "registered") {
		t.Fatalf("real child-shell mcp-cli registration did not succeed: %s", childResult)
	}
	callRegistration := func() ([]byte, error) {
		command := exec.Command(
			mcpCLI,
			"-c",
			".kankuro/mcp-servers.json",
			"call",
			"kankuro",
			"register_conversation",
			`{"conversation_id":"native-session-123","name":"Work"}`,
		)
		command.Dir = snapshot.WorkingDir
		command.Env = append(testEnvWithoutLegacyMCP(), "MCP_NO_DAEMON=1")
		return command.CombinedOutput()
	}
	sessionsDir := filepath.Join(repo.Dir(), "sessions")
	sessionRecordPath := filepath.Join(sessionsDir, sessionID+".json")
	if err := os.Remove(sessionRecordPath); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(sessionsDir); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sessionsDir, []byte("blocked"), 0o600); err != nil {
		t.Fatal(err)
	}
	output, err := callRegistration()
	if err != nil {
		t.Fatalf("mcp-cli registration: %v\n%s", err, output)
	}
	if !strings.Contains(string(output), `"persistence": "unsaved"`) {
		t.Fatalf("registration failure did not report unsaved metadata: %s", output)
	}
	snapshot, found = manager.GetSession(sessionID)
	if !found || len(snapshot.Conversations[card.ID]) != 1 {
		t.Fatalf("registration not persisted in session snapshot: %#v", snapshot)
	}
	binding := snapshot.Conversations[card.ID][0]
	if binding.ID != "native-session-123" || binding.Name != "Work" || binding.Driver != driverScript {
		t.Fatalf("registered binding = %#v", binding)
	}
	if snapshot.RegistrationPersistence != session.RegistrationUnsaved {
		t.Fatalf("registration persistence = %q, want unsaved", snapshot.RegistrationPersistence)
	}
	if err := os.Remove(sessionsDir); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(sessionsDir, 0o700); err != nil {
		t.Fatal(err)
	}
	output, err = callRegistration()
	if err != nil {
		t.Fatalf("mcp-cli retry: %v\n%s", err, output)
	}
	if !strings.Contains(string(output), `"persistence": "saved"`) {
		t.Fatalf("registration retry did not report saved metadata: %s", output)
	}
	snapshot, found = manager.GetSession(sessionID)
	if !found || snapshot.RegistrationPersistence != session.RegistrationSaved {
		t.Fatalf("registration retry did not clear unsaved state: %#v", snapshot)
	}
	token := strings.TrimPrefix(entry.Headers["Authorization"], "Bearer ")
	sessionRecord, err := os.ReadFile(sessionRecordPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(sessionRecord), token) {
		t.Fatal("activation credential leaked into the session record")
	}
	if _, err := manager.RegisterConversation(token, "native-session-123", "Renamed work"); err != nil {
		t.Fatalf("idempotent registration: %v", err)
	}
	if _, err := manager.RegisterConversation(token, "different-native-id", "Wrong identity"); err == nil {
		t.Fatal("conflicting conversation identity unexpectedly succeeded")
	}
	snapshot, found = manager.GetSession(sessionID)
	if !found || len(snapshot.Conversations[card.ID]) != 1 || snapshot.Conversations[card.ID][0].Name != "Renamed work" {
		t.Fatalf("duplicate/conflicting reports changed conversation identity: %#v", snapshot)
	}
	binding = snapshot.Conversations[card.ID][0]
	if err := manager.Kill(sessionID); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.RegisterConversation(token, "native-session-123", "Late report"); err == nil {
		t.Fatal("terminated activation accepted a stale registration")
	}
	if err := api.New(repo, manager).Relaunch(sessionID); err == nil || !strings.Contains(err.Error(), "selection required") {
		t.Fatalf("single saved conversation was resumed without an explicit choice: %v", err)
	}
	destination, err := repo.CreateIDCard(
		config.IDCard{ID: "destination", Name: "destination", Driver: "cat", WorkspaceID: workspaces[0].ID},
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := manager.SwitchAgent(sessionID, destination, nil, nil, nil); err != nil {
		t.Fatalf("switch to destination card: %v", err)
	}
	if err := api.New(repo, manager).DeleteCard(card.ID); err == nil || !strings.Contains(err.Error(), "referenced by session(s)") {
		t.Fatalf("conversation binding did not protect inactive card deletion: %v", err)
	}
	if err := manager.Kill(sessionID); err != nil {
		t.Fatal(err)
	}
	if err := api.New(repo, manager).
		SwitchAgent(sessionID, card.ID); err == nil ||
		!strings.Contains(err.Error(), "conversation selection required") {
		t.Fatalf("card with saved conversations accepted an omitted selection: %v", err)
	}
	manager.Shutdown()
	restored := newTestManager(t, repo.Dir(), repo)
	restoredSnapshot, found := restored.GetSession(sessionID)
	if !found || len(restoredSnapshot.Conversations[card.ID]) != 1 || restoredSnapshot.Conversations[card.ID][0] != binding {
		t.Fatalf("conversation binding did not survive restart: %#v", restoredSnapshot)
	}
	if restoredSnapshot.Registration != session.RegistrationUnavailable {
		t.Fatalf("runtime registration state survived process restart: %q", restoredSnapshot.Registration)
	}
	restored.SetMCPEndpoint(runtime.Endpoint())
	card.Env = map[string]string{
		driver.EnvKankuroMCPConfigPath: "/card/activation.json",
		driver.EnvKankuroMCPToken:      "card-token",
	}
	if _, err := repo.UpdateIDCard(card); err != nil {
		t.Fatal(err)
	}
	if err := restored.SwitchAgentWithChoice(
		sessionID, card, nil, nil, selectedMCPs, session.ConversationChoice{New: true},
	); err != nil {
		t.Fatalf("start a second conversation for the card: %v", err)
	}
	secondSnapshot, _ := restored.GetSession(sessionID)
	secondConfigPath := driver.MCPConfigPath(secondSnapshot.WorkingDir)
	secondToken := tokenFromConfig(t, secondConfigPath)
	if secondConfigPath != mcpConfigPath || secondToken == token {
		t.Fatal("same workspace path did not receive a fresh per-activation credential")
	}
	if _, err := restored.RegisterConversation(secondToken, "native-session-456", "Renamed work"); err != nil {
		t.Fatalf("register second conversation: %v", err)
	}
	secondSnapshot, _ = restored.GetSession(sessionID)
	if got := secondSnapshot.Conversations[card.ID]; len(got) != 2 || got[0].Name != got[1].Name || got[0].ID == got[1].ID {
		t.Fatalf("multiple same-name conversations were not retained independently: %#v", got)
	}
	if err := restored.Kill(sessionID); err != nil {
		t.Fatal(err)
	}
	restored.Shutdown()
	loadedAgain := newTestManager(t, repo.Dir(), repo)
	loadedSnapshot, found := loadedAgain.GetSession(sessionID)
	if !found || len(loadedSnapshot.Conversations[card.ID]) != 2 {
		t.Fatalf("multiple conversations did not survive another restart: %#v", loadedSnapshot)
	}
	otherSessionID, err := loadedAgain.Launch(card, nil, nil, nil)
	if err != nil {
		t.Fatalf("launch another Kankuro session: %v", err)
	}
	otherBindings, err := api.New(repo, loadedAgain).Conversations(otherSessionID, card.ID)
	if err != nil || len(otherBindings) != 0 {
		t.Fatalf("another Kankuro session inherited conversation bindings: %#v, %v", otherBindings, err)
	}
}

func tokenFromConfig(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var config struct {
		MCPServers map[string]struct {
			Headers map[string]string `json:"headers"`
		} `json:"mcpServers"`
	}
	if unmarshalErr := json.Unmarshal(data, &config); unmarshalErr != nil {
		t.Fatal(unmarshalErr)
	}
	return strings.TrimPrefix(config.MCPServers["kankuro"].Headers["Authorization"], "Bearer ")
}

//nolint:cyclop,gocognit,gocyclo,govet // Subtests exercise source/temp workspaces and both activation modes through a real child.
func TestLaunchEnvironmentDoesNotLeakReservedMCPCredentials(t *testing.T) {
	for _, test := range []struct {
		name          string
		selectKankuro bool
		useTmp        bool
	}{
		{name: "unselected strips parent and card values"},
		{name: "selected writes Kankuro workspace entry", selectKankuro: true},
		{name: "temporary copy replaces source MCP file", selectKankuro: true, useTmp: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv(driver.EnvKankuroMCPConfigPath, "/parent/activation.json")
			t.Setenv(driver.EnvKankuroMCPToken, "parent-token")
			t.Setenv("KANKURO_TMP_DIR", t.TempDir())
			repo, err := config.NewFileRepository(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			workspace := t.TempDir()
			if err := repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: workspace}}); err != nil {
				t.Fatal(err)
			}
			workspaces, err := repo.ListWorkspaces()
			if err != nil || len(workspaces) != 1 {
				t.Fatalf("workspaces: %v", err)
			}
			if err := os.MkdirAll(filepath.Join(workspace, ".kankuro"), 0o700); err != nil {
				t.Fatal(err)
			}
			staleMCPConfig := []byte(
				`{"mcpServers":{"stale":{"url":"https://old"},"kankuro":{"url":"https://old","headers":{"Authorization":"Bearer old-token"}}}}`,
			)
			if err := os.WriteFile(driver.MCPConfigPath(workspace), staleMCPConfig, 0o600); err != nil {
				t.Fatal(err)
			}
			manager := newTestManager(t, repo.Dir(), repo)
			runtime, err := mcp.Start(api.New(repo, manager))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := runtime.Shutdown(ctx); err != nil {
					t.Errorf("shutdown MCP: %v", err)
				}
			})
			t.Cleanup(func() { manager.Shutdown() })
			manager.SetMCPEndpoint(runtime.Endpoint())
			outputPath := filepath.Join(t.TempDir(), "child-env")
			childConfigPath := filepath.Join(t.TempDir(), "child-config")
			scriptPath := filepath.Join(t.TempDir(), "capture-env")
			script := "#!/bin/sh\nprintf '%s|%s\\n' \"${KANKURO_MCP_CONFIG_PATH-unset}\" \"${KANKURO_MCP_TOKEN-unset}\" > " + testShellQuote(
				outputPath,
			) + "\ncat .kankuro/mcp-servers.json > " + testShellQuote(
				childConfigPath,
			) + "\nexec cat\n"
			if err := os.WriteFile(scriptPath, []byte(script), 0o700); err != nil {
				t.Fatal(err)
			}
			card := config.IDCard{
				ID: "capture", Name: "capture", Driver: scriptPath, WorkspaceID: workspaces[0].ID, UseTmpWorkspace: test.useTmp,
				Env: map[string]string{
					driver.EnvKankuroMCPConfigPath: "/card/activation.json",
					driver.EnvKankuroMCPToken:      "card-token",
				},
			}
			var mcps []config.MCPServer
			if test.selectKankuro {
				mcps = config.BuiltinMCPs()
			}
			sessionID, err := manager.Launch(card, nil, nil, mcps)
			if err != nil {
				t.Fatalf("launch: %v", err)
			}
			snapshot, found := manager.GetSession(sessionID)
			if !found {
				t.Fatal("session snapshot not found")
			}
			deadline := time.Now().Add(2 * time.Second)
			var childEnv []byte
			for time.Now().Before(deadline) {
				childEnv, err = os.ReadFile(outputPath)
				if err == nil {
					break
				}
				time.Sleep(10 * time.Millisecond)
			}
			if err != nil {
				t.Fatalf("read child environment: %v", err)
			}
			parts := strings.Split(strings.TrimSpace(string(childEnv)), "|")
			if len(parts) != 2 || parts[1] != "unset" {
				t.Fatalf("reserved child environment = %q", childEnv)
			}
			if parts[0] != "unset" {
				t.Fatalf("activation config path leaked to child: %q", parts[0])
			}
			var workspaceConfig struct {
				MCPServers map[string]struct {
					URL     string            `json:"url"`
					Headers map[string]string `json:"headers"`
				} `json:"mcpServers"`
			}
			data := waitForTestFile(t, childConfigPath)
			if err := json.Unmarshal(data, &workspaceConfig); err != nil {
				t.Fatal(err)
			}
			workspaceData, err := os.ReadFile(driver.MCPConfigPath(snapshot.WorkingDir))
			if err != nil || !bytes.Equal(workspaceData, data) {
				t.Fatalf("config observed by child differs from its workspace file: err=%v", err)
			}
			if test.selectKankuro {
				if auth := workspaceConfig.MCPServers["kankuro"].Headers["Authorization"]; !strings.HasPrefix(auth, "Bearer ") {
					t.Fatalf("selected activation has no Kankuro credential: %#v", workspaceConfig.MCPServers)
				}
			} else if len(workspaceConfig.MCPServers) != 0 {
				t.Fatalf("unselected activation retained MCP entries: %#v", workspaceConfig.MCPServers)
			}
			if _, found := workspaceConfig.MCPServers["stale"]; found {
				t.Fatal("activation retained a server from the source workspace's old MCP config")
			}
			if test.useTmp {
				sourceData, err := os.ReadFile(driver.MCPConfigPath(workspace))
				if err != nil || !bytes.Equal(sourceData, staleMCPConfig) {
					t.Fatalf("temporary activation modified source MCP config: %s, %v", sourceData, err)
				}
			}
			if err := manager.Kill(sessionID); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func testShellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func testEnvWithoutLegacyMCP() []string {
	env := make([]string, 0, len(os.Environ()))
	for _, entry := range os.Environ() {
		key, _, ok := strings.Cut(entry, "=")
		if ok && (key == driver.EnvKankuroMCPConfigPath || key == driver.EnvKankuroMCPToken) {
			continue
		}
		env = append(env, entry)
	}
	return env
}

func waitForTestFile(t *testing.T, path string) []byte {
	t.Helper()
	deadline := time.Now().Add(testChildFileTimeout)
	for time.Now().Before(deadline) {
		data, err := os.ReadFile(path)
		if err == nil && len(data) > 0 {
			return data
		}
		time.Sleep(testChildPollInterval)
	}
	t.Fatalf("child did not write %s", path)
	return nil
}

//nolint:gocognit,govet // Controlled spawn hooks cover both staged-report outcomes.
func TestRegistrationBeforeSpawnPublicationIsCommittedOrDiscarded(t *testing.T) {
	for _, test := range []struct {
		name     string
		spawnErr error
	}{
		{name: "commit after successful spawn"},
		{name: "discard after spawn failure", spawnErr: errors.New("spawn failed")},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo, err := config.NewFileRepository(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			workspace := t.TempDir()
			if err := repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: workspace}}); err != nil {
				t.Fatal(err)
			}
			workspaces, err := repo.ListWorkspaces()
			if err != nil || len(workspaces) == 0 {
				t.Fatalf("workspaces: %v", err)
			}
			registry := driver.NewBuiltinRegistry()
			var reportingDriver *earlyReportingDriver
			var managerRef *session.Manager
			registry.Register("cat", func() driver.AgentDriver {
				reportingDriver = &earlyReportingDriver{
					GenericCLIDriver: driver.NewGenericCLIDriver(),
					manager:          managerRef,
					spawnErr:         test.spawnErr,
				}
				return reportingDriver
			})
			manager := newTestManager(t, repo.Dir(), repo, session.WithDriverRegistry(registry))
			managerRef = manager
			runtime, err := mcp.Start(api.New(repo, manager))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				_ = runtime.Shutdown(ctx)
			})
			t.Cleanup(func() { manager.Shutdown() })
			manager.SetMCPEndpoint(runtime.Endpoint())
			card := config.IDCard{ID: "card", Name: "card", Driver: "cat", WorkspaceID: workspaces[0].ID}
			sessionID, launchErr := manager.Launch(card, nil, nil, config.BuiltinMCPs())
			if test.spawnErr != nil {
				if !errors.Is(launchErr, test.spawnErr) {
					t.Fatalf("launch error = %v, want %v", launchErr, test.spawnErr)
				}
				workspaceConfigPath := driver.MCPConfigPath(workspace)
				data, readErr := os.ReadFile(workspaceConfigPath)
				if readErr != nil {
					t.Fatalf("failed activation deleted workspace MCP file: %v", readErr)
				}
				var workspaceConfig struct {
					MCPServers map[string]json.RawMessage `json:"mcpServers"`
				}
				if unmarshalErr := json.Unmarshal(data, &workspaceConfig); unmarshalErr != nil || len(workspaceConfig.MCPServers) != 0 {
					t.Fatalf("failed activation left its Kankuro entry: %s, %v", data, unmarshalErr)
				}
				if result, err := manager.RegisterConversation(reportingDriver.token, "late", "Late"); err == nil || result.Status != "" {
					t.Fatalf("staged credential remained valid after failed spawn: %#v, %v", result, err)
				}
				return
			}
			if launchErr != nil {
				t.Fatalf("launch: %v", launchErr)
			}
			if reportingDriver.result.Status != "pending_activation" ||
				reportingDriver.result.Persistence != session.RegistrationNotAttempted {
				t.Fatalf("early registration response = %#v", reportingDriver.result)
			}
			snapshot, found := manager.GetSession(sessionID)
			if !found || len(snapshot.Conversations[card.ID]) != 1 || snapshot.Conversations[card.ID][0].ID != "early-native-id" ||
				snapshot.ActiveConversationID != "early-native-id" || snapshot.Registration != session.RegistrationRegistered ||
				snapshot.RegistrationPersistence != session.RegistrationSaved {
				t.Fatalf("early registration was not committed: %#v", snapshot)
			}
		})
	}
}
