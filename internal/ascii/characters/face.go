package characters

import (
	"kankuro/internal/ascii"
)

type Face struct {
	ascii.BaseShape
	style     string
	useColors bool
}

func NewFace(opts ...ascii.Option) ascii.Shape {
	config := ascii.NewConfig()
	for _, opt := range opts {
		opt(config)
	}

	shape := &Face{
		style:     config.GetString("style", "neutral"),
		useColors: config.GetBool("useColor", false),
	}
	shape.BaseShape.Width = config.GetInt("width", 5)
	shape.BaseShape.Height = config.GetInt("height", 1)
	return shape
}

func (f *Face) Bounds() (width, height int) {
	return f.BaseShape.Bounds()
}

func (f *Face) Render(canvas *ascii.Canvas, x, y int) {
	switch f.style {
	case "happy":
		canvas.DrawText(x, y, "(^-^)")
	case "focused":
		canvas.DrawText(x, y, "[-.-]")
	case "thinking":
		canvas.DrawText(x, y, "(?_?)")
	default:
		canvas.DrawText(x, y, "(o.o)")
	}
}
