package app

import (
	"context"
	"errors"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/api"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
)

func (a sessionActions) launch(dashboardID string, generation uint64, cardID string) tea.Cmd {
	return func() tea.Msg {
		id, err := a.service.LaunchInDashboard(context.Background(), dashboardID, cardID)
		return messages.SessionLaunchCompletedMsg{SessionID: id, Err: err, DashboardID: dashboardID, Generation: generation}
	}
}

func (a sessionActions) relaunch(dashboardID string, generation uint64, sessionID string, choice session.ConversationChoice) tea.Cmd {
	return func() tea.Msg {
		err := a.service.RelaunchInDashboard(context.Background(), dashboardID, sessionID, choice)
		if err != nil && !errors.Is(err, session.ErrPersistFailed) {
			return messages.ScopedActionFinishedMsg{DashboardID: dashboardID, Generation: generation, Err: err}
		}
		return messages.SessionAttachRequestedMsg{SessionID: sessionID, Warning: err, DashboardID: dashboardID, Generation: generation}
	}
}

func (a sessionActions) switchAgent(
	dashboardID string,
	generation uint64,
	sessionID, cardID string,
	choice session.ConversationChoice,
) tea.Cmd {
	return func() tea.Msg {
		err := a.service.SwitchInDashboard(context.Background(), dashboardID, sessionID, cardID, choice)
		return messages.AgentSwitchedMsg{SessionID: sessionID, Err: err, DashboardID: dashboardID, Generation: generation}
	}
}

func (a sessionActions) launchIDE(dashboardID string, generation uint64, sessionID, command string) tea.Cmd {
	return a.launchWorkspace(dashboardID, generation, sessionID, command, session.ErrIDENotConfigured)
}

func (a sessionActions) launchTerminal(dashboardID string, generation uint64, sessionID, command string) tea.Cmd {
	return a.launchWorkspace(
		dashboardID,
		generation,
		sessionID,
		command,
		errors.New(strings.Join([]string{"Terminal", "command", "not configured"}, " ")),
	)
}

func (a sessionActions) launchWorkspace(
	dashboardID string,
	generation uint64,
	sessionID, command string,
	notConfigured error,
) tea.Cmd {
	return func() tea.Msg {
		err := a.service.CheckSessionInDashboard(dashboardID, sessionID)
		if err == nil {
			err = a.launchWorkspaceCommand(sessionID, command, notConfigured)
		}
		return messages.ScopedActionFinishedMsg{DashboardID: dashboardID, Generation: generation, Err: err}
	}
}

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
		return session.LaunchWorkspaceCommand(command, sess.WorkingDir)
	}
	return fmt.Errorf("session not found: %s", sessionID)
}
