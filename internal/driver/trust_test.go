//nolint:testpackage // Exercises configuration helpers alongside driver behavior.
package driver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/pelletier/go-toml/v2"
)

//nolint:gocognit,nestif,govet // Covers preservation across both configuration formats.
func TestTrustPreparation(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()
	workspace := t.TempDir()
	for _, tc := range []struct {
		name, path string
		prepare    func(context.Context, string) error
		initial    string
	}{
		{"codex", ".codex/config.toml", NewCodexCLIDriver().PrepareWorkspace, "model = 'example'\n[projects.'/other']\ntrust_level = 'trusted'\n"},
		{"claude", ".claude.json", NewClaudeCodeDriver().PrepareWorkspace, `{"other":{"keep":true},"projects":{"/other":{"hasTrustDialogAccepted":true}}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(os.Getenv("HOME"), tc.path)
			if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(tc.initial), 0640); err != nil {
				t.Fatal(err)
			}
			if err := tc.prepare(ctx, workspace); err != nil {
				t.Fatal(err)
			}
			first, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if err := tc.prepare(ctx, workspace); err != nil {
				t.Fatal(err)
			}
			second, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if string(first) != string(second) {
				t.Fatal("second update changed configuration")
			}
			info, err := os.Stat(path)
			if err != nil || info.Mode().Perm() != 0640 {
				t.Fatalf("permissions: %v %v", info, err)
			}
			if tc.name == "codex" {
				var cfg map[string]any
				if err := toml.Unmarshal(first, &cfg); err != nil {
					t.Fatal(err)
				}
				projects := cfg["projects"].(map[string]any)
				if cfg["model"] != "example" || projects["/other"] == nil ||
					projects[workspace].(map[string]any)["trust_level"] != "trusted" {
					t.Fatal("settings not preserved")
				}
			} else {
				var cfg map[string]json.RawMessage
				if err := json.Unmarshal(first, &cfg); err != nil {
					t.Fatal(err)
				}
				var projects map[string]map[string]json.RawMessage
				if err := json.Unmarshal(cfg["projects"], &projects); err != nil ||
					string(projects[workspace]["hasTrustDialogAccepted"]) != "true" || projects["/other"] == nil || cfg["other"] == nil {
					t.Fatal("configuration not preserved")
				}
			}
		})
	}
}

func TestTrustInvalidConfigurationPreserved(t *testing.T) {
	for _, tc := range []struct {
		name, path, content string
		prepare             func(context.Context, string) error
	}{
		{"invalid TOML", ".codex/config.toml", "[broken", NewCodexCLIDriver().PrepareWorkspace},
		{"invalid JSON", ".claude.json", "{broken", NewClaudeCodeDriver().PrepareWorkspace},
		{"wrong type", ".claude.json", `{"projects":"bad"}`, NewClaudeCodeDriver().PrepareWorkspace},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			path := filepath.Join(os.Getenv("HOME"), tc.path)
			if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(tc.content), 0600); err != nil {
				t.Fatal(err)
			}
			if err := tc.prepare(context.Background(), t.TempDir()); err == nil {
				t.Fatal("expected error")
			}
			got, err := os.ReadFile(path)
			if err != nil || string(got) != tc.content {
				t.Fatalf("configuration changed: %v", err)
			}
		})
	}
}

//nolint:govet // Test setup scopes errors at the operation under test.
func TestConcurrentTrustUpdates(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	d := NewCodexCLIDriver()
	paths := []string{t.TempDir(), t.TempDir()}
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for _, path := range paths {
		wg.Go(func() { errs <- d.PrepareWorkspace(context.Background(), path) })
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	data, err := os.ReadFile(filepath.Join(os.Getenv("HOME"), ".codex/config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	var cfg map[string]any
	if err := toml.Unmarshal(data, &cfg); err != nil {
		t.Fatal(err)
	}
	projects := cfg["projects"].(map[string]any)
	for _, path := range paths {
		if projects[path] == nil {
			t.Fatal("lost concurrent update")
		}
	}
}

//nolint:govet // Test setup scopes errors at the operation under test.
func TestGenericTrustPreparationAndCancellation(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := NewGenericCLIDriver().PrepareWorkspace(context.Background(), t.TempDir()); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(home)
	if err != nil || len(entries) != 0 {
		t.Fatal("generic driver modified configuration")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := NewCodexCLIDriver().PrepareWorkspace(ctx, t.TempDir()); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
	entries, err = os.ReadDir(home)
	if err != nil || len(entries) != 0 {
		t.Fatal("cancelled preparation modified configuration")
	}
}

//nolint:govet // Error variables are scoped to individual test assertions.
func TestRemoveWorkspaceTrust(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	workspace := t.TempDir()
	if err := RemoveWorkspaceTrust(context.Background(), workspace); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(home)
	if err != nil || len(entries) != 0 {
		t.Fatal("cleanup created configuration")
	}
	codexPath := filepath.Join(home, ".codex/config.toml")
	claudePath := filepath.Join(home, ".claude.json")
	if err := os.MkdirAll(filepath.Dir(codexPath), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(claudePath), 0700); err != nil {
		t.Fatal(err)
	}
	codex := "model = 'example'\n[projects." + strconv.Quote(
		workspace,
	) + "]\ntrust_level = 'trusted'\nother = 'keep'\n[projects.'/other']\ntrust_level = 'trusted'\n"
	claude := `{"other":true,"projects":{"/other":{"hasTrustDialogAccepted":true},` +
		strconv.Quote(workspace) + `:{"hasTrustDialogAccepted":true,"keep":"metadata"}}}`
	if err := os.WriteFile(codexPath, []byte(codex), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(claudePath, []byte(claude), 0600); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err := RemoveWorkspaceTrust(context.Background(), workspace); err != nil {
			t.Fatal(err)
		}
	}
	data, err := os.ReadFile(codexPath)
	if err != nil {
		t.Fatal(err)
	}
	var cfg map[string]any
	if err := toml.Unmarshal(data, &cfg); err != nil {
		t.Fatal(err)
	}
	projects := cfg["projects"].(map[string]any)
	project := projects[workspace].(map[string]any)
	if cfg["model"] != "example" || project["other"] != "keep" || project["trust_level"] != nil || projects["/other"] == nil {
		t.Fatal("Codex cleanup did not preserve other settings")
	}
	data, err = os.ReadFile(claudePath)
	if err != nil {
		t.Fatal(err)
	}
	var c map[string]json.RawMessage
	if err := json.Unmarshal(data, &c); err != nil {
		t.Fatal(err)
	}
	var projectsClaude map[string]map[string]json.RawMessage
	if err := json.Unmarshal(c["projects"], &projectsClaude); err != nil || projectsClaude["/other"] == nil ||
		projectsClaude[workspace]["hasTrustDialogAccepted"] != nil || string(projectsClaude[workspace]["keep"]) != `"metadata"` || c["other"] == nil {
		t.Fatal("Claude cleanup did not preserve other settings")
	}
}

//nolint:govet // Error variables are scoped to individual test assertions.
func TestRemoveTrustWriteFailurePreservesConfiguration(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root bypasses directory write permissions")
	}
	home := t.TempDir()
	t.Setenv("HOME", home)
	path := filepath.Join(home, ".codex/config.toml")
	workspace := t.TempDir()
	if err := NewCodexCLIDriver().PrepareWorkspace(context.Background(), workspace); err != nil {
		t.Fatal(err)
	}
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Dir(path)
	if err := os.Chmod(dir, 0500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0700) })
	if err := RemoveWorkspaceTrust(context.Background(), workspace); err == nil {
		t.Fatal("expected write failure")
	}
	after, err := os.ReadFile(path)
	if err != nil || string(after) != string(original) {
		t.Fatalf("configuration changed on failure: %v", err)
	}
}

func TestClaudeTrustCreationAndLegacyFileIsolation(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	workspace := filepath.Join(home, "actual", "project")
	legacy := filepath.Join(home, ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(legacy), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(legacy, []byte("{invalid legacy"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := NewClaudeCodeDriver().PrepareWorkspace(context.Background(), "~/actual/project"); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(home, ".claude.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var cfg struct {
		Projects map[string]map[string]json.RawMessage `json:"projects"`
	}
	if err = json.Unmarshal(raw, &cfg); err != nil {
		t.Fatal(err)
	}
	if string(cfg.Projects[workspace]["hasTrustDialogAccepted"]) != "true" {
		t.Fatalf("actual path grant missing: %s", raw)
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatalf("new config permissions: %v %v", info, err)
	}
	if err = RemoveWorkspaceTrust(context.Background(), workspace); err != nil {
		t.Fatal(err)
	}
	legacyData, err := os.ReadFile(legacy)
	if err != nil || string(legacyData) != "{invalid legacy" {
		t.Fatalf("legacy changed: %q %v", legacyData, err)
	}
}

//nolint:gocognit // This scenario checks preservation and idempotence across grant and cleanup.
func TestClaudeTrustPreservesValuesAndAvoidsUnnecessaryWrites(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	workspace := t.TempDir()
	path := filepath.Join(home, ".claude.json")
	initial := `{"big":900719925474099312345,"projects":{"/other":{"hasTrustDialogAccepted":true},` + strconv.Quote(
		workspace,
	) + `:{"nested":{"integer":900719925474099312345},"hasTrustDialogAccepted":false}}}`
	if err := os.WriteFile(path, []byte(initial), 0640); err != nil {
		t.Fatal(err)
	}
	if err := NewClaudeCodeDriver().PrepareWorkspace(context.Background(), workspace); err != nil {
		t.Fatal(err)
	}
	first, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(first), "900719925474099312345") {
		t.Fatalf("precision lost: %s", first)
	}
	var cfg map[string]json.RawMessage
	if err = json.Unmarshal(first, &cfg); err != nil {
		t.Fatal(err)
	}
	var projects map[string]map[string]json.RawMessage
	if err = json.Unmarshal(cfg["projects"], &projects); err != nil {
		t.Fatal(err)
	}
	if string(projects[workspace]["hasTrustDialogAccepted"]) != "true" || projects[workspace]["nested"] == nil ||
		projects["/other"] == nil {
		t.Fatalf("project data lost: %s", first)
	}
	if err = NewClaudeCodeDriver().PrepareWorkspace(context.Background(), workspace); err != nil {
		t.Fatal(err)
	}
	second, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(first, second) {
		t.Fatalf("idempotent grant changed bytes: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0640 {
		t.Fatalf("permissions: %v %v", info, err)
	}
	if err = RemoveWorkspaceTrust(context.Background(), workspace); err != nil {
		t.Fatal(err)
	}
	removed, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = RemoveWorkspaceTrust(context.Background(), workspace); err != nil {
		t.Fatal(err)
	}
	repeated, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(removed, repeated) {
		t.Fatalf("repeated cleanup changed bytes: %v", err)
	}
	if err = json.Unmarshal(removed, &cfg); err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(cfg["projects"], &projects); err != nil {
		t.Fatal(err)
	}
	if projects[workspace]["hasTrustDialogAccepted"] != nil || projects[workspace]["nested"] == nil || projects["/other"] == nil {
		t.Fatalf("cleanup lost metadata: %s", removed)
	}
	if !strings.Contains(string(removed), "900719925474099312345") {
		t.Fatalf("cleanup lost precision: %s", removed)
	}
}

func TestClaudeTrustRejectsInvalidObjectsWithoutRewriting(t *testing.T) {
	for _, tc := range []struct{ name, content string }{
		{"invalid JSON", "{"}, {"null root", "null"}, {"array root", "[]"},
		{"null projects", `{"projects":null}`}, {"invalid projects", `{"projects":[]}`},
		{"null project", `{"projects":{"/target":null}}`}, {"invalid project", `{"projects":{"/target":42}}`},
		{"null flag", `{"projects":{"/target":{"hasTrustDialogAccepted":null}}}`},
		{"invalid flag", `{"projects":{"/target":{"hasTrustDialogAccepted":"true"}}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			path := filepath.Join(home, ".claude.json")
			if err := os.WriteFile(path, []byte(tc.content), 0600); err != nil {
				t.Fatal(err)
			}
			for _, remove := range []bool{false, true} {
				if err := updateTrust(context.Background(), "/target", ".claude.json", false, remove); err == nil {
					t.Fatalf("remove=%v accepted invalid configuration", remove)
				}
				got, err := os.ReadFile(path)
				if err != nil || string(got) != tc.content {
					t.Fatalf("configuration changed: %q %v", got, err)
				}
			}
		})
	}
}

func TestClaudeTrustAbsentCleanupAndCancellation(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	path := filepath.Join(home, ".claude.json")
	workspace := t.TempDir()
	if err := RemoveWorkspaceTrust(context.Background(), workspace); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("cleanup created config: %v", err)
	}
	for _, content := range []string{`{"other":true}`, `{"projects":{"/other":{}}}`, `{"projects":{` + strconv.Quote(workspace) + `:{"keep":true}}}`} {
		if err := os.WriteFile(path, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
		if err := RemoveWorkspaceTrust(context.Background(), workspace); err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(path)
		if err != nil || string(got) != content {
			t.Fatalf("no-op cleanup rewrote config: %q %v", got, err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = NewClaudeCodeDriver().PrepareWorkspace(ctx, workspace); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled prepare: %v", err)
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("cancelled preparation changed config: %v", err)
	}
}

func TestConcurrentClaudeTrustUpdates(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	paths := []string{t.TempDir(), t.TempDir()}
	var wg sync.WaitGroup
	errs := make(chan error, len(paths))
	for _, path := range paths {
		wg.Go(func() { errs <- NewClaudeCodeDriver().PrepareWorkspace(context.Background(), path) })
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	raw, err := os.ReadFile(filepath.Join(home, ".claude.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cfg struct {
		Projects map[string]map[string]json.RawMessage `json:"projects"`
	}
	if err = json.Unmarshal(raw, &cfg); err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		if string(cfg.Projects[path]["hasTrustDialogAccepted"]) != "true" {
			t.Fatalf("lost grant for %s: %s", path, raw)
		}
	}
}

func TestClaudeTrustWriteFailurePreservesOriginal(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root bypasses directory write permissions")
	}
	home := t.TempDir()
	t.Setenv("HOME", home)
	path := filepath.Join(home, ".claude.json")
	initial := `{"projects":{"/other":{"hasTrustDialogAccepted":true}},"keep":123}`
	if err := os.WriteFile(path, []byte(initial), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(home, 0500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chmod(home, 0700); err != nil {
			t.Errorf("restore home permissions: %v", err)
		}
	})
	if err := NewClaudeCodeDriver().PrepareWorkspace(context.Background(), t.TempDir()); err == nil {
		t.Fatal("expected write failure")
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != initial {
		t.Fatalf("configuration changed: %q %v", got, err)
	}
}

func TestClaudeTrustCleanupWriteFailurePreservesOriginal(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root bypasses directory write permissions")
	}
	home := t.TempDir()
	t.Setenv("HOME", home)
	workspace := t.TempDir()
	path := filepath.Join(home, ".claude.json")
	if err := NewClaudeCodeDriver().PrepareWorkspace(context.Background(), workspace); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Chmod(home, 0500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if restoreErr := os.Chmod(home, 0700); restoreErr != nil {
			t.Errorf("restore home permissions: %v", restoreErr)
		}
	})
	if err = RemoveWorkspaceTrust(context.Background(), workspace); err == nil {
		t.Fatal("expected cleanup write failure")
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("failed cleanup changed config: %v", err)
	}
}
