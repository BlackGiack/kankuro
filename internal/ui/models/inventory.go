package models

import (
	"time"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
)

const puppetLookupKeysPerPuppet = 2

type DashboardRow struct {
	ID, Name, Description              string
	State                              contracts.DashboardState
	Recoverable                        bool
	Open, Edit, Delete, Inspect, Retry contracts.Availability
}

type CardRow struct {
	ID, OwnerDashboardID, Name, Driver, Model string
	Shared                                    bool
	SkillIDs, PersonaIDs                      []string
	SkillNames, PersonaNames                  []string
	PuppetName, FaceStyle                     string
	PuppetBuiltin                             bool
	PuppetLines                               []string
	DeletionState                             string
	Launch, Edit, Delete, Retry, Inspect      contracts.Availability
	FilterKey                                 string
}

type SessionRow struct {
	ID, DashboardID, CardID, Name, Status                    string
	StartedAt                                                time.Time
	PID                                                      int
	WorkingDir, OriginalDir                                  string
	TemporaryWorkspace                                       bool
	CardDeleted, CardUnavailable                             bool
	CardName, FaceStyle, PuppetName                          string
	PuppetBuiltin                                            bool
	PuppetLines                                              []string
	Registration, RegistrationPersistence, RegistrationError string
	ActiveConversationID, ActiveConversationName             string
	Attach, Relaunch, Stop, Remove, Switch, Rename           contracts.Availability
	OpenIDE, OpenTerminal                                    contracts.Availability
	FilterKey                                                string
}

func DashboardRows(values []contracts.DashboardSummary) []DashboardRow {
	rows := make([]DashboardRow, len(values))
	for i, value := range values {
		rows[i] = DashboardRow{
			ID: value.ID, Name: value.Name, Description: value.Description, State: value.State,
			Recoverable: value.Recoverable, Open: value.Open, Edit: value.Edit, Delete: value.Delete,
			Inspect: value.Inspect, Retry: value.Retry,
		}
	}
	return rows
}

// CloneCardRow detaches all nested slices in one presentation row.
func CloneCardRow(row CardRow) CardRow {
	row.SkillIDs = cloneStrings(row.SkillIDs)
	row.PersonaIDs = cloneStrings(row.PersonaIDs)
	row.SkillNames = cloneStrings(row.SkillNames)
	row.PersonaNames = cloneStrings(row.PersonaNames)
	row.PuppetLines = cloneStrings(row.PuppetLines)
	return row
}

// CloneCardRows detaches all nested slices in presentation rows.
func CloneCardRows(rows []CardRow) []CardRow {
	if rows == nil {
		return nil
	}
	cloned := append([]CardRow(nil), rows...)
	for i := range cloned {
		cloned[i].SkillIDs = cloneStrings(cloned[i].SkillIDs)
		cloned[i].PersonaIDs = cloneStrings(cloned[i].PersonaIDs)
		cloned[i].SkillNames = cloneStrings(cloned[i].SkillNames)
		cloned[i].PersonaNames = cloneStrings(cloned[i].PersonaNames)
		cloned[i].PuppetLines = cloneStrings(cloned[i].PuppetLines)
	}
	return cloned
}

// CloneSessionRows detaches artwork lines in presentation rows.
func CloneSessionRows(rows []SessionRow) []SessionRow {
	if rows == nil {
		return nil
	}
	cloned := append([]SessionRow(nil), rows...)
	for i := range cloned {
		cloned[i].PuppetLines = cloneStrings(cloned[i].PuppetLines)
	}
	return cloned
}

// CloneSessionRow detaches artwork lines in one presentation row.
func CloneSessionRow(row SessionRow) SessionRow {
	row.PuppetLines = cloneStrings(row.PuppetLines)
	return row
}

// CardRows resolves capability names and artwork before values reach renderers.
func CardRows(cards []contracts.CardSummary, choices contracts.DisplayChoices) []CardRow {
	skills := choiceNames(choices.Skills)
	personas := choiceNames(choices.Personas)
	puppets := make(map[string]contracts.PuppetChoice, len(choices.Puppets))
	for _, puppet := range choices.Puppets {
		puppets[puppet.ID] = puppet
		puppets[puppet.Name] = puppet
	}
	rows := make([]CardRow, 0, len(cards))
	for _, card := range cards {
		row := CardRow{
			ID: card.ID, OwnerDashboardID: card.OwnerDashboardID, Name: card.Name, Driver: card.Driver,
			Model: card.Model, Shared: card.Shared, SkillIDs: cloneStrings(card.SkillIDs),
			PersonaIDs: cloneStrings(card.PersonaIDs), PuppetName: card.Puppet, FaceStyle: card.FaceStyle,
			DeletionState: card.DeletionState, Launch: card.Launch, Edit: card.Edit,
			Delete: card.Delete, Retry: card.Retry, Inspect: card.Inspect,
			PuppetBuiltin: true,
		}
		row.SkillNames = resolveNames(card.SkillIDs, skills)
		row.PersonaNames = resolveNames(card.PersonaIDs, personas)
		if puppet, ok := puppets[card.Puppet]; ok {
			row.PuppetBuiltin = puppet.Builtin
			row.PuppetLines = cloneStrings(puppet.Lines)
		}
		row.FilterKey = card.Name + " " + card.Driver + " " + card.Model
		rows = append(rows, row)
	}
	return rows
}

// SessionRows resolves card labels and copies all strings needed by list views.
func SessionRows(sessions []contracts.SessionSummary, cards []contracts.CardSummary, choices contracts.DisplayChoices) []SessionRow {
	cardByID := make(map[string]contracts.CardSummary, len(cards))
	for _, card := range cards {
		cardByID[card.ID] = card
	}
	puppets := make(map[string]contracts.PuppetChoice, len(choices.Puppets)*puppetLookupKeysPerPuppet)
	for _, puppet := range choices.Puppets {
		puppets[puppet.ID] = puppet
		puppets[puppet.Name] = puppet
	}
	rows := make([]SessionRow, 0, len(sessions))
	for _, session := range sessions {
		row := SessionRow{
			ID: session.ID, DashboardID: session.DashboardID, CardID: session.CardID, Name: session.Name,
			Status: session.Status, StartedAt: session.StartedAt, PID: session.PID,
			WorkingDir: session.WorkingDir, OriginalDir: session.OriginalDir,
			TemporaryWorkspace: session.TemporaryWorkspace, CardDeleted: session.CardDeleted,
			CardUnavailable: session.CardUnavailable, FaceStyle: session.FaceStyle, PuppetName: session.Puppet,
			Registration: session.Registration, RegistrationPersistence: session.RegistrationPersistence,
			RegistrationError: session.RegistrationError, ActiveConversationID: session.ActiveConversationID,
			ActiveConversationName: session.ActiveConversationName, Attach: session.Attach, Relaunch: session.Relaunch,
			Stop: session.Stop, Remove: session.Remove, Switch: session.Switch, Rename: session.Rename,
			OpenIDE: session.OpenIDE, OpenTerminal: session.OpenTerminal,
			PuppetBuiltin: true,
		}
		if session.CardDeleted {
			row.CardName = "deleted card: " + session.CardID
		} else if card, ok := cardByID[session.CardID]; ok {
			row.CardName = card.Name
			row.FaceStyle, row.PuppetName = card.FaceStyle, card.Puppet
			if puppet, found := puppets[card.Puppet]; found {
				row.PuppetBuiltin = puppet.Builtin
				row.PuppetLines = cloneStrings(puppet.Lines)
			}
		} else {
			row.CardName = session.CardID
		}
		row.FilterKey = session.Name + " " + session.CardID
		rows = append(rows, row)
	}
	return rows
}

func choiceNames(choices []contracts.NamedChoice) map[string]string {
	result := make(map[string]string, len(choices))
	for _, choice := range choices {
		result[choice.ID] = choice.Name
	}
	return result
}

func cloneStrings(values []string) []string {
	if values == nil {
		return nil
	}
	return append([]string{}, values...)
}

func resolveNames(ids []string, choices map[string]string) []string {
	var names []string
	for _, id := range ids {
		if name, ok := choices[id]; ok {
			names = append(names, name)
		}
	}
	return names
}
