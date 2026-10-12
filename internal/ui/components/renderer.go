package components

import (
	"github.com/BlackGiack/kankuro/internal/ascii"
	"github.com/BlackGiack/kankuro/internal/ascii/characters"
	"github.com/BlackGiack/kankuro/internal/config"
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

func RenderAgentPuppet(puppetName string, puppets []config.Puppet) string {
	if puppetName == "" {
		puppetName = "kankuro"
	}

	for _, p := range puppets {
		if p.Name == puppetName {
			if p.Builtin {
				puppet := characters.NewPuppet(func(c *ascii.Config) { c.Set("style", puppetName) })
				w, h := puppet.Bounds()
				canvas := ascii.NewCanvas(w, h)
				puppet.Render(canvas, 0, 0)
				return canvas.String()
			}
			canvas := ascii.NewCanvas(maxLineWidth(p.Lines), len(p.Lines))
			for i, line := range p.Lines {
				canvas.DrawText(0, i, line)
			}
			return canvas.String()
		}
	}

	puppet := characters.NewPuppet(func(c *ascii.Config) { c.Set("style", "kankuro") })
	w, h := puppet.Bounds()
	canvas := ascii.NewCanvas(w, h)
	puppet.Render(canvas, 0, 0)
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
