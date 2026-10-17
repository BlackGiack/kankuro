// Package api contains UI-independent application operations shared by
// frontends. The session manager remains the owner of process lifecycle.
package api

import (
	"fmt"
	"strings"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/session"
)

// Repository is the configuration surface required by Service.
type Repository interface {
	ListIDCards() ([]config.IDCard, error)
	DeleteIDCard(id string) error
	ListSkills() ([]config.Skill, error)
	ListPersonas() ([]config.Persona, error)
	ListMCPs() ([]config.MCPServer, error)
}

// Service coordinates configuration resolution with the existing session manager.
type Service struct {
	repo     Repository
	sessions *session.Manager
}

type cardBundle struct {
	Card     config.IDCard
	Skills   []config.Skill
	Personas []config.Persona
	MCPs     []config.MCPServer
}

func New(repo Repository, sessions *session.Manager) *Service {
	return &Service{repo: repo, sessions: sessions}
}

// Launch resolves capabilities from the supplied card snapshot and starts it.
func (s *Service) Launch(card config.IDCard) (string, error) {
	bundle, err := s.resolveCapabilities(card)
	if err != nil {
		return "", err
	}
	return s.sessions.Launch(card, bundle.Skills, bundle.Personas, bundle.MCPs)
}

// Relaunch starts a session using its currently saved card configuration.
func (s *Service) Relaunch(sessionID string) error {
	var cardID string
	found := false
	for _, snapshot := range s.sessions.List() {
		if snapshot.ID == sessionID {
			cardID = snapshot.CardID
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("session not found: %s", sessionID)
	}
	card, err := s.loadCard(cardID)
	if err != nil {
		return err
	}
	return s.sessions.Relaunch(sessionID, card)
}

// SwitchAgent changes the card while the manager retains the session workspace.
func (s *Service) SwitchAgent(sessionID, cardID string) error {
	bundle, err := s.resolveCard(cardID)
	if err != nil {
		return err
	}
	return s.sessions.SwitchAgent(sessionID, bundle.Card, bundle.Skills, bundle.Personas, bundle.MCPs)
}

// DeleteCard refuses deletion while any session, including a terminated one, references it.
func (s *Service) DeleteCard(cardID string) error {
	var referencedBy []string
	for _, snapshot := range s.sessions.List() {
		if snapshot.CardID == cardID {
			referencedBy = append(referencedBy, snapshot.Name)
		}
	}
	if len(referencedBy) > 0 {
		return fmt.Errorf("cannot delete card: referenced by session(s) %s", strings.Join(referencedBy, ", "))
	}
	return s.repo.DeleteIDCard(cardID)
}

// RemoveSession delegates lifecycle guards and workspace cleanup to the manager.
func (s *Service) RemoveSession(sessionID string) error { return s.sessions.Remove(sessionID) }

func (s *Service) resolveCard(cardID string) (cardBundle, error) {
	card, err := s.loadCard(cardID)
	if err != nil {
		return cardBundle{}, err
	}
	resolved, err := s.resolveCapabilities(card)
	if err != nil {
		return cardBundle{}, err
	}
	resolved.Card = card
	return resolved, nil
}

func (s *Service) loadCard(cardID string) (config.IDCard, error) {
	cards, err := s.repo.ListIDCards()
	if err != nil {
		return config.IDCard{}, fmt.Errorf("list cards: %w", err)
	}
	for _, card := range cards {
		if card.ID == cardID {
			return card, nil
		}
	}
	return config.IDCard{}, fmt.Errorf("card not found: %s", cardID)
}

func (s *Service) resolveCapabilities(card config.IDCard) (cardBundle, error) {
	skills, err := s.repo.ListSkills()
	if err != nil {
		return cardBundle{}, fmt.Errorf("list skills: %w", err)
	}
	personas, err := s.repo.ListPersonas()
	if err != nil {
		return cardBundle{}, fmt.Errorf("list personas: %w", err)
	}
	mcps, err := s.repo.ListMCPs()
	if err != nil {
		return cardBundle{}, fmt.Errorf("list mcps: %w", err)
	}
	skillIDs, personaIDs, mcpIDs := idSet(card.SkillIDs), idSet(card.PersonaIDs), idSet(card.MCPIDs)
	bundle := cardBundle{}
	for _, item := range skills {
		if _, ok := skillIDs[item.ID]; ok {
			bundle.Skills = append(bundle.Skills, item)
		}
	}
	for _, item := range personas {
		if _, ok := personaIDs[item.ID]; ok {
			bundle.Personas = append(bundle.Personas, item)
		}
	}
	for _, item := range mcps {
		if _, ok := mcpIDs[item.ID]; ok {
			bundle.MCPs = append(bundle.MCPs, item)
		}
	}
	return bundle, nil
}

func idSet(ids []string) map[string]struct{} {
	set := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		set[id] = struct{}{}
	}
	return set
}
