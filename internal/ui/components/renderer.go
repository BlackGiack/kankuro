package components

import (
	"github.com/BlackGiack/kankuro/internal/ascii"
	"github.com/BlackGiack/kankuro/internal/ascii/characters"
)

func RenderAgentFace(faceStyle string) string {
	if faceStyle == "" {
		faceStyle = "neutral"
	}

	face := characters.NewFace(ascii.WithStyle(faceStyle))
	w, _ := face.Bounds()
	canvas := ascii.NewCanvas(w, 1)
	face.Render(canvas, 0, 0)
	return canvas.String()
}

// RenderPuppetArtwork renders focused artwork data without a registry type.
func RenderPuppetArtwork(puppetName string, builtin bool, lines []string) string {
	if puppetName == "" {
		puppetName, builtin = "kankuro", true
	}
	if builtin {
		puppet := characters.NewPuppet(func(c *ascii.Config) { c.Set("style", puppetName) })
		w, h := puppet.Bounds()
		canvas := ascii.NewCanvas(w, h)
		puppet.Render(canvas, 0, 0)
		return canvas.String()
	}
	canvas := ascii.NewCanvas(maxLineWidth(lines), len(lines))
	for i, line := range lines {
		canvas.DrawText(0, i, line)
	}
	return canvas.String()
}

func maxLineWidth(lines []string) int {
	w := 0
	for _, l := range lines {
		if len(l) > w {
			w = len(l)
		}
	}
	return w
}
