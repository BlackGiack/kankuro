package driver

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"kankuro/internal/config"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
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

	outputChanSize = 128
	readBufferSize = 4096
	promptTimeout  = 15 * time.Second
	copyTimeout    = 30 * time.Second
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

	SetAgentsMD(card config.IDCard, personas []config.Persona, mcps []config.MCPServer, workspaceRoot string, appendIfExist bool) error
	SetSkills(skills []config.Skill, workspaceRoot ...string) error
	SetPersonas(personas []config.Persona, workspaceRoot ...string) error
	SelectMCPs(mcps []config.MCPServer, workspaceRoot ...string) error
	SetModel(model string) []string

	// SendInput and SendInputBytes both write into the active PTY session,
	// synchronized against Kill via the driver's own lock. This is the only
	// safe way to write attach input: writing directly to a raw *os.File
	// PTY handle (as the session package used to) races Kill, which closes
	// and nils that same fd under lock. See debt_impl_plan_4.md step 4.
	SendInput(data string) error
	SendInputBytes(data []byte) (int, error)

	GetStatus() SessionStatus
	GetAgentFolder() string
}

type GenericCLIDriver struct {
	mu     sync.RWMutex
	status SessionStatus

	cmd      *exec.Cmd
	pty      *os.File
	done     chan struct{}
	once     sync.Once
	exitCode int
}

// NewGenericCLIDriver returns the default PTY-backed CLI driver implementation.
func NewGenericCLIDriver() *GenericCLIDriver {
	return &GenericCLIDriver{status: StatusTerminated}
}

func (d *GenericCLIDriver) GetAgentFolder() string { return ".agents" }

// SetAgentsMD ensures the agent instruction file (AGENTS.md) is present in the session workspace.
func (d *GenericCLIDriver) SetAgentsMD(
	card config.IDCard,
	personas []config.Persona,
	mcps []config.MCPServer,
	workspaceRoot string,
	appendIfExist bool,
) error {
	return syncInstructionFile(workspaceRoot, "AGENTS.md", d.GetAgentFolder(), card, personas, mcps, appendIfExist)
}

// ClaudeCodeDriver specializes the GenericCLIDriver to write CLAUDE.md instead of AGENTS.md.
type ClaudeCodeDriver struct {
	*GenericCLIDriver
}

// NewClaudeCodeDriver returns a specialized driver for Claude Code.
func NewClaudeCodeDriver() *ClaudeCodeDriver {
	return &ClaudeCodeDriver{GenericCLIDriver: NewGenericCLIDriver()}
}

func (d *ClaudeCodeDriver) GetAgentFolder() string { return ".claude" }

// SetAgentsMD ensures the agent instruction file (CLAUDE.md) is present in the session workspace.
func (d *ClaudeCodeDriver) SetAgentsMD(
	card config.IDCard,
	personas []config.Persona,
	mcps []config.MCPServer,
	workspaceRoot string,
	appendIfExist bool,
) error {
	return syncInstructionFile(workspaceRoot, "CLAUDE.md", d.GetAgentFolder(), card, personas, mcps, appendIfExist)
}

func (d *ClaudeCodeDriver) SetSkills(skills []config.Skill, workspaceRoot ...string) error {
	return setSkills(skills, d.GetAgentFolder(), workspaceRoot...)
}

func (d *ClaudeCodeDriver) SetPersonas(personas []config.Persona, workspaceRoot ...string) error {
	return setPersonas(personas, d.GetAgentFolder(), workspaceRoot...)
}

func syncInstructionFile(
	finalWD, filename, agentFolder string,
	card config.IDCard,
	personas []config.Persona,
	mcps []config.MCPServer,
	appendIfExist bool,
) error {
	content, err := resolveAgentsMDContent(card)
	if err != nil {
		return err
	}

	if len(personas) == 1 {
		content = appendPersonaInstruction(content, agentFolder, personas[0])
	}

	if len(mcps) > 0 {
		content = appendMCPInstruction(content, mcps)
	}

	if content == "" {
		return nil
	}

	return writeInstructionFile(finalWD, filename, content, appendIfExist)
}

func resolveAgentsMDContent(card config.IDCard) (string, error) {
	if card.AgentsMDPreferPath && card.AgentsMDPath != "" {
		data, err := os.ReadFile(card.AgentsMDPath)
		if err != nil {
			return "", fmt.Errorf("read agents md path %s: %w", card.AgentsMDPath, err)
		}
		return string(data), nil
	}
	return card.AgentsMD, nil
}

func appendPersonaInstruction(content, agentFolder string, p config.Persona) string {
	ext := filepath.Ext(p.Source)
	relPath := filepath.Join(agentFolder, "agents", p.Name+ext)
	instruction := fmt.Sprintf("Read the file %s and embody the persona described during this session,", relPath)
	if content != "" {
		return content + "\n\n" + instruction
	}
	return instruction
}

func appendMCPInstruction(content string, mcps []config.MCPServer) string {
	mcpNames := make([]string, 0, len(mcps))
	for _, mcp := range mcps {
		mcpNames = append(mcpNames, mcp.Name)
	}
	mcpJSON, _ := json.Marshal(mcpNames)
	mcpLine := fmt.Sprintf("Available mcp: %s", string(mcpJSON))
	if content != "" {
		return content + "\n" + mcpLine
	}
	return mcpLine
}

func writeInstructionFile(dir, filename, content string, appendIfExist bool) error {
	path := filepath.Join(dir, filename)
	if !appendIfExist {
		// G703: Path traversal via taint analysis.
		return os.WriteFile(path, []byte(content), 0o600)
	}

	// G703: Path traversal via taint analysis.
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open %s: %w", filename, err)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return fmt.Errorf("stat %s: %w", filename, err)
	}

	if stat.Size() > 0 {
		content = "\n\n" + content
	}

	_, err = f.WriteString(content)
	return err
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

	// G204: Subprocess launched with a potential tainted input or cmd arguments.
	// This is intentional as the driver is designed to run user-defined commands.
	// We use CommandContext with Background for the main process as its lifecycle
	// is managed via Kill() and watchProcess.
	//nolint:gosec // user-defined command execution
	cmd := exec.CommandContext(context.Background(), config.Driver, config.Args...)
	cmd.Dir = config.WorkingDir
	cmd.Env = mergeEnvWithPwd(config.Env, config.WorkingDir)

	f, err := pty.Start(cmd)
	if err != nil {
		return nil, fmt.Errorf("start PTY process: %w", err)
	}

	output := make(chan []byte, outputChanSize)
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
	buf := make([]byte, readBufferSize)
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
	if cmd.ProcessState != nil {
		d.exitCode = cmd.ProcessState.ExitCode()
	}
	if d.pty != nil {
		_ = d.pty.Close()
		d.pty = nil
	}
	d.cmd = nil
	d.closeDone()
}

// Kill terminates the underlying process and closes its PTY. A process that
// has already exited on its own is treated as a successful kill.
func (d *GenericCLIDriver) Kill() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.cmd == nil || d.cmd.Process == nil {
		return nil
	}
	err := d.cmd.Process.Kill()
	if errors.Is(err, os.ErrProcessDone) {
		err = nil
	}
	d.status = StatusTerminated
	if d.pty != nil {
		_ = d.pty.Close()
		d.pty = nil
	}
	d.closeDone()
	return err
}

// SetSkills ensures skill assets are present in the session workspace.
// For built-in skills (empty Source), content is written directly from embedded bytes.
func (d *GenericCLIDriver) SetSkills(skills []config.Skill, workspaceRoot ...string) error {
	return setSkills(skills, d.GetAgentFolder(), workspaceRoot...)
}

func setSkills(skills []config.Skill, agentsFolder string, workspaceRoot ...string) error {
	if len(skills) == 0 {
		return nil
	}
	root, errResolve := resolveWorkspaceRoot(workspaceRoot)
	if errResolve != nil {
		return errResolve
	}
	skillsRoot := filepath.Join(root, agentsFolder, "skills")
	if errMkdir := os.MkdirAll(skillsRoot, 0o700); errMkdir != nil {
		return fmt.Errorf("create skills dir: %w", errMkdir)
	}

	nameCounts := make(map[string]int)
	for _, skill := range skills {
		nameCounts[skill.Name]++
	}

	for _, skill := range skills {
		if errDeploy := deploySkill(skill, skillsRoot, nameCounts); errDeploy != nil {
			return errDeploy
		}
	}
	return nil
}

func deploySkill(skill config.Skill, skillsRoot string, nameCounts map[string]int) error {
	if skill.Name == "" {
		return errors.New("skill name is required")
	}
	destName := skill.Name
	if nameCounts[skill.Name] > 1 {
		destName = fmt.Sprintf("%s--%s", skill.Name, shortSkillHash(skill.Source))
	}
	destPath := filepath.Join(skillsRoot, destName)

	info, errStat := os.Stat(destPath)
	if errStat == nil {
		if !info.IsDir() {
			return fmt.Errorf("skill destination exists and is not a directory: %s", destPath)
		}
		return nil
	}
	if !errors.Is(errStat, os.ErrNotExist) {
		return fmt.Errorf("stat skill destination: %w", errStat)
	}

	// Handle built-in skills: write content directly from embedded bytes
	if skill.Builtin {
		return writeBuiltinSkill(skill.Name, destPath)
	}

	// Regular skills: copy from filesystem source directory
	if skill.Source == "" {
		return fmt.Errorf("skill %q has no source directory", skill.Name)
	}
	return copySkillFolder(skill.Source, destPath)
}

func resolveWorkspaceRoot(workspaceRoot []string) (string, error) {
	if len(workspaceRoot) == 0 {
		return "", errors.New("workspace root is required for skill sync")
	}
	root := strings.TrimSpace(workspaceRoot[0])
	if root == "" {
		return "", errors.New("workspace root is required for skill sync")
	}
	return root, nil
}

// SetPersonas ensures persona assets are present in the session workspace.
func (d *GenericCLIDriver) SetPersonas(personas []config.Persona, workspaceRoot ...string) error {
	return setPersonas(personas, d.GetAgentFolder(), workspaceRoot...)
}

func setPersonas(personas []config.Persona, agentsFolder string, workspaceRoot ...string) error {
	if len(personas) == 0 {
		return nil
	}
	root := ""
	if len(workspaceRoot) > 0 {
		root = strings.TrimSpace(workspaceRoot[0])
	}
	if root == "" {
		return errors.New("workspace root is required for persona sync")
	}
	personasRoot := filepath.Join(root, agentsFolder, "agents")
	if err := os.MkdirAll(personasRoot, 0o700); err != nil {
		return fmt.Errorf("create personas dir: %w", err)
	}

	nameCounts := make(map[string]int)
	for _, persona := range personas {
		nameCounts[persona.Name]++
	}

	for _, persona := range personas {
		if persona.Source == "" || persona.Name == "" {
			return errors.New("persona name and source are required")
		}
		ext := filepath.Ext(persona.Source)
		destName := persona.Name + ext
		if nameCounts[persona.Name] > 1 {
			destName = fmt.Sprintf("%s--%s%s", persona.Name, shortSkillHash(persona.Source), ext)
		}
		destPath := filepath.Join(personasRoot, destName)
		_, errStat := os.Stat(destPath)
		if errStat == nil {
			continue
		}
		if !errors.Is(errStat, os.ErrNotExist) {
			return fmt.Errorf("stat persona destination: %w", errStat)
		}
		if errCopy := copyPersonaFile(persona.Source, destPath); errCopy != nil {
			return errCopy
		}
	}
	return nil
}

// SelectMCPs ensures MCP configuration is present in the session workspace.
func (d *GenericCLIDriver) SelectMCPs(mcps []config.MCPServer, workspaceRoot ...string) error {
	if len(mcps) == 0 {
		return nil
	}
	root := ""
	if len(workspaceRoot) > 0 {
		root = strings.TrimSpace(workspaceRoot[0])
	}
	if root == "" {
		return errors.New("workspace root is required for MCP sync")
	}
	kankuroDir := filepath.Join(root, ".kankuro")
	if err := os.MkdirAll(kankuroDir, 0o700); err != nil {
		return fmt.Errorf("create .kankuro dir: %w", err)
	}

	type mcpEntry struct {
		Command string            `json:"command,omitempty"`
		Args    []string          `json:"args,omitempty"`
		Env     map[string]string `json:"env,omitempty"`
		CWD     string            `json:"cwd,omitempty"`
		URL     string            `json:"url,omitempty"`
		Headers map[string]string `json:"headers,omitempty"`
	}
	type mcpConfig struct {
		MCPServers map[string]mcpEntry `json:"mcpServers"`
	}

	cfg := mcpConfig{MCPServers: make(map[string]mcpEntry)}
	for _, mcp := range mcps {
		entry := mcpEntry{
			Args:    mcp.Args,
			Env:     mcp.Env,
			CWD:     mcp.CWD,
			Headers: mcp.Headers,
		}
		if mcp.Command != "" {
			entry.Command = mcp.Command
		}
		if mcp.URL != "" {
			entry.URL = mcp.URL
		}
		cfg.MCPServers[mcp.Name] = entry
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal mcp config: %w", err)
	}

	if err = os.WriteFile(filepath.Join(kankuroDir, "mcp-servers.json"), data, 0o600); err != nil {
		return fmt.Errorf("write mcp-servers.json: %w", err)
	}
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
	_, err := d.SendInputBytes([]byte(data))
	return err
}

// SendInputBytes writes raw input bytes into the active PTY session. It
// holds only a read lock: it doesn't mutate driver state, but Kill (which
// closes and nils the pty under a write lock) must never race a write to an
// already-closed fd - the RLock/Lock pairing is what serializes the two.
func (d *GenericCLIDriver) SendInputBytes(data []byte) (int, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.pty == nil {
		return 0, errors.New("no active PTY")
	}
	return d.pty.Write(data)
}

// GetStatus reports the current process lifecycle state.
func (d *GenericCLIDriver) GetStatus() SessionStatus {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.status
}

// ExitCode returns the child process's exit code once it has terminated.
// It is -1 for a signalled process and 0 before the process has exited.
func (d *GenericCLIDriver) ExitCode() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.exitCode
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
	keys := make([]string, 0, len(extra))
	for k := range extra {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		base = upsertEnv(base, k, extra[k])
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

func writeBuiltinSkill(name, destPath string) error {
	content, ok := config.GetBuiltinSkillContent(name)
	if !ok {
		return fmt.Errorf("built-in skill content not found: %s", name)
	}
	if err := os.MkdirAll(destPath, 0o700); err != nil {
		return fmt.Errorf("create skill dir: %w", err)
	}
	skillFile := filepath.Join(destPath, "SKILL.md")
	return os.WriteFile(skillFile, content, 0o600)
}

func copySkillFolder(source, dest string) error {
	info, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("stat skill source: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("skill source is not a directory: %s", source)
	}

	ctx, cancel := context.WithTimeout(context.Background(), copyTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "cp", "-r", source, dest)
	if out, errOutput := cmd.CombinedOutput(); errOutput != nil {
		return fmt.Errorf("copy skill %s: %s: %w", source, strings.TrimSpace(string(out)), errOutput)
	}
	return nil
}

func copyPersonaFile(source, dest string) error {
	info, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("stat persona source: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("persona source is a directory: %s", source)
	}

	ctx, cancel := context.WithTimeout(context.Background(), copyTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "cp", source, dest)
	if out, errOutput := cmd.CombinedOutput(); errOutput != nil {
		return fmt.Errorf("copy persona %s: %s: %w", source, strings.TrimSpace(string(out)), errOutput)
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
	st, errStat := os.Stat(abs)
	if errStat != nil {
		return "", errStat
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
	return now.Sub(lastWordAt) > promptTimeout
}
