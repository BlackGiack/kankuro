package ascii_test

import (
	"strings"
	"testing"

	"github.com/BlackGiack/kankuro/internal/ascii"
)

func TestCanvasCreation(t *testing.T) {
	c := ascii.NewCanvas(10, 5)
	want := strings.TrimSuffix(strings.Repeat(strings.Repeat(" ", 10)+"\n", 5), "\n")
	if got := c.String(); got != want {
		t.Fatalf("blank canvas = %q, want %q", got, want)
	}
}

func TestCanvasSetCell(t *testing.T) {
	c := ascii.NewCanvas(5, 3)
	c.SetCell(2, 1, 'X')
	if r := c.GetCell(2, 1); r != 'X' {
		t.Errorf("expected 'X', got %c", r)
	}
}

func TestCanvasDrawText(t *testing.T) {
	c := ascii.NewCanvas(10, 2)
	c.DrawText(0, 0, "Hello")
	expected := "Hello     \n          "
	if result := c.String(); result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}
