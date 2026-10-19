package session_test

import (
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/driver"
	"github.com/BlackGiack/kankuro/internal/session"
)

type workspaceValidationDriver struct {
	*driver.GenericCLIDriver

	spawns *atomic.Int32
}

func (d *workspaceValidationDriver) Spawn(cfg driver.AgentConfig) (*driver.SessionHandle, error) {
	d.spawns.Add(1)
	return d.GenericCLIDriver.Spawn(cfg)
}

//nolint:gocognit,govet // Each case verifies rejection, no mutation or spawn, and successful retry through the real lifecycle.
func TestActivationRejectsUnavailableSavedWorkspace(t *testing.T) {
	for _, test := range []struct {
		name        string
		useTmp      bool
		operation   string
		replacement string
	}{
		{"relaunch/missing/source", false, "relaunch", "missing"},
		{"relaunch/file/source", false, "relaunch", "file"},
		{"switch/missing/source", false, "switch", "missing"},
		{"switch/file/source", false, "switch", "file"},
		{"relaunch/missing/temporary", true, "relaunch", "missing"},
		{"relaunch/file/temporary", true, "relaunch", "file"},
		{"switch/missing/temporary", true, "switch", "missing"},
		{"switch/file/temporary", true, "switch", "file"},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo, workspaceID := workspaceValidationRepository(t)
			var spawns atomic.Int32
			registry := driver.NewRegistry()
			registry.Register("cat", func() driver.AgentDriver {
				return &workspaceValidationDriver{GenericCLIDriver: driver.NewGenericCLIDriver(), spawns: &spawns}
			})
			manager := newTestManager(t, repo.Dir(), repo, session.WithDriverRegistry(registry))
			card := config.IDCard{ID: "a", Name: "A", Driver: "cat", WorkspaceID: workspaceID, UseTmpWorkspace: test.useTmp}
			id, err := manager.Launch(card, nil, nil, nil)
			if err != nil {
				t.Fatal(err)
			}
			if err := manager.Kill(id); err != nil {
				t.Fatal(err)
			}
			before, found := manager.GetSession(id)
			if !found {
				t.Fatal("session missing")
			}
			backup := filepath.Join(t.TempDir(), "saved-workspace")
			if err := os.Rename(before.WorkingDir, backup); err != nil {
				t.Fatal(err)
			}
			if test.replacement == "file" {
				if err := os.WriteFile(before.WorkingDir, []byte("keep this file"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			activate := func() error { return manager.Relaunch(id, card) }
			if test.operation == "switch" {
				card.ID = "b"
				activate = func() error { return manager.SwitchAgent(id, card, nil, nil, nil) }
			}
			if err := activate(); err == nil || !strings.Contains(err.Error(), before.WorkingDir) {
				t.Fatalf("expected workspace error with saved path, got %v", err)
			}
			if spawns.Load() != 1 {
				t.Fatal("invalid workspace activation reached Spawn")
			}
			after, found := manager.GetSession(id)
			if !found || after.Status != session.StatusTerminated ||
				after.PID != before.PID || after.WorkingDir != before.WorkingDir {
				t.Fatalf("failed activation changed session identity or process: %#v", after)
			}
			if test.replacement == "missing" {
				if _, err := os.Lstat(before.WorkingDir); !os.IsNotExist(err) {
					t.Fatalf("missing workspace was recreated: %v", err)
				}
			} else {
				content, err := os.ReadFile(before.WorkingDir)
				if err != nil || string(content) != "keep this file" {
					t.Fatalf("replacement file changed: %q, %v", content, err)
				}
				if err := os.Remove(before.WorkingDir); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.Rename(backup, before.WorkingDir); err != nil {
				t.Fatal(err)
			}
			if err := activate(); err != nil {
				t.Fatalf("retry after restoring workspace: %v", err)
			}
			if spawns.Load() != 2 {
				t.Fatal("retry did not spawn exactly one agent")
			}
			content, err := os.ReadFile(filepath.Join(before.WorkingDir, "work.txt"))
			if err != nil || string(content) != "saved work" {
				t.Fatalf("saved work lost after retry: %q, %v", content, err)
			}
		})
	}
}

func workspaceValidationRepository(t *testing.T) (*config.FileRepository, string) {
	t.Helper()
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	workspace := t.TempDir()
	if writeErr := os.WriteFile(filepath.Join(workspace, "work.txt"), []byte("saved work"), 0o600); writeErr != nil {
		t.Fatal(writeErr)
	}
	if saveErr := repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: workspace}}); saveErr != nil {
		t.Fatal(saveErr)
	}
	workspaces, err := repo.ListWorkspaces()
	if err != nil || len(workspaces) != 1 {
		t.Fatalf("workspaces: %v", err)
	}
	return repo, workspaces[0].ID
}
