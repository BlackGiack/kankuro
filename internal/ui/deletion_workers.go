package ui

import (
	"context"
	"sync"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
)

type deletionWorkers struct {
	mu      sync.Mutex
	closed  bool
	cancels map[context.Context]context.CancelFunc
	wg      sync.WaitGroup
}

func (w *deletionWorkers) run(fn func(context.Context) (contracts.DeletionProgress, error)) (contracts.DeletionProgress, error) {
	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()
		return contracts.DeletionProgress{}, context.Canceled
	}
	ctx, cancel := context.WithTimeout(context.Background(), dashboardDeletionTimeout)
	if w.cancels == nil {
		w.cancels = make(map[context.Context]context.CancelFunc)
	}
	w.cancels[ctx] = cancel
	w.wg.Add(1)
	w.mu.Unlock()
	defer func() { cancel(); w.mu.Lock(); delete(w.cancels, ctx); w.mu.Unlock(); w.wg.Done() }()
	return fn(ctx)
}
func (w *deletionWorkers) close() {
	w.mu.Lock()
	w.closed = true
	for _, cancel := range w.cancels {
		cancel()
	}
	w.mu.Unlock()
	w.wg.Wait()
}
