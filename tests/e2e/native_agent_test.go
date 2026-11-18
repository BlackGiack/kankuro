//go:build e2e

package e2e_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tuitest "github.com/microsoft/tui-test/bindings/go"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/session"
)

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
	for _, name := range []string{"launched.png", "dashboard.png", "workspace-before-report.png", "live-report.png", "conversation-choice.png", "resume-ready.png", "resumed.png", "failure.png", "result.json"} {
		if removeErr := os.Remove(filepath.Join(artifacts, name)); removeErr != nil && !os.IsNotExist(removeErr) {
			t.Fatal(removeErr)
		}
	}
	run := &journeyRun{
		Agent: agent.name, Command: "./tests/e2e/run token-usage",
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
	if t.Failed() {
		run.Status = "failed"
	} else if run.Status == "" {
		run.Status = "passed"
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

type tokenUsage struct {
	Input  int `json:"input_tokens"`
	Cached int `json:"cached_input_tokens"`
	Output int `json:"output_tokens"`
}

type journeyRun struct {
	TokenUsage     *tokenUsage `json:"tokenUsage,omitempty"`
	Fingerprint    string      `json:"fingerprint"`
	Revision       string      `json:"revision"`
	Model          string      `json:"model,omitempty"`
	Agent          string      `json:"agent"`
	Command        string      `json:"command"`
	ConversationID string      `json:"conversationId,omitempty"`
	StartedAt      time.Time   `json:"startedAt"`
	FinishedAt     time.Time   `json:"finishedAt"`
	Status         string      `json:"status"`
	Stage          string      `json:"stage"`
	SkipReason     string      `json:"skipReason,omitempty"`
	Size           [2]int      `json:"size"`
	Assertions     []string    `json:"assertions"`
	Captures       []string    `json:"captures"`
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
