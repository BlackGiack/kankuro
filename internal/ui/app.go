package ui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
)

type App struct {
	coordinator *Coordinator
}

// New uses the shared API service for human and MCP work.
func New(ports Capabilities, startupError error) (*App, error) {
	return &App{coordinator: NewCoordinator(ports, startupError)}, nil
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

// Close releases selection-scoped background operations after the program stops.
func (a *App) Close() {
	a.coordinator.attachmentOwner.close()
	a.coordinator.deletionWorkers.close()
	a.coordinator.clearCenter()
}

// RuntimeError reports a runtime failure to the frontend.
func RuntimeError(err error) tea.Msg { return messages.MCPServerFailedMsg{Err: err} }
