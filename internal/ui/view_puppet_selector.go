package ui

import (
	"kankuro/internal/ascii"
	"kankuro/internal/ascii/characters"
)

func (a *App) viewPuppetSelectorRender() string {
	title := "Select Puppet"
	selected := puppetStyles()[a.cardForm.PuppetSelectorIndex]
	p := characters.NewPuppet(func(c *ascii.Config) { c.Set("style", selected) })
	w, h := p.Bounds()
	canvas := ascii.NewCanvas(w, h)
	p.Render(canvas, 0, 0)
	preview := "\n" + canvas.String() + "\n"
	return titleStyle.Render(title) + "\n\n" +
		"Use Left/Right to change puppet, Enter=confirm, Esc=cancel\n\n" +
		"Selected: " + selected + preview
}
