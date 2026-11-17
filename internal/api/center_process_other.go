//go:build !unix

package api

import (
	"os"
	"os/exec"
)

// Non-Unix platforms fall back to the process termination available in os/exec.
func configureGitProcess(cmd *exec.Cmd) { cmd.Cancel = func() error { return killGitProcess(cmd) } }
func killGitProcess(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return os.ErrProcessDone
	}
	return cmd.Process.Kill()
}
