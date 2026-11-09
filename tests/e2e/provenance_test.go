//go:build e2e

package e2e_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// sourceFingerprint uses the Appedia collector's canonical source inventory.
// Images/results are not source inputs; documentation does not affect this hash.
func sourceFingerprint(t *testing.T, root string) string {
	t.Helper()
	exclude := []string{".agents", ".kankuro", "appedia", "kankuro_skills/create-appedia", "kankuro_skills/create-appediff", "tmp"}
	skip := []string{
		".git",
		".hg",
		".svn",
		".maintenance",
		"node_modules",
		"vendor",
		"__pycache__",
		".venv",
		"venv",
		"appedia",
		"appediff",
		".next",
	}
	command := exec.Command("git", "ls-files", "-co", "--exclude-standard", "-z")
	command.Dir = root
	output, err := command.Output()
	must(t, err)
	sources := make(map[string]string)
	for path := range strings.SplitSeq(string(output), "\x00") {
		if !isSource(path) || excludedSource(path, append(skip, exclude...)) {
			continue
		}
		info, statErr := os.Lstat(filepath.Join(root, path))
		if os.IsNotExist(statErr) {
			continue
		}
		must(t, statErr)
		if !info.Mode().IsRegular() {
			continue
		}
		bytes, readErr := os.ReadFile(filepath.Join(root, path))
		must(t, readErr)
		sum := sha256.Sum256(bytes)
		sources[path] = hex.EncodeToString(sum[:])
	}
	slices.Sort(exclude)
	// Map keys are sorted by encoding/json, matching Python's sort_keys=True.
	canonical, err := json.Marshal(map[string]any{"sources": sources, "exclude": exclude})
	must(t, err)
	sum := sha256.Sum256(canonical)
	return hex.EncodeToString(sum[:])
}

func excludedSource(path string, excluded []string) bool {
	directory := filepath.ToSlash(filepath.Dir(path))
	for part := range strings.SplitSeq(directory, "/") {
		if slices.Contains(excluded, part) {
			return true
		}
	}
	return false
}

func isSource(path string) bool {
	extensions := []string{".go", ".py", ".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs", ".mts", ".cts", ".tf"}
	manifests := []string{
		"go.mod",
		"go.sum",
		"package.json",
		"package-lock.json",
		"pnpm-lock.yaml",
		"yarn.lock",
		"pyproject.toml",
		"uv.lock",
		"requirements.txt",
	}
	return slices.Contains(extensions, filepath.Ext(path)) || slices.Contains(manifests, filepath.Base(path))
}
