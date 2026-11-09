//go:build e2e

package e2e_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	tuitest "github.com/microsoft/tui-test/bindings/go"
	"github.com/microsoft/tui-test/bindings/go/tuitesttest"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/configstartup"
	"github.com/BlackGiack/kankuro/internal/session"
)

// TestNativeAgentJourney deliberately requires a configured installed agent.
// Missing executables skip; authentication, configuration and journey failures do not.
func TestNativeAgentJourney(t *testing.T) {
	j := newNativeJourney(t)
	first := j.launch()
	nonce := j.register(first)
	j.detach()
	j.resume(first, nonce)
	j.finish()
}

type nativeJourney struct {
	t         *testing.T
	root      string
	agent     nativeAgent
	artifacts string
	configDir string
	card      config.IDCard
	terminal  *tuitest.Client
	run       *journeyRun
}

func newNativeJourney(t *testing.T) *nativeJourney {
	t.Helper()
	if os.Getenv("TUI_TEST_RECORDING_FONT_FAMILIES") == "" {
		t.Setenv("TUI_TEST_RECORDING_FONT_FAMILIES", "DejaVu Sans Mono,DejaVu Sans")
	}
	agent := selectAgent(t)
	root, err := filepath.Abs("../..")
	must(t, err)
	run, artifacts := newJourneyRun(t, root, agent)
	j := &nativeJourney{t: t, root: root, agent: agent, artifacts: artifacts, run: run}
	verifyNativeCapture(t)
	sandbox := t.TempDir()
	environment := isolatedAgentEnvironment(t, sandbox, j.agent.name)
	workspace := filepath.Join(sandbox, "workspace")
	must(t, os.MkdirAll(workspace, 0o700))
	// Keep approvals confined to the disposable workspace; preserve the user's model/provider.
	if j.agent.name == "opencode" {
		j.run.Model = configureOpenCodeWorkspace(t, workspace)
	}
	configDir := filepath.Join(sandbox, "config")
	must(t, os.MkdirAll(configDir, 0o700))
	must(t, configstartup.PrepareStartup(configDir))
	repo, err := config.NewFileRepository(configDir)
	must(t, err)
	must(t, repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: workspace}}))
	workspaces, err := repo.ListWorkspaces()
	must(t, err)
	if len(workspaces) != 1 {
		t.Fatalf("workspace registry = %+v", workspaces)
	}
	card, err := repo.CreateIDCard(config.IDCard{
		DashboardID: "default",
		Name:        "Native E2E",
		Driver:      j.agent.name,
		WorkspaceID: workspaces[0].ID,
		Puppet:      "kankuro",
		SkillIDs:    []string{config.BuiltinSkillMCPCLI, config.BuiltinSkillKankuroPuppet},
		MCPIDs:      []string{config.BuiltinMCPKankuro},
		AgentsMD:    "This is a disposable terminal acceptance test. Follow the user's registration instructions exactly. Use the real native conversation ID, never invent one. Use mcp-cli with .kankuro/mcp-servers.json. Do not display credentials.",
	})
	must(t, err)
	binary := filepath.Join(sandbox, "kankuro")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	buildCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	build := exec.CommandContext(buildCtx, "go", "build", "-o", binary, ".")
	build.Dir = j.root
	if output, buildErr := build.CombinedOutput(); buildErr != nil {
		t.Fatalf("build Kankuro: %v\n%s", buildErr, output)
	}
	environment["KANKURO_CONFIG_DIR"] = configDir
	environment["KANKURO_TMP_DIR"] = filepath.Join(sandbox, "tmp")
	terminal := tuitesttest.New(t, tuitesttest.Options{
		Program: binary, Prefix: "kankuro-native",
		Client: tuitest.ClientOptions{Recording: &tuitest.AutomaticRecording{Mode: tuitest.RecordingDisabled}},
		Spawn: tuitest.SpawnOptions{
			Cols: new(uint16(120)), Rows: new(uint16(40)), Cwd: workspace,
			Env: environment, WaitReady: new(false),
		},
	})
	j.terminal, j.configDir, j.card = terminal, configDir, card
	registerTerminalCleanup(t, terminal, j.run, j.artifacts)
	return j
}

func newJourneyRun(t *testing.T, root string, agent nativeAgent) (*journeyRun, string) {
	t.Helper()
	artifacts := filepath.Join(root, "tests/e2e/artifacts/native-agent")
	if override := os.Getenv("KANKURO_E2E_ARTIFACT_DIR"); override != "" {
		resolved, pathErr := filepath.Abs(override)
		must(t, pathErr)
		artifacts = resolved
	}
	// Only remove this test's known artifacts; never clear an arbitrary supplied directory.
	must(t, os.MkdirAll(artifacts, 0o700))
	for _, name := range []string{"launched.png", "registered.png", "dashboard.png", "resumed.png", "failure.png", "result.json"} {
		if removeErr := os.Remove(filepath.Join(artifacts, name)); removeErr != nil && !os.IsNotExist(removeErr) {
			t.Fatal(removeErr)
		}
	}
	run := &journeyRun{
		Agent: agent.name, Command: "go test -tags=e2e ./tests/e2e -run '^TestNativeAgentJourney$' -count=1 -timeout=10m -v",
		Fingerprint: sourceFingerprint(t, root), Revision: "WORKTREE",
		StartedAt: time.Now().UTC(), Size: [2]int{120, 40}, Stage: "setup",
	}
	t.Cleanup(func() {
		if after := sourceFingerprint(t, root); after != run.Fingerprint {
			t.Error("source changed during native E2E run")
		}
		writeJourneyRun(t, run, artifacts)
	})
	return run, artifacts
}

func writeJourneyRun(t *testing.T, run *journeyRun, artifacts string) {
	t.Helper()
	run.FinishedAt = time.Now().UTC()
	run.Status = "passed"
	if t.Failed() {
		run.Status = "failed"
	}
	data, err := json.MarshalIndent(run, "", "  ")
	if err != nil {
		t.Errorf("encode run: %v", err)
		return
	}
	if writeErr := os.WriteFile(filepath.Join(artifacts, "result.json"), append(data, '\n'), 0o600); writeErr != nil {
		t.Errorf("record run: %v", writeErr)
	}
}

func registerTerminalCleanup(t *testing.T, terminal *tuitest.Client, run *journeyRun, artifacts string) {
	t.Helper()
	// Give Kankuro time to stop its managed process before the terminal closes.
	t.Cleanup(func() {
		if t.Failed() {
			if signalErr := terminal.Signal("SIGTERM"); signalErr != nil {
				t.Logf("signal Kankuro: %v", signalErr)
				return
			}
			timeout := 15 * time.Second
			if exitErr := terminal.WaitExit(tuitest.WaitOptions{Timeout: &timeout}); exitErr != nil {
				t.Errorf("Kankuro shutdown: %v", exitErr)
			}
		}
	})
	// Registered after terminal cleanup, so failure capture happens before closure.
	t.Cleanup(func() {
		if t.Failed() {
			if _, captureErr := terminal.Screenshot(
				filepath.Join(artifacts, "failure.png"),
				tuitest.ScreenshotOptions{},
			); captureErr != nil {
				t.Logf("failure screenshot: %v", captureErr)
			} else {
				run.Captures = append(run.Captures, "failure.png")
			}
			if text, textErr := terminal.Text(tuitest.TextOptions{}); textErr == nil {
				t.Logf("terminal at %s:\n%s", run.Stage, text)
			}
		}
	})
}

func (j *nativeJourney) launch() session.Data {
	t := j.t
	j.run.Stage = "launch"
	j.expect("Dashboards")
	j.press("Enter")
	j.expect("Native E2E")
	j.press("Enter")
	j.expect("Session launched successfully!")
	first := waitSession(t, j.configDir, 30*time.Second, func(data session.Data) bool { return data.CardID == j.card.ID && data.PID > 0 })
	j.press("Enter")
	// Wait for native UI, not the launcher or an assumed attachment banner.
	waitAgent(t, j.terminal, j.agent.name)
	handleTrust(t, j.terminal, j.agent.name)
	if j.agent.name == "opencode" {
		j.expect(openCodeTestModelName)
	}
	j.capture("launched", "Dashboard Enter launched a persisted native-agent process and Enter attached to the native UI.")

	return first
}

func (j *nativeJourney) register(first session.Data) string {
	t := j.t
	j.run.Stage = "register"
	secret := make([]byte, 12)
	_, err := rand.Read(secret)
	must(t, err)
	nonce := hex.EncodeToString(secret)
	// Later requests omit the nonce: returning it then requires native conversation history.
	instructions := "Discover your ACTUAL native conversation ID, follow the kankuro-puppet skill and register it with name Kankuro native E2E through mcp-cli."
	if j.agent.name == "opencode" {
		instructions = "Use your shell tool now. First run opencode session list --format json in this workspace. Take the id, not projectId, of this workspace's native session. Then run MCP_NO_DAEMON=1 mcp-cli -c .kankuro/mcp-servers.json call kankuro register_conversation with JSON arguments containing conversation_id set to that actual id and name set to Kankuro native E2E. Execute the commands; loading a skill alone does not complete this task."
	}
	prompt := fmt.Sprintf(
		"Remember the secret %s in this conversation only; do not write it to any file. %s Verify the registration result has status registered and persistence saved. Do not print credentials. Reply READY only after successful registration.",
		nonce,
		instructions,
	)
	must(t, j.terminal.Type(prompt))
	j.press("Enter")
	registered := waitSession(t, j.configDir, 4*time.Minute, func(data session.Data) bool {
		checkAgentFailure(t, j.terminal)
		return data.ID == first.ID && len(data.Conversations[j.card.ID]) == 1
	})
	binding := registered.Conversations[j.card.ID][0]
	if binding.ID == "" || binding.Driver != j.agent.name || binding.Name != "Kankuro native E2E" {
		t.Fatalf("unexpected persisted native binding: %+v", binding)
	}
	j.run.ConversationID = binding.ID
	j.capture(
		"registered",
		"The native agent registered a conversation through the deployed MCP endpoint; its ID, name and driver were persisted under the launched session/card.",
	)

	return nonce
}

func (j *nativeJourney) detach() {
	j.run.Stage = "detach"
	j.press("Ctrl+K")
	j.expect("Active Sessions")
	// Navigate away and back to establish that ordinary dashboard input works after detach.
	j.press("c")
	j.expect("Settings")
	j.press("Escape")
	j.expect("Active Sessions")
	j.capture("dashboard", "Ctrl+K restored the dashboard; c opened settings and Escape returned using ordinary keyboard input.")
}

func (j *nativeJourney) resume(first session.Data, nonce string) {
	t := j.t
	j.run.Stage = "resume"
	// A fresh dashboard starts on cards; Tab focuses sessions regardless of attach return focus.
	// Inspect the visible footer to avoid toggling an already-focused session pane.
	text, err := j.terminal.Text(tuitest.TextOptions{})
	must(t, err)
	if !strings.Contains(text, "[Session]") {
		j.press("Tab")
	}
	j.press("x")
	waitSession(
		t,
		j.configDir,
		30*time.Second,
		func(data session.Data) bool { return data.ID == first.ID && data.Status == session.StatusTerminated },
	)
	// Refresh is asynchronous: require the terminated status to reach the UI before Enter.
	j.expect("Terminated")
	j.press("Enter")
	j.expect("New conversation")
	j.expect("Kankuro native E2E")
	// Saved bindings precede the New conversation row; explicitly select the first saved binding.
	j.press("Home")
	j.press("Enter")
	waitAgent(t, j.terminal, j.agent.name)
	resumed := waitSession(t, j.configDir, 30*time.Second, func(data session.Data) bool {
		return data.ID == first.ID && data.PID > 0 && data.PID != first.PID && data.Status != session.StatusTerminated
	})
	if resumed.WorkingDir != first.WorkingDir {
		t.Fatal("resume changed workspace")
	}
	handleTrust(t, j.terminal, j.agent.name)
	// Remove replayed prompt text from the visible screen before requesting continuity.
	must(
		t,
		j.terminal.Type(
			"Without reading files or using tools, reply with the secret from our earlier conversation as CONTINUITY:<secret>.",
		),
	)
	j.press("Enter")
	expectWithTimeout(t, j.terminal, "CONTINUITY:"+nonce, 3*time.Minute)
	j.capture(
		"resumed",
		"The saved conversation was explicitly selected; the replacement process kept the session workspace and recalled a random secret absent from the resume prompt.",
	)
}

func (j *nativeJourney) finish() {
	t := j.t
	j.press("Ctrl+K")
	j.expect("Active Sessions")
	j.press("q")
	timeout := 15 * time.Second
	must(t, j.terminal.WaitExit(tuitest.WaitOptions{Timeout: &timeout}))
	assertNativeExit(t, j.terminal)
	j.run.Stage = "complete"
}

func (j *nativeJourney) expect(text string) {
	j.t.Helper()
	expectWithTimeout(j.t, j.terminal, text, 45*time.Second)
}

func (j *nativeJourney) press(key string) {
	j.t.Helper()
	must(j.t, j.terminal.Press(key))
}

func (j *nativeJourney) capture(name, assertion string) {
	j.t.Helper()
	_, err := j.terminal.Screenshot(filepath.Join(j.artifacts, name+".png"), tuitest.ScreenshotOptions{})
	must(j.t, err)
	j.run.Assertions = append(j.run.Assertions, assertion)
	j.run.Captures = append(j.run.Captures, name+".png")
}

type nativeAgent struct{ name string }

type journeyRun struct {
	Fingerprint    string    `json:"fingerprint"`
	Revision       string    `json:"revision"`
	Model          string    `json:"model,omitempty"`
	Agent          string    `json:"agent"`
	Command        string    `json:"command"`
	ConversationID string    `json:"conversationId,omitempty"`
	StartedAt      time.Time `json:"startedAt"`
	FinishedAt     time.Time `json:"finishedAt"`
	Status         string    `json:"status"`
	Stage          string    `json:"stage"`
	Size           [2]int    `json:"size"`
	Assertions     []string  `json:"assertions"`
	Captures       []string  `json:"captures"`
}

func selectAgent(t *testing.T) nativeAgent {
	t.Helper()
	for _, name := range []string{"opencode", "codex", "claude"} {
		if _, err := exec.LookPath(name); err == nil {
			return nativeAgent{name: name}
		}
	}
	t.Skip("no opencode, codex or claude executable available")
	return nativeAgent{}
}

func expectWithTimeout(t *testing.T, terminal *tuitest.Client, text string, timeout time.Duration) {
	t.Helper()
	must(t, terminal.GetByText(text, tuitest.TextSelectorOptions{}).First().Expect(tuitest.LocatorExpectOptions{Timeout: &timeout}))
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

// waitSession polls atomic application records, never treats terminal silence as completion.
func waitSession(t *testing.T, configDir string, timeout time.Duration, ready func(session.Data) bool) session.Data {
	t.Helper()
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()
	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()
	var latest session.Data
	for {
		paths, err := filepath.Glob(filepath.Join(configDir, "sessions", "*.json"))
		must(t, err)
		for _, path := range paths {
			bytes, readErr := os.ReadFile(path)
			if readErr != nil {
				continue
			}
			var data session.Data
			if json.Unmarshal(bytes, &data) != nil {
				continue
			}
			latest = data
			if ready(data) {
				return data
			}
		}
		select {
		case <-deadline.C:
			t.Fatalf("session state timed out: id=%s status=%s conversations=%v", latest.ID, latest.Status, latest.Conversations)
		case <-tick.C:
		}
	}
}

func handleTrust(t *testing.T, terminal *tuitest.Client, agent string) {
	t.Helper()
	// Trust prompts differ by CLI. Auto-confirm only known workspace prompts,
	// leaving login/model errors to fail the journey rather than skipping it.
	text, err := terminal.Text(tuitest.TextOptions{})
	must(t, err)
	if agent == "codex" && strings.Contains(text, "trust") {
		must(t, terminal.Press("Enter"))
	}
	if agent == "claude" && strings.Contains(text, "trust") {
		must(t, terminal.Press("Enter"))
	}
}

func waitAgent(t *testing.T, terminal *tuitest.Client, agent string) {
	t.Helper()
	pattern := "ctrl\\+p commands"
	switch agent {
	case "codex":
		pattern = "(?i)codex|trust this"
	case "claude":
		pattern = "(?i)claude|trust this"
	}
	timeout := 45 * time.Second
	must(
		t,
		terminal.GetByText("Session launched successfully!", tuitest.TextSelectorOptions{}).
			Wait(tuitest.LocatorWaitOptions{State: tuitest.Hidden, Timeout: &timeout}),
	)
	must(
		t,
		terminal.GetByText(pattern, tuitest.TextSelectorOptions{Regex: true}).
			First().
			Expect(tuitest.LocatorExpectOptions{Timeout: &timeout}),
	)
	checkAgentFailure(t, terminal)
}

func checkAgentFailure(t *testing.T, terminal *tuitest.Client) {
	t.Helper()
	text, err := terminal.Text(tuitest.TextOptions{})
	must(t, err)
	for _, failure := range []string{"Free limit reached", "Authentication required", "Invalid API key", "Server process exited", "Failed to start server"} {
		if strings.Contains(text, failure) {
			t.Fatalf("native agent reported %q", failure)
		}
	}
}
