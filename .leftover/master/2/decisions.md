# Decisions
- Decision: Use true terminal takeover for session attach instead of an embedded text-input terminal view.
  Rationale: Interactive CLIs need raw PTY semantics, live prompts, and full-screen behavior that the previous line-buffered Bubble Tea view could not provide.
  Tradeoffs: Attach temporarily leaves the dashboard UI and depends on terminal handoff logic.

- Decision: Keep one PTY reader per session and fan out byte chunks internally.
  Rationale: This avoids races where status tracking and the UI consume the same output stream and miss data.
  Tradeoffs: The session manager now owns more buffering/subscriber state.

- Decision: Replay bounded raw scrollback when attaching to an existing session.
  Rationale: Reattach should show recent context before live output resumes.
  Tradeoffs: Scrollback is bounded (currently 256 KiB), so very old output is dropped.

- Decision: Resize the session PTY from the real terminal during attach and on `SIGWINCH`.
  Rationale: Attached sessions should match the user terminal dimensions and respond to resizes.
  Tradeoffs: Resize handling is attach-time behavior; background sessions are not continuously resized when detached.

- Decision: Expose explicit session cleanup actions instead of auto-removing terminated sessions.
  Rationale: Users may want to inspect terminated sessions before removing them.
  Tradeoffs: The sidebar can accumulate terminated sessions until the user deletes them.
