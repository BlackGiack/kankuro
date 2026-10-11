package session_test

import (
	"kankuro/internal/config"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSessionAgentsMD(t *testing.T) {
	repo, _ := config.NewFileRepository(t.TempDir())
	wd := t.TempDir()
	_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	sm := newTestManager(t, repo.Dir(), repo)

	// Pre-create AGENTS.md in original workspace
	originalContent := "# Original Content"
	if errWrite := os.WriteFile(filepath.Join(wd, "AGENTS.md"), []byte(originalContent), 0o600); errWrite != nil {
		t.Fatalf("write original agents.md: %v", errWrite)
	}

	card := config.IDCard{
		ID:              "card-agents-md",
		Name:            "agents-md",
		Driver:          "cat",
		WorkspaceID:     workspaces[0].ID,
		UseTmpWorkspace: true,
		AgentsMD:        "# First Content",
	}

	sid, err := sm.Launch(card, nil, nil, nil)
	if err != nil {
		t.Fatalf("launch: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	s, ok := sm.GetSession(sid)
	if !ok {
		t.Fatalf("session not found")
	}
	finalWD := s.WorkingDir
	content, errRead := os.ReadFile(filepath.Join(finalWD, "AGENTS.md"))
	if errRead != nil {
		t.Fatalf("read AGENTS.md: %v", errRead)
	}
	// Should be appended because it's a new session and it existed in original wd
	expectedFirst := "# Original Content\n\n# First Content"
	if string(content) != expectedFirst {
		t.Fatalf("expected %q, got %q", expectedFirst, string(content))
	}

	if errKill := sm.Kill(sid); errKill != nil {
		t.Fatalf("kill: %v", errKill)
	}
	time.Sleep(100 * time.Millisecond)

	// Switch agent with different content -> should REPLACE
	card.AgentsMD = "# Second Content"
	if errSwitch := sm.SwitchAgent(sid, card, nil, nil, nil); errSwitch != nil {
		t.Fatalf("switch agent: %v", errSwitch)
	}
	time.Sleep(100 * time.Millisecond)

	content, errRead = os.ReadFile(filepath.Join(finalWD, "AGENTS.md"))
	if errRead != nil {
		t.Fatalf("read AGENTS.md after switch: %v", errRead)
	}
	expectedSecond := "# Second Content"
	if string(content) != expectedSecond {
		t.Fatalf("expected %q, got %q", expectedSecond, string(content))
	}

	if errKillAfter := sm.Kill(sid); errKillAfter != nil {
		t.Fatalf("kill after switch: %v", errKillAfter)
	}

	// Test path option
	tmpFile := filepath.Join(t.TempDir(), "external_agents.md")
	expectedPathContent := "# Content from file"
	if errWrite := os.WriteFile(tmpFile, []byte(expectedPathContent), 0o600); errWrite != nil {
		t.Fatalf("write tmp file: %v", errWrite)
	}

	card.AgentsMDPreferPath = true
	card.AgentsMDPath = tmpFile
	wd2 := t.TempDir()
	_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{
		{Path: wd, IsRoot: false},
		{Path: wd2, IsRoot: false},
	})
	ws2, _ := repo.GetWorkspaceByPath(wd2)
	card.WorkspaceID = ws2.ID
	sid2, errLaunch := sm.Launch(card, nil, nil, nil)
	if errLaunch != nil {
		t.Fatalf("launch with path: %v", errLaunch)
	}
	time.Sleep(100 * time.Millisecond)

	s2, ok := sm.GetSession(sid2)
	if !ok {
		t.Fatalf("session 2 not found")
	}
	finalWD2 := s2.WorkingDir
	content, errRead = os.ReadFile(filepath.Join(finalWD2, "AGENTS.md"))
	if errRead != nil {
		t.Fatalf("read AGENTS.md from path: %v", errRead)
	}
	// Should NOT append here because we didn't pre-create AGENTS.md in a fresh WD for sid2
	if string(content) != expectedPathContent {
		t.Fatalf("expected %q, got %q", expectedPathContent, string(content))
	}

	if errKill2 := sm.Kill(sid2); errKill2 != nil {
		t.Fatalf("kill: %v", errKill2)
	}
}

func TestLaunchCopiesSubWorkspaceAndRetainsManagedEdits(t *testing.T) {
	repo, _ := config.NewFileRepository(t.TempDir())
	mainDir, subDir := t.TempDir(), t.TempDir()
	if err := os.WriteFile(filepath.Join(subDir, "source.txt"), []byte("source"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: mainDir}, {Path: subDir}}); err != nil {
		t.Fatal(err)
	}
	workspaces, listErr := repo.ListWorkspaces()
	if listErr != nil {
		t.Fatal(listErr)
	}
	var mainID, subID string
	for _, workspace := range workspaces {
		if workspace.Path == mainDir {
			mainID = workspace.ID
		}
		if workspace.Path == subDir {
			subID = workspace.ID
		}
	}
	sm := newTestManager(t, repo.Dir(), repo)
	sid, launchErr := sm.Launch(config.IDCard{
		ID: "sub-card", Name: "sub", Driver: "cat", WorkspaceID: mainID,
		SubWorkspaceIDs: []string{subID},
	}, nil, nil, nil)
	if launchErr != nil {
		t.Fatalf("launch: %v", launchErr)
	}
	snapshot, ok := sm.GetSession(sid)
	if !ok {
		t.Fatal("session missing")
	}
	copyPath := filepath.Join(snapshot.WorkingDir, "tmp", filepath.Base(subDir))
	if _, statErr := os.Stat(filepath.Join(copyPath, "source.txt")); statErr != nil {
		t.Fatalf("copied source: %v", statErr)
	}
	edit := filepath.Join(copyPath, "agent.txt")
	if writeErr := os.WriteFile(edit, []byte("agent edit"), 0o600); writeErr != nil {
		t.Fatal(writeErr)
	}
	if killErr := sm.Kill(sid); killErr != nil {
		t.Fatal(killErr)
	}
	relaunchErr := sm.Relaunch(sid, config.IDCard{
		ID: "sub-card", Name: "sub", Driver: "cat", WorkspaceID: mainID,
		SubWorkspaceIDs: []string{subID},
	}, nil, nil, nil)
	if relaunchErr != nil {
		t.Fatalf("relaunch: %v", relaunchErr)
	}
	if content, readErr := os.ReadFile(edit); readErr != nil || string(content) != "agent edit" {
		t.Fatalf("managed edit not retained: %q, %v", content, readErr)
	}
	if killErr := sm.Kill(sid); killErr != nil {
		t.Fatal(killErr)
	}
}

func TestFailedTemporaryLaunchRemovesSessionRoot(t *testing.T) {
	repo, _ := config.NewFileRepository(t.TempDir())
	mainDir, subDir, tmpRoot := t.TempDir(), t.TempDir(), t.TempDir()
	t.Setenv("KANKURO_TMP_DIR", tmpRoot)
	if err := repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: mainDir}, {Path: subDir}}); err != nil {
		t.Fatal(err)
	}
	workspaces, err := repo.ListWorkspaces()
	if err != nil {
		t.Fatal(err)
	}
	var mainID, subID string
	for _, workspace := range workspaces {
		if workspace.Path == mainDir {
			mainID = workspace.ID
		}
		if workspace.Path == subDir {
			subID = workspace.ID
		}
	}
	sm := newTestManager(t, repo.Dir(), repo)
	for _, card := range []config.IDCard{
		{ID: "missing-cli", Name: "missing", Driver: filepath.Join(tmpRoot, "no-such-cli"), WorkspaceID: mainID, UseTmpWorkspace: true, SubWorkspaceIDs: []string{subID}},
		{ID: "missing-source", Name: "source", Driver: "cat", WorkspaceID: mainID, UseTmpWorkspace: true, SubWorkspaceIDs: []string{"missing-workspace"}},
	} {
		sid, launchErr := sm.Launch(card, nil, nil, nil)
		if launchErr == nil {
			t.Fatalf("Launch(%s) unexpectedly succeeded", card.ID)
		}
		if _, statErr := os.Lstat(filepath.Join(tmpRoot, sid)); !os.IsNotExist(statErr) {
			t.Fatalf("failed launch left temporary session root: %v", statErr)
		}
	}
}

func TestTemporaryLaunchSyncFailureRemovesSessionRoot(t *testing.T) {
	repo, _ := config.NewFileRepository(t.TempDir())
	mainDir, tmpRoot := t.TempDir(), t.TempDir()
	t.Setenv("KANKURO_TMP_DIR", tmpRoot)
	if err := os.Mkdir(filepath.Join(mainDir, "AGENTS.md"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: mainDir}}); err != nil {
		t.Fatal(err)
	}
	workspaces, err := repo.ListWorkspaces()
	if err != nil {
		t.Fatal(err)
	}
	sm := newTestManager(t, repo.Dir(), repo)
	sid, launchErr := sm.Launch(config.IDCard{
		ID: "sync-failure", Name: "sync", Driver: "generic", WorkspaceID: workspaces[0].ID,
		UseTmpWorkspace: true, AgentsMD: "instruction",
	}, nil, nil, nil)
	if launchErr == nil {
		t.Fatal("expected instruction-file sync failure")
	}
	if _, statErr := os.Lstat(filepath.Join(tmpRoot, sid)); !os.IsNotExist(statErr) {
		t.Fatalf("sync failure left temporary session root: %v", statErr)
	}
}

func TestSymlinkedSubWorkspaceCopiesContentsWithoutLinkingSource(t *testing.T) {
	repo, _ := config.NewFileRepository(t.TempDir())
	mainDir, sourceDir, aliasDir := t.TempDir(), t.TempDir(), filepath.Join(t.TempDir(), "source-alias")
	if err := os.WriteFile(filepath.Join(sourceDir, "data.txt"), []byte("original"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(sourceDir, aliasDir); err != nil {
		t.Fatal(err)
	}
	if err := repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: mainDir}, {Path: aliasDir}}); err != nil {
		t.Fatal(err)
	}
	workspaces, err := repo.ListWorkspaces()
	if err != nil {
		t.Fatal(err)
	}
	var mainID, subID string
	for _, workspace := range workspaces {
		if workspace.Path == mainDir {
			mainID = workspace.ID
		}
		if workspace.Path == aliasDir {
			subID = workspace.ID
		}
	}
	sm := newTestManager(t, repo.Dir(), repo)
	card := config.IDCard{ID: "symlink-sub", Name: "symlink", Driver: "cat", WorkspaceID: mainID, SubWorkspaceIDs: []string{subID}}
	sid, err := sm.Launch(card, nil, nil, nil)
	if err != nil {
		t.Fatalf("launch: %v", err)
	}
	snapshot, _ := sm.GetSession(sid)
	copyDir := filepath.Join(snapshot.WorkingDir, "tmp", filepath.Base(aliasDir))
	info, err := os.Lstat(copyDir)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("copy is not a real directory: %v, %v", info, err)
	}
	if writeErr := os.WriteFile(filepath.Join(copyDir, "data.txt"), []byte("edited"), 0o600); writeErr != nil {
		t.Fatal(writeErr)
	}
	original, err := os.ReadFile(filepath.Join(sourceDir, "data.txt"))
	if err != nil || string(original) != "original" {
		t.Fatalf("source was modified: %q, %v", original, err)
	}
	if killErr := sm.Kill(sid); killErr != nil {
		t.Fatal(killErr)
	}
	if relaunchErr := sm.Relaunch(sid, card, nil, nil, nil); relaunchErr != nil {
		t.Fatalf("relaunch managed symlink source: %v", relaunchErr)
	}
	if killErr := sm.Kill(sid); killErr != nil {
		t.Fatal(killErr)
	}
}

func TestSubWorkspaceContainmentResolvesSymlinkedAncestor(t *testing.T) {
	repo, _ := config.NewFileRepository(t.TempDir())
	realParent := t.TempDir()
	mainDir := filepath.Join(realParent, "main")
	if err := os.Mkdir(mainDir, 0o700); err != nil {
		t.Fatal(err)
	}
	aliasDir := filepath.Join(t.TempDir(), "parent-alias")
	if err := os.Symlink(realParent, aliasDir); err != nil {
		t.Fatal(err)
	}
	if err := repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: mainDir}, {Path: aliasDir}}); err != nil {
		t.Fatal(err)
	}
	workspaces, err := repo.ListWorkspaces()
	if err != nil {
		t.Fatal(err)
	}
	var mainID, parentID string
	for _, workspace := range workspaces {
		if workspace.Path == mainDir {
			mainID = workspace.ID
		}
		if workspace.Path == aliasDir {
			parentID = workspace.ID
		}
	}
	sm := newTestManager(t, repo.Dir(), repo)
	card := config.IDCard{
		ID: "contained", Name: "contained", Driver: "cat", WorkspaceID: mainID,
		SubWorkspaceIDs: []string{parentID},
	}
	if _, launchErr := sm.Launch(card, nil, nil, nil); launchErr == nil {
		t.Fatal("expected sub workspace containing the destination to be rejected")
	}
}

func TestSessionClaudeAgentsMD(t *testing.T) {
	// Setup dummy claude binary
	tmpPath := t.TempDir()
	claudePath := filepath.Join(tmpPath, "claude")
	if err := os.WriteFile(claudePath, []byte("#!/bin/sh\ncat"), 0o755); err != nil {
		t.Fatalf("failed to create dummy claude: %v", err)
	}
	oldPath := os.Getenv("PATH")
	t.Setenv("PATH", tmpPath+string(os.PathListSeparator)+oldPath)

	smDir := t.TempDir()
	repo, _ := config.NewFileRepository(smDir)
	wd := t.TempDir()
	_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	sm := newTestManager(t, repo.Dir(), repo)

	card := config.IDCard{
		ID:              "card-claude",
		Name:            "claude-test",
		Driver:          "claude", // Use Claude driver
		WorkspaceID:     workspaces[0].ID,
		UseTmpWorkspace: true,
		AgentsMD:        "# Claude Content",
	}

	sid, err := sm.Launch(card, nil, nil, nil)
	if err != nil {
		t.Fatalf("launch: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	s, ok := sm.GetSession(sid)
	if !ok {
		t.Fatalf("session not found")
	}
	finalWD := s.WorkingDir

	// Verify CLAUDE.md exists
	content, errRead := os.ReadFile(filepath.Join(finalWD, "CLAUDE.md"))
	if errRead != nil {
		t.Fatalf("read CLAUDE.md: %v", errRead)
	}
	if string(content) != "# Claude Content" {
		t.Fatalf("expected %q, got %q", "# Claude Content", string(content))
	}

	// Verify AGENTS.md does NOT exist
	if _, errStat := os.Stat(filepath.Join(finalWD, "AGENTS.md")); !os.IsNotExist(errStat) {
		t.Fatalf("expected AGENTS.md not to exist")
	}

	if errKill := sm.Kill(sid); errKill != nil {
		t.Fatalf("kill: %v", errKill)
	}
}

func TestSessionMCPs(t *testing.T) {
	repo, _ := config.NewFileRepository(t.TempDir())
	wd := t.TempDir()
	_ = repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd, IsRoot: false}})
	workspaces, _ := repo.ListWorkspaces()

	sm := newTestManager(t, repo.Dir(), repo)

	mcps := []config.MCPServer{
		{Name: "mcp1", Command: "node", Args: []string{"server.js"}},
		{Name: "mcp2", URL: "http://remote"},
	}

	card := config.IDCard{
		ID:              "card-mcp",
		Name:            "mcp-test",
		Driver:          "cat",
		WorkspaceID:     workspaces[0].ID,
		UseTmpWorkspace: true,
	}

	sid, err := sm.Launch(card, nil, nil, mcps)
	if err != nil {
		t.Fatalf("launch: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	s, ok := sm.GetSession(sid)
	if !ok {
		t.Fatalf("session not found")
	}
	finalWD := s.WorkingDir

	// Verify .kankuro/mcp-servers.json
	configPath := filepath.Join(finalWD, ".kankuro", "mcp-servers.json")
	if _, err = os.Stat(configPath); err != nil {
		t.Fatalf("mcp-servers.json missing: %v", err)
	}

	// Verify AGENTS.md
	agentsPath := filepath.Join(finalWD, "AGENTS.md")
	content, errRead := os.ReadFile(agentsPath)
	if errRead != nil {
		t.Fatalf("read AGENTS.md: %v", errRead)
	}
	expectedLine := `Available mcp: ["mcp1","mcp2"]`
	if !strings.Contains(string(content), expectedLine) {
		t.Fatalf("expected AGENTS.md to contain %q, got %q", expectedLine, string(content))
	}

	if errKill := sm.Kill(sid); errKill != nil {
		t.Fatalf("kill: %v", errKill)
	}
}
