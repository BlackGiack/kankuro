# Summary
Goal: Stabilize the KANKURO MVP terminal/session workflow after MVP testing, focusing on PTY attach fidelity, terminal sizing, dashboard navigation, and session lifecycle controls.
Current status: The app now uses raw PTY attach with Bubble Tea terminal release/restore, bounded scrollback replay, live PTY resize sync on attach, dashboard pane switching, and explicit kill/remove actions for sessions. `go test ./...` and `go build ./...` are passing as of 2026-03-09T00:18:39+01:00.
Key outcomes:
- Replaced the old line-based "wrapped terminal" with a raw attach path that detaches on `Ctrl+K` without killing the child process.
- Fixed missing output/prompt issues by moving to a single PTY reader with subscriber fanout and scrollback replay.
- Synced attached PTY size with the user terminal and updated it on `SIGWINCH`.
- Added dashboard focus switching between ID cards and sessions, plus kill/remove commands for sessions.
- Updated `kankuro_specs.md` to reflect the current behavior as the implementation changed.

# Key Context
- Handover created on 2026-03-09T00:18:39+01:00.
- Repo is still largely uncommitted; `git status --short` shows `kankuro_specs.md` modified and project files under `cmd/`, `internal/`, `go.mod`, `go.sum`, `.leftover/`, `.gitignore`, and `kankuro` as untracked.
- Latest handover folder is `/home/giack/Projects/kankuro/.leftover/2`.
