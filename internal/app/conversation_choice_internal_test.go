package app

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/api"
	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/configstartup"
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
	root := t.TempDir()
	if err := configstartup.PrepareStartup(root); err != nil {
		t.Fatal(err)
	}
	repo, err := config.NewFileRepository(root)
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
		config.IDCard{
			DashboardID: "default",
			ID:          "conversation-card",
			Name:        "Conversation card",
			Driver:      "cat",
			WorkspaceID: workspaces[0].ID,
			MCPIDs:      []string{config.BuiltinMCPKankuro},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	service := api.New(repo, manager)
	sessionID, err := service.LaunchInDashboard(context.Background(), "default", card.ID)
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
	if err := service.SwitchInDashboard(
		context.Background(), "default", sessionID, card.ID,
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

	preparation, err := service.ReadConversationPreparation(
		context.Background(),
		contracts.ConversationPreparationRequest{DashboardID: "default", SessionID: sessionID},
	)
	if err != nil || len(preparation.Options) != 2 {
		t.Fatalf("preparation: %+v %v", preparation, err)
	}
	originalName := preparation.Options[0].Name
	preparation.Options[0].Name = "caller mutation"
	again, err := service.ReadConversationPreparation(
		context.Background(),
		contracts.ConversationPreparationRequest{DashboardID: "default", SessionID: sessionID},
	)
	if err != nil || again.Options[0].Name != originalName {
		t.Fatalf("options not detached: %+v %v", again, err)
	}

	coord := NewCoordinator(manager, service)
	coord.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	coord.Update(
		messages.DashboardOpenRequestedMsg{DashboardID: "default", DashboardName: "Default", DashboardState: contracts.DashboardActive},
	)
	coord.Update(coord.refreshData()())
	request := coord.DashboardCallbacks().OnRelaunch(sessionID)
	if request == nil {
		t.Fatal("relaunch request returned no command")
	}
	coord.Update(request())
	choiceView := coord.Views()[ViewConversationChoice].(*views.ConversationChoiceView)
	choiceView.SetSize(80, 20)
	if selected, ok := choiceView.SelectedOption(); !ok || selected.NativeID != "codex-native-id" {
		t.Fatalf("initial compatible selection = %#v, found=%v", selected, ok)
	}

	// An error-triggered refresh must re-read the destination driver and reselect
	// the saved conversation that is compatible with that refreshed driver.
	card.Driver = "cat"
	if _, err := repo.UpdateIDCard(card); err != nil {
		t.Fatal(err)
	}
	relaunchRefresh := coord.refreshConversationChoiceContext()
	if relaunchRefresh == nil {
		t.Fatal("missing conversation refresh")
	}
	coord.Update(relaunchRefresh())
	if !strings.Contains(choiceView.View(), "no longer available or compatible") {
		t.Fatal("changed driver did not invalidate selected conversation")
	}
	if _, cmd := choiceView.Update(tea.KeyPressMsg{Code: tea.KeyEnter}); cmd != nil {
		t.Fatal("stale choice submitted")
	}
	selected, ok := choiceView.SelectedOption()
	if !ok || selected.NativeID != "cat-native-id" {
		t.Fatalf("compatible option: %+v", selected)
	}

	card.Driver = "codex"
	if _, err := repo.UpdateIDCard(card); err != nil {
		t.Fatal(err)
	}
	switchRefresh := coord.refreshConversationChoiceContext()
	if switchRefresh == nil {
		t.Fatal("missing second refresh")
	}
	coord.Update(switchRefresh())
	selected, ok = choiceView.SelectedOption()
	if !ok || selected.NativeID != "codex-native-id" {
		t.Fatalf("refreshed compatible option: %+v", selected)
	}

	preflight, err := service.PreflightCardDeletion("default", card.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.ConfirmCardDeletion(
		context.Background(),
		"default",
		card.ID,
		preflight.Token,
		preflight.RequiresForeignConfirmation,
	); err != nil {
		t.Fatal(err)
	}
	coord.Update(coord.refreshConversationChoiceContext()())
	if !strings.Contains(choiceView.View(), "card or dashboard is unavailable") {
		t.Fatalf("removed destination did not leave an actionable picker error: %s", choiceView.View())
	}
	if _, cmd := choiceView.Update(tea.KeyPressMsg{Code: tea.KeyEnter}); cmd != nil {
		t.Fatal("stale picker submitted an activation after its destination card disappeared")
	}
	if err := manager.Remove(sessionID); err != nil {
		t.Fatal(err)
	}
	coord.Update(coord.refreshConversationChoiceContext()())
	if !strings.Contains(choiceView.View(), "session is outside dashboard scope") {
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
