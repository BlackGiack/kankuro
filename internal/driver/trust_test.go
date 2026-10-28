//nolint:testpackage // Exercises configuration helpers alongside driver behavior.
package driver

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
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
		{"claude", ".claude/settings.json", NewClaudeCodeDriver().PrepareWorkspace, `{"other":{"keep":true},"trustedDirectories":["/other"]}`},
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
				var dirs []string
				if err := json.Unmarshal(
					cfg["trustedDirectories"],
					&dirs,
				); err != nil || len(dirs) != 2 || dirs[1] != workspace ||
					cfg["other"] == nil {
					t.Fatal("settings not preserved")
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
		{"invalid JSON", ".claude/settings.json", "{broken", NewClaudeCodeDriver().PrepareWorkspace},
		{"wrong type", ".claude/settings.json", `{"trustedDirectories":"bad"}`, NewClaudeCodeDriver().PrepareWorkspace},
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
	claudePath := filepath.Join(home, ".claude/settings.json")
	if err := os.MkdirAll(filepath.Dir(codexPath), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(claudePath), 0700); err != nil {
		t.Fatal(err)
	}
	codex := "model = 'example'\n[projects." + strconv.Quote(
		workspace,
	) + "]\ntrust_level = 'trusted'\nother = 'keep'\n[projects.'/other']\ntrust_level = 'trusted'\n"
	claude := `{"other":true,"trustedDirectories":["/other",` + strconv.Quote(workspace) + `,` + strconv.Quote(workspace) + `]}`
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
	var dirs []string
	if err := json.Unmarshal(c["trustedDirectories"], &dirs); err != nil || len(dirs) != 1 || dirs[0] != "/other" || c["other"] == nil {
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
