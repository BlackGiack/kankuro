# Work Done
- Modified `internal/session/manager.go` to add `LastWordAt` timestamp and track it using a new `wordRE` regex that matches alphanumeric characters.
- Refactored `updateSessionOutputState` to update `LastWordAt` whenever a chunk contains "words".
- Updated `IsPromptLike` in `internal/driver/driver.go` to mark a session as waiting if either a prompt character is seen or if `now - LastWordAt > 15s`.
- Removed `CaptureOutput` from the `AgentDriver` interface and all implementations (`GenericCLIDriver`, `stubDriver`).
- Removed the unused `output` field from `GenericCLIDriver`.
- Updated `internal/session/manager_test.go` and `internal/driver/driver_test.go` to reflect the new "word silence" logic and the removal of `CaptureOutput`.
- Verified that `LastOutputAt` was removed from `Session` and `Snapshot` structs.
- Confirmed that `go test ./...` and `go build ./...` pass.
