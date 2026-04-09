package ui

import "fmt"

func (a *App) viewSwitchAgentRender() string {
	header := "Switch Agent"
	if a.switchSessionID != "" {
		header = fmt.Sprintf("Switch Agent (session %s)", a.switchSessionID)
	}
	return titleStyle.Render(header) + "\n\n" +
		a.switchList.View() + "\n" +
		"Enter=switch Esc=cancel"
}
