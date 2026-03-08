# Files Touched
- `/home/giack/Projects/kankuro/internal/driver/driver.go`
  Change: Changed the driver output model from line strings to raw byte chunks and added session completion signaling.

- `/home/giack/Projects/kankuro/internal/session/manager.go`
  Change: Added scrollback buffering, PTY output subscriber fanout, attach handles, safe PTY exposure for resize sync, and session removal support.

- `/home/giack/Projects/kankuro/internal/session/manager_test.go`
  Change: Added coverage for attach output, scrollback replay, prompt detection without newline, and terminated-session removal.

- `/home/giack/Projects/kankuro/internal/ui/app.go`
  Change: Replaced the old terminal view flow with attach dispatch, added dashboard pane focus, and exposed session kill/remove actions.

- `/home/giack/Projects/kankuro/internal/ui/attach.go`
  Change: Added Bubble Tea `Exec`-based raw PTY attach, `Ctrl+K` detach interception, scrollback replay, live output forwarding, and PTY resize sync.

- `/home/giack/Projects/kankuro/internal/driver/factory_test.go`
  Change: Updated the stub driver to match the byte-stream `CaptureOutput` API.

- `/home/giack/Projects/kankuro/kankuro_specs.md`
  Change: Updated the specs to reflect raw PTY attach, byte-stream capture, resize behavior, dashboard focus switching, and session kill/remove keys.

- `/home/giack/Projects/kankuro/.leftover/2/summary.md`
  Change: Current-session handover summary.

- `/home/giack/Projects/kankuro/.leftover/2/work_done.md`
  Change: Concrete record of PTY/session/navigation work completed in this session.

- `/home/giack/Projects/kankuro/.leftover/2/decisions.md`
  Change: Decision log for the raw attach, fanout, resize, and session-cleanup choices.

- `/home/giack/Projects/kankuro/.leftover/2/next_steps.md`
  Change: Remaining follow-up work and open questions.

- `/home/giack/Projects/kankuro/.leftover/2/commands.md`
  Change: Command log and validation status for this session.
