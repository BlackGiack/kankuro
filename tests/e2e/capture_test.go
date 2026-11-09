//go:build e2e

package e2e_test

import (
	"path/filepath"
	"testing"
	"time"

	tuitest "github.com/microsoft/tui-test/bindings/go"
	"github.com/microsoft/tui-test/bindings/go/tuitesttest"
)

// This preflight makes no agent/model calls and leaves no persistent images.
func TestNativeCaptureGlyphs(t *testing.T) {
	verifyNativeCapture(t)
}

func verifyNativeCapture(t *testing.T) {
	t.Helper()
	terminal := tuitesttest.New(t, tuitesttest.Options{
		Prefix: "kankuro-capture-check", Program: "printf", Args: []string{"⬝ ⠋ ■"},
		Client: tuitest.ClientOptions{Recording: &tuitest.AutomaticRecording{Mode: tuitest.RecordingDisabled}},
		Spawn:  tuitest.SpawnOptions{Cols: new(uint16(12)), Rows: new(uint16(2)), WaitReady: new(false)},
	})
	timeout := 5 * time.Second
	must(t, terminal.WaitExit(tuitest.WaitOptions{Timeout: &timeout}))
	assertNativeExit(t, terminal)
	must(t, terminal.GetByText("⬝ ⠋ ■", tuitest.TextSelectorOptions{}).Expect(tuitest.LocatorExpectOptions{Timeout: &timeout}))
	_, err := terminal.Screenshot(filepath.Join(t.TempDir(), "glyph-check.png"), tuitest.ScreenshotOptions{})
	if err != nil {
		t.Fatalf("native screenshot preflight: %v; install an outline font covering U+2B1D (for example Noto Sans Symbols2)", err)
	}
}

// Direct Run processes expose State.Exited; ExpectExitCode tracks shell commands.
func assertNativeExit(t *testing.T, terminal *tuitest.Client) {
	t.Helper()
	state, err := terminal.State()
	must(t, err)
	if state.Exited == nil {
		t.Fatal("native process exit code was not recorded")
	}
	if *state.Exited != 0 {
		t.Fatalf("native process exit code = %d, want 0", *state.Exited)
	}
}
