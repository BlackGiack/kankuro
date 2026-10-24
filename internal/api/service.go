// Package api contains UI-independent application operations shared by
// frontends. The session manager remains the owner of process lifecycle.
package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"

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

type workspaceLister interface {
	ListWorkspaces() ([]config.Workspace, error)
}

// DashboardScopeResolver isolates dashboard membership policy from lifecycle logic.
type DashboardScopeResolver interface {
	DashboardForSession(sessionID string) string
	AllowsSession(callerSessionID, targetSessionID string) bool
	AllowsCard(callerSessionID, cardID string) bool
}

type dashboardSessionAssigner interface {
	AssignSession(callerSessionID, targetSessionID, cardID string)
}

type dashboardScopeLookup interface {
	SessionDashboard(sessionID string) (dashboardID string, exists bool)
	CardDashboard(cardID string) (dashboardID string, exists bool)
}

type singleDashboardScope struct{}

func (singleDashboardScope) DashboardForSession(string) string { return DefaultDashboardID }
func (singleDashboardScope) AllowsSession(string, string) bool { return true }
func (singleDashboardScope) AllowsCard(string, string) bool    { return true }

// CardDiscovery contains only configuration fields useful for choosing a card.
type CardDiscovery struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	Driver            string   `json:"driver"`
	Model             string   `json:"model,omitempty"`
	WorkspaceID       string   `json:"workspace_id"`
	WorkspacePath     string   `json:"workspace_path"`
	SubWorkspaceIDs   []string `json:"sub_workspace_ids"`
	SubWorkspacePaths []string `json:"sub_workspace_paths"`
	UseTmpWorkspace   bool     `json:"use_tmp_workspace"`
}

// SessionDiscovery is a safe summary, excluding terminal output and process data.
type SessionDiscovery struct {
	ID       string         `json:"session_id"`
	Name     string         `json:"name"`
	CardID   string         `json:"card_id"`
	Status   session.Status `json:"status"`
	Activity string         `json:"activity"`
}

// SessionInspection returns durable workspace identity and saved conversation metadata.
type SessionInspection struct {
	SessionDiscovery

	WorkingDir                    string                                   `json:"working_dir"`
	OriginalDir                   string                                   `json:"original_dir"`
	UseTmpWorkspace               bool                                     `json:"use_tmp_workspace"`
	Conversations                 map[string][]session.ConversationBinding `json:"conversations"`
	SubWorkspaces                 []session.SubWorkspaceDescriptor         `json:"sub_workspaces"`
	SubWorkspaceMetadataAvailable bool                                     `json:"sub_workspace_metadata_available"`
	ActiveConversationID          string                                   `json:"active_conversation_id,omitempty"`
	Registration                  session.RegistrationState                `json:"registration"`
	RegistrationPersistence       session.RegistrationPersistence          `json:"registration_persistence"`
}

// CallerContext identifies the authenticated caller separately from discovered targets.
type CallerContext struct {
	SessionID        string `json:"session_id"`
	DashboardID      string `json:"dashboard_id"`
	DashboardControl bool   `json:"dashboard_control"`
}

const DefaultDashboardID = "default"

// Service coordinates configuration resolution with the existing session manager.
type Service struct {
	repo       Repository
	sessions   *session.Manager
	operations *OperationRunner
	scope      DashboardScopeResolver
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
	return &Service{repo: repo, sessions: sessions, scope: singleDashboardScope{}}
}

// SetDashboardScopeResolver replaces the initial single-dashboard policy.
func (s *Service) SetDashboardScopeResolver(scope DashboardScopeResolver) {
	if scope != nil {
		s.scope = scope
	}
}

// SetOperationRunner wires the durable lifecycle operation coordinator before MCP starts.
func (s *Service) SetOperationRunner(runner *OperationRunner) { s.operations = runner }

// Launch resolves capabilities from the supplied card snapshot and starts it.
func (s *Service) Launch(card config.IDCard) (string, error) {
	release, err := s.reserveCardReference(card.ID)
	if err != nil {
		return "", err
	}
	defer release()
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
	release, err := s.reserveCardReference(cardID)
	if err != nil {
		return err
	}
	defer release()
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
	release, err := s.reserveCardReference(cardID)
	if err != nil {
		return err
	}
	defer release()
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
	return s.sessions.DeleteCardIfUnreferenced(cardID, func() error {
		return s.repo.DeleteIDCard(cardID)
	})
}

func (s *Service) reserveCardReference(cardID string) (func(), error) {
	return s.sessions.ReserveCardReference(cardID, func() (bool, error) {
		_, err := s.loadCard(cardID)
		if errors.Is(err, errCardNotFound) {
			return false, nil
		}
		return err == nil, err
	})
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

// ResolveCaller derives lifecycle authority solely from the activation credential.
func (s *Service) ResolveCaller(token string) (CallerContext, error) {
	identity, err := s.sessions.ResolveDashboardCaller(token)
	if err != nil {
		return CallerContext{}, reject("forbidden", "caller activation is not authorized for dashboard control")
	}
	return CallerContext{
		SessionID: identity.SessionID, DashboardID: s.scope.DashboardForSession(identity.SessionID),
		DashboardControl: true,
	}, nil
}

func (s *Service) validateOperationScope(token string, operation Operation) error {
	caller, err := s.ResolveCaller(token)
	if err != nil {
		return err
	}
	if caller.DashboardID != operation.DashboardID {
		return errors.New("operation is outside the caller's current dashboard scope")
	}
	if operation.TargetSession != "" && operation.Kind != createSessionOperation &&
		!s.scope.AllowsSession(caller.SessionID, operation.TargetSession) {
		return errors.New("session is outside the caller's dashboard scope")
	}
	var cardID string
	switch operation.Kind {
	case createSessionOperation:
		var arguments createArguments
		if decodeErr := json.Unmarshal(operation.Arguments, &arguments); decodeErr != nil {
			return errors.New("invalid persisted operation arguments")
		}
		cardID = arguments.CardID
	case "switch_session_card":
		var arguments switchArguments
		if decodeErr := json.Unmarshal(operation.Arguments, &arguments); decodeErr != nil {
			return errors.New("invalid persisted operation arguments")
		}
		cardID = arguments.CardID
	case "relaunch_session":
		snapshot, found := s.sessions.GetSession(operation.TargetSession)
		if !found {
			return errors.New("session not found")
		}
		cardID = snapshot.CardID
	}
	if cardID != "" && !s.scope.AllowsCard(caller.SessionID, cardID) {
		return errors.New("card is outside the caller's dashboard scope")
	}
	return nil
}

func (s *Service) authorizeExistingOperation(caller CallerContext, operation Operation) error {
	if operation.DashboardID != caller.DashboardID {
		return reject("forbidden", "operation is outside the caller's dashboard scope")
	}
	lookup, hasLookup := s.scope.(dashboardScopeLookup)
	if operation.TargetSession != "" {
		if hasLookup {
			dashboard, exists := lookup.SessionDashboard(operation.TargetSession)
			if exists && dashboard != caller.DashboardID {
				return reject("forbidden", "operation target is outside the caller's dashboard scope")
			}
			// Missing targets are historical tombstones; they remain readable to
			// their originating dashboard, while a moved live target is denied.
		} else if !s.scope.AllowsSession(caller.SessionID, operation.TargetSession) {
			return reject("forbidden", "operation target is outside the caller's dashboard scope")
		}
	}
	cardID := operationCardID(operation)
	if cardID == "" {
		return nil
	}
	if hasLookup {
		dashboard, exists := lookup.CardDashboard(cardID)
		if exists && dashboard != caller.DashboardID {
			return reject("forbidden", "operation card is outside the caller's dashboard scope")
		}
		return nil
	}
	if !s.scope.AllowsCard(caller.SessionID, cardID) {
		return reject("forbidden", "operation card is outside the caller's dashboard scope")
	}
	return nil
}

func operationCardID(operation Operation) string {
	switch operation.Kind {
	case createSessionOperation:
		var args createArguments
		if json.Unmarshal(operation.Arguments, &args) == nil {
			return args.CardID
		}
	case "switch_session_card":
		var args switchArguments
		if json.Unmarshal(operation.Arguments, &args) == nil {
			return args.CardID
		}
	}
	return ""
}

// ListCards authenticates and restricts destination cards in the shared service.
func (s *Service) ListCards(token string) ([]CardDiscovery, error) {
	caller, err := s.ResolveCaller(token)
	if err != nil {
		return nil, err
	}
	return s.listCardsForCaller(caller)
}

func (s *Service) listCardsForCaller(caller CallerContext) ([]CardDiscovery, error) {
	cards, err := s.repo.ListIDCards()
	if err != nil {
		return nil, fmt.Errorf("list cards: %w", err)
	}
	workspaces, ok := s.repo.(workspaceLister)
	if !ok {
		return nil, errors.New("workspace discovery is unavailable")
	}
	items, err := workspaces.ListWorkspaces()
	if err != nil {
		return nil, fmt.Errorf("list workspaces: %w", err)
	}
	byID := make(map[string]string, len(items))
	for _, item := range items {
		byID[item.ID] = item.Path
	}
	out := make([]CardDiscovery, 0, len(cards))
	for _, card := range cards {
		if caller.SessionID != "" && !s.scope.AllowsCard(caller.SessionID, card.ID) {
			continue
		}
		entry := CardDiscovery{ID: card.ID, Name: card.Name, Driver: card.Driver, Model: card.Model,
			WorkspaceID: card.WorkspaceID, WorkspacePath: byID[card.WorkspaceID],
			SubWorkspaceIDs: append([]string(nil), card.SubWorkspaceIDs...), UseTmpWorkspace: card.UseTmpWorkspace}
		for _, id := range card.SubWorkspaceIDs {
			entry.SubWorkspacePaths = append(entry.SubWorkspacePaths, byID[id])
		}
		out = append(out, entry)
	}
	return out, nil
}

// ListSessions returns summaries in the current dashboard, excluding the caller.
// ListSessions authenticates the caller before returning dashboard summaries.
func (s *Service) ListSessions(token string) ([]SessionDiscovery, error) {
	caller, err := s.ResolveCaller(token)
	if err != nil {
		return nil, err
	}
	out := make([]SessionDiscovery, 0)
	for _, item := range s.sessions.List() {
		if item.ID == caller.SessionID || !s.scope.AllowsSession(caller.SessionID, item.ID) {
			continue
		}
		out = append(out, SessionDiscovery{
			ID: item.ID, Name: item.Name, CardID: item.CardID,
			Status: item.Status, Activity: item.LifecycleActivity,
		})
	}
	return out, nil
}

// GetSession authenticates and authorizes a non-caller target inside the service.
func (s *Service) GetSession(token, id string) (SessionInspection, error) {
	caller, err := s.ResolveCaller(token)
	if err != nil {
		return SessionInspection{}, err
	}
	if id == caller.SessionID {
		return SessionInspection{}, errors.New("cannot inspect caller session")
	}
	if !s.scope.AllowsSession(caller.SessionID, id) {
		return SessionInspection{}, errors.New("session not found")
	}
	item, ok := s.sessions.GetSession(id)
	if !ok {
		return SessionInspection{}, errors.New("session not found")
	}
	subWorkspaces := item.SubWorkspaces
	metadataAvailable := item.SubWorkspaceMetadataAvailable
	if !metadataAvailable {
		subWorkspaces, metadataAvailable = session.ReadSubWorkspaceDescriptors(item.WorkingDir)
	}
	if subWorkspaces == nil {
		subWorkspaces = []session.SubWorkspaceDescriptor{}
	}
	conversations := item.Conversations
	if conversations == nil {
		conversations = map[string][]session.ConversationBinding{}
	}
	return SessionInspection{
		ID: item.ID, Name: item.Name, CardID: item.CardID,
		Status: item.Status, Activity: item.LifecycleActivity,
		WorkingDir: item.WorkingDir, OriginalDir: item.OriginalDir,
		UseTmpWorkspace: item.UseTmpWorkspace, Conversations: conversations,
		SubWorkspaces: subWorkspaces, SubWorkspaceMetadataAvailable: metadataAvailable,
		ActiveConversationID: item.ActiveConversationID, Registration: item.Registration,
		RegistrationPersistence: item.RegistrationPersistence,
	}, nil
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
	return config.IDCard{}, fmt.Errorf("%w: %s", errCardNotFound, cardID)
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
