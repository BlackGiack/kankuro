package app

import (
	"context"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
)

// sessionActionPort publishes detached values. Commands capture requests before
// leaving the coordinator goroutine; the consumer owns returned option slices.
type sessionActionPort interface {
	LaunchSession(context.Context, contracts.LaunchSessionRequest) contracts.LifecycleResult
	RelaunchDashboardSession(context.Context, contracts.RelaunchSessionRequest) contracts.LifecycleResult
	SwitchSession(context.Context, contracts.SwitchSessionRequest) contracts.LifecycleResult
	ReadConversationPreparationResult(context.Context, contracts.ConversationPreparationRequest) contracts.ConversationPreparation
	RenameDashboardSession(context.Context, contracts.RenameSessionRequest) contracts.RenameSessionResult
	StopDashboardSession(context.Context, contracts.ScopedSessionRequest) contracts.StopSessionResult
	RemoveDashboardSession(context.Context, contracts.ScopedSessionRequest) contracts.RemoveSessionResult
	OpenSessionWorkspace(context.Context, contracts.OpenSessionWorkspaceRequest) contracts.OpenSessionWorkspaceResult
}
