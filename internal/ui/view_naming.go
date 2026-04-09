package ui

func (a *App) viewNamingRender() string {
	var statusLine string
	if a.namingLaunchDone {
		if a.namingLaunchErr != nil {
			statusLine = "Launch failed: " + a.namingLaunchErr.Error()
		} else {
			statusLine = "Session ready. Enter to confirm name and attach."
		}
	} else {
		statusLine = "Launching session..."
	}
	return titleStyle.Render("Name Your Session") + "\n\n" +
		statusLine + "\n\n" +
		"Name: " + a.namingInput.View() + "\n\n" +
		"Enter=confirm  (Leave empty to use session ID as name)"
}
