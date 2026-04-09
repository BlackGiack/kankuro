package ui

import "fmt"

func (a *App) viewDeleteConfirmRender() string {
	return titleStyle.Render("Delete Session?") + "\n\n" +
		fmt.Sprintf("Are you sure you want to delete session %s?\n", a.pendingDeleteID) +
		"This will also delete its temporary workspace if it exists.\n\n" +
		"y=Confirm  n=Cancel"
}
