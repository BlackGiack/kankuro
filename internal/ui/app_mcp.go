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
			return a.handleMCPFormSubmit()
		}
	}
	cmds := make([]tea.Cmd, len(a.mcpInputs))
	for i := range a.mcpInputs {
		a.mcpInputs[i], cmds[i] = a.mcpInputs[i].Update(msg)
	}
	return a, tea.Batch(cmds...)
}

func (a *App) handleMCPFormSubmit() (tea.Model, tea.Cmd) {
	args := parseCommaList(a.mcpInputs[2].Value())
	env := parseKeyValueList(a.mcpInputs[5].Value())
	headers := parseKeyValueList(a.mcpInputs[6].Value())

	_, err := a.repo.UpsertMCP(config.MCPServer{
		Name:    strings.TrimSpace(a.mcpInputs[0].Value()),
		Command: strings.TrimSpace(a.mcpInputs[1].Value()),
		Args:    args,
		CWD:     strings.TrimSpace(a.mcpInputs[3].Value()),
		URL:     strings.TrimSpace(a.mcpInputs[4].Value()),
		Env:     env,
		Headers: headers,
	})
	if err != nil {
		a.err = err
		return a, nil
	}
	a.mcps, _ = a.repo.ListMCPs() // Refresh full list to get ID etc
	a.mode = a.mcpFormPrevMode
	return a, nil
}

func parseCommaList(input string) []string {
	parts := strings.Split(input, ",")
	var result []string
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			result = append(result, s)
		}
	}
	return result
}

func parseKeyValueList(input string) map[string]string {
	result := make(map[string]string)
	for pair := range strings.SplitSeq(input, ",") {
		if k, v, ok := strings.Cut(pair, "="); ok {
			result[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	return result
}
