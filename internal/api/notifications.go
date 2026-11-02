package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/BlackGiack/kankuro/internal/driver"
	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/google/uuid"
)

const (
	notificationRetryInitial    = time.Second
	notificationStaleActivation = "stale_activation"
	notificationRetryMaximum    = 30 * time.Second
	notificationRetryMultiplier = 2
)

type notificationAuthority struct {
	service            *Service
	messageID, replyID string
}

func (a notificationAuthority) validate(op Operation) error {
	store := a.service.sessions.Messages()
	if store == nil {
		return errors.New("report history unavailable")
	}
	for _, item := range store.Notifications() {
		if item.Message.ID != a.messageID || item.Reply.ID != a.replyID {
			continue
		}
		if item.Message.DashboardID != op.DashboardID || item.Message.SenderSessionID != op.TargetSession ||
			a.service.scope.DashboardForSession(op.TargetSession) != op.DashboardID {
			return errors.New("notification dashboard changed")
		}
		target, ok := a.service.sessions.GetSession(op.TargetSession)
		if !ok || target.ActivationID != item.Message.SenderActivationID || !target.MessageDeliveryAvailable {
			return errors.New("original activation unavailable")
		}
		return nil
	}
	return errors.New("report history unavailable")
}

func (a notificationAuthority) monitor(op Operation) error { return a.validate(op) }

type notificationSchedule struct {
	due   time.Time
	delay time.Duration
}

func (r *OperationRunner) startDispatcher() {
	if r.service == nil || r.service.sessions.Messages() == nil || r.loadErr != nil {
		return
	}
	r.mu.Lock()
	if r.closed || r.dispatchCancel != nil {
		r.mu.Unlock()
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	r.dispatchCancel = cancel
	r.dispatcher.Add(1)
	r.mu.Unlock()
	go r.dispatchNotifications(ctx)
}

func (r *OperationRunner) dispatchNotifications(ctx context.Context) {
	defer r.dispatcher.Done()
	store := r.service.sessions.Messages()
	schedule := make(map[string]notificationSchedule)
	for {
		changed := store.Changed()
		now := time.Now()
		items := store.Notifications()
		retained := make(map[string]bool, len(items))
		for _, item := range items {
			retained[item.Reply.ID] = true
		}
		for _, item := range dueNotifications(items, schedule, now) {
			entry := schedule[item.Reply.ID]
			r.dispatchNotification(item)
			entry.delay = nextNotificationDelay(entry.delay)
			entry.due = now.Add(entry.delay)
			schedule[item.Reply.ID] = entry
		}
		for id := range schedule {
			if !retained[id] {
				delete(schedule, id)
			}
		}
		timer := time.NewTimer(notificationRetryInitial)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-changed:
			timer.Stop()
		case <-timer.C:
		}
	}
}

func dueNotifications(
	items []session.PendingNotification,
	schedule map[string]notificationSchedule,
	now time.Time,
) []session.PendingNotification {
	due := make([]session.PendingNotification, 0, len(items))
	blockedRecipients := make(map[string]bool)
	for _, item := range items {
		recipient := item.Message.SenderSessionID + "\x00" + item.Message.SenderActivationID
		if blockedRecipients[recipient] {
			continue
		}
		blockedRecipients[recipient] = true
		if !now.Before(schedule[item.Reply.ID].due) {
			due = append(due, item)
		}
	}
	return due
}

func nextNotificationDelay(previous time.Duration) time.Duration {
	if previous == 0 {
		return notificationRetryInitial
	}
	return min(previous*notificationRetryMultiplier, notificationRetryMaximum)
}

//nolint:funlen // The admission transaction must keep identity, reservation, and journal acceptance together.
func (r *OperationRunner) dispatchNotification(item session.PendingNotification) {
	store := r.service.sessions.Messages()
	n := item.Reply.Notification
	if n.State == session.NotificationAccepted {
		r.reconcileNotification(item)
		return
	}
	requestID := n.RequestID
	if requestID == "" {
		requestID = uuid.NewSHA1(uuid.NameSpaceURL, []byte(fmt.Sprintf("kankuro-report-notification:%s:%d", item.Reply.ID, n.Attempt))).
			String()
		if err := store.UpdateNotification(
			item.Message.ID,
			item.Reply.ID,
			n.Attempt,
			session.NotificationPending,
			func(v *session.ReplyNotification) { v.RequestID = requestID },
		); err != nil {
			return
		}
	}
	target, found := r.service.sessions.GetSession(item.Message.SenderSessionID)
	inventory, inventoryErr := r.service.scope.inventory()
	if inventoryErr != nil {
		return
	}
	if !found || target.ActivationID != item.Message.SenderActivationID || target.Status == session.StatusTerminated ||
		!target.KankuroMCPEnabled ||
		!inventory.allowsSession(item.Message.DashboardID, item.Message.SenderSessionID) {
		_ = store.UpdateNotification(
			item.Message.ID,
			item.Reply.ID,
			n.Attempt,
			session.NotificationPending,
			func(v *session.ReplyNotification) {
				v.State = session.NotificationUnavailable
				v.FailureCode = notificationStaleActivation
				v.FailureMessage = "original sender activation is unavailable"
			},
		)
		return
	}
	admission, err := r.service.sessions.ReserveReportNotification(item.Message.ID, item.Reply.ID)
	if err != nil {
		var rejected *session.MessagingError
		if errors.As(err, &rejected) && rejected.Code == notificationStaleActivation {
			_ = store.UpdateNotification(
				item.Message.ID,
				item.Reply.ID,
				n.Attempt,
				session.NotificationPending,
				func(v *session.ReplyNotification) {
					v.State = session.NotificationUnavailable
					v.FailureCode = notificationStaleActivation
					v.FailureMessage = "original sender activation is unavailable"
				},
			)
		}
		return
	}
	messageID := uuid.NewSHA1(uuid.NameSpaceURL, []byte("kankuro-report-message:"+requestID)).String()
	intent := session.MessageRecord{
		ID:                    messageID,
		Kind:                  driver.PromptReportNotification,
		OriginalMessageID:     item.Message.ID,
		ReplyID:               item.Reply.ID,
		ReportKind:            item.Reply.Kind,
		DashboardID:           item.Message.DashboardID,
		SenderSessionID:       item.Reply.SenderSessionID,
		SenderActivationID:    item.Reply.SenderActivationID,
		RecipientSessionID:    item.Message.SenderSessionID,
		RecipientActivationID: item.Message.SenderActivationID,
		Text:                  "report notification",
	}
	delivery := &preparedDelivery{service: r.service, store: store, admission: admission, intent: intent, requestID: requestID}
	defer delivery.releaseUnpublished()
	fingerprintData, err := json.Marshal(intent)
	if err != nil {
		return
	}
	delivery.fingerprint = session.SendFingerprint(intent.RecipientSessionID, intent.RecipientActivationID, string(fingerprintData))
	args, err := json.Marshal(delivery.arguments())
	if err != nil {
		return
	}
	authority := notificationAuthority{service: r.service, messageID: item.Message.ID, replyID: item.Reply.ID}
	operation, _, err := r.accept(
		intent.SenderSessionID,
		intent.DashboardID,
		operationKey{caller: intent.SenderSessionID, request: requestID},
		requestID,
		messageOperation,
		intent.RecipientSessionID,
		args,
		operationFingerprint(messageOperation, intent.RecipientSessionID, args),
		authority,
		delivery.submission(),
		delivery.cancel,
		delivery.publish,
	)
	if err == nil && operation.ID != "" {
		_ = store.UpdateNotification(
			item.Message.ID,
			item.Reply.ID,
			n.Attempt,
			session.NotificationAccepted,
			func(v *session.ReplyNotification) { v.OperationID = operation.ID },
		)
	}
}

func (r *OperationRunner) reconcileNotification(item session.PendingNotification) {
	store := r.service.sessions.Messages()
	n := item.Reply.Notification
	if n.OperationID == "" && n.RequestID != "" {
		r.mu.Lock()
		operation := r.records[operationKey{caller: item.Reply.SenderSessionID, request: n.RequestID}]
		operationID := ""
		if operation != nil {
			operationID = operation.ID
		}
		r.mu.Unlock()
		if operationID != "" {
			_ = store.UpdateNotification(
				item.Message.ID,
				item.Reply.ID,
				n.Attempt,
				session.NotificationAccepted,
				func(v *session.ReplyNotification) { v.OperationID = operationID },
			)
		}
	}
	message, err := store.GetMessage(n.MessageID)
	if err != nil {
		_ = store.UpdateNotification(
			item.Message.ID,
			item.Reply.ID,
			n.Attempt,
			session.NotificationAccepted,
			func(v *session.ReplyNotification) {
				v.State = session.NotificationUncertain
				v.FailureCode = "history_unavailable"
				v.FailureMessage = "delivery evidence is unavailable"
			},
		)
		return
	}
	switch message.Delivery {
	case session.MessageAccepted, session.MessageSubmitting:
		return
	case session.MessageSubmitted:
		_ = store.UpdateNotification(
			item.Message.ID,
			item.Reply.ID,
			n.Attempt,
			session.NotificationAccepted,
			func(v *session.ReplyNotification) { v.State = session.NotificationSubmitted },
		)
	case session.MessageUncertain:
		_ = store.UpdateNotification(
			item.Message.ID,
			item.Reply.ID,
			n.Attempt,
			session.NotificationAccepted,
			func(v *session.ReplyNotification) {
				v.State = session.NotificationUncertain
				v.FailureCode = "delivery_uncertain"
				v.FailureMessage = "prompt delivery could not be established safely"
			},
		)
	case session.MessageNotSubmitted:
		_ = store.UpdateNotification(
			item.Message.ID,
			item.Reply.ID,
			n.Attempt,
			session.NotificationAccepted,
			func(v *session.ReplyNotification) {
				v.State = session.NotificationPending
				v.Attempt++
				v.RequestID = ""
				v.MessageID = ""
				v.OperationID = ""
			},
		)
	}
}
