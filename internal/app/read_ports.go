package app

import (
	"github.com/BlackGiack/kankuro/internal/api/contracts"
)

// dashboardInventoryReader is the frontend-owned read capability for focused
// dashboard inventory and display data.
type dashboardInventoryReader interface {
	ReadDashboardInventory() ([]contracts.DashboardSummary, error)
	ReadDashboardOverview(dashboardID string) (contracts.DashboardInventory, error)
	ReadDisplayChoices() contracts.DisplayChoicesResult
	ReadRegistryChangeToken() (contracts.RegistryChangeToken, error)
}
