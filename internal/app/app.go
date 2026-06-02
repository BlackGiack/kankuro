package app

import (
	tea "charm.land/bubbletea/v2"

	"kankuro/internal/config"
	"kankuro/internal/session"
)

type App struct {
	coordinator *Coordinator
}

func New(repo config.Repository, sm *session.Manager) (*App, error) {
	coord := NewCoordinator(repo, sm)
	return &App{coordinator: coord}, nil
}

func (a *App) Coordinator() *Coordinator {
	return a.coordinator
}

func (a *App) Init() tea.Cmd {
	return a.coordinator.Init()
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	_, cmd := a.coordinator.Update(msg)
	return a, cmd
}

func (a *App) View() tea.View {
	view := tea.NewView(a.coordinator.View())
	view.AltScreen = true
	return view
}

func (a *App) SetSize(width, height int) {
	a.coordinator.width = width
	a.coordinator.height = height
	a.coordinator.resizeAll()
}
