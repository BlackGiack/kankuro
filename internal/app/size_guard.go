package app

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/BlackGiack/kankuro/internal/ui/components"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

// The views are laid out for a terminal of at least views.MinTerminalWidth ×
// views.MinTerminalHeight. Below that they have no defined shape (the dashboard's
// preview pane would get a negative width, the card editor cannot fit), so the
// coordinator replaces whatever the current view would draw with a notice, and
// drops keys so that a keypress aimed at a screen the user cannot see does not
// launch or delete anything. Everything that is not a keypress keeps flowing,
// so the view behind the notice is current the moment the terminal grows.

// tooSmall reports whether the terminal is below the minimum. A size of 0×0 is
// "no WindowSizeMsg yet", not a tiny terminal.
func (c *Coordinator) tooSmall() bool {
	if c.width == 0 && c.height == 0 {
		return false
	}
	return c.width < views.MinTerminalWidth || c.height < views.MinTerminalHeight
}

// guardKey decides what a message does while the terminal is too small. It
// returns handled=true for input it consumed: quit still quits (it is the
// user's way out), every other keypress or paste is dropped.
func (c *Coordinator) guardKey(msg tea.Msg) (bool, tea.Cmd) {
	switch m := msg.(type) {
	case tea.KeyPressMsg:
		if s := m.String(); s == "q" || s == "ctrl+c" {
			return true, tea.Quit
		}
		return true, nil
	case tea.PasteMsg:
		return true, nil
	}
	return false, nil
}

// tooSmallView draws the notice, centred in exactly the terminal's size. When
// the terminal is narrower than the single-line text it is split in two, and
// each line is clipped, so the notice itself never wraps.
func (c *Coordinator) tooSmallView() string {
	text := fmt.Sprintf("Terminal too small: need %d×%d, have %d×%d",
		views.MinTerminalWidth, views.MinTerminalHeight, c.width, c.height)
	lines := []string{text}
	if lipgloss.Width(text) > c.width {
		head, tail, _ := strings.Cut(text, ": ")
		lines = []string{head, tail}
	}
	if len(lines) > c.height {
		lines = lines[:max(c.height, 0)]
	}
	for i, l := range lines {
		lines[i] = components.TruncateHead(l, c.width)
	}
	// JoinVertical pads the lines to one width: Place leaves a block as it is
	// when it is already as wide as the terminal, ragged lines included.
	return lipgloss.Place(c.width, c.height, lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center, lines...))
}
