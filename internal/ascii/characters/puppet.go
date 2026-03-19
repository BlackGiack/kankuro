package characters

import (
	"kankuro/internal/ascii"
)

const defaultPuppetWidth = 7
const defaultPuppetHeight = 4

// Puppet renders a simple ASCII representation of one of Kankuro's puppets.
// It occupies a 7x4 character area by default.
type Puppet struct {
	ascii.BaseShape

	style     string
	useColors bool
}

func NewPuppet(opts ...ascii.Option) ascii.Shape {
	config := ascii.NewConfig()
	for _, opt := range opts {
		opt(config)
	}

	p := &Puppet{
		style:     config.GetString("style", "default"),
		useColors: config.GetBool("useColor", false),
	}
	p.BaseShape.Width = config.GetInt("width", defaultPuppetWidth)
	p.BaseShape.Height = config.GetInt("height", defaultPuppetHeight)
	return p
}

func (p *Puppet) Bounds() (int, int) {
	return p.BaseShape.Bounds()
}

func (p *Puppet) Render(canvas *ascii.Canvas, x, y int) {
	// Four lines, each 7 characters wide
	canvas.DrawText(x, y+0, "  ___  ")
	canvas.DrawText(x, y+1, " (o o) ")
	canvas.DrawText(x, y+2, "/|_=_|\\")
	canvas.DrawText(x, y+3, " /   \\ ")
}
