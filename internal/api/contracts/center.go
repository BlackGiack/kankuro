package contracts

import (
	"errors"
	"time"
)

var ErrSessionScope = errors.New("session is outside dashboard scope")

type SessionReportsRequest struct {
	ScopedSessionRequest

	Limit int
}

type SessionReport struct {
	ID, MessageID, AuthorSessionID, AuthorActivationID, Kind, Text string
	CreatedAt                                                      time.Time
	Sequence                                                       uint64
}

type SessionReportsResult struct {
	Reports                 []SessionReport
	Limit, Maximum, Omitted int
}

// ReportSubscription delivers coalesced changes until Release or cancellation.
// Err reports why the signal closed, and is safe to call after closure.
type ReportSubscription interface {
	Changes() <-chan struct{}
	Err() error
	Release()
}

// WorkspaceKind distinguishes a clean Git tree from a directory listing.
type WorkspaceKind string

const (
	WorkspaceGit       WorkspaceKind = "git"
	WorkspaceDirectory WorkspaceKind = "directory"
)

type GitEntry struct {
	IndexStatus, WorktreeStatus byte
	Path, OriginalPath          string
	Conflict, Submodule         bool
}
type DirectoryEntry struct {
	Name          string
	Mode          uint32
	LinkCount     *uint64
	Owner, Group  *uint32
	Size          int64
	ModifiedAt    time.Time
	SymlinkTarget *string
}
type SessionWorkspaceResult struct {
	WorkingDir, RepositoryRoot string
	Kind                       WorkspaceKind
	GitEntries                 []GitEntry
	Entries                    []DirectoryEntry
	Truncated                  bool
	Warnings                   []string
}
