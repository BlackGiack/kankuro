package characters_test

import (
	"kankuro/internal/ascii"
	"kankuro/internal/ascii/characters"
	"testing"
)

func TestFaceDefaultStyle(t *testing.T) {
	face := characters.NewFace()
	width, height := face.Bounds()
	if width != 5 {
		t.Errorf("expected width 5, got %d", width)
	}
	if height != 1 {
		t.Errorf("expected height 1, got %d", height)
	}
}

func TestFaceNeutralStyle(t *testing.T) {
	face := characters.NewFace(ascii.WithStyle("neutral"))
	canvas := ascii.NewCanvas(5, 1)
	face.Render(canvas, 0, 0)
	expected := "(o.o)"
	if result := canvas.String(); result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFaceHappyStyle(t *testing.T) {
	face := characters.NewFace(ascii.WithStyle("happy"))
	canvas := ascii.NewCanvas(5, 1)
	face.Render(canvas, 0, 0)
	expected := "(^-^)"
	if result := canvas.String(); result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFaceFocusedStyle(t *testing.T) {
	face := characters.NewFace(ascii.WithStyle("focused"))
	canvas := ascii.NewCanvas(5, 1)
	face.Render(canvas, 0, 0)
	expected := "[-.-]"
	if result := canvas.String(); result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFaceThinkingStyle(t *testing.T) {
	face := characters.NewFace(ascii.WithStyle("thinking"))
	canvas := ascii.NewCanvas(5, 1)
	face.Render(canvas, 0, 0)
	expected := "(?_?)"
	if result := canvas.String(); result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFaceCustomWidth(t *testing.T) {
	face := characters.NewFace(ascii.WithWidth(7))
	width, _ := face.Bounds()
	if width != 7 {
		t.Errorf("expected width 7, got %d", width)
	}
}
