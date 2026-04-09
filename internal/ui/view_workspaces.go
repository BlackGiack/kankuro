package ui

func (a *App) viewWorkspacesRender() string {
	return a.renderWorkspacesList() + "\n\n" +
		"Enter/Space=edit n=new d=delete r=refresh Esc=back"
}

func (a *App) viewWorkspaceSourceFormRender() string {
	title := "New Workspace Source"
	if a.editingWorkspaceSourceID != "" {
		title = "Edit Workspace Source"
	}
	return titleStyle.Render(title) + "\n\n" +
		"Path: " + a.workspaceInputs[0].View() + "\n" +
		"Is Root (y/n): " + a.workspaceInputs[1].View() + "\n\n" +
		"Enter=save Tab=next Esc=cancel"
}

func (a *App) viewWorkspacePickerRender() string {
	return a.renderWorkspacePicker() + "\n\n" +
		"Enter/Space=select Esc=cancel"
}
