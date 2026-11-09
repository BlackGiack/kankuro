package ascii

import "strings"

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

func (c *Canvas) DrawText(x, y int, text string) {
	for i, r := range text {
		c.SetCell(x+i, y, r)
	}
}

func (c *Canvas) String() string {
	var sb strings.Builder
	for y := range c.height {
		sb.WriteString(string(c.cells[y]))
		if y < c.height-1 {
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}
