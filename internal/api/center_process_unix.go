//go:build unix

package api

import (
	"errors"
	"os"
	"os/exec"
	"syscall"
)

// configureGitProcess gives each inspection an exclusive process group so
// cancellation cannot leave its Git helpers running or affect other sessions.
func configureGitProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return killGitProcess(cmd) }
}

func killGitProcess(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return os.ErrProcessDone
	}
	err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	if errors.Is(err, syscall.ESRCH) {
		return os.ErrProcessDone
	}
	return err
}
