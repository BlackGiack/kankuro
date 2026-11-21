//go:build e2e

package e2e_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	tuitest "github.com/microsoft/tui-test/bindings/go"
	"github.com/microsoft/tui-test/bindings/go/tuitesttest"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/configstartup"
	"github.com/BlackGiack/kankuro/internal/session"
)

// TestNativeLaunchFocusReturn edits the launch card and starts an interactive
// agent without submitting a prompt. Enter after detach must reattach to the
// created session.
func TestNativeLaunchFocusReturn(t *testing.T) {
	j := newLaunchFocusJourney(t, "")
	editAndReopenReasoningEffort(j)
	runLaunchFocus(j)
	j.finish()
}

func editAndReopenReasoningEffort(j *nativeJourney) {
	j.t.Helper()
	j.run.Stage = "edit and reopen launch card"
	j.press("e")
	j.expect("Edit ID Card")
	j.press("Tab") // Driver
	j.press("Tab") // Model
	j.press("Tab") // Reasoning effort
	must(j.t, j.terminal.Type("high"))
	j.press("Enter")
	j.expect("Active Sessions")

	j.press("e")
	j.expect("Edit ID Card")
	const effortRow = `Reasoning effort:[ >]*high\b`
	timeout := 10 * time.Second
	must(j.t, j.terminal.GetByText(effortRow, tuitest.TextSelectorOptions{Regex: true}).
		Expect(tuitest.LocatorExpectOptions{Timeout: &timeout}))
	j.capture("card-effort-reopened", "Reopening the saved card shows high on the Reasoning effort row.")
	j.press("Escape")
	j.expect("Active Sessions")
}

func runLaunchFocus(j *nativeJourney) session.Data {
	t := j.t
	j.run.Stage = "launch"
	j.press("Enter")
	j.expect("Sub Workspaces:")
	j.press("Enter")
	j.expect("Session launched successfully!")
	created := waitSession(t, j.configDir, 30*time.Second, func(data session.Data) bool {
		return data.ID != "older-session" && data.CardID == j.card.ID && data.PID > 0
	})
	// Type only into Kankuro's naming form, before the agent terminal is attached.
	must(t, j.terminal.Type("Created focus session"))
	j.press("Enter")
	expectIdleCodex(t, j.terminal)
	j.capture("attached", "The real Codex input screen is visible after card launch and name confirmation; no agent prompt was entered.")

	j.run.Stage = "detach"
	j.press("Ctrl+K")
	j.expect("Active Sessions")
	j.expect("[Session] Enter=Attach")
	j.expect("Created focus session")
	j.expect("Older session")
	j.expect("Started:")
	j.expect("Card: Focus E2E")
	state, err := j.terminal.State()
	must(t, err)
	if state.Cols != 120 || state.Rows != 40 {
		t.Fatalf("dashboard terminal size = %dx%d, want 120x40", state.Cols, state.Rows)
	}
	j.capture("dashboard", "Ctrl+K restores the dashboard with both sessions, action hints and the selected session header.")

	// Check keyboard restoration without changing the sessions selection.
	j.press("c")
	j.expect("Settings")
	j.press("Escape")
	j.expect("[Session] Enter=Attach")
	j.run.Stage = "reattach"
	j.press("Enter")
	expectIdleCodex(t, j.terminal)
	assertLaunchFocusSessions(t, j.configDir, created)
	j.capture(
		"reattached",
		"Enter without pane navigation reattaches to the same live process; exactly two persisted sessions remain and the older session stays terminated.",
	)
	return created
}

func newLaunchFocusJourney(t *testing.T, workspaceMode string) *nativeJourney {
	t.Helper()
	if _, err := exec.LookPath("codex"); err != nil {
		t.Skip("startup-only launch-focus test requires Codex on PATH")
	}
	if os.Getenv("TUI_TEST_RECORDING_FONT_FAMILIES") == "" {
		t.Setenv("TUI_TEST_RECORDING_FONT_FAMILIES", "DejaVu Sans Mono,DejaVu Sans")
	}
	verifyNativeCapture(t)
	root, err := filepath.Abs("../..")
	must(t, err)
	artifacts := filepath.Join(root, "tests/e2e/artifacts/launch-focus")
	if override := os.Getenv("KANKURO_E2E_ARTIFACT_DIR"); override != "" {
		artifacts, err = filepath.Abs(override)
		must(t, err)
	}
	if workspaceMode != "" {
		artifacts = filepath.Join(artifacts, workspaceMode)
	}
	must(t, os.MkdirAll(artifacts, 0o700))
	for _, name := range []string{
		"card-effort-reopened.png", "attached.png", "dashboard.png", "reattached.png", "failure.png", "result.json",
		"workspace-initial.png", "workspace-updated.png", "supported-minimum.png", "minimum-size.png",
		"stopped.png", "relaunched.png", "report-after-relaunch.png", "restart-report.png",
	} {
		if removeErr := os.Remove(filepath.Join(artifacts, name)); removeErr != nil && !os.IsNotExist(removeErr) {
			t.Fatal(removeErr)
		}
	}
	run := &journeyRun{
		Agent: "codex", Revision: "WORKTREE", Fingerprint: sourceFingerprint(t, root),
		Command:   "go test -tags=e2e ./tests/e2e -run '^TestNativeLaunchFocusReturn$' -count=1 -timeout=3m -v",
		StartedAt: time.Now().UTC(), Size: [2]int{120, 40}, Stage: "setup",
	}
	if workspaceMode != "" {
		run.Command = "go test -tags=e2e ./tests/e2e -run '^TestCodexOfflineWorkspaceAndReports$/" + workspaceMode + "$' -count=1 -timeout=3m -v"
	}
	t.Cleanup(func() {
		if sourceFingerprint(t, root) != run.Fingerprint {
			t.Error("source changed during launch-focus E2E run")
		}
		writeJourneyRun(t, run, artifacts)
	})

	sandbox := t.TempDir()
	environment := startupOnlyCodexEnvironment(t, sandbox)
	// Recent Codex versions start a daemon outside the CLI's process group.
	// Stop only the daemon in this disposable home before removing its files.
	t.Cleanup(func() {
		defer stopCodexUpdater(t, environment["CODEX_HOME"])
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		stop := exec.CommandContext(ctx, "codex", "app-server", "daemon", "stop")
		stop.Env = os.Environ()
		for key, value := range environment {
			stop.Env = append(stop.Env, key+"="+value)
		}
		if output, stopErr := stop.CombinedOutput(); stopErr != nil {
			t.Errorf("stop isolated Codex daemon: %v\n%s", stopErr, output)
		}
	})
	workspace := filepath.Join(sandbox, "workspace")
	must(t, os.MkdirAll(workspace, 0o700))
	if workspaceMode == "git" {
		git := exec.Command("git", "init", "-q")
		git.Dir = workspace
		must(t, git.Run())
		must(t, os.WriteFile(filepath.Join(workspace, ".git/info/exclude"), []byte(".kankuro/\n.agents/\nAGENTS.md\n"), 0600))
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
		DashboardID: "default", Name: "Focus E2E", Driver: "codex", WorkspaceID: workspaces[0].ID, Puppet: "kankuro",
	})
	must(t, err)
	older := session.Data{
		SchemaVersion: 2, ID: "older-session", DashboardID: "default", CardID: card.ID,
		Name: "Older session", WorkingDir: workspace, OriginalDir: workspace,
		Status: session.StatusTerminated, StartedAt: time.Unix(1, 0),
	}
	data, err := json.Marshal(older)
	must(t, err)
	must(t, os.MkdirAll(filepath.Join(configDir, "sessions"), 0o700))
	must(t, os.WriteFile(filepath.Join(configDir, "sessions/older-session.json"), data, 0o600))
	if workspaceMode != "" {
		seedOfflineReport(t, configDir, older.ID)
	}
	binary := filepath.Join(sandbox, "kankuro")
	buildCtx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	defer cancel()
	build := exec.CommandContext(buildCtx, "go", "build", "-o", binary, ".")
	build.Dir = root
	if output, buildErr := build.CombinedOutput(); buildErr != nil {
		t.Fatalf("build Kankuro: %v\n%s", buildErr, output)
	}
	environment["KANKURO_CONFIG_DIR"] = configDir
	environment["KANKURO_TMP_DIR"] = filepath.Join(sandbox, "tmp")
	terminal := tuitesttest.New(t, tuitesttest.Options{
		Program: binary, Prefix: "kankuro-launch-focus",
		Client: tuitest.ClientOptions{Recording: &tuitest.AutomaticRecording{Mode: tuitest.RecordingDisabled}},
		Spawn: tuitest.SpawnOptions{
			Cols: new(uint16(120)), Rows: new(uint16(40)), Cwd: workspace,
			Env: environment, WaitReady: new(false),
		},
	})
	registerTerminalCleanup(t, terminal, run, artifacts)
	j := &nativeJourney{t: t, root: root, agent: nativeAgent{name: "codex"},
		artifacts: artifacts, configDir: configDir, card: card, terminal: terminal, run: run}
	j.expect("Dashboards")
	j.press("Enter")
	j.expect("Focus E2E")
	return j
}

// The daemon stop command leaves its updater running in some Codex versions.
// Match both the private executable path and exact subcommand; never stop an
// updater belonging to the user's real home or another test.
func stopCodexUpdater(t *testing.T, codexHome string) {
	t.Helper()
	output, err := exec.Command("ps", "-eo", "pid=,args=").Output()
	must(t, err)
	for line := range strings.SplitSeq(string(output), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 5 || !strings.HasPrefix(fields[1], codexHome+string(os.PathSeparator)) ||
			fields[2] != "app-server" || fields[3] != "daemon" || fields[4] != "pid-update-loop" {
			continue
		}
		pid, parseErr := strconv.Atoi(fields[0])
		must(t, parseErr)
		process, findErr := os.FindProcess(pid)
		must(t, findErr)
		if killErr := process.Kill(); killErr != nil && !errors.Is(killErr, os.ErrProcessDone) && !os.IsNotExist(killErr) {
			t.Errorf("stop isolated Codex updater: %v", killErr)
		}
	}
}

// Use a synthetic key and an unreachable loopback provider for idle Codex tests.
func startupOnlyCodexEnvironment(t *testing.T, sandbox string) map[string]string {
	t.Helper()
	home := filepath.Join(sandbox, "home")
	codexHome := filepath.Join(home, ".codex")
	must(t, os.MkdirAll(codexHome, 0o700))
	const endpoint = "http://127.0.0.1:1"
	const fixtureKey = "appediff-offline-fixture"
	must(t, os.WriteFile(filepath.Join(codexHome, "auth.json"), []byte(`{"OPENAI_API_KEY":"`+fixtureKey+`"}`), 0600))
	settings := "check_for_update_on_startup = false\nmodel = \"gpt-6-sol\"\nopenai_base_url = \"" + endpoint + "/v1\"\n"
	must(t, os.WriteFile(filepath.Join(codexHome, "config.toml"), []byte(settings), 0600))
	return map[string]string{
		"HOME": home, "CODEX_HOME": codexHome, "XDG_CONFIG_HOME": filepath.Join(home, ".config"),
		"XDG_DATA_HOME": filepath.Join(home, ".local/share"), "XDG_STATE_HOME": filepath.Join(home, ".local/state"),
		"XDG_CACHE_HOME": filepath.Join(home, ".cache"), "TERM": "xterm-256color", "COLORTERM": "truecolor",
		"NO_COLOR": "", "CLICOLOR_FORCE": "1",
		"OPENAI_API_KEY": fixtureKey, "OPENAI_BASE_URL": endpoint + "/v1",
		"HTTP_PROXY": endpoint, "HTTPS_PROXY": endpoint, "ALL_PROXY": endpoint, "NO_PROXY": "localhost,127.0.0.1",
	}
}

func expectIdleCodex(t *testing.T, terminal *tuitest.Client) {
	t.Helper()
	timeout := 45 * time.Second
	for _, text := range []string{"Active Sessions", "Session launched successfully!"} {
		must(t, terminal.GetByText(text, tuitest.TextSelectorOptions{}).
			Wait(tuitest.LocatorWaitOptions{State: tuitest.Hidden, Timeout: &timeout}))
	}
	expectWithTimeout(t, terminal, "OpenAI Codex", 45*time.Second)
	expectWithTimeout(t, terminal, "Ask Codex to do anything", 45*time.Second)
	// The empty input exists during the splash too; wait for model metadata so
	// captures show the ready native UI rather than its transient loading frame.
	must(t, terminal.GetByText("(?i)GPT-|model:", tuitest.TextSelectorOptions{Regex: true}).
		First().Expect(tuitest.LocatorExpectOptions{Timeout: &timeout}))
	checkAgentFailure(t, terminal)
}

func assertLaunchFocusSessions(t *testing.T, configDir string, created session.Data) {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(configDir, "sessions/*.json"))
	must(t, err)
	if len(paths) != 2 {
		t.Fatalf("session count = %d, want older session and exactly one created session", len(paths))
	}
	for _, path := range paths {
		data, readErr := os.ReadFile(path)
		must(t, readErr)
		var current session.Data
		must(t, json.Unmarshal(data, &current))
		if current.ID == created.ID {
			if current.PID != created.PID || current.Name != "Created focus session" || current.Status == session.StatusTerminated {
				t.Fatalf("created session identity or activation changed: %+v", current)
			}
		} else if current.ID != "older-session" || current.PID != 0 || current.Status != session.StatusTerminated {
			t.Fatalf("Enter selected or relaunched the wrong session: %+v", current)
		}
	}
}
