// Package mcp provides Kankuro's authenticated Streamable HTTP MCP endpoint.
package mcp

import (
	"context"
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

type contextKey struct{}

type registrationInput struct {
	ConversationID string `json:"conversation_id" jsonschema:"the actual native conversation ID reported by the CLI"`
	Name           string `json:"name"            jsonschema:"a short meaningful display name for this conversation"`
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
