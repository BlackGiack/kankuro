package api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/config"
)

type sourceRepository interface {
	ListSkillSources() ([]string, error)
	ListPersonaSources() ([]string, error)
	ListPuppetSources() ([]string, error)
	ListWorkspaceSources() ([]config.WorkspaceSource, error)
	UpdateSkillSources(func([]string) ([]string, error)) error
	UpdatePersonaSources(func([]string) ([]string, error)) error
	UpdatePuppetSources(func([]string) ([]string, error)) error
	UpdateWorkspaceSources(func([]config.WorkspaceSource) ([]config.WorkspaceSource, error)) error
}

var errSourceConflict = errors.New("source changed or no longer exists; reload sources and try again")

func sourceToken(value any) string {
	data, _ := json.Marshal(value)
	sum := sha256.Sum256(data)
	return "v1:" + hex.EncodeToString(sum[:])
}

func pathIdentity(kind contracts.SourceKind, paths []string, index int) string {
	return sourceToken(struct {
		Kind  contracts.SourceKind
		Paths []string
		Index int
	}{kind, paths, index})
}

func workspaceIdentity(source config.WorkspaceSource) string {
	return source.ID + ":" + sourceToken(source)
}

// ReadSources returns detached rows with identities tied to the observed contents.
func (s *Service) ReadSources(kind contracts.SourceKind) ([]contracts.SourceRow, error) {
	repo, ok := s.repo.(sourceRepository)
	if !ok {
		return nil, errors.New("source repository unavailable")
	}
	if kind == contracts.SourceWorkspaces {
		sources, err := repo.ListWorkspaceSources()
		if err != nil {
			return nil, err
		}
		rows := make([]contracts.SourceRow, 0, len(sources))
		for _, source := range sources {
			rows = append(rows, contracts.SourceRow{Identity: workspaceIdentity(source), Path: source.Path, IsRoot: source.IsRoot})
		}
		return rows, nil
	}
	var paths []string
	var err error
	switch kind { //nolint:exhaustive // Workspaces are handled above.
	case contracts.SourceSkills:
		paths, err = repo.ListSkillSources()
	case contracts.SourcePersonas:
		paths, err = repo.ListPersonaSources()
	case contracts.SourcePuppets:
		paths, err = repo.ListPuppetSources()
	default:
		return nil, fmt.Errorf("unknown source kind %q", kind)
	}
	if err != nil {
		return nil, err
	}
	rows := make([]contracts.SourceRow, 0, len(paths))
	for i, path := range paths {
		rows = append(rows, contracts.SourceRow{Identity: pathIdentity(kind, paths, i), Path: path})
	}
	return rows, nil
}

// MutateSource validates the observed identity under the repository's update lock.
func (s *Service) MutateSource(ctx context.Context, input contracts.SourceMutation) contracts.SourceMutationResult {
	if err := ctx.Err(); err != nil {
		return contracts.SourceMutationResult{Err: err}
	}
	repo, ok := s.repo.(sourceRepository)
	if !ok {
		return contracts.SourceMutationResult{Err: errors.New("source repository unavailable")}
	}
	if !input.Delete && strings.TrimSpace(input.Path) == "" {
		return contracts.SourceMutationResult{Err: errors.New("source path is required")}
	}
	if input.Delete && input.Identity == "" {
		return contracts.SourceMutationResult{Err: errSourceConflict, Reload: true}
	}
	var err error
	if input.Kind == contracts.SourceWorkspaces {
		err = updateWorkspaceSource(ctx, repo, input)
	} else {
		err = updatePathSource(ctx, repo, input)
	}

	if err == nil {
		return contracts.SourceMutationResult{Saved: true, Refresh: true}
	}
	if _, syncFailed := errors.AsType[*config.SyncFailedError](err); syncFailed {
		return contracts.SourceMutationResult{Saved: true, Refresh: true, Warning: err}
	}
	return contracts.SourceMutationResult{Err: err, Reload: errors.Is(err, errSourceConflict)}
}

func updateWorkspaceSource(ctx context.Context, repo sourceRepository, input contracts.SourceMutation) error {
	return repo.UpdateWorkspaceSources(func(sources []config.WorkspaceSource) ([]config.WorkspaceSource, error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if input.Identity == "" {
			return append(sources, config.WorkspaceSource{Path: input.Path, IsRoot: input.IsRoot}), nil
		}
		for i, source := range sources {
			if workspaceIdentity(source) != input.Identity {
				continue
			}
			if input.Delete {
				return append(sources[:i], sources[i+1:]...), nil
			}
			sources[i].Path, sources[i].IsRoot = input.Path, input.IsRoot
			return sources, nil
		}
		return nil, errSourceConflict
	})
}

func updatePathSource(ctx context.Context, repo sourceRepository, input contracts.SourceMutation) error {
	mutate := func(paths []string) ([]string, error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if input.Identity == "" {
			return append(paths, input.Path), nil
		}
		for i := range paths {
			if pathIdentity(input.Kind, paths, i) != input.Identity {
				continue
			}
			if input.Delete {
				return append(paths[:i], paths[i+1:]...), nil
			}
			paths[i] = input.Path
			return paths, nil
		}
		return nil, errSourceConflict
	}
	switch input.Kind { //nolint:exhaustive // Workspaces use their stable IDs above.
	case contracts.SourceSkills:
		return repo.UpdateSkillSources(mutate)
	case contracts.SourcePersonas:
		return repo.UpdatePersonaSources(mutate)
	case contracts.SourcePuppets:
		return repo.UpdatePuppetSources(mutate)
	default:
		return fmt.Errorf("unknown source kind %q", input.Kind)
	}
}

type sourceSyncRepository interface {
	SyncWorkspacesFromSources() error
	SyncPuppetsFromSources() error
}

// SyncSources rescans the source kinds that support explicit synchronization.
func (s *Service) SyncSources(ctx context.Context, kind contracts.SourceKind) contracts.SourceSyncResult {
	if err := ctx.Err(); err != nil {
		return contracts.SourceSyncResult{Err: err}
	}
	repo, ok := s.repo.(sourceSyncRepository)
	if !ok {
		return contracts.SourceSyncResult{Err: errors.New("source synchronization unavailable")}
	}
	var err error
	switch kind {
	case contracts.SourceWorkspaces:
		err = repo.SyncWorkspacesFromSources()
	case contracts.SourcePuppets:
		err = repo.SyncPuppetsFromSources()
	case contracts.SourceSkills, contracts.SourcePersonas:
	default:
		return contracts.SourceSyncResult{Err: fmt.Errorf("unknown source kind %q", kind)}
	}
	return contracts.SourceSyncResult{Refresh: true, Err: err}
}
