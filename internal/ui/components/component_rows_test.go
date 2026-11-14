package components_test

import (
	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/BlackGiack/kankuro/internal/ui/components"
	"github.com/BlackGiack/kankuro/internal/ui/models"
)

func cardRowsFromConfig(cards []config.IDCard, skills []config.Skill, personas []config.Persona, puppets []config.Puppet) []models.CardRow {
	summaries := make([]contracts.CardSummary, 0, len(cards))
	for _, card := range cards {
		summaries = append(summaries, contracts.CardSummary{ID: card.ID, OwnerDashboardID: card.DashboardID, Shared: card.Shared,
			Name: card.Name, Driver: card.Driver, Model: card.Model, SkillIDs: card.SkillIDs, PersonaIDs: card.PersonaIDs,
			Puppet: card.Puppet, FaceStyle: card.FaceStyle, DeletionState: card.DeletionState})
	}
	choices := contracts.DisplayChoices{}
	for _, skill := range skills {
		choices.Skills = append(choices.Skills, contracts.NamedChoice{ID: skill.ID, Name: skill.Name})
	}
	for _, persona := range personas {
		choices.Personas = append(choices.Personas, contracts.NamedChoice{ID: persona.ID, Name: persona.Name})
	}
	for _, puppet := range puppets {
		choices.Puppets = append(choices.Puppets, contracts.PuppetChoice{
			ID: puppet.ID, Name: puppet.Name, Lines: puppet.Lines, Builtin: puppet.Builtin,
		})
	}
	return models.CardRows(summaries, choices)
}

func setPreviewSession(preview *components.SessionPreview, snapshot *session.Snapshot, card *config.IDCard) {
	if snapshot == nil {
		preview.SetSession(nil)
		return
	}
	row := models.SessionRow{
		ID: snapshot.ID, DashboardID: snapshot.DashboardID, CardID: snapshot.CardID,
		Name: snapshot.Name, Status: string(snapshot.Status), StartedAt: snapshot.StartedAt,
		PID: snapshot.PID, WorkingDir: snapshot.WorkingDir, OriginalDir: snapshot.OriginalDir,
		TemporaryWorkspace: snapshot.UseTmpWorkspace, CardDeleted: snapshot.CardDeleted,
		Registration: string(snapshot.Registration), RegistrationPersistence: string(snapshot.RegistrationPersistence),
		RegistrationError: snapshot.RegistrationErr, ActiveConversationID: snapshot.ActiveConversationID,
	}
	if card != nil {
		row.CardName, row.PuppetName, row.FaceStyle = card.Name, card.Puppet, card.FaceStyle
		row.PuppetBuiltin = true
	}
	preview.SetSession(&row)
}

func setPreviewPuppets(preview *components.SessionPreview, puppets []config.Puppet) {
	choices := make([]contracts.PuppetChoice, 0, len(puppets))
	for _, puppet := range puppets {
		choices = append(choices, contracts.PuppetChoice{
			ID: puppet.ID, Name: puppet.Name, Lines: append([]string(nil), puppet.Lines...), Builtin: puppet.Builtin,
		})
	}
	preview.SetPuppetChoices(choices)
}

func sessionRowsFromSnapshots(sessions []session.Snapshot, cards []config.IDCard) []models.SessionRow {
	cardSummaries := make([]contracts.CardSummary, 0, len(cards))
	for _, card := range cards {
		cardSummaries = append(cardSummaries, contracts.CardSummary{
			ID: card.ID, Name: card.Name, FaceStyle: card.FaceStyle, Puppet: card.Puppet,
		})
	}
	summaries := make([]contracts.SessionSummary, 0, len(sessions))
	for _, item := range sessions {
		row := contracts.SessionSummary{
			ID: item.ID, DashboardID: item.DashboardID, CardID: item.CardID, Name: item.Name,
			Status: string(item.Status), StartedAt: item.StartedAt, PID: item.PID,
			WorkingDir: item.WorkingDir, OriginalDir: item.OriginalDir,
			TemporaryWorkspace: item.UseTmpWorkspace, CardDeleted: item.CardDeleted,
			Registration: string(item.Registration), RegistrationPersistence: string(item.RegistrationPersistence),
			RegistrationError: item.RegistrationErr, ActiveConversationID: item.ActiveConversationID,
		}
		for _, binding := range item.Conversations[item.CardID] {
			if binding.ID == item.ActiveConversationID && item.ActiveConversationID != "" {
				row.ActiveConversationName = binding.Name
				break
			}
		}
		summaries = append(summaries, row)
	}
	return models.SessionRows(summaries, cardSummaries, contracts.DisplayChoices{})
}
