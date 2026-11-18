//go:build e2e

package e2e_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	tuitest "github.com/microsoft/tui-test/bindings/go"
	"github.com/microsoft/tui-test/bindings/go/tuitesttest"

	"github.com/BlackGiack/kankuro/internal/session"
)

const fixtureReport = "Stored report from older activation"

func TestCodexOfflineWorkspaceAndReports(t *testing.T) {
	for _, mode := range []string{"directory", "git"} {
		t.Run(mode, func(t *testing.T) {
			j := newLaunchFocusJourney(t, mode)
			created := runLaunchFocus(j)
			j.press("Ctrl+K")
			j.expect("Active Sessions")
			checkOfflineWorkspace(j, created.WorkingDir, mode)
			stopAndRelaunchOffline(j, created)
			checkOfflineReportAfterRestart(j, created.WorkingDir)
		})
	}
}

func checkOfflineWorkspace(j *nativeJourney, workspace, mode string) {
	t := j.t
	j.run.Stage = "workspace refresh"
	initial := "Directory:"
	if mode == "git" {
		initial = "Working tree clean"
	}
	j.expect(initial)
	j.capture("workspace-initial", "The center displays "+initial+" for the selected idle Codex workspace.")
	const file = "refresh-check.txt"
	must(t, os.WriteFile(filepath.Join(workspace, file), []byte("offline fixture"), 0600))
	expectWithTimeout(t, j.terminal, file, 45*time.Second)
	j.capture("workspace-updated", "The live workspace refresh displays the new file.")
	must(t, j.terminal.Resize(60, 20))
	j.expect("Terminal too small")
	j.capture("minimum-size", "At 60×20 cells the dashboard displays its minimum-size warning.")
	must(t, j.terminal.Resize(120, 40))
	j.expect("Active Sessions")
}

func stopAndRelaunchOffline(j *nativeJourney, created session.Data) {
	t := j.t
	j.run.Stage = "stop and relaunch"
	j.press("x")
	stopped := waitSession(t, j.configDir, 30*time.Second, func(data session.Data) bool {
		return data.ID == created.ID && data.Status == session.StatusTerminated
	})
	if stopped.WorkingDir != created.WorkingDir {
		t.Fatal("stop changed workspace")
	}
	j.capture("stopped", "The selected session is stopped and its persisted workspace is unchanged.")
	j.press("Enter")
	expectIdleCodex(t, j.terminal)
	relaunched := waitSession(t, j.configDir, 30*time.Second, func(data session.Data) bool {
		return data.ID == created.ID && data.PID > 0 && data.PID != created.PID
	})
	if relaunched.WorkingDir != created.WorkingDir {
		t.Fatal("relaunch changed workspace")
	}
	j.capture("relaunched", "The same session and workspace have a new Codex process.")
	j.press("Ctrl+K")
	j.expect("Active Sessions")
	j.press("Down")
	j.expect(fixtureReport)
	j.press("Enter")
	expectIdleCodex(t, j.terminal)
	waitSession(t, j.configDir, 30*time.Second, func(data session.Data) bool {
		return data.ID == "older-session" && data.PID > 0 && data.Status != session.StatusTerminated
	})
	j.press("Ctrl+K")
	j.expect(fixtureReport)
	j.capture("report-after-relaunch", "The seeded report remains visible after the author session relaunches with Codex.")
	quitOfflineDashboard(j)
}

func quitOfflineDashboard(j *nativeJourney) {
	j.press("q")
	timeout := 15 * time.Second
	must(j.t, j.terminal.WaitExit(tuitest.WaitOptions{Timeout: &timeout}))
	assertNativeExit(j.t, j.terminal)
}

func checkOfflineReportAfterRestart(j *nativeJourney, workspace string) {
	t := j.t
	j.run.Stage = "report reconstruction"
	sandbox := filepath.Dir(j.configDir)
	env := startupOnlyCodexEnvironment(t, sandbox)
	env["KANKURO_CONFIG_DIR"] = j.configDir
	env["KANKURO_TMP_DIR"] = filepath.Join(sandbox, "tmp")
	j.terminal = tuitesttest.New(t, tuitesttest.Options{
		Program: filepath.Join(sandbox, "kankuro"), Prefix: "kankuro-offline-restart",
		Client: tuitest.ClientOptions{Recording: &tuitest.AutomaticRecording{Mode: tuitest.RecordingDisabled}},
		Spawn:  tuitest.SpawnOptions{Cols: new(uint16(120)), Rows: new(uint16(40)), Cwd: workspace, Env: env, WaitReady: new(false)},
	})
	registerTerminalCleanup(t, j.terminal, j.run, j.artifacts)
	j.expect("Dashboards")
	j.press("Enter")
	j.expect("Active Sessions")
	j.expect(fixtureReport)
	j.capture("restart-report", "Kankuro reconstructs the selected session’s persisted report after restart.")
	quitOfflineDashboard(j)
	j.run.Stage = "complete"
}

func seedOfflineReport(t *testing.T, configDir, recipient string) {
	t.Helper()
	store, err := session.NewMessageStore(configDir)
	must(t, err)
	message, err := store.CreateIntent(session.MessageRecord{
		DashboardID: "default", SenderSessionID: "fixture-orchestrator", SenderActivationID: "fixture-sender",
		RecipientSessionID: recipient, RecipientActivationID: "fixture-activation", Text: "Fixture only",
	}, "fixture-send", session.SendFingerprint(recipient, "fixture-activation", "Fixture only"))
	must(t, err)
	_, err = store.RecordReply(recipient, "fixture-activation", "fixture-report", message.ID, "progress", fixtureReport)
	must(t, err)
}
