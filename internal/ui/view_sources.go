package ui

// Skill Sources.
func (a *App) viewSkillSourcesRender() string {
	return titleStyle.Render("Manage Skill Sources") + "\n\n" +
		"Defined sources:\n" + a.renderSkillSourcesList() + "\n\n" +
		"Enter/Space=edit n=new d=delete Esc=back"
}

func (a *App) viewSkillSourceFormRender() string {
	title := "New Skill Source"
	if a.skillSources.EditingIdx != -1 {
		title = "Edit Skill Source"
	}
	return titleStyle.Render(title) + "\n\n" +
		"Path: " + a.skillSources.Input.View() + "\n\n" +
		"Enter=save Esc=cancel"
}

// Persona Sources.
func (a *App) viewPersonaSourcesRender() string {
	return titleStyle.Render("Manage Persona Sources") + "\n\n" +
		"Defined sources:\n" + a.renderPersonaSourcesList() + "\n\n" +
		"Enter/Space=edit n=new d=delete Esc=back"
}

func (a *App) viewPersonaSourceFormRender() string {
	title := "New Persona Source"
	if a.personaSources.EditingIdx != -1 {
		title = "Edit Persona Source"
	}
	return titleStyle.Render(title) + "\n\n" +
		"Path: " + a.personaSources.Input.View() + "\n\n" +
		"Enter=save Esc=cancel"
}
