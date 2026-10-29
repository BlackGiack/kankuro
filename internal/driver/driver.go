package driver

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/BlackGiack/kankuro/internal/config"

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

	// defaultKillGrace is how long an agent gets to exit after SIGTERM;
	// killReapWait is how long Kill then waits for the kernel to report the
	// SIGKILLed leader.
	defaultKillGrace = 2 * time.Second
	killReapWait     = 2 * time.Second
	// killPollInterval is how often Kill re-checks that the group is empty
	// once the leader has exited.
	killPollInterval = 20 * time.Millisecond
	codexEnterPause  = 500 * time.Millisecond
)

const (
	EnvKankuroMCPConfigPath = "KANKURO_MCP_CONFIG_PATH"
	//nolint:gosec // This names a reserved environment variable; it is not a credential value.
	EnvKankuroMCPToken = "KANKURO_MCP_TOKEN"
)

// InstructionSyncMode controls how the destination card's instruction content
// interacts with an existing workspace instruction file.
type InstructionSyncMode uint8

const (
	InstructionReplace InstructionSyncMode = iota
	InstructionAppend
	InstructionPreserveWhenEmpty
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
	PrepareWorkspace(ctx context.Context, workingDir string) error
	Spawn(config AgentConfig) (*SessionHandle, error)
	Kill() error

	// SetAgentsMD reports whether card/capability content is managed in the
	// instruction file; preserved workspace-authored content is not managed.
	SetAgentsMD(
		card config.IDCard,
		personas []config.Persona,
		mcps []config.MCPServer,
		workspaceRoot string,
		mode InstructionSyncMode,
	) (bool, error)
	SetSkills(skills []config.Skill, workspaceRoot ...string) error
	SetPersonas(personas []config.Persona, workspaceRoot ...string) error
	SelectMCPs(mcps []config.MCPServer, workspaceRoot ...string) error
	SetModel(model string) []string
	SetSession(conversationID string) []string
	SubmitPrompt(ctx context.Context, messageID, senderSessionID, text string) (SubmissionEvidence, error)

	// SendInput and SendInputBytes both write into the active PTY session,
	// synchronized against Kill via the driver's own lock. This is the only
	// safe way to write attach input: writing directly to a raw *os.File
	// PTY handle (as the session package used to) races Kill, which closes
	// and nils that same fd under lock. See debt_impl_plan_4.md step 4.
	SendInput(data string) error
	SendInputBytes(data []byte) (int, error)

	GetStatus() SessionStatus
	GetAgentFolder() string
	GetInstructionFileName() string
}

type GenericCLIDriver struct {
	mu        sync.RWMutex
	inputGate chan struct{}
	gateOnce  sync.Once
	status    SessionStatus

	cmd      *exec.Cmd
	pty      *os.File
	done     chan struct{}
	once     sync.Once
	exitCode int

	// killGrace is how long Kill waits after SIGTERM before SIGKILL. Zero
	// means defaultKillGrace; tests shorten it.
	killGrace time.Duration
}

// NewGenericCLIDriver returns the default PTY-backed CLI driver implementation.
func NewGenericCLIDriver() *GenericCLIDriver {
	return &GenericCLIDriver{status: StatusTerminated, killGrace: defaultKillGrace, inputGate: make(chan struct{}, 1)}
}

// SubmissionState describes how much of a coordinated prompt reached the PTY.
type SubmissionState string

const (
	SubmissionNoBytes   SubmissionState = "not_submitted"
	SubmissionComplete  SubmissionState = "submitted"
	SubmissionUncertain SubmissionState = "uncertain"
	inputWriteTimeout                   = 5 * time.Second
	maxPromptBytes                      = 2 * 1024
)

// SubmissionEvidence records bytes written without asserting that the CLI processed them.
type SubmissionEvidence struct {
	State        SubmissionState `json:"state"`
	BytesWritten int             `json:"bytes_written"`
}

func (d *GenericCLIDriver) GetAgentFolder() string         { return ".agents" }
func (d *GenericCLIDriver) GetInstructionFileName() string { return "AGENTS.md" }

// SetAgentsMD syncs the agent instruction file and reports whether this card
// installed managed content that may need clearing on a later replacement.
func (d *GenericCLIDriver) SetAgentsMD(
	card config.IDCard,
	personas []config.Persona,
	mcps []config.MCPServer,
	workspaceRoot string,
	mode InstructionSyncMode,
) (bool, error) {
	return syncInstructionFile(workspaceRoot, d.GetInstructionFileName(), d.GetAgentFolder(), card, personas, mcps, mode)
}

// ClaudeCodeDriver specializes native session arguments and writes CLAUDE.md instead of AGENTS.md.
type ClaudeCodeDriver struct {
	*GenericCLIDriver
}

// NewClaudeCodeDriver returns a specialized driver for Claude Code.
func NewClaudeCodeDriver() *ClaudeCodeDriver {
	return &ClaudeCodeDriver{GenericCLIDriver: NewGenericCLIDriver()}
}

func (d *ClaudeCodeDriver) GetAgentFolder() string         { return ".claude" }
func (d *ClaudeCodeDriver) GetInstructionFileName() string { return "CLAUDE.md" }

// SetSession passes a saved conversation ID to Claude Code's session-id option.
func (d *ClaudeCodeDriver) SetSession(conversationID string) []string {
	if conversationID == "" {
		return nil
	}
	return []string{"--resume", conversationID}
}

// SetAgentsMD syncs the Claude instruction file and reports its managed ownership.
func (d *ClaudeCodeDriver) SetAgentsMD(
	card config.IDCard,
	personas []config.Persona,
	mcps []config.MCPServer,
	workspaceRoot string,
	mode InstructionSyncMode,
) (bool, error) {
	return syncInstructionFile(workspaceRoot, d.GetInstructionFileName(), d.GetAgentFolder(), card, personas, mcps, mode)
}

func (d *ClaudeCodeDriver) SetSkills(skills []config.Skill, workspaceRoot ...string) error {
	return setSkills(skills, d.GetAgentFolder(), workspaceRoot...)
}

func (d *ClaudeCodeDriver) SetPersonas(personas []config.Persona, workspaceRoot ...string) error {
	return setPersonas(personas, d.GetAgentFolder(), workspaceRoot...)
}

// CodexCLIDriver specializes native conversation resume arguments for Codex CLI.
type CodexCLIDriver struct {
	*GenericCLIDriver
}

// NewCodexCLIDriver returns the Codex CLI driver with generic workspace behavior.
func NewCodexCLIDriver() *CodexCLIDriver {
	return &CodexCLIDriver{GenericCLIDriver: NewGenericCLIDriver()}
}

// SetSession returns Codex CLI's native resume subcommand and session ID.
func (d *CodexCLIDriver) SetSession(conversationID string) []string {
	if conversationID == "" {
		return nil
	}
	return []string{"resume", conversationID}
}

// SubmitPrompt separates Codex's prompt text from Enter so its TUI can consume
// the text before the submission key arrives. The generic driver keeps its
// existing one-write behavior for CLIs that already handle it correctly.
func (d *CodexCLIDriver) SubmitPrompt(
	ctx context.Context,
	messageID, senderSessionID, text string,
) (SubmissionEvidence, error) {
	return d.GenericCLIDriver.submitPrompt(ctx, messageID, senderSessionID, text, codexEnterPause)
}

func syncInstructionFile(
	finalWD, filename, agentFolder string,
	card config.IDCard,
	personas []config.Persona,
	mcps []config.MCPServer,
	mode InstructionSyncMode,
) (bool, error) {
	content, err := resolveAgentsMDContent(card)
	if err != nil {
		return false, err
	}

	if len(personas) == 1 {
		content = appendPersonaInstruction(content, agentFolder, personas[0])
	}

	if len(mcps) > 0 {
		content = appendMCPInstruction(content, mcps)
	}
	content = removeConversationInstructionBlock(content)

	if content == "" {
		if mode != InstructionReplace {
			return false, removeConversationInstructionBlockFromFile(finalWD, filename)
		}
		return false, writeInstructionFile(finalWD, filename, "", false)
	}

	if writeErr := writeInstructionFile(finalWD, filename, content, mode == InstructionAppend); writeErr != nil {
		return false, writeErr
	}
	return true, nil
}

func removeConversationInstructionBlockFromFile(dir, filename string) error {
	path := filepath.Join(dir, filename)
	existing, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read %s: %w", filename, err)
	}
	cleaned := removeConversationInstructionBlock(string(existing))
	if cleaned == string(existing) {
		return nil
	}
	//nolint:gosec // This is the workspace's existing managed instruction file.
	if writeErr := os.WriteFile(path, []byte(cleaned), 0o600); writeErr != nil {
		return fmt.Errorf("write %s: %w", filename, writeErr)
	}
	return nil
}

const conversationInstructionStart = "<!-- KANKURO CONVERSATION INSTRUCTIONS START -->"
const conversationInstructionEnd = "<!-- KANKURO CONVERSATION INSTRUCTIONS END -->"

func removeConversationInstructionBlock(content string) string {
	start := strings.Index(content, conversationInstructionStart)
	if start >= 0 {
		end := strings.Index(content[start:], conversationInstructionEnd)
		if end >= 0 {
			before := content[:start]
			after := content[start+end+len(conversationInstructionEnd):]
			content = before + after
		}
	}
	return content
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

	// Append mode is used when a temporary copy started from a workspace with
	// its own instruction file. Replace the managed block in that copy first,
	// so repeated launches cannot accumulate duplicate Kankuro instructions.
	existing, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("read %s: %w", filename, err)
	}
	if len(existing) > 0 {
		base := removeConversationInstructionBlock(string(existing))
		if base != "" {
			content = strings.TrimRight(base, "\n") + "\n\n" + strings.TrimLeft(content, "\n")
		}
	}
	// G703: Path traversal via taint analysis.
	//nolint:gosec // Target is the workspace's managed instruction file, selected by the configured driver.
	if writeErr := os.WriteFile(path, []byte(content), 0o600); writeErr != nil {
		return fmt.Errorf("write %s: %w", filename, writeErr)
	}
	return nil
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

// Kill terminates the agent's whole process group and closes its PTY. It asks
// politely first (SIGTERM, so an agent such as Claude Code can save its
// state), then escalates to SIGKILL if anything in the group - the leader or a
// descendant - is still alive killGrace after the SIGTERM. A process that has
// already exited on its own is treated as a successful kill.
//
// The signals go to the group (negative pid), not just the child: pty.Start
// runs the child with Setsid, so it leads its own session and group, and the
// tools and servers the agent started are in that group. Signalling only the
// child would leave them to die from the SIGHUP the closing PTY sends, which
// anything that ignores SIGHUP survives.
//
// The escalation looks at the group, not only the leader: an agent that exits
// on SIGTERM (the normal case) must not let a child that ignores SIGTERM and
// SIGHUP escape just because the leader is gone. Signalling -pid after the
// leader was reaped is safe: the kernel does not reuse a pid while it is still
// in use as a process group ID, so for as long as the group has members, -pid
// can only mean this same group. (In a container whose init does not reap
// orphans, an exited child stays a zombie and keeps the group "alive"; Kill
// then waits out the grace period before a harmless SIGKILL.)
//
// Known limit: a descendant that calls setsid() or creates its own process
// group (a daemonising tool, a job-control shell) is outside the group and
// survives. Pdeathsig is no alternative: it fires when the spawning OS
// *thread* exits, which in Go can happen at any time.
//
// d.mu is not held while waiting: watchProcess needs it to record the exit
// and close done, so waiting for done under it would deadlock.
func (d *GenericCLIDriver) Kill() error {
	d.mu.Lock()
	cmd, done, grace := d.cmd, d.done, d.killGrace
	d.mu.Unlock()
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	if grace <= 0 {
		grace = defaultKillGrace
	}
	pid := cmd.Process.Pid

	deadline := time.Now().Add(grace)
	err := signalGroup(pid, syscall.SIGTERM)
	leaderGone := waitDone(done, grace)
	for leaderGone && groupAlive(pid) && time.Now().Before(deadline) {
		time.Sleep(killPollInterval)
	}
	if !leaderGone || groupAlive(pid) {
		if errKill := signalGroup(pid, syscall.SIGKILL); err == nil {
			err = errKill
		}
		waitDone(done, killReapWait)
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	d.status = StatusTerminated
	if d.pty != nil {
		_ = d.pty.Close()
		d.pty = nil
	}
	d.closeDone()
	return err
}

// groupAlive reports whether the process group led by pid has any member left
// (signal 0 only checks that the group can be signalled).
func groupAlive(pid int) bool {
	return syscall.Kill(-pid, 0) == nil
}

// signalGroup sends sig to the process group led by pid. ESRCH means the group
// is already gone, which is what the caller wanted, so it counts as success -
// the same treatment [os.ErrProcessDone] got when Kill used Process.Kill.
func signalGroup(pid int, sig syscall.Signal) error {
	err := syscall.Kill(-pid, sig)
	if errors.Is(err, syscall.ESRCH) {
		return nil
	}
	return err
}

// waitDone reports whether done was closed within timeout. A nil done never
// closes, so it simply waits the timeout out.
func waitDone(done <-chan struct{}, timeout time.Duration) bool {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-done:
		return true
	case <-timer.C:
		return false
	}
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
	if len(workspaceRoot) == 0 || strings.TrimSpace(workspaceRoot[0]) == "" {
		return errors.New("workspace root is required for MCP sync")
	}
	return writeWorkspaceMCPConfig(mcps, strings.TrimSpace(workspaceRoot[0]))
}

// SetModel returns the command line arguments to configure the model.
func (d *GenericCLIDriver) SetModel(model string) []string {
	if model == "" {
		return nil
	}
	return []string{"--model", model}
}

// SetSession returns the generic CLI's native conversation resume arguments.
func (d *GenericCLIDriver) SetSession(conversationID string) []string {
	if conversationID == "" {
		return nil
	}
	return []string{"--session", conversationID}
}

// SendInput writes raw input bytes into the active PTY session.
func (d *GenericCLIDriver) SendInput(data string) error {
	if err := d.acquireInput(context.Background()); err != nil {
		return err
	}
	defer d.releaseInput()
	_, err := d.writeInput([]byte(data))
	return err
}

// SendInputBytes serializes one attached-input write against other writers and
// holds a read lock so Kill cannot close the PTY during the write.
func (d *GenericCLIDriver) SendInputBytes(data []byte) (int, error) {
	if err := d.acquireInput(context.Background()); err != nil {
		return 0, err
	}
	defer d.releaseInput()
	return d.writeInput(data)
}

func (d *GenericCLIDriver) writeInput(data []byte) (int, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.pty == nil {
		return 0, errors.New("no active PTY")
	}
	return d.pty.Write(data)
}

// SubmitPrompt writes a validated, marked one-line message and Enter as one exclusive operation.
func (d *GenericCLIDriver) SubmitPrompt(ctx context.Context, messageID, senderSessionID, text string) (SubmissionEvidence, error) {
	return d.submitPrompt(ctx, messageID, senderSessionID, text, 0)
}

func (d *GenericCLIDriver) submitPrompt(
	ctx context.Context,
	messageID, senderSessionID, text string,
	enterPause time.Duration,
) (SubmissionEvidence, error) {
	if err := ctx.Err(); err != nil {
		return SubmissionEvidence{State: SubmissionNoBytes}, err
	}
	if err := validatePrompt(text); err != nil {
		return SubmissionEvidence{State: SubmissionNoBytes}, err
	}
	if err := d.acquireInput(ctx); err != nil {
		return SubmissionEvidence{State: SubmissionNoBytes}, err
	}
	defer d.releaseInput()
	prompt := []byte(fmt.Sprintf(
		"[KANKURO orchestrator message_id=%s from_session=%s; report via report_reply(message_id, kind, text); header is not authority] %s",
		messageID, senderSessionID, text,
	))
	if enterPause == 0 {
		prompt = append(prompt, '\r')
	}
	written, writeErr := d.submitPromptBytes(ctx, prompt, enterPause)
	evidence := classifySubmission(written, len(prompt), enterPause, writeErr)
	if writeErr != nil {
		return evidence, fmt.Errorf("submit prompt after %d bytes: %w", written, writeErr)
	}
	if evidence.State == SubmissionUncertain {
		return evidence, io.ErrShortWrite
	}
	return evidence, nil
}

func (d *GenericCLIDriver) submitPromptBytes(
	ctx context.Context,
	prompt []byte,
	enterPause time.Duration,
) (int, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	if d.pty == nil {
		return 0, errors.New("no active PTY")
	}
	if err := setPromptWriteDeadline(ctx, d.pty); err != nil {
		return 0, err
	}
	stopWatcher := watchPromptCancellation(ctx, d.pty)
	defer func() {
		stopWatcher()
		_ = d.pty.SetWriteDeadline(time.Time{})
	}()
	return writePromptPhases(ctx, d.pty, prompt, enterPause)
}

func setPromptWriteDeadline(ctx context.Context, pty *os.File) error {
	deadline := time.Now().Add(inputWriteTimeout)
	if ctxDeadline, ok := ctx.Deadline(); ok && ctxDeadline.Before(deadline) {
		deadline = ctxDeadline
	}
	if err := pty.SetWriteDeadline(deadline); err != nil {
		return fmt.Errorf("set PTY write deadline: %w", err)
	}
	return nil
}

func watchPromptCancellation(ctx context.Context, pty *os.File) func() {
	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		select {
		case <-ctx.Done():
			_ = pty.SetWriteDeadline(time.Now())
		case <-stop:
		}
	}()
	return func() {
		close(stop)
		<-done
	}
}

func writePromptPhases(ctx context.Context, writer io.Writer, prompt []byte, enterPause time.Duration) (int, error) {
	written, err := writeAll(writer, prompt)
	if err != nil || enterPause == 0 || written != len(prompt) {
		return written, err
	}
	if err = waitForPromptEnter(ctx, enterPause); err != nil {
		return written, err
	}
	enterWritten, err := writePromptEnter(ctx, writer)
	return written + enterWritten, err
}

func waitForPromptEnter(ctx context.Context, pause time.Duration) error {
	timer := time.NewTimer(pause)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func writePromptEnter(ctx context.Context, writer io.Writer) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	return writeAll(writer, []byte{'\r'})
}

func classifySubmission(written, promptLength int, enterPause time.Duration, writeErr error) SubmissionEvidence {
	wantWritten := promptLength
	if enterPause > 0 {
		wantWritten++
	}
	evidence := SubmissionEvidence{State: SubmissionComplete, BytesWritten: written}
	if written == 0 {
		evidence.State = SubmissionNoBytes
	} else if written != wantWritten || writeErr != nil {
		evidence.State = SubmissionUncertain
	}
	return evidence
}

func validatePrompt(text string) error {
	if !utf8.ValidString(text) || strings.TrimSpace(text) == "" {
		return errors.New("message must be nonempty UTF-8 text")
	}
	if len([]byte(text)) > maxPromptBytes {
		return errors.New("message exceeds 2 KiB; put complex instructions in a file and send its path")
	}
	for _, r := range text {
		if r == '\n' || r == '\r' || r < 0x20 || r == 0x7f {
			return errors.New(
				"message must be one line without terminal control characters; put complex instructions in a file and send its path",
			)
		}
	}
	return nil
}

// ValidatePrompt checks generic CLI input constraints before an operation is accepted.
func ValidatePrompt(text string) error { return validatePrompt(text) }

func (d *GenericCLIDriver) acquireInput(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	d.gateOnce.Do(func() {
		if d.inputGate == nil {
			d.inputGate = make(chan struct{}, 1)
		}
	})
	select {
	case d.inputGate <- struct{}{}:
		if err := ctx.Err(); err != nil {
			<-d.inputGate
			return err
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (d *GenericCLIDriver) releaseInput() { <-d.inputGate }

func writeAll(writer io.Writer, data []byte) (int, error) {
	written := 0
	for written < len(data) {
		n, err := writer.Write(data[written:])
		written += n
		if err != nil {
			return written, err
		}
		if n == 0 {
			return written, io.ErrShortWrite
		}
	}
	return written, nil
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
	base := make([]string, 0, len(os.Environ())+len(extra)+1)
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if key == EnvKankuroMCPConfigPath || key == EnvKankuroMCPToken {
			continue
		}
		base = append(base, entry)
	}
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
