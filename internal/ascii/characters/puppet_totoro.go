package characters

import (
	"kankuro/internal/ascii"
)

// NewPuppetTotoro returns a TextPuppet for the totoro design.
func NewPuppetTotoro(opts ...ascii.Option) ascii.Shape {
	lines := []string{
		"  ^---^  ",
		" ( o o ) ",
		"(  vvv  )",
		" \\_____/",
	}
	return NewTextPuppet(lines, opts...)
}
