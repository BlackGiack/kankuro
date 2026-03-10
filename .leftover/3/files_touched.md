# Files Touched
- `internal/session/manager.go`: Added `LastWordAt`, updated output tracking, and used new `IsPromptLike` signature.
- `internal/driver/driver.go`: Refactored `IsPromptLike`, removed `CaptureOutput` and `output` field.
- `internal/session/manager_test.go`: Added `TestSessionIdleWordDetection` and updated existing tests.
- `internal/driver/driver_test.go`: Updated `TestIsPromptLike` to match new logic.
- `internal/driver/factory_test.go`: Removed `CaptureOutput` from `stubDriver`.
