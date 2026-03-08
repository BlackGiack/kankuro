# Work Done
- Reworked PTY/session handling so the session layer owns the only PTY read stream and distributes output to status tracking plus attach subscribers.
- Added raw attach execution in the UI using Bubble Tea `Exec`, raw terminal mode, scrollback replay, `Ctrl+K` detach interception, and live PTY output forwarding.
- Added PTY resize sync at attach start and on terminal resize signals.
- Added dashboard pane focus state so the user can move between ID cards and active sessions with `Tab`/`Shift+Tab`/arrow keys.
- Added session actions from the sidebar: `x` to kill a running session and `Delete`/`Backspace` to remove a terminated session.
- Added tests for attach output, scrollback replay, prompt detection without newline, and removal of terminated sessions only.
- Updated the specs document to match the byte-stream driver API, raw PTY attach flow, resize behavior, and session management keys.

# Notes
- The UI still has no dedicated Bubble Tea model tests; validation is currently via package tests plus `go build ./...`.
- The repo contains a top-level `kankuro` artifact in `git status`; it was not touched in this session.
