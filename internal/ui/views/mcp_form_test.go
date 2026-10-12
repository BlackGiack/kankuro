package views_test

import (
	"reflect"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/views"
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

// TestMCPFormRoundTripsEveryField sets an MCPServer with a distinct value in
// every field, submits it through Update(Enter) and expects the same server
// back. A swapped input index in SetMCP or handleSubmit moves a value into
// the wrong field and fails here; the drift test cannot see that.
func TestMCPFormRoundTripsEveryField(t *testing.T) {
	var saved config.MCPServer
	saveCalls := 0
	f := views.NewMCPForm(views.MCPFormCallbacks{
		OnSave: func(mcp config.MCPServer) tea.Cmd {
			saved = mcp
			saveCalls++
			return nil
		},
	})

	want := config.MCPServer{
		ID:      "id-1",
		Name:    "n",
		Command: "cmd",
		Args:    []string{"a1", "a,2"},
		CWD:     "/cwd",
		URL:     "http://u",
		Env:     map[string]string{"E": "1"},
		Headers: map[string]string{"H": "2"},
	}
	f.SetMCP(want)
	f.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if saveCalls != 1 {
		t.Fatalf("expected OnSave to run once, ran %d times", saveCalls)
	}
	if saved.ID != want.ID {
		t.Errorf("ID: got %q, want %q", saved.ID, want.ID)
	}
	if saved.Name != want.Name {
		t.Errorf("Name: got %q, want %q", saved.Name, want.Name)
	}
	if saved.Command != want.Command {
		t.Errorf("Command: got %q, want %q", saved.Command, want.Command)
	}
	if !reflect.DeepEqual(saved.Args, want.Args) {
		t.Errorf("Args: got %q, want %q", saved.Args, want.Args)
	}
	if saved.CWD != want.CWD {
		t.Errorf("CWD: got %q, want %q", saved.CWD, want.CWD)
	}
	if saved.URL != want.URL {
		t.Errorf("URL: got %q, want %q", saved.URL, want.URL)
	}
	if !reflect.DeepEqual(saved.Env, want.Env) {
		t.Errorf("Env: got %v, want %v", saved.Env, want.Env)
	}
	if !reflect.DeepEqual(saved.Headers, want.Headers) {
		t.Errorf("Headers: got %v, want %v", saved.Headers, want.Headers)
	}
}

// TestMCPFormLabelsMatchPlaceholders renders an empty form and expects each
// label to precede its own input's placeholder, which catches a field-table
// row whose label and input have drifted apart.
func TestMCPFormLabelsMatchPlaceholders(t *testing.T) {
	f := views.NewMCPForm(views.MCPFormCallbacks{})
	view := ansi.Strip(f.View())

	pairs := []struct{ label, placeholder string }{
		{"Name", "Name"},
		{"Command", "Command"},
		{"Args", "Args (comma-separated"},
		{"Working Directory", "Working directory"},
		{"URL", "URL"},
		{"Environment", "Environment (key=value"},
		{"Headers", "Headers (key=value"},
	}

	// Walk the view in order: each label, then its placeholder on a later
	// line, then the next label.
	pos := 0
	for _, p := range pairs {
		li := strings.Index(view[pos:], p.label+"\n")
		if li < 0 {
			t.Fatalf("label %q not found after offset %d in:\n%s", p.label, pos, view)
		}
		pos += li + len(p.label) + 1
		pi := strings.Index(view[pos:], p.placeholder)
		nextLabelGap := strings.Index(view[pos:], "\n\n")
		if pi < 0 || (nextLabelGap >= 0 && pi > nextLabelGap) {
			t.Fatalf("placeholder %q does not directly follow label %q in:\n%s", p.placeholder, p.label, view)
		}
		pos += pi + len(p.placeholder)
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
