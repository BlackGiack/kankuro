# Next Steps
1. Add dedicated driver constructors and register them.
- Example targets: `claude`, `gemini`, etc.
- Implement in `internal/driver` and register during init/startup.

2. Improve wrapped terminal behavior.
- Current output loop is line-driven from scanner/channel.
- Move toward streaming byte chunks or viewport-based PTY rendering for better UX.

3. Expand TUI CRUD coverage.
- Existing forms create/update basics.
- Add delete flows and richer editing for skills/MCP fields (`description`, `args`, `env`, etc.).

4. Strengthen test coverage.
- Add integration tests for attach/detach interaction and factory wiring through session launch.
- Add UI-model tests for key flows.

5. Optional cleanup.
- Decide whether to keep or remove build artifact: `/home/giack/Projects/kankuro/cmd/kankuro/kankuro`.

# Blockers / Open Questions
- Which concrete drivers should be first-class and how should each map card config to actual CLI args/env?
- Desired terminal fidelity level for MVP vs next iteration.
