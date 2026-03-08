package driver

import (
	"errors"
	"fmt"
	"io"
	"kankuro/internal/config"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/creack/pty"
)

type SessionStatus string

const (
	StatusRunning    SessionStatus = "running"
	StatusWaiting    SessionStatus = "waiting"
	StatusTerminated SessionStatus = "terminated"
)

type AgentConfig struct {
	Driver            string
	Model             string
	WorkingDir        string
	SystemInstruction string
	Env               map[string]string
}

type SessionHandle struct {
	Cmd    *exec.Cmd
	PTY    *os.File
	Output <-chan []byte
	Done   <-chan struct{}
}

type AgentDriver interface {
	Spawn(config AgentConfig) (*SessionHandle, error)
	Kill() error

	SetSkills(skills []config.Skill) error
	SetMcps(mcps []config.MCPServer) error

	SendInput(data string) error
	CaptureOutput() (chan []byte, error)

	GetStatus() SessionStatus
}

type GenericCLIDriver struct {
	mu     sync.RWMutex
	status SessionStatus

	cmd    *exec.Cmd
	pty    *os.File
	output chan []byte
	done   chan struct{}
	once   sync.Once
}

// NewGenericCLIDriver returns the default PTY-backed CLI driver implementation.
func NewGenericCLIDriver() *GenericCLIDriver {
	return &GenericCLIDriver{status: StatusTerminated}
}

// Spawn starts an interactive CLI process in a PTY and wires streaming output.
func (d *GenericCLIDriver) Spawn(config AgentConfig) (*SessionHandle, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.cmd != nil && d.status != StatusTerminated {
		return nil, errors.New("session already running")
	}
	if config.Driver == "" {
		return nil, errors.New("driver command is required")
	}
	if config.WorkingDir == "" {
		return nil, errors.New("working directory is required")
	}

	cmd := exec.Command(config.Driver)
	cmd.Dir = config.WorkingDir
	cmd.Env = mergeEnv(config.Env)

	f, err := pty.Start(cmd)
	if err != nil {
		return nil, fmt.Errorf("start PTY process: %w", err)
	}

	output := make(chan []byte, 128)
	done := make(chan struct{})
	d.cmd = cmd
	d.pty = f
	d.output = output
	d.done = done
	d.status = StatusRunning

	go d.streamOutput(f, output)
	go d.watchProcess(cmd)

	return &SessionHandle{Cmd: cmd, PTY: f, Output: output, Done: done}, nil
}

func (d *GenericCLIDriver) streamOutput(f *os.File, out chan []byte) {
	buf := make([]byte, 4096)
	for {
		n, err := f.Read(buf)
		if n > 0 {
			chunk := make([]byte, n)
			copy(chunk, buf[:n])
			out <- chunk
		}
		if err != nil {
			break
		}
	}
	close(out)
}

func (d *GenericCLIDriver) watchProcess(cmd *exec.Cmd) {
	_ = cmd.Wait()
	d.mu.Lock()
	defer d.mu.Unlock()
	d.status = StatusTerminated
	if d.pty != nil {
		_ = d.pty.Close()
		d.pty = nil
	}
	d.cmd = nil
	d.closeDone()
}

// Kill terminates the underlying process and closes its PTY.
func (d *GenericCLIDriver) Kill() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.cmd == nil || d.cmd.Process == nil {
		return nil
	}
	err := d.cmd.Process.Kill()
	d.status = StatusTerminated
	if d.pty != nil {
		_ = d.pty.Close()
		d.pty = nil
	}
	d.closeDone()
	return err
}

// SetSkills is a placeholder hook for provider-specific skill wiring.
func (d *GenericCLIDriver) SetSkills(skills []config.Skill) error {
	_ = skills
	return nil
}

// SetMcps is a placeholder hook for provider-specific MCP wiring.
func (d *GenericCLIDriver) SetMcps(mcps []config.MCPServer) error {
	_ = mcps
	return nil
}

// SendInput writes raw input bytes into the active PTY session.
func (d *GenericCLIDriver) SendInput(data string) error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.pty == nil {
		return errors.New("no active PTY")
	}
	_, err := io.WriteString(d.pty, data)
	return err
}

// CaptureOutput returns the driver's byte-stream output channel.
func (d *GenericCLIDriver) CaptureOutput() (chan []byte, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.output == nil {
		return nil, errors.New("no active output stream")
	}
	return d.output, nil
}

// GetStatus reports the current process lifecycle state.
func (d *GenericCLIDriver) GetStatus() SessionStatus {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.status
}

// PTY exposes the active pseudo-terminal file descriptor.
func (d *GenericCLIDriver) PTY() *os.File {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.pty
}

// PID returns the child process identifier when running.
func (d *GenericCLIDriver) PID() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.cmd == nil || d.cmd.Process == nil {
		return 0
	}
	return d.cmd.Process.Pid
}

func (d *GenericCLIDriver) closeDone() {
	if d.done == nil {
		return
	}
	d.once.Do(func() {
		close(d.done)
	})
}

func mergeEnv(extra map[string]string) []string {
	base := os.Environ()
	for k, v := range extra {
		base = append(base, fmt.Sprintf("%s=%s", k, v))
	}
	return base
}

// SanitizeWorkingDir resolves and validates the session working directory.
func SanitizeWorkingDir(path string) (string, error) {
	if path == "" {
		return "", errors.New("working directory is required")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	st, err := os.Stat(abs)
	if err != nil {
		return "", err
	}
	if !st.IsDir() {
		return "", fmt.Errorf("not a directory: %s", abs)
	}
	return abs, nil
}

// IsPromptLike applies a simple heuristic for "waiting for input" prompts.
func IsPromptLike(last string, lastOutput time.Time, now time.Time) bool {
	if now.Sub(lastOutput) > 15*time.Second {
		return false
	}
	if last == "" {
		return false
	}
	s := last[len(last)-1]
	return s == ':' || s == '?' || s == '>'
}
