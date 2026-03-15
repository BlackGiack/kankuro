package ui

import "github.com/charmbracelet/bubbles/textinput"

func (a *App) cycleFocus(inputs []textinput.Model) {
	focused := -1
	for i := range inputs {
		if inputs[i].Focused() {
			focused = i
			inputs[i].Blur()
			break
		}
	}
	next := (focused + 1) % len(inputs)
	inputs[next].Focus()
}
