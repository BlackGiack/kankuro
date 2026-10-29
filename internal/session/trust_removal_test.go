package session_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/driver"
	"github.com/BlackGiack/kankuro/internal/session"
)

type trustTestDriver struct {
	*driver.GenericCLIDriver

	prepare    func(context.Context, string) error
	spawnCheck func(string) error
	fail       bool
}

func (d *trustTestDriver) PrepareWorkspace(ctx context.Context, path string) error {
	return d.prepare(ctx, path)
}
func (d *trustTestDriver) Spawn(cfg driver.AgentConfig) (*driver.SessionHandle, error) {
	if d.spawnCheck != nil {
		if err := d.spawnCheck(cfg.WorkingDir); err != nil {
			return nil, err
		}
	}
	if d.fail {
		return nil, errors.New("spawn failed")
	}
	cfg.Driver = "cat"
	return d.GenericCLIDriver.Spawn(cfg)
}

//nolint:govet,gocognit,gocyclo,cyclop // Covers both trust grants, restart, cleanup failure and retry.
func TestTemporarySessionTrustCleanupSurvivesRestartAndCanRetry(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	repo, workspaceID := workspaceValidationRepository(t)
	registry := driver.NewRegistry()
	registry.Register("codex-test", func() driver.AgentDriver {
		return &trustTestDriver{GenericCLIDriver: driver.NewGenericCLIDriver(), prepare: driver.NewCodexCLIDriver().PrepareWorkspace}
	})
	registry.Register("claude-test", func() driver.AgentDriver {
		return &trustTestDriver{GenericCLIDriver: driver.NewGenericCLIDriver(), prepare: driver.NewClaudeCodeDriver().PrepareWorkspace,
			spawnCheck: func(path string) error {
				data, err := os.ReadFile(filepath.Join(home, ".claude.json"))
				if err != nil {
					return err
				}
				var cfg struct {
					Projects map[string]map[string]json.RawMessage `json:"projects"`
				}
				if err = json.Unmarshal(data, &cfg); err != nil {
					return err
				}
				if string(cfg.Projects[path]["hasTrustDialogAccepted"]) != "true" {
					return errors.New("Claude trust absent before spawn")
				}
				return nil
			}}
	})
	legacy := filepath.Join(home, ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(legacy), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(legacy, []byte("{legacy-invalid"), 0600); err != nil {
		t.Fatal(err)
	}
	manager := newTestManager(t, repo.Dir(), repo, session.WithDriverRegistry(registry))
	card := config.IDCard{ID: "a", Name: "a", Driver: "codex-test", WorkspaceID: workspaceID, UseTmpWorkspace: true}
	id, err := manager.Launch(card, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	snap, ok := manager.GetSession(id)
	if !ok {
		t.Fatal("missing session")
	}
	if err := manager.Kill(id); err != nil {
		t.Fatal(err)
	}
	claudePath := filepath.Join(home, ".claude.json")
	seed := `{"root":"preserve","projects":{"/other":{"hasTrustDialogAccepted":true},` +
		strconv.Quote(snap.WorkingDir) + `:{"keep":"metadata"}}}`
	if err := os.WriteFile(claudePath, []byte(seed), 0600); err != nil {
		t.Fatal(err)
	}
	card.ID = "b"
	card.Driver = "claude-test"
	if err := manager.SwitchAgent(id, card, nil, nil, nil); err != nil {
		t.Fatal(err)
	}
	var before struct {
		Projects map[string]map[string]json.RawMessage `json:"projects"`
	}
	claudeData, err := os.ReadFile(claudePath)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(claudeData, &before); err != nil {
		t.Fatal(err)
	}
	if string(before.Projects[snap.WorkingDir]["hasTrustDialogAccepted"]) != "true" || before.Projects[snap.OriginalDir] != nil {
		t.Fatalf("wrong workspace trust: %s", claudeData)
	}
	if err := manager.Kill(id); err != nil {
		t.Fatal(err)
	}
	manager.Shutdown()
	restored := newTestManager(t, repo.Dir(), repo)
	valid, err := os.ReadFile(claudePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(claudePath, []byte("{"), 0600); err != nil {
		t.Fatal(err)
	}
	result, err := restored.RemoveDetailed(id)
	if !errors.Is(err, session.ErrTrustCleanup) || result.InMemoryRemoved {
		t.Fatalf("expected retryable failure: %#v %v", result, err)
	}
	if _, ok := restored.GetSession(id); !ok {
		t.Fatal("failed deletion discarded session")
	}
	if _, err := os.Stat(snap.WorkingDir); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(claudePath, valid, 0600); err != nil {
		t.Fatal(err)
	}
	result, err = restored.RemoveDetailed(id)
	if err != nil || !result.InMemoryRemoved || !result.TemporaryWorkspaceRemoved {
		t.Fatalf("retry failed: %#v %v", result, err)
	}
	codex, err := os.ReadFile(filepath.Join(home, ".codex/config.toml"))
	if err != nil || strings.Contains(string(codex), snap.WorkingDir) {
		t.Fatalf("Codex trust remains: %v", err)
	}
	cleaned, err := os.ReadFile(claudePath)
	if err != nil {
		t.Fatal(err)
	}
	var after struct {
		Root     string                                `json:"root"`
		Projects map[string]map[string]json.RawMessage `json:"projects"`
	}
	if err = json.Unmarshal(cleaned, &after); err != nil {
		t.Fatal(err)
	}
	if after.Root != "preserve" || after.Projects["/other"] == nil ||
		after.Projects[snap.WorkingDir]["hasTrustDialogAccepted"] != nil ||
		string(after.Projects[snap.WorkingDir]["keep"]) != `"metadata"` {
		t.Fatalf("cleanup lost project data: %s", cleaned)
	}
	legacyData, err := os.ReadFile(legacy)
	if err != nil || string(legacyData) != "{legacy-invalid" {
		t.Fatalf("legacy settings changed: %q %v", legacyData, err)
	}
}

//nolint:govet // Error variables are scoped to individual test assertions.
func TestNonTemporarySessionDeletionPreservesTrust(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	repo, workspaceID := workspaceValidationRepository(t)
	manager := newTestManager(t, repo.Dir(), repo)
	id, err := manager.Launch(config.IDCard{ID: "a", Name: "a", Driver: "cat", WorkspaceID: workspaceID}, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	snap, _ := manager.GetSession(id)
	if err := driver.NewCodexCLIDriver().PrepareWorkspace(context.Background(), snap.WorkingDir); err != nil {
		t.Fatal(err)
	}
	if err := driver.NewClaudeCodeDriver().PrepareWorkspace(context.Background(), snap.WorkingDir); err != nil {
		t.Fatal(err)
	}
	if err := manager.Kill(id); err != nil {
		t.Fatal(err)
	}
	if err := manager.Remove(id); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(home, ".codex/config.toml"))
	if err != nil || !strings.Contains(string(data), snap.WorkingDir) {
		t.Fatal("non-temporary trust was removed")
	}
	claude, err := os.ReadFile(filepath.Join(home, ".claude.json"))
	if err != nil || !strings.Contains(string(claude), "hasTrustDialogAccepted") || !strings.Contains(string(claude), snap.WorkingDir) {
		t.Fatal("non-temporary Claude trust was removed")
	}
}

func TestFailedTemporaryLaunchRemovesTrust(t *testing.T) {
	for _, tc := range []struct {
		name, path, grant string
		prepare           func(context.Context, string) error
	}{
		{"codex", ".codex/config.toml", "trust_level", driver.NewCodexCLIDriver().PrepareWorkspace},
		{"claude", ".claude.json", "hasTrustDialogAccepted", driver.NewClaudeCodeDriver().PrepareWorkspace},
	} {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			repo, workspaceID := workspaceValidationRepository(t)
			registry := driver.NewRegistry()
			registry.Register("fail", func() driver.AgentDriver {
				return &trustTestDriver{GenericCLIDriver: driver.NewGenericCLIDriver(), prepare: tc.prepare, fail: true}
			})
			manager := newTestManager(t, repo.Dir(), repo, session.WithDriverRegistry(registry))
			_, err := manager.Launch(
				config.IDCard{ID: "a", Name: "a", Driver: "fail", WorkspaceID: workspaceID, UseTmpWorkspace: true},
				nil,
				nil,
				nil,
			)
			if err == nil {
				t.Fatal("expected spawn failure")
			}
			data, err := os.ReadFile(filepath.Join(home, tc.path))
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(data), tc.grant) {
				t.Fatalf("failed launch left trust grant: %s", data)
			}
		})
	}
}
