package ui

import tea "github.com/charmbracelet/bubbletea"

func (a *App) updateAgentsEditor(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case keyCtrlC:
			return a, tea.Quit
		case keyEsc:
			a.mode = viewCardForm
			return a, nil
		case "ctrl+s":
			a.cardForm.Content = a.cardForm.Editor.Value()
			a.mode = viewCardForm
			return a, nil
		}
	}
	var cmd tea.Cmd
	a.cardForm.Editor, cmd = a.cardForm.Editor.Update(msg)
	return a, cmd
}

func (a *App) updateAgentsPathInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case keyCtrlC:
			return a, tea.Quit
		case keyEsc:
			a.mode = viewCardForm
			return a, nil
		case keyEnter:
			a.cardForm.Path = a.cardForm.PathInput.Value()
			a.mode = viewCardForm
			return a, nil
		}
	}
	var cmd tea.Cmd
	a.cardForm.PathInput, cmd = a.cardForm.PathInput.Update(msg)
	return a, cmd
}
