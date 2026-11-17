package api

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sort"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/google/uuid"
)

type dashboardWriter interface {
	CreateDashboard(config.Dashboard) (config.Dashboard, error)
	UpdateDashboardDetails(id, name, description string) (config.Dashboard, error)
}

type scopedCardWriter interface {
	CreateIDCard(config.IDCard) (config.IDCard, error)
	UpdateIDCard(config.IDCard) (config.IDCard, error)
}

func (s *Service) listCardTombstones() ([]config.CardTombstone, error) {
	return s.repo.ListCardTombstones()
}

func (s *Service) loadMembershipInventory(sessions []session.Snapshot) (*membershipInventory, error) {
	if err := s.sessions.LoadErrors(); err != nil {
		return nil, fmt.Errorf("session inventory is unavailable: %w", err)
	}
	dashboards, err := s.ListDashboards()
	if err != nil {
		return nil, err
	}
	cards, err := s.repo.ListIDCards()
	if err != nil {
		return nil, err
	}
	tombstones, err := s.listCardTombstones()
	if err != nil {
		return nil, err
	}
	return newMembershipInventory(dashboards, cards, sessions, tombstones)
}

// ListDashboards returns the human inventory, including recoverable deletions.
//

func (s *Service) ListDashboards() ([]config.Dashboard, error) {
	dashboards, err := s.repo.ListDashboards()
	if err != nil {
		return nil, err
	}
	manifests, err := s.repo.ListDashboardDeletions()
	if err != nil {
		return nil, err
	}
	for _, manifest := range manifests {
		found := false
		for index := range dashboards {
			if dashboards[index].ID == manifest.DashboardID {
				dashboards[index].State = config.DashboardDeleting
				found = true
				break
			}
		}
		if !found {
			dashboards = append(
				dashboards,
				config.Dashboard{
					ID:          manifest.DashboardID,
					Name:        manifest.Name,
					Description: manifest.Description,
					State:       config.DashboardDeleting,
				},
			)
		}
	}
	deletedDashboards, err := s.repo.ListDashboardTombstones()
	if err != nil {
		return nil, err
	}
	for _, tombstone := range deletedDashboards {
		for _, dashboard := range dashboards {
			if dashboard.ID == tombstone.ID && dashboard.State != config.DashboardDeleting {
				return nil, errors.New("active dashboard conflicts with deleted identity")
			}
		}
	}
	sort.Slice(dashboards, func(i, j int) bool {
		if dashboards[i].Name == dashboards[j].Name {
			return dashboards[i].ID < dashboards[j].ID
		}
		return dashboards[i].Name < dashboards[j].Name
	})
	return dashboards, nil
}

// CreateDashboard creates a dashboard with an allocated immutable identity.
func (s *Service) CreateDashboard(name, description string) (config.Dashboard, error) {
	writer, ok := s.repo.(dashboardWriter)
	if !ok {
		return config.Dashboard{}, errors.New("dashboard inventory unavailable")
	}
	return writer.CreateDashboard(config.Dashboard{Name: name, Description: description})
}

// UpdateDashboardDetails changes the name and description of an active dashboard.
func (s *Service) UpdateDashboardDetails(id, name, description string) (config.Dashboard, error) {
	writer, ok := s.repo.(dashboardWriter)
	if !ok {
		return config.Dashboard{}, errors.New("dashboard inventory unavailable")
	}
	var updated config.Dashboard
	err := s.sessions.WithAdmissionSnapshot(func(snapshot session.AdmissionSnapshot) error {
		inventory, err := s.loadMembershipInventory(snapshot.Sessions)
		if err != nil {
			return err
		}
		if !inventory.activeDashboard(id) {
			return errors.New("dashboard unavailable")
		}
		updated, err = writer.UpdateDashboardDetails(id, name, description)
		return err
	})
	return updated, err
}

// CheckSessionInDashboard revalidates human access without using mutable UI state.
func (s *Service) CheckSessionInDashboard(dashboardID, sessionID string) error {
	return s.sessions.WithAdmissionSnapshot(func(snapshot session.AdmissionSnapshot) error {
		inventory, err := s.loadMembershipInventory(snapshot.Sessions)
		if err != nil {
			return err
		}
		if !inventory.allowsSession(dashboardID, sessionID) {
			return contracts.ErrSessionScope
		}
		return nil
	})
}

// AttachInDashboard validates membership at the subscription boundary.
func (s *Service) AttachInDashboard(ctx context.Context, dashboardID, sessionID string) (contracts.Attachment, error) {
	return s.attachInDashboard(ctx, dashboardID, sessionID, "")
}

// AttachExpectedProcessInDashboard attaches only to the specified process generation.
func (s *Service) AttachExpectedProcessInDashboard(
	ctx context.Context,
	dashboardID, sessionID, processID string,
) (contracts.Attachment, error) {
	if processID == "" {
		return nil, errors.New("expected process identity is required")
	}
	return s.attachInDashboard(ctx, dashboardID, sessionID, processID)
}

func (s *Service) attachInDashboard(ctx context.Context, dashboardID, sessionID, processID string) (contracts.Attachment, error) {
	check := func(sessions []session.Snapshot) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		inventory, err := s.loadMembershipInventory(sessions)
		if err != nil {
			return err
		}
		if !inventory.allowsSession(dashboardID, sessionID) {
			return errors.New("session is outside dashboard scope")
		}
		return nil
	}
	var handle *session.AttachHandle
	var err error
	if processID != "" {
		handle, err = s.sessions.AttachExpectedProcessChecked(sessionID, processID, check)
	} else {
		handle, err = s.sessions.AttachChecked(sessionID, check)
	}
	if err != nil {
		return nil, err
	}
	resource := &attachment{handle: handle, closed: make(chan struct{})}
	if contextErr := ctx.Err(); contextErr != nil {
		_ = resource.Close()
		return nil, contextErr
	}
	return resource, nil
}

// sessionScopeCheck executes while manager admission is held.
func (s *Service) sessionScopeCheck(dashboardID, sessionID string) func([]session.Snapshot) error {
	return s.sessionScopeCheckContext(context.Background(), dashboardID, sessionID)
}

func (s *Service) sessionScopeCheckContext(ctx context.Context, dashboardID, sessionID string) func([]session.Snapshot) error {
	return func(sessions []session.Snapshot) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		inventory, err := s.loadMembershipInventory(sessions)
		if err != nil {
			return err
		}
		if !inventory.allowsSession(dashboardID, sessionID) {
			return errors.New("session is outside dashboard scope")
		}
		return ctx.Err()
	}
}

// RenameInDashboard validates membership at the mutation boundary.
func (s *Service) RenameInDashboard(dashboardID, sessionID, name string) error {
	_, err := s.sessions.SetNameDetailedChecked(sessionID, name, s.sessionScopeCheck(dashboardID, sessionID))
	return err
}

// StopInDashboard validates membership at the stop reservation boundary.
func (s *Service) StopInDashboard(dashboardID, sessionID string) error {
	_, err := s.sessions.KillDetailedChecked(sessionID, s.sessionScopeCheck(dashboardID, sessionID))
	return err
}

// RemoveInDashboard validates membership at the deletion reservation boundary.
func (s *Service) RemoveInDashboard(dashboardID, sessionID string) error {
	_, err := s.sessions.RemoveDetailedChecked(sessionID, s.sessionScopeCheck(dashboardID, sessionID))
	return err
}

func hasForeignCardReferences(impact session.CardReferenceImpact, ownerID string) bool {
	for _, reference := range impact.Sessions {
		if reference.DashboardID != ownerID {
			return true
		}
	}
	for dashboardID := range impact.PendingByDashboard {
		if dashboardID != ownerID {
			return true
		}
	}
	return false
}

// LaunchInDashboard reserves a card and revalidates its configuration before
// publishing a new session. Dashboard and card mutations must use the same
// manager-owned admission boundary for this check to provide isolation.
func (s *Service) LaunchInDashboard(ctx context.Context, dashboardID, cardID string) (string, error) {
	return s.launchDashboardWithID(ctx, dashboardID, cardID, uuid.NewString(), nil)
}

func (s *Service) launchDashboardWithID(ctx context.Context, dashboardID, cardID, sessionID string, name *string) (string, error) {
	return s.launchDashboardWithIDEvidence(ctx, dashboardID, cardID, sessionID, name, nil)
}

func (s *Service) launchDashboardWithIDEvidence(
	ctx context.Context,
	dashboardID, cardID, sessionID string,
	name *string,
	evidence *session.ActivationResult,
) (string, error) {
	launchCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	release, err := s.sessions.ReserveCardReferenceForDashboardWithCancel(
		cardID,
		dashboardID,
		sessionID,
		cancel,
		func(sessions []session.Snapshot) (bool, error) {
			_, available, checkErr := s.availableDashboardCard(dashboardID, cardID, sessions)
			return available, checkErr
		},
	)
	if err != nil {
		return "", err
	}
	defer release()
	bundle, err := s.resolveCard(cardID)
	if err != nil {
		return "", err
	}
	publicationCheck := func(sessions []session.Snapshot) error {
		current, available, checkErr := s.availableDashboardCard(dashboardID, cardID, sessions)
		if checkErr != nil {
			return checkErr
		}
		if !available || !reflect.DeepEqual(current, bundle.Card) {
			return errors.New("card or dashboard changed before activation publication")
		}
		return nil
	}
	result := s.sessions.LaunchInDashboardWithIDResult(launchCtx, sessionID, dashboardID, bundle.Card,
		bundle.Skills, bundle.Personas, bundle.MCPs, publicationCheck, name)
	if evidence != nil {
		*evidence = result
	}
	return sessionID, result.Err
}

// SwitchInDashboard retains session membership while revalidating the target
// session and destination card before activation publication.
func (s *Service) SwitchInDashboard(ctx context.Context, dashboardID, sessionID, cardID string, choice session.ConversationChoice) error {
	return s.switchInDashboardEvidence(ctx, dashboardID, sessionID, cardID, choice, nil)
}

func (s *Service) switchInDashboardEvidence(
	ctx context.Context,
	dashboardID, sessionID, cardID string,
	choice session.ConversationChoice,
	evidence *session.ActivationResult,
) error {
	activationCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	release, err := s.sessions.ReserveCardReferenceForDashboardWithCancel(
		cardID,
		dashboardID,
		sessionID,
		cancel,
		func(sessions []session.Snapshot) (bool, error) {
			_, available, checkErr := s.availableDashboardCard(dashboardID, cardID, sessions)
			if checkErr != nil {
				return false, checkErr
			}
			if !sessionInDashboard(sessions, dashboardID, sessionID) {
				return false, errors.New("session is outside dashboard scope")
			}
			return available, nil
		},
	)
	if err != nil {
		return err
	}
	defer release()
	bundle, err := s.resolveCard(cardID)
	if err != nil {
		return err
	}
	publicationCheck := func(sessions []session.Snapshot) error {
		current, available, checkErr := s.availableDashboardCard(dashboardID, cardID, sessions)
		if checkErr != nil {
			return checkErr
		}
		if !available || !sessionInDashboard(sessions, dashboardID, sessionID) || !reflect.DeepEqual(current, bundle.Card) {
			return errors.New("session, card or dashboard changed before activation publication")
		}
		return nil
	}
	result := s.sessions.SwitchAgentWithChoiceCheckedResult(activationCtx, sessionID, bundle.Card,
		bundle.Skills, bundle.Personas, bundle.MCPs, choice, publicationCheck)
	if evidence != nil {
		*evidence = result
	}
	return result.Err
}

func sessionInDashboard(sessions []session.Snapshot, dashboardID, sessionID string) bool {
	for _, item := range sessions {
		if item.ID == sessionID {
			return item.DashboardID == dashboardID
		}
	}
	return false
}

// RelaunchInDashboard preserves saved workspace and deployment settings while
// rejecting a deleted or no-longer-available current card.
func (s *Service) RelaunchInDashboard(ctx context.Context, dashboardID, sessionID string, choice session.ConversationChoice) error {
	return s.relaunchInDashboardEvidence(ctx, dashboardID, sessionID, "", choice, nil)
}

func (s *Service) relaunchInDashboardEvidence(
	ctx context.Context,
	dashboardID, sessionID, expectedCardID string,
	choice session.ConversationChoice,
	evidence *session.ActivationResult,
) error {
	snapshot, exists := s.sessions.GetSession(sessionID)
	if !exists || snapshot.DashboardID != dashboardID {
		return errors.New("session is outside dashboard scope")
	}
	cardID := snapshot.CardID
	if expectedCardID != "" && expectedCardID != cardID {
		return errors.New("session card changed before activation")
	}
	activationCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	release, err := s.sessions.ReserveCardReferenceForDashboardWithCancel(
		cardID,
		dashboardID,
		sessionID,
		cancel,
		func(sessions []session.Snapshot) (bool, error) {
			_, available, checkErr := s.availableDashboardCard(dashboardID, cardID, sessions)
			return available && sessionUsesCard(sessions, dashboardID, sessionID, cardID), checkErr
		},
	)
	if err != nil {
		return fmt.Errorf("card unavailable; choose another available card: %w", err)
	}
	defer release()
	card, err := s.loadCard(cardID)
	if err != nil {
		return err
	}
	publicationCheck := func(sessions []session.Snapshot) error {
		current, available, checkErr := s.availableDashboardCard(dashboardID, cardID, sessions)
		if checkErr != nil {
			return checkErr
		}
		if !available || !sessionUsesCard(sessions, dashboardID, sessionID, cardID) || !reflect.DeepEqual(current, card) {
			return errors.New("session, card or dashboard changed before activation publication")
		}
		return nil
	}
	result := s.sessions.RelaunchWithChoiceCheckedResult(activationCtx, sessionID, card, choice, publicationCheck)
	if evidence != nil {
		*evidence = result
	}
	return result.Err
}

func sessionUsesCard(sessions []session.Snapshot, dashboardID, sessionID, cardID string) bool {
	for _, item := range sessions {
		if item.ID == sessionID {
			return item.DashboardID == dashboardID && item.CardID == cardID
		}
	}
	return false
}

func (s *Service) availableDashboardCard(dashboardID, cardID string, sessions []session.Snapshot) (config.IDCard, bool, error) {
	inventory, err := s.loadMembershipInventory(sessions)
	if err != nil {
		return config.IDCard{}, false, err
	}
	if !inventory.allowsCard(dashboardID, cardID) {
		return config.IDCard{}, false, nil
	}
	return inventory.cards[cardID], true, nil
}

// dashboardData is a backend snapshot used only while deriving focused results.
type dashboardData struct {
	Cards    []config.IDCard
	Sessions []session.Snapshot
}

// readDashboardData returns cards available to and sessions owned by dashboardID.
// It refuses an invalid inventory rather than broadening the requested scope.
func (s *Service) readDashboardData(dashboardID string) (dashboardData, error) {
	sessions := s.sessions.List()
	inventory, err := s.loadMembershipInventory(sessions)
	if err != nil {
		return dashboardData{}, err
	}
	if !inventory.activeDashboard(dashboardID) {
		return dashboardData{}, errors.New("dashboard unavailable")
	}
	data := dashboardData{Cards: []config.IDCard{}, Sessions: []session.Snapshot{}}
	for _, card := range inventory.cards {
		if inventory.allowsCard(dashboardID, card.ID) {
			data.Cards = append(data.Cards, card)
		}
	}
	for _, tombstone := range inventory.tombstones {
		if tombstone.DashboardID == dashboardID && tombstone.State == config.CardDeleting {
			data.Cards = append(
				data.Cards,
				config.IDCard{
					ID:            tombstone.ID,
					Name:          tombstone.Name + " [deleting]",
					DashboardID:   dashboardID,
					DeletionState: config.CardDeleting,
				},
			)
		}
	}
	sort.Slice(data.Cards, func(i, j int) bool {
		if data.Cards[i].Name == data.Cards[j].Name {
			return data.Cards[i].ID < data.Cards[j].ID
		}
		return data.Cards[i].Name < data.Cards[j].Name
	})
	for _, item := range sessions {
		if inventory.allowsSession(dashboardID, item.ID) {
			item.CardDeleted = inventory.hasTombstone(item.CardID)
			data.Sessions = append(data.Sessions, item)
		}
	}
	return data, nil
}

// membershipInventory is a validated snapshot, not an admission reservation.
// Callers must rebuild and revalidate it under the shared admission boundary
// before publishing a mutation.
type membershipInventory struct {
	dashboards map[string]config.Dashboard
	cards      map[string]config.IDCard
	sessions   map[string]session.Snapshot
	tombstones map[string]config.CardTombstone
}

//nolint:gocognit,gocyclo,cyclop // Validation of all membership domains forms one inventory boundary.
func newMembershipInventory(
	dashboards []config.Dashboard,
	cards []config.IDCard,
	sessions []session.Snapshot,
	tombstones []config.CardTombstone,
) (*membershipInventory, error) {
	inventory := &membershipInventory{
		dashboards: make(map[string]config.Dashboard, len(dashboards)),
		cards:      make(map[string]config.IDCard, len(cards)),
		sessions:   make(map[string]session.Snapshot, len(sessions)),
		tombstones: make(map[string]config.CardTombstone, len(tombstones)),
	}
	for _, dashboard := range dashboards {
		if dashboard.ID == "" || (dashboard.State != config.DashboardActive && dashboard.State != config.DashboardDeleting) {
			return nil, errors.New("invalid dashboard membership inventory")
		}
		if _, exists := inventory.dashboards[dashboard.ID]; exists {
			return nil, fmt.Errorf("duplicate dashboard ID %q", dashboard.ID)
		}
		inventory.dashboards[dashboard.ID] = dashboard
	}
	for _, card := range cards {
		if card.ID == "" || card.DashboardID == "" {
			return nil, errors.New("card has missing identity or dashboard membership")
		}
		if _, exists := inventory.dashboards[card.DashboardID]; !exists {
			return nil, fmt.Errorf("card %q has unknown dashboard", card.ID)
		}
		if _, exists := inventory.cards[card.ID]; exists {
			return nil, fmt.Errorf("duplicate card ID %q", card.ID)
		}
		inventory.cards[card.ID] = card
	}
	for _, tombstone := range tombstones {
		if tombstone.ID == "" || tombstone.DashboardID == "" ||
			(tombstone.State != config.CardDeleting && tombstone.State != config.CardDeleted) {
			return nil, errors.New("invalid card tombstone")
		}
		if _, exists := inventory.tombstones[tombstone.ID]; exists {
			return nil, fmt.Errorf("duplicate card tombstone %q", tombstone.ID)
		}
		if card, exists := inventory.cards[tombstone.ID]; exists &&
			(card.DashboardID != tombstone.DashboardID || tombstone.State == config.CardDeleted) {
			return nil, fmt.Errorf("card %q conflicts with its tombstone", tombstone.ID)
		}
		inventory.tombstones[tombstone.ID] = tombstone
	}
	for _, item := range sessions {
		if item.ID == "" || item.DashboardID == "" {
			return nil, errors.New("session has missing identity or dashboard membership")
		}
		if _, exists := inventory.dashboards[item.DashboardID]; !exists {
			return nil, fmt.Errorf("session %q has unknown dashboard", item.ID)
		}
		if _, exists := inventory.cards[item.CardID]; !exists && !inventory.hasTombstone(item.CardID) {
			return nil, fmt.Errorf("session %q has unknown current card", item.ID)
		}
		if card, exists := inventory.cards[item.CardID]; exists && !inventory.hasTombstone(item.CardID) && !card.Shared &&
			card.DashboardID != item.DashboardID {
			return nil, fmt.Errorf("session %q references a foreign exclusive card", item.ID)
		}
		for cardID := range item.Conversations {
			if _, exists := inventory.cards[cardID]; !exists && !inventory.hasTombstone(cardID) {
				return nil, fmt.Errorf("session %q has unknown saved card", item.ID)
			}
			if card, exists := inventory.cards[cardID]; exists && !inventory.hasTombstone(cardID) && !card.Shared &&
				card.DashboardID != item.DashboardID {
				return nil, fmt.Errorf("session %q has a foreign exclusive saved card", item.ID)
			}
		}
		if _, exists := inventory.sessions[item.ID]; exists {
			return nil, fmt.Errorf("duplicate session ID %q", item.ID)
		}
		inventory.sessions[item.ID] = item
	}
	return inventory, nil
}

func (i *membershipInventory) activeDashboard(id string) bool {
	dashboard, exists := i.dashboards[id]
	return exists && dashboard.State == config.DashboardActive
}

func (i *membershipInventory) allowsCard(dashboardID, cardID string) bool {
	card, exists := i.cards[cardID]
	return exists && !i.hasTombstone(cardID) && i.activeDashboard(dashboardID) && i.activeDashboard(card.DashboardID) &&
		(card.DashboardID == dashboardID || card.Shared)
}

func (i *membershipInventory) hasTombstone(cardID string) bool {
	_, exists := i.tombstones[cardID]
	return exists
}

func (i *membershipInventory) allowsSession(dashboardID, sessionID string) bool {
	item, exists := i.sessions[sessionID]
	return exists && i.activeDashboard(dashboardID) && item.DashboardID == dashboardID
}
