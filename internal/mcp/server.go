// Package mcp provides Kankuro's authenticated Streamable HTTP MCP endpoint.
package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/BlackGiack/kankuro/internal/api"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	maxRequestBytes      int64 = 1 << 20
	mcpIdleTimeout             = 60 * time.Second
	mcpReadHeaderTimeout       = 5 * time.Second
	mcpReadTimeout             = 15 * time.Second
)

type registrar interface {
	RegisterConversation(token, conversationID, name string) (api.RegistrationResult, error)
	HasActivationCredential(token string) bool
}

type discoveryService interface {
	ResolveCaller(string) (api.CallerContext, error)
	ListCards(string) ([]api.CardDiscovery, error)
	ListSessions(string) ([]api.SessionDiscovery, error)
	GetSession(string, string) (api.SessionInspection, error)
	CreateSession(context.Context, string, string, string, ...*string) (api.OperationResult, error)
	StopSession(context.Context, string, string, string) (api.OperationResult, error)
	RelaunchSession(context.Context, string, string, string, *api.ConversationRequest) (api.OperationResult, error)
	SwitchSessionCard(context.Context, string, string, string, string, *api.ConversationRequest) (api.OperationResult, error)
	DeleteSession(context.Context, string, string, string) (api.OperationResult, error)
	GetOperation(string, string) (api.OperationResult, error)
	CancelOperation(string, string) (api.OperationResult, error)
}

type messagingService interface {
	SendMessage(context.Context, string, string, string, string, string) (api.OperationResult, error)
	GetMessage(string, string) (api.MessageRecord, error)
	ReportReply(string, string, string, string, string) (api.MessageReply, error)
	GetReplies(context.Context, string, string, string, int, time.Duration) (api.ReplyPage, error)
}

type contextKey struct{}

type registrationInput struct {
	ConversationID string `json:"conversation_id" jsonschema:"the actual native conversation ID reported by the CLI"`
	Name           string `json:"name"            jsonschema:"a short meaningful display name for this conversation"`
}

type sessionInput struct {
	SessionID string `json:"session_id" jsonschema:"session ID in the linked dashboard"`
}

type createSessionInput struct {
	RequestID string  `json:"request_id"     jsonschema:"caller-generated UUID, stable across retries"`
	CardID    string  `json:"card_id"`
	Name      *string `json:"name,omitempty"`
}

type targetSessionInput struct {
	RequestID string `json:"request_id" jsonschema:"UUID, stable across retries"`
	SessionID string `json:"session_id" jsonschema:"target session ID"`
}

type activationInput struct {
	RequestID    string                   `json:"request_id"`
	SessionID    string                   `json:"session_id"`
	Conversation *api.ConversationRequest `json:"conversation,omitempty" jsonschema:"mode new or resume; required when bindings exist"`
}

type switchSessionInput struct {
	RequestID    string                   `json:"request_id"`
	SessionID    string                   `json:"session_id"`
	CardID       string                   `json:"card_id"`
	Conversation *api.ConversationRequest `json:"conversation,omitempty" jsonschema:"mode new or resume; required when bindings exist"`
}

type operationInput struct {
	OperationID string `json:"operation_id" jsonschema:"operation ID returned by a lifecycle mutation"`
}

type sendMessageInput struct {
	RequestID    string `json:"request_id"    jsonschema:"caller-generated UUID, stable across retries"`
	SessionID    string `json:"session_id"    jsonschema:"recipient session ID"`
	ActivationID string `json:"activation_id" jsonschema:"recipient activation ID from get_session"`
	Text         string `json:"text"          jsonschema:"short single-line instruction, up to 2 KiB"`
}
type messageInput struct {
	MessageID string `json:"message_id"`
}
type reportReplyInput struct {
	RequestID string `json:"request_id" jsonschema:"caller-generated UUID, stable across retries"`
	MessageID string `json:"message_id"`
	Kind      string `json:"kind"       jsonschema:"completed, question, blocked, or progress"`
	Text      string `json:"text"       jsonschema:"explicit report text, up to 16 KiB"`
}
type repliesInput struct {
	MessageID   string `json:"message_id,omitempty"`
	Cursor      string `json:"cursor,omitempty"`
	Limit       int    `json:"limit,omitempty"`
	WaitSeconds int    `json:"wait_seconds,omitempty" jsonschema:"zero by default, at most ten"`
}

type requestRejection struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type lifecycleToolResponse struct {
	Operation *api.OperationResult `json:"operation,omitempty"`
	Error     *requestRejection    `json:"error,omitempty"`
}

func lifecycleResponse(result api.OperationResult, err error) (*sdk.CallToolResult, lifecycleToolResponse, error) {
	if err == nil {
		return nil, lifecycleToolResponse{Operation: &result}, nil
	}
	rejected := api.SafeRejection(err)
	payload := requestRejection{Code: rejected.Code, Message: rejected.Message}
	encoded, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return nil, lifecycleToolResponse{}, marshalErr
	}
	return &sdk.CallToolResult{
		IsError: true,
		Content: []sdk.Content{&sdk.TextContent{Text: string(encoded)}},
	}, lifecycleToolResponse{Error: &payload}, nil
}

type discoveryResult[T any] struct {
	Caller api.CallerContext `json:"caller"`
	Items  T                 `json:"items"`
}

type Runtime struct {
	endpoint string
	server   *http.Server
	listener net.Listener
	done     chan error
	failures chan error
}

// Start binds an OS-assigned loopback port and serves the registration tool.
func Start(registrar registrar) (*Runtime, error) {
	var listenConfig net.ListenConfig
	listener, err := listenConfig.Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("listen for Kankuro MCP: %w", err)
	}
	endpoint := "http://" + listener.Addr().String() + "/mcp"
	implementation := &sdk.Implementation{Name: "kankuro", Version: "1.0.0"}
	mcpServer := sdk.NewServer(implementation, nil)
	sdk.AddTool(mcpServer, &sdk.Tool{
		Name:        "register_conversation",
		Description: "Register the actual native conversation ID and a short display name for the authenticated Kankuro activation. If activation is still starting, retry the same report after one second, up to five times. A pending response is not confirmation or proof of persistence.",
	}, func(ctx context.Context, _ *sdk.CallToolRequest, input registrationInput) (*sdk.CallToolResult, api.RegistrationResult, error) {
		token, _ := ctx.Value(contextKey{}).(string)
		result, registerErr := registrar.RegisterConversation(token, input.ConversationID, input.Name)
		return nil, result, registerErr
	})
	if discovery, ok := registrar.(discoveryService); ok {
		addDiscoveryTools(mcpServer, discovery)
	}
	if messaging, ok := registrar.(messagingService); ok {
		addMessagingTools(mcpServer, messaging)
	}
	stream := sdk.NewStreamableHTTPHandler(
		func(*http.Request) *sdk.Server { return mcpServer },
		&sdk.StreamableHTTPOptions{Stateless: true},
	)
	handler := authorize(registrar, endpointHost(endpoint), stream)
	// Streamable HTTP may hold an SSE response open; bound request reads and idle
	// connections, but do not impose a write deadline on the streaming response.
	server := &http.Server{
		Handler:           handler,
		IdleTimeout:       mcpIdleTimeout,
		ReadHeaderTimeout: mcpReadHeaderTimeout,
		ReadTimeout:       mcpReadTimeout,
	}
	runtime := &Runtime{endpoint: endpoint, server: server, listener: listener, done: make(chan error, 1), failures: make(chan error, 1)}
	go func() {
		serveErr := server.Serve(listener)
		runtime.done <- serveErr
		if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			runtime.failures <- serveErr
		}
	}()
	return runtime, nil
}

func addDiscoveryTools(server *sdk.Server, service discoveryService) {
	sdk.AddTool(server, &sdk.Tool{
		Name:        "list_cards",
		Description: "List cards and resolved workspace paths in the linked dashboard. Requires the caller card's Dashboard Control grant and selected Kankuro MCP.",
	},
		func(ctx context.Context, _ *sdk.CallToolRequest, _ struct{}) (*sdk.CallToolResult, discoveryResult[[]api.CardDiscovery], error) {
			token := callerToken(ctx)
			caller, err := service.ResolveCaller(token)
			if err != nil {
				return nil, discoveryResult[[]api.CardDiscovery]{}, err
			}
			items, err := service.ListCards(token)
			return nil, discoveryResult[[]api.CardDiscovery]{Caller: caller, Items: items}, err
		})
	sdk.AddTool(server, &sdk.Tool{
		Name: "list_sessions",
		Description: "List other sessions in the linked dashboard. Waiting sessions are running; " +
			"stop before relaunching or switching cards.",
	},
		func(
			ctx context.Context,
			_ *sdk.CallToolRequest,
			_ struct{},
		) (*sdk.CallToolResult, discoveryResult[[]api.SessionDiscovery], error) {
			token := callerToken(ctx)
			caller, err := service.ResolveCaller(token)
			if err != nil {
				return nil, discoveryResult[[]api.SessionDiscovery]{}, err
			}
			items, err := service.ListSessions(token)
			return nil, discoveryResult[[]api.SessionDiscovery]{Caller: caller, Items: items}, err
		})
	sdk.AddTool(server, &sdk.Tool{
		Name:        "get_session",
		Description: "Inspect another session in the linked dashboard, including its actual workspace paths and saved conversation bindings. Does not expose secrets or terminal output.",
	},
		func(
			ctx context.Context,
			_ *sdk.CallToolRequest,
			input sessionInput,
		) (*sdk.CallToolResult, discoveryResult[api.SessionInspection], error) {
			token := callerToken(ctx)
			caller, err := service.ResolveCaller(token)
			if err != nil {
				return nil, discoveryResult[api.SessionInspection]{}, err
			}
			item, err := service.GetSession(token, input.SessionID)
			return nil, discoveryResult[api.SessionInspection]{Caller: caller, Items: item}, err
		})
	addLifecycleTools(server, service)
}

func addLifecycleTools(server *sdk.Server, service discoveryService) {
	sdk.AddTool(server, &sdk.Tool{
		Name: "create_session",
		Description: "Create a session from an existing card. Returns after completion or ten seconds " +
			"with an operation ID to poll. The optional name is part of retry identity. The result is a historical snapshot; get_session returns current delivery availability.",
	}, func(
		ctx context.Context, _ *sdk.CallToolRequest, input createSessionInput,
	) (*sdk.CallToolResult, lifecycleToolResponse, error) {
		result, err := service.CreateSession(ctx, callerToken(ctx), input.RequestID, input.CardID, input.Name)
		return lifecycleResponse(result, err)
	})
	sdk.AddTool(server, &sdk.Tool{
		Name: "stop_session",
		Description: "Stop a running or waiting session. Waiting still means running. Stop must complete " +
			"before relaunch or card switch. The same request UUID is safe to retry.",
	}, func(
		ctx context.Context, _ *sdk.CallToolRequest, input targetSessionInput,
	) (*sdk.CallToolResult, lifecycleToolResponse, error) {
		result, err := service.StopSession(ctx, callerToken(ctx), input.RequestID, input.SessionID)
		return lifecycleResponse(result, err)
	})
	sdk.AddTool(server, &sdk.Tool{
		Name: "relaunch_session",
		Description: "Relaunch a terminated session in its saved workspace. If saved conversations exist, " +
			"specify conversation mode new or resume. Poll its operation ID after ten seconds if pending.",
	}, func(
		ctx context.Context, _ *sdk.CallToolRequest, input activationInput,
	) (*sdk.CallToolResult, lifecycleToolResponse, error) {
		result, err := service.RelaunchSession(ctx, callerToken(ctx), input.RequestID, input.SessionID, input.Conversation)
		return lifecycleResponse(result, err)
	})
	sdk.AddTool(server, &sdk.Tool{
		Name: "switch_session_card",
		Description: "Switch a terminated session to another card while preserving its workspace. " +
			"Waiting sessions are running: stop and await completion first. Conversation choice may be required.",
	}, func(
		ctx context.Context, _ *sdk.CallToolRequest, input switchSessionInput,
	) (*sdk.CallToolResult, lifecycleToolResponse, error) {
		result, err := service.SwitchSessionCard(ctx, callerToken(ctx), input.RequestID, input.SessionID, input.CardID, input.Conversation)
		return lifecycleResponse(result, err)
	})
	sdk.AddTool(server, &sdk.Tool{
		Name: "delete_session",
		Description: "Delete an idle terminated session. Running, waiting, activating, and stopping " +
			"sessions cannot be deleted. Cleanup warnings describe files or records that may remain.",
	}, func(
		ctx context.Context, _ *sdk.CallToolRequest, input targetSessionInput,
	) (*sdk.CallToolResult, lifecycleToolResponse, error) {
		result, err := service.DeleteSession(ctx, callerToken(ctx), input.RequestID, input.SessionID)
		return lifecycleResponse(result, err)
	})
	sdk.AddTool(server, &sdk.Tool{
		Name: "get_operation",
		Description: "Read the outcome of an operation from this Kankuro session. This is an immediate " +
			"snapshot; poll as needed.",
	}, func(
		ctx context.Context, _ *sdk.CallToolRequest, input operationInput,
	) (*sdk.CallToolResult, lifecycleToolResponse, error) {
		result, err := service.GetOperation(callerToken(ctx), input.OperationID)
		return lifecycleResponse(result, err)
	})
	sdk.AddTool(server, &sdk.Tool{
		Name: "cancel_operation",
		Description: "Request cancellation of this session's pending or preparing operation. Cancellation " +
			"may be too late after publication or irreversible cleanup; inspect state and effects.",
	}, func(
		ctx context.Context, _ *sdk.CallToolRequest, input operationInput,
	) (*sdk.CallToolResult, lifecycleToolResponse, error) {
		result, err := service.CancelOperation(callerToken(ctx), input.OperationID)
		return lifecycleResponse(result, err)
	})
}

func addMessagingTools(server *sdk.Server, service messagingService) {
	sdk.AddTool(
		server,
		&sdk.Tool{
			Name:        "send_message",
			Description: "Send a short single-line prompt to a pinned live session activation. Inspect the current activation and delivery availability first. A starting rejection can be retried with the same request UUID. Submission means the prompt and Enter were written, not that the CLI processed it.",
		},
		func(ctx context.Context, _ *sdk.CallToolRequest, input sendMessageInput) (*sdk.CallToolResult, lifecycleToolResponse, error) {
			result, err := service.SendMessage(ctx, callerToken(ctx), input.RequestID, input.SessionID, input.ActivationID, input.Text)
			return lifecycleResponse(result, err)
		},
	)
	sdk.AddTool(
		server,
		&sdk.Tool{Name: "get_message", Description: "Read retained content and process-submission evidence for a message."},
		func(ctx context.Context, _ *sdk.CallToolRequest, input messageInput) (*sdk.CallToolResult, any, error) {
			result, err := service.GetMessage(callerToken(ctx), input.MessageID)
			if err != nil {
				return messageError(err)
			}
			return nil, map[string]any{"message": result}, nil
		},
	)
	sdk.AddTool(
		server,
		&sdk.Tool{
			Name:        "report_reply",
			Description: "Persist an explicit completed, question, blocked, or progress report for a message assigned to this activation. Progress is informational and does not complete a workflow. Use a stable request UUID for retries.",
		},
		func(ctx context.Context, _ *sdk.CallToolRequest, input reportReplyInput) (*sdk.CallToolResult, any, error) {
			result, err := service.ReportReply(callerToken(ctx), input.RequestID, input.MessageID, input.Kind, input.Text)
			if err != nil {
				return messageError(err)
			}
			return nil, map[string]any{"reply": result}, nil
		},
	)
	sdk.AddTool(
		server,
		&sdk.Tool{
			Name:        "get_replies",
			Description: "Retrieve explicit correlated reports with a bounded cursor page. Empty results and timeouts do not imply completion; cursor gaps may repeat retained reply IDs.",
		},
		func(ctx context.Context, _ *sdk.CallToolRequest, input repliesInput) (*sdk.CallToolResult, any, error) {
			result, err := service.GetReplies(
				ctx,
				callerToken(ctx),
				input.MessageID,
				input.Cursor,
				input.Limit,
				time.Duration(input.WaitSeconds)*time.Second,
			)
			if err != nil {
				return messageError(err)
			}
			return nil, result, nil
		},
	)
}

func messageError(err error) (*sdk.CallToolResult, any, error) {
	rejected := api.SafeRejection(err)
	payload := requestRejection{Code: rejected.Code, Message: rejected.Message}
	encoded, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return nil, nil, marshalErr
	}
	return &sdk.CallToolResult{
		IsError: true,
		Content: []sdk.Content{&sdk.TextContent{Text: string(encoded)}},
	}, map[string]any{
		"error": payload,
	}, nil
}

func callerToken(ctx context.Context) string {
	token, _ := ctx.Value(contextKey{}).(string)
	return token
}

func (r *Runtime) Endpoint() string { return r.endpoint }

// Failures reports an unexpected server termination, if one occurs.
func (r *Runtime) Failures() <-chan error { return r.failures }

// Shutdown drains active MCP requests and joins the serving goroutine.
func (r *Runtime) Shutdown(ctx context.Context) error {
	err := r.server.Shutdown(ctx)
	if err != nil {
		_ = r.server.Close()
	}
	serveErr := <-r.done
	if errors.Is(serveErr, http.ErrServerClosed) {
		serveErr = nil
	}
	return errors.Join(err, serveErr)
}

func endpointHost(endpoint string) string {
	parsed, _ := url.Parse(endpoint)
	return parsed.Host
}

func authorize(registrar registrar, endpointHost string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host != endpointHost || !strings.HasPrefix(endpointHost, "127.0.0.1:") {
			http.Error(w, "invalid host", http.StatusForbidden)
			return
		}
		if origin := r.Header.Get("Origin"); origin != "" {
			parsed, err := url.Parse(origin)
			if err != nil || parsed.Host != endpointHost || (parsed.Scheme != "http" && parsed.Scheme != "https") {
				http.Error(w, "invalid origin", http.StatusForbidden)
				return
			}
		}
		const prefix = "Bearer "
		authorization := r.Header.Get("Authorization")
		if !strings.HasPrefix(authorization, prefix) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		token := strings.TrimPrefix(authorization, prefix)
		if token == "" || !registrar.HasActivationCredential(token) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, maxRequestBytes)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), contextKey{}, token)))
	})
}
