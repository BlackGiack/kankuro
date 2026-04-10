package ui

func (a *App) viewNamingRender() string {
	var statusLine string
	if a.naming.LaunchDone {
		if a.naming.LaunchErr != nil {
			statusLine = "Launch failed: " + a.naming.LaunchErr.Error()
		} else {
			statusLine = "Session ready. Enter to confirm name and attach."
		}
	} else {
		statusLine = "Launching session..."
	}
	return titleStyle.Render("Name Your Session") + "\n\n" +
		statusLine + "\n\n" +
		"Name: " + a.naming.Input.View() + "\n\n" +
		"Enter=confirm  (Leave empty to use session ID as name)"
}
