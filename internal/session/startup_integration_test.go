package session_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

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

type testStartupTimer struct{ ready chan time.Time }

func (t *testStartupTimer) C() <-chan time.Time { return t.ready }
func (*testStartupTimer) Stop() bool            { return true }

type testStartupClock struct {
	timer   *testStartupTimer
	entered chan struct{}
}

func (*testStartupClock) Now() time.Time { return time.Unix(100, 0) }
func (c *testStartupClock) NewTimer(_ time.Duration) session.StartupTimer {
	close(c.entered)
	return c.timer
}

func TestActivationPublishedBeforeStartupDelay(t *testing.T) {
	repo, workspaceID := workspaceValidationRepository(t)
	registry := driver.NewRegistry()
	registry.Register("cat", func() driver.AgentDriver { return &startupDriver{driver.NewGenericCLIDriver(), make(chan string, 1)} })
	clock := &testStartupClock{timer: &testStartupTimer{ready: make(chan time.Time, 1)}, entered: make(chan struct{})}
	manager := newTestManager(
		t,
		repo.Dir(),
		repo,
		session.WithDriverRegistry(registry),
		session.WithStartupClock(clock),
		session.WithStartupDelay(10*time.Second),
	)
	id := uuid.NewString()
	_, err := manager.LaunchWithIDContext(
		context.Background(),
		id,
		config.IDCard{ID: "a", Name: "a", Driver: "cat", WorkspaceID: workspaceID, UseTmpWorkspace: true},
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	<-clock.entered
	snap, ok := manager.GetSession(id)
	if !ok || !snap.UseTmpWorkspace || !snap.ProcessStarted || snap.Phase != "starting" || snap.MessageDeliveryAvailable {
		t.Fatalf("unexpected startup snapshot: %+v", snap)
	}
	if handle, attachErr := manager.Attach(id); attachErr != nil {
		t.Fatal(attachErr)
	} else {
		handle.Close()
	}
	clock.timer.ready <- time.Unix(110, 0)
	deadline := time.After(2 * time.Second)
	for {
		snap, _ = manager.GetSession(id)
		if snap.Phase == "active" {
			break
		}
		select {
		case <-deadline:
			t.Fatalf("startup transition did not complete: %+v", snap)
		default:
			time.Sleep(time.Millisecond)
		}
	}
	if snap.MessageDeliveryAvailable {
		t.Fatal("MCP-disabled activation accepted delivery")
	}
}

//nolint:gocognit // The table checks publication, disk persistence, and cleanup together.
func TestCreationNameIsSavedAtPublication(t *testing.T) {
	repo, workspaceID := workspaceValidationRepository(t)
	registry := driver.NewRegistry()
	registry.Register("cat", func() driver.AgentDriver { return &startupDriver{driver.NewGenericCLIDriver(), make(chan string, 1)} })
	manager := newTestManager(t, repo.Dir(), repo, session.WithDriverRegistry(registry))
	card := config.IDCard{ID: "card", Name: "default", Driver: "cat", WorkspaceID: workspaceID}
	for _, tc := range []struct {
		label string
		name  *string
		want  string
	}{
		{label: "omitted", want: "default"},
		{label: "unicode", name: new("  仕事  "), want: "仕事"},
		{label: "empty", name: new("  ")},
	} {
		t.Run(tc.label, func(t *testing.T) {
			id := uuid.NewString()
			if _, err := manager.LaunchWithIDContext(context.Background(), id, card, nil, nil, nil, tc.name); err != nil {
				t.Fatal(err)
			}
			want := tc.want
			if tc.name != nil && want == "" {
				want = id
			}
			snapshot, ok := manager.GetSession(id)
			if !ok || snapshot.Name != want {
				t.Fatalf("snapshot: %+v, want %q", snapshot, want)
			}
			raw, err := os.ReadFile(filepath.Join(repo.Dir(), "sessions", id+".json"))
			if err != nil {
				t.Fatal(err)
			}
			var record struct {
				Name string `json:"name"`
			}
			if decodeErr := json.Unmarshal(raw, &record); decodeErr != nil {
				t.Fatal(decodeErr)
			}
			if record.Name != want {
				t.Fatalf("saved name = %q, want %q", record.Name, want)
			}
			if killErr := manager.Kill(id); killErr != nil {
				t.Fatal(killErr)
			}
		})
	}
	if card.Name != "default" {
		t.Fatalf("card was modified: %+v", card)
	}
}
