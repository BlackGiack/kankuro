package characters

import (
	"kankuro/internal/ascii"
)

// NewPuppetKankuro returns a TextPuppet for the Kankuro design.
func NewPuppetKankuro(opts ...ascii.Option) ascii.Shape {
	lines := []string{
		" /\\_/\\ ",
		"( -.- )  ",
		" >(T)<   ",
		"/|||||\\ ",
		" || ||   ",
	}
	return NewTextPuppet(lines, opts...)
}
