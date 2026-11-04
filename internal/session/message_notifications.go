package session

import (
	"context"
	"errors"
	"sort"

	"github.com/BlackGiack/kankuro/internal/config"
)

// Notifications returns a stable snapshot in report order. Legacy reports without
// notification metadata are deliberately excluded.
func (s *MessageStore) Notifications() []PendingNotification {
	s.mu.Lock()
	defer s.mu.Unlock()
	var result []PendingNotification
	for _, message := range s.data.Messages {
		for _, reply := range message.Replies {
			if reply.Notification != nil &&
				(reply.Notification.State == NotificationPending || reply.Notification.State == NotificationAccepted) {
				message.Replies = nil
				result = append(result, PendingNotification{Message: message, Reply: cloneReply(reply)})
			}
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Reply.Sequence < result[j].Reply.Sequence })
	return result
}

// UpdateNotification updates only the named retained report, conditional on its
// current attempt. The mutation and attempt advance are saved together.
func (s *MessageStore) UpdateNotification(
	messageID, replyID string,
	attempt uint64,
	expected NotificationState,
	update func(*ReplyNotification),
) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.data.Messages {
		message := &s.data.Messages[i]
		if message.ID != messageID {
			continue
		}
		for j := range message.Replies {
			reply := &message.Replies[j]
			if reply.ID != replyID || reply.Notification == nil {
				continue
			}
			if reply.Notification.Attempt != attempt || reply.Notification.State != expected {
				return messagingError("attempt_conflict", "notification attempt changed", nil)
			}
			old := *reply.Notification
			update(reply.Notification)
			if err := s.saveLocked(); err != nil {
				*reply.Notification = old
				return messagingError("storage_unavailable", "notification state could not be saved", err)
			}
			s.signalLocked()
			return nil
		}
	}
	return messagingError("history_unavailable", "report history is unavailable", nil)
}

func (s *MessageStore) notification(messageID, replyID string) (PendingNotification, bool) {
	for _, notification := range s.Notifications() {
		if notification.Message.ID == messageID && notification.Reply.ID == replyID {
			return notification, true
		}
	}
	return PendingNotification{}, false
}

// ReserveReportNotification reserves only the exact destination in a retained report.
func (m *Manager) ReserveReportNotification(messageID, replyID string) (*MessageAdmission, error) {
	if m.messages == nil {
		return nil, messagingError("history_unavailable", "messaging history is unavailable", nil)
	}
	notification, ok := m.messages.notification(messageID, replyID)
	if !ok {
		return nil, messagingError("history_unavailable", "report history is unavailable", nil)
	}
	original := notification.Message
	m.mu.Lock()
	defer m.mu.Unlock()
	target := m.sessions[original.SenderSessionID]
	dashboardAvailable, dashboardErr := m.reportDashboardAvailableLocked(original.DashboardID)
	if dashboardErr != nil {
		return nil, messagingError("storage_unavailable", "dashboard state is unavailable", dashboardErr)
	}
	if target == nil || target.DashboardID != original.DashboardID || m.dashboardDeletionWorkers[original.DashboardID] ||
		!dashboardAvailable ||
		target.ActivationID != original.SenderActivationID ||
		target.Status == StatusTerminated ||
		target.stopping ||
		target.deleting ||
		target.activating ||
		!target.KankuroMCPEnabled {
		return nil, messagingError("stale_activation", "original sender activation is unavailable", nil)
	}
	if target.starting {
		return nil, messagingError("starting", "original sender is starting", nil)
	}
	if !m.deliveryAvailableLocked(target) {
		return nil, messagingError("stale_activation", "original sender activation is unavailable", nil)
	}
	ctx, cancel := context.WithCancel(context.Background())
	if target.activeSubmissions == 0 {
		target.submissionsDone = make(chan struct{})
	}
	target.activeSubmissions++
	admission := &MessageAdmission{manager: m, recipient: target, senderID: notification.Reply.SenderSessionID,
		senderActivationID: notification.Reply.SenderActivationID, recipientID: target.ID,
		recipientActivationID: target.ActivationID, ctx: ctx, cancel: cancel,
		reportMessageID: messageID, reportReplyID: replyID}
	if m.messageAdmissions == nil {
		m.messageAdmissions = make(map[*MessageAdmission]struct{})
	}
	m.messageAdmissions[admission] = struct{}{}
	return admission, nil
}

func (m *Manager) reportDashboardAvailableLocked(dashboardID string) (bool, error) {
	repo, ok := m.repo.(interface {
		ListDashboardDeletions() ([]config.DashboardDeletion, error)
		ListDashboards() ([]config.Dashboard, error)
	})
	if !ok {
		return false, errors.New("dashboard inventory unavailable")
	}
	deletions, err := repo.ListDashboardDeletions()
	if err != nil {
		return false, err
	}
	for _, deletion := range deletions {
		if deletion.DashboardID == dashboardID {
			return false, nil
		}
	}
	dashboards, err := repo.ListDashboards()
	if err != nil {
		return false, err
	}
	for _, dashboard := range dashboards {
		if dashboard.ID == dashboardID {
			return true, nil
		}
	}
	return false, nil
}

func (m *Manager) reportAdmissionValidLocked(admission *MessageAdmission) bool {
	if admission.recipient == nil {
		return false
	}
	report, found := m.messages.notification(admission.reportMessageID, admission.reportReplyID)
	if !found || report.Message.SenderSessionID != admission.recipientID ||
		report.Message.SenderActivationID != admission.recipientActivationID ||
		admission.recipient.ActivationID != admission.recipientActivationID ||
		admission.recipient.DashboardID != report.Message.DashboardID || m.dashboardDeletionWorkers[report.Message.DashboardID] {
		return false
	}
	available, err := m.reportDashboardAvailableLocked(report.Message.DashboardID)
	return err == nil && available
}
