package characters

import (
	"kankuro/internal/ascii"
)

// NewPuppetCat returns a TextPuppet for the cat design.
func NewPuppetCat(opts ...ascii.Option) ascii.Shape {
	lines := []string{
		"  /\\___/\\ ",
		" (  o o  )",
		"  |  ^  | ",
		"  | \\_/ | ",
		"   -----  ",
	}
	return NewTextPuppet(lines, opts...)
}
