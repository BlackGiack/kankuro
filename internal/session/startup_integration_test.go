package session_test

import (
	"context"
	"errors"
	"testing"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/driver"
	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/google/uuid"
)

type startupDriver struct {
	*driver.GenericCLIDriver

	prepared chan string
}

func (d *startupDriver) PrepareWorkspace(ctx context.Context, path string) error {
	d.prepared <- path
	return ctx.Err()
}
func (d *startupDriver) Spawn(cfg driver.AgentConfig) (*driver.SessionHandle, error) {
	select {
	case path := <-d.prepared:
		if path != cfg.WorkingDir {
			return nil, errors.New("wrong workspace")
		}
	default:
		return nil, errors.New("spawn before trust preparation")
	}
	return d.GenericCLIDriver.Spawn(cfg)
}

func TestActivationWaitsBeforePublication(t *testing.T) {
	repo, workspaceID := workspaceValidationRepository(t)
	registry := driver.NewRegistry()
	registry.Register("cat", func() driver.AgentDriver { return &startupDriver{driver.NewGenericCLIDriver(), make(chan string, 1)} })
	entered, release := make(chan struct{}), make(chan struct{})
	wait := func(ctx context.Context, shutdown, done <-chan struct{}, _ <-chan []byte) ([]byte, error) {
		close(entered)
		select {
		case <-release:
			return nil, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-shutdown:
			return nil, session.ErrManagerClosed
		case <-done:
			return nil, errors.New("process exited")
		}
	}
	manager := newTestManager(t, repo.Dir(), repo, session.WithDriverRegistry(registry), session.WithStartupWait(wait))
	id := uuid.NewString()
	result := make(chan error, 1)
	go func() {
		_, err := manager.LaunchWithIDContext(
			context.Background(),
			id,
			config.IDCard{ID: "a", Name: "a", Driver: "cat", WorkspaceID: workspaceID, UseTmpWorkspace: true},
			nil,
			nil,
			nil,
		)
		result <- err
	}()
	<-entered
	if _, ok := manager.GetSession(id); ok {
		t.Fatal("session published during startup")
	}
	close(release)
	if err := <-result; err != nil {
		t.Fatal(err)
	}
	snap, ok := manager.GetSession(id)
	if !ok || !snap.UseTmpWorkspace {
		t.Fatal("session was not published")
	}
}
