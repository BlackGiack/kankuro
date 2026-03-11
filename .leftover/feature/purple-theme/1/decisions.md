# Decisions: Purple/Black Theme and Status Coloring

- **Style Refactoring:** Moved color and style definitions to a new `internal/ui/styles.go` file to share them across different UI components (`app.go` and `attach.go`).
- **Nerd Font Icon:** Chose to use a puppet icon (`󰚩 `) for sessions waiting for input to provide a clearer visual indicator.
- **Dimmed Colors:** Adjusted "dimmed" description colors to `"236"` to ensure they remain subtle but legible on black backgrounds.
- **Black Theme:** Interpreted the "black and purple" request as prioritizing dark backgrounds (the natural terminal state) and using purple as the primary accent, focusing on high-contrast, status-driven coloring.
