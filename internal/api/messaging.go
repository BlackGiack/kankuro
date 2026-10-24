package api

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/BlackGiack/kankuro/internal/driver"
	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/google/uuid"
)

const (
	messageOperation = "send_message"
	maxReplyWait     = 10 * time.Second
)

type MessageRecord = session.MessageRecord
type MessageReply = session.MessageReply
type ReplyPage = session.ReplyPage

type sendMessageArguments struct {
	RecipientSessionID    string `json:"recipient_session_id"`
	RecipientActivationID string `json:"recipient_activation_id"`
	Fingerprint           string `json:"fingerprint"`
}

// SendMessage admits one idempotent prompt against a pinned live activation.
//
//nolint:gocognit // Idempotency replay must be resolved before target or prompt side effects.
func (s *Service) SendMessage(ctx context.Context, token, requestID, recipientID, activationID, text string) (OperationResult, error) {
	caller, err := s.ResolveCaller(token)
	if err != nil {
		return OperationResult{}, err
	}
	parsed, err := uuid.Parse(requestID)
	if err != nil {
		return OperationResult{}, reject("invalid_request", "request_id must be a UUID")
	}
	requestID = parsed.String()
	if err = driver.ValidatePrompt(text); err != nil {
		return OperationResult{}, reject("invalid_request", err.Error())
	}
	fingerprint := session.SendFingerprint(recipientID, activationID, text)
	args := sendMessageArguments{recipientID, activationID, fingerprint}
	if replay, found, replayErr := s.replayMutation(ctx, token, requestID, messageOperation, recipientID, args); replayErr != nil {
		return OperationResult{}, replayErr
	} else if found {
		return operationResult(replay), nil
	}
	store := s.sessions.Messages()
	if store == nil {
		return OperationResult{}, reject("messaging_unavailable", "messaging history is unavailable")
	}
	if receipt, found, receiptErr := store.FindSend(caller.SessionID, requestID); receiptErr != nil {
		return OperationResult{}, messagingRejection(receiptErr)
	} else if found {
		if receipt.Fingerprint != fingerprint {
			return OperationResult{}, reject("request_id_conflict", "request_id was already used with different delivery arguments")
		}
		return OperationResult{}, reject(
			"history_unavailable",
			"this send was already accepted; its original operation record is unavailable and it will not be submitted again",
		)
	}
	if caller.SessionID == recipientID {
		return OperationResult{}, reject("invalid_request", "cannot send a message to the caller session")
	}
	if !s.scope.AllowsSession(caller.SessionID, recipientID) {
		return OperationResult{}, reject("not_found", "recipient session was not found")
	}
	target, found := s.sessions.GetSession(recipientID)
	if !found || target.Status == session.StatusTerminated || target.LifecycleActivity != "idle" {
		return OperationResult{}, reject("lifecycle_conflict", "recipient session is not available for messages")
	}
	if target.ActivationID != activationID || !target.KankuroMCPEnabled {
		return OperationResult{}, reject("stale_activation", "recipient activation is stale or does not have Kankuro MCP selected")
	}
	if s.operations == nil {
		return OperationResult{}, errors.New("operation runner unavailable")
	}
	admission, admissionErr := s.sessions.ReserveMessageAdmission(token, recipientID, activationID)
	if admissionErr != nil {
		return OperationResult{}, messagingRejection(admissionErr)
	}
	messageID := uuid.NewSHA1(uuid.NameSpaceURL, []byte("kankuro-message:"+caller.SessionID+":"+requestID)).String()
	intent := session.MessageRecord{ID: messageID, DashboardID: caller.DashboardID, SenderSessionID: caller.SessionID,
		SenderActivationID: caller.ActivationID, RecipientSessionID: recipientID, RecipientActivationID: activationID, Text: text}
	acceptedIntent := false
	beforeAccept := func(persistAcceptance func() error) error {
		createErr := s.sessions.PublishMessageAdmission(admission, func() error {
			_, persistErr := store.CreateIntentWithAdmission(intent, requestID, fingerprint, persistAcceptance)
			return persistErr
		})
		if createErr != nil {
			return messagingRejection(createErr)
		}
		acceptedIntent = true
		return nil
	}
	operation, submitErr := s.operations.SubmitWithAdmission(ctx, token, requestID, messageOperation, recipientID, args, beforeAccept,
		func(operationCtx context.Context, _ Operation) (json.RawMessage, error) {
			record, sendErr := s.sessions.SubmitMessage(operationCtx, admission, token,
				messageID, recipientID, activationID, text)
			effects := OperationEffects{MessageID: messageID, Delivery: string(record.Delivery)}
			encoded, marshalErr := json.Marshal(effects)
			if marshalErr != nil {
				return nil, marshalErr
			}
			if sendErr != nil && record.Delivery != session.MessageNotSubmitted && record.Delivery != "" {
				return encoded, operationWarningError{
					"prompt delivery is uncertain; partial input may remain in the CLI input buffer; inspect get_message before deciding what to do",
				}
			}
			return encoded, errors.Join(sendErr, marshalErr)
		}, func(Operation) {
			_ = store.UpdateDelivery(messageID, session.MessageNotSubmitted, "operation was canceled before prompt submission began")
			admission.Release()
		})
	if !acceptedIntent {
		admission.Release()
	}
	if operation.ID == "" {
		return operationResult(operation), submitErr
	}
	result := operationResult(operation)
	if result.Result == nil {
		result.Result = &OperationEffects{MessageID: messageID, Delivery: string(session.MessageAccepted)}
	}
	return result, submitErr
}

// GetMessage returns retained message content to an authorized dashboard controller.
func (s *Service) GetMessage(token, messageID string) (session.MessageRecord, error) {
	caller, err := s.ResolveCaller(token)
	if err != nil {
		return session.MessageRecord{}, err
	}
	store := s.sessions.Messages()
	if store == nil {
		return session.MessageRecord{}, reject("history_unavailable", "messaging history is unavailable")
	}
	message, err := store.GetMessage(messageID)
	if err != nil {
		return session.MessageRecord{}, messagingRejection(err)
	}
	if message.DashboardID != caller.DashboardID {
		return session.MessageRecord{}, reject("not_found", "message was not found")
	}
	return message, nil
}

// ReportReply stores an explicit report from the authenticated recipient activation.
func (s *Service) ReportReply(token, requestID, messageID, kind, text string) (session.MessageReply, error) {
	parsed, err := uuid.Parse(requestID)
	if err != nil {
		return session.MessageReply{}, reject("invalid_request", "request_id must be a UUID")
	}
	if messageID == "" {
		return session.MessageReply{}, reject("invalid_request", "message_id is required")
	}
	reply, err := s.sessions.ReportMessage(token, parsed.String(), messageID, kind, text)
	return reply, messagingRejection(err)
}

// GetReplies returns a bounded page, optionally waiting up to ten seconds.
//
//nolint:gocognit // Authorization is rechecked through each bounded wait and retry.
func (s *Service) GetReplies(
	ctx context.Context,
	token, messageID, cursor string,
	limit int,
	wait time.Duration,
) (session.ReplyPage, error) {
	if wait < 0 || wait > maxReplyWait {
		return session.ReplyPage{}, reject("invalid_request", "wait_seconds must be between zero and ten")
	}
	caller, err := s.ResolveCaller(token)
	if err != nil {
		return session.ReplyPage{}, err
	}
	store := s.sessions.Messages()
	if store == nil {
		return session.ReplyPage{}, reject("history_unavailable", "messaging history is unavailable")
	}
	if wait == 0 {
		page, pageErr := store.GetReplies(caller.DashboardID, messageID, cursor, limit)
		if pageErr != nil {
			return session.ReplyPage{}, messagingRejection(pageErr)
		}
		if _, err = s.ResolveCaller(token); err != nil {
			return session.ReplyPage{}, err
		}
		return page, nil
	}
	if !store.AcquireWait() {
		return session.ReplyPage{}, reject("too_many_waiters", "the dashboard already has 32 waiting reply requests")
	}
	defer store.ReleaseWait()
	deadline := time.NewTimer(wait)
	defer deadline.Stop()
	for {
		changed := store.Changed()
		page, pageErr := store.GetReplies(caller.DashboardID, messageID, cursor, limit)
		if pageErr != nil {
			return session.ReplyPage{}, messagingRejection(pageErr)
		}
		if len(page.Items) > 0 {
			if _, err = s.ResolveCaller(token); err != nil {
				return session.ReplyPage{}, err
			}
			return page, nil
		}
		if messageID != "" {
			live, liveErr := s.sessions.MessageRecipientLive(messageID)
			if liveErr != nil {
				return session.ReplyPage{}, messagingRejection(liveErr)
			}
			if !live {
				return session.ReplyPage{}, reject(
					"recipient_unavailable",
					"the message recipient activation has terminated or been replaced",
				)
			}
		}
		select {
		case <-ctx.Done():
			return session.ReplyPage{}, ctx.Err()
		case <-deadline.C:
			if _, err = s.ResolveCaller(token); err != nil {
				return session.ReplyPage{}, err
			}
			if messageID != "" {
				live, liveErr := s.sessions.MessageRecipientLive(messageID)
				if liveErr != nil {
					return session.ReplyPage{}, messagingRejection(liveErr)
				}
				if !live {
					return session.ReplyPage{}, reject(
						"recipient_unavailable",
						"the message recipient activation has terminated or been replaced",
					)
				}
			}
			return page, nil
		case <-changed:
			caller, err = s.ResolveCaller(token)
			if err != nil {
				return session.ReplyPage{}, err
			}
		}
	}
}

func messagingRejection(err error) error {
	if messagingErr, ok := errors.AsType[*session.MessagingError](err); ok {
		return reject(messagingErr.Code, messagingErr.Message)
	}
	return err
}
