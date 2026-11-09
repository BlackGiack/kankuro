package characters

import (
	"github.com/BlackGiack/kankuro/internal/ascii"
)

// TextPuppet is a generic ASCII puppet backed by fixed lines of text.
// It centralizes rendering logic to avoid duplicated constructors and types.
type TextPuppet struct {
	ascii.BaseShape

	lines []string
}

// NewTextPuppet creates a TextPuppet sized to its first line and line count.
func NewTextPuppet(lines []string) ascii.Shape {
	p := &TextPuppet{lines: lines}

	if len(lines) > 0 {
		p.BaseShape.Width = len(lines[0])
	}
	p.BaseShape.Height = len(lines)

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
