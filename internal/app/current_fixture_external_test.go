package app_test

import (
	"context"

	"github.com/BlackGiack/kankuro/internal/api"
	"github.com/BlackGiack/kankuro/internal/app"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/layout"
	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
)

func newCurrentRepository(dir string) (*config.FileRepository, error) {
	if err := layout.PrepareStartup(dir); err != nil {
		return nil, err
	}
	return config.NewFileRepository(dir)
}

func newScopedCoordinator(repo config.Repository, manager *session.Manager) *app.Coordinator {
	return app.NewCoordinator(repo, manager, api.New(repo, manager))
}

func openDefaultDashboard(coordinator *app.Coordinator) {
	coordinator.Update(
		messages.DashboardOpenRequestedMsg{Dashboard: config.Dashboard{ID: "default", Name: "Default", State: config.DashboardActive}},
	)
}

func newOpenedCoordinator(repo config.Repository, manager *session.Manager) *app.Coordinator {
	coordinator := newScopedCoordinator(repo, manager)
	openDefaultDashboard(coordinator)
	return coordinator
}

func createCurrentCard(repo config.Repository, card config.IDCard) (config.IDCard, error) {
	if card.DashboardID == "" {
		card.DashboardID = "default"
	}
	return repo.CreateIDCard(card)
}

func launchCurrentSession(repo config.Repository, manager *session.Manager, card config.IDCard) (string, error) {
	return api.New(repo, manager).LaunchInDashboard(context.Background(), card.DashboardID, card.ID)
}
