package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
)

const (
	workspaceMaxEntries   = 1000
	workspaceBatchSize    = 128
	workspaceDotEntries   = 2
	gitMinimumRecordBytes = 4
	gitOutputLimit        = 4 << 20
	gitErrorLimit         = 8 << 10
	gitTimeout            = 10 * time.Second
	gitPipeDrainTimeout   = 250 * time.Millisecond
)

// ReadSessionWorkspace inspects the saved actual workspace, revalidating scope
// after external I/O. Git paths are relative to the repository root.
func (s *Service) ReadSessionWorkspace(ctx context.Context, req contracts.ScopedSessionRequest) (contracts.SessionWorkspaceResult, error) {
	if err := s.CheckSessionInDashboard(req.DashboardID, req.SessionID); err != nil {
		return contracts.SessionWorkspaceResult{}, err
	}
	snap, ok := s.sessions.GetSession(req.SessionID)
	if !ok || snap.DashboardID != req.DashboardID || snap.WorkingDir == "" || !filepath.IsAbs(snap.WorkingDir) {
		return contracts.SessionWorkspaceResult{}, errors.New("saved session workspace unavailable")
	}
	result, inspectErr := inspectWorkspace(ctx, snap.WorkingDir)
	if err := s.CheckSessionInDashboard(req.DashboardID, req.SessionID); err != nil {
		return contracts.SessionWorkspaceResult{}, err
	}
	current, ok := s.sessions.GetSession(req.SessionID)
	if !ok || current.DashboardID != req.DashboardID || current.WorkingDir != snap.WorkingDir {
		return contracts.SessionWorkspaceResult{}, errors.New("session workspace changed during inspection")
	}
	return result, inspectErr
}

//nolint:nestif // Repository probe outcomes determine the explicit fallback classification.
func inspectWorkspace(ctx context.Context, dir string) (contracts.SessionWorkspaceResult, error) {
	result := contracts.SessionWorkspaceResult{WorkingDir: dir}
	info, err := os.Stat(dir)
	if err != nil {
		return result, fmt.Errorf("inspect workspace: %w", err)
	}
	if !info.IsDir() {
		return result, errors.New("saved workspace is not a directory")
	}
	metadata, err := findGitMetadata(dir)
	if err != nil {
		return result, err
	}
	root, probeErr := runGit(ctx, dir, "rev-parse", "--show-toplevel")
	if probeErr == nil {
		repositoryRoot := strings.TrimSuffix(string(root), "\n")
		if repositoryRoot == "" || !filepath.IsAbs(repositoryRoot) {
			probeErr = errors.New("git returned an invalid repository root")
		} else {
			result.RepositoryRoot = repositoryRoot
			data, statusErr := runGit(ctx, repositoryRoot, "status", "--porcelain=v2", "-z", "--untracked-files=all")
			if statusErr == nil {
				entries, truncated, parseErr := parseGitStatus(data)
				if parseErr == nil {
					result.Kind = contracts.WorkspaceGit
					result.GitEntries = entries
					result.Truncated = truncated
					return result, nil
				}
				statusErr = parseErr
			}
			probeErr = fmt.Errorf("git status: %w", statusErr)
		}
	}
	if contextErr := ctx.Err(); contextErr != nil {
		return result, contextErr
	}
	// A normal exit 128 without metadata is Git's controlled non-repository
	// outcome. Other failures, including missing Git, are inspection errors.
	var exitErr *exec.ExitError
	nonGit := !metadata && errors.As(probeErr, &exitErr) && exitErr.ExitCode() == 128 &&
		strings.Contains(probeErr.Error(), "fatal: not a git repository (or any of the parent directories): .git")
	if !nonGit {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Git inspection failed; directory fallback: %v", probeErr))
	}
	result.Kind = contracts.WorkspaceDirectory
	entries, truncated, warnings, err := listWorkspace(ctx, dir)
	result.Entries, result.Truncated = entries, truncated
	result.Warnings = append(result.Warnings, warnings...)
	return result, err
}

func findGitMetadata(dir string) (bool, error) {
	for {
		_, err := os.Lstat(filepath.Join(dir, ".git"))
		if err == nil {
			return true, nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return false, fmt.Errorf("inspect Git metadata: %w", err)
		}
		// A bare repository is an unsupported context, not a normal directory.
		if _, err = os.Stat(filepath.Join(dir, "HEAD")); err == nil {
			if st, e := os.Stat(filepath.Join(dir, "objects")); e == nil && st.IsDir() {
				return true, nil
			}
		} else if !errors.Is(err, os.ErrNotExist) {
			return false, fmt.Errorf("inspect repository metadata: %w", err)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return false, nil
		}
		dir = parent
	}
}

type limitedBuffer struct {
	buffer   bytes.Buffer
	limit    int
	exceeded bool
	kill     func()
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	if len(p) > b.limit-b.buffer.Len() {
		b.exceeded = true
		if b.kill != nil {
			b.kill()
		}
		return 0, errors.New("output limit exceeded")
	}
	return b.buffer.Write(p)
}
func runGit(ctx context.Context, dir string, args ...string) ([]byte, error) {
	cmdCtx, cancel := context.WithTimeout(ctx, gitTimeout)
	defer cancel()
	cmd := exec.CommandContext(cmdCtx, "git", args...)
	configureGitProcess(cmd)
	// A descendant may retain inherited pipes after the direct process exits.
	// Bound pipe draining after either cancellation or output-limit termination.
	cmd.WaitDelay = gitPipeDrainTimeout
	cmd.Dir = dir
	// Do not let a process-level Git context redirect inspection elsewhere.
	for _, variable := range os.Environ() {
		key, _, _ := strings.Cut(variable, "=")
		if !strings.HasPrefix(key, "GIT_") && key != "LC_ALL" {
			cmd.Env = append(cmd.Env, variable)
		}
	}
	cmd.Env = append(cmd.Env, "LC_ALL=C", "GIT_OPTIONAL_LOCKS=0", "GIT_DISCOVERY_ACROSS_FILESYSTEM=1")
	out := &limitedBuffer{limit: gitOutputLimit}
	stderr := &limitedBuffer{limit: gitErrorLimit}
	cmd.Stdout, cmd.Stderr = out, stderr
	kill := func() { _ = killGitProcess(cmd) }
	out.kill, stderr.kill = kill, kill
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	err := cmd.Wait()
	if cmdCtx.Err() != nil {
		return nil, cmdCtx.Err()
	}
	if out.exceeded || stderr.exceeded {
		return nil, errors.New("git output exceeded inspection limit")
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.buffer.String()))
	}
	return out.buffer.Bytes(), nil
}

//nolint:gocognit,nestif // Porcelain records and rename pairs require ordered validation.
func parseGitStatus(data []byte) ([]contracts.GitEntry, bool, error) {
	entries := make([]contracts.GitEntry, 0, min(workspaceMaxEntries, len(data)/gitMinimumRecordBytes))
	truncated := false
	next := func() (string, bool) {
		i := bytes.IndexByte(data, 0)
		if i < 0 {
			return "", false
		}
		record := string(data[:i])
		data = data[i+1:]
		return record, true
	}
	for len(data) > 0 {
		rec, ok := next()
		if !ok {
			return nil, false, errors.New("truncated Git status")
		}
		var e contracts.GitEntry
		if strings.HasPrefix(rec, "? ") {
			if len(rec) == len("? ") {
				return nil, false, errors.New("malformed untracked Git path")
			}
			e = contracts.GitEntry{IndexStatus: '?', WorktreeStatus: '?', Path: rec[2:]}
		} else {
			var count int
			switch {
			case strings.HasPrefix(rec, "1 "):
				count = 9
			case strings.HasPrefix(rec, "2 "):
				count = 10
			case strings.HasPrefix(rec, "u "):
				count = 11
			default:
				return nil, false, errors.New("malformed Git status record")
			}
			fields := strings.SplitN(rec, " ", count)
			if len(fields) != count || len(fields[1]) != 2 || len(fields[2]) != 4 || fields[count-1] == "" {
				return nil, false, errors.New("malformed Git status fields")
			}
			e = contracts.GitEntry{
				IndexStatus:    fields[1][0],
				WorktreeStatus: fields[1][1],
				Path:           fields[count-1],
				Conflict:       fields[0] == "u",
				Submodule:      fields[2][0] == 'S',
			}
			if fields[0] == "2" {
				origin, originOK := next()
				if !originOK || origin == "" {
					return nil, false, errors.New("malformed Git rename")
				}
				e.OriginalPath = origin
			}
		}
		if len(entries) < workspaceMaxEntries {
			entries = append(entries, e)
		} else {
			truncated = true
		}
	}
	return entries, truncated, nil
}

func listWorkspace(ctx context.Context, dir string) ([]contracts.DirectoryEntry, bool, []string, error) {
	f, err := os.Open(dir)
	if err != nil {
		return nil, false, nil, fmt.Errorf("open workspace: %w", err)
	}
	defer f.Close()
	entries := make([]contracts.DirectoryEntry, 0, workspaceMaxEntries)
	var warnings []string
	add := func(name string) {
		entry, e := directoryEntry(dir, name)
		if e != nil {
			warnings = append(warnings, fmt.Sprintf("%s: %v", name, e))
			return
		}
		entries = append(entries, entry)
	}
	add(".")
	add("..")
	truncated := false
	// Bound enumeration even when individual metadata calls fail.
	remaining := workspaceMaxEntries - workspaceDotEntries
	for remaining > 0 {
		if contextErr := ctx.Err(); contextErr != nil {
			return nil, false, nil, contextErr
		}
		batch, readErr := f.ReadDir(min(remaining, workspaceBatchSize))
		for _, entry := range batch {
			add(entry.Name())
			remaining--
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return entries, false, warnings, fmt.Errorf("read workspace directory: %w", readErr)
		}
	}
	if remaining == 0 {
		more, readErr := f.ReadDir(1)
		if len(more) > 0 {
			truncated = true
		}
		if readErr != nil && readErr != io.EOF {
			return entries, truncated, warnings, fmt.Errorf("read workspace directory: %w", readErr)
		}
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })
	return entries, truncated, warnings, nil
}
func directoryEntry(dir, name string) (contracts.DirectoryEntry, error) {
	path := filepath.Join(dir, name)
	info, err := os.Lstat(path)
	if err != nil {
		return contracts.DirectoryEntry{}, err
	}
	e := contracts.DirectoryEntry{Name: name, Mode: uint32(info.Mode()), Size: info.Size(), ModifiedAt: info.ModTime()}
	setDirectoryMetadata(&e, info)
	if info.Mode()&os.ModeSymlink != 0 {
		target, linkErr := os.Readlink(path)
		if linkErr != nil {
			return e, linkErr
		}
		e.SymlinkTarget = &target
	}
	return e, nil
}
