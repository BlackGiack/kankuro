package views_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"kankuro/internal/config"
	"kankuro/internal/ui/commands"
	"kankuro/internal/ui/messages"
	"kankuro/internal/ui/views"
)

func TestMCPForm_New(t *testing.T) {
	f := views.NewMCPForm(views.MCPFormCallbacks{}, &commands.Commands{})
	if f == nil {
		t.Fatal("NewMCPForm returned nil")
	}
	if len(f.Inputs()) != views.MCPInputCount {
		t.Errorf("Expected %d inputs, got %d", views.MCPInputCount, len(f.Inputs()))
	}
}

func TestMCPForm_SetMCP(t *testing.T) {
	f := views.NewMCPForm(views.MCPFormCallbacks{}, &commands.Commands{})
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
	f := views.NewMCPForm(callbacks, &commands.Commands{})

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

func TestMCPForm_Navigation(t *testing.T) {
	f := views.NewMCPForm(views.MCPFormCallbacks{}, &commands.Commands{})

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
