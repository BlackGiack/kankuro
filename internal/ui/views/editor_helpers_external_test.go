package views_test

import (
	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/ui/models"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

func loadEditor(e *views.CardEditor, c config.IDCard) {
	e.LoadEditor(
		contracts.CardEditorData{
			CardID:           c.ID,
			DashboardID:      c.DashboardID,
			OwnerDashboardID: c.DashboardID,
			CanEdit:          true,
			Values: contracts.CardValues{
				Name:               c.Name,
				Driver:             c.Driver,
				Model:              c.Model,
				WorkspaceID:        c.WorkspaceID,
				SubWorkspaceIDs:    c.SubWorkspaceIDs,
				FaceStyle:          c.FaceStyle,
				Puppet:             c.Puppet,
				UseTmpWorkspace:    c.UseTmpWorkspace,
				DashboardControl:   c.DashboardControl,
				Shared:             c.Shared,
				AgentsMD:           c.AgentsMD,
				AgentsMDPath:       c.AgentsMDPath,
				AgentsMDPreferPath: c.AgentsMDPreferPath,
				SkillIDs:           c.SkillIDs,
				PersonaIDs:         c.PersonaIDs,
				MCPIDs:             c.MCPIDs,
			},
		},
	)
}
func valuesCard(v contracts.CardValues) config.IDCard {
	return config.IDCard{
		Name:               v.Name,
		Driver:             v.Driver,
		Model:              v.Model,
		WorkspaceID:        v.WorkspaceID,
		SubWorkspaceIDs:    v.SubWorkspaceIDs,
		FaceStyle:          v.FaceStyle,
		Puppet:             v.Puppet,
		UseTmpWorkspace:    v.UseTmpWorkspace,
		DashboardControl:   v.DashboardControl,
		Shared:             v.Shared,
		AgentsMD:           v.AgentsMD,
		AgentsMDPath:       v.AgentsMDPath,
		AgentsMDPreferPath: v.AgentsMDPreferPath,
		SkillIDs:           v.SkillIDs,
		PersonaIDs:         v.PersonaIDs,
		MCPIDs:             v.MCPIDs,
	}
}
func editorCard(e *views.CardEditor) config.IDCard {
	c := valuesCard(e.Draft().Values)
	c.ID = e.EditingID()
	return c
}
func setEditorSettings(e *views.CardEditor, s config.AppSettings) {
	e.SetSettings(models.EditorSettings{ShowPersonas: s.ShowPersonasInCardEditor, ShowMCPs: s.ShowMCPsInCardEditor})
}

type editorTestData struct {
	Workspaces []config.Workspace
	Skills     []config.Skill
	Personas   []config.Persona
	MCPs       []config.MCPServer
	Puppets    []config.Puppet
}

func setEditorData(e *views.CardEditor, d editorTestData) {
	var c contracts.CardEditorChoices
	if d.Workspaces != nil {
		c.Workspaces.Attempted = true
		c.Workspaces.Value = make([]contracts.WorkspaceOption, 0, len(d.Workspaces))
		for _, v := range d.Workspaces {
			c.Workspaces.Value = append(c.Workspaces.Value, contracts.WorkspaceOption{ID: v.ID, Name: v.Name, Path: v.Path})
		}
	}
	if d.Skills != nil {
		c.Skills.Attempted = true
		c.Skills.Value = make([]contracts.EditorOption, 0, len(d.Skills))
		for _, v := range d.Skills {
			c.Skills.Value = append(c.Skills.Value, contracts.EditorOption{ID: v.ID, Name: v.Name, Description: v.Description})
		}
	}
	if d.Personas != nil {
		c.Personas.Attempted = true
		c.Personas.Value = make([]contracts.EditorOption, 0, len(d.Personas))
		for _, v := range d.Personas {
			c.Personas.Value = append(c.Personas.Value, contracts.EditorOption{ID: v.ID, Name: v.Name, Description: v.Source})
		}
	}
	if d.MCPs != nil {
		c.MCPs.Attempted = true
		c.MCPs.Value = make([]contracts.EditorOption, 0, len(d.MCPs))
		for _, v := range d.MCPs {
			c.MCPs.Value = append(c.MCPs.Value, contracts.EditorOption{ID: v.ID, Name: v.Name, Description: v.Description})
		}
	}
	if d.Puppets != nil {
		c.Puppets.Attempted = true
		c.Puppets.Value = make([]contracts.PuppetChoice, 0, len(d.Puppets))
		for _, v := range d.Puppets {
			c.Puppets.Value = append(c.Puppets.Value, contracts.PuppetChoice{ID: v.ID, Name: v.Name, Lines: v.Lines, Builtin: v.Builtin})
		}
	}
	e.SetChoices(c)
}

type testCapability struct{ ID, Name string }

func (v testCapability) GetID() string          { return v.ID }
func (v testCapability) GetLabel() string       { return v.Name }
func (v testCapability) GetDescription() string { return "" }
func workspaceOptions(in []config.Workspace) []models.WorkspaceOption {
	out := make([]models.WorkspaceOption, 0, len(in))
	for _, v := range in {
		out = append(out, models.WorkspaceOption{ID: v.ID, Name: v.Name, Path: v.Path})
	}
	return out
}
func puppetOptions(in []config.Puppet) []models.PuppetOption {
	out := make([]models.PuppetOption, 0, len(in))
	for _, v := range in {
		out = append(out, models.PuppetOption{ID: v.ID, Name: v.Name, Lines: v.Lines, Builtin: v.Builtin})
	}
	return out
}
