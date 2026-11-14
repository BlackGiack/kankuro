package ui

import (
	"context"
	"errors"
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

const dashboardDeletionTimeout = 5 * time.Minute

func (c *Coordinator) dashboardListCallbacks() views.DashboardListCallbacks {
	service := c.deletion
	workers := c.deletionWorkers
	return views.DashboardListCallbacks{
		Open: func(dashboard contracts.DashboardSummary) tea.Cmd {
			generation := c.viewGeneration
			return func() tea.Msg {
				return messages.DashboardOpenRequestedMsg{
					DashboardID: dashboard.ID, DashboardName: dashboard.Name,
					DashboardState: dashboard.State, Generation: generation,
				}
			}
		},
		Save: func(id, name, description string) tea.Cmd {
			generation := c.viewGeneration
			return func() tea.Msg {
				var err error
				if id == "" {
					err = service.CreateHumanDashboard(name, description)
				} else {
					err = service.UpdateHumanDashboard(id, name, description)
				}
				return messages.DashboardMutationFinishedMsg{Err: err, Generation: generation}
			}
		},
		PreflightDelete: func(id string) tea.Cmd {
			c.dashboardDeleteRead++
			request, visit := c.dashboardDeleteRead, c.deletionVisit
			return func() tea.Msg {
				preflight, err := service.PrepareHumanDashboardDeletion(id)
				return messages.DashboardDeletionPreparedMsg{
					DashboardID: id, Request: request, Visit: visit,
					Token: preflight.Token,
					Summary: dashboardDeletionSummary(
						preflight,
					),
					RequiresForeignConfirmation: preflight.RequiresForeignConfirmation,
					Err:                         err,
				}
			}
		},
		ConfirmDelete: func(id, token string, foreign bool) tea.Cmd {
			if c.dashboardDeleting {
				return nil
			}
			c.dashboardDeleting = true
			c.dashboardDeleteTarget = id
			c.dashboardDeleteAction++
			request, visit := c.dashboardDeleteAction, c.deletionVisit
			generation := c.viewGeneration
			return func() tea.Msg {
				progress, err := workers.run(func(ctx context.Context) (contracts.DeletionProgress, error) {
					return service.ConfirmHumanDashboardDeletion(ctx, id, token, foreign)
				})
				return messages.DashboardMutationFinishedMsg{
					Err:        err,
					Generation: generation,
					Progress:   progress,
					Request:    request,
					Visit:      visit,
				}
			}
		},
		RetryDelete: func(id string) tea.Cmd {
			if c.dashboardDeleting {
				return nil
			}
			c.dashboardDeleting = true
			c.dashboardDeleteTarget = id
			c.dashboardDeleteAction++
			request, visit := c.dashboardDeleteAction, c.deletionVisit
			generation := c.viewGeneration
			return func() tea.Msg {
				progress, err := workers.run(func(ctx context.Context) (contracts.DeletionProgress, error) {
					return service.RetryHumanDashboardDeletion(ctx, id)
				})
				return messages.DashboardMutationFinishedMsg{
					Err:        err,
					Generation: generation,
					Progress:   progress,
					Request:    request,
					Visit:      visit,
				}
			}
		},
		InspectDelete: func(id string) tea.Cmd {
			c.dashboardDeleteRead++
			request, visit := c.dashboardDeleteRead, c.deletionVisit
			return func() tea.Msg {
				manifest, err := service.ReadHumanDashboardDeletion(id)
				return messages.DashboardDeletionStatusMsg{
					DashboardID: id,
					Request:     request,
					Visit:       visit,
					Summary:     dashboardDeletionStatus(manifest),
					Err:         err,
				}
			}
		},
	}
}

func dashboardDeletionStatus(manifest contracts.DeletionProgress) []string {
	lines := []string{fmt.Sprintf("Admissions drained: %t; history removed: %t", manifest.AdmissionsDrained, manifest.HistoryRemoved)}
	for _, item := range manifest.Sessions {
		lines = append(
			lines,
			fmt.Sprintf(
				"Session %s: stopped=%t removed=%t record deleted=%t temporary workspace removed=%t %s",
				item.ID,
				item.Stopped,
				item.Removed,
				item.RecordDeleted,
				item.TemporaryRemoved,
				item.Diagnostic,
			),
		)
	}
	for _, item := range manifest.Cards {
		lines = append(lines, fmt.Sprintf("Card %s: deleted=%t %s", item.ID, item.Deleted, item.Diagnostic))
	}
	for _, item := range manifest.Pending {
		lines = append(lines, fmt.Sprintf("Pending session %s: cleaned=%t %s", item.SessionID, item.Cleaned, item.Diagnostic))
	}
	return lines
}

func dashboardDeletionSummary(preflight contracts.DeletionPreparation) []string {
	manifest := preflight
	lines := []string{"Owned sessions:"}
	if len(manifest.Sessions) == 0 {
		lines = append(lines, "  none")
	}
	for _, item := range manifest.Sessions {
		lines = append(lines, fmt.Sprintf("  %s (%s, %s)", item.Name, item.ID, item.Status))
	}
	lines = append(lines, "Owned cards (shared cards are deleted globally):")
	if len(manifest.Cards) == 0 {
		lines = append(lines, "  none")
	}
	for _, card := range manifest.Cards {
		lines = append(lines, fmt.Sprintf("  %s (%s)", card.Name, card.ID))
		for _, ref := range card.References {
			if ref.DashboardID != manifest.DashboardID {
				lines = append(lines, fmt.Sprintf("    foreign dashboard %s: session %s", ref.DashboardID, ref.SessionID))
			}
		}
	}
	for _, pending := range manifest.Pending {
		lines = append(lines, fmt.Sprintf("Pending activation: dashboard %s, session %s", pending.DashboardID, pending.SessionID))
	}
	return append(lines, "Foreign sessions and global registries survive.")
}

func (c *Coordinator) refreshDashboardList() tea.Cmd {
	c.dashboardListRequest++
	return dashboardListCommand(c.reads, c.dashboardListRequest)
}

func dashboardListCommand(service dashboardInventoryReader, requestID uint64) tea.Cmd {
	return func() tea.Msg {
		dashboards, err := service.ReadDashboardInventory()
		return messages.DashboardListLoadedMsg{Dashboards: dashboards, Err: err, RequestID: requestID}
	}
}

func dashboardCardCommand(
	service dashboardInventoryReader,
	dashboardID, cardID string,
	generation, request uint64,
	action messages.DashboardCardAction,
) tea.Cmd {
	return func() tea.Msg {
		inventory, err := service.ReadDashboardOverview(dashboardID)
		if err != nil {
			return messages.DashboardCardLoadedMsg{
				Request:     request,
				DashboardID: dashboardID,
				Generation:  generation,
				CardID:      cardID,
				Action:      action,
				Err:         err,
			}
		}
		for _, card := range inventory.Cards {
			if card.ID == cardID {
				return messages.DashboardCardLoadedMsg{
					Request:     request,
					DashboardID: dashboardID,
					Generation:  generation,
					CardID:      cardID,
					Action:      action,
					Card:        card,
				}
			}
		}
		return messages.DashboardCardLoadedMsg{
			Request:     request,
			DashboardID: dashboardID,
			Generation:  generation,
			CardID:      cardID,
			Action:      action,
			Err:         errors.New("card unavailable"),
		}
	}
}

func cardDeletionPreflightCommand(service deletionPort, dashboardID, cardID string, generation, request, visit uint64) tea.Cmd {
	return func() tea.Msg {
		preflight, err := service.PrepareHumanCardDeletion(dashboardID, cardID)
		var summary []string
		for _, item := range preflight.References {
			if item.Pending {
				summary = append(summary, fmt.Sprintf("Pending: dashboard %s, session %s", item.DashboardID, item.SessionID))
				continue
			}
			summary = append(
				summary,
				fmt.Sprintf(
					"Dashboard %s, session %s (current: %t, saved: %t)",
					item.DashboardID,
					item.SessionID,
					item.Current,
					item.Saved,
				),
			)
		}
		summary = append(summary, "Other dashboards' sessions and workspaces survive.")
		return messages.CardDeletionPreparedMsg{
			DashboardID:                 dashboardID,
			Generation:                  generation,
			CardID:                      cardID,
			Request:                     request,
			Visit:                       visit,
			Token:                       preflight.Token,
			Summary:                     summary,
			RequiresForeignConfirmation: preflight.RequiresForeignConfirmation,
			Err:                         err,
		}
	}
}

func confirmCardDeletionCommand(
	service deletionPort,
	workers *deletionWorkers,
	dashboardID, cardID, token string,
	generation, request, visit uint64,
	foreign bool,
) tea.Cmd {
	return func() tea.Msg {
		progress, err := workers.run(func(ctx context.Context) (contracts.DeletionProgress, error) {
			return service.ConfirmHumanCardDeletion(ctx, dashboardID, cardID, token, foreign)
		})
		return messages.ScopedActionFinishedMsg{
			DashboardID: dashboardID,
			Generation:  generation,
			Err:         err,
			Progress:    progress,
			Request:     request,
			Visit:       visit,
		}
	}
}

func retryCardDeletionCommand(
	service deletionPort,
	workers *deletionWorkers,
	dashboardID, cardID string,
	generation, request, visit uint64,
) tea.Cmd {
	return func() tea.Msg {
		progress, err := workers.run(func(ctx context.Context) (contracts.DeletionProgress, error) {
			return service.RetryHumanCardDeletion(ctx, dashboardID, cardID)
		})
		return messages.ScopedActionFinishedMsg{
			DashboardID: dashboardID,
			Generation:  generation,
			Err:         err,
			Progress:    progress,
			Request:     request,
			Visit:       visit,
		}
	}
}

func (c *Coordinator) confirmCardDeletion(dashboardID, cardID, token string, generation uint64, foreign bool) tea.Cmd {
	if c.cardDeleting {
		return nil
	}
	c.cardDeleting = true
	c.cardDeleteTarget = cardID
	c.cardDeleteAction++
	return confirmCardDeletionCommand(
		c.deletion,
		c.deletionWorkers,
		dashboardID,
		cardID,
		token,
		generation,
		c.cardDeleteAction,
		c.deletionVisit,
		foreign,
	)
}
