# Summary
Goal: Improve session "waiting" status detection and clean up redundant driver code.
Current status: Implemented "word silence" idle detection (15s without alphanumeric characters) and removed the unused `CaptureOutput` method. All tests and builds are passing as of 2026-03-09.
Key outcomes:
- Sessions now correctly transition to `StatusWaiting` if they stop sending meaningful output (words) for 15 seconds, even if they continue to show animations (non-word characters).
- Redundant `CaptureOutput` method and the `output` attribute in `GenericCLIDriver` have been removed to simplify the driver interface.
- `LastOutputAt` was removed from the session manager as it became redundant with `LastWordAt` and `now`.
