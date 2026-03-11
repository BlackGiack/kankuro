# Summary: Purple/Black Theme and Status Coloring

I implemented a new visual theme for KANKURO, transitioning to a purple and black aesthetic with status-aware session indicators.

## Key Changes:
- **Global Palette:** Defined a new color scheme in `internal/ui/styles.go` (`purple`, `hotPink`, `darkPurple`, `gray`).
- **Themed UI:** Updated the dashboard borders, titles, and list delegates to use the new palette.
- **Status Indicators:** Sessions in `StatusWaiting` (waiting for input) now "pulse" with a hot pink color and a puppet icon (`󰚩 `) in the sidebar.
- **PTY Styling:** The attachment header message now matches the theme.
