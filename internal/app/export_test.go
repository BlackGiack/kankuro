package app

import (
	tea "charm.land/bubbletea/v2"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

// These adapters expose the same coordinator behavior to external-package tests.
func (c *Coordinator) NavigateTo(v ViewType) tea.Cmd                { return c.navigateTo(v) }
func (c *Coordinator) ResizeAll()                                   { c.resizeAll() }
func (c *Coordinator) DashboardCallbacks() views.DashboardCallbacks { return c.dashboardCallbacks() }
func (c *Coordinator) SourceManagerCallbacks(t views.SourceType) views.SourceManagerCallbacks {
	return c.sourceManagerCallbacks(t)
}
func (c *Coordinator) AgentSwitchCallbacks() views.AgentSwitchCallbacks {
	return c.agentSwitchCallbacks()
}
func (c *Coordinator) RefreshData() tea.Cmd { return c.refreshData() }

func (c *Coordinator) refreshData() tea.Cmd {
	cmd := c.refreshDataSeeded()
	return func() tea.Msg { return unwrapRefresh(cmd()) }
}

func unwrapRefresh(msg tea.Msg) tea.Msg {
	if result, ok := msg.(configFingerprintMsg); ok && result.data != nil {
		return *result.data
	}
	return msg
}
