package session

import (
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
