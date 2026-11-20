//nolint:testpackage // Tests exercise internal storage and inspection boundaries.
package api

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
)

func workspaceGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(
		os.Environ(),
		"GIT_AUTHOR_NAME=test",
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=test",
		"GIT_COMMITTER_EMAIL=test@example.com",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
}
func TestCenterWorkspaceGitAndDirectory(t *testing.T) {
	dir := t.TempDir()
	plain, err := inspectWorkspace(context.Background(), dir)
	if err != nil || plain.Kind != contracts.WorkspaceDirectory || len(plain.Warnings) != 0 || len(plain.Entries) != 2 {
		t.Fatalf("non-git = %#v, %v", plain, err)
	}
	workspaceGit(t, dir, "init", "-q")
	if err = os.WriteFile(filepath.Join(dir, "old name"), []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}
	workspaceGit(t, dir, "add", ".")
	workspaceGit(t, dir, "commit", "-qm", "initial")
	clean, err := inspectWorkspace(context.Background(), dir)
	if err != nil || clean.Kind != contracts.WorkspaceGit || len(clean.GitEntries) != 0 {
		t.Fatalf("clean = %#v, %v", clean, err)
	}
	workspaceGit(t, dir, "mv", "old name", "new name")
	if err = os.WriteFile(filepath.Join(dir, "new name"), []byte("changed"), 0600); err != nil {
		t.Fatal(err)
	}
	if err = os.Mkdir(filepath.Join(dir, "sub"), 0700); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dir, "untracked\nfile"), []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	got, err := inspectWorkspace(context.Background(), filepath.Join(dir, "sub"))
	if err != nil || got.Kind != contracts.WorkspaceGit || got.RepositoryRoot != dir || len(got.GitEntries) != 2 {
		t.Fatalf("git = %#v, %v", got, err)
	}
	if got.GitEntries[0].Path != "new name" || got.GitEntries[0].OriginalPath != "old name" || got.GitEntries[0].IndexStatus != 'R' ||
		got.GitEntries[0].WorktreeStatus != 'M' ||
		got.GitEntries[1].Path != "untracked\nfile" {
		t.Fatalf("paths = %#v", got.GitEntries)
	}
}
func TestCenterWorkspaceBoundedListingAndErrors(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".hidden"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(".hidden", filepath.Join(dir, "link")); err != nil {
		t.Fatal(err)
	}
	for i := range 1000 {
		if err := os.WriteFile(filepath.Join(dir, strings.Repeat("a", 4)+string(rune(0x100+i))), nil, 0600); err != nil {
			t.Fatal(err)
		}
	}
	entries, truncated, _, err := listWorkspace(context.Background(), dir)
	if err != nil || !truncated || len(entries) != workspaceMaxEntries {
		t.Fatalf("listing count %d truncated %v err %v", len(entries), truncated, err)
	}
	if _, err = inspectWorkspace(context.Background(), filepath.Join(dir, "missing")); err == nil {
		t.Fatal("missing workspace accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = inspectWorkspace(ctx, dir); err == nil {
		t.Fatal("cancellation ignored")
	}
	for _, bad := range [][]byte{[]byte("1 bad\x00"), []byte("? incomplete"), []byte("2 R. N... 100644 100644 100644 a b R100 name\x00")} {
		if _, _, err = parseGitStatus(bad); err == nil {
			t.Fatalf("accepted malformed %q", bad)
		}
	}
}

func TestCenterWorkspaceSymlinkAndMissingGit(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".hidden"), []byte("data"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(".hidden", filepath.Join(dir, "link")); err != nil {
		t.Fatal(err)
	}
	entries, truncated, warnings, err := listWorkspace(context.Background(), dir)
	if err != nil || truncated || len(warnings) > 0 || len(entries) != 4 {
		t.Fatalf("listing=%+v %v %v %v", entries, truncated, warnings, err)
	}
	var found bool
	for _, e := range entries {
		if e.Name == "link" {
			found = e.SymlinkTarget != nil && *e.SymlinkTarget == ".hidden"
		}
	}
	if !found {
		t.Fatal("symlink target missing")
	}
	t.Setenv("PATH", t.TempDir())
	got, err := inspectWorkspace(context.Background(), dir)
	if err != nil || got.Kind != contracts.WorkspaceDirectory || len(got.Warnings) == 0 {
		t.Fatalf("missing Git must be a warned fallback: %+v, %v", got, err)
	}
}

func TestCenterWorkspaceLinkedWorktree(t *testing.T) {
	repo := t.TempDir()
	workspaceGit(t, repo, "init", "-q")
	if err := os.WriteFile(filepath.Join(repo, "tracked"), []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	workspaceGit(t, repo, "add", ".")
	workspaceGit(t, repo, "commit", "-qm", "initial")
	linked := filepath.Join(t.TempDir(), "worktree")
	workspaceGit(t, repo, "worktree", "add", "-qb", "test-branch", linked)
	if err := os.WriteFile(filepath.Join(linked, "tracked"), []byte("modified"), 0600); err != nil {
		t.Fatal(err)
	}
	got, err := inspectWorkspace(context.Background(), linked)
	if err != nil || got.Kind != contracts.WorkspaceGit || got.RepositoryRoot != linked || len(got.GitEntries) != 1 ||
		got.GitEntries[0].Path != "tracked" {
		t.Fatalf("linked worktree=%+v, %v", got, err)
	}
}

func TestCenterWorkspaceGitTimeoutAndTruncation(t *testing.T) {
	data := []byte(strings.Repeat("? file\x00", workspaceMaxEntries+1))
	entries, truncated, err := parseGitStatus(data)
	if err != nil || !truncated || len(entries) != workspaceMaxEntries {
		t.Fatalf("bounded git status: %d %v %v", len(entries), truncated, err)
	}
	bin := t.TempDir()
	fake := filepath.Join(bin, "git")
	if err = os.WriteFile(fake, []byte("#!/bin/sh\nwhile :; do :; done\n"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	started := time.Now()
	_, err = runGit(ctx, t.TempDir(), "status")
	if !errors.Is(err, context.DeadlineExceeded) || time.Since(started) > 2*time.Second {
		t.Fatalf("deadline: %v after %s", err, time.Since(started))
	}
}

func TestCenterWorkspaceKillsGitOnOutputLimit(t *testing.T) {
	bin := t.TempDir()
	script := "#!/bin/sh\nwhile :; do printf '" + strings.Repeat("0123456789", 800) + "'; done\n"
	if err := os.WriteFile(filepath.Join(bin, "git"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := runGit(ctx, t.TempDir(), "status")
	if err == nil || !strings.Contains(err.Error(), "output exceeded") || ctx.Err() != nil {
		t.Fatalf("output limit was not enforced promptly: %v", err)
	}
}

func TestCenterWorkspaceBoundsInheritedPipes(t *testing.T) {
	if _, err := os.Stat("/bin/sleep"); err != nil {
		t.Skip("test requires /bin/sleep")
	}
	for _, tc := range []struct {
		name, body string
		timeout    time.Duration
		want       string
	}{
		{name: "cancellation", body: "wait", timeout: 100 * time.Millisecond, want: "deadline"},
		{name: "output limit", body: "while :; do printf '" + strings.Repeat("0123456789", 800) + "'; done", timeout: 5 * time.Second, want: "output exceeded"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bin := t.TempDir()
			script := "#!/bin/sh\n/bin/sleep 3 &\n" + tc.body + "\n"
			if err := os.WriteFile(filepath.Join(bin, "git"), []byte(script), 0700); err != nil {
				t.Fatal(err)
			}
			t.Setenv("PATH", bin)
			ctx, cancel := context.WithTimeout(t.Context(), tc.timeout)
			defer cancel()
			started := time.Now()
			_, err := runGit(ctx, t.TempDir(), "status")
			elapsed := time.Since(started)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected %s, got %v", tc.want, err)
			}
			if elapsed >= 1500*time.Millisecond {
				t.Fatalf("inherited pipes delayed completion: %v", elapsed)
			}
		})
	}
}

func TestCenterWorkspaceIgnoresConfiguredGitHelpers(t *testing.T) {
	dir := t.TempDir()
	workspaceGit(t, dir, "init", "-q")
	if err := os.WriteFile(filepath.Join(dir, "tracked.txt"), []byte("initial\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".gitattributes"), []byte("tracked.txt filter=sentinel\n"), 0600); err != nil {
		t.Fatal(err)
	}
	workspaceGit(t, dir, "add", ".")
	workspaceGit(t, dir, "commit", "-qm", "initial")
	workspaceGit(t, dir, "update-index", "--split-index")
	sentinel := filepath.Join(dir, "helper-ran")
	helper := filepath.Join(t.TempDir(), "sentinel-helper")
	if err := os.WriteFile(helper, []byte("#!/bin/sh\nprintf called > '"+sentinel+"'\n"), 0700); err != nil {
		t.Fatal(err)
	}
	workspaceGit(t, dir, "config", "core.fsmonitor", helper)
	if got, err := inspectWorkspace(t.Context(), dir); err != nil || got.Kind != contracts.WorkspaceGit || len(got.GitEntries) != 0 {
		t.Fatalf("expected clean status with fsmonitor disabled: %+v, %v", got, err)
	}
	workspaceGit(t, dir, "-c", "core.fsmonitor=false", "config", "filter.sentinel.clean", helper)
	if err := os.WriteFile(filepath.Join(dir, "tracked.txt"), []byte("changed"), 0600); err != nil {
		t.Fatal(err)
	}
	got, err := inspectWorkspace(t.Context(), dir)
	if err != nil || got.Kind != contracts.WorkspaceDirectory || !strings.Contains(strings.Join(got.Warnings, " "), "clean filters") {
		t.Fatalf("expected labelled fallback for configured filters: %+v, %v", got, err)
	}
	if _, statErr := os.Stat(sentinel); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("helper ran or sentinel could not be checked: %v", statErr)
	}
}

func TestCenterWorkspaceFallsBackAfterMetadataProbeFailure(t *testing.T) {
	dir := t.TempDir()
	if err := os.Symlink("HEAD", filepath.Join(dir, "HEAD")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "visible.txt"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	got, err := inspectWorkspace(t.Context(), dir)
	if err != nil || got.Kind != contracts.WorkspaceDirectory ||
		!strings.Contains(strings.Join(got.Warnings, " "), "inspect repository metadata") {
		t.Fatalf("expected a labelled fallback: %+v, %v", got, err)
	}
	for _, item := range got.Entries {
		if item.Name == "visible.txt" {
			return
		}
	}
	t.Fatal("directory fallback lost the readable file")
}

func TestCenterWorkspaceLinkedWorktreeStatus(t *testing.T) {
	root := t.TempDir()
	workspaceGit(t, root, "init", "-q")
	workspaceGit(t, root, "commit", "--allow-empty", "-qm", "initial")
	worktree := filepath.Join(t.TempDir(), "linked")
	workspaceGit(t, root, "worktree", "add", "--detach", worktree)
	if err := os.WriteFile(filepath.Join(worktree, "untracked"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	got, err := inspectWorkspace(t.Context(), worktree)
	if err != nil || got.Kind != contracts.WorkspaceGit || got.RepositoryRoot != worktree ||
		len(got.GitEntries) != 1 || got.GitEntries[0].Path != "untracked" {
		t.Fatalf("linked worktree status: %+v, %v", got, err)
	}
}

func TestCenterWorkspaceSplitIndexStatus(t *testing.T) {
	root := t.TempDir()
	workspaceGit(t, root, "init", "-q")
	if err := os.WriteFile(filepath.Join(root, "tracked.txt"), []byte("initial\n"), 0600); err != nil {
		t.Fatal(err)
	}
	workspaceGit(t, root, "add", "tracked.txt")
	workspaceGit(t, root, "commit", "-qm", "initial")
	linked := filepath.Join(t.TempDir(), "linked")
	workspaceGit(t, root, "worktree", "add", "--detach", linked)
	for _, tc := range []struct{ name, dir string }{{"main", root}, {"linked", linked}} {
		t.Run(tc.name, func(t *testing.T) {
			workspaceGit(t, tc.dir, "update-index", "--split-index")
			if err := os.WriteFile(filepath.Join(tc.dir, "tracked.txt"), []byte("changed\n"), 0600); err != nil {
				t.Fatal(err)
			}
			got, err := inspectWorkspace(t.Context(), tc.dir)
			if err != nil || got.Kind != contracts.WorkspaceGit || len(got.Warnings) != 0 || len(got.GitEntries) != 1 {
				t.Fatalf("expected one Git change with a split index: %+v, %v", got, err)
			}
			if entry := got.GitEntries[0]; entry.Path != "tracked.txt" || entry.WorktreeStatus != 'M' {
				t.Fatalf("unexpected Git change: %+v", entry)
			}
		})
	}
}

func TestCenterWorkspaceReportsSubmoduleFallback(t *testing.T) {
	dir := t.TempDir()
	workspaceGit(t, dir, "init", "-q")
	workspaceGit(t, dir, "commit", "--allow-empty", "-qm", "initial")
	head, err := runGit(t.Context(), dir, "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	workspaceGit(t, dir, "update-index", "--add", "--cacheinfo", "160000,"+strings.TrimSpace(string(head))+",module")
	got, err := inspectWorkspace(t.Context(), dir)
	if err != nil || got.Kind != contracts.WorkspaceDirectory || !strings.Contains(strings.Join(got.Warnings, " "), "submodules") {
		t.Fatalf("expected labelled fallback for a gitlink: %+v, %v", got, err)
	}
}
