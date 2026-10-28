package driver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/pelletier/go-toml/v2"
)

// trustConfigMu serializes updates within this process. Independent Kankuro
// processes and other CLI config writers are not covered by this lock.
//
//nolint:gochecknoglobals // One lock coordinates trust writes across all driver instances.
var trustConfigMu sync.Mutex

func (d *GenericCLIDriver) PrepareWorkspace(ctx context.Context, _ string) error {
	return ctx.Err()
}

func (d *CodexCLIDriver) PrepareWorkspace(ctx context.Context, workingDir string) error {
	return updateTrust(ctx, workingDir, ".codex/config.toml", true, false)
}
func (d *ClaudeCodeDriver) PrepareWorkspace(ctx context.Context, workingDir string) error {
	return updateTrust(ctx, workingDir, ".claude/settings.json", false, false)
}

// RemoveWorkspaceTrust removes the exact workspace grant from both supported
// CLI configurations, including grants from earlier card switches.
func RemoveWorkspaceTrust(ctx context.Context, workingDir string) error {
	if workingDir == "" {
		return errors.New("workspace path is empty")
	}
	return errors.Join(
		updateTrust(ctx, workingDir, ".codex/config.toml", true, true),
		updateTrust(ctx, workingDir, ".claude/settings.json", false, true),
	)
}

func updateTrust(ctx context.Context, workingDir, configPath string, codex, remove bool) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolve home directory: %w", err)
	}
	if workingDir == "~" {
		workingDir = home
	} else if strings.HasPrefix(workingDir, "~/") {
		workingDir = filepath.Join(home, workingDir[2:])
	}
	workingDir, err = filepath.Abs(workingDir)
	if err != nil {
		return fmt.Errorf("resolve workspace path: %w", err)
	}
	path := filepath.Join(home, configPath)
	trustConfigMu.Lock()
	defer trustConfigMu.Unlock()
	if ctxErr := ctx.Err(); ctxErr != nil {
		return ctxErr
	}
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("read trust configuration %s: %w", path, err)
	}
	if remove && errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if !codex && err == nil && len(data) == 0 {
		return fmt.Errorf("update trust configuration %s: settings must be a valid JSON object", path)
	}
	var result []byte
	if codex {
		result, err = updateCodexTrust(data, workingDir, remove)
	} else {
		result, err = updateClaudeTrust(data, workingDir, remove)
	}
	if err != nil {
		return fmt.Errorf("update trust configuration %s: %w", path, err)
	}
	if result == nil {
		return nil
	}
	if ctxErr := ctx.Err(); ctxErr != nil {
		return ctxErr
	}
	if writeErr := atomicTrustWrite(path, result); writeErr != nil {
		return fmt.Errorf("write trust configuration %s: %w", path, writeErr)
	}
	return nil
}

//nolint:gocognit // Validate and preserve the complete configuration while updating one entry.
func updateCodexTrust(data []byte, workspace string, remove bool) ([]byte, error) {
	var cfg map[string]any
	if len(data) > 0 {
		if err := toml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("invalid TOML: %w", err)
		}
	}
	if cfg == nil {
		cfg = make(map[string]any)
	}
	projects, ok := cfg["projects"].(map[string]any)
	if !ok && cfg["projects"] != nil {
		return nil, errors.New("projects must be a table")
	}
	if projects == nil && remove {
		return nil, nil
	}
	if projects == nil {
		projects = make(map[string]any)
		cfg["projects"] = projects
	}
	project, ok := projects[workspace].(map[string]any)
	if !ok && projects[workspace] != nil {
		return nil, errors.New("workspace project must be a table")
	}
	if project == nil && remove {
		return nil, nil
	}
	if project == nil {
		project = make(map[string]any)
		projects[workspace] = project
	}
	if remove {
		if level, exists := project["trust_level"]; !exists {
			return nil, nil
		} else if _, isString := level.(string); !isString {
			return nil, errors.New("trust_level must be a string")
		}
		delete(project, "trust_level")
		if len(project) == 0 {
			delete(projects, workspace)
		}
		return toml.Marshal(cfg)
	}
	if level, exists := project["trust_level"]; exists {
		if _, isString := level.(string); !isString {
			return nil, errors.New("trust_level must be a string")
		}
		if level == "trusted" {
			return nil, nil
		}
	}
	project["trust_level"] = "trusted"
	return toml.Marshal(cfg)
}

//nolint:gocognit // Validate and preserve the complete configuration while updating one entry.
func updateClaudeTrust(data []byte, workspace string, remove bool) ([]byte, error) {
	cfg := make(map[string]json.RawMessage)
	if len(data) > 0 {
		if err := json.Unmarshal(data, &cfg); err != nil || cfg == nil {
			return nil, errors.New("settings must be a valid JSON object")
		}
	}
	var dirs []string
	if raw, exists := cfg["trustedDirectories"]; exists {
		var values []json.RawMessage
		if string(raw) == "null" || json.Unmarshal(raw, &values) != nil {
			return nil, errors.New("trustedDirectories must be an array of strings")
		}
		for _, value := range values {
			var dir string
			if string(value) == "null" || json.Unmarshal(value, &dir) != nil {
				return nil, errors.New("trustedDirectories must be an array of strings")
			}
			dirs = append(dirs, dir)
		}
	}
	if remove {
		if !slices.Contains(dirs, workspace) {
			return nil, nil
		}
		dirs = slices.DeleteFunc(dirs, func(dir string) bool { return dir == workspace })
		if dirs == nil {
			dirs = []string{}
		}
	} else {
		if slices.Contains(dirs, workspace) {
			return nil, nil
		}
		dirs = append(dirs, workspace)
	}
	cfg["trustedDirectories"], _ = json.Marshal(dirs)
	return json.MarshalIndent(cfg, "", "  ")
}

func atomicTrustWrite(path string, data []byte) (err error) {
	if err = os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	mode := os.FileMode(0o600) //nolint:mnd // Private configuration for a newly created file.
	if info, statErr := os.Stat(path); statErr == nil {
		mode = info.Mode().Perm()
	} else if !errors.Is(statErr, os.ErrNotExist) {
		return statErr
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".kankuro-trust-*")
	if err != nil {
		return err
	}
	tempName := f.Name()
	defer os.Remove(tempName)
	defer func() {
		if f != nil {
			err = errors.Join(err, f.Close())
		}
	}()
	if err = f.Chmod(mode); err != nil {
		return err
	}
	if _, err = f.Write(data); err != nil {
		return err
	}
	if err = f.Sync(); err != nil {
		return err
	}
	if err = f.Close(); err != nil {
		f = nil
		return err
	}
	f = nil
	return os.Rename(tempName, path)
}
