# Commands
- Command: `go test ./internal/session/...`
  Purpose: Verify new idle detection logic.
  Notable output: Passed.

- Command: `go test ./internal/driver/...`
  Purpose: Verify `IsPromptLike` changes.
  Notable output: Initially failed, then passed after updating `driver_test.go`.

- Command: `go test ./... && go build ./...`
  Purpose: Full project validation.
  Notable output: All tests passed, build successful.
