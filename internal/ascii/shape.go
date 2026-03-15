package ascii

type Shape interface {
	Render(canvas *Canvas, x, y int)
	Bounds() (int, int)
}

type BaseShape struct {
	Width  int
	Height int
}

func (b *BaseShape) Bounds() (int, int) {
	return b.Width, b.Height
}
