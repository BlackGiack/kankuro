package api //nolint:testpackage // Authorization fixtures and repository probes are package-private.

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/BlackGiack/kankuro/internal/config"
)

type workspaceProbeRepository struct {
	*testRepository

	items []config.Workspace
	err   error
	reads int
}

func (r *workspaceProbeRepository) ListWorkspaces() ([]config.Workspace, error) {
	r.reads++
	return r.items, r.err
}

func TestListWorkspacesAuthorizationAndRegistry(t *testing.T) {
	skill, err := os.ReadFile("../../kankuro_skills/mcp-cli/SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	config.SetBuiltinMCPCLISkillContent(skill)
	t.Cleanup(func() { config.SetBuiltinMCPCLISkillContent(nil) })
	service, repo, manager, workspace, workspaceID := newServiceFixture(t)
	manager.SetMCPEndpoint("http://127.0.0.1:12345/mcp")
	callerCard := saveCard(
		t,
		repo.FileRepository,
		config.IDCard{
			Name:             "caller",
			Driver:           "cat",
			WorkspaceID:      workspaceID,
			DashboardControl: true,
			MCPIDs:           []string{config.BuiltinMCPKankuro},
		},
	)
	callerID, err := launchCardTest(service, callerCard)
	if err != nil {
		t.Fatal(err)
	}
	token := workspaceMCPToken(t, workspace)
	probe := &workspaceProbeRepository{testRepository: repo, items: []config.Workspace{
		{ID: "z", Name: "Same", Path: "/z"}, {ID: "a", Name: "Same", Path: "/a"}, {ID: "x", Name: "Alpha", Path: "/missing"},
	}}
	discovery := New(probe, manager)
	caller, items, err := discovery.ListWorkspaces(token)
	if err != nil || caller.SessionID != callerID {
		t.Fatalf("authorized list: caller %+v items %+v err %v", caller, items, err)
	}
	want := []WorkspaceDiscovery{
		{ID: "x", Name: "Alpha", Path: "/missing"},
		{ID: "a", Name: "Same", Path: "/a"},
		{ID: "z", Name: "Same", Path: "/z"},
	}
	if !reflect.DeepEqual(items, want) {
		t.Fatalf("workspaces = %+v, want %+v", items, want)
	}
	encoded, err := json.Marshal(items[0])
	if err != nil || string(encoded) != `{"id":"x","name":"Alpha","path":"/missing"}` {
		t.Fatalf("payload %s, %v", encoded, err)
	}
	probe.items = nil
	_, items, err = discovery.ListWorkspaces(token)
	if err != nil || items == nil || len(items) != 0 {
		t.Fatalf("empty list = %#v, %v", items, err)
	}
	failure := errors.New("registry failed")
	probe.err = failure
	if _, _, err = discovery.ListWorkspaces(token); !errors.Is(err, failure) {
		t.Fatalf("registry error = %v", err)
	}
	unprivileged := saveCard(
		t,
		repo.FileRepository,
		config.IDCard{
			Name:            "unprivileged",
			Driver:          "cat",
			WorkspaceID:     workspaceID,
			UseTmpWorkspace: true,
			MCPIDs:          []string{config.BuiltinMCPKankuro},
		},
	)
	unprivilegedID, launchErr := launchCardTest(service, unprivileged)
	if launchErr != nil {
		t.Fatal(launchErr)
	}
	snapshot, ok := manager.GetSession(unprivilegedID)
	if !ok {
		t.Fatal("unprivileged session missing")
	}
	unprivilegedToken := workspaceMCPToken(t, snapshot.WorkingDir)
	reads := probe.reads
	if _, _, err = discovery.ListWorkspaces(unprivilegedToken); err == nil || probe.reads != reads {
		t.Fatalf("unprivileged credential: %v; reads %d", err, probe.reads)
	}
	if _, _, err = discovery.ListWorkspaces("invalid"); err == nil || probe.reads != reads {
		t.Fatalf("invalid credential: %v; reads %d", err, probe.reads)
	}
	if killErr := manager.Kill(callerID); killErr != nil {
		t.Fatal(killErr)
	}
	if _, _, err = discovery.ListWorkspaces(token); err == nil || probe.reads != reads {
		t.Fatalf("revoked credential: %v; reads %d", err, probe.reads)
	}
}
