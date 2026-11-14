package ui

import (
	"fmt"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

// deletionResultLines turns an authoritative execution result into guidance;
// an inventory read is never used to infer the operation's outcome.
func deletionResultLines(p contracts.DeletionProgress, operationErr error) []string {
	var outcome string
	switch p.Outcome {
	case contracts.DeletionNoIntent:
		outcome = "No deletion intent was saved. Review and confirm the deletion again."
	case contracts.DeletionPendingOutcome:
		outcome = "Deletion is pending. Retry resumes the saved deletion."
	case contracts.DeletionCompleted:
		outcome = "Deletion completed."
	case contracts.DeletionUncertain:
		outcome = "Deletion outcome is uncertain. Inspect the deletion status before taking further action."
	default:
		outcome = "Deletion outcome is uncertain. Inspect the deletion status before taking further action."
	}
	lines := []string{outcome}
	if p.Diagnostic != "" {
		lines = append(lines, "Detail: "+p.Diagnostic)
	} else if operationErr != nil {
		lines = append(lines, "Detail: "+operationErr.Error())
	}
	if p.DashboardID != "" {
		lines = append(lines, "Dashboard: "+p.DashboardID)
	}
	if p.CardID != "" {
		lines = append(lines, "Card: "+p.CardID)
	}
	if (p.CardID == "" && p.Outcome == contracts.DeletionPendingOutcome) || p.AdmissionsDrained || p.HistoryRemoved {
		lines = append(lines, fmt.Sprintf("Admissions drained: %t; history removed: %t", p.AdmissionsDrained, p.HistoryRemoved))
	}
	for _, s := range p.Sessions {
		lines = append(
			lines,
			fmt.Sprintf(
				"Session %s: stopped=%t removed=%t record deleted=%t temporary workspace removed=%t %s",
				s.ID,
				s.Stopped,
				s.Removed,
				s.RecordDeleted,
				s.TemporaryRemoved,
				s.Diagnostic,
			),
		)
	}
	for _, c := range p.Cards {
		lines = append(lines, fmt.Sprintf("Card %s: deleted=%t %s", c.ID, c.Deleted, c.Diagnostic))
	}
	for _, pending := range p.Pending {
		lines = append(lines, fmt.Sprintf("Pending session %s: cleaned=%t %s", pending.SessionID, pending.Cleaned, pending.Diagnostic))
	}
	return lines
}

func (c *Coordinator) showCardDeletionResult(p contracts.DeletionProgress, request uint64, dashboardID string, err error) bool {
	if request != 0 && c.cardDeleteTarget != "" && p.CardID != "" && p.CardID != c.cardDeleteTarget {
		return false
	}
	if p.CardID == "" {
		p.CardID = c.cardDeleteTarget
	}
	if p.DashboardID == "" {
		p.DashboardID = dashboardID
	}
	if p.DashboardID != dashboardID {
		return false
	}
	if dashboard, ok := c.views[ViewDashboard].(*views.Dashboard); ok {
		dashboard.ShowCardDeletionResult(deletionResultLines(p, err))
	}
	return true
}
