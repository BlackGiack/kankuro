# Commands
- Command: `sed -n '1,260p' internal/driver/driver.go`, `sed -n '1,260p' internal/session/manager.go`, `sed -n '1,260p' internal/ui/app.go`, `sed -n '1,220p' ptyExample.goex`
  Purpose: Inspect the original MVP PTY/session/UI design and compare it with the proposed raw PTY example.
  Notable output: Confirmed the old design was line-buffered, single-consumer, and not a true terminal wrapper.

- Command: `rg -n "ReleaseTerminal|RestoreTerminal|ExecProcess|Suspend" /home/giack/go/pkg/mod/github.com/charmbracelet/bubbletea@v1.3.10 -g '*.go'`
  Purpose: Verify Bubble Tea terminal handoff support before implementing raw attach.
  Notable output: Confirmed `Program.ReleaseTerminal`, `Program.RestoreTerminal`, and `tea.Exec` support.

- Command: `gofmt -w ...`, `go test ./...`, `go build ./...`
  Purpose: Format and validate each code change set.
  Notable output: Final validation passed. One intermediate failed test exposed partial-line prompt handling, which led to a parser fix.

- Command: `rg -n "InheritSize|Setsize|GetsizeFull|SIGWINCH" $(go env GOPATH)/pkg/mod/github.com/creack/pty* -g '*.go'`
  Purpose: Check PTY resize support before fixing terminal sizing.
  Notable output: Confirmed `GetsizeFull` and `Setsize` APIs, used for attach-time PTY resize sync.

- Command: `python /home/giack/.codex/skills/handover/scripts/init_handover.py --root /home/giack/Projects/kankuro`
  Purpose: Create the next handover package.
  Notable output: Created `/home/giack/Projects/kankuro/.leftover/2`.

# Important outputs / side effects
- `go test ./...` passed on 2026-03-09 after the PTY/session/navigation changes.
- `go build ./...` passed on 2026-03-09 after the PTY/session/navigation changes.
