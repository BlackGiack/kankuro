package mcp_test

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BlackGiack/kankuro/internal/api"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/configstartup"
	"github.com/BlackGiack/kankuro/internal/mcp"
	"github.com/BlackGiack/kankuro/internal/session"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

//nolint:gocognit,gocyclo,cyclop // The scenario exercises both callers through the actual transport.
func TestDashboardScopeOverHTTP(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	for _, skill := range []struct {
		path string
		set  func([]byte)
	}{
		{"../../kankuro_skills/mcp-cli/SKILL.md", config.SetBuiltinMCPCLISkillContent},
		{"../../kankuro_skills/kankuro-puppet/SKILL.md", config.SetBuiltinKankuroPuppetSkillContent},
	} {
		data, err := os.ReadFile(skill.path)
		if err != nil {
			t.Fatal(err)
		}
		skill.set(data)
		t.Cleanup(func() { skill.set(nil) })
	}
	root := t.TempDir()
	t.Setenv("KANKURO_TMP_DIR", filepath.Join(root, "tmp"))
	if err := configstartup.PrepareStartup(root); err != nil {
		t.Fatal(err)
	}
	repo, err := config.NewFileRepository(root)
	if err != nil {
		t.Fatal(err)
	}
	second, err := repo.CreateDashboard(config.Dashboard{Name: "Second"})
	if err != nil {
		t.Fatal(err)
	}
	workspace := t.TempDir()
	if errAt50 := repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: workspace}}); errAt50 != nil {
		t.Fatal(errAt50)
	}
	sources, err := repo.ListWorkspaces()
	if err != nil || len(sources) != 1 {
		t.Fatalf("workspace: %v %v", sources, err)
	}
	card, err := repo.CreateIDCard(
		config.IDCard{
			Name:             "shared",
			Driver:           "cat",
			WorkspaceID:      sources[0].ID,
			DashboardID:      "default",
			Shared:           true,
			UseTmpWorkspace:  true,
			DashboardControl: true,
			MCPIDs:           []string{config.BuiltinMCPKankuro},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	manager := session.NewManager(root, repo, session.WithStartupDelay(0))
	t.Cleanup(func() { manager.Shutdown() })
	service := api.New(repo, manager)
	runtime, err := mcp.Start(service)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), testShutdownTimeout)
		defer cancel()
		if errAt84 := runtime.Shutdown(ctx); errAt84 != nil {
			t.Error(errAt84)
		}
	})
	manager.SetMCPEndpoint(runtime.Endpoint())
	ids := map[string][]string{}
	for _, dashboard := range []string{"default", second.ID} {
		for range 2 {
			id, errAt92 := service.LaunchInDashboard(context.Background(), dashboard, card.ID)
			if errAt92 != nil {
				t.Fatal(errAt92)
			}
			ids[dashboard] = append(ids[dashboard], id)
		}
	}
	for dashboard, sessionIDs := range ids {
		item, ok := manager.GetSession(sessionIDs[0])
		if !ok {
			t.Fatal("caller missing")
		}
		raw, errAt104 := os.ReadFile(filepath.Join(item.WorkingDir, ".kankuro", "mcp-servers.json"))
		if errAt104 != nil {
			t.Fatal(errAt104)
		}
		var activation struct {
			Servers map[string]struct {
				Headers map[string]string `json:"headers"`
			} `json:"mcpServers"`
		}
		if errAt104 = json.Unmarshal(raw, &activation); errAt104 != nil {
			t.Fatal(errAt104)
		}
		token := strings.TrimPrefix(activation.Servers["kankuro"].Headers["Authorization"], "Bearer ")
		if token == "" {
			t.Fatal("missing token")
		}
		client := sdk.NewClient(&sdk.Implementation{Name: "dashboard-scope-test", Version: "test"}, nil)
		transportHTTP := &http.Transport{DisableKeepAlives: true}
		ctx, cancel := context.WithTimeout(context.Background(), testShutdownTimeout)
		connection, errAt104 := client.Connect(
			ctx,
			&sdk.StreamableClientTransport{
				Endpoint:   runtime.Endpoint(),
				HTTPClient: &http.Client{Transport: authRoundTripper{next: transportHTTP, token: token}},
			},
			nil,
		)
		if errAt104 != nil {
			cancel()
			t.Fatal(errAt104)
		}
		result, errAt104 := connection.CallTool(ctx, &sdk.CallToolParams{Name: "list_sessions", Arguments: map[string]any{}})
		if errAt104 != nil || result.IsError {
			cancel()
			t.Fatalf("list_sessions: %v %#v", errAt104, result)
		}
		content, errAt104 := json.Marshal(result.StructuredContent)
		if errAt104 != nil {
			cancel()
			t.Fatal(errAt104)
		}
		var response struct {
			Items []struct {
				SessionID string `json:"session_id"`
			} `json:"items"`
		}
		if errAt150 := json.Unmarshal(content, &response); errAt150 != nil {
			t.Fatal(errAt150)
		}
		items, errAt104 := json.Marshal(response.Items)
		if errAt104 != nil {
			t.Fatal(errAt104)
		}
		if !strings.Contains(string(items), sessionIDs[1]) || strings.Contains(string(items), sessionIDs[0]) {
			t.Errorf("dashboard %s missing peer or contains caller: %s", dashboard, content)
		}
		for other, foreign := range ids {
			if other != dashboard {
				for _, id := range foreign {
					if strings.Contains(string(items), id) {
						t.Errorf("dashboard %s leaked %s", dashboard, id)
					}
				}
			}
		}
		closeTestClientSession(t, connection, transportHTTP)
		cancel()
	}
}
