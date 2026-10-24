package session

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/driver"
)

// ErrPersistFailed marks a "started, but its record could not be saved"
// outcome: the agent process is live and the in-memory session is fully
// usable, but the on-disk copy did not update, so it will not survive a
// restart. Callers should use [errors.Is] to tell this apart from a launch
// that never started a process at all, and can still attach/use the session
// while surfacing this as a (non-blocking) warning.
var ErrPersistFailed = errors.New("session state could not be saved")

// ErrManagerClosed is returned by Launch, Relaunch and SwitchAgent once
// Shutdown has started: the manager no longer starts processes it could not
// stop or watch.
var ErrManagerClosed = errors.New("session manager is shut down")

var (
	ErrConversationChoiceInvalid  = errors.New("conversation choice is invalid")
	ErrConversationChoiceRequired = errors.New("conversation choice is required")
	ErrLifecycleConflict          = errors.New("session lifecycle conflict")
	ErrSessionNotFound            = errors.New("session not found")
	ErrWorkspaceUnavailable       = errors.New("saved session workspace is unavailable")
)

// syncDriverFiles clears temp directories (if needed) and syncs skill/persona/MCP files.
func (m *Manager) syncDriverFiles(
	d driver.AgentDriver,
	card config.IDCard,
	skills []config.Skill,
	personas []config.Persona,
	mcps []config.MCPServer,
	workingDir string,
	clearTmp bool,
	instructionMode driver.InstructionSyncMode,
) (driver.AgentDriver, bool, error) {
	if clearTmp {
		// TODO: if other names become relevant, implement a more elegant way of handling this
		agentDirs := []string{".agents", ".claude"}
		if d.GetAgentFolder() != agentDirs[0] && d.GetAgentFolder() != agentDirs[1] {
			agentDirs = append(agentDirs, d.GetAgentFolder())
		}
		for _, agentDir := range agentDirs {
			// Clear existing .agents directories in tmp workspace
			skillsDir := filepath.Join(workingDir, agentDir, "skills")
			if err := os.RemoveAll(skillsDir); err != nil {
				return nil, false, fmt.Errorf("clear tmp skills dir: %w", err)
			}
			personasDir := filepath.Join(workingDir, agentDir, "agents")
			if err := os.RemoveAll(personasDir); err != nil {
				return nil, false, fmt.Errorf("clear tmp personas dir: %w", err)
			}
		}
	}

	instructionsManaged, err := d.SetAgentsMD(card, personas, mcps, workingDir, instructionMode)
	if err != nil {
		return nil, false, fmt.Errorf("set agents.md: %w", err)
	}
	if skillsErr := d.SetSkills(skills, workingDir); skillsErr != nil {
		return nil, false, fmt.Errorf("set skills: %w", skillsErr)
	}
	if personasErr := d.SetPersonas(personas, workingDir); personasErr != nil {
		return nil, false, fmt.Errorf("set personas: %w", personasErr)
	}
	return d, instructionsManaged, nil
}

func (m *Manager) Launch(card config.IDCard, skills []config.Skill, personas []config.Persona, mcps []config.MCPServer) (string, error) {
	return m.LaunchWithID(uuid.NewString(), card, skills, personas, mcps)
}

// LaunchWithID starts a new session using an ID allocated by Kankuro's
// operation journal. External callers must not choose session identities.
func (m *Manager) LaunchWithID(
	sid string,
	card config.IDCard,
	skills []config.Skill,
	personas []config.Persona,
	mcps []config.MCPServer,
) (string, error) {
	return m.LaunchWithIDContext(context.Background(), sid, card, skills, personas, mcps)
}

// LaunchWithIDContext starts an activation that can be cancelled before publication.
func (m *Manager) LaunchWithIDContext(
	ctx context.Context,
	sid string,
	card config.IDCard,
	skills []config.Skill,
	personas []config.Persona,
	mcps []config.MCPServer,
) (string, error) {
	if _, err := uuid.Parse(sid); err != nil {
		return "", fmt.Errorf("invalid preallocated session ID: %w", err)
	}
	// The closed check and the Add share one critical section, the same
	// ordering argument as startWatchLocked: Shutdown sets closed under m.mu
	// before it Waits, so no Add can race a Wait that started at zero.
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return sid, ErrManagerClosed
	}
	if _, exists := m.sessions[sid]; exists {
		m.mu.Unlock()
		return sid, fmt.Errorf("session already exists: %s", sid)
	}
	if _, exists := m.launching[sid]; exists {
		m.mu.Unlock()
		return sid, fmt.Errorf("session launch already reserved: %s", sid)
	}
	m.launching[sid] = struct{}{}
	m.activations.Add(1)
	m.mu.Unlock()
	defer func() {
		m.mu.Lock()
		delete(m.launching, sid)
		m.mu.Unlock()
		m.activations.Done()
	}()
	return sid, m.startSession(ctx, sid, card, skills, personas, mcps, true, "", "", nil, "")
}

// activationSnapshot holds the session fields Relaunch/SwitchAgent need while
// the lock is not held, copied under m.mu so they are never read racily.
type activationSnapshot struct {
	WorkingDir           string
	OriginalDir          string
	UseTmpWorkspace      bool
	InstructionOwnership map[string]bool
	KankuroMCPEnabled    bool
	DashboardControl     bool
}

func validateConversationChoice(s *Session, cardID, driverName string, choice ConversationChoice) (string, error) {
	bindings := s.Conversations[cardID]
	if choice.New && choice.ID != "" {
		return "", fmt.Errorf("%w: choose either new or one saved conversation", ErrConversationChoiceInvalid)
	}
	if len(bindings) == 0 {
		if choice.ID != "" {
			return "", fmt.Errorf("%w: conversation is not saved for this card", ErrConversationChoiceInvalid)
		}
		return "", nil
	}
	if choice.New {
		return "", nil
	}
	if choice.ID == "" {
		return "", fmt.Errorf("%w: conversation selection required for this card; choose saved or start new", ErrConversationChoiceRequired)
	}
	hasIDForAnotherDriver := false
	for _, binding := range bindings {
		if binding.ID == choice.ID && binding.Driver == driverName {
			return binding.ID, nil
		}
		if binding.ID == choice.ID {
			hasIDForAnotherDriver = true
		}
	}
	if hasIDForAnotherDriver {
		return "", fmt.Errorf("%w: conversation belongs to a different driver", ErrConversationChoiceInvalid)
	}
	return "", fmt.Errorf("%w: conversation is not saved for this card", ErrConversationChoiceInvalid)
}

// reserveForActivation checks, under m.mu, that the session can be
// (re)activated and marks it as activating. It takes the lock itself; the
// caller must call releaseActivation on every failure path (success clears
// the mark in activateSessionLocked).
//
// On success it has also registered the activation with m.activations (in the
// same critical section as the !closed check), and the caller must call
// m.activations.Done() when the activation is over, after any
// releaseActivation, so Shutdown does not return while it is still working.
func (m *Manager) reserveForActivation(
	id, cardID, driverName, notTerminatedMsg string,
	choice ConversationChoice,
) (activationSnapshot, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return activationSnapshot{}, "", ErrManagerClosed
	}
	s, ok := m.sessions[id]
	if !ok {
		return activationSnapshot{}, "", fmt.Errorf("%w: %s", ErrSessionNotFound, id)
	}
	if s.Status != StatusTerminated {
		return activationSnapshot{}, "", fmt.Errorf("%w: %s", ErrLifecycleConflict, notTerminatedMsg)
	}
	if s.activating {
		return activationSnapshot{}, "", fmt.Errorf("%w: session is already being started", ErrLifecycleConflict)
	}
	if s.stopping || s.deleting {
		return activationSnapshot{}, "", fmt.Errorf("%w: another lifecycle operation is in progress", ErrLifecycleConflict)
	}
	conversationID, err := validateConversationChoice(s, cardID, driverName, choice)
	if err != nil {
		return activationSnapshot{}, "", err
	}
	m.revokeSessionActivationLocked(s)
	s.Registration = RegistrationUnavailable
	s.RegistrationPersistence = RegistrationNotAttempted
	s.RegistrationErr = ""
	s.ExpectedConversationID = ""
	s.activating = true
	m.activations.Add(1)
	return activationSnapshot{
		WorkingDir: s.WorkingDir, OriginalDir: s.OriginalDir,
		UseTmpWorkspace: s.UseTmpWorkspace, InstructionOwnership: sessionInstructionOwnership(s),
		KankuroMCPEnabled: s.KankuroMCPEnabled,
		DashboardControl:  s.DashboardControl,
	}, conversationID, nil
}

type instructionFileSnapshot struct {
	content []byte
	mode    os.FileMode
	exists  bool
}

func snapshotInstructionFile(path string) (instructionFileSnapshot, error) {
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return instructionFileSnapshot{}, nil
	}
	if err != nil {
		return instructionFileSnapshot{}, err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return instructionFileSnapshot{}, err
	}
	return instructionFileSnapshot{content: content, mode: info.Mode().Perm(), exists: true}, nil
}

func (s instructionFileSnapshot) restore(path string) error {
	if !s.exists {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("restore instruction file: %w", err)
		}
		return nil
	}
	if err := os.WriteFile(path, s.content, s.mode); err != nil {
		return fmt.Errorf("restore instruction file: %w", err)
	}
	return nil
}

// releaseActivation clears the reservation taken by reserveForActivation after
// an activation that did not happen.
func (m *Manager) releaseActivation(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s, ok := m.sessions[id]; ok {
		s.activating = false
		s.Registration = RegistrationFailed
		s.RegistrationErr = "activation failed before conversation registration"
	}
}

// Relaunch restarts a terminated session using its existing workspace (if tmp).
func (m *Manager) Relaunch(id string, card config.IDCard) error {
	return m.RelaunchWithChoice(id, card, ConversationChoice{})
}

func (m *Manager) RelaunchWithChoice(
	id string,
	card config.IDCard,
	choice ConversationChoice,
) error {
	return m.RelaunchWithChoiceContext(context.Background(), id, card, choice)
}

// RelaunchWithChoiceContext propagates cancellation through workspace preparation.
func (m *Manager) RelaunchWithChoiceContext(
	ctx context.Context,
	id string,
	card config.IDCard,
	choice ConversationChoice,
) error {
	snap, conversationID, err := m.reserveForActivation(id, card.ID, card.Driver, "session is already running", choice)
	if err != nil {
		return err
	}
	defer m.activations.Done()
	card.UseTmpWorkspace = snap.UseTmpWorkspace
	card.DashboardControl = snap.DashboardControl
	var mcps []config.MCPServer
	if snap.KankuroMCPEnabled {
		mcps = config.BuiltinMCPs()
	}
	err = m.startSession(
		ctx,
		id,
		card,
		nil,
		nil,
		mcps,
		false,
		snap.WorkingDir,
		snap.OriginalDir,
		snap.InstructionOwnership,
		conversationID,
	)
	// ErrPersistFailed means the activation did happen (and already cleared
	// the reservation); any other error means it did not.
	if err != nil && !errors.Is(err, ErrPersistFailed) {
		m.releaseActivation(id)
	}
	return err
}

// SwitchAgent restarts a terminated session using its existing workspace, but with a new agent card.
func (m *Manager) SwitchAgent(
	id string,
	card config.IDCard,
	skills []config.Skill,
	personas []config.Persona,
	mcps []config.MCPServer,
) error {
	return m.SwitchAgentWithChoice(id, card, skills, personas, mcps, ConversationChoice{})
}

func (m *Manager) SwitchAgentWithChoice(
	id string,
	card config.IDCard,
	skills []config.Skill,
	personas []config.Persona,
	mcps []config.MCPServer,
	choice ConversationChoice,
) error {
	return m.SwitchAgentWithChoiceContext(context.Background(), id, card, skills, personas, mcps, choice)
}

// SwitchAgentWithChoiceContext propagates cancellation until activation publication.
//
//nolint:gocognit,funlen // Reservation, rollback, spawn publication, and persistence form one activation transaction.
func (m *Manager) SwitchAgentWithChoiceContext(
	ctx context.Context,
	id string,
	card config.IDCard,
	skills []config.Skill,
	personas []config.Persona,
	mcps []config.MCPServer,
	choice ConversationChoice,
) (returnErr error) {
	snap, conversationID, err := m.reserveForActivation(
		id,
		card.ID,
		card.Driver,
		"session must be terminated before switching agents",
		choice,
	)
	if err != nil {
		return err
	}
	defer m.activations.Done() // registered first, so it runs after releaseActivation below
	activated := false
	defer func() {
		if !activated {
			m.releaseActivation(id)
		}
	}()

	finalWD := snap.WorkingDir
	if workspaceErr := validateSessionWorkspace(finalWD); workspaceErr != nil {
		return workspaceErr
	}
	if prepareErr := m.prepareSubWorkspaces(ctx, card, finalWD); prepareErr != nil {
		return prepareErr
	}
	subWorkspaces, subWorkspaceMetadataAvailable := preparedSubWorkspaceDescriptors(finalWD)

	// Clear tmp directories if using tmp workspace
	clearTmp := snap.UseTmpWorkspace
	// Replace content owned by the previous card. If it had no managed
	// instructions, an empty destination must preserve workspace-authored rules.
	d := m.drivers.Get(card.Driver)
	instructionFile := d.GetInstructionFileName()
	instructionMode := driver.InstructionReplace
	if !snap.InstructionOwnership[instructionFile] {
		instructionMode = driver.InstructionPreserveWhenEmpty
	}
	previousFile, previousFileErr := snapshotInstructionFile(filepath.Join(finalWD, instructionFile))
	if previousFileErr != nil {
		return previousFileErr
	}
	defer func() {
		if !activated {
			returnErr = errors.Join(returnErr, previousFile.restore(filepath.Join(finalWD, instructionFile)))
		}
	}()
	if ctxErr := ctx.Err(); ctxErr != nil {
		return ctxErr
	}
	d, instructionsManaged, err := m.syncDriverFiles(d, card, skills, personas, mcps, finalWD, clearTmp, instructionMode)
	if err != nil {
		return err
	}
	if ctxErr := ctx.Err(); ctxErr != nil {
		return ctxErr
	}

	token, tokenErr := m.prepareActivationMCPConfig(id, card, d, mcps, conversationID, finalWD, false)
	if tokenErr != nil {
		return tokenErr
	}
	published := false
	defer func() {
		if !published {
			m.revokeActivation(token)
		}
	}()
	if ctxErr := ctx.Err(); ctxErr != nil {
		return ctxErr
	}
	args := append(d.SetModel(card.Model), d.SetSession(conversationID)...)
	h, errSpawn := d.Spawn(driver.AgentConfig{
		Driver:            card.Driver,
		Model:             card.Model,
		Args:              args,
		WorkingDir:        finalWD,
		SystemInstruction: card.SystemInstruction,
		Env:               activationEnv(card.Env),
	})
	if errSpawn != nil {
		return errSpawn
	}

	now := time.Now().UTC()
	pid := 0
	if h.Cmd != nil && h.Cmd.Process != nil {
		pid = h.Cmd.Process.Pid
	}

	m.mu.Lock()
	if ctxErr := ctx.Err(); ctxErr != nil {
		m.mu.Unlock()
		return errors.Join(ctxErr, d.Kill())
	}
	if m.closed {
		// Shutdown started during the unlocked spawn. The process is already
		// running and no watcher will ever own it: kill it here, outside the
		// lock (Kill can block on teardown). The deferred releaseActivation
		// runs after this return.
		m.mu.Unlock()
		_ = d.Kill()
		return ErrManagerClosed
	}
	defer m.mu.Unlock()
	s, ok := m.sessions[id]
	if !ok {
		return fmt.Errorf("%w: %s", ErrSessionNotFound, id)
	}

	m.activateSessionLocked(s, card, d, h, pid, now)
	s.InstructionOwnership = updateInstructionOwnership(snap.InstructionOwnership, instructionFile, instructionsManaged)
	s.InstructionsManaged = instructionsManaged
	s.KankuroMCPEnabled = token != ""
	s.DashboardControl = card.DashboardControl && token != ""
	s.SubWorkspaces = subWorkspaces
	s.SubWorkspaceMetadataAvailable = subWorkspaceMetadataAvailable
	m.publishActivationLocked(token, s)
	if token == "" {
		s.ExpectedConversationID = conversationID
	}
	published = true
	activated = true
	s.PendingAttachHeader = fmt.Appendf(nil, "\r\n--- Switched agent to %s at %s ---\r\n\r\n",
		card.Name, now.Format(time.RFC3339))
	gen := s.generation

	go m.trackOutput(id, gen, h.Output)
	m.startWatchLocked(id, gen, d)

	// A persist failure here must not unwind the spawn: the agent process is
	// already running. Surface the error so the caller can report it, but
	// leave the in-memory session as-is. Wrapped in ErrPersistFailed so
	// callers can tell "did not start" apart from "started, but its record
	// did not save" and still offer the normal attach flow.
	if persistErr := m.persistSessionLockedErr(s); persistErr != nil {
		if s.Registration == RegistrationRegistered {
			s.RegistrationPersistence = RegistrationUnsaved
			s.RegistrationErr = conversationPersistenceWarning
		}
		return fmt.Errorf("%w: %w", ErrPersistFailed, persistErr)
	}
	return nil
}

// resolveWorkspacePath returns the override when given, else the path of the
// card's registered workspace.
func (m *Manager) resolveWorkspacePath(card config.IDCard, override string) (string, error) {
	if override != "" {
		return override, nil
	}
	workspaces, err := m.repo.ListWorkspaces()
	if err != nil {
		return "", fmt.Errorf("list workspaces: %w", err)
	}
	for _, w := range workspaces {
		if w.ID == card.WorkspaceID && w.Path != "" {
			return w.Path, nil
		}
	}
	return "", fmt.Errorf("workspace not found: %s", card.WorkspaceID)
}

func validateSessionWorkspace(path string) error {
	if path == "" {
		return errors.New("session workspace is empty")
	}
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("%w: %q; restore the directory before retrying: %w", ErrWorkspaceUnavailable, path, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%w: %q is not a directory; restore the directory before retrying", ErrWorkspaceUnavailable, path)
	}
	return nil
}

// startSession runs one activation. Its callers hold an m.activations
// registration (Launch, or reserveForActivation for Relaunch), which is also
// what guarantees Shutdown had not started when they began.
//
//nolint:funlen,gocognit // Workspace preparation, sync, spawn, publication, and persistence form one owned activation path.
func (m *Manager) startSession(
	ctx context.Context,
	sid string,
	card config.IDCard,
	skills []config.Skill,
	personas []config.Persona,
	mcps []config.MCPServer,
	isNew bool,
	workspaceOverride string,
	originalOverride string,
	previousInstructionOwnership map[string]bool,
	conversationID string,
) (returnErr error) {
	var wd, finalWD string
	if isNew {
		var err error
		wd, err = m.resolveWorkspacePath(card, workspaceOverride)
		if err != nil {
			return err
		}
		var errPrep error
		finalWD, errPrep = m.prepareWorkspace(ctx, sid, card, wd, true)
		if errPrep != nil {
			return m.cleanupNewTemporaryWorkspace(sid, card, true, errPrep)
		}
	} else {
		wd, finalWD = originalOverride, workspaceOverride
		if err := validateSessionWorkspace(finalWD); err != nil {
			return err
		}
	}
	d := m.drivers.Get(card.Driver)
	instructionFile := d.GetInstructionFileName()
	instructionsManaged := previousInstructionOwnership[instructionFile]
	committedInstructions := false
	var deployedSubWorkspaces []SubWorkspaceDescriptor
	subWorkspaceMetadataAvailable := false
	if isNew {
		var rollback func() error
		var prepareErr error
		instructionsManaged, rollback, prepareErr = m.prepareNewSessionFiles(ctx, d, card, skills, personas, mcps, finalWD)
		if prepareErr != nil {
			return m.cleanupNewTemporaryWorkspace(sid, card, true, prepareErr)
		}
		deployedSubWorkspaces, subWorkspaceMetadataAvailable = preparedSubWorkspaceDescriptors(finalWD)
		defer func() {
			if !committedInstructions {
				returnErr = errors.Join(returnErr, rollback())
			}
		}()
	}
	if ctxErr := ctx.Err(); ctxErr != nil {
		return m.cleanupNewTemporaryWorkspace(sid, card, isNew, ctxErr)
	}

	token, tokenErr := m.prepareActivationMCPConfig(sid, card, d, mcps, conversationID, finalWD, !isNew)
	if tokenErr != nil {
		return m.cleanupNewTemporaryWorkspace(sid, card, isNew, tokenErr)
	}
	published := false
	defer func() {
		if !published {
			m.revokeActivation(token)
		}
	}()
	if ctxErr := ctx.Err(); ctxErr != nil {
		return m.cleanupNewTemporaryWorkspace(sid, card, isNew, ctxErr)
	}
	args := append(d.SetModel(card.Model), d.SetSession(conversationID)...)
	if ctxErr := ctx.Err(); ctxErr != nil {
		return m.cleanupNewTemporaryWorkspace(sid, card, isNew, ctxErr)
	}
	h, errSpawn := d.Spawn(driver.AgentConfig{
		Driver:            card.Driver,
		Model:             card.Model,
		Args:              args,
		WorkingDir:        finalWD,
		SystemInstruction: card.SystemInstruction,
		Env:               activationEnv(card.Env),
	})
	if errSpawn != nil {
		return m.cleanupNewTemporaryWorkspace(sid, card, isNew, errSpawn)
	}

	now := time.Now().UTC()
	pid := 0
	if h.Cmd != nil && h.Cmd.Process != nil {
		pid = h.Cmd.Process.Pid
	}

	m.mu.Lock()
	if ctxErr := ctx.Err(); ctxErr != nil {
		m.mu.Unlock()
		return m.cleanupNewTemporaryWorkspace(sid, card, isNew, errors.Join(ctxErr, d.Kill()))
	}
	if m.closed {
		// Shutdown started during the unlocked spawn; see SwitchAgent. A new
		// session has no record, so nothing would ever point at its tmp copy.
		m.mu.Unlock()
		return m.cleanupNewTemporaryWorkspace(sid, card, isNew, errors.Join(ErrManagerClosed, d.Kill()))
	}
	defer m.mu.Unlock()

	s, exists := m.sessions[sid]
	if !exists {
		s = &Session{
			ID:          sid,
			Subscribers: make(map[chan []byte]struct{}),
		}
		m.sessions[sid] = s
	}

	m.activateSessionLocked(s, card, d, h, pid, now)
	s.InstructionOwnership = updateInstructionOwnership(previousInstructionOwnership, instructionFile, instructionsManaged)
	s.InstructionsManaged = instructionsManaged
	s.KankuroMCPEnabled = token != ""
	s.DashboardControl = card.DashboardControl && token != ""
	if isNew {
		s.SubWorkspaces = deployedSubWorkspaces
		s.SubWorkspaceMetadataAvailable = subWorkspaceMetadataAvailable
	}
	m.publishActivationLocked(token, s)
	if token == "" {
		s.ExpectedConversationID = conversationID
	}
	published = true
	committedInstructions = true
	s.WorkingDir = finalWD
	s.OriginalDir = wd
	s.UseTmpWorkspace = card.UseTmpWorkspace
	if isNew {
		s.Scrollback = nil
		s.PendingAttachHeader = nil
	} else {
		// Show the relaunch marker only to the next attacher (not stored in
		// scrollback so it doesn't replay on every subsequent reattach).
		s.PendingAttachHeader = fmt.Appendf(nil, "\r\n--- Relaunched at %s ---\r\n\r\n",
			now.Format(time.RFC3339))
	}
	gen := s.generation

	go m.trackOutput(sid, gen, h.Output)
	m.startWatchLocked(sid, gen, d)

	// A persist failure here must not unwind the spawn: the agent process is
	// already running. Surface the error so the caller can report it, but
	// leave the in-memory session as-is. Wrapped in ErrPersistFailed so
	// callers can tell "did not start" apart from "started, but its record
	// did not save" and still offer the normal attach flow.
	if persistErr := m.persistSessionLockedErr(s); persistErr != nil {
		if s.Registration == RegistrationRegistered {
			s.RegistrationPersistence = RegistrationUnsaved
			s.RegistrationErr = conversationPersistenceWarning
		}
		return fmt.Errorf("%w: %w", ErrPersistFailed, persistErr)
	}
	return nil
}

func (m *Manager) prepareNewSessionFiles(
	ctx context.Context,
	d driver.AgentDriver, card config.IDCard, skills []config.Skill,
	personas []config.Persona, mcps []config.MCPServer, workingDir string,
) (bool, func() error, error) {
	if err := m.prepareSubWorkspaces(ctx, card, workingDir); err != nil {
		return false, nil, err
	}
	instructionPath := filepath.Join(workingDir, d.GetInstructionFileName())
	previousFile, err := snapshotInstructionFile(instructionPath)
	if err != nil {
		return false, nil, err
	}
	rollback := func() error { return previousFile.restore(instructionPath) }
	mode := driver.InstructionPreserveWhenEmpty
	if card.UseTmpWorkspace {
		mode = driver.InstructionAppend
	}
	_, managed, err := m.syncDriverFiles(d, card, skills, personas, mcps, workingDir, card.UseTmpWorkspace, mode)
	if err != nil {
		return false, nil, errors.Join(err, rollback())
	}
	return managed, rollback, nil
}

func activationEnv(cardEnv map[string]string) map[string]string {
	env := make(map[string]string, len(cardEnv))
	for key, value := range cardEnv {
		if key == driver.EnvKankuroMCPToken || key == driver.EnvKankuroMCPConfigPath {
			continue
		}
		env[key] = value
	}
	return env
}

func (m *Manager) cleanupNewTemporaryWorkspace(sid string, card config.IDCard, isNew bool, cause error) error {
	if !isNew || !card.UseTmpWorkspace {
		return cause
	}
	return errors.Join(cause, os.RemoveAll(filepath.Join(m.tmpRoot(), sid)))
}

func (m *Manager) activateSessionLocked(
	s *Session,
	card config.IDCard,
	d driver.AgentDriver,
	h *driver.SessionHandle,
	pid int,
	now time.Time,
) {
	s.generation++
	s.activating = false
	s.CardID = card.ID
	if s.Name == "" {
		s.Name = card.Name
	}
	s.PID = pid
	s.StartedAt = now
	s.LastWordAt = now
	s.Status = StatusRunning
	s.Pinged = false
	s.Driver = d
	s.PTY = h.PTY
	s.Output = h.Output
	s.Done = h.Done
}

// persistSessionLockedErr writes s to disk and returns any error. Callers
// must hold m.mu.
func (m *Manager) persistSessionLockedErr(s *Session) error {
	sessionsDir := filepath.Join(m.configDir, "sessions")
	if err := saveSession(sessionsDir, s.data()); err != nil {
		return err
	}
	if s.Registration == RegistrationRegistered {
		s.RegistrationPersistence = RegistrationSaved
		s.RegistrationErr = ""
	}
	return nil
}

// persistSessionLocked is the background-safe variant: failures cannot
// propagate to a caller (e.g. the watch ticker), so they are logged instead
// of discarded. See debt_impl_plan_2.md step 3 for the logging decision:
// callers must never log unconditionally to stderr (it corrupts the
// rendered TUI frame), so this goes through the Manager's injected logger
// (WithLogger), which main.go points at the debug log file, or discards
// entirely when --debug is not set.
func (m *Manager) persistSessionLocked(s *Session) {
	if err := m.persistSessionLockedErr(s); err != nil {
		m.logger.Warn("session persist failed", "session_id", s.ID, "error", err)
	}
}

// Kill terminates one managed session. A teardown error is returned after the
// manager checks the driver's observed process state; an active process keeps
// its prior status and remains eligible for another stop attempt.
func (m *Manager) Kill(id string) error {
	_, err := m.KillDetailed(id)
	return err
}

type StopResult struct {
	ProcessTerminated      bool
	SessionRecordPersisted bool
}

// KillDetailed separates process teardown from session-record durability.
func (m *Manager) KillDetailed(id string) (StopResult, error) {
	return m.killDetailed(id, false)
}

func (m *Manager) killDetailed(id string, duringShutdown bool) (StopResult, error) {
	m.mu.Lock()
	if m.closed && !duringShutdown {
		m.mu.Unlock()
		return StopResult{}, ErrManagerClosed
	}
	s, ok := m.sessions[id]
	var d driver.AgentDriver
	var gen uint64
	if ok {
		if s.activating || s.stopping || s.deleting {
			m.mu.Unlock()
			return StopResult{}, fmt.Errorf("%w: %s", ErrLifecycleConflict, id)
		}
		if s.Status == StatusTerminated {
			persistErr := m.persistSessionLockedErr(s)
			m.mu.Unlock()
			return StopResult{ProcessTerminated: true, SessionRecordPersisted: persistErr == nil}, persistErr
		}
		s.stopping = true
		m.stops.Add(1)
		d, gen = s.Driver, s.generation
	}
	m.mu.Unlock()
	if !ok {
		return StopResult{}, fmt.Errorf("%w: %s", ErrSessionNotFound, id)
	}
	defer m.stops.Done()

	// Called with no lock held on purpose: it can block on process teardown,
	// and watch needs m.mu in the meantime.
	var killErr error
	if d != nil {
		killErr = d.Kill()
	}
	processTerminated := d == nil || d.GetStatus() == StatusTerminated

	m.mu.Lock()
	defer m.mu.Unlock()
	// Only mark terminated if this is still the same activation of the same
	// session. The lock was released above, so a Remove in that window must
	// not have its deleted record re-created by our persist, and a Relaunch
	// must not have its new, running activation marked dead.
	if cur, still := m.sessions[id]; !still || cur != s || s.generation != gen {
		return StopResult{ProcessTerminated: processTerminated}, killErr
	}
	s.stopping = false
	if !processTerminated {
		return StopResult{}, errors.Join(killErr, errors.New("session process remains active"))
	}
	s.Status = StatusTerminated
	s.Pinged = false
	if s.Registration == RegistrationPending {
		s.Registration = RegistrationUnconfirmed
	}
	m.revokeSessionActivationLocked(s)
	persistErr := m.persistSessionLockedErr(s)
	return StopResult{ProcessTerminated: true, SessionRecordPersisted: persistErr == nil}, errors.Join(killErr, persistErr)
}

// Remove deletes a terminated session from the registry.
func (m *Manager) Remove(id string) error {
	_, err := m.RemoveDetailed(id)
	return err
}

type RemovalResult struct {
	InMemoryRemoved           bool
	SessionRecordDeleted      bool
	TemporaryWorkspaceRemoved bool
}

// RemoveDetailed reports each deletion effect independently.
func (m *Manager) RemoveDetailed(id string) (RemovalResult, error) {
	m.mu.Lock()
	s, ok := m.sessions[id]
	if !ok {
		m.mu.Unlock()
		return RemovalResult{}, fmt.Errorf("%w: %s", ErrSessionNotFound, id)
	}
	if s.Status != StatusTerminated {
		m.mu.Unlock()
		return RemovalResult{}, fmt.Errorf("%w: session still running or waiting", ErrLifecycleConflict)
	}
	// A relaunch/switch in its unlocked phase is still Terminated; removing
	// now would delete the tmp workspace it is syncing into.
	if s.activating {
		m.mu.Unlock()
		return RemovalResult{}, fmt.Errorf("%w: session is activating", ErrLifecycleConflict)
	}
	if s.stopping || s.deleting {
		m.mu.Unlock()
		return RemovalResult{}, fmt.Errorf("%w: %s", ErrLifecycleConflict, id)
	}
	s.deleting = true
	m.revokeSessionActivationLocked(s)
	m.closeSubscribersLocked(id)
	tmp := s.UseTmpWorkspace
	m.mu.Unlock()

	sessionsDir := filepath.Join(m.configDir, "sessions")
	deleteErr := deleteSession(sessionsDir, id)
	// A file that is already gone is not a failure: the caller asked for it
	// not to exist, and it doesn't. Surfacing this as an error (as Plan 2
	// originally did) produced a red banner for a delete that did exactly
	// what was asked - the same shape as os.ErrProcessDone in Kill.
	if errors.Is(deleteErr, os.ErrNotExist) {
		deleteErr = nil
	}

	var rmErr error
	if tmp {
		sessionRoot := filepath.Join(m.tmpRoot(), id)
		// G703: Path traversal via taint analysis.
		// sessionRoot is derived from managed ID.
		rmErr = os.RemoveAll(sessionRoot)
	}

	// Always drop the in-memory session so the UI does not wedge on a row it
	// cannot delete, even if the on-disk cleanup below failed.
	m.mu.Lock()
	if m.sessions[id] == s {
		delete(m.sessions, id)
	}
	m.mu.Unlock()

	var errs []error
	if deleteErr != nil {
		errs = append(errs, fmt.Errorf("delete session file: %w", deleteErr))
	}
	if rmErr != nil {
		errs = append(errs, fmt.Errorf("remove tmp workspace: %w", rmErr))
	}
	return RemovalResult{InMemoryRemoved: true, SessionRecordDeleted: deleteErr == nil,
		TemporaryWorkspaceRemoved: tmp && rmErr == nil}, errors.Join(errs...)
}

// RunningCount reports how many sessions are running or waiting, so a caller
// can announce a Shutdown before it blocks.
func (m *Manager) RunningCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	n := 0
	for _, s := range m.sessions {
		if s.Status != StatusTerminated {
			n++
		}
	}
	return n
}

// Shutdown terminates all managed sessions, waits for activations that were
// already in flight to finish (each notices the shutdown at its final check,
// kills the process it just spawned and removes a new session's tmp copy), and
// stops the watchers. It returns only once no process it could still be
// responsible for is running and no watcher can write to disk any more. It is
// idempotent. It returns how many sessions were running or waiting when it
// started.
//
// An activation that passed its final check before closed was set is already
// Running; Shutdown collects its live list in the same critical section that
// sets closed, so it kills that session like any other.
//
// Sessions are killed concurrently: each Kill may spend a grace period on a
// graceful stop, and doing them one after another would multiply it by the
// session count. Kill comes before the watchers are stopped, so each record is
// persisted as Terminated by Kill itself; the watchers then exit on m.stop
// without their redundant final persist. One consequence: a watcher no longer
// records ExitCode after a Shutdown kill - for a killed process that is just
// the signal, so nothing is lost.
//
// Sessions already Terminated are left alone: their record is current, and
// rewriting every historic record on each quit would be pure churn.
func (m *Manager) Shutdown() int {
	m.mu.Lock()
	if m.shutdownDone != nil {
		done := m.shutdownDone
		select {
		case <-done:
			m.mu.Unlock()
			return 0
		default:
		}
		m.mu.Unlock()
		<-done
		return m.shutdownCount
	}
	m.shutdownDone = make(chan struct{})
	m.closed = true
	var live []string
	for id, s := range m.sessions {
		if s.Status != StatusTerminated {
			live = append(live, id)
		}
	}
	m.mu.Unlock()

	var wg sync.WaitGroup
	for _, id := range live {
		wg.Go(func() {
			if _, err := m.killDetailed(id, true); err != nil && !errors.Is(err, ErrLifecycleConflict) {
				m.logger.Warn("session shutdown stop failed", "session_id", id, "error", err)
			}
		})
	}
	wg.Wait()

	m.stops.Wait()
	m.activations.Wait()

	m.stopOnce.Do(func() { close(m.stop) })
	m.watchers.Wait()
	m.mu.Lock()
	m.shutdownCount = len(live)
	close(m.shutdownDone)
	m.mu.Unlock()
	return len(live)
}
