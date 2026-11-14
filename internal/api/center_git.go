package api

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
)

const (
	privateGitFileMode = 0600
	privateGitDirMode  = 0700
	gitSHA1HexLength   = 40
	gitSHA256HexLength = 64
)

func inspectGitWorkspace(ctx context.Context, result *contracts.SessionWorkspaceResult) error {
	ctx, cancel := context.WithTimeout(ctx, gitTimeout)
	defer cancel()

	data, err := runGit(ctx, result.WorkingDir, "-c", "core.fsmonitor=false", "rev-parse", "--path-format=absolute",
		"--show-toplevel", "--absolute-git-dir", "--git-common-dir", "--shared-index-path")
	if err != nil {
		return err
	}
	paths := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	const requiredPaths = 3
	if len(paths) < requiredPaths || len(paths) > requiredPaths+1 ||
		!filepath.IsAbs(paths[0]) || !filepath.IsAbs(paths[1]) || !filepath.IsAbs(paths[2]) {
		return errors.New("git returned invalid workspace metadata paths")
	}
	root, gitDir, commonDir := paths[0], paths[1], paths[2]
	var sharedIndex string
	if len(paths) > requiredPaths {
		sharedIndex = paths[requiredPaths]
	}
	workspacePath, err := filepath.EvalSymlinks(result.WorkingDir)
	if err != nil {
		return err
	}
	relative, err := filepath.Rel(root, workspacePath)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return errors.New("git worktree does not contain the saved workspace")
	}
	result.RepositoryRoot = root
	settings, err := gitStatusSettings(ctx, root)
	if err != nil {
		return err
	}
	private, err := os.MkdirTemp("", "kankuro-git-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(private)

	if linkErr := linkGitStatusMetadata(private, gitDir, commonDir, sharedIndex); linkErr != nil {
		return linkErr
	}
	env := []string{"GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=" + os.DevNull, "GIT_NO_LAZY_FETCH=1"}
	args := []string{"--git-dir=" + private, "--work-tree=" + root,
		"-c", "core.fsmonitor=false", "-c", "core.hooksPath=" + os.DevNull, "-c", "submodule.recurse=false"}
	args = append(args, settings...)
	index, err := runGitWithEnv(ctx, root, env, append(append([]string(nil), args...), "ls-files", "--stage", "-z")...)
	if err != nil {
		return fmt.Errorf("inspect Git index: %w", err)
	}
	for record := range bytes.SplitSeq(index, []byte{0}) {
		if bytes.HasPrefix(record, []byte("160000 ")) {
			return errors.New("background Git inspection skips repositories with submodules")
		}
	}
	data, err = runGitWithEnv(ctx, root, env, append(args, "status", "--porcelain=v2", "-z",
		"--untracked-files=all", "--ignore-submodules=all")...)
	if err != nil {
		return fmt.Errorf("git status: %w", err)
	}
	entries, truncated, err := parseGitStatus(data)
	if err != nil {
		return fmt.Errorf("git status: %w", err)
	}
	result.Kind, result.GitEntries, result.Truncated = contracts.WorkspaceGit, entries, truncated
	return nil
}

// Read effective settings once. Status runs with only the allowlisted values;
// it never rereads the original repository, system or user configuration.
func gitStatusSettings(ctx context.Context, root string) ([]string, error) {
	data, err := runGit(ctx, root, "config", "--null", "--list", "--includes")
	if err != nil {
		return nil, fmt.Errorf("inspect Git configuration: %w", err)
	}
	var settings []string
	for record := range strings.SplitSeq(strings.TrimSuffix(string(data), "\x00"), "\x00") {
		key, value, hasValue := strings.Cut(record, "\n")
		if strings.HasPrefix(key, "filter.") && (strings.HasSuffix(key, ".clean") || strings.HasSuffix(key, ".process")) {
			return nil, errors.New("background Git inspection skips repositories with configured clean filters")
		}
		switch key {
		case "core.repositoryformatversion", "extensions.objectformat", "core.filemode", "core.ignorecase",
			"core.symlinks", "core.autocrlf", "core.eol", "core.excludesfile", "core.attributesfile",
			"core.checkstat", "core.trustctime", "status.renames", "status.renamelimit", "diff.renamelimit":
			if !hasValue {
				value = "true"
			}
			settings = append(settings, "-c", key+"="+value)
		case "core.sparsecheckout", "extensions.refstorage", "extensions.partialclone":
			return nil, fmt.Errorf("background Git inspection does not support %s", key)
		}
	}
	return settings, nil
}

// Link individual data paths into a private Git directory. In particular, no
// config, config.worktree, commondir, hooks or submodule metadata is linked.
func linkGitStatusMetadata(private, gitDir, commonDir, sharedIndex string) error {
	headPath := filepath.Join(gitDir, "HEAD")
	info, err := os.Lstat(headPath)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Size() > gitErrorLimit {
		return errors.New("unsupported Git HEAD file")
	}
	head, err := os.ReadFile(headPath)
	if err != nil {
		return err
	}
	privateRoot, err := os.OpenRoot(private)
	if err != nil {
		return err
	}
	defer privateRoot.Close()
	if writeErr := privateRoot.WriteFile("HEAD", head, privateGitFileMode); writeErr != nil {
		return writeErr
	}
	if mkdirErr := privateRoot.Mkdir("info", privateGitDirMode); mkdirErr != nil {
		return mkdirErr
	}
	for _, item := range []struct{ name, dir string }{
		{"index", gitDir}, {"objects", commonDir}, {"refs", commonDir},
		{"packed-refs", commonDir}, {"info/exclude", commonDir}, {"info/attributes", commonDir},
	} {
		source := filepath.Join(item.dir, item.name)
		if _, statErr := os.Stat(source); errors.Is(statErr, os.ErrNotExist) {
			continue
		} else if statErr != nil {
			return statErr
		}
		if linkErr := privateRoot.Symlink(source, item.name); linkErr != nil {
			return linkErr
		}
	}
	return linkSharedGitIndex(privateRoot, gitDir, sharedIndex)
}

func linkSharedGitIndex(private *os.Root, gitDir, path string) error {
	if path == "" {
		return nil
	}
	name := filepath.Base(path)
	hash, hasPrefix := strings.CutPrefix(name, "sharedindex.")
	if !hasPrefix || (len(hash) != gitSHA1HexLength && len(hash) != gitSHA256HexLength) || filepath.Join(gitDir, name) != path {
		return errors.New("git returned an invalid shared index path")
	}
	if _, err := hex.DecodeString(hash); err != nil {
		return errors.New("git returned an invalid shared index name")
	}
	info, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("inspect Git shared index: %w", err)
	}
	if !info.Mode().IsRegular() {
		return errors.New("unsupported Git shared index file")
	}
	return private.Symlink(path, name)
}
