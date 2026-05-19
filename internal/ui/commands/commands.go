package commands

import (
	"kankuro/internal/config"
)

type Commands struct {
	LaunchSession   func(card config.IDCard, skills []config.Skill, personas []config.Persona, mcps []config.MCPServer, name string) (string, error)
	AttachSession   func(sessionID string) error
	KillSession     func(sessionID string) error
	RemoveSession   func(sessionID string) error
	RelaunchSession func(sessionID string, card config.IDCard, skills []config.Skill, personas []config.Persona, mcps []config.MCPServer) error

	SaveCard   func(card config.IDCard) (config.IDCard, error)
	DeleteCard func(cardID string) error
	GetCard    func(cardID string) (config.IDCard, error)

	SaveMCP   func(mcp config.MCPServer) (config.MCPServer, error)
	DeleteMCP func(mcpID string) error

	SaveSkillSource   func(path string) error
	DeleteSkillSource func(path string) error

	SavePersonaSource   func(path string) error
	DeletePersonaSource func(path string) error

	SaveWorkspaceSource   func(src config.WorkspaceSource) error
	DeleteWorkspaceSource func(id string) error

	SavePuppetSource   func(path string) error
	DeletePuppetSource func(path string) error

	SaveAppSettings func(config.AppSettings) error
	OpenIDE         func(sessionID string) error

	RefreshData func()
}
