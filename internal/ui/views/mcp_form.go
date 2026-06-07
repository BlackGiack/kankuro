package views

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"kankuro/internal/config"
	"kankuro/internal/ui/commands"
	"kankuro/internal/ui/component"
	"kankuro/internal/ui/messages"
)

const (
	mcpInputCount = 7
)

// MCPInputCount is the number of inputs in the MCP form.
const MCPInputCount = mcpInputCount

// MCPFormCallbacks defines callbacks for MCP form actions.
type MCPFormCallbacks struct {
	OnSave   func(mcp config.MCPServer) tea.Cmd
	OnDelete func(mcpID string) error
}

// MCPForm is a simple form for creating/editing MCP servers.
type MCPForm struct {
	component.Base

	inputs []textinput.Model
	mcp    config.MCPServer

	callbacks MCPFormCallbacks
	cmds      *commands.Commands

	err error
}

// NewMCPForm creates a new MCP form.
func NewMCPForm(callbacks MCPFormCallbacks, cmds *commands.Commands) *MCPForm {
	inputs := make([]textinput.Model, mcpInputCount)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Name"
	inputs[0].SetWidth(defaultInputWidth)

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Command"
	inputs[1].SetWidth(defaultInputWidth)

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "Args (comma-separated)"
	inputs[2].SetWidth(defaultInputWidth)

	inputs[3] = textinput.New()
	inputs[3].Placeholder = "Working directory"
	inputs[3].SetWidth(defaultInputWidth)

	inputs[4] = textinput.New()
	inputs[4].Placeholder = "URL"
	inputs[4].SetWidth(defaultInputWidth)

	inputs[5] = textinput.New()
	inputs[5].Placeholder = "Environment (key=value, comma-separated)"
	inputs[5].SetWidth(defaultInputWidth)

	inputs[6] = textinput.New()
	inputs[6].Placeholder = "Headers (key=value, comma-separated)"
	inputs[6].SetWidth(defaultInputWidth)

	inputs[0].Focus()

	return &MCPForm{
		inputs:    inputs,
		callbacks: callbacks,
		cmds:      cmds,
	}
}

// Inputs returns the underlying text input models.
func (f *MCPForm) Inputs() []textinput.Model {
	return f.inputs
}

// HandleSubmit manually triggers the submit logic.
func (f *MCPForm) HandleSubmit() tea.Cmd {
	return f.handleSubmit()
}

// SetMCP sets the MCP server to edit.
func (f *MCPForm) SetMCP(mcp config.MCPServer) {
	f.mcp = mcp
	f.inputs[0].SetValue(mcp.Name)
	f.inputs[1].SetValue(mcp.Command)
	f.inputs[2].SetValue(strings.Join(mcp.Args, ","))
	f.inputs[3].SetValue(mcp.CWD)
	f.inputs[4].SetValue(mcp.URL)
	f.inputs[5].SetValue(formatKeyValueMap(mcp.Env))
	f.inputs[6].SetValue(formatKeyValueMap(mcp.Headers))
}

// Reset clears the form.
func (f *MCPForm) Reset() {
	f.mcp = config.MCPServer{}
	for i := range f.inputs {
		f.inputs[i].SetValue("")
		f.inputs[i].Blur()
	}
	f.inputs[0].Focus()
	f.err = nil
}

// Init initializes the component.
func (f *MCPForm) Init() tea.Cmd {
	f.Reset()
	return nil
}

// Update handles messages.
func (f *MCPForm) Update(msg tea.Msg) (component.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		if keyMsg.String() == keyCtrlC {
			return f, tea.Quit
		}
		switch keyMsg.Code {
		case tea.KeyEsc:
			return f, func() tea.Msg { return messages.NavigateBackMsg{} }
		case tea.KeyTab:
			f.cycleFocus()
			return f, nil
		case tea.KeyEnter:
			return f, f.handleSubmit()
		}
	}

	cmds := make([]tea.Cmd, len(f.inputs))
	for i := range f.inputs {
		f.inputs[i], cmds[i] = f.inputs[i].Update(msg)
	}
	return f, tea.Batch(cmds...)
}

// View renders the form.
func (f *MCPForm) View() string {
	var content strings.Builder

	if f.err != nil {
		content.WriteString(errorStyle.Render(f.err.Error()) + "\n\n")
	}

	content.WriteString(titleStyle.Render("MCP Server Configuration") + "\n\n")

	fields := []struct {
		label string
		input *textinput.Model
	}{
		{"Name", &f.inputs[0]},
		{"Command", &f.inputs[1]},
		{"Args", &f.inputs[2]},
		{"Working Directory", &f.inputs[3]},
		{"URL", &f.inputs[4]},
		{"Environment", &f.inputs[5]},
		{"Headers", &f.inputs[6]},
	}

	for _, field := range fields {
		content.WriteString(labelStyle.Render(field.label) + "\n")
		content.WriteString(field.input.View() + "\n\n")
	}

	content.WriteString(helpStyle.Render("Tab: Next field | Enter: Save | Esc: Back"))

	return content.String()
}

func (f *MCPForm) cycleFocus() {
	for i := range f.inputs {
		if f.inputs[i].Focused() {
			f.inputs[i].Blur()
			next := (i + 1) % len(f.inputs)
			f.inputs[next].Focus()
			return
		}
	}
	f.inputs[0].Focus()
}

func (f *MCPForm) handleSubmit() tea.Cmd {
	args := parseCommaList(f.inputs[2].Value())
	env := parseKeyValueList(f.inputs[5].Value())
	headers := parseKeyValueList(f.inputs[6].Value())

	mcp := config.MCPServer{
		ID:      f.mcp.ID,
		Name:    strings.TrimSpace(f.inputs[0].Value()),
		Command: strings.TrimSpace(f.inputs[1].Value()),
		Args:    args,
		CWD:     strings.TrimSpace(f.inputs[3].Value()),
		URL:     strings.TrimSpace(f.inputs[4].Value()),
		Env:     env,
		Headers: headers,
	}

	if f.callbacks.OnSave != nil {
		return f.callbacks.OnSave(mcp)
	}

	return func() tea.Msg { return messages.NavigateBackMsg{} }
}

func parseCommaList(input string) []string {
	parts := strings.Split(input, ",")
	var result []string
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			result = append(result, s)
		}
	}
	return result
}

func parseKeyValueList(input string) map[string]string {
	result := make(map[string]string)
	for pair := range strings.SplitSeq(input, ",") {
		if k, v, ok := strings.Cut(pair, "="); ok {
			result[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	return result
}

func formatKeyValueMap(m map[string]string) string {
	var parts []string
	for k, v := range m {
		parts = append(parts, k+"="+v)
	}
	return strings.Join(parts, ",")
}
