package api

import (
	"errors"
	"slices"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/session"
)

var errCardRevisionConflict = errors.New("card changed; reload before saving")

// LoadCardEditor returns only editable values, with an opaque concurrency identity.
func (s *Service) LoadCardEditor(dashboardID, cardID string) (contracts.CardEditorData, error) {
	data, err := s.readDashboardData(dashboardID)
	if err != nil {
		return contracts.CardEditorData{}, err
	}
	var card config.IDCard
	found := false
	for _, candidate := range data.Cards {
		if candidate.ID == cardID && candidate.DeletionState == "" {
			card, found = candidate, true
			break
		}
	}
	if !found {
		return contracts.CardEditorData{}, errors.New("card unavailable")
	}
	revision, err := cardRevision(card)
	if err != nil {
		return contracts.CardEditorData{}, err
	}
	result := contracts.CardEditorData{
		DashboardID:      dashboardID,
		OwnerDashboardID: card.DashboardID,
		CardID:           card.ID,
		Revision:         revision,
		CanEdit:          card.DashboardID == dashboardID,
		Values:           cardValues(card),
	}
	if !result.CanEdit {
		result.Reason = "Only the owning dashboard can edit this card"
	}
	return result, nil
}

func cardValues(card config.IDCard) contracts.CardValues {
	return contracts.CardValues{
		Name:               card.Name,
		Driver:             card.Driver,
		Model:              card.Model,
		WorkspaceID:        card.WorkspaceID,
		SubWorkspaceIDs:    slices.Clone(card.SubWorkspaceIDs),
		FaceStyle:          card.FaceStyle,
		Puppet:             card.Puppet,
		UseTmpWorkspace:    card.UseTmpWorkspace,
		DashboardControl:   card.DashboardControl,
		Shared:             card.Shared,
		AgentsMD:           card.AgentsMD,
		AgentsMDPath:       card.AgentsMDPath,
		AgentsMDPreferPath: card.AgentsMDPreferPath,
		SkillIDs:           slices.Clone(card.SkillIDs),
		PersonaIDs:         slices.Clone(card.PersonaIDs),
		MCPIDs:             slices.Clone(card.MCPIDs),
	}
}

// NewCardEditor returns the application defaults for a new card.
func (s *Service) NewCardEditor(dashboardID string) (contracts.CardEditorData, error) {
	data, err := s.readDashboardData(dashboardID)
	if err != nil {
		return contracts.CardEditorData{}, err
	}
	_ = data
	return contracts.CardEditorData{DashboardID: dashboardID, CanEdit: true, Values: contracts.CardValues{Puppet: "kankuro"}}, nil
}

// CreateEditorCard allocates identity and timestamps in the repository.
func (s *Service) CreateEditorCard(input contracts.CreateCardInput) (contracts.CardSaveResult, error) {
	card := config.IDCard{Puppet: "kankuro"}
	card.Name = input.Values.Name
	card.Driver = input.Values.Driver
	card.Model = input.Values.Model
	card.WorkspaceID = input.Values.WorkspaceID
	card.SubWorkspaceIDs = slices.Clone(input.Values.SubWorkspaceIDs)
	card.FaceStyle = input.Values.FaceStyle
	if input.Values.Puppet != "" {
		card.Puppet = input.Values.Puppet
	}
	card.UseTmpWorkspace = input.Values.UseTmpWorkspace
	card.DashboardControl = input.Values.DashboardControl
	card.Shared = input.Values.Shared
	card.AgentsMD = input.Values.AgentsMD
	card.AgentsMDPath = input.Values.AgentsMDPath
	card.AgentsMDPreferPath = input.Values.AgentsMDPreferPath
	card.SkillIDs = slices.Clone(input.Values.SkillIDs)
	card.PersonaIDs = slices.Clone(input.Values.PersonaIDs)
	card.MCPIDs = slices.Clone(input.Values.MCPIDs)
	writer, ok := s.repo.(scopedCardWriter)
	if !ok {
		return contracts.CardSaveResult{}, errors.New("card storage unavailable")
	}
	if _, err := cardRevision(card); err != nil {
		return contracts.CardSaveResult{}, err
	}
	var saved config.IDCard
	err := s.sessions.WithCardReferencesAndSnapshot("", func(_ session.CardReferenceImpact, sessions []session.Snapshot) error {
		inventory, err := s.loadMembershipInventory(sessions)
		if err != nil {
			return err
		}
		if !inventory.activeDashboard(input.DashboardID) {
			return errors.New("dashboard unavailable")
		}
		card.DashboardID = input.DashboardID
		saved, err = writer.CreateIDCard(card)
		return err
	})
	if err != nil {
		return contracts.CardSaveResult{}, err
	}
	revision, err := cardRevision(saved)
	if err != nil {
		return contracts.CardSaveResult{}, err
	}
	return contracts.CardSaveResult{CardID: saved.ID, Revision: revision, Status: contracts.CardSaveSucceeded}, nil
}

// UpdateEditorCard revalidates authority and the revision within card admission.
func (s *Service) UpdateEditorCard(input contracts.UpdateCardInput) (contracts.CardSaveResult, error) {
	writer, ok := s.repo.(scopedCardWriter)
	if !ok {
		return contracts.CardSaveResult{}, errors.New("card storage unavailable")
	}
	var saved config.IDCard
	err := s.sessions.WithCardReferencesAndSnapshot(
		input.CardID,
		func(impact session.CardReferenceImpact, sessions []session.Snapshot) error {
			current, err := s.cardOwnedBy(input.DashboardID, input.CardID, impact, sessions)
			if err != nil {
				return err
			}
			revision, err := cardRevision(current)
			if err != nil {
				return err
			}
			if input.ExpectedRevision == "" || revision != input.ExpectedRevision {
				return errCardRevisionConflict
			}
			updated := applyCardChanges(current, input.Changes)
			if current.Shared && !updated.Shared && hasForeignCardReferences(impact, input.DashboardID) {
				return errors.New("cannot unshare a card referenced by another dashboard")
			}
			if _, revisionErr := cardRevision(updated); revisionErr != nil {
				return revisionErr
			}
			saved, err = writer.UpdateIDCard(updated)
			return err
		},
	)
	if errors.Is(err, errCardRevisionConflict) {
		return contracts.CardSaveResult{Status: contracts.CardSaveConflict, Reason: err.Error()}, err
	}
	if err != nil {
		return contracts.CardSaveResult{}, err
	}
	revision, err := cardRevision(saved)
	if err != nil {
		return contracts.CardSaveResult{}, err
	}
	return contracts.CardSaveResult{CardID: saved.ID, Revision: revision, Status: contracts.CardSaveSucceeded}, nil
}

// ReadCardEditorChoices reads registries independently so one failure does not
// discard the last successful value of another registry.
//
//nolint:gocognit // Each registry read has independent failure and value semantics.
func (s *Service) ReadCardEditorChoices() contracts.CardEditorChoices {
	var result contracts.CardEditorChoices
	result.Workspaces.Attempted = true
	if repo, ok := s.repo.(workspaceLister); ok {
		values, err := repo.ListWorkspaces()
		result.Workspaces.Err = err
		if err == nil {
			result.Workspaces.Value = make([]contracts.WorkspaceOption, 0, len(values))
			for _, v := range values {
				result.Workspaces.Value = append(result.Workspaces.Value, contracts.WorkspaceOption{ID: v.ID, Name: v.Name, Path: v.Path})
			}
		}
	} else {
		result.Workspaces.Err = errors.New("workspace inventory unavailable")
	}
	result.Skills.Attempted = true
	skills, err := s.repo.ListSkills()
	result.Skills.Err = err
	if err == nil {
		result.Skills.Value = make([]contracts.EditorOption, 0, len(skills))
		for _, v := range skills {
			result.Skills.Value = append(result.Skills.Value, contracts.EditorOption{ID: v.ID, Name: v.Name, Description: v.Description})
		}
	}
	result.Personas.Attempted = true
	personas, err := s.repo.ListPersonas()
	result.Personas.Err = err
	if err == nil {
		result.Personas.Value = make([]contracts.EditorOption, 0, len(personas))
		for _, v := range personas {
			result.Personas.Value = append(result.Personas.Value, contracts.EditorOption{ID: v.ID, Name: v.Name, Description: v.Source})
		}
	}
	result.MCPs.Attempted = true
	mcps, err := s.repo.ListMCPs()
	result.MCPs.Err = err
	if err == nil {
		result.MCPs.Value = make([]contracts.EditorOption, 0, len(mcps))
		for _, v := range mcps {
			result.MCPs.Value = append(result.MCPs.Value, contracts.EditorOption{ID: v.ID, Name: v.Name, Description: v.Description})
		}
	}
	result.Puppets.Attempted = true
	if repo, ok := s.repo.(puppetLister); ok {
		puppets, puppetErr := repo.ListPuppets()
		result.Puppets.Err = puppetErr
		if puppetErr == nil {
			result.Puppets.Value = make([]contracts.PuppetChoice, 0, len(puppets))
			for _, v := range puppets {
				result.Puppets.Value = append(
					result.Puppets.Value,
					contracts.PuppetChoice{ID: v.ID, Name: v.Name, Lines: slices.Clone(v.Lines), Builtin: v.Builtin},
				)
			}
		}
	} else {
		result.Puppets.Err = errors.New("puppet inventory unavailable")
	}
	return result
}

func applyCardChanges(current config.IDCard, changes contracts.CardChanges) config.IDCard {
	updated := current
	if changes.Name.Present {
		updated.Name = changes.Name.Value
	}
	if changes.Driver.Present {
		updated.Driver = changes.Driver.Value
	}
	if changes.Model.Present {
		updated.Model = changes.Model.Value
	}
	if changes.WorkspaceID.Present {
		updated.WorkspaceID = changes.WorkspaceID.Value
	}
	if changes.SubWorkspaceIDs.Present {
		updated.SubWorkspaceIDs = slices.Clone(changes.SubWorkspaceIDs.Value)
	}
	if changes.FaceStyle.Present {
		updated.FaceStyle = changes.FaceStyle.Value
	}
	if changes.Puppet.Present {
		updated.Puppet = changes.Puppet.Value
	}
	if changes.UseTmpWorkspace.Present {
		updated.UseTmpWorkspace = changes.UseTmpWorkspace.Value
	}
	if changes.DashboardControl.Present {
		updated.DashboardControl = changes.DashboardControl.Value
	}
	if changes.Shared.Present {
		updated.Shared = changes.Shared.Value
	}
	if changes.AgentsMD.Present {
		updated.AgentsMD = changes.AgentsMD.Value
	}
	if changes.AgentsMDPath.Present {
		updated.AgentsMDPath = changes.AgentsMDPath.Value
	}
	if changes.AgentsMDPreferPath.Present {
		updated.AgentsMDPreferPath = changes.AgentsMDPreferPath.Value
	}
	if changes.SkillIDs.Present {
		updated.SkillIDs = slices.Clone(changes.SkillIDs.Value)
	}
	if changes.PersonaIDs.Present {
		updated.PersonaIDs = slices.Clone(changes.PersonaIDs.Value)
	}
	if changes.MCPIDs.Present {
		updated.MCPIDs = slices.Clone(changes.MCPIDs.Value)
	}

	return updated
}
