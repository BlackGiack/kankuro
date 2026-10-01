package driver

import "time"

// SetKillGraceForTest shortens how long Kill waits after SIGTERM before it
// escalates to SIGKILL.
func (d *GenericCLIDriver) SetKillGraceForTest(grace time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.killGrace = grace
}
