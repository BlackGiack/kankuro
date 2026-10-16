package app

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/driver"
	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

type coordinatorCodexDriver struct{ *driver.GenericCLIDriver }

func (d *coordinatorCodexDriver) Spawn(cfg driver.AgentConfig) (*driver.SessionHandle, error) {
	cfg.Driver = "cat"
	return d.GenericCLIDriver.Spawn(cfg)
}

//nolint:cyclop,gocognit,gocyclo,govet // Exercises refresh recovery for both relaunch and switch errors against real manager state.
func TestConversationChoiceRefreshesDriverAndRejectsStaleContext(t *testing.T) {
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
	registry := driver.NewBuiltinRegistry()
	registry.Register("codex", func() driver.AgentDriver {
		return &coordinatorCodexDriver{GenericCLIDriver: driver.NewGenericCLIDriver()}
	})
	manager := newTestManager(t, repo.Dir(), repo, session.WithDriverRegistry(registry))
	manager.SetMCPEndpoint("http://127.0.0.1:43210/mcp")
	card, err := repo.CreateIDCard(
		config.IDCard{ID: "conversation-card", Name: "Conversation card", Driver: "cat", WorkspaceID: workspaces[0].ID},
	)
	if err != nil {
		t.Fatal(err)
	}
	sessionID, err := manager.Launch(card, nil, nil, config.BuiltinMCPs())
	if err != nil {
		t.Fatal(err)
	}
	catSnapshot, _ := manager.GetSession(sessionID)
	if _, err := manager.RegisterConversation(
		readMCPToken(t, driver.MCPConfigPath(catSnapshot.WorkingDir)),
		"cat-native-id",
		"Cat conversation",
	); err != nil {
		t.Fatal(err)
	}
	if err := manager.Kill(sessionID); err != nil {
		t.Fatal(err)
	}
	card.Driver = "codex"
	if _, err := repo.UpdateIDCard(card); err != nil {
		t.Fatal(err)
	}
	if err := manager.SwitchAgentWithChoice(
		sessionID,
		card,
		nil,
		nil,
		config.BuiltinMCPs(),
		session.ConversationChoice{New: true},
	); err != nil {
		t.Fatal(err)
	}
	codexSnapshot, _ := manager.GetSession(sessionID)
	if _, err := manager.RegisterConversation(
		readMCPToken(t, driver.MCPConfigPath(codexSnapshot.WorkingDir)),
		"codex-native-id",
		"Codex conversation",
	); err != nil {
		t.Fatal(err)
	}
	if err := manager.Kill(sessionID); err != nil {
		t.Fatal(err)
	}

	coord := NewCoordinator(repo, manager)
	coord.Update(messages.DataLoadedMsg{Cards: []config.IDCard{card}, Sessions: manager.List()})
	request := coord.DashboardCallbacks().OnRelaunch(sessionID)
	if request == nil {
		t.Fatal("relaunch request returned no command")
	}
	coord.Update(request())
	choiceView := coord.Views()[ViewConversationChoice].(*views.ConversationChoiceView)
	choiceView.SetSize(80, 20)
	if selected, ok := choiceView.SelectedBinding(); !ok || selected.ID != "codex-native-id" {
		t.Fatalf("initial compatible selection = %#v, found=%v", selected, ok)
	}

	// An error-triggered refresh must re-read the destination driver and reselect
	// the saved conversation that is compatible with that refreshed driver.
	card.Driver = "cat"
	if _, err := repo.UpdateIDCard(card); err != nil {
		t.Fatal(err)
	}
	_, relaunchRefresh := coord.Update(messages.ErrorMsg{Err: errors.New("stale relaunch choice")})
	if relaunchRefresh == nil {
		t.Fatal("relaunch failure did not request refreshed context")
	}
	coord.Update(relaunchRefresh())
	selected, ok := choiceView.SelectedBinding()
	if !ok || selected.ID != "cat-native-id" || selected.Driver != "cat" {
		t.Fatalf("relaunch refresh selection = %#v, found=%v", selected, ok)
	}
	if !strings.Contains(choiceView.View(), "incompatible driver") {
		t.Fatal("picker did not mark the other driver's saved binding incompatible")
	}

	card.Driver = "codex"
	if _, err := repo.UpdateIDCard(card); err != nil {
		t.Fatal(err)
	}
	_, switchRefresh := coord.Update(messages.AgentSwitchedMsg{SessionID: sessionID, Err: errors.New("stale switch choice")})
	if switchRefresh == nil {
		t.Fatal("switch failure did not request refreshed context")
	}
	coord.Update(switchRefresh())
	selected, ok = choiceView.SelectedBinding()
	if !ok || selected.ID != "codex-native-id" || selected.Driver != "codex" {
		t.Fatalf("switch refresh selection = %#v, found=%v", selected, ok)
	}

	if err := repo.DeleteIDCard(card.ID); err != nil {
		t.Fatal(err)
	}
	coord.Update(messages.DataLoadedMsg{Cards: []config.IDCard{}, Sessions: manager.List()})
	if !strings.Contains(choiceView.View(), "destination card was removed") {
		t.Fatalf("removed destination did not leave an actionable picker error: %s", choiceView.View())
	}
	if _, cmd := choiceView.Update(tea.KeyPressMsg{Code: tea.KeyEnter}); cmd != nil {
		t.Fatal("stale picker submitted an activation after its destination card disappeared")
	}
	if err := manager.Remove(sessionID); err != nil {
		t.Fatal(err)
	}
	coord.Update(messages.SessionsRefreshedMsg{Sessions: manager.List()})
	if !strings.Contains(choiceView.View(), "selected session was removed") {
		t.Fatalf("removed session did not leave an actionable picker error: %s", choiceView.View())
	}
}

func readMCPToken(t *testing.T, path string) string {
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
