package ascii

type Shape interface {
	Render(canvas *Canvas, x, y int)
	Bounds() (width, height int)
}

type BaseShape struct {
	Width  int
	Height int
}

func (b *BaseShape) Bounds() (width, height int) {
	return b.Width, b.Height
}
