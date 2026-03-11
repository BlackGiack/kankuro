# Summary
Goal: Implement KANKURO MVP from `kankuro_specs.md`, then refactor for readability and introduce an `AgentDriver` factory that is actually used.
Current status: MVP implemented and compiling; config package split by subject; comments added to main functions; `NewCLIDriver(driver string) AgentDriver` factory added and wired into session launch path.
Key outcomes:
- Full Go module scaffold and runnable Bubble Tea app created.
- PTY-backed session manager and generic CLI driver implemented.
- Persistent YAML config for ID cards, skills, and MCPs implemented with validation and atomic writes.
- Driver factory + registration map implemented with generic fallback.
- Tests and build are currently green (`go test ./...`, `go build ./...`).

# Key Context
- Date/time of handover creation: 2026-03-08T02:18:25+01:00.
- Repo is still uncommitted (`git status` shows untracked project files).
- A local binary artifact exists at `/home/giack/Projects/kankuro/cmd/kankuro/kankuro` from `go build`.
