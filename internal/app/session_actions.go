package app

import (
	"errors"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/api"
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
	service  *api.Service
}

// launch resolves card's capabilities and spawns a session, reporting the
// outcome as a SessionLaunchCompletedMsg.
func (a sessionActions) launch(card config.IDCard) tea.Cmd {
	return func() tea.Msg {
		id, err := a.service.Launch(card)
		return messages.SessionLaunchCompletedMsg{SessionID: id, Err: err}
	}
}

// relaunch restarts a terminated session with its own card and reports a
// SessionAttachRequestedMsg (with a Warning when only the record failed to
// save) or an ErrorMsg.
func (a sessionActions) relaunch(sessionID string) tea.Cmd {
	return func() tea.Msg {
		relaunchErr := a.service.Relaunch(sessionID)
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
		err := a.service.SwitchAgent(sessionID, newCardID)
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
		if err := a.service.DeleteCard(cardID); err != nil {
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
