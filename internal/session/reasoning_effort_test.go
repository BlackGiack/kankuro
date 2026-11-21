package session_test

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/BlackGiack/kankuro/internal/api"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/driver"
	"github.com/BlackGiack/kankuro/internal/mcp"
	"github.com/BlackGiack/kankuro/internal/session"
)

type captureCodexDriver struct {
	*captureClaudeDriver
}

func (d *captureCodexDriver) SetReasoningEffort(effort string) []string {
	return driver.NewCodexCLIDriver().SetReasoningEffort(effort)
}

func (d *captureCodexDriver) SetSession(id string) []string {
	return driver.NewCodexCLIDriver().SetSession(id)
}

func TestCodexResumeEffortOrdering(t *testing.T) {
	repo, workspaceID := workspaceValidationRepository(t)
	registry := driver.NewBuiltinRegistry()
	capture := &claudeCapture{}
	registry.Register("codex", func() driver.AgentDriver {
		return &captureCodexDriver{captureClaudeDriver: &captureClaudeDriver{
			ClaudeCodeDriver: driver.NewClaudeCodeDriver(), capture: capture,
		}}
	})
	manager := newTestManager(t, repo.Dir(), repo, session.WithDriverRegistry(registry))
	defer manager.Shutdown()
	runtime, err := mcp.Start(api.New(repo, manager))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if shutdownErr := runtime.Shutdown(ctx); shutdownErr != nil {
			t.Errorf("shutdown MCP runtime: %v", shutdownErr)
		}
	}()
	manager.SetMCPEndpoint(runtime.Endpoint())
	card := config.IDCard{
		ID: "codex-card", Name: "Codex", Driver: "codex", Model: "model",
		ReasoningEffort: "high", WorkspaceID: workspaceID,
	}
	id, err := testLaunch(manager, card, nil, nil, config.BuiltinMCPs())
	if err != nil {
		t.Fatal(err)
	}
	snapshot, found := manager.GetSession(id)
	if !found {
		t.Fatal("session was not published")
	}
	token := tokenFromConfig(t, driver.MCPConfigPath(snapshot.WorkingDir))
	const conversationID = "6cbf35c6-7e9d-4cb5-aefc-0ad55096ed8d"
	if _, registerErr := manager.RegisterConversation(token, conversationID, "Saved work"); registerErr != nil {
		t.Fatal(registerErr)
	}
	if killErr := manager.Kill(id); killErr != nil {
		t.Fatal(killErr)
	}
	if relaunchErr := testRelaunchWithChoice(manager, id, card, session.ConversationChoice{ID: conversationID}); relaunchErr != nil {
		t.Fatal(relaunchErr)
	}
	want := []string{"--model", "model", "-c", "model_reasoning_effort=high", "resume", conversationID}
	if got := capture.capturedArgs(1); !slices.Equal(got, want) {
		t.Fatalf("resumed Codex args = %#v, want %#v", got, want)
	}
}
