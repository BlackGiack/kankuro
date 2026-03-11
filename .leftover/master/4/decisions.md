# Decisions
- Cap `PendingLine` at 8KB while keeping scrollback at 256KB to stabilize heavy, no-newline output without losing normal history.
- Stop writing to terminal immediately on detach (drop any remaining buffered output from the previous session to avoid bleed).
- Override `PWD` in the child process environment to the configured working directory to align CLIs that rely on PWD.
