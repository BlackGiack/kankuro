package app

import (
	"context"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
)

func (a sessionActions) launch(dashboardID string, generation uint64, cardID string, request uint64) tea.Cmd {
	return func() tea.Msg {
		result := a.port.LaunchSession(context.Background(), contracts.LaunchSessionRequest{DashboardID: dashboardID, CardID: cardID})
		return messages.SessionLaunchCompletedMsg{
			Request:     request,
			CardID:      cardID,
			Result:      result,
			DashboardID: dashboardID,
			Generation:  generation,
		}
	}
}

func (a sessionActions) activate(
	dashboardID string,
	generation, request uint64,
	sessionID, cardID string,
	relaunch bool,
	choice contracts.ConversationChoice,
) tea.Cmd {
	return func() tea.Msg {
		var result contracts.LifecycleResult
		if relaunch {
			result = a.port.RelaunchDashboardSession(
				context.Background(),
				contracts.RelaunchSessionRequest{DashboardID: dashboardID, SessionID: sessionID, CardID: cardID, Choice: choice},
			)
		} else {
			result = a.port.SwitchSession(
				context.Background(),
				contracts.SwitchSessionRequest{DashboardID: dashboardID, SessionID: sessionID, CardID: cardID, Choice: choice},
			)
		}
		return messages.SessionActivationCompletedMsg{
			DashboardID: dashboardID,
			Generation:  generation,
			Request:     request,
			SessionID:   sessionID,
			CardID:      cardID,
			Result:      result,
		}
	}
}

func (a sessionActions) launchWorkspace(
	dashboardID string,
	generation, request uint64,
	sessionID string,
	kind contracts.WorkspaceCommandKind,
) tea.Cmd {
	return func() tea.Msg {
		result := a.port.OpenSessionWorkspace(
			context.Background(),
			contracts.OpenSessionWorkspaceRequest{DashboardID: dashboardID, SessionID: sessionID, Kind: kind},
		)
		return messages.WorkspaceOpenedMsg{
			DashboardID: dashboardID,
			Generation:  generation,
			Request:     request,
			SessionID:   sessionID,
			Kind:        kind,
			Result:      result,
		}
	}
}

// sessionActions captures request values before executing blocking work.
type sessionActions struct{ port sessionActionPort }
