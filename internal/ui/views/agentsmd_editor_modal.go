package views

import (
	"strings"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/BlackGiack/kankuro/internal/ui/component"
	"github.com/BlackGiack/kankuro/internal/ui/components"
	"github.com/BlackGiack/kankuro/internal/ui/styles"
)

// AgentsMDEditorModal allows editing agents.md content or path.
type AgentsMDEditorModal struct {
	component.Base
	components.ModalState

	mode       int // 0 = content, 1 = path
	editor     textarea.Model
	pathInput  textinput.Model
	preferPath bool

	// modalWidth is the box's width, border and padding included. The editor
	// and path input are sized from it by SetSize, from the inside out.
	modalWidth int
}

// Layout of the box. The fixed text rows are the title, a blank, "Content:",
// a blank, "Path:", the path input, a blank and the help line; the text area
// takes whatever height is left.
const (
	agentsMDMinModalWidth = 50
	agentsMDMarkerWidth   = 2 // "> " in front of the focused field
	agentsMDBorder        = 2
	agentsMDBoxChrome     = agentsMDBorder + 2*hPadding // columns the box takes from its width
	agentsMDFixedRows     = 8
	agentsMDMinEditorRows = 3
	agentsMDTextInputPad  = 3 // a text input draws its "> " prompt and cursor beyond its width
)

// NewAgentsMDEditorModal creates a new agents.md editor.
func NewAgentsMDEditorModal() *AgentsMDEditorModal {
	const (
		editorWidth  = 40
		editorHeight = 10
		modalPadding = 2
	)
	editor := textarea.New()
	editor.SetWidth(editorWidth)
	editor.SetHeight(editorHeight)
	editor.Placeholder = "Enter AGENTS.md content..."

	pathInput := textinput.New()
	pathInput.Placeholder = "Path to AGENTS.md file..."
	pathInput.SetWidth(smallInputWidth)

	return &AgentsMDEditorModal{
		editor:     editor,
		pathInput:  pathInput,
		modalWidth: agentsMDMinModalWidth,
	}
}

func (m *AgentsMDEditorModal) SetState(state AgentsMDState) {
	m.editor.SetValue(state.Content)
	m.pathInput.SetValue(state.Path)
	m.preferPath = state.PreferPath
	if m.preferPath {
		m.mode = 1
		m.pathInput.Focus()
	} else {
		m.mode = 0
		m.editor.Focus()
	}
}

func (m *AgentsMDEditorModal) State() AgentsMDState {
	return AgentsMDState{
		Content:    m.editor.Value(),
		Path:       m.pathInput.Value(),
		PreferPath: m.preferPath,
	}
}

func (m *AgentsMDEditorModal) Init() tea.Cmd { return nil }

// SetSize implements component.Model. Sizes are derived from the inside out so
// that nothing is wider than the box it sits in: the box is at most the space
// given, its inner width is the box minus border and padding, and the fields
// get the inner width minus the marker column.
func (m *AgentsMDEditorModal) SetSize(width, height int) {
	m.Base.SetSize(width, height)

	m.modalWidth = agentsMDMinModalWidth
	if width > 0 {
		m.modalWidth = min(width, max(agentsMDMinModalWidth, width*4/5)) //nolint:mnd // responsive sizing
	}
	inner := m.modalWidth - agentsMDBoxChrome
	fieldWidth := max(1, inner-agentsMDMarkerWidth)
	m.editor.SetWidth(fieldWidth)
	m.pathInput.SetWidth(max(1, fieldWidth-agentsMDTextInputPad))

	chrome := agentsMDBorder + 2*vPadding + agentsMDFixedRows
	editorHeight := max(agentsMDMinEditorRows, min(max(10, height/3), height-chrome)) //nolint:mnd // responsive sizing
	m.editor.SetHeight(editorHeight)
}

func (m *AgentsMDEditorModal) Update(msg tea.Msg) (component.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch keyMsg.Code {
		case tea.KeyEnter:
			if m.mode == 1 { // Path mode - enter confirms
				m.preferPath = true
				m.Confirm()
				return m, nil
			}
		case tea.KeyEsc:
			m.Cancel()
			return m, nil
		case tea.KeyTab:
			m.mode = 1 - m.mode
			if m.mode == 0 {
				m.editor.Focus()
				m.pathInput.Blur()
			} else {
				m.editor.Blur()
				m.pathInput.Focus()
			}
			return m, nil
		}

		if keyMsg.String() == "ctrl+enter" {
			m.preferPath = m.mode == 1
			m.Confirm()
			return m, nil
		}
	}

	var cmd tea.Cmd
	if m.mode == 0 {
		m.editor, cmd = m.editor.Update(msg)
	} else {
		m.pathInput, cmd = m.pathInput.Update(msg)
	}
	return m, cmd
}

// markedField draws a field with the focus marker as its own column. Prefixing the
// text with "> " would mark only its first line and make the rows wider than
// the area; a column keeps both focus states the same width.
func markedField(view string, focused bool) string {
	marker := strings.Repeat(" ", agentsMDMarkerWidth)
	if focused {
		marker = lipgloss.NewStyle().Foreground(styles.HighlightColor).Render("> ")
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, marker, view)
}

func (m *AgentsMDEditorModal) View() string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#8E44AD")).
		Padding(vPadding, hPadding).
		Width(m.modalWidth)

	var content string
	content += "AGENTS.md Editor\n\n"
	content += "Content:\n"
	content += markedField(m.editor.View(), m.mode == 0) + "\n"
	content += "\nPath:\n"
	content += markedField(m.pathInput.View(), m.mode == 1) + "\n"
	content += "\nTab=switch Ctrl+Enter=confirm Esc=cancel"

	return boxStyle.Render(content)
}
