package characters

import (
	"kankuro/internal/ascii"
)

// NewPuppetDarth returns a TextPuppet for the darth vader design.
func NewPuppetDarth(opts ...ascii.Option) ascii.Shape {
	lines := []string{
		"  .---.  ",
		" /_|_|_\\",
		" / \\_/ \\",
		"| >[#]< |",
		" \\_____/ ",
	}
	return NewTextPuppet(lines, opts...)
}
