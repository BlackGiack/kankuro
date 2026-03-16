package ui

import tea "github.com/charmbracelet/bubbletea"

func (a *App) updateAgentsEditor(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "ctrl+c":
			return a, tea.Quit
		case keyEsc:
			a.mode = viewCardForm
			return a, nil
		case "ctrl+s":
			a.agentsMD = a.agentsEditor.Value()
			a.mode = viewCardForm
			return a, nil
		}
	}
	var cmd tea.Cmd
	a.agentsEditor, cmd = a.agentsEditor.Update(msg)
	return a, cmd
}

func (a *App) updateAgentsPathInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "ctrl+c":
			return a, tea.Quit
		case keyEsc:
			a.mode = viewCardForm
			return a, nil
		case keyEnter:
			a.agentsMDPath = a.agentsPathInput.Value()
			a.mode = viewCardForm
			return a, nil
		}
	}
	var cmd tea.Cmd
	a.agentsPathInput, cmd = a.agentsPathInput.Update(msg)
	return a, cmd
}
