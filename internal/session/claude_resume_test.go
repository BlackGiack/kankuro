package session_test

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/BlackGiack/kankuro/internal/api"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/driver"
	"github.com/BlackGiack/kankuro/internal/mcp"
	"github.com/BlackGiack/kankuro/internal/session"
)

type captureClaudeDriver struct {
	*driver.ClaudeCodeDriver

	output  chan []byte
	done    chan struct{}
	status  atomic.Bool
	stop    sync.Once
	capture *claudeCapture
}

type claudeCapture struct {
	mu      sync.Mutex
	configs []driver.AgentConfig
}

func (d *captureClaudeDriver) PrepareWorkspace(context.Context, string) error { return nil }

func (d *captureClaudeDriver) Spawn(cfg driver.AgentConfig) (*driver.SessionHandle, error) {
	d.capture.mu.Lock()
	d.capture.configs = append(d.capture.configs, cfg)
	d.capture.mu.Unlock()
	d.output = make(chan []byte)
	d.done = make(chan struct{})
	d.status.Store(true)
	return &driver.SessionHandle{Output: d.output, Done: d.done}, nil
}

func (d *captureClaudeDriver) Kill() error {
	d.stop.Do(func() {
		close(d.output)
		close(d.done)
		d.status.Store(false)
	})
	return nil
}

func (d *captureClaudeDriver) GetStatus() driver.SessionStatus {
	if d.status.Load() {
		return driver.StatusRunning
	}
	return driver.StatusTerminated
}

func (c *claudeCapture) capturedArgs(index int) []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return slices.Clone(c.configs[index].Args)
}

//nolint:gocognit,govet,golines // One lifecycle integration keeps registration and per-card resume assertions together.
func TestClaudeResumeLifecycleUsesSelectedNativeID(t *testing.T) {
	repo, workspaceID := workspaceValidationRepository(t)
	registry := driver.NewBuiltinRegistry()
	capture := &claudeCapture{}
	registry.Register("claude", func() driver.AgentDriver {
		return &captureClaudeDriver{ClaudeCodeDriver: driver.NewClaudeCodeDriver(), capture: capture}
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
		if err := runtime.Shutdown(ctx); err != nil {
			t.Errorf("shutdown MCP runtime: %v", err)
		}
	}()
	manager.SetMCPEndpoint(runtime.Endpoint())

	card := config.IDCard{ID: "claude-card", Name: "Claude", Driver: "claude", Model: "sonnet", WorkspaceID: workspaceID}
	sessionID, err := manager.Launch(card, nil, nil, config.BuiltinMCPs())
	if err != nil {
		t.Fatalf("fresh launch: %v", err)
	}
	wantNew := []string{"--model", "sonnet"}
	if got := capture.capturedArgs(0); !slices.Equal(got, wantNew) {
		t.Fatalf("fresh Claude args = %#v, want %#v", got, wantNew)
	}
	snapshot, found := manager.GetSession(sessionID)
	if !found {
		t.Fatal("fresh session was not published")
	}
	originalWorkingDir := snapshot.WorkingDir
	markerName := "workspace-marker"
	if err := os.WriteFile(filepath.Join(originalWorkingDir, markerName), []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}
	assertWorkspace := func(stage string) {
		t.Helper()
		current, found := manager.GetSession(sessionID)
		if !found {
			t.Fatalf("%s: session was not published", stage)
		}
		if current.WorkingDir != originalWorkingDir {
			t.Fatalf("%s: working directory = %q, want original %q", stage, current.WorkingDir, originalWorkingDir)
		}
		content, err := os.ReadFile(filepath.Join(current.WorkingDir, markerName))
		if err != nil || string(content) != "keep" {
			t.Fatalf("%s: workspace marker = %q, %v", stage, content, err)
		}
	}
	token := tokenFromConfig(t, driver.MCPConfigPath(snapshot.WorkingDir))
	if _, err := manager.RegisterConversation(token, "6cbf35c6-7e9d-4cb5-aefc-0ad55096ed8d", "Saved work"); err != nil {
		t.Fatalf("register native Claude conversation: %v", err)
	}
	if err := manager.Kill(sessionID); err != nil {
		t.Fatalf("stop fresh activation: %v", err)
	}
	if err := manager.RelaunchWithChoice(sessionID, card, session.ConversationChoice{ID: "6cbf35c6-7e9d-4cb5-aefc-0ad55096ed8d"}); err != nil {
		t.Fatalf("resume saved Claude conversation: %v", err)
	}
	wantResume := []string{"--model", "sonnet", "--resume", "6cbf35c6-7e9d-4cb5-aefc-0ad55096ed8d"}
	if got := capture.capturedArgs(1); !slices.Equal(got, wantResume) {
		t.Fatalf("resumed Claude args = %#v, want %#v", got, wantResume)
	}
	assertWorkspace("saved-conversation relaunch")
	if err := manager.Kill(sessionID); err != nil {
		t.Fatalf("stop resumed activation: %v", err)
	}
	if err := manager.RelaunchWithChoice(sessionID, card, session.ConversationChoice{New: true}); err != nil {
		t.Fatalf("start new Claude conversation: %v", err)
	}
	if got := capture.capturedArgs(2); !slices.Equal(got, wantNew) {
		t.Fatalf("new Claude args = %#v, want %#v", got, wantNew)
	}
	assertWorkspace("new-conversation relaunch")
	if err := manager.Kill(sessionID); err != nil {
		t.Fatalf("stop new activation: %v", err)
	}
	cardB := config.IDCard{ID: "claude-card-b", Name: "Claude B", Driver: "claude", Model: "sonnet", WorkspaceID: workspaceID}
	if err := manager.SwitchAgentWithChoice(sessionID, cardB, nil, nil, config.BuiltinMCPs(), session.ConversationChoice{New: true}); err != nil {
		t.Fatalf("switch to second Claude card: %v", err)
	}
	snapshot, found = manager.GetSession(sessionID)
	if !found {
		t.Fatal("second card session was not published")
	}
	assertWorkspace("switch to second card")
	secondToken := tokenFromConfig(t, driver.MCPConfigPath(snapshot.WorkingDir))
	secondID := "2af0f23d-6595-4ab2-8046-08642e1875b3"
	if _, err := manager.RegisterConversation(secondToken, secondID, "Other card"); err != nil {
		t.Fatalf("register second card conversation: %v", err)
	}
	if err := manager.Kill(sessionID); err != nil {
		t.Fatalf("stop second card activation: %v", err)
	}
	if err := manager.SwitchAgentWithChoice(sessionID, card, nil, nil, config.BuiltinMCPs(), session.ConversationChoice{ID: "6cbf35c6-7e9d-4cb5-aefc-0ad55096ed8d"}); err != nil {
		t.Fatalf("switch back to saved Claude conversation: %v", err)
	}
	if got := capture.capturedArgs(4); !slices.Equal(got, wantResume) {
		t.Fatalf("A→B→A Claude args = %#v, want %#v", got, wantResume)
	}
	assertWorkspace("switch back to first card")
	snapshot, found = manager.GetSession(sessionID)
	if !found || len(snapshot.Conversations[card.ID]) != 1 || len(snapshot.Conversations[cardB.ID]) != 1 ||
		snapshot.Conversations[card.ID][0].ID == snapshot.Conversations[cardB.ID][0].ID {
		t.Fatalf("conversation bindings were not kept per card: %#v", snapshot.Conversations)
	}
}
