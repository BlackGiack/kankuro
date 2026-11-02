package layout

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

// Lock holds the config-root writer lock until Close. The lock file is never
// removed: unlinking it while another process waits would permit split locks.
type Lock struct{ file *os.File }

// AcquireLock obtains a nonblocking exclusive lock. It must eventually be used
// by both application startup and migration before either accesses the layout.
func AcquireLock(dir string) (*Lock, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("create config directory: %w", err)
	}
	file, err := os.OpenFile(filepath.Join(dir, ".dashboard-layout.lock"), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open layout lock: %w", err)
	}
	if errAt25 := unix.Flock(int(file.Fd()), unix.LOCK_EX|unix.LOCK_NB); errAt25 != nil {
		_ = file.Close()
		return nil, fmt.Errorf("acquire layout lock (stop other Kankuro or migration processes): %w", errAt25)
	}
	return &Lock{file: file}, nil
}

// Close releases the lock and its file descriptor.
func (l *Lock) Close() error {
	if l.file == nil {
		return nil
	}
	err := l.file.Close()
	l.file = nil
	return err
}
