package ui

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
)

type configurationPort interface {
	ReadMCPRows() ([]contracts.MCPRow, error)
	LoadMCPEditor(string) (contracts.MCPEditor, error)
	SaveMCP(contracts.MCPEditor) (contracts.MCPEditor, error)
	DeleteConfiguredMCP(string) error
	ReadSources(contracts.SourceKind) ([]contracts.SourceRow, error)
	MutateSource(context.Context, contracts.SourceMutation) contracts.SourceMutationResult
	SyncSources(context.Context, contracts.SourceKind) contracts.SourceSyncResult
	DefaultSettings() contracts.Settings
	ReadSettings() (contracts.Settings, error)
	SaveSettings(contracts.Settings) error
}

func (c *Coordinator) loadSettings() tea.Cmd {
	c.settingsLoading = true
	c.settingsRequest++
	service, request := c.configuration, c.settingsRequest
	return func() tea.Msg {
		settings, err := service.ReadSettings()
		return messages.SettingsLoadedMsg{Settings: settings, Err: err, Request: request}
	}
}

// mcpDeleter is captured by deletion commands before they leave the coordinator.
type mcpDeleter interface{ DeleteConfiguredMCP(string) error }

func deleteMCPCommand(service mcpDeleter, m messages.MCPDeleteRequestedMsg) tea.Cmd {
	return func() tea.Msg {
		if err := service.DeleteConfiguredMCP(m.MCPID); err != nil {
			return messages.ErrorMsg{Err: err}
		}
		return messages.MCPDeletedMsg(m)
	}
}

func deleteMCPFromManagerCommand(service mcpDeleter, id, dashboardID string, generation, request uint64) tea.Cmd {
	return func() tea.Msg {
		return messages.MCPManagerDeletedMsg{
			MCPID:       id,
			Err:         service.DeleteConfiguredMCP(id),
			DashboardID: dashboardID,
			Generation:  generation,
			Request:     request,
		}
	}
}
