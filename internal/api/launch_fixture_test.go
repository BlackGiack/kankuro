//nolint:testpackage // Uses the package integration fixtures.
package api

import (
	"context"
	"errors"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
)

func launchSessionForTest(ctx context.Context, service *Service, dashboardID, cardID string) (string, error) {
	result := service.LaunchSession(ctx, contracts.LaunchSessionRequest{DashboardID: dashboardID, CardID: cardID})
	if result.Diagnostic != "" {
		return result.SessionID, errors.New(result.Diagnostic)
	}
	return result.SessionID, nil
}
