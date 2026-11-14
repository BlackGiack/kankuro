package api

import (
	"context"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/config"
)

const deletionNoIntent = "no_intent"

func (s *Service) CreateHumanDashboard(name, description string) error {
	_, err := s.CreateDashboard(name, description)
	return err
}
func (s *Service) UpdateHumanDashboard(id, name, description string) error {
	_, err := s.UpdateDashboardDetails(id, name, description)
	return err
}

func (s *Service) PrepareHumanDashboardDeletion(id string) (contracts.DeletionPreparation, error) {
	p, err := s.PreflightDashboardDeletion(id)
	if err != nil {
		return contracts.DeletionPreparation{}, err
	}
	v := projectDashboardDeletion(p.Manifest)
	return contracts.DeletionPreparation{
		DashboardID:                 id,
		Name:                        p.Manifest.Name,
		Token:                       p.Token,
		RequiresForeignConfirmation: p.RequiresForeignConfirmation,
		Sessions:                    v.Sessions,
		Cards:                       v.Cards,
		Pending:                     v.Pending,
	}, nil
}
func (s *Service) PrepareHumanCardDeletion(dashboardID, cardID string) (contracts.DeletionPreparation, error) {
	p, err := s.PreflightCardDeletion(dashboardID, cardID)
	if err != nil {
		return contracts.DeletionPreparation{}, err
	}
	v := contracts.DeletionPreparation{
		DashboardID:                 dashboardID,
		CardID:                      cardID,
		Name:                        p.Card.Name,
		Token:                       p.Token,
		RequiresForeignConfirmation: p.RequiresForeignConfirmation,
	}
	for _, ref := range p.Impact.Sessions {
		v.References = append(
			v.References,
			contracts.DeletionReference{DashboardID: ref.DashboardID, SessionID: ref.SessionID, Current: ref.Current, Saved: ref.Binding},
		)
	}
	for _, ref := range p.Impact.Pending {
		v.References = append(
			v.References,
			contracts.DeletionReference{DashboardID: ref.DashboardID, SessionID: ref.SessionID, Pending: true},
		)
	}
	return v, nil
}
func (s *Service) ConfirmHumanDashboardDeletion(ctx context.Context, id, token string, foreign bool) (contracts.DeletionProgress, error) {
	result, err := s.ConfirmDashboardDeletion(ctx, id, token, foreign)
	return dashboardDeletionProgress(id, result, err), err
}
func (s *Service) RetryHumanDashboardDeletion(ctx context.Context, id string) (contracts.DeletionProgress, error) {
	result, err := s.RetryDashboardDeletion(ctx, id)
	return dashboardDeletionProgress(id, result, err), err
}
func (s *Service) ReadHumanDashboardDeletion(id string) (contracts.DeletionProgress, error) {
	manifest, err := s.GetDashboardDeletion(id)
	if err != nil {
		return contracts.DeletionProgress{DashboardID: id, Outcome: contracts.DeletionUncertain}, err
	}
	progress := projectDashboardDeletion(manifest)
	progress.Outcome = contracts.DeletionPendingOutcome
	return progress, nil
}

func (s *Service) ConfirmHumanCardDeletion(
	ctx context.Context,
	dashboardID, cardID, token string,
	foreign bool,
) (contracts.DeletionProgress, error) {
	result, err := s.ConfirmCardDeletion(ctx, dashboardID, cardID, token, foreign)
	return cardDeletionProgress(dashboardID, cardID, result, err), err
}
func (s *Service) RetryHumanCardDeletion(ctx context.Context, dashboardID, cardID string) (contracts.DeletionProgress, error) {
	result, err := s.RetryCardDeletion(ctx, dashboardID, cardID)
	return cardDeletionProgress(dashboardID, cardID, result, err), err
}
func deletionOutcome(state string, err error) contracts.DeletionOutcome {
	switch state {
	case deletionNoIntent:
		return contracts.DeletionNoIntent
	case config.CardDeleted:
		return contracts.DeletionCompleted
	case config.CardDeleting:
		return contracts.DeletionPendingOutcome
	}
	if err != nil {
		return contracts.DeletionUncertain
	}
	return contracts.DeletionNoIntent
}
func cardDeletionProgress(dashboardID, cardID string, result CardDeletionResult, err error) contracts.DeletionProgress {
	p := contracts.DeletionProgress{DashboardID: dashboardID, CardID: cardID, Outcome: deletionOutcome(result.State, err)}
	if err != nil {
		p.Diagnostic = err.Error()
	}
	return p
}
func dashboardDeletionProgress(id string, result DashboardDeletionResult, err error) contracts.DeletionProgress {
	p := projectDashboardDeletion(result.Manifest)
	p.DashboardID = id
	p.Outcome = deletionOutcome(result.State, err)
	if err != nil {
		p.Diagnostic = err.Error()
	}
	return p
}
func projectDashboardDeletion(m config.DashboardDeletion) contracts.DeletionProgress {
	p := contracts.DeletionProgress{DashboardID: m.DashboardID, AdmissionsDrained: m.AdmissionsDrained, HistoryRemoved: m.HistoryRemoved}
	for _, s := range m.Sessions {
		p.Sessions = append(
			p.Sessions,
			contracts.DeletionSession{
				ID:               s.ID,
				Name:             s.Name,
				Status:           s.Status,
				Diagnostic:       s.Error,
				Stopped:          s.Stopped,
				Removed:          s.Removed,
				RecordDeleted:    s.RecordDeleted,
				TemporaryRemoved: s.TemporaryRemoved,
			},
		)
	}
	for _, c := range m.Cards {
		card := contracts.DeletionCard{ID: c.ID, Name: c.Name, Shared: c.Shared, Deleted: c.Deleted, Diagnostic: c.Error}
		for _, r := range c.References {
			card.References = append(
				card.References,
				contracts.DeletionReference{
					DashboardID: r.DashboardID,
					SessionID:   r.SessionID,
					Current:     r.Current,
					Saved:       r.Binding,
					Pending:     r.Pending,
				},
			)
		}
		p.Cards = append(p.Cards, card)
	}
	for _, q := range m.Pending {
		p.Pending = append(
			p.Pending,
			contracts.DeletionPending{
				CardID:      q.CardID,
				DashboardID: q.DashboardID,
				SessionID:   q.SessionID,
				Diagnostic:  q.Error,
				Cleaned:     q.Cleaned,
			},
		)
	}
	return p
}
