package app

import (
	"kankuro/internal/config"
	"kankuro/internal/session"
	"kankuro/internal/ui/commands"
)

func WireCommands(repo config.Repository, sm *session.Manager) *commands.Commands {
	cmds := &commands.Commands{}
	wireSessionCommands(cmds, sm)
	wireCardCommands(cmds, repo)
	wireMCPCommands(cmds, repo)
	wireSourceCommands(cmds, repo)
	wireSettingsCommands(cmds, repo)
	cmds.RefreshData = func() {}
	return cmds
}

func wireSessionCommands(cmds *commands.Commands, sm *session.Manager) {
	cmds.LaunchSession = func(card config.IDCard, skills []config.Skill, personas []config.Persona, mcps []config.MCPServer, _ string) (string, error) {
		return sm.Launch(card, skills, personas, mcps)
	}
	cmds.AttachSession = func(sessionID string) error {
		_, err := sm.Attach(sessionID)
		return err
	}
	cmds.KillSession = func(sessionID string) error {
		return sm.Kill(sessionID)
	}
	cmds.RemoveSession = func(sessionID string) error {
		return sm.Remove(sessionID)
	}
	cmds.RelaunchSession = func(sessionID string, card config.IDCard, skills []config.Skill, personas []config.Persona, mcps []config.MCPServer) error {
		return sm.Relaunch(sessionID, card, skills, personas, mcps)
	}
}

func wireCardCommands(cmds *commands.Commands, repo config.Repository) {
	cmds.SaveCard = func(card config.IDCard) (config.IDCard, error) {
		if card.ID == "" {
			return repo.CreateIDCard(card)
		}
		return repo.UpdateIDCard(card)
	}
	cmds.DeleteCard = func(cardID string) error {
		return repo.DeleteIDCard(cardID)
	}
	cmds.GetCard = func(cardID string) (config.IDCard, error) {
		cards, _ := repo.ListIDCards()
		for _, card := range cards {
			if card.ID == cardID {
				return card, nil
			}
		}
		return config.IDCard{}, nil
	}
}

func wireMCPCommands(cmds *commands.Commands, repo config.Repository) {
	cmds.SaveMCP = func(mcp config.MCPServer) (config.MCPServer, error) {
		return repo.UpsertMCP(mcp)
	}
	cmds.DeleteMCP = func(mcpID string) error {
		return repo.DeleteMCP(mcpID)
	}
}

func wireSourceCommands(cmds *commands.Commands, repo config.Repository) {
	cmds.SaveSkillSource = func(path string) error {
		sources, _ := repo.ListSkillSources()
		sources = append(sources, path)
		return repo.SaveSkillSources(sources)
	}
	cmds.DeleteSkillSource = func(path string) error {
		sources, _ := repo.ListSkillSources()
		var filtered []string
		for _, s := range sources {
			if s != path {
				filtered = append(filtered, s)
			}
		}
		return repo.SaveSkillSources(filtered)
	}
	cmds.SavePersonaSource = func(path string) error {
		sources, _ := repo.ListPersonaSources()
		sources = append(sources, path)
		return repo.SavePersonaSources(sources)
	}
	cmds.DeletePersonaSource = func(path string) error {
		sources, _ := repo.ListPersonaSources()
		var filtered []string
		for _, s := range sources {
			if s != path {
				filtered = append(filtered, s)
			}
		}
		return repo.SavePersonaSources(filtered)
	}
	cmds.SaveWorkspaceSource = func(src config.WorkspaceSource) error {
		sources, _ := repo.ListWorkspaceSources()
		sources = append(sources, src)
		return repo.SaveWorkspaceSources(sources)
	}
	cmds.DeleteWorkspaceSource = func(id string) error {
		sources, _ := repo.ListWorkspaceSources()
		var filtered []config.WorkspaceSource
		for _, s := range sources {
			if s.ID != id {
				filtered = append(filtered, s)
			}
		}
		return repo.SaveWorkspaceSources(filtered)
	}
	cmds.SavePuppetSource = func(path string) error {
		sources, _ := repo.ListPuppetSources()
		sources = append(sources, path)
		return repo.SavePuppetSources(sources)
	}
	cmds.DeletePuppetSource = func(path string) error {
		sources, _ := repo.ListPuppetSources()
		var filtered []string
		for _, s := range sources {
			if s != path {
				filtered = append(filtered, s)
			}
		}
		return repo.SavePuppetSources(filtered)
	}
}

func wireSettingsCommands(cmds *commands.Commands, repo config.Repository) {
	cmds.SaveAppSettings = func(cfg config.AppSettings) error {
		return repo.SaveAppSettings(cfg)
	}
	cmds.OpenIDE = func(_ string) error {
		return nil
	}
}
