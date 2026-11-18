package views_test

import (
	"strings"

	"github.com/BlackGiack/kankuro/internal/config"
)

const borderCols = 2

// userPuppet builds a user puppet whose art is w columns wide and h lines
// tall, for testing puppets larger than the box they are drawn into.
func userPuppet(w, h int) config.Puppet {
	lines := make([]string, h)
	for i := range lines {
		switch {
		case w <= 0:
		case i == 0 || i == h-1:
			lines[i] = "<" + strings.Repeat("=", max(0, w-borderCols)) + ">"
		default:
			lines[i] = "|" + strings.Repeat("#", max(0, w-borderCols)) + "|"
		}
	}
	return config.Puppet{ID: "user-puppet", Name: "user-puppet", Lines: lines}
}
