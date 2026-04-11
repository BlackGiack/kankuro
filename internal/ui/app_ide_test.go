package ui //nolint:testpackage // IDE tests need unexported helpers.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSplitCommand(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"simple", "code /path", []string{"code", "/path"}},
		{"quoted path", `code "/path with spaces"`, []string{"code", "/path with spaces"}},
		{"single quotes", `code '/path with spaces'`, []string{"code", "/path with spaces"}},
		{"extra whitespace", "  code   /path  ", []string{"code", "/path"}},
		{"empty", "", nil},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := splitCommand(tc.input)
			if len(got) != len(tc.want) {
				t.Fatalf("expected %d parts, got %d", len(tc.want), len(got))
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("part %d: expected %q, got %q", i, tc.want[i], got[i])
				}
			}
		})
	}
}

func TestIDEPathReplacement(t *testing.T) {
	tmpDir := t.TempDir()
	cmd := "code $PATH"
	got := strings.ReplaceAll(cmd, idePathPlaceholder, tmpDir)
	want := "code " + tmpDir
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestResolveWorkspacePath(t *testing.T) {
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "workspace")
	if err := os.MkdirAll(testDir, 0o700); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}

	// Absolute path
	got, err := resolveWorkspacePath(testDir)
	if err != nil {
		t.Fatalf("expected no error for absolute path, got %v", err)
	}
	if got != testDir {
		t.Fatalf("expected %q, got %q", testDir, got)
	}

	// Non-existent path
	_, err = resolveWorkspacePath("/nonexistent/path")
	if err == nil {
		t.Fatalf("expected error for non-existent path")
	}

	// Empty path
	_, err = resolveWorkspacePath("")
	if err == nil {
		t.Fatalf("expected error for empty path")
	}
}
