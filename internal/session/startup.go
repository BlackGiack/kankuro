package session

import (
	"context"
	"errors"
	"time"
)

// startupDelay is a pragmatic pause, not a CLI readiness guarantee. It also
// adds latency for workspaces that were already trusted.
const startupDelay = 10 * time.Second

func waitForStartup(ctx context.Context, shutdown, done <-chan struct{}, output <-chan []byte) ([]byte, error) {
	return waitForStartupDuration(ctx, shutdown, done, output, startupDelay)
}

func waitForStartupDuration(
	ctx context.Context,
	shutdown, done <-chan struct{},
	output <-chan []byte,
	delay time.Duration,
) ([]byte, error) {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	var scrollback []byte
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-shutdown:
			return nil, ErrManagerClosed
		case <-done:
			return nil, errors.New("agent process exited during startup")
		case chunk, ok := <-output:
			if !ok {
				output = nil
				continue
			}
			scrollback = appendScrollback(scrollback, chunk)
		case <-timer.C:
			// Prefer a pending termination over publication if both became ready.
			select {
			case <-done:
				return nil, errors.New("agent process exited during startup")
			default:
			}
			return scrollback, nil
		}
	}
}

func drainStartupOutput(output <-chan []byte) {
	for range output {
	}
}
