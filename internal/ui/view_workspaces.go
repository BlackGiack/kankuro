package ui

func (a *App) viewWorkspacesRender() string {
	return a.renderWorkspacesList() + "\n\n" +
		"Enter/Space=edit n=new d=delete r=refresh Esc=back"
}

func (a *App) viewWorkspaceSourceFormRender() string {
	title := "New Workspace Source"
	if a.workspaceForm.EditingID != "" {
		title = "Edit Workspace Source"
	}
	return titleStyle.Render(title) + "\n\n" +
		"Path: " + a.workspaceForm.Inputs[0].View() + "\n" +
		"Is Root (y/n): " + a.workspaceForm.Inputs[1].View() + "\n\n" +
		"Enter=save Tab=next Esc=cancel"
}

func (a *App) viewWorkspacePickerRender() string {
	return a.renderWorkspacePicker() + "\n\n" +
		"Enter/Space=select Esc=cancel"
}
