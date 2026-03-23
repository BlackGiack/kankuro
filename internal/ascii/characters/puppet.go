package characters

import (
	"kankuro/internal/ascii"
)

func NewPuppet(opts ...ascii.Option) ascii.Shape {
	config := ascii.NewConfig()
	for _, opt := range opts {
		opt(config)
	}

	style := config.GetString("style", "default")
	switch style {
	case "cat":
		return NewPuppetCat(opts...)
	case "strawhat":
		return NewPuppetStrawhat(opts...)
	case "kankuro":
		return NewPuppetKankuro(opts...)
	case "darth":
		return NewPuppetDarth(opts...)
	case "totoro":
		return NewPuppetTotoro(opts...)
	default:
		return NewBasePuppet(opts...)
	}
}

// NewBasePuppet returns a TextPuppet for the cat design.
func NewBasePuppet(opts ...ascii.Option) ascii.Shape {
	lines := []string{
		"  ___  ",
		" (o o) ",
		"/|_=_|\\",
		" /   \\ ",
	}
	return NewTextPuppet(lines, opts...)
}
