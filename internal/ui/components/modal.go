package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"kankuro/internal/ui/component"
)

// Modal is a reusable overlay component for dialogs.
type Modal struct {
	component.Base

	title      string
	content    string
	open       bool
	onConfirm  func()
	onCancel   func()
	modalWidth int
}

// NewModal creates a new modal with the given title.
func NewModal(title string) *Modal {
	return &Modal{
		title:      title,
		open:       false,
		modalWidth: 50, //nolint:mnd // responsive sizing
	}
}

// SetSize implements component.Model.
func (m *Modal) SetSize(width, height int) {
	m.Base.SetSize(width, height)
	m.modalWidth = max(40, width*3/5) //nolint:mnd // responsive sizing
}

// SetContent sets the modal's content string.
func (m *Modal) SetContent(content string) {
	m.content = content
}

// Open makes the modal visible.
func (m *Modal) Open() {
	m.open = true
}

// Close hides the modal.
func (m *Modal) Close() {
	m.open = false
}

// IsOpen returns whether modal is currently visible.
func (m *Modal) IsOpen() bool {
	return m.open
}

// SetConfirm sets the confirm callback.
func (m *Modal) SetConfirm(callback func()) {
	m.onConfirm = callback
}

// SetCancel sets the cancel callback.
func (m *Modal) SetCancel(callback func()) {
	m.onCancel = callback
}

// Init implements component.Model.
func (m *Modal) Init() tea.Cmd {
	return nil
}

// Update handles modal key events.
func (m *Modal) Update(_ tea.Msg) (component.Model, tea.Cmd) {
	// Key handling is done by parent view
	// Modal just renders and exposes Open/Close/IsOpen
	return m, nil
}

// View renders the modal overlay.
func (m *Modal) View() string {
	if !m.open {
		return ""
	}

	dialog := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#8E44AD")).
		Padding(modalPaddingHeight, modalPaddingWidth).
		Width(m.modalWidth)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF"))

	content := titleStyle.Render(m.title)
	if m.content != "" {
		content += "\n\n" + m.content
	}

	return dialog.Render(content)
}
