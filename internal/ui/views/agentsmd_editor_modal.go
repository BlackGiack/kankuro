package views

import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"kankuro/internal/ui/component"
)

// AgentsMDEditorModal allows editing agents.md content or path.
type AgentsMDEditorModal struct {
	component.Base

	mode       int // 0 = content, 1 = path
	editor     textarea.Model
	pathInput  textinput.Model
	preferPath bool
	confirmed  bool
	cancelled  bool
	open       bool
}

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
	pathInput.Width = 30

	return &AgentsMDEditorModal{
		editor:    editor,
		pathInput: pathInput,
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

func (m *AgentsMDEditorModal) Open()           { m.open = true; m.confirmed = false; m.cancelled = false }
func (m *AgentsMDEditorModal) Close()          { m.open = false; m.confirmed = false; m.cancelled = false }
func (m *AgentsMDEditorModal) IsOpen() bool    { return m.open }
func (m *AgentsMDEditorModal) Confirmed() bool { return m.confirmed }
func (m *AgentsMDEditorModal) Cancelled() bool { return m.cancelled }

func (m *AgentsMDEditorModal) Init() tea.Cmd { return nil }

// SetSize implements component.Model.
func (m *AgentsMDEditorModal) SetSize(width, height int) {
	m.Base.SetSize(width, height)

	editorWidth := max(40, width*8/10) //nolint:mnd // responsive sizing
	editorHeight := max(10, height/3)  //nolint:mnd // responsive sizing
	m.editor.SetWidth(editorWidth)
	m.editor.SetHeight(editorHeight)

	m.pathInput.Width = max(30, width*3/5) //nolint:mnd // responsive sizing
}

func (m *AgentsMDEditorModal) Update(msg tea.Msg) (component.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case keyEnter:
			if m.mode == 1 { // Path mode - enter confirms
				m.preferPath = true
				m.confirmed = true
				return m, nil
			}
		case "ctrl+enter":
			m.preferPath = m.mode == 1
			m.confirmed = true
			return m, nil
		case keyEsc:
			m.cancelled = true
			return m, nil
		case keyTab:
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
	}

	var cmd tea.Cmd
	if m.mode == 0 {
		m.editor, cmd = m.editor.Update(msg)
	} else {
		m.pathInput, cmd = m.pathInput.Update(msg)
	}
	return m, cmd
}

func (m *AgentsMDEditorModal) View() string {
	modalWidth := max(50, m.Width*4/5) //nolint:mnd // responsive sizing
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#8E44AD")).
		Padding(vPadding, hPadding).
		Width(modalWidth)

	var content string
	content += "AGENTS.md Editor\n\n"
	content += "Content:\n"
	if m.mode == 0 {
		content += lipgloss.NewStyle().Foreground(highlightColor).Render("> "+m.editor.View()) + "\n"
	} else {
		content += m.editor.View() + "\n"
	}
	content += "\nPath:\n"
	if m.mode == 1 {
		content += lipgloss.NewStyle().Foreground(highlightColor).Render("> "+m.pathInput.View()) + "\n"
	} else {
		content += m.pathInput.View() + "\n"
	}
	content += "\nTab=switch Ctrl+Enter=confirm Esc=cancel"

	return boxStyle.Render(content)
}
