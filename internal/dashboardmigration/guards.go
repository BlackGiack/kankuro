package dashboardmigration

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
)

func historyGuards(dir string) (map[string]string, error) {
	guards := make(map[string]string)
	entries, err := os.ReadDir(filepath.Join(dir, "operations"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != jsonExtension {
			continue
		}
		if !entry.Type().IsRegular() {
			return nil, fmt.Errorf("operation %s is not a regular file", entry.Name())
		}
		path := filepath.Join("operations", entry.Name())
		raw, readErr := os.ReadFile(filepath.Join(dir, path))
		if readErr != nil {
			return nil, readErr
		}
		guards[path] = digest(raw)
	}
	path := filepath.Join("messaging", "history.json")
	info, err := os.Lstat(filepath.Join(dir, path))
	if errors.Is(err, os.ErrNotExist) {
		return guards, nil
	}
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("message history is not a regular file")
	}
	raw, err := os.ReadFile(filepath.Join(dir, path))
	if err != nil {
		return nil, err
	}
	guards[path] = digest(raw)
	return guards, nil
}

func verifyHistoryGuards(dir string, expected map[string]string) error {
	if expected == nil {
		return errors.New("migration manifest is missing history guards")
	}
	current, err := historyGuards(dir)
	if err != nil {
		return err
	}
	if !maps.Equal(current, expected) {
		return errors.New("operation or message history changed since staging")
	}
	return nil
}
