package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// expandPath handles ~ expansion, env var expansion, and returns an absolute path.
func expandPath(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", errors.New("path is empty")
	}
	value = os.ExpandEnv(value)
	if strings.HasPrefix(value, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		suffix := strings.TrimPrefix(value, "~")
		suffix = strings.TrimPrefix(suffix, string(os.PathSeparator))
		value = filepath.Join(home, suffix)
	}
	return filepath.Abs(value)
}

// ScanEntries iterates over source paths, expanding each, reading directory entries,
// and calling onEntry for each entry that passes shouldInclude.
//
// Expansion and read errors are wrapped with the source path for diagnostics.
// Callback errors are returned unchanged.
//
// For sync flows where invalid sources should be silently skipped, use
// ScanEntriesWithSkip instead.
func ScanEntries(
	sources []string,
	shouldInclude func(entry os.DirEntry) bool,
	onEntry func(src, expanded string, entry os.DirEntry) error,
) error {
	for _, src := range sources {
		expanded, expErr := expandPath(src)
		if expErr != nil {
			return fmt.Errorf("expand source %q: %w", src, expErr)
		}
		entries, readErr := os.ReadDir(expanded)
		if readErr != nil {
			return fmt.Errorf("read source %q: %w", src, readErr)
		}
		for _, entry := range entries {
			if !shouldInclude(entry) {
				continue
			}
			if cbErr := onEntry(src, expanded, entry); cbErr != nil {
				return cbErr
			}
		}
	}
	return nil
}

// ScanEntriesWithSkip is like ScanEntries but skips sources that fail to expand
// or read, returning the first onEntry error instead.
func ScanEntriesWithSkip(
	sources []string,
	shouldInclude func(entry os.DirEntry) bool,
	onEntry func(src, expanded string, entry os.DirEntry) error,
) error {
	for _, src := range sources {
		expanded, expErr := expandPath(src)
		if expErr != nil {
			continue
		}
		entries, readErr := os.ReadDir(expanded)
		if readErr != nil {
			continue
		}
		for _, entry := range entries {
			if !shouldInclude(entry) {
				continue
			}
			if cbErr := onEntry(src, expanded, entry); cbErr != nil {
				return cbErr
			}
		}
	}
	return nil
}
