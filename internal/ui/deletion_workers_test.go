//nolint:testpackage // Exercises package-private resource ownership.
package ui

import (
	"context"
	"testing"
	"time"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
)

func TestDeletionWorkersCancelAndJoin(t *testing.T) {
	w := &deletionWorkers{}
	started, exited := make(chan struct{}), make(chan struct{})
	go func() {
		defer close(exited)
		_, _ = w.run(func(ctx context.Context) (contracts.DeletionProgress, error) {
			close(started)
			<-ctx.Done()
			return contracts.DeletionProgress{Outcome: contracts.DeletionPendingOutcome}, ctx.Err()
		})
	}()
	<-started
	w.close()
	select {
	case <-exited:
	case <-time.After(time.Second):
		t.Fatal("worker was not joined")
	}
	if _, err := w.run(func(context.Context) (contracts.DeletionProgress, error) {
		t.Fatal("started after shutdown")
		return contracts.DeletionProgress{}, nil
	}); err == nil {
		t.Fatal("shutdown accepted worker")
	}
}

func TestDeletionPreparationRejectsSupersededAndOldVisit(t *testing.T) {
	c, _ := newLaunchFocusCoordinator(t)
	c.cardDeleteRead = 2
	c.deletionVisit = 4
	for _, result := range []messages.CardDeletionPreparedMsg{
		{DashboardID: c.activeDashboard, Generation: c.viewGeneration, CardID: "card", Token: "old", Request: 1, Visit: 4},
		{DashboardID: c.activeDashboard, Generation: c.viewGeneration, CardID: "card", Token: "old", Request: 2, Visit: 3},
	} {
		_, cmd := c.Update(result)
		if cmd != nil {
			t.Fatal("stale preparation dispatched a confirmation")
		}
	}
}
