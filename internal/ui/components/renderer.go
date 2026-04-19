package components

import (
	"kankuro/internal/ascii"
	"kankuro/internal/ascii/characters"
)

// RenderAgentFace renders ASCII face for a session/card.
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

// RenderAgentPuppet renders ASCII puppet for a card.
func RenderAgentPuppet(puppetStyle string) string {
	if puppetStyle == "" {
		puppetStyle = "kankuro"
	}

	puppet := characters.NewPuppet(func(c *ascii.Config) { c.Set("style", puppetStyle) })
	w, h := puppet.Bounds()
	canvas := ascii.NewCanvas(w, h)
	puppet.Render(canvas, 0, 0)
	return canvas.String()
}
