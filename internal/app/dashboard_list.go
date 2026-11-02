package app

import (
	"context"
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/api"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

const dashboardDeletionTimeout = 5 * time.Minute

func (c *Coordinator) dashboardListCallbacks() views.DashboardListCallbacks {
	service := c.actions.service
	return views.DashboardListCallbacks{
		Open: func(dashboard config.Dashboard) tea.Cmd {
			generation := c.viewGeneration
			return func() tea.Msg {
				return messages.DashboardOpenRequestedMsg{Dashboard: dashboard, Generation: generation}
			}
		},
		Save: func(id, name, description string) tea.Cmd {
			generation := c.viewGeneration
			return func() tea.Msg {
				var err error
				if id == "" {
					_, err = service.CreateDashboard(name, description)
				} else {
					_, err = service.UpdateDashboardDetails(id, name, description)
				}
				return messages.DashboardMutationFinishedMsg{Err: err, Generation: generation}
			}
		},
		PreflightDelete: func(id string) tea.Cmd {
			return func() tea.Msg {
				preflight, err := service.PreflightDashboardDeletion(id)
				return messages.DashboardDeletionPreparedMsg{
					DashboardID: id,
					Token:       preflight.Token,
					Summary: dashboardDeletionSummary(
						preflight,
					),
					RequiresForeignConfirmation: preflight.RequiresForeignConfirmation,
					Err:                         err,
				}
			}
		},
		ConfirmDelete: func(id, token string, foreign bool) tea.Cmd {
			generation := c.viewGeneration
			return func() tea.Msg {
				ctx, cancel := context.WithTimeout(context.Background(), dashboardDeletionTimeout)
				defer cancel()
				_, err := service.ConfirmDashboardDeletion(ctx, id, token, foreign)
				return messages.DashboardMutationFinishedMsg{Err: err, Generation: generation}
			}
		},
		RetryDelete: func(id string) tea.Cmd {
			generation := c.viewGeneration
			return func() tea.Msg {
				ctx, cancel := context.WithTimeout(context.Background(), dashboardDeletionTimeout)
				defer cancel()
				_, err := service.RetryDashboardDeletion(ctx, id)
				return messages.DashboardMutationFinishedMsg{Err: err, Generation: generation}
			}
		},
		InspectDelete: func(id string) tea.Cmd {
			return func() tea.Msg {
				manifest, err := service.GetDashboardDeletion(id)
				return messages.DashboardDeletionStatusMsg{DashboardID: id, Summary: dashboardDeletionStatus(manifest), Err: err}
			}
		},
	}
}

func dashboardDeletionStatus(manifest config.DashboardDeletion) []string {
	lines := []string{fmt.Sprintf("Admissions drained: %t; history removed: %t", manifest.AdmissionsDrained, manifest.HistoryRemoved)}
	for _, item := range manifest.Sessions {
		lines = append(lines, fmt.Sprintf("Session %s: stopped=%t removed=%t %s", item.ID, item.Stopped, item.Removed, item.Error))
	}
	for _, item := range manifest.Cards {
		lines = append(lines, fmt.Sprintf("Card %s: deleted=%t %s", item.ID, item.Deleted, item.Error))
	}
	for _, item := range manifest.Pending {
		lines = append(lines, fmt.Sprintf("Pending session %s: cleaned=%t %s", item.SessionID, item.Cleaned, item.Error))
	}
	return lines
}

func dashboardDeletionSummary(preflight api.DashboardDeletionPreflight) []string {
	manifest := preflight.Manifest
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
	return dashboardListCommand(c.actions.service, c.dashboardListRequest)
}

func dashboardListCommand(service *api.Service, requestID uint64) tea.Cmd {
	return func() tea.Msg {
		dashboards, err := service.ListDashboards()
		return messages.DashboardListLoadedMsg{Dashboards: dashboards, Err: err, RequestID: requestID}
	}
}

func cardDeletionPreflightCommand(service *api.Service, dashboardID, cardID string, generation uint64) tea.Cmd {
	return func() tea.Msg {
		preflight, err := service.PreflightCardDeletion(dashboardID, cardID)
		var summary []string
		for _, item := range preflight.Impact.Sessions {
			summary = append(
				summary,
				fmt.Sprintf(
					"Dashboard %s, session %s (current: %t, saved: %t)",
					item.DashboardID,
					item.SessionID,
					item.Current,
					item.Binding,
				),
			)
		}
		for _, item := range preflight.Impact.Pending {
			summary = append(summary, fmt.Sprintf("Pending: dashboard %s, session %s", item.DashboardID, item.SessionID))
		}
		summary = append(summary, "Other dashboards' sessions and workspaces survive.")
		return messages.CardDeletionPreparedMsg{DashboardID: dashboardID, Generation: generation, CardID: cardID,
			Token: preflight.Token, Summary: summary, RequiresForeignConfirmation: preflight.RequiresForeignConfirmation, Err: err}
	}
}

func confirmCardDeletionCommand(service *api.Service, dashboardID, cardID, token string, generation uint64, foreign bool) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), dashboardDeletionTimeout)
		defer cancel()
		_, err := service.ConfirmCardDeletion(ctx, dashboardID, cardID, token, foreign)
		return messages.ScopedActionFinishedMsg{DashboardID: dashboardID, Generation: generation, Err: err}
	}
}

func retryCardDeletionCommand(service *api.Service, dashboardID, cardID string, generation uint64) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), dashboardDeletionTimeout)
		defer cancel()
		_, err := service.RetryCardDeletion(ctx, dashboardID, cardID)
		return messages.ScopedActionFinishedMsg{DashboardID: dashboardID, Generation: generation, Err: err}
	}
}
