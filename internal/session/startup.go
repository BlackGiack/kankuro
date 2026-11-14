package session

import (
	"errors"
	"fmt"
	"time"

	"github.com/BlackGiack/kankuro/internal/driver"
)

// startupDelay gates MCP message admission; it does not establish CLI readiness.
const startupDelay = 10 * time.Second

// teardownUnpublished drains while Kill runs, then joins the drain even if the
// driver does not close its output channel after teardown.
func teardownUnpublished(d driver.AgentDriver, output <-chan []byte) error {
	stop, drained := make(chan struct{}), make(chan struct{})
	go func() {
		defer close(drained)
		for {
			select {
			case _, ok := <-output:
				if !ok {
					return
				}
			case <-stop:
				return
			}
		}
	}()
	err := d.Kill()
	close(stop)
	<-drained
	return err
}

// teardownActivation records whether the driver observed termination. A kill
// error is retained even when termination is confirmed.
func teardownActivation(
	d driver.AgentDriver,
	h *driver.SessionHandle,
	result *ActivationResult,
	sessionID, workingDir string,
	cause error,
) error {
	pid := 0
	if h.Cmd != nil && h.Cmd.Process != nil {
		pid = h.Cmd.Process.Pid
	}
	killErr := teardownUnpublished(d, h.Output)
	*result = ActivationResult{SessionID: sessionID, PID: pid, WorkingDir: workingDir, Spawned: true}
	if d.GetStatus() != driver.StatusTerminated {
		result.Effect = ActivationUncertain
		return errors.Join(
			cause,
			killErr,
			fmt.Errorf("spawned process termination was not confirmed; inspect workspace %q and process %d", workingDir, pid),
		)
	}
	return errors.Join(cause, killErr, errors.New("spawned process terminated before publication"))
}
