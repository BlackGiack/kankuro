package session

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"maps"
	"strings"
	"time"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/driver"
	"github.com/google/uuid"
)

const (
	maxConversationIDLength        = 512
	maxConversationNameLength      = 128
	conversationPersistenceWarning = "registered, but saving conversation metadata failed"
)

type activationAuthority struct {
	sessionID    string
	activationID string
	cardID       string
	driver       string
	expectedID   string
	configPath   string
	acceptedID   string
	failure      string
	published    bool
	revoked      bool
	staged       *ConversationBinding
}

// CallerIdentity is the manager-derived identity of a published activation.
// It intentionally contains no credential and is not available during startup.
type CallerIdentity struct {
	SessionID        string
	ActivationID     string
	CardID           string
	Driver           string
	DashboardControl bool
}

func (m *Manager) prepareActivationMCPConfig(
	sessionID string,
	card config.IDCard,
	d driver.AgentDriver,
	mcps []config.MCPServer,
	expectedID string,
	workingDir string,
	preserveDeployed bool,
) (string, error) {
	resolvedMCPs := append([]config.MCPServer(nil), mcps...)
	hasKankuro := false
	for i := range resolvedMCPs {
		if resolvedMCPs[i].ID != config.BuiltinMCPKankuro {
			continue
		}
		hasKankuro = true
	}
	token := ""
	activationID := ""
	if hasKankuro {
		var err error
		token, err = newActivationToken()
		if err != nil {
			return "", err
		}
		activationID = uuid.NewString()
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return "", ErrManagerClosed
	}
	if hasKankuro {
		if m.mcpEndpoint == "" {
			return "", errors.New("kankuro MCP endpoint is not configured")
		}
		for i := range resolvedMCPs {
			if resolvedMCPs[i].ID != config.BuiltinMCPKankuro {
				continue
			}
			resolvedMCPs[i].URL = m.mcpEndpoint
			headers := make(map[string]string, len(resolvedMCPs[i].Headers)+1)
			maps.Copy(headers, resolvedMCPs[i].Headers)
			headers["Authorization"] = "Bearer " + token
			resolvedMCPs[i].Headers = headers
		}
	}
	if err := syncActivationMCPConfig(d, resolvedMCPs, workingDir, preserveDeployed); err != nil {
		return "", fmt.Errorf("select MCPs: %w", err)
	}

	if hasKankuro {
		m.authorities[token] = &activationAuthority{
			sessionID:    sessionID,
			activationID: activationID,
			cardID:       card.ID,
			driver:       card.Driver,
			expectedID:   expectedID,
			configPath:   driver.MCPConfigPath(workingDir),
		}
	}
	return token, nil
}

func syncActivationMCPConfig(d driver.AgentDriver, mcps []config.MCPServer, workingDir string, preserveDeployed bool) error {
	if !preserveDeployed {
		return d.SelectMCPs(mcps, workingDir)
	}
	for _, server := range mcps {
		if server.ID == config.BuiltinMCPKankuro {
			return driver.RefreshKankuroMCP(server, workingDir)
		}
	}
	return nil
}

func (m *Manager) revokeActivation(token string) {
	if token == "" {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.revokeAuthorityLocked(token)
}

func (m *Manager) revokeSessionActivationLocked(s *Session) {
	m.revokeAuthorityLocked(s.ActivationToken)
	s.ActivationToken = ""
	if m.messages != nil {
		m.messages.Wake()
	}
}

func (m *Manager) revokeAuthorityLocked(token string) {
	if token == "" {
		return
	}
	authority := m.authorities[token]
	if authority == nil {
		return
	}
	authority.revoked = true
	delete(m.authorities, token)
	if err := driver.RemoveKankuroMCPIfAuthorizationMatches(authority.configPath, token); err != nil {
		m.logger.Warn("workspace MCP config cleanup failed", "session_id", authority.sessionID,
			"warning", "revoked Kankuro MCP entry could not be safely removed")
	}
}

func (m *Manager) publishActivationLocked(token string, s *Session) {
	if token == "" {
		s.Registration = RegistrationUnavailable
		s.RegistrationPersistence = RegistrationNotAttempted
		s.RegistrationErr = ""
		s.ActivationID = ""
		s.ActivationToken = ""
		s.ExpectedConversationID = ""
		return
	}
	authority := m.authorities[token]
	if authority == nil || authority.revoked {
		s.Registration = RegistrationFailed
		return
	}
	authority.published = true
	s.ActivationID = authority.activationID
	s.ActivationToken = token
	s.ExpectedConversationID = authority.expectedID
	s.Registration = RegistrationPending
	s.RegistrationPersistence = RegistrationNotAttempted
	s.RegistrationAt = time.Now().UTC()
	if authority.failure != "" {
		s.Registration = RegistrationFailed
		s.RegistrationErr = authority.failure
	}
	if authority.staged != nil {
		addConversationBinding(s, authority.cardID, *authority.staged)
		s.ExpectedConversationID = authority.staged.ID
		s.Registration = RegistrationRegistered
		s.RegistrationPersistence = RegistrationNotAttempted
		if authority.failure != "" {
			s.RegistrationErr = authority.failure
		}
	}
}

type RegistrationResult struct {
	Status      RegistrationState       `json:"status"`
	Persistence RegistrationPersistence `json:"persistence"`
	Warning     string                  `json:"warning,omitempty"`
}

func newActivationToken() (string, error) {
	var bytes [32]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", fmt.Errorf("generate activation credential: %w", err)
	}
	return hex.EncodeToString(bytes[:]), nil
}

func validConversation(conversationID, name string) error {
	if strings.TrimSpace(conversationID) == "" || len(conversationID) > maxConversationIDLength {
		return errors.New("conversation ID must be non-empty and at most 512 bytes")
	}
	if strings.TrimSpace(name) == "" || len(name) > maxConversationNameLength {
		return errors.New("conversation name must be non-empty and at most 128 bytes")
	}
	return nil
}

// RegisterConversation binds an authenticated activation's native conversation
// to its own session/card. The credential, not the tool arguments, establishes
// caller identity.
func (m *Manager) RegisterConversation(token, conversationID, name string) (RegistrationResult, error) {
	if err := validConversation(conversationID, name); err != nil {
		m.recordRegistrationFailure(token, "conversation registration input was invalid")
		return RegistrationResult{}, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	authority, err := m.activationAuthorityLocked(token)
	if err != nil {
		return RegistrationResult{}, err
	}
	if validateErr := m.validateReportedConversationLocked(authority, conversationID); validateErr != nil {
		return RegistrationResult{}, validateErr
	}
	binding := ConversationBinding{ID: conversationID, Name: name, Driver: authority.driver}
	return m.registerConversationLocked(authority, binding)
}

func (m *Manager) activationAuthorityLocked(token string) (*activationAuthority, error) {
	if m.closed {
		return nil, ErrManagerClosed
	}
	authority := m.authorities[token]
	if authority == nil || authority.revoked {
		return nil, errors.New("activation credential is stale")
	}
	return authority, nil
}

func (m *Manager) validateReportedConversationLocked(authority *activationAuthority, conversationID string) error {
	if authority.expectedID == "" || authority.expectedID == conversationID {
		return nil
	}
	if authority.acceptedID == "" {
		m.setRegistrationFailureLocked(authority, "reported conversation did not match the requested resume")
	}
	return errors.New("reported conversation does not match the requested resume")
}

func (m *Manager) registerConversationLocked(authority *activationAuthority, binding ConversationBinding) (RegistrationResult, error) {
	if authority.acceptedID != "" && authority.acceptedID != binding.ID {
		m.setRegistrationFailureLocked(authority, "activation rejected a conflicting conversation identity")
		return RegistrationResult{}, errors.New("activation already reported a different conversation")
	}
	authority.acceptedID = binding.ID
	authority.failure = ""
	if !authority.published {
		authority.staged = &binding
		return RegistrationResult{Status: "pending_activation", Persistence: RegistrationNotAttempted}, nil
	}
	s := m.sessions[authority.sessionID]
	if s == nil || s.ActivationID != authority.activationID || s.Status == StatusTerminated {
		return RegistrationResult{}, errors.New("activation is no longer live")
	}
	addConversationBinding(s, authority.cardID, binding)
	s.ExpectedConversationID = binding.ID
	s.Registration = RegistrationRegistered
	s.RegistrationPersistence = RegistrationNotAttempted
	s.RegistrationErr = ""
	if err := m.persistSessionLockedErr(s); err != nil {
		s.RegistrationPersistence = RegistrationUnsaved
		s.RegistrationErr = conversationPersistenceWarning
		// A live activation with unsaved metadata is a successful registration outcome.
		//nolint:nilerr // The transport returns durability in the structured result.
		return RegistrationResult{Status: RegistrationRegistered, Persistence: RegistrationUnsaved, Warning: s.RegistrationErr}, nil
	}
	return RegistrationResult{Status: RegistrationRegistered, Persistence: RegistrationSaved}, nil
}

func (m *Manager) setRegistrationFailureLocked(authority *activationAuthority, message string) {
	authority.failure = message
	if !authority.published {
		return
	}
	s := m.sessions[authority.sessionID]
	if s == nil || s.ActivationID != authority.activationID {
		return
	}
	s.RegistrationErr = message
	if authority.acceptedID == "" {
		s.Registration = RegistrationFailed
	}
}

func (m *Manager) recordRegistrationFailure(token, message string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return
	}
	authority := m.authorities[token]
	if authority == nil || authority.revoked || authority.acceptedID != "" {
		return
	}
	authority.failure = message
	if authority.published {
		if s := m.sessions[authority.sessionID]; s != nil && s.ActivationID == authority.activationID {
			s.Registration = RegistrationFailed
			s.RegistrationErr = message
		}
	}
}

// HasActivationCredential validates transport credentials without exposing
// activation identity to the transport layer.
func (m *Manager) HasActivationCredential(token string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	authority, ok := m.authorities[token]
	return !m.closed && ok && !authority.revoked
}

// ResolveDashboardCaller validates a current published activation and returns
// only its safe identity. Lifecycle discovery/control requires both an active
// Kankuro credential and the permission deployed by that activation.
func (m *Manager) ResolveDashboardCaller(token string) (CallerIdentity, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.closed {
		return CallerIdentity{}, ErrManagerClosed
	}
	authority := m.authorities[token]
	if authority == nil || authority.revoked || !authority.published {
		return CallerIdentity{}, errors.New("activation credential is stale or unpublished")
	}
	s := m.sessions[authority.sessionID]
	if s == nil || s.ActivationID != authority.activationID || s.Status == StatusTerminated {
		return CallerIdentity{}, errors.New("activation is no longer live")
	}
	if !s.DashboardControl {
		return CallerIdentity{}, errors.New("dashboard control is not enabled for this activation")
	}
	return CallerIdentity{
		SessionID:        authority.sessionID,
		ActivationID:     authority.activationID,
		CardID:           authority.cardID,
		Driver:           authority.driver,
		DashboardControl: true,
	}, nil
}

func addConversationBinding(s *Session, cardID string, binding ConversationBinding) {
	if s.Conversations == nil {
		s.Conversations = make(map[string][]ConversationBinding)
	}
	bindings := s.Conversations[cardID]
	for i := range bindings {
		if bindings[i].Driver == binding.Driver && bindings[i].ID == binding.ID {
			bindings[i].Name = binding.Name
			s.Conversations[cardID] = bindings
			return
		}
	}
	s.Conversations[cardID] = append(bindings, binding)
}
