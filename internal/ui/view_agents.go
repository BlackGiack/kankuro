package ui

func (a *App) viewAgentsEditorRender() string {
	return titleStyle.Render("Editing AGENTS.md") + "\n\n" +
		a.cardForm.Editor.View() + "\n\n" +
		"Ctrl+s=save Esc=cancel"
}

func (a *App) viewAgentsPathInputRender() string {
	return titleStyle.Render("Enter AGENTS.md Path") + "\n\n" +
		a.cardForm.PathInput.View() + "\n\n" +
		"Enter=save Esc=cancel"
}
