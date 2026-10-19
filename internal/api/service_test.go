package api //nolint:testpackage // private resolver is tested alongside its public callers.

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/BlackGiack/kankuro/internal/testenv"
)

func TestMain(m *testing.M) { os.Exit(testenv.Run(m, "api")) }

type testRepository struct {
	*config.FileRepository

	cardsErr, skillsErr, personasErr, mcpsErr, deleteErr error
	cardReads, skillReads, personaReads, mcpReads        int
}

func (r *testRepository) ListIDCards() ([]config.IDCard, error) {
	r.cardReads++
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
	r.skillReads++
	if r.skillsErr != nil {
		return nil, r.skillsErr
	}
	return r.FileRepository.ListSkills()
}
func (r *testRepository) ListPersonas() ([]config.Persona, error) {
	r.personaReads++
	if r.personasErr != nil {
		return nil, r.personasErr
	}
	return r.FileRepository.ListPersonas()
}
func (r *testRepository) ListMCPs() ([]config.MCPServer, error) {
	r.mcpReads++
	if r.mcpsErr != nil {
		return nil, r.mcpsErr
	}
	return r.FileRepository.ListMCPs()
}

//nolint:govet // short declarations localize setup failures in this test helper.
func newServiceFixture(t *testing.T) (*Service, *testRepository, *session.Manager, string, string) {
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
	sessions := session.NewManager(repoDir, repo)
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
	if countMCPCLI(withMCP.Skills) != 1 {
		t.Fatalf("mcp-cli access skill count = %d, want one", countMCPCLI(withMCP.Skills))
	}
	withMCPAndSkill, err := service.resolveCapabilities(config.IDCard{
		MCPIDs: []string{config.BuiltinMCPKankuro}, SkillIDs: []string{config.BuiltinSkillMCPCLI},
	})
	if err != nil {
		t.Fatal(err)
	}
	if countMCPCLI(withMCPAndSkill.Skills) != 1 {
		t.Fatalf("explicit mcp-cli skill duplicated: %#v", withMCPAndSkill.Skills)
	}
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
			api, repo, sessions, _, _ := newServiceFixture(t)
			want := errors.New("injected " + tc.name + " error")
			switch tc.registry {
			case "skills":
				repo.skillsErr = want
			case "personas":
				repo.personasErr = want
			case "mcps":
				repo.mcpsErr = want
			}
			_, err := api.Launch(config.IDCard{}) // empty selections still read each registry.
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
	selected.ID, selected.Name = "snapshot-id", "selected-snapshot"
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
