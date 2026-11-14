package ui

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

// cardEditorPort is the frontend-owned card editing capability.
type cardEditorPort interface {
	LoadCardEditor(string, string) (contracts.CardEditorData, error)
	NewCardEditor(string) (contracts.CardEditorData, error)
	ReadCardEditorChoices() contracts.CardEditorChoices
	CreateEditorCard(contracts.CreateCardInput) (contracts.CardSaveResult, error)
	UpdateEditorCard(contracts.UpdateCardInput) (contracts.CardSaveResult, error)
}
