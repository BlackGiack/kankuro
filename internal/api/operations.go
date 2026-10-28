package api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/BlackGiack/kankuro/internal/fsutil"
	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/google/uuid"
)

const (
	createSessionOperation   = "create_session"
	codeCancelled            = "cancelled"
	codeConversationInvalid  = "conversation_invalid"
	codeConversationRequired = "conversation_choice_required"
	codeLifecycleConflict    = "lifecycle_conflict"
	codeNotFound             = "not_found"
	codeShutdown             = "shutdown"
	codeWorkspaceUnavailable = "workspace_unavailable"
	operationWorkers         = 4
	operationQueueCapacity   = 32
	operationWait            = 10 * time.Second
	operationTimeout         = 5 * time.Minute
	operationPollInterval    = 20 * time.Millisecond
	operationFileMode        = 0o600
	operationCallerPoll      = 100 * time.Millisecond
)

type OperationState string

const (
	OperationPending              OperationState = "pending"
	OperationRunning              OperationState = "running"
	OperationSucceeded            OperationState = "succeeded"
	OperationSucceededWithWarning OperationState = "succeeded_with_warnings"
	OperationFailed               OperationState = "failed"
	OperationCancelled            OperationState = "cancelled"
	OperationInterrupted          OperationState = "interrupted"
)

// Operation is a durable, credential-free summary of one accepted request.
type Operation struct {
	ID                    string          `json:"operation_id"`
	CallerSession         string          `json:"caller_session_id"`
	DashboardID           string          `json:"dashboard_id"`
	RequestID             string          `json:"request_id"`
	Kind                  string          `json:"kind"`
	State                 OperationState  `json:"state"`
	CancellationRequested bool            `json:"cancellation_requested,omitempty"`
	TargetSession         string          `json:"target_session_id,omitempty"`
	Fingerprint           string          `json:"fingerprint"`
	Arguments             json.RawMessage `json:"arguments,omitempty"`
	Result                json.RawMessage `json:"result,omitempty"`
	ErrorCode             string          `json:"error_code,omitempty"`
	ErrorMessage          string          `json:"error_message,omitempty"`
	Warning               string          `json:"warning,omitempty"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
}

// OperationResult is the safe transport view; journal identity and fingerprints stay private.
type OperationResult struct {
	ID                    string            `json:"operation_id"`
	RequestID             string            `json:"request_id"`
	Kind                  string            `json:"kind"`
	State                 OperationState    `json:"state"`
	CancellationRequested bool              `json:"cancellation_requested,omitempty"`
	TargetSession         string            `json:"target_session_id,omitempty"`
	Result                *OperationEffects `json:"result,omitempty"`
	ErrorCode             string            `json:"error_code,omitempty"`
	ErrorMessage          string            `json:"error_message,omitempty"`
	Warning               string            `json:"warning,omitempty"`
	CreatedAt             time.Time         `json:"created_at"`
	UpdatedAt             time.Time         `json:"updated_at"`
}

// OperationEffects is the structured lifecycle outcome returned over MCP.
// The journal stores its JSON encoding as RawMessage independently.
type OperationEffects struct {
	MessageID                 string             `json:"message_id,omitempty"`
	Delivery                  string             `json:"delivery,omitempty"`
	SessionID                 string             `json:"session_id,omitempty"`
	ProcessTerminated         bool               `json:"process_terminated,omitempty"`
	SessionRecordPersisted    bool               `json:"session_record_persisted,omitempty"`
	InMemoryRemoved           bool               `json:"in_memory_removed,omitempty"`
	SessionRecordDeleted      bool               `json:"session_record_deleted,omitempty"`
	TemporaryWorkspaceRemoved bool               `json:"temporary_workspace_removed,omitempty"`
	Inspection                *SessionInspection `json:"inspection,omitempty"`
}

func operationResult(operation Operation) OperationResult {
	var result *OperationEffects
	if len(operation.Result) > 0 {
		result = new(OperationEffects)
		if err := json.Unmarshal(operation.Result, result); err != nil {
			result = nil
		}
	}
	return OperationResult{
		ID: operation.ID, RequestID: operation.RequestID, Kind: operation.Kind,
		State: operation.State, CancellationRequested: operation.CancellationRequested,
		TargetSession: operation.TargetSession, Result: result,
		ErrorCode: operation.ErrorCode, ErrorMessage: operation.ErrorMessage, Warning: operation.Warning,
		CreatedAt: operation.CreatedAt, UpdatedAt: operation.UpdatedAt,
	}
}

type operationKey struct{ caller, request string }
type operationTask struct {
	operationID string
	token       string
	ctx         context.Context
	cancel      context.CancelFunc
	apply       func(context.Context, Operation) (json.RawMessage, error)
	onCancel    func(Operation)
}

// OperationRunner serializes durable request acceptance and bounds lifecycle work.
type OperationRunner struct {
	mu              sync.Mutex
	dir             string
	service         *Service
	writeFileAtomic func(string, []byte) error
	records         map[operationKey]*Operation
	byID            map[string]operationKey
	running         map[string]context.CancelFunc
	queue           chan operationTask
	closed          bool
	loadErr         error
	workers         sync.WaitGroup
}

type ConversationRequest struct {
	Mode           string `json:"mode"`
	ConversationID string `json:"conversation_id,omitempty"`
}

type createArguments struct {
	CardID string `json:"card_id"`
}
type conversationArguments struct {
	Conversation *ConversationRequest `json:"conversation,omitempty"`
}
type switchArguments struct {
	CardID       string               `json:"card_id"`
	Conversation *ConversationRequest `json:"conversation,omitempty"`
}
type lifecycleEffects = OperationEffects

type operationWarningError struct{ message string }

func (e operationWarningError) Error() string { return e.message }

func (s *Service) CreateSession(ctx context.Context, token, requestID, cardID string) (OperationResult, error) {
	if s.operations == nil {
		return OperationResult{}, errors.New("operation runner unavailable")
	}
	caller, err := s.ResolveCaller(token)
	if err != nil {
		return OperationResult{}, err
	}
	if strings.TrimSpace(cardID) == "" {
		return OperationResult{}, reject("invalid_request", "card_id is required")
	}
	args := createArguments{CardID: cardID}
	if replay, found, replayErr := s.replayMutation(ctx, token, requestID, createSessionOperation, "", args); replayErr != nil {
		return OperationResult{}, replayErr
	} else if found {
		return operationResult(replay), nil
	}
	if !s.scope.AllowsCard(caller.SessionID, cardID) {
		return OperationResult{}, reject("forbidden", "card is outside the caller's dashboard scope")
	}
	if _, loadErr := s.loadCard(cardID); loadErr != nil {
		return OperationResult{}, classifyCardLoadError(loadErr)
	}
	apply := func(operationCtx context.Context, operation Operation) (json.RawMessage, error) {
		sid, launchErr := s.launchReservedSession(operationCtx, cardID, operation.TargetSession)
		if launchErr != nil && !errors.Is(launchErr, session.ErrPersistFailed) {
			return nil, launchErr
		}
		if assigner, ok := s.scope.(dashboardSessionAssigner); ok {
			assigner.AssignSession(caller.SessionID, sid, cardID)
		}
		inspection, inspectErr := s.inspectSession(sid)
		if inspectErr != nil {
			return nil, errors.Join(launchErr, inspectErr)
		}
		encoded, marshalErr := json.Marshal(lifecycleEffects{SessionID: sid, Inspection: &inspection})
		if marshalErr != nil {
			return nil, marshalErr
		}
		if errors.Is(launchErr, session.ErrPersistFailed) {
			return encoded, fmt.Errorf("activation started but session record could not be saved: %w", launchErr)
		}
		return encoded, nil
	}
	operation, err := s.operations.Submit(ctx, token, requestID, createSessionOperation, "", args, apply)
	return operationResult(operation), err
}

func (s *Service) launchReservedSession(ctx context.Context, cardID, sessionID string) (string, error) {
	release, err := s.reserveCardReference(cardID)
	if err != nil {
		return "", err
	}
	defer release()

	bundle, err := s.resolveCard(cardID)
	if err != nil {
		return "", err
	}
	return s.sessions.LaunchWithIDContext(
		ctx, sessionID, bundle.Card, bundle.Skills, bundle.Personas, bundle.MCPs,
	)
}

func (s *Service) StopSession(ctx context.Context, token, requestID, sessionID string) (OperationResult, error) {
	apply := func(context.Context) (lifecycleEffects, error) {
		result, err := s.sessions.KillDetailed(sessionID)
		effects := lifecycleEffects{
			SessionID: sessionID, ProcessTerminated: result.ProcessTerminated,
			SessionRecordPersisted: result.SessionRecordPersisted,
		}
		if err != nil && result.ProcessTerminated {
			return effects, operationWarningError{"process stopped, but teardown or metadata persistence reported a warning"}
		}
		return effects, err
	}
	operation, err := s.submitSession(ctx, token, requestID, "stop_session", sessionID, struct{}{}, apply)
	return operationResult(operation), err
}

func (s *Service) RelaunchSession(
	ctx context.Context,
	token, requestID, sessionID string,
	conversation *ConversationRequest,
) (OperationResult, error) {
	if _, err := conversationChoice(conversation); err != nil {
		return OperationResult{}, reject("invalid_request", "conversation choice is invalid")
	}
	args := conversationArguments{Conversation: conversation}
	apply := func(operationCtx context.Context) (lifecycleEffects, error) {
		choice, parseErr := conversationChoice(conversation)
		if parseErr != nil {
			return lifecycleEffects{}, parseErr
		}
		snapshot, found := s.sessions.GetSession(sessionID)
		if !found {
			return lifecycleEffects{}, errors.New("session not found")
		}
		card, loadErr := s.loadCard(snapshot.CardID)
		if loadErr != nil {
			return lifecycleEffects{}, loadErr
		}
		activationErr := s.sessions.RelaunchWithChoiceContext(operationCtx, sessionID, card, choice)
		if activationErr != nil && !errors.Is(activationErr, session.ErrPersistFailed) {
			return lifecycleEffects{}, activationErr
		}
		inspection, inspectErr := s.inspectSession(sessionID)
		if inspectErr != nil {
			return lifecycleEffects{}, errors.Join(activationErr, inspectErr)
		}
		out := lifecycleEffects{SessionID: sessionID, Inspection: &inspection}
		if errors.Is(activationErr, session.ErrPersistFailed) {
			return out, fmt.Errorf("activation started but session record could not be saved: %w", activationErr)
		}
		return out, nil
	}
	operation, err := s.submitSession(ctx, token, requestID, "relaunch_session", sessionID, args, apply)
	return operationResult(operation), err
}

func (s *Service) SwitchSessionCard(
	ctx context.Context,
	token, requestID, sessionID, cardID string,
	conversation *ConversationRequest,
) (OperationResult, error) {
	caller, err := s.ResolveCaller(token)
	if err != nil {
		return OperationResult{}, err
	}
	if strings.TrimSpace(cardID) == "" {
		return OperationResult{}, reject("invalid_request", "card_id is required")
	}
	if _, choiceErr := conversationChoice(conversation); choiceErr != nil {
		return OperationResult{}, reject("invalid_request", "conversation choice is invalid")
	}
	args := switchArguments{CardID: cardID, Conversation: conversation}
	if replay, found, replayErr := s.replayMutation(ctx, token, requestID, "switch_session_card", sessionID, args); replayErr != nil {
		return OperationResult{}, replayErr
	} else if found {
		return operationResult(replay), nil
	}
	if !s.scope.AllowsCard(caller.SessionID, cardID) {
		return OperationResult{}, reject("forbidden", "destination card is outside the caller's dashboard scope")
	}
	if _, loadErr := s.loadCard(cardID); loadErr != nil {
		return OperationResult{}, classifyCardLoadError(loadErr)
	}
	apply := func(operationCtx context.Context) (lifecycleEffects, error) {
		release, reserveErr := s.reserveCardReference(cardID)
		if reserveErr != nil {
			return lifecycleEffects{}, reserveErr
		}
		defer release()
		choice, parseErr := conversationChoice(conversation)
		if parseErr != nil {
			return lifecycleEffects{}, parseErr
		}
		bundle, resolveErr := s.resolveCard(cardID)
		if resolveErr != nil {
			return lifecycleEffects{}, resolveErr
		}
		activationErr := s.sessions.SwitchAgentWithChoiceContext(
			operationCtx, sessionID, bundle.Card, bundle.Skills, bundle.Personas, bundle.MCPs, choice,
		)
		if activationErr != nil && !errors.Is(activationErr, session.ErrPersistFailed) {
			return lifecycleEffects{}, activationErr
		}
		inspection, inspectErr := s.inspectSession(sessionID)
		if inspectErr != nil {
			return lifecycleEffects{}, errors.Join(activationErr, inspectErr)
		}
		out := lifecycleEffects{SessionID: sessionID, Inspection: &inspection}
		if errors.Is(activationErr, session.ErrPersistFailed) {
			return out, fmt.Errorf("activation started but session record could not be saved: %w", activationErr)
		}
		return out, nil
	}
	operation, err := s.submitSession(ctx, token, requestID, "switch_session_card", sessionID, args, apply)
	return operationResult(operation), err
}

func (s *Service) DeleteSession(ctx context.Context, token, requestID, sessionID string) (OperationResult, error) {
	apply := func(context.Context) (lifecycleEffects, error) {
		result, err := s.sessions.RemoveDetailed(sessionID)
		if errors.Is(err, session.ErrTrustCleanup) {
			return lifecycleEffects{
				SessionID: sessionID,
			}, operationWarningError{
				"temporary workspace trust cleanup may be incomplete; session retained: fix the CLI configuration and retry deletion",
			}
		}
		if err != nil && result.InMemoryRemoved {
			return lifecycleEffects{SessionID: sessionID, InMemoryRemoved: result.InMemoryRemoved,
					SessionRecordDeleted:      result.SessionRecordDeleted,
					TemporaryWorkspaceRemoved: result.TemporaryWorkspaceRemoved},
				operationWarningError{"session removed in memory, but its record or temporary workspace may remain"}
		}
		return lifecycleEffects{SessionID: sessionID, InMemoryRemoved: result.InMemoryRemoved,
			SessionRecordDeleted:      result.SessionRecordDeleted,
			TemporaryWorkspaceRemoved: result.TemporaryWorkspaceRemoved}, err
	}
	operation, err := s.submitSession(ctx, token, requestID, "delete_session", sessionID, struct{}{}, apply)
	return operationResult(operation), err
}

func (s *Service) submitSession(
	ctx context.Context,
	token, requestID, kind, sessionID string,
	arguments any,
	apply func(context.Context) (lifecycleEffects, error),
) (Operation, error) {
	if s.operations == nil {
		return Operation{}, errors.New("operation runner unavailable")
	}
	caller, err := s.ResolveCaller(token)
	if err != nil {
		return Operation{}, err
	}
	if caller.SessionID == sessionID {
		return Operation{}, reject("forbidden", "cannot target the caller session")
	}
	if previous, found, replayErr := s.replayMutation(ctx, token, requestID, kind, sessionID, arguments); replayErr != nil {
		return Operation{}, replayErr
	} else if found {
		return previous, nil
	}
	if !s.scope.AllowsSession(caller.SessionID, sessionID) {
		return Operation{}, reject("forbidden", "session is outside the caller's dashboard scope")
	}
	if _, exists := s.sessions.GetSession(sessionID); !exists {
		return Operation{}, reject("not_found", "session was not found")
	}
	return s.operations.Submit(ctx, token, requestID, kind, sessionID, arguments,
		func(operationCtx context.Context, _ Operation) (json.RawMessage, error) {
			out, applyErr := apply(operationCtx)
			encoded, marshalErr := json.Marshal(out)
			return encoded, errors.Join(applyErr, marshalErr)
		})
}

func (s *Service) replayMutation(
	ctx context.Context, token, requestID, kind, target string, arguments any,
) (Operation, bool, error) {
	if s.operations == nil {
		return Operation{}, false, errors.New("operation runner unavailable")
	}
	return s.operations.Replay(ctx, token, requestID, kind, target, arguments)
}

func (s *Service) GetOperation(token, operationID string) (OperationResult, error) {
	if s.operations == nil {
		return OperationResult{}, errors.New("operation runner unavailable")
	}
	operation, err := s.operations.Get(token, operationID)
	return operationResult(operation), err
}

func (s *Service) CancelOperation(token, operationID string) (OperationResult, error) {
	if s.operations == nil {
		return OperationResult{}, errors.New("operation runner unavailable")
	}
	operation, err := s.operations.Cancel(token, operationID)
	return operationResult(operation), err
}

func conversationChoice(request *ConversationRequest) (session.ConversationChoice, error) {
	if request == nil {
		return session.ConversationChoice{}, nil
	}
	switch request.Mode {
	case "new":
		if request.ConversationID != "" {
			return session.ConversationChoice{}, fmt.Errorf("%w: new cannot include conversation_id", session.ErrConversationChoiceInvalid)
		}
		return session.ConversationChoice{New: true}, nil
	case "resume":
		if request.ConversationID == "" {
			return session.ConversationChoice{}, fmt.Errorf("%w: resume requires conversation_id", session.ErrConversationChoiceInvalid)
		}
		return session.ConversationChoice{ID: request.ConversationID}, nil
	default:
		return session.ConversationChoice{}, fmt.Errorf("%w: mode must be new or resume", session.ErrConversationChoiceInvalid)
	}
}

func (s *Service) inspectSession(sessionID string) (SessionInspection, error) {
	item, ok := s.sessions.GetSession(sessionID)
	if !ok {
		return SessionInspection{}, errors.New("session not found after lifecycle operation")
	}
	subWorkspaces := item.SubWorkspaces
	metadataAvailable := item.SubWorkspaceMetadataAvailable
	if !metadataAvailable {
		subWorkspaces, metadataAvailable = session.ReadSubWorkspaceDescriptors(item.WorkingDir)
	}
	if subWorkspaces == nil {
		subWorkspaces = []session.SubWorkspaceDescriptor{}
	}
	conversations := item.Conversations
	if conversations == nil {
		conversations = map[string][]session.ConversationBinding{}
	}
	return SessionInspection{ID: item.ID, Name: item.Name, CardID: item.CardID, Status: item.Status, Activity: item.LifecycleActivity,
		WorkingDir: item.WorkingDir, OriginalDir: item.OriginalDir, UseTmpWorkspace: item.UseTmpWorkspace,
		Conversations: conversations, SubWorkspaces: subWorkspaces, SubWorkspaceMetadataAvailable: metadataAvailable,
		ActiveConversationID: item.ActiveConversationID,
		Registration:         item.Registration, RegistrationPersistence: item.RegistrationPersistence}, nil
}

// NewOperationRunner loads the durable idempotency journal and starts workers.
func NewOperationRunner(configDir string, service *Service) *OperationRunner {
	r := &OperationRunner{
		dir: filepath.Join(configDir, "operations"), service: service,
		records: make(map[operationKey]*Operation), byID: make(map[string]operationKey),
		running: make(map[string]context.CancelFunc), queue: make(chan operationTask, operationQueueCapacity),
	}
	r.loadErr = r.load()
	if r.loadErr == nil {
		for range operationWorkers {
			r.workers.Add(1)
			go r.worker()
		}
	}
	return r
}

// LoadError reports a journal problem that disables lifecycle mutation admission.
func (r *OperationRunner) LoadError() error { return r.loadErr }

// Submit accepts or replays an idempotent mutation. HTTP cancellation only ends
// this wait; the accepted operation runs under its independent deadline.
func (r *OperationRunner) Submit(
	ctx context.Context, token, requestID, kind, target string, arguments any,
	apply func(context.Context, Operation) (json.RawMessage, error),
) (Operation, error) {
	return r.submit(ctx, token, requestID, kind, target, arguments, nil, apply, nil)
}

// SubmitWithCancel installs cleanup for an accepted operation canceled before its worker starts.
func (r *OperationRunner) SubmitWithCancel(
	ctx context.Context, token, requestID, kind, target string, arguments any,
	apply func(context.Context, Operation) (json.RawMessage, error), onCancel func(Operation),
) (Operation, error) {
	return r.submitWithRecovery(ctx, token, requestID, kind, target, arguments, nil, apply, onCancel)
}

// SubmitWithAdmission runs beforeAccept after queue capacity is reserved and
// persists the operation from inside its transaction before a worker can start.
func (r *OperationRunner) SubmitWithAdmission(
	ctx context.Context, token, requestID, kind, target string, arguments any,
	beforeAccept func(persistAcceptance func() error) error,
	apply func(context.Context, Operation) (json.RawMessage, error), onCancel func(Operation),
) (Operation, error) {
	return r.submitWithRecovery(ctx, token, requestID, kind, target, arguments, beforeAccept, apply, onCancel)
}

func (r *OperationRunner) submitWithRecovery(
	ctx context.Context, token, requestID, kind, target string, arguments any,
	beforeAccept func(func() error) error,
	apply func(context.Context, Operation) (json.RawMessage, error), onCancel func(Operation),
) (Operation, error) {
	operation, err := r.submit(ctx, token, requestID, kind, target, arguments, beforeAccept, apply, onCancel)
	if operation.ID != "" || err == nil {
		return operation, err
	}
	caller, authErr := r.service.ResolveCaller(token)
	if authErr != nil {
		return operation, err
	}
	parsed, parseErr := uuid.Parse(requestID)
	if parseErr != nil {
		return operation, err
	}
	key := operationKey{caller: caller.SessionID, request: parsed.String()}
	r.mu.Lock()
	if accepted := r.records[key]; accepted != nil {
		operation = cloneOperation(accepted)
	}
	r.mu.Unlock()
	return operation, err
}

func (r *OperationRunner) submit(
	ctx context.Context, token, requestID, kind, target string, arguments any,
	beforeAccept func(func() error) error,
	apply func(context.Context, Operation) (json.RawMessage, error), onCancel func(Operation),
) (Operation, error) {
	if _, err := uuid.Parse(requestID); err != nil {
		return Operation{}, reject("invalid_request", "request_id must be a UUID")
	}
	caller, err := r.service.ResolveCaller(token)
	if err != nil {
		return Operation{}, err
	}
	encoded, err := json.Marshal(arguments)
	if err != nil {
		return Operation{}, fmt.Errorf("encode operation arguments: %w", err)
	}
	fingerprint := operationFingerprint(kind, target, encoded)
	key := operationKey{caller: caller.SessionID, request: requestID}
	op, accepted, err := r.accept(
		caller.SessionID,
		caller.DashboardID,
		key,
		requestID,
		kind,
		target,
		encoded,
		fingerprint,
		token,
		apply,
		onCancel,
		beforeAccept,
	)
	if err != nil {
		return op, err
	}
	if !accepted {
		if authErr := r.service.authorizeExistingOperation(caller, op); authErr != nil {
			return Operation{}, authErr
		}
		if isOperationTerminal(op.State) {
			return op, nil
		}
		op, err = r.wait(ctx, op.ID)
		if err != nil {
			return Operation{}, err
		}
		return op, r.service.authorizeExistingOperation(caller, op)
	}
	return r.wait(ctx, op.ID)
}

func operationFingerprint(kind, target string, encoded []byte) string {
	if kind == createSessionOperation {
		target = ""
	}
	digest := sha256.Sum256(append([]byte(kind+"\x00"+target+"\x00"), encoded...))
	return hex.EncodeToString(digest[:])
}

// Replay checks durable idempotency before rejecting an absent target.
func (r *OperationRunner) Replay(ctx context.Context, token, requestID, kind, target string, arguments any) (Operation, bool, error) {
	if _, err := uuid.Parse(requestID); err != nil {
		return Operation{}, false, reject("invalid_request", "request_id must be a UUID")
	}
	caller, err := r.service.ResolveCaller(token)
	if err != nil {
		return Operation{}, false, err
	}
	encoded, err := json.Marshal(arguments)
	if err != nil {
		return Operation{}, false, fmt.Errorf("encode operation arguments: %w", err)
	}
	fingerprint := operationFingerprint(kind, target, encoded)
	key := operationKey{caller: caller.SessionID, request: requestID}
	r.mu.Lock()
	if r.loadErr != nil {
		r.mu.Unlock()
		return Operation{}, false, reject("journal_unavailable", "lifecycle operations are temporarily unavailable")
	}
	previous := r.records[key]
	if previous == nil {
		r.mu.Unlock()
		return Operation{}, false, nil
	}
	if previous.DashboardID != caller.DashboardID || previous.Fingerprint != fingerprint {
		r.mu.Unlock()
		return Operation{}, false, reject("request_id_conflict", "request_id was already used with different arguments")
	}
	cloned := cloneOperation(previous)
	r.mu.Unlock()
	if authErr := r.service.authorizeExistingOperation(caller, cloned); authErr != nil {
		return Operation{}, false, authErr
	}
	if !isOperationTerminal(cloned.State) {
		cloned, err = r.wait(ctx, cloned.ID)
		if err != nil {
			return Operation{}, false, err
		}
		if authErr := r.service.authorizeExistingOperation(caller, cloned); authErr != nil {
			return Operation{}, false, authErr
		}
		return cloned, true, nil
	}
	return cloned, true, nil
}

func (r *OperationRunner) accept(
	callerID, dashboardID string, key operationKey, requestID, kind, target string, encoded []byte,
	fingerprint, token string,
	apply func(context.Context, Operation) (json.RawMessage, error),
	onCancel func(Operation),
	beforeAccept func(func() error) error,
) (Operation, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return Operation{}, false, reject("shutting_down", "Kankuro is shutting down")
	}
	if r.loadErr != nil {
		return Operation{}, false, reject("journal_unavailable", "lifecycle operations are temporarily unavailable")
	}
	if previous := r.records[key]; previous != nil {
		if previous.DashboardID != dashboardID || previous.Fingerprint != fingerprint {
			return Operation{}, false, reject("request_id_conflict", "request_id was already used with different arguments")
		}
		return cloneOperation(previous), false, nil
	}
	if len(r.queue) == cap(r.queue) {
		return Operation{}, false, reject("capacity_exhausted", "lifecycle operation capacity is full; retry later")
	}
	if kind == createSessionOperation && target == "" {
		target = uuid.NewString()
	}
	now := time.Now().UTC()
	op := &Operation{ID: uuid.NewString(), CallerSession: callerID, DashboardID: dashboardID, RequestID: requestID,
		Kind: kind, State: OperationPending, TargetSession: target, Fingerprint: fingerprint,
		Arguments: encoded, CreatedAt: now, UpdatedAt: now}
	persistAcceptance := func() error { return r.saveLocked(op) }
	if beforeAccept != nil {
		if admitErr := beforeAccept(persistAcceptance); admitErr != nil {
			return Operation{}, false, admitErr
		}
	} else if saveErr := persistAcceptance(); saveErr != nil {
		return Operation{}, false, reject("journal_unavailable", "lifecycle operations are temporarily unavailable")
	}
	r.records[key], r.byID[op.ID] = op, key
	opCtx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	r.running[op.ID] = cancel
	r.queue <- operationTask{operationID: op.ID, token: token, ctx: opCtx, cancel: cancel, apply: apply, onCancel: onCancel}
	return cloneOperation(op), true, nil
}

func (r *OperationRunner) wait(ctx context.Context, operationID string) (Operation, error) {
	deadline := time.NewTimer(operationWait)
	defer deadline.Stop()
	for {
		current := r.snapshot(operationID)
		if isOperationTerminal(current.State) {
			return current, nil
		}
		select {
		case <-ctx.Done():
			return current, nil
		case <-deadline.C:
			return current, nil
		case <-time.After(operationPollInterval):
		}
	}
}

// Get returns a copied operation outcome to its originating session only.
func (r *OperationRunner) Get(token, operationID string) (Operation, error) {
	caller, err := r.service.ResolveCaller(token)
	if err != nil {
		return Operation{}, err
	}
	return r.getForCaller(caller, operationID)
}

func (r *OperationRunner) getForCaller(caller CallerContext, operationID string) (Operation, error) {
	r.mu.Lock()
	key, ok := r.byID[operationID]
	if !ok || key.caller != caller.SessionID || r.records[key].DashboardID != caller.DashboardID {
		r.mu.Unlock()
		return Operation{}, reject("not_found", "operation was not found")
	}
	operation := cloneOperation(r.records[key])
	r.mu.Unlock()
	if err := r.service.authorizeExistingOperation(caller, operation); err != nil {
		return Operation{}, err
	}
	return operation, nil
}

func (r *OperationRunner) snapshot(operationID string) Operation {
	r.mu.Lock()
	defer r.mu.Unlock()
	key, ok := r.byID[operationID]
	if !ok {
		return Operation{}
	}
	return cloneOperation(r.records[key])
}

// Cancel requests cancellation after revalidating the caller activation.
func (r *OperationRunner) Cancel(token, operationID string) (Operation, error) {
	_, err := r.Get(token, operationID)
	if err != nil {
		return Operation{}, err
	}
	r.mu.Lock()
	key := r.byID[operationID]
	op := r.records[key]
	if !isOperationTerminal(op.State) {
		op.CancellationRequested = true
		op.UpdatedAt = time.Now().UTC()
		if cancel := r.running[operationID]; cancel != nil {
			cancel()
		}
		if saveErr := r.saveLocked(op); saveErr != nil {
			op.Warning = "cancellation requested but journal update failed"
		}
	}
	cloned := cloneOperation(op)
	r.mu.Unlock()
	return cloned, nil
}

// Shutdown stops admission, cancels queued/preparing work, and joins workers.
func (r *OperationRunner) Shutdown() {
	r.mu.Lock()
	if !r.closed {
		r.closed = true
		for _, cancel := range r.running {
			cancel()
		}
		close(r.queue)
	}
	r.mu.Unlock()
	r.workers.Wait()
}

func (r *OperationRunner) worker() {
	defer r.workers.Done()
	for task := range r.queue {
		r.execute(task)
	}
}

func (r *OperationRunner) execute(task operationTask) {
	r.mu.Lock()
	key := r.byID[task.operationID]
	op := r.records[key]
	if op == nil || isOperationTerminal(op.State) {
		r.mu.Unlock()
		task.cancel()
		return
	}
	if task.ctx.Err() != nil {
		snapshot := cloneOperation(op)
		op.State = OperationCancelled
		op.ErrorCode = "cancelled"
		op.ErrorMessage = "operation cancelled before execution"
		r.finishLocked(op)
		r.mu.Unlock()
		if task.onCancel != nil {
			task.onCancel(snapshot)
		}
		task.cancel()
		return
	}
	if err := r.service.validateOperationScope(task.token, *op); err != nil {
		snapshot := cloneOperation(op)
		op.State = OperationCancelled
		op.ErrorCode = "caller_revoked"
		op.ErrorMessage = "caller activation is no longer authorized"
		r.finishLocked(op)
		r.mu.Unlock()
		if task.onCancel != nil {
			task.onCancel(snapshot)
		}
		task.cancel()
		return
	}
	op.State = OperationRunning
	op.UpdatedAt = time.Now().UTC()
	_ = r.saveLocked(op)
	snapshot := cloneOperation(op)
	r.mu.Unlock()

	monitorDone := make(chan struct{})
	go func() {
		defer close(monitorDone)
		r.monitorCaller(task)
	}()
	result, err := task.apply(task.ctx, snapshot)
	r.mu.Lock()
	r.completeLocked(op, task, result, err)
	r.mu.Unlock()
	task.cancel()
	<-monitorDone
}

func (r *OperationRunner) monitorCaller(task operationTask) {
	ticker := time.NewTicker(operationCallerPoll)
	defer ticker.Stop()
	for {
		select {
		case <-task.ctx.Done():
			return
		case <-ticker.C:
			if _, err := r.service.ResolveCaller(task.token); err != nil {
				task.cancel()
				return
			}
		}
	}
}

func (r *OperationRunner) completeLocked(op *Operation, task operationTask, result json.RawMessage, err error) {
	op.Result = append(json.RawMessage(nil), result...)
	r.setOperationOutcome(op, task, err)
	r.finishLocked(op)
}

//nolint:funlen // Lifecycle outcomes are mapped together for consistent transport semantics.
func (r *OperationRunner) setOperationOutcome(op *Operation, task operationTask, err error) {
	var warning operationWarningError
	var cleanupErr *session.TemporaryTrustCleanupError
	switch {
	case errors.As(err, &warning):
		op.State = OperationSucceededWithWarning
		op.Warning = warning.message
	case errors.As(err, &cleanupErr):
		op.State = OperationFailed
		op.ErrorCode = "trust_cleanup_failed"
		op.ErrorMessage = fmt.Sprintf(
			"startup failed and temporary workspace trust cleanup may be incomplete; fix CLI configuration and remove the exact path %q manually",
			cleanupErr.WorkingDir,
		)
	case errors.Is(err, session.ErrPersistFailed):
		op.State = OperationSucceededWithWarning
		op.Warning = "activation started but session metadata was not saved"
		op.ErrorCode = "persistence_warning"
		op.ErrorMessage = "activation started; inspect the returned session and persistence status"
	case errors.Is(err, session.ErrConversationChoiceRequired):
		op.State = OperationFailed
		op.ErrorCode = codeConversationRequired
		op.ErrorMessage = "choose a saved conversation or explicitly start a new one"
	case errors.Is(err, session.ErrConversationChoiceInvalid):
		op.State = OperationFailed
		op.ErrorCode = codeConversationInvalid
		op.ErrorMessage = "conversation choice is invalid for this card and session"
	case errors.Is(err, session.ErrSessionNotFound):
		op.State = OperationFailed
		op.ErrorCode = codeNotFound
		op.ErrorMessage = "session was not found in the linked dashboard"
	case errors.Is(err, session.ErrLifecycleConflict):
		op.State = OperationFailed
		op.ErrorCode = codeLifecycleConflict
		op.ErrorMessage = "session state does not allow this lifecycle operation"
	case errors.Is(err, session.ErrWorkspaceUnavailable):
		op.State = OperationFailed
		op.ErrorCode = codeWorkspaceUnavailable
		op.ErrorMessage = "saved session workspace is unavailable; restore it before retrying"
	case errors.Is(err, session.ErrManagerClosed):
		op.State = OperationFailed
		op.ErrorCode = codeShutdown
		op.ErrorMessage = "Kankuro is shutting down"
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		op.State = OperationCancelled
		op.ErrorCode = codeCancelled
		op.ErrorMessage = "operation cancelled or deadline exceeded"
	case task.ctx.Err() != nil && err != nil:
		op.State = OperationCancelled
		op.ErrorCode = codeCancelled
		op.ErrorMessage = "operation cancelled or deadline exceeded"
	case err != nil:
		op.State = OperationFailed
		op.ErrorCode = "lifecycle_error"
		op.ErrorMessage = "operation failed; correct its preconditions or inspect the session state"
	default:
		if task.ctx.Err() != nil && op.CancellationRequested {
			op.State = OperationSucceededWithWarning
			op.Warning = "operation completed; cancellation arrived too late"
		} else {
			op.State = OperationSucceeded
		}
	}
}

func (r *OperationRunner) finishLocked(op *Operation) {
	op.UpdatedAt = time.Now().UTC()
	delete(r.running, op.ID)
	if err := r.saveLocked(op); err != nil {
		op.Warning = "outcome observed but operation journal update failed"
		if op.State == OperationSucceeded {
			op.State = OperationSucceededWithWarning
		}
	}
}

func (r *OperationRunner) load() error {
	entries, err := os.ReadDir(r.dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		if loadErr := r.loadRecord(entry.Name()); loadErr != nil {
			return loadErr
		}
	}
	return nil
}

func (r *OperationRunner) loadRecord(name string) error {
	path := filepath.Join(r.dir, name)
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var op Operation
	if decodeErr := json.Unmarshal(data, &op); decodeErr != nil {
		return fmt.Errorf("corrupt operation record %s: %w", name, decodeErr)
	}
	key := operationKey{caller: op.CallerSession, request: op.RequestID}
	if key.caller == "" || key.request == "" || op.ID == "" {
		return fmt.Errorf("incomplete operation identity in journal record %s", name)
	}
	if r.records[key] != nil || r.byID[op.ID].request != "" {
		return fmt.Errorf("duplicate operation identity in journal: %s", name)
	}
	if !isOperationTerminal(op.State) {
		op.State = OperationInterrupted
		op.Warning = "effect may have occurred; inspect current sessions before retrying"
		op.UpdatedAt = time.Now().UTC()
		if saveErr := r.saveLocked(&op); saveErr != nil {
			return saveErr
		}
	}
	r.records[key], r.byID[op.ID] = &op, key
	return nil
}

func (r *OperationRunner) saveLocked(op *Operation) error {
	if err := os.MkdirAll(r.dir, 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(op, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(r.dir, op.CallerSession+"-"+op.RequestID+".json")
	if r.writeFileAtomic != nil {
		return r.writeFileAtomic(path, data)
	}
	return fsutil.WriteFileAtomic(path, data, operationFileMode)
}

func cloneOperation(op *Operation) Operation {
	cloned := *op
	cloned.Arguments = append(json.RawMessage(nil), op.Arguments...)
	cloned.Result = append(json.RawMessage(nil), op.Result...)
	return cloned
}

func isOperationTerminal(state OperationState) bool {
	switch state {
	case OperationSucceeded, OperationSucceededWithWarning, OperationFailed, OperationCancelled, OperationInterrupted:
		return true
	case OperationPending, OperationRunning:
		return false
	default:
		return false
	}
}
