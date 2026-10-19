package mcp_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/BlackGiack/kankuro/internal/api"
	"github.com/BlackGiack/kankuro/internal/mcp"
	"github.com/BlackGiack/kankuro/internal/session"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

const testShutdownTimeout = 5 * time.Second

type testRegistrar struct {
	registered bool
}

func (r *testRegistrar) HasActivationCredential(token string) bool {
	return token == "valid-test-token"
}
func (r *testRegistrar) RegisterConversation(_ string, id, name string) (api.RegistrationResult, error) {
	r.registered = id == "native-id" && name == "Work"
	return api.RegistrationResult{Status: session.RegistrationRegistered, Persistence: session.RegistrationSaved}, nil
}

type authRoundTripper struct{ next http.RoundTripper }

func (t authRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	request.Header.Set("Authorization", "Bearer valid-test-token")
	return t.next.RoundTrip(request)
}

//nolint:govet // Each request owns short-lived independent errors in the HTTP integration test.
func TestStreamableMCPRegistrationAndAuthentication(t *testing.T) {
	registrar := &testRegistrar{}
	runtime, err := mcp.Start(registrar)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), testShutdownTimeout)
		defer cancel()
		if err := runtime.Shutdown(ctx); err != nil {
			t.Errorf("shutdown MCP: %v", err)
		}
	})

	for _, authorization := range []string{"", "Bearer invalid-test-token"} {
		request, err := http.NewRequest(http.MethodPost, runtime.Endpoint(), strings.NewReader(`{}`))
		if err != nil {
			t.Fatal(err)
		}
		request.Header.Set("Content-Type", "application/json")
		if authorization != "" {
			request.Header.Set("Authorization", authorization)
		}
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = io.Copy(io.Discard, response.Body)
		_ = response.Body.Close()
		if response.StatusCode != http.StatusUnauthorized {
			t.Fatalf("unauthorized HTTP request status = %d, want %d", response.StatusCode, http.StatusUnauthorized)
		}
	}

	client := sdk.NewClient(&sdk.Implementation{Name: "test-client", Version: "test"}, nil)
	transport := &sdk.StreamableClientTransport{
		Endpoint:   runtime.Endpoint(),
		HTTPClient: &http.Client{Transport: authRoundTripper{next: http.DefaultTransport}},
	}
	ctx, cancel := context.WithTimeout(context.Background(), testShutdownTimeout)
	defer cancel()
	clientSession, err := client.Connect(ctx, transport, nil)
	if err != nil {
		t.Fatalf("connect MCP client: %v", err)
	}
	defer clientSession.Close()
	tools, err := clientSession.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("list MCP tools: %v", err)
	}
	if len(tools.Tools) != 1 || tools.Tools[0].Name != "register_conversation" {
		t.Fatalf("MCP tools = %#v, want only register_conversation", tools.Tools)
	}
	result, err := clientSession.CallTool(ctx, &sdk.CallToolParams{
		Name:      "register_conversation",
		Arguments: map[string]any{"conversation_id": "native-id", "name": "Work"},
	})
	if err != nil {
		t.Fatalf("call registration tool: %v", err)
	}
	if result.IsError || !registrar.registered {
		t.Fatalf("registration result = %#v, registered=%v", result, registrar.registered)
	}
}

//nolint:govet // Each subtest builds an independent request and error value.
func TestRejectsInvalidHostAndOrigin(t *testing.T) {
	registrar := &testRegistrar{}
	runtime, err := mcp.Start(registrar)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = runtime.Shutdown(ctx)
	})

	for _, test := range []struct {
		name   string
		host   string
		origin string
	}{
		{name: "host", host: "example.com"},
		{name: "origin", origin: "https://attacker.example"},
	} {
		t.Run(test.name, func(t *testing.T) {
			request, err := http.NewRequest(http.MethodPost, runtime.Endpoint(), strings.NewReader(`{}`))
			if err != nil {
				t.Fatal(err)
			}
			request.Header.Set("Authorization", "Bearer valid-test-token")
			if test.host != "" {
				request.Host = test.host
			}
			if test.origin != "" {
				request.Header.Set("Origin", test.origin)
			}
			response, err := http.DefaultClient.Do(request)
			if err != nil {
				t.Fatal(err)
			}
			_, _ = io.Copy(io.Discard, response.Body)
			_ = response.Body.Close()
			if response.StatusCode != http.StatusForbidden {
				t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusForbidden)
			}
		})
	}
}
