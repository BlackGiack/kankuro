package views

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/ui/component"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/styles"
)

// Indices of the form's text inputs. Every field is a text input, so the
// index is the position in MCPForm.inputs.
const (
	mcpNameIdx = iota
	mcpCommandIdx
	mcpArgsIdx
	mcpCWDIdx
	mcpURLIdx
	mcpEnvIdx
	mcpHeadersIdx

	mcpInputCount // keep last
)

type mcpField struct {
	idx         int
	label       string // shown above the input in View
	placeholder string
}

// mcpFieldTable is the single source for the form's field order, labels and
// placeholders. NewMCPForm and View both read it; SetMCP and handleSubmit
// address inputs by the named constants above.
func mcpFieldTable() []mcpField {
	return []mcpField{
		{mcpNameIdx, "Name", "Name"},
		{mcpCommandIdx, "Command", "Command"},
		{mcpArgsIdx, "Args", "Args (comma-separated, use \\, for a literal comma)"},
		{mcpCWDIdx, "Working Directory", "Working directory"},
		{mcpURLIdx, "URL", "URL"},
		{mcpEnvIdx, "Environment", "Environment (key=value, comma-separated, use \\, for a literal comma)"},
		{mcpHeadersIdx, "Headers", "Headers (key=value, comma-separated, use \\, for a literal comma)"},
	}
}

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
	for _, field := range mcpFieldTable() {
		in := textinput.New()
		in.Placeholder = field.placeholder
		in.SetWidth(defaultInputWidth)
		inputs[field.idx] = in
	}

	inputs[mcpNameIdx].Focus()

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
	f.inputs[mcpNameIdx].SetValue(mcp.Name)
	f.inputs[mcpCommandIdx].SetValue(mcp.Command)
	f.inputs[mcpArgsIdx].SetValue(formatCommaList(mcp.Args))
	f.inputs[mcpCWDIdx].SetValue(mcp.CWD)
	f.inputs[mcpURLIdx].SetValue(mcp.URL)
	f.inputs[mcpEnvIdx].SetValue(formatKeyValueMap(mcp.Env))
	f.inputs[mcpHeadersIdx].SetValue(formatKeyValueMap(mcp.Headers))
}

// Reset clears the form.
func (f *MCPForm) Reset() {
	f.mcp = config.MCPServer{}
	for i := range f.inputs {
		f.inputs[i].SetValue("")
		f.inputs[i].Blur()
	}
	f.inputs[mcpNameIdx].Focus()
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

	for _, field := range mcpFieldTable() {
		content.WriteString(styles.LabelStyle.Render(field.label) + "\n")
		content.WriteString(f.inputs[field.idx].View() + "\n\n")
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
	f.inputs[mcpNameIdx].Focus()
}

func (f *MCPForm) handleSubmit() tea.Cmd {
	f.err = nil

	args := parseCommaList(f.inputs[mcpArgsIdx].Value())

	env, err := parseKeyValueList(f.inputs[mcpEnvIdx].Value())
	if err != nil {
		f.err = fmt.Errorf("environment: %w", err)
		return nil
	}

	headers, err := parseKeyValueList(f.inputs[mcpHeadersIdx].Value())
	if err != nil {
		f.err = fmt.Errorf("headers: %w", err)
		return nil
	}

	mcp := config.MCPServer{
		ID:      f.mcp.ID,
		Name:    strings.TrimSpace(f.inputs[mcpNameIdx].Value()),
		Command: strings.TrimSpace(f.inputs[mcpCommandIdx].Value()),
		Args:    args,
		CWD:     strings.TrimSpace(f.inputs[mcpCWDIdx].Value()),
		URL:     strings.TrimSpace(f.inputs[mcpURLIdx].Value()),
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
