package api

import (
	"context"
	"errors"
	"sync"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/session"
)

// ReadSessionReports returns retained reports authored by a dashboard session.
func (s *Service) ReadSessionReports(ctx context.Context, req contracts.SessionReportsRequest) (contracts.SessionReportsResult, error) {
	if err := ctx.Err(); err != nil {
		return contracts.SessionReportsResult{}, err
	}
	if err := s.CheckSessionInDashboard(req.DashboardID, req.SessionID); err != nil {
		return contracts.SessionReportsResult{}, err
	}
	store := s.sessions.Messages()
	if store == nil {
		return contracts.SessionReportsResult{}, errMessagingUnavailable
	}
	page, err := store.RecentAuthoredReports(ctx, req.DashboardID, req.SessionID, req.Limit)
	if err != nil {
		return contracts.SessionReportsResult{}, err
	}
	if err = s.CheckSessionInDashboard(req.DashboardID, req.SessionID); err != nil {
		return contracts.SessionReportsResult{}, err
	}
	limit := min(req.Limit, session.MaxAuthoredReports)
	out := contracts.SessionReportsResult{
		Limit:   limit,
		Maximum: session.MaxAuthoredReports,
		Omitted: page.Omitted,
		Reports: make([]contracts.SessionReport, 0, len(page.Items)),
	}
	for _, item := range page.Items {
		out.Reports = append(
			out.Reports,
			contracts.SessionReport{
				ID:                 item.ID,
				MessageID:          item.MessageID,
				AuthorSessionID:    item.SenderSessionID,
				AuthorActivationID: item.SenderActivationID,
				Kind:               item.Kind,
				Text:               item.Text,
				CreatedAt:          item.CreatedAt,
				Sequence:           item.Sequence,
			},
		)
	}
	return out, nil
}

var errMessagingUnavailable = errors.New("messaging history is unavailable")

type reportSubscription struct {
	changes chan struct{}
	cancel  context.CancelFunc
	once    sync.Once
	mu      sync.Mutex
	err     error
}

func (s *reportSubscription) Changes() <-chan struct{} { return s.changes }
func (s *reportSubscription) Release()                 { s.once.Do(s.cancel) }
func (s *reportSubscription) Err() error               { s.mu.Lock(); defer s.mu.Unlock(); return s.err }
func (s *reportSubscription) finish(err error) {
	s.mu.Lock()
	s.err = err
	s.mu.Unlock()
	close(s.changes)
}

// SubscribeSessionReports captures the store signal before returning, so callers
// can safely perform their initial read after setup completes.
func (s *Service) SubscribeSessionReports(ctx context.Context, req contracts.ScopedSessionRequest) (contracts.ReportSubscription, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := s.CheckSessionInDashboard(req.DashboardID, req.SessionID); err != nil {
		return nil, err
	}
	store := s.sessions.Messages()
	if store == nil {
		return nil, errMessagingUnavailable
	}
	signal := store.ReportsChanged()
	scope := store.ReportScopeChanged()
	shutdown := s.sessions.ShutdownStarted()
	select {
	case <-shutdown:
		return nil, errors.New("session manager shut down")
	default:
	}
	if err := s.CheckSessionInDashboard(req.DashboardID, req.SessionID); err != nil {
		return nil, err
	}
	subctx, cancel := context.WithCancel(ctx)
	sub := &reportSubscription{changes: make(chan struct{}, 1), cancel: cancel}
	go watchReportChanges(subctx, cancel, shutdown, signal, scope, store.ReportsChanged, store.ReportScopeChanged, func() error {
		return s.CheckSessionInDashboard(req.DashboardID, req.SessionID)
	}, sub)
	return sub, nil
}

// watchReportChanges rearms before notifying consumers so a change during
// rearming remains observable even when notices coalesce.
func watchReportChanges(
	ctx context.Context,
	cancel context.CancelFunc,
	shutdown <-chan struct{},
	signal, scope <-chan struct{},
	changed, scopeChanged func() <-chan struct{},
	check func() error,
	sub *reportSubscription,
) {
	defer cancel()
	for {
		select {
		case <-shutdown:
			sub.finish(errors.New("session manager shut down"))
			return
		case <-ctx.Done():
			sub.finish(ctx.Err())
			return
		case <-scope:
			scope = scopeChanged()
			if err := check(); err != nil {
				sub.finish(err)
				return
			}
		case <-signal:
			signal = changed()
			if err := check(); err != nil {
				sub.finish(err)
				return
			}
			select {
			case sub.changes <- struct{}{}:
			default:
			}
		}
	}
}
