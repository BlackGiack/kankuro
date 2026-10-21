package app

import (
	"errors"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
)

// sessionActions runs blocking session work (capability resolution, spawn)
// inside tea.Cmd closures. Like refresher, it holds no reference to the
// Coordinator or its views, so a closure it builds can only reach repo,
// sessions and the values passed in as arguments: invariant 1 holds here by
// construction. Anything the work needs from coordinator state (a view, the
// settings) must be read by the caller *before* and passed in as a value.
type sessionActions struct {
	repo     config.Repository
	sessions *session.Manager
}

// launch resolves card's capabilities and spawns a session, reporting the
// outcome as a SessionLaunchCompletedMsg.
func (a sessionActions) launch(card config.IDCard) tea.Cmd {
	return func() tea.Msg {
		// Resolve capabilities inside the Cmd (off the UI goroutine) -
		// each registry is read once here rather than by three
		// separate LoadXxxForCard calls, and a failed read now aborts
		// the launch instead of silently starting the agent with no
		// skills/personas/MCPs (see debt_impl_plan_5.md step 2).
		skills, personas, mcps, err := a.resolveCapabilities(card)
		if err != nil {
			return messages.SessionLaunchCompletedMsg{Err: err}
		}
		id, err := a.sessions.Launch(card, skills, personas, mcps)
		return messages.SessionLaunchCompletedMsg{SessionID: id, Err: err}
	}
}

// relaunch restarts a terminated session with its own card and reports a
// SessionAttachRequestedMsg (with a Warning when only the record failed to
// save) or an ErrorMsg.
func (a sessionActions) relaunch(sessionID string) tea.Cmd {
	return func() tea.Msg {
		var sess *session.Snapshot
		for _, s := range a.sessions.List() {
			if s.ID == sessionID {
				sess = &s
				break
			}
		}
		if sess == nil {
			return messages.ErrorMsg{Err: fmt.Errorf("session not found: %s", sessionID)}
		}
		card, err := a.loadCard(sess.CardID)
		if err != nil {
			return messages.ErrorMsg{Err: err}
		}
		relaunchErr := a.sessions.Relaunch(sessionID, card)
		if relaunchErr != nil {
			if !errors.Is(relaunchErr, session.ErrPersistFailed) {
				return messages.ErrorMsg{Err: relaunchErr}
			}
			// The process did restart; only its on-disk record failed to
			// save. Still attach, but let the user know.
			return messages.SessionAttachRequestedMsg{SessionID: sessionID, Warning: relaunchErr}
		}
		return messages.SessionAttachRequestedMsg{SessionID: sessionID}
	}
}

// switchAgent performs the actual (blocking) agent switch - it syncs driver
// files to disk and spawns a new PTY process - and reports the outcome as an
// AgentSwitchedMsg. The caller must hand the returned Cmd to Bubble Tea,
// never run it synchronously inside Update. A synchronous wrapper used to
// exist for DashboardCallbacks.OnSwitch, which the real Dashboard never
// invoked; it was deleted (review_dp_4_2.md finding 4) rather than left as a
// dead, reachable-by-one-wiring-change copy of the blocking-work-inside-Update
// bug.
func (a sessionActions) switchAgent(sessionID, newCardID string) tea.Cmd {
	return func() tea.Msg {
		bundle, err := a.resolveCard(newCardID)
		if err == nil {
			err = a.sessions.SwitchAgent(sessionID, bundle.Card, bundle.Skills, bundle.Personas, bundle.MCPs)
		}
		return messages.AgentSwitchedMsg{SessionID: sessionID, Err: err}
	}
}

// deleteCard removes an ID card, refusing when a persisted session still
// references it. DeleteIDCard itself has no reference guard (internal/config
// cannot see sessions), so the check lives here where the session manager is
// in hand: a card a session still points at cannot be deleted, or that
// session's relaunch/switch would break with a bare "card not found".
func (a sessionActions) deleteCard(cardID string) tea.Cmd {
	return func() tea.Msg {
		var referencedBy []string
		for _, s := range a.sessions.List() {
			if s.CardID == cardID {
				referencedBy = append(referencedBy, s.Name)
			}
		}
		if len(referencedBy) > 0 {
			return messages.ErrorMsg{Err: fmt.Errorf(
				"cannot delete card: referenced by session(s) %s",
				strings.Join(referencedBy, ", "),
			)}
		}
		if err := a.repo.DeleteIDCard(cardID); err != nil {
			return messages.ErrorMsg{Err: err}
		}
		return messages.DataRefreshRequestedMsg{}
	}
}

// deleteMCP removes an MCP server off the UI goroutine - disk I/O deferred
// to a tea.Cmd, consistent with every other mutation (it used to run
// synchronously inside Update).
func (a sessionActions) deleteMCP(m messages.MCPDeleteRequestedMsg) tea.Cmd {
	return func() tea.Msg {
		if err := a.repo.DeleteMCP(m.MCPID); err != nil {
			return messages.ErrorMsg{Err: err}
		}
		return messages.MCPDeletedMsg(m)
	}
}

// launchIDE spawns ideCommand on a session's workspace, so the (potentially
// slow) process spawn runs off the UI goroutine; completion is reported via
// messages.IDELaunchedMsg. ideCommand is passed in by value because the
// settings it comes from are coordinator-goroutine state: reading them
// inside the closure would race with SettingsSavedMsg.
func (a sessionActions) launchIDE(sessionID, ideCommand string) tea.Cmd {
	return func() tea.Msg {
		return messages.IDELaunchedMsg{Err: a.launchWorkspaceCommand(sessionID, ideCommand, session.ErrIDENotConfigured)}
	}
}

func (a sessionActions) launchTerminal(sessionID, command string) tea.Cmd {
	return func() tea.Msg {
		notConfigured := errors.New(strings.Join([]string{"Terminal", "command", "not configured"}, " "))
		err := a.launchWorkspaceCommand(sessionID, command, notConfigured)
		return messages.TerminalLaunchedMsg{Err: err}
	}
}

func (a sessionActions) launchWorkspaceCommand(sessionID, command string, notConfigured error) error {
	for _, sess := range a.sessions.List() {
		if sess.ID != sessionID {
			continue
		}
		if sess.WorkingDir == "" {
			return errors.New("session has no workspace path")
		}
		if command == "" || (!errors.Is(notConfigured, session.ErrIDENotConfigured) && strings.TrimSpace(command) == "") {
			return notConfigured
		}
		if errors.Is(notConfigured, session.ErrIDENotConfigured) {
			return session.NewIDELauncher(command).Launch(sess.WorkingDir)
		}
		return session.LaunchWorkspaceCommand(command, sess.WorkingDir)
	}
	return fmt.Errorf("session not found: %s", sessionID)
}

// cardBundle groups an ID card with its resolved capabilities, so launch and
// switch only have to read each registry once - see
// debt_impl_plan_5.md step 2.
type cardBundle struct {
	Card     config.IDCard
	Skills   []config.Skill
	Personas []config.Persona
	MCPs     []config.MCPServer
}

// resolveCard reads the card registry once, finds cardID, and resolves its
// capabilities via resolveCapabilities. It folds the old getCardByID's
// not-found lookup together with propagating a failed registry read (step
// 3): a read failure now surfaces as "list cards: ...", not the misleading
// "card not found" a discarded error used to produce.
func (a sessionActions) resolveCard(cardID string) (cardBundle, error) {
	card, err := a.loadCard(cardID)
	if err != nil {
		return cardBundle{}, err
	}
	skills, personas, mcps, err := a.resolveCapabilities(card)
	if err != nil {
		return cardBundle{}, err
	}
	return cardBundle{Card: card, Skills: skills, Personas: personas, MCPs: mcps}, nil
}

func (a sessionActions) loadCard(cardID string) (config.IDCard, error) {
	cards, err := a.repo.ListIDCards()
	if err != nil {
		return config.IDCard{}, fmt.Errorf("list cards: %w", err)
	}
	var card config.IDCard
	found := false
	for _, cd := range cards {
		if cd.ID == cardID {
			card = cd
			found = true
			break
		}
	}
	if !found {
		return config.IDCard{}, fmt.Errorf("card not found: %s", cardID)
	}
	return card, nil
}

// resolveCapabilities reads the skill/persona/MCP registries once each and
// filters them down to card's referenced IDs, iterating each registry in its
// own (name-sorted) order so results come out in registry order rather than
// card-ID order - preserving the order previously written to disk/AGENTS.md
// by the three separate ListXxx-then-nested-loop calls this replaces. Unlike
// those calls, a failed registry read here aborts (returns an error) instead
// of silently launching/relaunching/switching with no capabilities.
func (a sessionActions) resolveCapabilities(card config.IDCard) ([]config.Skill, []config.Persona, []config.MCPServer, error) {
	allSkills, err := a.repo.ListSkills()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("list skills: %w", err)
	}
	allPersonas, err := a.repo.ListPersonas()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("list personas: %w", err)
	}
	allMCPs, err := a.repo.ListMCPs()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("list mcps: %w", err)
	}

	skillIDs := toIDSet(card.SkillIDs)
	personaIDs := toIDSet(card.PersonaIDs)
	mcpIDs := toIDSet(card.MCPIDs)

	var skills []config.Skill
	for _, s := range allSkills {
		if _, ok := skillIDs[s.ID]; ok {
			skills = append(skills, s)
		}
	}
	var personas []config.Persona
	for _, p := range allPersonas {
		if _, ok := personaIDs[p.ID]; ok {
			personas = append(personas, p)
		}
	}
	var mcps []config.MCPServer
	for _, m := range allMCPs {
		if _, ok := mcpIDs[m.ID]; ok {
			mcps = append(mcps, m)
		}
	}
	return skills, personas, mcps, nil
}

func toIDSet(ids []string) map[string]struct{} {
	set := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		set[id] = struct{}{}
	}
	return set
}
