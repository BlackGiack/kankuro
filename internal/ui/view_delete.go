package ui

import "fmt"

func (a *App) viewDeleteConfirmRender() string {
	if a.sessionView.PendingDeleteCardID != "" {
		cardName := a.sessionView.PendingDeleteCardID
		for _, c := range a.cards {
			if c.ID == a.sessionView.PendingDeleteCardID {
				cardName = c.Name
				break
			}
		}
		return titleStyle.Render("Delete ID Card?") + "\n\n" +
			fmt.Sprintf("Are you sure you want to delete ID card %s?\n", cardName) +
			"This action cannot be undone.\n\n" +
			"y=Confirm  n=Cancel"
	}

	return titleStyle.Render("Delete Session?") + "\n\n" +
		fmt.Sprintf("Are you sure you want to delete session %s?\n", a.sessionView.PendingDeleteID) +
		"This will also delete its temporary workspace if it exists.\n\n" +
		"y=Confirm  n=Cancel"
}
