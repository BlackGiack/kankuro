package app_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/app"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
)

func TestTerminalCommandHelperProcess(t *testing.T) {
	if os.Getenv("KANKURO_TERMINAL_HELPER") != "1" {
		return
	}
	record, err := json.Marshal(struct {
		Args []string `json:"args"`
		Dir  string   `json:"dir"`
	}{Args: os.Args, Dir: mustTerminalHelperWorkingDir()})
	if err != nil {
		t.Fatal(err)
	}
	writeErr := os.WriteFile(os.Getenv("KANKURO_TERMINAL_RECORD"), record, 0o600)
	if writeErr != nil {
		t.Fatal(writeErr)
	}
}

func mustTerminalHelperWorkingDir() string {
	workingDir, _ := os.Getwd()
	return workingDir
}

func TestTerminalLaunchThroughDashboard(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	sm := newTestManager(t, repo.Dir(), repo)
	workspace := t.TempDir()
	if err = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: workspace}}); err != nil {
		t.Fatal(err)
	}
	workspaces, err := repo.ListWorkspaces()
	if err != nil {
		t.Fatal(err)
	}
	card, err := repo.CreateIDCard(config.IDCard{
		Name: "terminal-test", Driver: "cat", WorkspaceID: workspaces[0].ID,
		UseTmpWorkspace: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	sessionID, err := sm.Launch(card, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	coord := app.NewCoordinator(repo, sm)
	coord.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	coord.Update(coord.RefreshData()())
	snapshot, ok := sm.GetSession(sessionID)
	if !ok {
		t.Fatal("session disappeared")
	}
	if snapshot.WorkingDir == workspace {
		t.Fatal("expected temporary session workspace")
	}
	if snapshot.OriginalDir != workspace {
		t.Fatalf("session original directory %q, want source workspace %q", snapshot.OriginalDir, workspace)
	}
	recordPath := filepath.Join(t.TempDir(), "terminal-launch.json")
	t.Setenv("KANKURO_TERMINAL_HELPER", "1")
	t.Setenv("KANKURO_TERMINAL_RECORD", recordPath)
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	settings := config.DefaultAppSettings()
	settings.TerminalCommand = fmt.Sprintf("\"%s\" -test.run=TestTerminalCommandHelperProcess -- \"$PATH\"", binary)
	coord.Update(messages.SettingsSavedMsg{Settings: settings})
	coord.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	_, command := coord.Update(tea.KeyPressMsg{Code: 0, Text: "t"})
	if command == nil {
		t.Fatal("terminal key did not return a command")
	}
	msg := command()
	if result, isTerminalResult := msg.(messages.TerminalLaunchedMsg); !isTerminalResult || result.Err != nil {
		t.Fatalf("terminal launch result %#v", msg)
	}
	coord.Update(msg)
	if coord.CurrentView() != app.ViewDashboard {
		t.Fatalf("successful launch navigated to %v", coord.CurrentView())
	}
	child := readTerminalLaunchRecord(t, recordPath)
	if child.Dir != snapshot.WorkingDir {
		t.Errorf("terminal working directory %q, want session workspace %q", child.Dir, snapshot.WorkingDir)
	}
	if child.Dir == snapshot.OriginalDir {
		t.Errorf("terminal ran in original workspace %q", child.Dir)
	}
	if !containsExact(child.Args, snapshot.WorkingDir) {
		t.Errorf("terminal arguments %q lack expanded session workspace %q", child.Args, snapshot.WorkingDir)
	}
}

func readTerminalLaunchRecord(t *testing.T, path string) struct {
	Args []string `json:"args"`
	Dir  string   `json:"dir"`
} {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		data, err := os.ReadFile(path)
		if err == nil {
			var result struct {
				Args []string `json:"args"`
				Dir  string   `json:"dir"`
			}
			decodeErr := json.Unmarshal(data, &result)
			if decodeErr == nil {
				return result
			}
			// The child writes this test record directly; a concurrent read may
			// observe the file between truncate and completion, especially under -race.
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("terminal helper did not write %s before timeout", path)
	return struct {
		Args []string `json:"args"`
		Dir  string   `json:"dir"`
	}{}
}

func containsExact(values []string, want string) bool {
	return slices.Contains(values, want)
}

func TestTerminalLaunchReportsConfigurationAndLookupErrors(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	sm := newTestManager(t, repo.Dir(), repo)
	coord := app.NewCoordinator(repo, sm)
	callbacks := coord.DashboardCallbacks()
	missing := callbacks.OnLaunchTerminal("missing")()
	if result := missing.(messages.TerminalLaunchedMsg); result.Err == nil || result.Err.Error() != "session not found: missing" {
		t.Fatalf("missing session result: %+v", result)
	}
	saveErr := repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: t.TempDir()}})
	if saveErr != nil {
		t.Fatal(saveErr)
	}
	workspaces, err := repo.ListWorkspaces()
	if err != nil {
		t.Fatal(err)
	}
	card, err := repo.CreateIDCard(config.IDCard{Name: "error-test", Driver: "cat", WorkspaceID: workspaces[0].ID})
	if err != nil {
		t.Fatal(err)
	}
	sessionID, err := sm.Launch(card, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	result := callbacks.OnLaunchTerminal(sessionID)().(messages.TerminalLaunchedMsg)
	if result.Err == nil || result.Err.Error() != "Terminal command not configured" {
		t.Fatalf("unconfigured terminal result: %+v", result)
	}
	settings := config.DefaultAppSettings()
	settings.TerminalCommand = "kankuro-terminal-executable-missing"
	coord.Update(messages.SettingsSavedMsg{Settings: settings})
	result = coord.DashboardCallbacks().OnLaunchTerminal(sessionID)().(messages.TerminalLaunchedMsg)
	if result.Err == nil {
		t.Fatal("expected command spawn failure")
	}
	coord.Update(result)
	if coord.Err() == nil {
		t.Fatal("spawn failure was not shown in the coordinator error banner")
	}
	settings.TerminalCommand = "/bin/true"
	coord.Update(messages.SettingsSavedMsg{Settings: settings})
	snapshot, ok := sm.GetSession(sessionID)
	if !ok {
		t.Fatal("session disappeared")
	}
	removeErr := os.RemoveAll(snapshot.WorkingDir)
	if removeErr != nil {
		t.Fatal(removeErr)
	}
	result = coord.DashboardCallbacks().OnLaunchTerminal(sessionID)().(messages.TerminalLaunchedMsg)
	if result.Err == nil {
		t.Fatal("expected missing workspace error")
	}
}
