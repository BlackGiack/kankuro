# Work Done: Purple/Black Theme and Status Coloring

- [x] Defined global colors and styles in `internal/ui/styles.go`.
- [x] Refactored `internal/ui/app.go` to remove inline color codes and use the new styles.
- [x] Customized `list.DefaultDelegate` in `New()` to ensure consistent theme application across ID Cards and Sessions.
- [x] Updated the `item` struct to include a `status` field, allowing status-based title and description rendering.
- [x] Implemented a "waiting" visual state for session items using a hot pink color and a puppet icon (`󰚩 `).
- [x] Styled the PTY attachment message in `internal/ui/attach.go` with the new theme colors.
- [x] Verified the changes by rebuilding the project.
