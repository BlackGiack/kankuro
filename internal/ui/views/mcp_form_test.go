package views_test

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"kankuro/internal/config"
	"kankuro/internal/ui/messages"
	"kankuro/internal/ui/views"
)

func TestMCPForm_New(t *testing.T) {
	f := views.NewMCPForm(views.MCPFormCallbacks{})
	if f == nil {
		t.Fatal("NewMCPForm returned nil")
	}
	if len(f.Inputs()) != views.MCPInputCount {
		t.Errorf("Expected %d inputs, got %d", views.MCPInputCount, len(f.Inputs()))
	}
}

func TestMCPForm_SetMCP(t *testing.T) {
	f := views.NewMCPForm(views.MCPFormCallbacks{})
	mcp := config.MCPServer{
		Name:    "Test MCP",
		Command: "node",
		Args:    []string{"index.js"},
		Env:     map[string]string{"DEBUG": "1"},
	}

	f.SetMCP(mcp)

	if f.Inputs()[0].Value() != "Test MCP" {
		t.Errorf("Expected name 'Test MCP', got '%s'", f.Inputs()[0].Value())
	}
	if f.Inputs()[1].Value() != "node" {
		t.Errorf("Expected command 'node', got '%s'", f.Inputs()[1].Value())
	}
	if f.Inputs()[2].Value() != "index.js" {
		t.Errorf("Expected args 'index.js', got '%s'", f.Inputs()[2].Value())
	}
}

func TestMCPForm_Submit(t *testing.T) {
	var savedMCP config.MCPServer
	callbacks := views.MCPFormCallbacks{
		OnSave: func(mcp config.MCPServer) tea.Cmd {
			savedMCP = mcp
			return nil
		},
	}
	f := views.NewMCPForm(callbacks)

	f.Inputs()[0].SetValue("New MCP")
	f.Inputs()[1].SetValue("python")
	f.Inputs()[2].SetValue("main.py,arg1")
	f.Inputs()[5].SetValue("KEY=VAL")

	f.HandleSubmit()

	if savedMCP.Name != "New MCP" {
		t.Errorf("Expected name 'New MCP', got '%s'", savedMCP.Name)
	}
	if savedMCP.Command != "python" {
		t.Errorf("Expected command 'python', got '%s'", savedMCP.Command)
	}
	if len(savedMCP.Args) != 2 || savedMCP.Args[1] != "arg1" {
		t.Errorf("Expected 2 args, got %v", savedMCP.Args)
	}
	if savedMCP.Env["KEY"] != "VAL" {
		t.Errorf("Expected Env[KEY]=VAL, got %s", savedMCP.Env["KEY"])
	}
}

func TestMCPForm_EnvRoundTrip_CommasAndOrder(t *testing.T) {
	f := views.NewMCPForm(views.MCPFormCallbacks{})
	mcp := config.MCPServer{
		Name:    "Test MCP",
		Command: "node",
		Env: map[string]string{
			"ZEBRA":  "last",
			"ACCEPT": "a,b",
			"ARG":    "value=with=equals",
		},
	}

	f.SetMCP(mcp)
	rendered := f.Inputs()[5].Value()

	// Rendering must be stable/sorted (ACCEPT < ARG < ZEBRA).
	wantOrder := "ACCEPT=a\\,b,ARG=value=with=equals,ZEBRA=last"
	if rendered != wantOrder {
		t.Fatalf("expected sorted, escaped rendering %q, got %q", wantOrder, rendered)
	}

	// Re-rendering must be identical (no map-order flakiness).
	f.SetMCP(mcp)
	if again := f.Inputs()[5].Value(); again != rendered {
		t.Fatalf("rendering is not stable: %q vs %q", rendered, again)
	}

	var saved config.MCPServer
	f2 := views.NewMCPForm(views.MCPFormCallbacks{
		OnSave: func(mcp config.MCPServer) tea.Cmd {
			saved = mcp
			return nil
		},
	})
	f2.Inputs()[0].SetValue("Test MCP")
	f2.Inputs()[1].SetValue("node")
	f2.Inputs()[5].SetValue(rendered)
	f2.HandleSubmit()

	if saved.Env["ACCEPT"] != "a,b" {
		t.Errorf("expected ACCEPT=%q, got %q", "a,b", saved.Env["ACCEPT"])
	}
	if saved.Env["ARG"] != "value=with=equals" {
		t.Errorf("expected ARG=%q, got %q", "value=with=equals", saved.Env["ARG"])
	}
	if saved.Env["ZEBRA"] != "last" {
		t.Errorf("expected ZEBRA=%q, got %q", "last", saved.Env["ZEBRA"])
	}
}

func TestMCPForm_MalformedEnvPair_ProducesError(t *testing.T) {
	f := views.NewMCPForm(views.MCPFormCallbacks{})
	f.Inputs()[0].SetValue("Test MCP")
	f.Inputs()[1].SetValue("node")
	f.Inputs()[5].SetValue("NOVALUE")

	cmd := f.HandleSubmit()
	if cmd != nil {
		t.Errorf("expected no save command on malformed input, got %v", cmd)
	}
	if !strings.Contains(f.View(), "malformed key=value pair") {
		t.Errorf("expected error rendered in view, got: %s", f.View())
	}
}

func TestMCPForm_Navigation(t *testing.T) {
	f := views.NewMCPForm(views.MCPFormCallbacks{})

	// Test Esc
	_, cmd := f.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	msg := cmd()
	if _, ok := msg.(messages.NavigateBackMsg); !ok {
		t.Errorf("Expected NavigateBackMsg, got %T", msg)
	}

	// Test Tab
	if !f.Inputs()[0].Focused() {
		t.Error("First input should be focused")
	}
	f.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	if f.Inputs()[0].Focused() {
		t.Error("First input should be blurred after Tab")
	}
	if !f.Inputs()[1].Focused() {
		t.Error("Second input should be focused after Tab")
	}
}
