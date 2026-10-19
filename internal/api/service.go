// Package api contains UI-independent application operations shared by
// frontends. The session manager remains the owner of process lifecycle.
package api

import (
	"fmt"
	"slices"
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

// RegistrationResult is the structured result shared by the MCP transport and manager.
type RegistrationResult = session.RegistrationResult

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

// Relaunch restarts a session without redeploying its workspace configuration.
func (s *Service) Relaunch(sessionID string) error {
	return s.RelaunchWithChoice(sessionID, session.ConversationChoice{})
}

// RelaunchWithChoice restarts a session with an explicit native conversation choice.
func (s *Service) RelaunchWithChoice(sessionID string, choice session.ConversationChoice) error {
	snapshot, found := s.sessions.GetSession(sessionID)
	if !found {
		return fmt.Errorf("session not found: %s", sessionID)
	}
	card, err := s.loadCard(snapshot.CardID)
	if err != nil {
		return err
	}
	return s.sessions.RelaunchWithChoice(sessionID, card, choice)
}

// SwitchAgent changes the card while the manager retains the session workspace.
func (s *Service) SwitchAgent(sessionID, cardID string) error {
	bundle, err := s.resolveCard(cardID)
	if err != nil {
		return err
	}
	return s.sessions.SwitchAgentWithChoice(
		sessionID,
		bundle.Card,
		bundle.Skills,
		bundle.Personas,
		bundle.MCPs,
		session.ConversationChoice{},
	)
}

// SwitchAgentWithChoice changes cards and optionally resumes a conversation
// saved for that destination card.
func (s *Service) SwitchAgentWithChoice(sessionID, cardID string, choice session.ConversationChoice) error {
	bundle, err := s.resolveCard(cardID)
	if err != nil {
		return err
	}
	return s.sessions.SwitchAgentWithChoice(sessionID, bundle.Card, bundle.Skills, bundle.Personas, bundle.MCPs, choice)
}

// Conversations returns the saved bindings for one session/card pair.
func (s *Service) Conversations(sessionID, cardID string) ([]session.ConversationBinding, error) {
	snapshot, found := s.sessions.GetSession(sessionID)
	if !found {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}
	return append([]session.ConversationBinding(nil), snapshot.Conversations[cardID]...), nil
}

// DeleteCard refuses deletion while any session, including a terminated one, references it.
func (s *Service) DeleteCard(cardID string) error {
	var referencedBy []string
	for _, snapshot := range s.sessions.List() {
		_, hasBinding := snapshot.Conversations[cardID]
		if snapshot.CardID == cardID || hasBinding {
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

// RegisterConversation delegates an authenticated activation report to the
// manager, which validates that the credential still owns a live activation.
func (s *Service) RegisterConversation(token, conversationID, name string) (RegistrationResult, error) {
	return s.sessions.RegisterConversation(token, conversationID, name)
}

// HasActivationCredential validates an MCP bearer token without exposing its principal.
func (s *Service) HasActivationCredential(token string) bool {
	return s.sessions.HasActivationCredential(token)
}

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
	appendKankuroAccessSkill(card, &bundle, skills)
	return bundle, nil
}

func appendKankuroAccessSkill(card config.IDCard, bundle *cardBundle, skills []config.Skill) {
	if slices.Contains(card.SkillIDs, config.BuiltinSkillMCPCLI) {
		return
	}
	for _, mcp := range bundle.MCPs {
		if mcp.ID != config.BuiltinMCPKankuro {
			continue
		}
		for _, skill := range skills {
			if skill.ID == config.BuiltinSkillMCPCLI {
				bundle.Skills = append(bundle.Skills, skill)
				return
			}
		}
	}
}

func idSet(ids []string) map[string]struct{} {
	set := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		set[id] = struct{}{}
	}
	return set
}
