package app

import (
	tea "charm.land/bubbletea/v2"
	"github.com/BlackGiack/kankuro/internal/api"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
)

type legacyActions struct {
	repo    config.Repository
	service *api.Service
}

// deleteMCP removes an MCP server off the UI goroutine - disk I/O deferred
// to a tea.Cmd, consistent with every other mutation (it used to run
// synchronously inside Update).
func (a legacyActions) deleteMCP(m messages.MCPDeleteRequestedMsg) tea.Cmd {
	return func() tea.Msg {
		if err := a.repo.DeleteMCP(m.MCPID); err != nil {
			return messages.ErrorMsg{Err: err}
		}
		return messages.MCPDeletedMsg(m)
	}
}

func (a legacyActions) deleteMCPFromManager(id, dashboardID string, generation, request uint64) tea.Cmd {
	return func() tea.Msg {
		return messages.MCPManagerDeletedMsg{
			MCPID: id, Err: a.repo.DeleteMCP(id), DashboardID: dashboardID, Generation: generation, Request: request,
		}
	}
}
