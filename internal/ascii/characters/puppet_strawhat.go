package characters

import (
	"kankuro/internal/ascii"
)

// NewPuppetStrawhat returns a TextPuppet for the straw-hat design.
func NewPuppetStrawhat(opts ...ascii.Option) ascii.Shape {
	lines := []string{
		"  _...._  ",
		" /  __  \\",
		"(  (oo)  )",
		" \\__--__/ ",
		"   /||\\  ",
	}
	return NewTextPuppet(lines, opts...)
}
