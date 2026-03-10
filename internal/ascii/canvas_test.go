package ascii

import (
	"testing"
)

func TestCanvasCreation(t *testing.T) {
	c := NewCanvas(10, 5)
	if c.Width() != 10 {
		t.Errorf("expected width 10, got %d", c.Width())
	}
	if c.Height() != 5 {
		t.Errorf("expected height 5, got %d", c.Height())
	}
}

func TestCanvasSetCell(t *testing.T) {
	c := NewCanvas(5, 3)
	c.SetCell(2, 1, 'X')
	if r := c.GetCell(2, 1); r != 'X' {
		t.Errorf("expected 'X', got %c", r)
	}
}

func TestCanvasDrawText(t *testing.T) {
	c := NewCanvas(10, 2)
	c.DrawText(0, 0, "Hello")
	expected := "Hello     \n          "
	if result := c.String(); result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestCanvasFill(t *testing.T) {
	c := NewCanvas(5, 3)
	c.Fill(Rect{0, 0, 5, 3}, '#')
	expected := "#####\n#####\n#####"
	if result := c.String(); result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}
