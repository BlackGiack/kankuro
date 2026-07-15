package config

import "fmt"

// SyncFailedError reports that a source-list mutation (add/edit/delete) was
// itself durably saved, but the derived registry built from that list (e.g.
// skills.yaml from skills_sources.yaml) failed to rebuild - for example
// because one of the sources is an unreadable or nonexistent directory, or a
// path with an undefined environment variable (see expandPath).
//
// Callers should treat this differently from a hard failure: the mutation
// the user asked for did happen and the UI must refresh to show it, but the
// user also needs to know the associated registry is now stale.
type SyncFailedError struct {
	// Err is the underlying sync error (e.g. from ScanEntries).
	Err error
}

func (e *SyncFailedError) Error() string {
	return fmt.Sprintf("source saved, but sync failed: %v", e.Err)
}

func (e *SyncFailedError) Unwrap() error { return e.Err }
