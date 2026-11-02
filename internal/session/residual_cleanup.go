package session

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BlackGiack/kankuro/internal/driver"
)

// RemoveResidual retries disk, workspace and history cleanup after a session row
// has already disappeared. The identity must come from a durable deletion manifest.
//
//nolint:gocognit // Report each independent residual cleanup effect.
func (m *Manager) RemoveResidual(ctx context.Context, id, workingDir, temporaryRoot string, useTmp bool) (RemovalResult, error) {
	if id == "" || id == "." || id == ".." || filepath.Base(id) != id {
		return RemovalResult{}, errors.New("invalid session identity")
	}
	if useTmp {
		root, err := savedTemporaryRoot(id, workingDir)
		if err != nil {
			return RemovalResult{}, err
		}
		if !filepath.IsAbs(temporaryRoot) || filepath.Clean(root) != filepath.Clean(temporaryRoot) {
			return RemovalResult{}, errors.New("temporary cleanup path does not match saved workspace")
		}
	}
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return RemovalResult{}, ErrManagerClosed
	}
	if _, exists := m.sessions[id]; exists {
		m.mu.Unlock()
		return RemovalResult{}, errors.New("session is still present")
	}
	if _, exists := m.launching[id]; exists {
		m.mu.Unlock()
		return RemovalResult{}, errors.New("session activation is in progress")
	}
	if m.launching == nil {
		m.launching = make(map[string]struct{})
	}
	m.launching[id] = struct{}{}
	m.activations.Add(1)
	m.mu.Unlock()
	defer func() { m.mu.Lock(); delete(m.launching, id); m.mu.Unlock(); m.activations.Done() }()
	if useTmp {
		if err := driver.RemoveWorkspaceTrust(ctx, workingDir); err != nil {
			return RemovalResult{}, fmt.Errorf("%w: %w", ErrTrustCleanup, err)
		}
	}
	deleteErr := deleteSession(filepath.Join(m.configDir, "sessions"), id)
	if errors.Is(deleteErr, os.ErrNotExist) {
		deleteErr = nil
	}
	var removeErr error
	if useTmp {
		removeErr = os.RemoveAll(temporaryRoot)
	}
	if deleteErr != nil || removeErr != nil {
		return RemovalResult{
			SessionRecordDeleted:      deleteErr == nil,
			TemporaryWorkspaceRemoved: useTmp && removeErr == nil,
		}, errors.Join(
			deleteErr,
			removeErr,
		)
	}
	if m.messages != nil {
		if err := m.messages.DeleteRecipient(id); err != nil {
			return RemovalResult{SessionRecordDeleted: true, TemporaryWorkspaceRemoved: useTmp}, err
		}
	}
	return RemovalResult{InMemoryRemoved: true, SessionRecordDeleted: true, TemporaryWorkspaceRemoved: useTmp}, nil
}
