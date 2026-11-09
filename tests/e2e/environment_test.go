//go:build e2e

package e2e_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

// isolatedAgentEnvironment copies only configuration, credentials and model
// preferences, never historical conversations or a running server's state.
func isolatedAgentEnvironment(t *testing.T, sandbox, agent string) map[string]string {
	t.Helper()
	originalHome, err := os.UserHomeDir()
	must(t, err)
	home := filepath.Join(sandbox, "home")
	configHome := filepath.Join(home, ".config")
	dataHome := filepath.Join(home, ".local/share")
	stateHome := filepath.Join(home, ".local/state")
	for _, dir := range []string{home, configHome, dataHome, stateHome} {
		must(t, os.MkdirAll(dir, 0o700))
	}
	originalConfig := os.Getenv("XDG_CONFIG_HOME")
	if originalConfig == "" {
		originalConfig = filepath.Join(originalHome, ".config")
	}
	originalData := os.Getenv("XDG_DATA_HOME")
	if originalData == "" {
		originalData = filepath.Join(originalHome, ".local/share")
	}
	originalState := os.Getenv("XDG_STATE_HOME")
	if originalState == "" {
		originalState = filepath.Join(originalHome, ".local/state")
	}
	environment := map[string]string{
		"HOME": home, "XDG_CONFIG_HOME": configHome, "XDG_DATA_HOME": dataHome,
		"XDG_STATE_HOME": stateHome, "XDG_CACHE_HOME": filepath.Join(home, ".cache"),
		"TERM": "xterm-256color", "COLORTERM": "truecolor",
	}
	switch agent {
	case "opencode":
		copyOpenCodeConfiguration(t, filepath.Join(originalConfig, "opencode"), filepath.Join(configHome, "opencode"))
		copyConfiguration(t, filepath.Join(originalData, "opencode/auth.json"), filepath.Join(dataHome, "opencode/auth.json"))
		copyConfiguration(t, filepath.Join(originalState, "opencode/model.json"), filepath.Join(stateHome, "opencode/model.json"))
		copyConfiguration(t, filepath.Join(originalState, "opencode/kv.json"), filepath.Join(stateHome, "opencode/kv.json"))
		if _, statErr := os.Stat(filepath.Join(originalData, "opencode/opencode.db")); statErr == nil {
			copyOpenCodeCredentials(t, environment)
		} else if !os.IsNotExist(statErr) {
			t.Fatal(statErr)
		}
	case "codex":
		source := os.Getenv("CODEX_HOME")
		if source == "" {
			source = filepath.Join(originalHome, ".codex")
		}
		destination := filepath.Join(home, ".codex")
		for _, name := range []string{"config.toml", "auth.json", "instructions.md", "rules"} {
			copyConfiguration(t, filepath.Join(source, name), filepath.Join(destination, name))
		}
		environment["CODEX_HOME"] = destination
	case "claude":
		copyConfiguration(t, filepath.Join(originalHome, ".claude.json"), filepath.Join(home, ".claude.json"))
		for _, name := range []string{"settings.json", "settings.local.json", ".credentials.json"} {
			copyConfiguration(t, filepath.Join(originalHome, ".claude", name), filepath.Join(home, ".claude", name))
		}
		environment["CLAUDE_CONFIG_DIR"] = filepath.Join(home, ".claude")
	}
	return environment
}

func copyConfiguration(t *testing.T, source, destination string) {
	t.Helper()
	info, err := os.Stat(source)
	if os.IsNotExist(err) {
		return
	}
	must(t, err)
	if !info.IsDir() {
		copyConfigurationFile(t, source, destination)
		return
	}
	must(t, filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, relativeErr := filepath.Rel(source, path)
		if relativeErr != nil {
			return relativeErr
		}
		target := filepath.Join(destination, relative)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o700)
		}
		if entry.Type().IsRegular() {
			copyConfigurationFile(t, path, target)
		}
		return nil
	}))
}

func copyConfigurationFile(t *testing.T, source, destination string) {
	t.Helper()
	bytes, err := os.ReadFile(source)
	must(t, err)
	must(t, os.MkdirAll(filepath.Dir(destination), 0o700))
	must(t, os.WriteFile(destination, bytes, 0o600))
}
