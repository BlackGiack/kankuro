package ui

import "fmt"

func (a *App) viewSwitchAgentRender() string {
	header := "Switch Agent"
	if a.sessionView.SwitchID != "" {
		header = fmt.Sprintf("Switch Agent (session %s)", a.sessionView.SwitchID)
	}
	return titleStyle.Render(header) + "\n\n" +
		a.switchList.View() + "\n" +
		"Enter=switch Esc=cancel"
}
