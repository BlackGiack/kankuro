//go:build e2e

package e2e_test

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tuitest "github.com/microsoft/tui-test/bindings/go"
	"github.com/microsoft/tui-test/bindings/go/tuitesttest"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/configstartup"
)

const mcpJourneyCommand = "go test -tags=e2e ./tests/e2e -run '^TestMCPManagerJourney$' -count=1 -timeout=3m -v"

// TestMCPManagerJourney drives the real configuration screens without launching
// an agent or connecting to a model provider. Captures are checkpoints, not
// pixel-baseline comparisons.
func TestMCPManagerJourney(t *testing.T) {
	j := newMCPJourney(t)
	j.expect("Dashboards")
	j.press("Enter")
	j.expect("Active Sessions")
	j.press("m")
	j.expect("MCP Servers")
	j.expect("kankuro (built-in, read-only)")
	j.expect("[ New MCP ]")
	j.capture("list", "Only the protected built-in and New are initially available.")

	j.press("Down")
	j.expect("> [ New MCP ]")
	j.press("Enter")
	j.expect("MCP Server Configuration")
	j.absent("Alpha MCP")
	j.capture("new", "New opens a blank MCP form with save/back instructions.")
	must(t, j.terminal.Type("Alpha MCP"))
	j.press("Tab")
	must(t, j.terminal.Type("mcp-placeholder"))
	j.press("Enter")
	j.expect("MCP Servers")
	j.expect("Alpha MCP")
	created := j.customServer()
	if created.Name != "Alpha MCP" || created.Command != "mcp-placeholder" || created.ID == "" {
		t.Fatalf("create did not persist form values: %+v", created)
	}

	// The New row stays selected after the registry refresh; return to Alpha.
	j.press("Up")
	j.press("Up")
	j.press("Enter")
	j.expect("MCP Server Configuration")
	j.expect("Alpha MCP")
	j.expect("mcp-placeholder")
	j.capture("edit", "Existing entry loads its name and command in the editor.")
	j.press("Ctrl+A")
	j.press("Ctrl+K")
	must(t, j.terminal.Type("Beta MCP"))
	j.press("Enter")
	j.expect("MCP Servers")
	j.expect("Beta MCP")
	j.absent("Alpha MCP")
	updated := j.customServer()
	if updated.ID != created.ID || updated.Name != "Beta MCP" || updated.Command != created.Command {
		t.Fatalf("edit changed identity or lost unedited command: %+v", updated)
	}
	j.resize(80, 24)
	j.expect("MCP Servers")
	j.capture("edited-list", "At 80×24 the renamed server, protected built-in, New and help are visible.")

	j.press("Delete")
	j.expect("Delete MCP?")
	j.expect("Delete Beta MCP?")
	j.capture("confirm", "Delete opens a named confirmation with confirm/cancel instructions at 80×24.")
	if got := j.customServer(); got.ID != created.ID {
		t.Fatal("opening confirmation mutated the registry")
	}
	j.press("Escape")
	j.expect("MCP Servers")
	j.expect("Beta MCP")
	if got := j.customServer(); got.ID != created.ID {
		t.Fatal("canceling confirmation removed the server")
	}
	j.press("Delete")
	j.expect("Delete Beta MCP?")
	j.press("Enter")
	j.expect("MCP Servers")
	j.absent("Beta MCP")
	servers, err := j.repo.ListMCPs()
	must(t, err)
	if len(servers) != 1 || servers[0].ID != config.BuiltinMCPKankuro {
		t.Fatalf("confirmed deletion did not preserve exactly the built-in: %+v", servers)
	}
	j.capture("deleted", "Confirmed deletion removes the custom definition and retains built-in/New at 80×24.")
	j.press("Escape")
	j.expect("Active Sessions")
	j.press("m")
	j.expect("MCP Servers")
	j.absent("Beta MCP")
	j.press("Escape")
	j.expect("Active Sessions")
	j.press("q")
	timeout := 10 * time.Second
	must(t, j.terminal.WaitExit(tuitest.WaitOptions{Timeout: &timeout}))
	assertNativeExit(t, j.terminal)
}

type mcpJourney struct {
	t         *testing.T
	terminal  *tuitest.Client
	repo      *config.FileRepository
	artifacts string
	run       *mcpJourneyRun
	size      [2]int
}

type mcpJourneyRun struct {
	Command     string       `json:"command"`
	Revision    string       `json:"revision"`
	Fingerprint string       `json:"fingerprint"`
	StartedAt   time.Time    `json:"startedAt"`
	FinishedAt  time.Time    `json:"finishedAt"`
	Status      string       `json:"status"`
	Captures    []mcpCapture `json:"captures"`
}

type mcpCapture struct {
	File      string    `json:"file"`
	Size      [2]int    `json:"size"`
	Assertion string    `json:"assertion"`
	At        time.Time `json:"at"`
}

func newMCPJourney(t *testing.T) *mcpJourney {
	t.Helper()
	root, err := filepath.Abs("../..")
	must(t, err)
	sandbox := t.TempDir()
	configDir := filepath.Join(sandbox, "config")
	must(t, os.MkdirAll(configDir, 0o700))
	must(t, configstartup.PrepareStartup(configDir))
	repo, err := config.NewFileRepository(configDir)
	must(t, err)
	artifacts := filepath.Join(root, "tests/e2e/artifacts/mcp-manager")
	if override := os.Getenv("KANKURO_MCP_E2E_ARTIFACT_DIR"); override != "" {
		artifacts, err = filepath.Abs(override)
		must(t, err)
	}
	must(t, os.MkdirAll(artifacts, 0o700))
	for _, name := range []string{"list.png", "new.png", "edit.png", "edited-list.png", "confirm.png", "deleted.png", "failure.png", "result.json"} {
		if removeErr := os.Remove(filepath.Join(artifacts, name)); removeErr != nil && !os.IsNotExist(removeErr) {
			t.Fatal(removeErr)
		}
	}
	run := &mcpJourneyRun{
		Command:     mcpJourneyCommand,
		Revision:    "WORKTREE",
		Fingerprint: sourceFingerprint(t, root),
		StartedAt:   time.Now().UTC(),
	}
	t.Cleanup(func() { finishMCPJourney(t, root, artifacts, run) })
	binary := filepath.Join(sandbox, "kankuro")
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	build := exec.CommandContext(ctx, "go", "build", "-o", binary, ".")
	build.Dir = root
	if output, buildErr := build.CombinedOutput(); buildErr != nil {
		t.Fatalf("build Kankuro: %v\n%s", buildErr, output)
	}
	// No native-agent config, conversations, credentials or user files are copied.
	env := map[string]string{
		"HOME": sandbox, "XDG_CONFIG_HOME": filepath.Join(sandbox, "xdg-config"),
		"XDG_DATA_HOME": filepath.Join(sandbox, "xdg-data"), "XDG_STATE_HOME": filepath.Join(sandbox, "xdg-state"),
		"XDG_CACHE_HOME": filepath.Join(sandbox, "xdg-cache"), "KANKURO_CONFIG_DIR": configDir,
		"KANKURO_TMP_DIR": filepath.Join(sandbox, "tmp"), "TERM": "xterm-256color", "COLORTERM": "truecolor",
	}
	terminal := tuitesttest.New(t, tuitesttest.Options{
		Prefix: "kankuro-mcp-manager", Program: binary,
		Client: tuitest.ClientOptions{Recording: &tuitest.AutomaticRecording{Mode: tuitest.RecordingDisabled}},
		Spawn:  tuitest.SpawnOptions{Cols: new(uint16(120)), Rows: new(uint16(40)), Cwd: sandbox, Env: env, WaitReady: new(false)},
	})
	j := &mcpJourney{t: t, terminal: terminal, repo: repo, artifacts: artifacts, run: run, size: [2]int{120, 40}}
	t.Cleanup(func() {
		if !t.Failed() {
			return
		}
		if _, captureErr := terminal.Screenshot(filepath.Join(artifacts, "failure.png"), tuitest.ScreenshotOptions{}); captureErr != nil {
			t.Logf("failure screenshot: %v", captureErr)
		}
		if text, textErr := terminal.Text(tuitest.TextOptions{}); textErr == nil {
			t.Logf("MCP journey terminal:\n%s", text)
		}
		if signalErr := terminal.Signal("SIGTERM"); signalErr != nil {
			t.Logf("shutdown signal: %v", signalErr)
		}
	})
	return j
}

func (j *mcpJourney) press(key string) {
	j.t.Helper()
	must(j.t, j.terminal.Press(key))
}

func (j *mcpJourney) expect(text string) {
	j.t.Helper()
	expectWithTimeout(j.t, j.terminal, text, 10*time.Second)
}

func (j *mcpJourney) absent(text string) {
	j.t.Helper()
	timeout := 10 * time.Second
	must(j.t, j.terminal.GetByText(text, tuitest.TextSelectorOptions{}).Expect(tuitest.LocatorExpectOptions{Not: true, Timeout: &timeout}))
}

func (j *mcpJourney) resize(cols, rows uint16) {
	j.t.Helper()
	must(j.t, j.terminal.Resize(cols, rows))
	j.size = [2]int{int(cols), int(rows)}
}

func (j *mcpJourney) capture(name, assertion string) {
	j.t.Helper()
	// Assert instructions as well as the main state before capturing the frame.
	switch name {
	case "confirm":
		j.expect("Enter/y=confirm")
	case "new", "edit":
		j.expect("Enter: Save")
	default:
		j.expect("Del=delete")
		j.expect("[ New MCP ]")
	}
	_, err := j.terminal.Screenshot(filepath.Join(j.artifacts, name+".png"), tuitest.ScreenshotOptions{})
	must(j.t, err)
	j.run.Captures = append(j.run.Captures, mcpCapture{File: name + ".png", Size: j.size, Assertion: assertion, At: time.Now().UTC()})
}

func (j *mcpJourney) customServer() config.MCPServer {
	j.t.Helper()
	servers, err := j.repo.ListMCPs()
	must(j.t, err)
	for _, server := range servers {
		if server.ID != config.BuiltinMCPKankuro && strings.HasSuffix(server.Name, "MCP") {
			return server
		}
	}
	j.t.Fatal("custom MCP definition not persisted")
	return config.MCPServer{}
}

func finishMCPJourney(t *testing.T, root, artifacts string, run *mcpJourneyRun) {
	t.Helper()
	if sourceFingerprint(t, root) != run.Fingerprint {
		t.Error("source changed during MCP journey")
	}
	run.FinishedAt, run.Status = time.Now().UTC(), "passed"
	if t.Failed() {
		run.Status = "failed"
	}
	data, marshalErr := json.MarshalIndent(run, "", "  ")
	if marshalErr != nil {
		t.Error(marshalErr)
		return
	}
	if writeErr := os.WriteFile(filepath.Join(artifacts, "result.json"), append(data, '\n'), 0o600); writeErr != nil {
		t.Error(writeErr)
	}
}
