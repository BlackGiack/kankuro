package ascii

type Canvas struct {
	width  int
	height int
	cells  [][]rune
}

func NewCanvas(width, height int) *Canvas {
	cells := make([][]rune, height)
	for i := range cells {
		cells[i] = make([]rune, width)
		for j := range cells[i] {
			cells[i][j] = ' '
		}
	}
	return &Canvas{
		width:  width,
		height: height,
		cells:  cells,
	}
}

func (c *Canvas) Width() int {
	return c.width
}

func (c *Canvas) Height() int {
	return c.height
}

func (c *Canvas) SetCell(x, y int, r rune) {
	if x < 0 || x >= c.width || y < 0 || y >= c.height {
		return
	}
	c.cells[y][x] = r
}

func (c *Canvas) GetCell(x, y int) rune {
	if x < 0 || x >= c.width || y < 0 || y >= c.height {
		return ' '
	}
	return c.cells[y][x]
}

func (c *Canvas) Fill(rect Rect, r rune) {
	for y := rect.Y; y < rect.Y+rect.Height; y++ {
		for x := rect.X; x < rect.X+rect.Width; x++ {
			c.SetCell(x, y, r)
		}
	}
}

func (c *Canvas) DrawLine(x1, y1, x2, y2 int, r rune) {
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	sx := 1
	if x1 > x2 {
		sx = -1
	}
	sy := 1
	if y1 > y2 {
		sy = -1
	}
	err := dx - dy

	for {
		c.SetCell(x1, y1, r)
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}

func (c *Canvas) DrawRect(rect Rect, r rune, filled bool) {
	if filled {
		c.Fill(rect, r)
	} else {
		for x := rect.X; x < rect.X+rect.Width; x++ {
			c.SetCell(x, rect.Y, r)
			c.SetCell(x, rect.Y+rect.Height-1, r)
		}
		for y := rect.Y + 1; y < rect.Y+rect.Height-1; y++ {
			c.SetCell(rect.X, y, r)
			c.SetCell(rect.X+rect.Width-1, y, r)
		}
	}
}

func (c *Canvas) DrawText(x, y int, text string) {
	for i, r := range text {
		c.SetCell(x+i, y, r)
	}
}

func (c *Canvas) String() string {
	var result string
	for y := 0; y < c.height; y++ {
		result += string(c.cells[y])
		if y < c.height-1 {
			result += "\n"
		}
	}
	return result
}

type Rect struct {
	X, Y, Width, Height int
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
