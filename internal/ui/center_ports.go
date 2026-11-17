package ui

import (
	"context"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
)

type centerPort interface {
	ReadSessionReports(context.Context, contracts.SessionReportsRequest) (contracts.SessionReportsResult, error)
	SubscribeSessionReports(context.Context, contracts.ScopedSessionRequest) (contracts.ReportSubscription, error)
	ReadSessionWorkspace(context.Context, contracts.ScopedSessionRequest) (contracts.SessionWorkspaceResult, error)
}
