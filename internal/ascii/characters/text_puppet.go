package characters

import (
	"github.com/BlackGiack/kankuro/internal/ascii"
)

// TextPuppet is a generic ASCII puppet backed by fixed lines of text.
// It centralizes rendering logic to avoid duplicated constructors and types.
type TextPuppet struct {
	ascii.BaseShape

	lines     []string
	style     string
	useColors bool
}

// NewTextPuppet creates a TextPuppet from the provided lines. Width/Height can be
// overridden via ascii.Option; otherwise they default to the first line length and
// number of lines.
func NewTextPuppet(lines []string, opts ...ascii.Option) ascii.Shape {
	config := ascii.NewConfig()
	for _, opt := range opts {
		opt(config)
	}

	p := &TextPuppet{
		lines:     lines,
		style:     config.GetString("style", "default"),
		useColors: config.GetBool("useColor", false),
	}

	if l := config.GetInt("width", 0); l != 0 {
		p.BaseShape.Width = l
	} else if len(lines) > 0 {
		p.BaseShape.Width = len(lines[0])
	}

	if h := config.GetInt("height", 0); h != 0 {
		p.BaseShape.Height = h
	} else {
		p.BaseShape.Height = len(lines)
	}

	return p
}

func (p *TextPuppet) Bounds() (int, int) {
	return p.BaseShape.Bounds()
}

func (p *TextPuppet) Render(canvas *ascii.Canvas, x, y int) {
	for i, line := range p.lines {
		canvas.DrawText(x, y+i, line)
	}
}
