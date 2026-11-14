package api

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/session"
)

type puppetLister interface {
	ListPuppets() ([]config.Puppet, error)
}
type registryFingerprinter interface {
	Fingerprint() (map[string]config.FileStamp, error)
}

// ReadDashboardInventory returns display-focused rows derived from the checked
// dashboard inventory. These facts do not authorize later operations.
func (s *Service) ReadDashboardInventory() ([]contracts.DashboardSummary, error) {
	dashboards, err := s.ListDashboards()
	if err != nil {
		return nil, err
	}
	rows := make([]contracts.DashboardSummary, 0, len(dashboards))
	for _, dashboard := range dashboards {
		row := contracts.DashboardSummary{ID: dashboard.ID, Name: dashboard.Name, Description: dashboard.Description}
		switch dashboard.State {
		case config.DashboardActive:
			row.State = contracts.DashboardActive
			row.Open = enabled()
			row.Inspect = enabled()
			row.Edit, row.Delete = enabled(), enabled()
		case config.DashboardDeleting:
			row.State, row.Recoverable = contracts.DashboardDeleting, true
			row.Inspect, row.Retry = enabled(), enabled()
		default:
			row.State = contracts.DashboardUnknown
		}
		rows = append(rows, row)
	}
	return rows, nil
}

// ReadDashboardOverview is a display-focused, dashboard-scoped read.
func (s *Service) ReadDashboardOverview(dashboardID string) (contracts.DashboardInventory, error) {
	data, err := s.readDashboardData(dashboardID)
	if err != nil {
		return contracts.DashboardInventory{}, err
	}
	dashboards, err := s.ReadDashboardInventory()
	if err != nil {
		return contracts.DashboardInventory{}, err
	}
	var dashboard contracts.DashboardSummary
	for _, row := range dashboards {
		if row.ID == dashboardID {
			dashboard = row
			break
		}
	}
	if dashboard.State != contracts.DashboardActive {
		return contracts.DashboardInventory{}, errors.New("dashboard unavailable")
	}
	result := contracts.DashboardInventory{
		Dashboard: dashboard,
		Cards:     make([]contracts.CardSummary, 0, len(data.Cards)),
		Sessions:  make([]contracts.SessionSummary, 0, len(data.Sessions)),
	}
	cards := make(map[string]config.IDCard, len(data.Cards))
	for _, card := range data.Cards {
		cards[card.ID] = card
		result.Cards = append(result.Cards, summarizeCard(card, dashboardID))
	}
	for _, item := range data.Sessions {
		result.Sessions = append(result.Sessions, summarizeSession(item, cards))
	}
	return result, nil
}

func summarizeCard(card config.IDCard, dashboardID string) contracts.CardSummary {
	row := contracts.CardSummary{
		ID: card.ID, OwnerDashboardID: card.DashboardID, Shared: card.Shared, Name: card.Name,
		Driver: card.Driver, Model: card.Model, SkillIDs: append([]string(nil), card.SkillIDs...),
		PersonaIDs: append([]string(nil), card.PersonaIDs...), Puppet: card.Puppet, FaceStyle: card.FaceStyle,
		DeletionState: card.DeletionState,
	}
	if card.DeletionState == config.CardDeleting {
		row.Retry, row.Delete = enabled(), disabled("card deletion is already in progress")
		return row
	}
	row.Launch = enabled()
	row.Inspect = enabled()
	if card.DashboardID == dashboardID {
		row.Edit, row.Delete = enabled(), enabled()
		return row
	}
	row.Edit = disabled("only the owning dashboard may edit this card")
	row.Delete = disabled("only the owning dashboard may delete this card")
	return row
}

func summarizeSession(item session.Snapshot, cards map[string]config.IDCard) contracts.SessionSummary {
	row := contracts.SessionSummary{
		ID: item.ID, DashboardID: item.DashboardID, CardID: item.CardID, Name: item.Name,
		Status: string(item.Status), StartedAt: item.StartedAt, PID: item.PID, WorkingDir: item.WorkingDir,
		OriginalDir: item.OriginalDir, TemporaryWorkspace: item.UseTmpWorkspace, CardDeleted: item.CardDeleted,
		Registration: string(item.Registration), RegistrationPersistence: string(item.RegistrationPersistence),
		RegistrationError: item.RegistrationErr, ActiveConversationID: item.ActiveConversationID,
	}
	card, ok := cards[item.CardID]
	if ok {
		row.FaceStyle, row.Puppet = card.FaceStyle, card.Puppet
	} else {
		row.CardUnavailable = true
	}
	for _, binding := range item.Conversations[item.CardID] {
		if binding.ID == item.ActiveConversationID && item.ActiveConversationID != "" {
			row.ActiveConversationName = binding.Name
			break
		}
	}
	unknown := disabled("session state is unknown")
	switch item.Status {
	case session.StatusRunning, session.StatusWaiting:
		row.Attach, row.Stop, row.Rename, row.OpenIDE, row.OpenTerminal = enabled(), enabled(), enabled(), enabled(), enabled()
		row.Remove, row.Relaunch = disabled("stop the session before removing it"), disabled("session is already active")
	case session.StatusTerminated:
		row.Remove, row.Relaunch, row.Switch, row.Rename = enabled(), enabled(), enabled(), enabled()
		row.Attach, row.Stop = disabled("session is stopped"), disabled("session is stopped")
		row.OpenIDE, row.OpenTerminal = enabled(), enabled()
	default:
		row.Attach, row.Relaunch, row.Stop, row.Remove, row.Switch, row.Rename = unknown, unknown, unknown, unknown, unknown, unknown
	}
	if item.CardDeleted || row.CardUnavailable {
		row.Relaunch = disabled("card is unavailable")
	}
	return row
}

// ReadDisplayChoices returns only the values used to label card rows.
func (s *Service) ReadDisplayChoices() contracts.DisplayChoicesResult {
	var result contracts.DisplayChoicesResult
	skills, skillErr := s.repo.ListSkills()
	result.Skills.Attempted = true
	if skillErr != nil {
		result.Skills.Err = fmt.Errorf("list skills: %w", skillErr)
	} else {
		result.Skills.Value = make([]contracts.NamedChoice, 0, len(skills))
		for _, v := range skills {
			result.Skills.Value = append(result.Skills.Value, contracts.NamedChoice{ID: v.ID, Name: v.Name})
		}
	}
	personas, personaErr := s.repo.ListPersonas()
	result.Personas.Attempted = true
	if personaErr != nil {
		result.Personas.Err = fmt.Errorf("list personas: %w", personaErr)
	} else {
		result.Personas.Value = make([]contracts.NamedChoice, 0, len(personas))
		for _, v := range personas {
			result.Personas.Value = append(result.Personas.Value, contracts.NamedChoice{ID: v.ID, Name: v.Name})
		}
	}
	puppetRepo, ok := s.repo.(puppetLister)
	if !ok {
		result.Puppets = contracts.ReadResult[[]contracts.PuppetChoice]{Attempted: true, Err: errors.New("puppet inventory unavailable")}
		return result
	}
	puppets, puppetErr := puppetRepo.ListPuppets()
	result.Puppets.Attempted = true
	if puppetErr != nil {
		result.Puppets.Err = fmt.Errorf("list puppets: %w", puppetErr)
	} else {
		result.Puppets.Value = make([]contracts.PuppetChoice, 0, len(puppets))
		for _, v := range puppets {
			result.Puppets.Value = append(result.Puppets.Value, contracts.PuppetChoice{
				ID: v.ID, Name: v.Name, Lines: append([]string(nil), v.Lines...), Builtin: v.Builtin,
			})
		}
	}
	return result
}

// ReadRegistryChangeToken maps the application's existing metadata fingerprint
// to an opaque frontend-comparable value.
func (s *Service) ReadRegistryChangeToken() (contracts.RegistryChangeToken, error) {
	fingerprinter, ok := s.repo.(registryFingerprinter)
	if !ok {
		return "", errors.New("registry fingerprint unavailable")
	}
	stamps, err := fingerprinter.Fingerprint()
	if err != nil {
		return "", err
	}
	keys := make([]string, 0, len(stamps))
	for key := range stamps {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var serialized strings.Builder
	for _, key := range keys {
		stamp := stamps[key]
		fmt.Fprintf(&serialized, "%s\x00%d\x00%d\n", key, stamp.ModTime.UnixNano(), stamp.Size)
	}
	digest := sha256.Sum256([]byte(serialized.String()))
	return contracts.RegistryChangeToken(hex.EncodeToString(digest[:])), nil
}

func enabled() contracts.Availability               { return contracts.Availability{Enabled: true} }
func disabled(reason string) contracts.Availability { return contracts.Availability{Reason: reason} }
