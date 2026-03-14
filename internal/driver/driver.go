package driver

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"kankuro/internal/config"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	Args              []string
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

	SetSkills(skills []config.Skill, workspaceRoot ...string) error
	SetMcps(mcps []config.MCPServer) error
	SetModel(model string) []string

	SendInput(data string) error

	GetStatus() SessionStatus
}

type GenericCLIDriver struct {
	mu     sync.RWMutex
	status SessionStatus

	cmd  *exec.Cmd
	pty  *os.File
	done chan struct{}
	once sync.Once
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

	cmd := exec.Command(config.Driver, config.Args...)
	cmd.Dir = config.WorkingDir
	cmd.Env = mergeEnvWithPwd(config.Env, config.WorkingDir)

	f, err := pty.Start(cmd)
	if err != nil {
		return nil, fmt.Errorf("start PTY process: %w", err)
	}

	output := make(chan []byte, 128)
	done := make(chan struct{})
	d.cmd = cmd
	d.pty = f
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

// SetSkills ensures skill assets are present in the session workspace.
func (d *GenericCLIDriver) SetSkills(skills []config.Skill, workspaceRoot ...string) error {
	if len(skills) == 0 {
		return nil
	}
	root := ""
	if len(workspaceRoot) > 0 {
		root = strings.TrimSpace(workspaceRoot[0])
	}
	if root == "" {
		return errors.New("workspace root is required for skill sync")
	}
	skillsRoot := filepath.Join(root, ".agents", "skills")
	if err := os.MkdirAll(skillsRoot, 0o755); err != nil {
		return fmt.Errorf("create skills dir: %w", err)
	}

	nameCounts := make(map[string]int)
	for _, skill := range skills {
		nameCounts[skill.Name]++
	}

	for _, skill := range skills {
		if skill.Source == "" || skill.Name == "" {
			return errors.New("skill name and source are required")
		}
		destName := skill.Name
		if nameCounts[skill.Name] > 1 {
			destName = fmt.Sprintf("%s--%s", skill.Name, shortSkillHash(skill.Source))
		}
		destPath := filepath.Join(skillsRoot, destName)
		info, err := os.Stat(destPath)
		if err == nil {
			if !info.IsDir() {
				return fmt.Errorf("skill destination exists and is not a directory: %s", destPath)
			}
			continue
		}
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("stat skill destination: %w", err)
		}
		if err := copySkillFolder(skill.Source, destPath); err != nil {
			return err
		}
	}
	return nil
}

// SetMcps is a placeholder hook for provider-specific MCP wiring.
func (d *GenericCLIDriver) SetMcps(mcps []config.MCPServer) error {
	_ = mcps
	return nil
}

// SetModel returns the command line arguments to configure the model.
func (d *GenericCLIDriver) SetModel(model string) []string {
	if model == "" {
		return nil
	}
	return []string{"--model", model}
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

func mergeEnvWithPwd(extra map[string]string, pwd string) []string {
	base := os.Environ()
	for k, v := range extra {
		base = append(base, fmt.Sprintf("%s=%s", k, v))
	}
	return upsertEnv(base, "PWD", pwd)
}

func upsertEnv(env []string, key string, value string) []string {
	if key == "" {
		return env
	}
	prefix := key + "="
	for i, entry := range env {
		if strings.HasPrefix(entry, prefix) {
			env[i] = prefix + value
			return env
		}
	}
	return append(env, prefix+value)
}

func shortSkillHash(path string) string {
	sum := sha256.Sum256([]byte(path))
	return hex.EncodeToString(sum[:])[:8]
}

func copySkillFolder(source, dest string) error {
	info, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("stat skill source: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("skill source is not a directory: %s", source)
	}
	cmd := exec.Command("cp", "-r", source, dest)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("copy skill %s: %s: %w", source, strings.TrimSpace(string(out)), err)
	}
	return nil
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

// IsPromptLike applies a heuristic for "waiting for input" prompts.
// It returns true if the last line ends with a prompt-like character,
// or if no "words" (alphanumeric characters) have been received for more than 15 seconds.
func IsPromptLike(last string, lastWordAt time.Time, now time.Time) bool {
	if last != "" {
		s := last[len(last)-1]
		if s == ':' || s == '?' || s == '>' {
			return true
		}
	}
	return now.Sub(lastWordAt) > 15*time.Second
}
