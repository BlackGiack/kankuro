package api

import (
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/config"
)

type settingsRepository interface {
	GetAppSettings() (config.AppSettings, error)
	SaveAppSettings(config.AppSettings) error
}

func settingsFromConfig(s config.AppSettings) contracts.Settings {
	return contracts.Settings{
		IDECommand:               s.IDECommand,
		TerminalCommand:          s.TerminalCommand,
		ShowPersonasInCardEditor: s.ShowPersonasInCardEditor,
		ShowMCPsInCardEditor:     s.ShowMCPsInCardEditor,
	}
}

// DefaultSettings returns application defaults without accessing storage.
func (s *Service) DefaultSettings() contracts.Settings {
	return settingsFromConfig(config.DefaultAppSettings())
}

func (s *Service) ReadSettings() (contracts.Settings, error) {
	repo, ok := s.repo.(settingsRepository)
	if !ok {
		return contracts.Settings{}, errors.New("settings repository unavailable")
	}
	value, err := repo.GetAppSettings()
	return settingsFromConfig(value), err
}

func (s *Service) SaveSettings(value contracts.Settings) error {
	repo, ok := s.repo.(settingsRepository)
	if !ok {
		return errors.New("settings repository unavailable")
	}
	return repo.SaveAppSettings(
		config.AppSettings{
			IDECommand:               value.IDECommand,
			TerminalCommand:          value.TerminalCommand,
			ShowPersonasInCardEditor: value.ShowPersonasInCardEditor,
			ShowMCPsInCardEditor:     value.ShowMCPsInCardEditor,
		},
	)
}

type mcpRepository interface {
	UpsertMCP(config.MCPServer) (config.MCPServer, error)
	DeleteMCP(string) error
}

func (s *Service) ReadMCPRows() ([]contracts.MCPRow, error) {
	entries, err := s.repo.ListMCPs()
	if err != nil {
		return nil, err
	}
	rows := make([]contracts.MCPRow, 0, len(entries))
	for _, entry := range entries {
		row := contracts.MCPRow{
			ID:      entry.ID,
			Name:    entry.Name,
			Builtin: entry.ID == config.BuiltinMCPKankuro,
			Edit:    enabled(),
			Delete:  enabled(),
		}
		if entry.ID == config.BuiltinMCPKankuro {
			row.Edit = disabled("the built-in kankuro MCP cannot be edited")
			row.Delete = disabled("the built-in kankuro MCP cannot be deleted")
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func mcpEditor(entry config.MCPServer) contracts.MCPEditor {
	return contracts.MCPEditor{
		ID:          entry.ID,
		Name:        entry.Name,
		Command:     entry.Command,
		Args:        slices.Clone(entry.Args),
		Env:         maps.Clone(entry.Env),
		CWD:         entry.CWD,
		URL:         entry.URL,
		Headers:     maps.Clone(entry.Headers),
		Description: entry.Description,
	}
}

func (s *Service) LoadMCPEditor(id string) (contracts.MCPEditor, error) {
	if id == "" {
		return contracts.MCPEditor{}, nil
	}
	entries, err := s.repo.ListMCPs()
	if err != nil {
		return contracts.MCPEditor{}, err
	}
	for _, entry := range entries {
		if entry.ID == id {
			if id == config.BuiltinMCPKankuro {
				return contracts.MCPEditor{}, errors.New("the built-in kankuro MCP cannot be edited")
			}
			return mcpEditor(entry), nil
		}
	}
	return contracts.MCPEditor{}, fmt.Errorf("mcp not found: %s", id)
}

func (s *Service) SaveMCP(input contracts.MCPEditor) (contracts.MCPEditor, error) {
	repo, ok := s.repo.(mcpRepository)
	if !ok {
		return contracts.MCPEditor{}, errors.New("mcp repository unavailable")
	}
	entry, err := repo.UpsertMCP(
		config.MCPServer{
			ID:          input.ID,
			Name:        input.Name,
			Command:     input.Command,
			Args:        slices.Clone(input.Args),
			Env:         maps.Clone(input.Env),
			CWD:         input.CWD,
			URL:         input.URL,
			Headers:     maps.Clone(input.Headers),
			Description: input.Description,
		},
	)
	return mcpEditor(entry), err
}

func (s *Service) DeleteConfiguredMCP(id string) error {
	repo, ok := s.repo.(mcpRepository)
	if !ok {
		return errors.New("mcp repository unavailable")
	}
	return repo.DeleteMCP(id)
}
