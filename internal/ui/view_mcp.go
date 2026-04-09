package ui

func (a *App) viewMCPFormRender() string {
	return titleStyle.Render("Add MCP") + "\n\n" +
		"Name: " + a.mcpInputs[0].View() + "\n" +
		"Command: " + a.mcpInputs[1].View() + "\n" +
		"Args: " + a.mcpInputs[2].View() + "\n" +
		"CWD: " + a.mcpInputs[3].View() + "\n" +
		"URL: " + a.mcpInputs[4].View() + "\n" +
		"Env: " + a.mcpInputs[5].View() + "\n" +
		"Headers: " + a.mcpInputs[6].View() + "\n\n" +
		"Enter=save Tab=next Esc=cancel"
}
