package views

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"kankuro/internal/config"
	"kankuro/internal/ui/component"
	"kankuro/internal/ui/messages"
	"kankuro/internal/ui/styles"
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

	err error
}

// NewMCPForm creates a new MCP form.
func NewMCPForm(callbacks MCPFormCallbacks) *MCPForm {
	inputs := make([]textinput.Model, mcpInputCount)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Name"
	inputs[0].SetWidth(defaultInputWidth)

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Command"
	inputs[1].SetWidth(defaultInputWidth)

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "Args (comma-separated, use \\, for a literal comma)"
	inputs[2].SetWidth(defaultInputWidth)

	inputs[3] = textinput.New()
	inputs[3].Placeholder = "Working directory"
	inputs[3].SetWidth(defaultInputWidth)

	inputs[4] = textinput.New()
	inputs[4].Placeholder = "URL"
	inputs[4].SetWidth(defaultInputWidth)

	inputs[5] = textinput.New()
	inputs[5].Placeholder = "Environment (key=value, comma-separated, use \\, for a literal comma)"
	inputs[5].SetWidth(defaultInputWidth)

	inputs[6] = textinput.New()
	inputs[6].Placeholder = "Headers (key=value, comma-separated, use \\, for a literal comma)"
	inputs[6].SetWidth(defaultInputWidth)

	inputs[0].Focus()

	return &MCPForm{
		inputs:    inputs,
		callbacks: callbacks,
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
	f.inputs[2].SetValue(formatCommaList(mcp.Args))
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
		content.WriteString(styles.ErrorStyle.Render(f.err.Error()) + "\n\n")
	}

	content.WriteString(styles.TitleStyle.Render("MCP Server Configuration") + "\n\n")

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
		content.WriteString(styles.LabelStyle.Render(field.label) + "\n")
		content.WriteString(field.input.View() + "\n\n")
	}

	content.WriteString(styles.HelpStyle.Render("Tab: Next field | Enter: Save | Esc: Back"))

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
	f.err = nil

	args := parseCommaList(f.inputs[2].Value())

	env, err := parseKeyValueList(f.inputs[5].Value())
	if err != nil {
		f.err = fmt.Errorf("environment: %w", err)
		return nil
	}

	headers, err := parseKeyValueList(f.inputs[6].Value())
	if err != nil {
		f.err = fmt.Errorf("headers: %w", err)
		return nil
	}

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

// Args, env, and header values are entered as a single comma-separated line.
// Since commas are common inside those values (e.g. "Accept: a,b", a JSON
// arg, a PATH-like list), a literal comma or backslash in a value must be
// escaped with a backslash (`\,`, `\\`) to distinguish it from the
// delimiter. splitEscaped/escapeItem implement that round-trip.

// splitEscaped splits input on unescaped commas, unescaping `\,` and `\\`
// into a literal `,` and `\` in the returned tokens.
func splitEscaped(input string) []string {
	var result []string
	var cur strings.Builder
	escaped := false
	for _, r := range input {
		switch {
		case escaped:
			cur.WriteRune(r)
			escaped = false
		case r == '\\':
			escaped = true
		case r == ',':
			result = append(result, cur.String())
			cur.Reset()
		default:
			cur.WriteRune(r)
		}
	}
	result = append(result, cur.String())
	return result
}

// escapeItem escapes backslashes and commas so the value survives a
// round-trip through splitEscaped.
func escapeItem(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `,`, `\,`)
	return s
}

func parseCommaList(input string) []string {
	var result []string
	for _, p := range splitEscaped(input) {
		if s := strings.TrimSpace(p); s != "" {
			result = append(result, s)
		}
	}
	return result
}

func formatCommaList(items []string) string {
	parts := make([]string, len(items))
	for i, item := range items {
		parts[i] = escapeItem(item)
	}
	return strings.Join(parts, ",")
}

// parseKeyValueList parses a comma-separated list of key=value pairs. It
// returns an error instead of silently dropping a pair that has no `=`, so a
// typo is surfaced rather than vanishing.
func parseKeyValueList(input string) (map[string]string, error) {
	result := make(map[string]string)
	for _, pair := range splitEscaped(input) {
		if strings.TrimSpace(pair) == "" {
			continue
		}
		k, v, ok := strings.Cut(pair, "=")
		if !ok {
			return nil, fmt.Errorf("malformed key=value pair: %q", pair)
		}
		result[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	return result, nil
}

// formatKeyValueMap renders a map as a sorted, comma-separated key=value
// list. Sorting keeps the rendered text stable across calls (Go randomizes
// map iteration order, which previously made the form's env/header fields
// visibly reshuffle as the user edited them).
func formatKeyValueMap(m map[string]string) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, escapeItem(k)+"="+escapeItem(m[k]))
	}
	return strings.Join(parts, ",")
}
