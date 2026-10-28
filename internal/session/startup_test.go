//nolint:testpackage // Exercises the unexported startup wait directly.
package session

import (
	"context"
	"errors"
	"testing"
	"time"
)

//nolint:gocritic // The branches select distinct interruption signals.
func TestStartupWait(t *testing.T) {
	t.Run("preserves output", func(t *testing.T) {
		output := make(chan []byte, 1)
		output <- []byte("startup output")
		got, err := waitForStartupDuration(context.Background(), make(chan struct{}), make(chan struct{}), output, 10*time.Millisecond)
		if err != nil || string(got) != "startup output" {
			t.Fatalf("got %q, %v", got, err)
		}
	})
	for _, tc := range []struct {
		name     string
		cancel   bool
		shutdown bool
	}{{"cancellation", true, false}, {"shutdown", false, true}, {"process exit", false, false}} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			shutdown, done := make(chan struct{}), make(chan struct{})
			if tc.cancel {
				cancel()
			} else if tc.shutdown {
				close(shutdown)
			} else {
				close(done)
			}
			_, err := waitForStartupDuration(ctx, shutdown, done, nil, time.Hour)
			if err == nil {
				t.Fatal("expected interruption")
			}
			if tc.cancel && !errors.Is(err, context.Canceled) {
				t.Fatalf("got %v", err)
			}
			if tc.shutdown && !errors.Is(err, ErrManagerClosed) {
				t.Fatalf("got %v", err)
			}
		})
	}
}
