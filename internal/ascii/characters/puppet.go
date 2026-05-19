package characters

import (
	"kankuro/internal/ascii"
)

// BuiltinPuppetDef represents a built-in puppet definition with its name and lines.
type BuiltinPuppetDef struct {
	Name  string
	Lines []string
}

// BuiltinPuppetLines returns definitions for all built-in puppets.
// This is the single source of truth for built-in ASCII art.
func BuiltinPuppetLines() []BuiltinPuppetDef {
	return []BuiltinPuppetDef{
		{Name: "default", Lines: []string{"  ___  ", " (o o) ", "/|_=_|\\", " /   \\ "}},
		{Name: "kankuro", Lines: []string{" /\\_/\\ ", "( -.- )  ", " >(T)<   ", "/|||||\\ ", " || ||   "}},
		{Name: "cat", Lines: []string{"  /\\___/\\ ", " (  o o  )", "  |  ^  | ", "  | \\_/ | ", "   -----  "}},
		{Name: "strawhat", Lines: []string{"  _...._  ", " /  __  \\", "(  (oo)  )", " \\__--__/ ", "   /||\\  "}},
		{Name: "darth", Lines: []string{"  .---.  ", " /_|_|_\\", " / \\_/ \\", "| >[#]< |", " \\_____/ "}},
		{Name: "totoro", Lines: []string{"  ^---^  ", " ( o o ) ", "(  vvv  )", " \\_____/", ""}},
	}
}

func NewPuppet(opts ...ascii.Option) ascii.Shape {
	config := ascii.NewConfig()
	for _, opt := range opts {
		opt(config)
	}

	style := config.GetString("style", "default")
	for _, def := range BuiltinPuppetLines() {
		if def.Name == style {
			return NewTextPuppet(def.Lines, opts...)
		}
	}
	return NewTextPuppet(BuiltinPuppetLines()[0].Lines, opts...)
}
