package ui

import (
	"context"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
)

type deletionPort interface {
	CreateHumanDashboard(string, string) error
	UpdateHumanDashboard(string, string, string) error
	PrepareHumanDashboardDeletion(string) (contracts.DeletionPreparation, error)
	PrepareHumanCardDeletion(string, string) (contracts.DeletionPreparation, error)
	ConfirmHumanDashboardDeletion(context.Context, string, string, bool) (contracts.DeletionProgress, error)
	RetryHumanDashboardDeletion(context.Context, string) (contracts.DeletionProgress, error)
	ReadHumanDashboardDeletion(string) (contracts.DeletionProgress, error)
	ConfirmHumanCardDeletion(context.Context, string, string, string, bool) (contracts.DeletionProgress, error)
	RetryHumanCardDeletion(context.Context, string, string) (contracts.DeletionProgress, error)
}
