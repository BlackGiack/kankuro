package session

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/BlackGiack/kankuro/internal/driver"
	"github.com/BlackGiack/kankuro/internal/fsutil"
	"github.com/google/uuid"
)

const (
	messageHistoryLimitBytes = 64 << 20
	messageHistoryLimitCount = 10000
	messageReportLimitBytes  = 16 << 10
	messageReportsPerMessage = 128
	maxReplyPageSize         = 100
	maxReplyWaiters          = 32
	messageFileMode          = 0o600
)

// MessageDelivery records process input evidence, not proof of CLI processing.
type MessageDelivery string

const (
	MessageAccepted     MessageDelivery = "accepted"
	MessageSubmitting   MessageDelivery = "submitting"
	MessageSubmitted    MessageDelivery = "submitted"
	MessageNotSubmitted MessageDelivery = "not_submitted"
	MessageUncertain    MessageDelivery = "uncertain"
)

// MessageRecord is retained history for one orchestrator prompt.
type MessageRecord struct {
	ID                    string          `json:"message_id"`
	DashboardID           string          `json:"dashboard_id"`
	SenderSessionID       string          `json:"sender_session_id"`
	SenderActivationID    string          `json:"sender_activation_id"`
	RecipientSessionID    string          `json:"recipient_session_id"`
	RecipientActivationID string          `json:"recipient_activation_id"`
	Text                  string          `json:"text"`
	Delivery              MessageDelivery `json:"delivery"`
	Warning               string          `json:"warning,omitempty"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
	Replies               []MessageReply  `json:"replies,omitempty"`
	Sequence              uint64          `json:"sequence"`
}

// MessageReply is an explicit report from the pinned recipient activation.
type MessageReply struct {
	ID                 string    `json:"reply_id"`
	MessageID          string    `json:"message_id"`
	SenderSessionID    string    `json:"sender_session_id"`
	SenderActivationID string    `json:"sender_activation_id"`
	Kind               string    `json:"kind"`
	Text               string    `json:"text"`
	CreatedAt          time.Time `json:"created_at"`
	Sequence           uint64    `json:"sequence"`
}

type sendReceipt struct {
	SenderSessionID string `json:"sender_session_id"`
	RequestID       string `json:"request_id"`
	Fingerprint     string `json:"fingerprint"`
	MessageID       string `json:"message_id"`
}
type reportReceipt struct {
	RecipientSessionID    string `json:"recipient_session_id"`
	RecipientActivationID string `json:"recipient_activation_id"`
	MessageID             string `json:"message_id"`
	RequestID             string `json:"request_id"`
	Fingerprint           string `json:"fingerprint"`
	ReplyID               string `json:"reply_id"`
}
type messageStoreData struct {
	Messages             []MessageRecord   `json:"messages"`
	Sends                []sendReceipt     `json:"send_receipts,omitempty"`
	Reports              []reportReceipt   `json:"report_receipts,omitempty"`
	Sequence             uint64            `json:"sequence"`
	DashboardGenerations map[string]uint64 `json:"dashboard_generations,omitempty"`
}

// MessageStore serializes messaging history independently from the manager lock.
type MessageStore struct {
	mu              sync.Mutex
	path            string
	data            messageStoreData
	changed         chan struct{}
	waiters         int
	runtimeDelivery map[string]runtimeDeliveryOutcome
}

type runtimeDeliveryOutcome struct {
	state   MessageDelivery
	warning string
}

// MessageAdmission pins sender and recipient sessions until submission begins
// or lifecycle teardown invalidates the pending operation.
type MessageAdmission struct {
	manager               *Manager
	sender                *Session
	recipient             *Session
	authority             *activationAuthority
	senderID              string
	senderActivationID    string
	recipientID           string
	recipientActivationID string
	token                 string
	ctx                   context.Context
	cancel                context.CancelFunc
	started               bool
	invalidated           bool
	publishing            bool
	published             bool
	released              bool
}

// Release drops the lifecycle reservation. It is safe to call more than once.
func (a *MessageAdmission) Release() {
	if a == nil || a.manager == nil {
		return
	}
	m := a.manager
	m.mu.Lock()
	defer m.mu.Unlock()
	m.releaseMessageAdmissionLocked(a)
}

// MessagingError is a safe, stable failure category for API transports.
type MessagingError struct {
	Code    string
	Message string
	Cause   error
}

func (e *MessagingError) Error() string { return e.Message }
func (e *MessagingError) Unwrap() error { return e.Cause }

func messagingError(code, message string, cause error) error {
	return &MessagingError{Code: code, Message: message, Cause: cause}
}

// NewMessageStore loads the durable message journal from a private directory.
func NewMessageStore(configDir string) (*MessageStore, error) {
	dir := filepath.Join(configDir, "messaging")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("create messaging directory: %w", err)
	}
	store := &MessageStore{
		path:            filepath.Join(dir, "history.json"),
		changed:         make(chan struct{}),
		data:            messageStoreData{DashboardGenerations: make(map[string]uint64)},
		runtimeDelivery: make(map[string]runtimeDeliveryOutcome),
	}
	bytes, err := os.ReadFile(store.path)
	if errors.Is(err, os.ErrNotExist) {
		return store, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read messaging history: %w", err)
	}
	if err = json.Unmarshal(bytes, &store.data); err != nil {
		return nil, fmt.Errorf("decode messaging history: %w", err)
	}
	if store.data.DashboardGenerations == nil {
		store.data.DashboardGenerations = make(map[string]uint64)
	}
	return store, nil
}

// SendFingerprint returns a stable digest of the complete delivery target and content.
func SendFingerprint(recipientSessionID, activationID, text string) string {
	sum := sha256.Sum256([]byte(recipientSessionID + "\x00" + activationID + "\x00" + text))
	return hex.EncodeToString(sum[:])
}

// FindSend returns a prior receipt for stable request replay.
func (s *MessageStore) FindSend(senderID, requestID string) (sendReceipt, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, receipt := range s.data.Sends {
		if receipt.SenderSessionID == senderID && receipt.RequestID == requestID {
			return receipt, true, nil
		}
	}
	return sendReceipt{}, false, nil
}

// RollbackIntent removes an intent only when operation admission itself failed.
func (s *MessageStore) RollbackIntent(senderID, requestID, messageID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, message := range s.data.Messages {
		if message.ID == messageID && message.SenderSessionID == senderID {
			s.data.Messages = append(s.data.Messages[:i], s.data.Messages[i+1:]...)
			break
		}
	}
	for i, receipt := range s.data.Sends {
		if receipt.SenderSessionID == senderID && receipt.RequestID == requestID {
			s.data.Sends = append(s.data.Sends[:i], s.data.Sends[i+1:]...)
			break
		}
	}
	if err := s.saveLocked(); err != nil {
		// Admission never reached a worker, so forgetting it in memory is safe;
		// startup recovery will mark a stale on-disk intent uncertain.
		s.signalLocked()
		return err
	}
	s.signalLocked()
	return nil
}

// WaitTurn keeps prompts for one activation in durable acceptance order.
func (s *MessageStore) WaitTurn(ctx context.Context, messageID string) error {
	for {
		s.mu.Lock()
		var current *MessageRecord
		for i := range s.data.Messages {
			if s.data.Messages[i].ID == messageID {
				current = &s.data.Messages[i]
				break
			}
		}
		if current == nil {
			s.mu.Unlock()
			return errors.New("message history is unavailable")
		}
		blocked := false
		for i := range s.data.Messages {
			previous := &s.data.Messages[i]
			if s.messageBlocksTurnLocked(previous, current) {
				blocked = true
				break
			}
		}
		changed := s.changed
		s.mu.Unlock()
		if !blocked {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-changed:
		}
	}
}

func (s *MessageStore) messageBlocksTurnLocked(previous, current *MessageRecord) bool {
	if previous.RecipientSessionID != current.RecipientSessionID ||
		previous.RecipientActivationID != current.RecipientActivationID || previous.Sequence >= current.Sequence {
		return false
	}
	if previous.Delivery != MessageAccepted && previous.Delivery != MessageSubmitting {
		return false
	}
	return !s.runtimeDeliveryFinishedLocked(previous.ID)
}

// CreateIntent durably records content and a compact request receipt before PTY input.
func (s *MessageStore) CreateIntent(record MessageRecord, requestID, fingerprint string) (MessageRecord, error) {
	return s.createIntent(record, requestID, fingerprint, nil)
}

// CreateIntentWithAdmission persists intent and retention changes as one
// transaction with durable operation admission.
func (s *MessageStore) CreateIntentWithAdmission(
	record MessageRecord,
	requestID, fingerprint string,
	persistAcceptance func() error,
) (MessageRecord, error) {
	return s.createIntent(record, requestID, fingerprint, persistAcceptance)
}

//nolint:gocognit // Persistence and operation admission are one rollback-capable history transaction.
func (s *MessageStore) createIntent(
	record MessageRecord,
	requestID, fingerprint string,
	persistAcceptance func() error,
) (MessageRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, receipt := range s.data.Sends {
		if receipt.SenderSessionID == record.SenderSessionID && receipt.RequestID == requestID {
			if receipt.Fingerprint != fingerprint {
				return MessageRecord{}, messagingError(
					"request_id_conflict",
					"request_id was already used with different delivery arguments",
					nil,
				)
			}
			for _, existing := range s.data.Messages {
				if existing.ID == receipt.MessageID {
					return existing, nil
				}
			}
			return MessageRecord{
				ID:       receipt.MessageID,
				Delivery: MessageUncertain,
			}, messagingError("history_unavailable", "message history was evicted or deleted; retry will not submit it again", nil)
		}
	}
	if record.ID == "" {
		record.ID = uuid.NewString()
	}
	if record.CreatedAt.IsZero() {
		record.CreatedAt = time.Now().UTC()
	}
	record.UpdatedAt = record.CreatedAt
	record.Delivery = MessageAccepted
	old, cloneErr := cloneMessageData(s.data)
	if cloneErr != nil {
		return MessageRecord{}, cloneErr
	}
	s.data.Sequence++
	record.Sequence = s.data.Sequence
	s.data.Messages = append(s.data.Messages, record)
	s.data.Sends = append(s.data.Sends, sendReceipt{record.SenderSessionID, requestID, fingerprint, record.ID})
	s.evictLocked()
	if err := s.saveLocked(); err != nil {
		s.data = old
		return MessageRecord{}, messagingError(
			"storage_unavailable",
			"message history could not be saved; retry with the same request_id",
			err,
		)
	}
	if persistAcceptance != nil {
		if err := persistAcceptance(); err != nil {
			s.data = old
			rollbackErr := s.saveLocked()
			s.signalLocked()
			return MessageRecord{}, errors.Join(
				messagingError("journal_unavailable", "operation admission could not be saved; retry with the same request_id", err),
				rollbackErr,
			)
		}
	}
	s.pruneRuntimeDeliveryLocked()
	s.signalLocked()
	return record, nil
}

// UpdateDelivery persists state evidence for an admitted prompt.
func (s *MessageStore) UpdateDelivery(messageID string, state MessageDelivery, warning string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.data.Messages {
		if s.data.Messages[i].ID != messageID {
			continue
		}
		old, cloneErr := cloneMessageData(s.data)
		if cloneErr != nil {
			return cloneErr
		}
		s.data.Messages[i].Delivery, s.data.Messages[i].Warning = state, warning
		s.data.Messages[i].UpdatedAt = time.Now().UTC()
		s.evictLocked()
		if err := s.saveLocked(); err != nil {
			s.data = old
			s.rememberUnsavedDeliveryLocked(messageID, state, warning)
			return messagingError("storage_unavailable", "message delivery state could not be saved", err)
		}
		s.pruneRuntimeDeliveryLocked()
		delete(s.runtimeDelivery, messageID)
		s.signalLocked()
		return nil
	}
	return messagingError("history_unavailable", "message history is unavailable", nil)
}

func (s *MessageStore) rememberUnsavedDeliveryLocked(messageID string, state MessageDelivery, warning string) {
	if !isTerminalMessageDelivery(state) {
		return
	}
	if s.runtimeDelivery == nil {
		s.runtimeDelivery = make(map[string]runtimeDeliveryOutcome)
	}
	outcome := runtimeDeliveryOutcome{state: state, warning: warning}
	if state == MessageSubmitted {
		outcome.state = MessageUncertain
		outcome.warning = "prompt bytes were written but the delivery outcome could not be saved"
	}
	s.runtimeDelivery[messageID] = outcome
	s.signalLocked()
}

func isTerminalMessageDelivery(state MessageDelivery) bool {
	return state == MessageSubmitted || state == MessageNotSubmitted || state == MessageUncertain
}

func (s *MessageStore) runtimeDeliveryFinishedLocked(messageID string) bool {
	_, found := s.runtimeDelivery[messageID]
	return found
}

// RecordReply persists an idempotent report before returning its stable identity.
//
//nolint:funlen,gocognit,nestif // Idempotency, activation binding, quota enforcement, and durable append are atomic.
func (s *MessageStore) RecordReply(recipientID, activationID, requestID, messageID, kind, text string) (MessageReply, error) {
	if !utf8.ValidString(text) || len(text) == 0 {
		return MessageReply{}, messagingError("invalid_request", "report text must be nonempty UTF-8", nil)
	}
	if len([]byte(text)) > messageReportLimitBytes {
		return MessageReply{}, messagingError("invalid_request", "report text exceeds 16 KiB", nil)
	}
	if kind != "completed" && kind != "question" && kind != "blocked" {
		return MessageReply{}, messagingError("invalid_request", "kind must be completed, question, or blocked", nil)
	}
	fingerprint := SendFingerprint(messageID, kind, text)
	s.mu.Lock()
	defer s.mu.Unlock()
	messageIndex := -1
	for i := range s.data.Messages {
		if s.data.Messages[i].ID == messageID {
			messageIndex = i
			break
		}
	}
	for _, receipt := range s.data.Reports {
		if receipt.RecipientSessionID == recipientID && receipt.MessageID == messageID && receipt.RequestID == requestID {
			if messageIndex >= 0 {
				message := s.data.Messages[messageIndex]
				if message.RecipientSessionID != recipientID || message.RecipientActivationID != activationID {
					return MessageReply{}, messagingError("stale_activation", "message is not assigned to this activation", nil)
				}
			}
			if receipt.RecipientActivationID != activationID {
				return MessageReply{}, messagingError("stale_activation", "report request belongs to a different recipient activation", nil)
			}
			if receipt.Fingerprint != fingerprint {
				return MessageReply{}, messagingError(
					"request_id_conflict",
					"request_id was already used with different report content",
					nil,
				)
			}
			if messageIndex < 0 {
				return MessageReply{}, messagingError(
					"history_unavailable",
					"message history is unavailable; report was previously accepted",
					nil,
				)
			}
			for _, reply := range s.data.Messages[messageIndex].Replies {
				if reply.ID == receipt.ReplyID {
					return reply, nil
				}
			}
			return MessageReply{}, messagingError(
				"history_unavailable",
				"message history is unavailable; report was previously accepted",
				nil,
			)
		}
	}
	if messageIndex < 0 {
		return MessageReply{}, messagingError("history_unavailable", "message history is unavailable", nil)
	}
	message := &s.data.Messages[messageIndex]
	if message.RecipientSessionID != recipientID || message.RecipientActivationID != activationID {
		return MessageReply{}, messagingError("stale_activation", "message is not assigned to this activation", nil)
	}
	if len(message.Replies) >= messageReportsPerMessage {
		return MessageReply{}, messagingError("capacity_exhausted", "message already has 128 reports", nil)
	}
	old, err := cloneMessageData(s.data)
	if err != nil {
		return MessageReply{}, err
	}
	s.data.Sequence++
	reply := MessageReply{ID: uuid.NewString(), MessageID: messageID, SenderSessionID: recipientID,
		SenderActivationID: activationID, Kind: kind, Text: text, CreatedAt: time.Now().UTC(), Sequence: s.data.Sequence}
	message.Replies = append(message.Replies, reply)
	s.data.Reports = append(s.data.Reports, reportReceipt{RecipientSessionID: recipientID, RecipientActivationID: activationID,
		MessageID: messageID, RequestID: requestID, Fingerprint: fingerprint, ReplyID: reply.ID})
	s.evictLocked()
	if err = s.saveLocked(); err != nil {
		s.data = old
		return MessageReply{}, messagingError("storage_unavailable", "report could not be saved; retry with the same request_id", err)
	}
	s.pruneRuntimeDeliveryLocked()
	s.signalLocked()
	return reply, nil
}

// GetMessage returns a retained message or a clear unavailable error.
func (s *MessageStore) GetMessage(id string) (MessageRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, message := range s.data.Messages {
		if message.ID == id {
			if outcome, found := s.runtimeDelivery[id]; found {
				message.Delivery = outcome.state
				message.Warning = outcome.warning
			}
			return message, nil
		}
	}
	return MessageRecord{}, messagingError(
		"history_unavailable",
		"message content is unavailable because it was evicted or its recipient session was deleted",
		nil,
	)
}

func (s *MessageStore) MessageRecipient(messageID string) (string, string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, message := range s.data.Messages {
		if message.ID == messageID {
			return message.RecipientSessionID, message.RecipientActivationID, nil
		}
	}
	return "", "", messagingError("history_unavailable", "message history is unavailable", nil)
}

// MessageCursor is an opaque continuation bound to a dashboard and message filter.
type MessageCursor struct {
	DashboardID string `json:"d"`
	MessageID   string `json:"m"`
	Sequence    uint64 `json:"s"`
	Generation  uint64 `json:"g"`
}
type ReplyPage struct {
	Items      []MessageReply `json:"items"`
	NextCursor string         `json:"next_cursor,omitempty"`
	HasMore    bool           `json:"has_more"`
	Gap        bool           `json:"gap"`
}

// GetReplies returns a bounded page and reports cursor gaps after retention changes.
//
//nolint:gocognit,funlen // Cursor validation and bounded paging use one consistent store snapshot.
func (s *MessageStore) GetReplies(dashboardID, messageID, cursor string, limit int) (ReplyPage, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > maxReplyPageSize {
		limit = maxReplyPageSize
	}
	start, generation := uint64(0), uint64(0)
	if cursor != "" {
		decoded, err := base64.RawURLEncoding.DecodeString(cursor)
		if err != nil {
			return ReplyPage{}, messagingError("invalid_cursor", "cursor is malformed", err)
		}
		var parsed MessageCursor
		if json.Unmarshal(decoded, &parsed) != nil {
			return ReplyPage{}, messagingError("invalid_cursor", "cursor is malformed", nil)
		}
		if parsed.DashboardID != dashboardID || parsed.MessageID != messageID {
			return ReplyPage{}, messagingError("invalid_cursor", "cursor does not match this dashboard or message filter", nil)
		}
		start, generation = parsed.Sequence, parsed.Generation
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	currentGeneration := s.data.DashboardGenerations[dashboardID]
	gap := cursor != "" && generation != currentGeneration
	items := make([]MessageReply, 0, limit)
	all := make([]MessageReply, 0)
	for _, message := range s.data.Messages {
		if message.DashboardID == dashboardID && (messageID == "" || message.ID == messageID) {
			all = append(all, message.Replies...)
		}
	}
	sort.Slice(all, func(i, j int) bool { return all[i].Sequence < all[j].Sequence })
	oldest := uint64(0)
	if len(all) > 0 {
		oldest = all[0].Sequence
	}
	if gap && oldest > 0 {
		start = oldest - 1
	}
	byteCount := 0
	for _, reply := range all {
		if reply.Sequence <= start {
			continue
		}
		if len(items) == limit {
			break
		}
		encoded, _ := json.Marshal(reply)
		if byteCount+len(encoded) > 256<<10 {
			break
		}
		items = append(items, reply)
		byteCount += len(encoded)
	}
	hasMore := false
	if len(items) > 0 {
		last := items[len(items)-1].Sequence
		for _, reply := range all {
			if reply.Sequence > last {
				hasMore = true
				break
			}
		}
	}
	page := ReplyPage{Items: items, HasMore: hasMore, Gap: gap}
	if len(items) > 0 {
		next, _ := json.Marshal(
			MessageCursor{
				DashboardID: dashboardID,
				MessageID:   messageID,
				Sequence:    items[len(items)-1].Sequence,
				Generation:  currentGeneration,
			},
		)
		page.NextCursor = base64.RawURLEncoding.EncodeToString(next)
	}
	return page, nil
}

// Changed returns the current notification channel for optional report waits.
func (s *MessageStore) Changed() <-chan struct{} { s.mu.Lock(); defer s.mu.Unlock(); return s.changed }

// Wake interrupts bounded waiters when lifecycle or authorization state changes.
func (s *MessageStore) Wake() { s.mu.Lock(); defer s.mu.Unlock(); s.signalLocked() }

// AcquireWait bounds report waiters across every frontend sharing this manager.
func (s *MessageStore) AcquireWait() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.waiters >= maxReplyWaiters {
		return false
	}
	s.waiters++
	return true
}

// ReleaseWait releases one bounded report waiter slot.
func (s *MessageStore) ReleaseWait() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.waiters > 0 {
		s.waiters--
	}
}

// DeleteRecipient removes incoming history after successful session deletion.
func (s *MessageStore) DeleteRecipient(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	old, err := cloneMessageData(s.data)
	if err != nil {
		return err
	}
	kept := s.data.Messages[:0]
	removed := false
	for _, message := range s.data.Messages {
		if message.RecipientSessionID == sessionID {
			removed = true
			s.data.DashboardGenerations[message.DashboardID]++
			continue
		}
		kept = append(kept, message)
	}
	s.data.Messages = kept
	if !removed {
		return nil
	}
	s.pruneReportsLocked()
	if saveErr := s.saveLocked(); saveErr != nil {
		s.data = old
		return saveErr
	}
	s.pruneRuntimeDeliveryLocked()
	s.signalLocked()
	return nil
}

// Recover marks unfinished submissions uncertain and drops records for deleted recipients.
func (s *MessageStore) Recover(sessions map[string]*Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	old, cloneErr := cloneMessageData(s.data)
	if cloneErr != nil {
		return cloneErr
	}
	kept := s.data.Messages[:0]
	changed := false
	retained := make(map[string]struct{}, len(s.data.Messages))
	for _, message := range s.data.Messages {
		if sessions[message.RecipientSessionID] == nil {
			changed = true
			s.data.DashboardGenerations[message.DashboardID]++
			continue
		}
		if message.Delivery == MessageAccepted || message.Delivery == MessageSubmitting {
			message.Delivery = MessageUncertain
			message.Warning = "Kankuro restarted while submission was in progress; input may have been delivered"
			message.UpdatedAt = time.Now().UTC()
			changed = true
		}
		kept = append(kept, message)
		retained[message.ID] = struct{}{}
	}
	s.data.Messages = kept
	if !changed {
		return nil
	}
	s.pruneReportsToMessagesLocked(retained)
	s.evictLocked()
	if err := s.saveLocked(); err != nil {
		s.data = old
		return err
	}
	s.signalLocked()
	return nil
}

func cloneMessageData(data messageStoreData) (messageStoreData, error) {
	encoded, err := json.Marshal(data)
	if err != nil {
		return messageStoreData{}, err
	}
	var clone messageStoreData
	if err = json.Unmarshal(encoded, &clone); err != nil {
		return messageStoreData{}, err
	}
	return clone, nil
}

//nolint:gocognit // Retention limits, oldest-message selection, generations, and receipt pruning are one locked mutation.
func (s *MessageStore) evictLocked() {
	removed := make(map[string]struct{})
	for {
		counts, sizes := make(map[string]int), make(map[string]int)
		overLimit := ""
		for _, message := range s.data.Messages {
			counts[message.DashboardID]++
			encoded, _ := json.Marshal(message)
			sizes[message.DashboardID] += len(encoded)
			if counts[message.DashboardID] > messageHistoryLimitCount || sizes[message.DashboardID] > messageHistoryLimitBytes {
				overLimit = message.DashboardID
			}
		}
		if overLimit == "" {
			break
		}
		oldest := -1
		for i, message := range s.data.Messages {
			if message.DashboardID == overLimit && (oldest < 0 || message.Sequence < s.data.Messages[oldest].Sequence) {
				oldest = i
			}
		}
		if oldest < 0 {
			break
		}
		s.data.DashboardGenerations[overLimit]++
		removed[s.data.Messages[oldest].ID] = struct{}{}
		s.data.Messages = append(s.data.Messages[:oldest], s.data.Messages[oldest+1:]...)
	}
	if len(removed) > 0 {
		kept := s.data.Reports[:0]
		for _, receipt := range s.data.Reports {
			if _, found := removed[receipt.MessageID]; !found {
				kept = append(kept, receipt)
			}
		}
		s.data.Reports = kept
	}
}

func (s *MessageStore) pruneRuntimeDeliveryLocked() {
	retained := make(map[string]struct{}, len(s.data.Messages))
	for _, message := range s.data.Messages {
		retained[message.ID] = struct{}{}
	}
	for messageID := range s.runtimeDelivery {
		if _, found := retained[messageID]; !found {
			delete(s.runtimeDelivery, messageID)
		}
	}
}

func (s *MessageStore) pruneReportsLocked() {
	retained := make(map[string]struct{}, len(s.data.Messages))
	for _, message := range s.data.Messages {
		retained[message.ID] = struct{}{}
	}
	s.pruneReportsToMessagesLocked(retained)
}

func (s *MessageStore) pruneReportsToMessagesLocked(retained map[string]struct{}) {
	kept := s.data.Reports[:0]
	for _, receipt := range s.data.Reports {
		if _, found := retained[receipt.MessageID]; found {
			kept = append(kept, receipt)
		}
	}
	s.data.Reports = kept
}
func (s *MessageStore) saveLocked() error {
	encoded, err := json.Marshal(s.data)
	if err != nil {
		return err
	}
	return fsutil.WriteFileAtomic(s.path, encoded, messageFileMode)
}
func (s *MessageStore) signalLocked() { close(s.changed); s.changed = make(chan struct{}) }

// ReserveMessageAdmission validates and pins both activations before history
// admission, so deletion and replacement cannot race the durable intent.
func (m *Manager) ReserveMessageAdmission(token, recipientID, recipientActivationID string) (*MessageAdmission, error) {
	caller, err := m.ResolveDashboardCaller(token)
	if err != nil {
		return nil, messagingError("forbidden", "caller activation is no longer authorized", err)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	authority := m.authorities[token]
	sender := m.sessions[caller.SessionID]
	recipient := m.sessions[recipientID]
	if m.closed || authority == nil || authority.revoked || !authority.published || sender == nil ||
		authority.sessionID != caller.SessionID || authority.activationID != caller.ActivationID ||
		sender.ActivationID != caller.ActivationID || !sender.DashboardControl || sender.Status == StatusTerminated ||
		sender.stopping || sender.activating || sender.deleting {
		return nil, messagingError("forbidden", "caller activation is no longer authorized", nil)
	}
	if recipient == nil || recipientID == caller.SessionID || recipient.Status == StatusTerminated || recipient.stopping ||
		recipient.activating || recipient.deleting {
		return nil, messagingError("lifecycle_conflict", "recipient session is not available for messages", nil)
	}
	if recipient.ActivationID != recipientActivationID || !recipient.KankuroMCPEnabled {
		return nil, messagingError("stale_activation", "recipient activation is stale or does not have Kankuro MCP selected", nil)
	}
	if recipient.Driver == nil {
		return nil, messagingError("lifecycle_conflict", "recipient driver is unavailable", nil)
	}
	admissionCtx, cancel := context.WithCancel(context.Background())
	for _, current := range []*Session{sender, recipient} {
		if current.activeSubmissions == 0 {
			current.submissionsDone = make(chan struct{})
		}
		current.activeSubmissions++
	}
	admission := &MessageAdmission{
		manager: m, sender: sender, recipient: recipient, authority: authority,
		senderID: caller.SessionID, senderActivationID: caller.ActivationID,
		recipientID: recipientID, recipientActivationID: recipientActivationID, token: token,
		ctx: admissionCtx, cancel: cancel,
	}
	if m.messageAdmissions == nil {
		m.messageAdmissions = make(map[*MessageAdmission]struct{})
	}
	m.messageAdmissions[admission] = struct{}{}
	return admission, nil
}

func (m *Manager) releaseMessageAdmissionLocked(admission *MessageAdmission) {
	if admission.released {
		return
	}
	admission.released = true
	delete(m.messageAdmissions, admission)
	if admission.cancel != nil {
		admission.cancel()
	}
	for _, current := range []*Session{admission.sender, admission.recipient} {
		if current == nil || current.activeSubmissions == 0 {
			continue
		}
		current.activeSubmissions--
		if current.activeSubmissions == 0 {
			close(current.submissionsDone)
			current.submissionsDone = nil
		}
	}
}

// cancelPendingMessageAdmissionsLocked invalidates reservations that have not
// reached PTY submission, so lifecycle workers never wait for queued work.
func (m *Manager) cancelPendingMessageAdmissionsLocked(session *Session) {
	for admission := range m.messageAdmissions {
		if admission.started || admission.publishing || (admission.sender != session && admission.recipient != session) {
			continue
		}
		admission.invalidated = true
		m.releaseMessageAdmissionLocked(admission)
	}
}

// PublishMessageAdmission coordinates durable intent publication with lifecycle
// invalidation. A publisher already in progress holds its reservation until the
// publication transaction finishes; a lifecycle operation can invalidate it
// before publication starts.
func (m *Manager) PublishMessageAdmission(admission *MessageAdmission, publish func() error) error {
	if admission == nil || admission.manager != m || publish == nil {
		return messagingError("lifecycle_conflict", "message admission is no longer available", nil)
	}
	m.mu.Lock()
	if admission.invalidated || admission.released || admission.published || admission.publishing || m.closed ||
		m.sessions[admission.senderID] != admission.sender || m.sessions[admission.recipientID] != admission.recipient ||
		admission.sender.stopping || admission.sender.deleting || admission.sender.Status == StatusTerminated ||
		admission.recipient.stopping || admission.recipient.deleting || admission.recipient.Status == StatusTerminated ||
		admission.sender.ActivationID != admission.senderActivationID || admission.recipient.ActivationID != admission.recipientActivationID {
		m.mu.Unlock()
		return messagingError("lifecycle_conflict", "message admission was canceled before intent publication", nil)
	}
	admission.publishing = true
	m.mu.Unlock()

	err := publish()

	m.mu.Lock()
	admission.publishing = false
	if err == nil {
		admission.published = true
	}
	if m.closed || admission.sender.stopping || admission.sender.deleting || admission.sender.Status == StatusTerminated ||
		admission.recipient.stopping || admission.recipient.deleting || admission.recipient.Status == StatusTerminated {
		admission.invalidated = true
		m.releaseMessageAdmissionLocked(admission)
	}
	m.mu.Unlock()
	return err
}

func (m *Manager) beginMessageSubmission(admission *MessageAdmission, token, recipientID, activationID string) (driver.AgentDriver, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if admission == nil || admission.manager != m || admission.invalidated || admission.released ||
		!admission.published || admission.started ||
		admission.token != token || admission.recipientID != recipientID || admission.recipientActivationID != activationID {
		return nil, false
	}
	sender, recipient := admission.sender, admission.recipient
	if m.closed || sender == nil || recipient == nil || sender.stopping || sender.deleting || recipient.stopping ||
		recipient.deleting || sender.Status == StatusTerminated || recipient.Status == StatusTerminated ||
		m.sessions[sender.ID] != sender || m.sessions[recipientID] != recipient ||
		sender.ActivationID != admission.senderActivationID || recipient.ActivationID != activationID || !sender.DashboardControl {
		return nil, false
	}
	authority := m.authorities[token]
	if authority != admission.authority || authority == nil || authority.revoked || !authority.published {
		return nil, false
	}
	admission.started = true
	return recipient.Driver, recipient.Driver != nil
}

// SubmitMessage admits a prompt against a pinned live recipient activation.
func (m *Manager) SubmitMessage(
	ctx context.Context,
	admission *MessageAdmission,
	token, messageID, recipientID, recipientActivationID, text string,
) (MessageRecord, error) {
	if m.messages == nil {
		return MessageRecord{}, errors.New("messaging history is unavailable")
	}
	failBeforeWrite := func(message string, cause error) (MessageRecord, error) {
		persistErr := m.messages.UpdateDelivery(messageID, MessageNotSubmitted, message)
		record, getErr := m.messages.GetMessage(messageID)
		if record.ID == "" {
			record = MessageRecord{ID: messageID, Delivery: MessageNotSubmitted, Warning: message}
		}
		return record, errors.Join(cause, persistErr, getErr)
	}
	if admission == nil || admission.manager != m || admission.token != token || admission.recipientID != recipientID ||
		admission.recipientActivationID != recipientActivationID {
		return failBeforeWrite("message admission reservation was unavailable", errors.New("message admission was not reserved"))
	}
	defer admission.Release()
	submissionCtx, cancelSubmission := context.WithCancel(ctx)
	stopAdmissionCancel := context.AfterFunc(admission.ctx, cancelSubmission)
	defer func() {
		stopAdmissionCancel()
		cancelSubmission()
	}()
	caller := CallerIdentity{SessionID: admission.senderID}

	record, err := m.messages.GetMessage(messageID)
	if err != nil {
		return failBeforeWrite("message intent was unavailable before prompt submission", err)
	}
	if record.Delivery != MessageAccepted {
		return record, nil
	}
	if err = m.messages.WaitTurn(submissionCtx, messageID); err != nil {
		_ = m.messages.UpdateDelivery(messageID, MessageNotSubmitted, "submission was canceled before writing prompt bytes")
		updated, getErr := m.messages.GetMessage(messageID)
		return updated, errors.Join(err, getErr)
	}
	if err = m.messages.UpdateDelivery(messageID, MessageSubmitting, ""); err != nil {
		_ = m.messages.UpdateDelivery(messageID, MessageNotSubmitted, "submission state could not be saved before writing prompt bytes")
		record.Delivery = MessageNotSubmitted
		return record, err
	}

	d, valid := m.beginMessageSubmission(admission, token, recipientID, recipientActivationID)
	if !valid {
		persistErr := m.messages.UpdateDelivery(
			messageID,
			MessageNotSubmitted,
			"session lifecycle canceled the queued prompt before submission",
		)
		updated, getErr := m.messages.GetMessage(messageID)
		return updated, errors.Join(context.Canceled, persistErr, getErr)
	}
	evidence, submitErr := d.SubmitPrompt(submissionCtx, messageID, caller.SessionID, text)
	state := MessageDelivery(evidence.State)
	warning := ""
	if state == MessageUncertain {
		warning = "delivery is uncertain; partial input may remain in the CLI input buffer"
	}
	if submitErr != nil && evidence.BytesWritten > 0 && state == MessageNotSubmitted {
		state = MessageUncertain
		warning = "delivery is uncertain; partial input may remain in the CLI input buffer"
	}
	if persistErr := m.messages.UpdateDelivery(messageID, state, warning); persistErr != nil {
		record.Delivery = MessageUncertain
		record.Warning = "prompt bytes were written but the delivery outcome could not be saved"
		return record, errors.Join(submitErr, fmt.Errorf("persist delivery outcome: %w", persistErr))
	}
	updated, getErr := m.messages.GetMessage(messageID)
	return updated, errors.Join(submitErr, getErr)
}

// ReportMessage accepts a report only from its authenticated recipient activation.
func (m *Manager) ReportMessage(token, requestID, messageID, kind, text string) (MessageReply, error) {
	if m.messages == nil {
		return MessageReply{}, messagingError("history_unavailable", "messaging history is unavailable", nil)
	}
	m.mu.RLock()
	if m.closed {
		m.mu.RUnlock()
		return MessageReply{}, ErrManagerClosed
	}
	authority := m.authorities[token]
	if authority == nil || authority.revoked || !authority.published {
		m.mu.RUnlock()
		return MessageReply{}, messagingError("forbidden", "activation credential is stale or unpublished", nil)
	}
	current := m.sessions[authority.sessionID]
	if current == nil || current.ActivationID != authority.activationID || current.Status == StatusTerminated {
		m.mu.RUnlock()
		return MessageReply{}, messagingError("forbidden", "activation is no longer live", nil)
	}
	sessionID, activationID := authority.sessionID, authority.activationID
	m.mu.RUnlock()
	return m.messages.RecordReply(sessionID, activationID, requestID, messageID, kind, text)
}

// MessageRecipientLive reports whether a retained message still targets the
// same live recipient activation.
func (m *Manager) MessageRecipientLive(messageID string) (bool, error) {
	if m.messages == nil {
		return false, messagingError("history_unavailable", "messaging history is unavailable", nil)
	}
	recipientID, activationID, err := m.messages.MessageRecipient(messageID)
	if err != nil {
		return false, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	recipient := m.sessions[recipientID]
	if recipient == nil {
		return false, messagingError("history_unavailable", "message recipient session was deleted", nil)
	}
	return !m.closed && recipient.ActivationID == activationID &&
		recipient.Status != StatusTerminated && !recipient.stopping && !recipient.deleting && !recipient.activating, nil
}
