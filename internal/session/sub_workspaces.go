package session

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/uuid"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/fsutil"
)

const (
	subWorkspaceManifestVersion = 1
	subWorkspaceManifestMode    = 0o600
)

type subWorkspaceManifest struct {
	SchemaVersion int                         `json:"schema_version"`
	Copies        map[string]subWorkspaceCopy `json:"copies"`
}

// SubWorkspaceDescriptor identifies a deployed copy, not a current card selection.
type SubWorkspaceDescriptor struct {
	WorkspaceID     string `json:"workspace_id"`
	SourcePath      string `json:"source_path"`
	DestinationPath string `json:"destination_path"`
}

// ReadSubWorkspaceDescriptors reads the workspace manifest without preparing or
// recreating anything. The availability result distinguishes absent/invalid legacy metadata.
func ReadSubWorkspaceDescriptors(workingDir string) ([]SubWorkspaceDescriptor, bool) {
	if validateSessionWorkspace(workingDir) != nil {
		return nil, false
	}
	manifestPath := filepath.Join(workingDir, ".kankuro", "sub-workspaces.json")
	manifest, err := loadSubWorkspaceManifest(manifestPath)
	if err != nil {
		return nil, false
	}
	if _, statErr := os.Stat(manifestPath); statErr != nil {
		return nil, false
	}
	names := make([]string, 0, len(manifest.Copies))
	for name := range manifest.Copies {
		names = append(names, name)
	}
	sort.Strings(names)
	descriptors := make([]SubWorkspaceDescriptor, 0, len(names))
	for _, name := range names {
		entry := manifest.Copies[name]
		destination, absErr := filepath.Abs(filepath.Join(workingDir, "tmp", name))
		if absErr != nil {
			return nil, false
		}
		descriptors = append(descriptors, SubWorkspaceDescriptor{
			WorkspaceID: entry.WorkspaceID, SourcePath: entry.SourcePath, DestinationPath: destination,
		})
	}
	return descriptors, true
}

func preparedSubWorkspaceDescriptors(workingDir string) ([]SubWorkspaceDescriptor, bool) {
	descriptors, available := ReadSubWorkspaceDescriptors(workingDir)
	if available {
		return descriptors, true
	}
	manifestPath := filepath.Join(workingDir, ".kankuro", "sub-workspaces.json")
	if _, err := os.Stat(manifestPath); errors.Is(err, os.ErrNotExist) {
		return nil, true
	}
	return nil, false
}

type subWorkspaceCopy struct {
	WorkspaceID string `json:"workspace_id"`
	SourcePath  string `json:"source_path"`
}

type subWorkspaceSource struct {
	id   string
	path string
	name string
}

type subWorkspaceTransaction struct {
	stages    []string
	published []string
}

// prepareSubWorkspaces serializes manifest decisions and copy publication for a working directory.
func (m *Manager) prepareSubWorkspaces(ctx context.Context, card config.IDCard, workingDir string) error {
	if len(card.SubWorkspaceIDs) == 0 {
		return nil
	}
	select {
	case m.subWorkspaceGate <- struct{}{}:
		defer func() { <-m.subWorkspaceGate }()
	case <-ctx.Done():
		return ctx.Err()
	}

	canonicalDir, canonicalErr := canonicalDirectory(workingDir)
	if canonicalErr != nil {
		return fmt.Errorf("resolve session workspace: %w", canonicalErr)
	}
	sources, resolveErr := resolveSubWorkspaceSources(m.repo, card.SubWorkspaceIDs, canonicalDir)
	if resolveErr != nil {
		return resolveErr
	}
	manifestPath := filepath.Join(canonicalDir, ".kankuro", "sub-workspaces.json")
	manifest, manifestErr := loadSubWorkspaceManifest(manifestPath)
	if manifestErr != nil {
		return manifestErr
	}
	tmpDir := filepath.Join(canonicalDir, "tmp")
	if prepareErr := prepareSubWorkspaceDirectory(tmpDir); prepareErr != nil {
		return prepareErr
	}

	tx := &subWorkspaceTransaction{}
	updated := copySubWorkspaceManifest(manifest, len(sources))
	for _, source := range sources {
		shouldCopy, ownershipErr := shouldCopySubWorkspace(source, tmpDir, manifest)
		if ownershipErr != nil {
			return tx.rollback(ownershipErr)
		}
		if !shouldCopy {
			continue
		}
		if copyErr := copyAndPublishSubWorkspace(ctx, source, tmpDir, tx); copyErr != nil {
			return tx.rollback(copyErr)
		}
		updated.Copies[source.name] = subWorkspaceCopy{WorkspaceID: source.id, SourcePath: source.path}
	}
	if persistErr := persistSubWorkspaceManifest(manifestPath, updated); persistErr != nil {
		return tx.rollback(persistErr)
	}
	return nil
}

func resolveSubWorkspaceSources(repo WorkspaceReader, ids []string, workingDir string) ([]subWorkspaceSource, error) {
	workspaces, listErr := repo.ListWorkspaces()
	if listErr != nil {
		return nil, fmt.Errorf("list sub workspaces: %w", listErr)
	}
	byID := make(map[string]config.Workspace, len(workspaces))
	for _, workspace := range workspaces {
		byID[workspace.ID] = workspace
	}

	sources := make([]subWorkspaceSource, 0, len(ids))
	usedNames := make(map[string]bool, len(ids))
	tmpDir := filepath.Join(workingDir, "tmp")
	for _, id := range ids {
		workspace, ok := byID[id]
		if !ok {
			return nil, fmt.Errorf("sub workspace not found: %s", id)
		}
		resolved, err := canonicalDirectory(workspace.Path)
		if err != nil {
			return nil, fmt.Errorf("resolve sub workspace %s: %w", id, err)
		}
		info, err := os.Stat(resolved)
		if err != nil {
			return nil, fmt.Errorf("sub workspace %s unavailable: %w", id, err)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("sub workspace %s is not a directory", id)
		}
		if containmentErr := ensureDestinationOutsideSource(resolved, tmpDir, id); containmentErr != nil {
			return nil, containmentErr
		}
		name := filepath.Base(filepath.Clean(workspace.Path))
		if !safeSubWorkspaceName(name) || usedNames[name] {
			return nil, fmt.Errorf("sub workspace destination collision: %s", name)
		}
		usedNames[name] = true
		sources = append(sources, subWorkspaceSource{id: id, path: resolved, name: name})
	}
	return sources, nil
}

func canonicalDirectory(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(absolute)
}

func ensureDestinationOutsideSource(source, destination, workspaceID string) error {
	rel, err := filepath.Rel(source, destination)
	if err != nil {
		return fmt.Errorf("check sub workspace containment %s: %w", workspaceID, err)
	}
	if rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))) {
		return fmt.Errorf("sub workspace contains its destination: %s", workspaceID)
	}
	return nil
}

func loadSubWorkspaceManifest(path string) (subWorkspaceManifest, error) {
	metadataDir := filepath.Dir(path)
	if err := rejectSymlink(metadataDir); err != nil {
		return subWorkspaceManifest{}, err
	}
	if err := rejectSymlink(path); err != nil {
		return subWorkspaceManifest{}, err
	}
	manifest := subWorkspaceManifest{
		SchemaVersion: subWorkspaceManifestVersion,
		Copies:        make(map[string]subWorkspaceCopy),
	}
	data, readErr := os.ReadFile(path)
	if errors.Is(readErr, os.ErrNotExist) {
		return manifest, nil
	}
	if readErr != nil {
		return subWorkspaceManifest{}, fmt.Errorf("read sub workspace manifest: %w", readErr)
	}
	if unmarshalErr := json.Unmarshal(data, &manifest); unmarshalErr != nil {
		return subWorkspaceManifest{}, fmt.Errorf("read sub workspace manifest: %w", unmarshalErr)
	}
	if manifest.SchemaVersion != subWorkspaceManifestVersion || manifest.Copies == nil {
		return subWorkspaceManifest{}, errors.New("unsupported sub workspace manifest")
	}
	for name := range manifest.Copies {
		if !safeSubWorkspaceName(name) {
			return subWorkspaceManifest{}, fmt.Errorf("unsafe sub workspace manifest destination: %s", name)
		}
	}
	return manifest, nil
}

func safeSubWorkspaceName(name string) bool {
	return name != "" && name != "." && name != ".." && filepath.Base(name) == name && !strings.ContainsAny(name, `/\\`)
}

func prepareSubWorkspaceDirectory(path string) error {
	if err := rejectSymlink(path); err != nil {
		return err
	}
	if err := os.MkdirAll(path, 0o700); err != nil {
		return fmt.Errorf("create sub workspace directory: %w", err)
	}
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return errors.New("sub workspace tmp path is not a directory")
	}
	return nil
}

func copySubWorkspaceManifest(manifest subWorkspaceManifest, additional int) subWorkspaceManifest {
	updated := subWorkspaceManifest{
		SchemaVersion: subWorkspaceManifestVersion,
		Copies:        make(map[string]subWorkspaceCopy, len(manifest.Copies)+additional),
	}
	maps.Copy(updated.Copies, manifest.Copies)
	return updated
}

func shouldCopySubWorkspace(source subWorkspaceSource, tmpDir string, manifest subWorkspaceManifest) (bool, error) {
	destination := filepath.Join(tmpDir, source.name)
	if err := rejectSymlink(destination); err != nil {
		return false, err
	}
	entry, managed := manifest.Copies[source.name]
	info, err := os.Lstat(destination)
	if err == nil {
		if !managed || entry.WorkspaceID != source.id || entry.SourcePath != source.path {
			return false, fmt.Errorf("sub workspace destination collision: %s", source.name)
		}
		if !info.IsDir() {
			return false, fmt.Errorf("managed sub workspace is not a directory: %s", source.name)
		}
		return false, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return false, err
	}
	if managed && (entry.WorkspaceID != source.id || entry.SourcePath != source.path) {
		return false, fmt.Errorf("sub workspace destination collision: %s", source.name)
	}
	return true, nil
}

func copyAndPublishSubWorkspace(ctx context.Context, source subWorkspaceSource, tmpDir string, tx *subWorkspaceTransaction) error {
	stage := filepath.Join(tmpDir, ".kankuro-stage-"+uuid.NewString())
	tx.stages = append(tx.stages, stage)
	if err := os.Mkdir(stage, 0o700); err != nil {
		return fmt.Errorf("create sub workspace staging directory: %w", err)
	}
	//nolint:gosec // explicit argv keeps registered path arguments out of shell interpretation.
	cmd := exec.CommandContext(ctx, "cp", "-a", "--", source.path+string(filepath.Separator)+".", stage)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("copy sub workspace %s: %s: %w", source.id, output, err)
	}
	stageInfo, stageErr := os.Lstat(stage)
	if stageErr != nil || !stageInfo.IsDir() || stageInfo.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("sub workspace staging result is not a real directory: %s", source.name)
	}
	destination := filepath.Join(tmpDir, source.name)
	if _, err := os.Lstat(destination); err == nil {
		return fmt.Errorf("sub workspace destination collision: %s", source.name)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.Rename(stage, destination); err != nil {
		return fmt.Errorf("publish sub workspace %s: %w", source.id, err)
	}
	tx.stages = tx.stages[:len(tx.stages)-1]
	tx.published = append(tx.published, destination)
	return nil
}

func persistSubWorkspaceManifest(path string, manifest subWorkspaceManifest) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	encoded, marshalErr := json.MarshalIndent(manifest, "", "  ")
	if marshalErr != nil {
		return marshalErr
	}
	if writeErr := fsutil.WriteFileAtomic(path, encoded, subWorkspaceManifestMode); writeErr != nil {
		return fmt.Errorf("write sub workspace manifest: %w", writeErr)
	}
	return nil
}

func (tx *subWorkspaceTransaction) rollback(cause error) error {
	var cleanupErrs []error
	for _, path := range tx.stages {
		cleanupErrs = append(cleanupErrs, os.RemoveAll(path))
	}
	for _, path := range tx.published {
		cleanupErrs = append(cleanupErrs, os.RemoveAll(path))
	}
	return errors.Join(cause, errors.Join(cleanupErrs...))
}

func rejectSymlink(path string) error {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect sub workspace path %s: %w", path, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("sub workspace path is a symlink: %s", path)
	}
	return nil
}
