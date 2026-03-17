package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"kankuro/internal/config"
)

func (a *App) updateMCPForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case keyCtrlC:
			return a, tea.Quit
		case keyEsc:
			a.mode = viewDashboard
			return a, nil
		case keyTab:
			a.cycleFocus(a.mcpInputs)
			return a, nil
		case keyEnter:
			mcp, err := a.repo.UpsertMCP(config.MCPServer{
				Name:    strings.TrimSpace(a.mcpInputs[0].Value()),
				Command: strings.TrimSpace(a.mcpInputs[1].Value()),
			})
			if err != nil {
				a.err = err
				return a, nil
			}
			a.mcps = append(a.mcps, mcp)
			a.mode = viewDashboard
			return a, nil
		}
	}
	cmds := make([]tea.Cmd, len(a.mcpInputs))
	for i := range a.mcpInputs {
		a.mcpInputs[i], cmds[i] = a.mcpInputs[i].Update(msg)
	}
	return a, tea.Batch(cmds...)
}
