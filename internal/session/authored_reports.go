package session

import (
	"container/heap"
	"context"
	"errors"
	"sort"
)

// MaxAuthoredReports bounds the amount of retained history projected for a panel.
const (
	MaxAuthoredReports      = 256
	authoredReportsMaxBytes = 4 << 20
)

// AuthoredReports is a detached, newest-first projection of retained reports.
// Omitted counts matching retained reports excluded by the requested limit.
type AuthoredReports struct {
	Items   []MessageReply
	Omitted int
}

// RecentAuthoredReports selects reports by their author, not the parent sender.
// It holds at most limit candidates while scanning the bounded history.
func (s *MessageStore) RecentAuthoredReports(ctx context.Context, dashboardID, authorID string, limit int) (AuthoredReports, error) {
	if dashboardID == "" || authorID == "" || limit <= 0 {
		return AuthoredReports{}, errors.New("dashboard, author and positive report limit are required")
	}
	if limit > MaxAuthoredReports {
		limit = MaxAuthoredReports
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	candidates := make(reportHeap, 0, limit)
	count := 0
	for _, message := range s.data.Messages {
		if err := ctx.Err(); err != nil {
			return AuthoredReports{}, err
		}
		if message.DashboardID != dashboardID {
			continue
		}
		for _, reply := range message.Replies {
			if err := ctx.Err(); err != nil {
				return AuthoredReports{}, err
			}
			if reply.SenderSessionID != authorID {
				continue
			}
			count++
			// Notification is not part of the authored report projection.
			reply.Notification = nil
			if len(candidates) < limit {
				heap.Push(&candidates, reply)
			} else if reportNewer(reply, candidates[0]) {
				candidates[0] = reply
				heap.Fix(&candidates, 0)
			}
		}
	}
	sort.Slice(candidates, func(i, j int) bool { return reportNewer(candidates[i], candidates[j]) })
	return boundAuthoredReports(candidates, count)
}

// Include attribution bytes in the response budget, keeping complete records.
func boundAuthoredReports(candidates reportHeap, count int) (AuthoredReports, error) {
	total, kept := 0, 0
	for _, item := range candidates {
		size := len(item.ID) + len(item.MessageID) + len(item.SenderSessionID)
		size += len(item.SenderActivationID) + len(item.Kind) + len(item.Text)
		if len(item.Text) > messageReportLimitBytes {
			return AuthoredReports{}, errors.New("retained report exceeds size limit")
		}
		if size > authoredReportsMaxBytes-total {
			break
		}
		total += size
		kept++
	}
	if kept == 0 && count > 0 {
		return AuthoredReports{}, errors.New("retained report exceeds response budget")
	}
	return AuthoredReports{Items: []MessageReply(candidates[:kept]), Omitted: count - kept}, nil
}

func reportNewer(a, b MessageReply) bool {
	if !a.CreatedAt.Equal(b.CreatedAt) {
		return a.CreatedAt.After(b.CreatedAt)
	}
	if a.Sequence != b.Sequence {
		return a.Sequence > b.Sequence
	}
	return a.ID < b.ID
}

type reportHeap []MessageReply

func (h *reportHeap) Len() int           { return len(*h) }
func (h *reportHeap) Less(i, j int) bool { return reportNewer((*h)[j], (*h)[i]) }
func (h *reportHeap) Swap(i, j int)      { (*h)[i], (*h)[j] = (*h)[j], (*h)[i] }
func (h *reportHeap) Push(x any) {
	item, ok := x.(MessageReply)
	if !ok {
		panic("invalid report heap item")
	}
	*h = append(*h, item)
}
func (h *reportHeap) Pop() any {
	old := *h
	x := old[len(old)-1]
	*h = old[:len(old)-1]
	return x
}
