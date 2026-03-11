# Work Done
- Added a pending-line cap to prevent unbounded no-newline output buffering in session tracking.
- Stopped attach output goroutine immediately on detach to reduce cross-session output bleed.
- Ensured spawned CLI processes get `PWD` set to the configured working directory, fixing workspace issues for CLIs that read PWD instead of cwd.
- Added a unit test for pending line cap.
