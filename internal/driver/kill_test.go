package driver_test

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/BlackGiack/kankuro/internal/driver"
)

var pidLineRE = regexp.MustCompile(`(\d+)\r?\n`)

// spawnScript runs script under sh in a PTY, drains its output into the
// returned channel, and registers a cleanup that SIGKILLs the process group
// whatever the test did - a failing kill test must not leave processes (a
// `sleep 1000`, say) running on the developer's machine.
func spawnScript(t *testing.T, script string, env map[string]string) (*driver.GenericCLIDriver, *driver.SessionHandle, <-chan string) {
	t.Helper()
	d := driver.NewGenericCLIDriver()
	h, err := d.Spawn(driver.AgentConfig{
		Driver:     "sh",
		Args:       []string{"-c", script},
		WorkingDir: t.TempDir(),
		Env:        env,
	})
	if err != nil {
		t.Fatalf("spawn: %v", err)
	}
	pid := h.Cmd.Process.Pid
	t.Cleanup(func() { _ = syscall.Kill(-pid, syscall.SIGKILL) })

	out := make(chan string, 64)
	go func() {
		for chunk := range h.Output {
			select {
			case out <- string(chunk):
			default:
			}
		}
	}()
	return d, h, out
}

// waitForMatch accumulates output until pred accepts it.
func waitForMatch(t *testing.T, out <-chan string, what string, pred func(string) bool) string {
	t.Helper()
	var acc strings.Builder
	timeout := time.After(5 * time.Second)
	for {
		select {
		case chunk := <-out:
			acc.WriteString(chunk)
			if pred(acc.String()) {
				return acc.String()
			}
		case <-timeout:
			t.Fatalf("timed out waiting for %s; output so far: %q", what, acc.String())
		}
	}
}

func waitForText(t *testing.T, out <-chan string, text string) {
	t.Helper()
	waitForMatch(t, out, text, func(s string) bool { return strings.Contains(s, text) })
}

// processGone reports whether pid no longer runs. A zombie counts as gone: in
// a container whose init does not reap orphans, an exited grandchild stays a
// zombie forever and kill(pid, 0) keeps succeeding.
func processGone(pid int) bool {
	if err := syscall.Kill(pid, 0); errors.Is(err, syscall.ESRCH) {
		return true
	}
	stat, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/stat")
	if err != nil {
		return false
	}
	// "<pid> (<comm>) <state> ..."; comm may contain spaces, so cut at the last ')'.
	s := string(stat)
	if i := strings.LastIndexByte(s, ')'); i >= 0 && i+2 < len(s) {
		return s[i+2] == 'Z'
	}
	return false
}

// TestKillTerminatesTheProcessGroup: a descendant that ignores SIGHUP used to
// survive Kill (only the direct child was signalled, and the SIGHUP from the
// closing PTY does not reach a process ignoring it).
func TestKillTerminatesTheProcessGroup(t *testing.T) {
	// The trap is set before the background job is forked so the job inherits
	// the ignore: no window in which it could still die of the PTY's SIGHUP.
	d, _, out := spawnScript(t, `trap "" HUP; sleep 1000 & echo $!; wait`, nil)

	got := waitForMatch(t, out, "the grandchild's pid", func(s string) bool { return pidLineRE.MatchString(s) })
	gpid, err := strconv.Atoi(pidLineRE.FindStringSubmatch(got)[1])
	if err != nil {
		t.Fatalf("parse pid: %v", err)
	}
	t.Cleanup(func() { _ = syscall.Kill(gpid, syscall.SIGKILL) })
	if processGone(gpid) {
		t.Fatalf("grandchild %d is not running before Kill", gpid)
	}

	if errKill := d.Kill(); errKill != nil {
		t.Fatalf("kill: %v", errKill)
	}

	deadline := time.Now().Add(5 * time.Second)
	for !processGone(gpid) {
		if time.Now().After(deadline) {
			t.Fatalf("grandchild %d survived Kill", gpid)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// TestKillIsGraceful: SIGTERM comes first, so the agent can run its own
// shutdown. A bare SIGKILL gives the trap no chance.
func TestKillIsGraceful(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "bye")
	d, _, out := spawnScript(t,
		`trap 'echo bye > "$F"; exit 0' TERM; echo ready; while :; do sleep 0.1; done`,
		map[string]string{"F": marker})
	waitForText(t, out, "ready")

	if err := d.Kill(); err != nil {
		t.Fatalf("kill: %v", err)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("the TERM handler did not run: %v", err)
	}
}

// TestKillEscalatesWhenTERMIsIgnored: an agent that ignores SIGTERM must still
// die, after the grace period and not much later.
func TestKillEscalatesWhenTERMIsIgnored(t *testing.T) {
	d, h, out := spawnScript(t, `trap "" TERM; echo ready; while :; do sleep 0.1; done`, nil)
	d.SetKillGraceForTest(200 * time.Millisecond)
	waitForText(t, out, "ready")
	pid := h.Cmd.Process.Pid

	start := time.Now()
	if err := d.Kill(); err != nil {
		t.Fatalf("kill: %v", err)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("Kill took %v, want about the 200ms grace period", elapsed)
	}
	if !processGone(pid) {
		t.Fatalf("process %d survived Kill", pid)
	}
	if d.GetStatus() != driver.StatusTerminated {
		t.Fatalf("status = %s, want terminated", d.GetStatus())
	}
}

// TestKillReachesDescendantsThatIgnoreTERM: the leader exits on SIGTERM (the
// normal case), but a child it started ignores SIGTERM and SIGHUP. The SIGKILL
// escalation must depend on the group still having members, not on the leader
// still running, or such a child is orphaned and lives on.
func TestKillReachesDescendantsThatIgnoreTERM(t *testing.T) {
	// The child inherits the ignored TERM and HUP; the shell then restores the
	// default TERM action for itself and only then announces the pid, so Kill
	// cannot run before the shell is ready to die of it.
	d, _, out := spawnScript(t, `trap "" HUP TERM; sleep 1000 & P=$!; trap - TERM; echo $P; wait`, nil)
	d.SetKillGraceForTest(300 * time.Millisecond)

	got := waitForMatch(t, out, "the descendant's pid", func(s string) bool { return pidLineRE.MatchString(s) })
	gpid, err := strconv.Atoi(pidLineRE.FindStringSubmatch(got)[1])
	if err != nil {
		t.Fatalf("parse pid: %v", err)
	}
	t.Cleanup(func() { _ = syscall.Kill(gpid, syscall.SIGKILL) })
	if processGone(gpid) {
		t.Fatalf("descendant %d is not running before Kill", gpid)
	}

	if errKill := d.Kill(); errKill != nil {
		t.Fatalf("kill: %v", errKill)
	}

	// No assertion on how long Kill took: where init does not reap orphans an
	// exited child stays a zombie and keeps the group "alive", so Kill then
	// legitimately spends the whole grace period.
	deadline := time.Now().Add(5 * time.Second)
	for !processGone(gpid) {
		if time.Now().After(deadline) {
			t.Fatalf("descendant %d (ignoring TERM and HUP) survived Kill", gpid)
		}
		time.Sleep(20 * time.Millisecond)
	}
}
