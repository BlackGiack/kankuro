package characters

import (
	"kankuro/internal/ascii"
)

// NewPuppetStrawhat returns a TextPuppet for the straw-hat design.
func NewPuppetStrawhat(opts ...ascii.Option) ascii.Shape {
	lines := []string{
		"  ______  ",
		" /      \\ ",
		"/ (o_o)  \\",
		"|  \\_/   |",
		"  / | \\   ",
	}
	return NewTextPuppet(lines, opts...)
}
