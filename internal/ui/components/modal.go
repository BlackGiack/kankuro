package components

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/BlackGiack/kankuro/internal/ui/component"
)

// ModalState is the open/confirmed/cancelled bookkeeping shared by every
// modal. Embed it (by value) in a modal type to get Open, Close, IsOpen,
// Confirmed and Cancelled; call Confirm/Cancel from the modal's Update. A
// modal that needs extra work on open defines its own Open and calls
// ModalState.Open from it, otherwise a stale confirm/cancel survives a reopen.
type ModalState struct {
	open, confirmed, cancelled bool
}

// Open shows the modal and clears any previous confirm/cancel.
func (s *ModalState) Open() { s.open = true; s.confirmed = false; s.cancelled = false }

// Close hides the modal and clears any confirm/cancel.
func (s *ModalState) Close() { s.open = false; s.confirmed = false; s.cancelled = false }

// IsOpen returns whether the modal is currently visible.
func (s *ModalState) IsOpen() bool { return s.open }

// Confirmed returns whether the user confirmed since the modal was opened.
func (s *ModalState) Confirmed() bool { return s.confirmed }

// Cancelled returns whether the user cancelled since the modal was opened.
func (s *ModalState) Cancelled() bool { return s.cancelled }

// Confirm records a confirmation. It does not close the modal.
func (s *ModalState) Confirm() { s.confirmed = true }

// Cancel records a cancellation. It does not close the modal.
func (s *ModalState) Cancel() { s.cancelled = true }

// Modal is a reusable overlay component for dialogs.
type Modal struct {
	component.Base
	ModalState

	title      string
	content    string
	legend     string
	modalWidth int
}

// NewModal creates a new modal with the given title.
func NewModal(title string) *Modal {
	return &Modal{
		title:      title,
		modalWidth: 40, //nolint:mnd // responsive sizing
	}
}

// SetSize implements component.Model.
func (m *Modal) SetSize(width, height int) {
	m.Base.SetSize(width, height)
	m.modalWidth = min(width, max(40, width*3/5)) //nolint:mnd // responsive sizing; never wider than the space given
}

// SetContent sets the modal's content string.
func (m *Modal) SetContent(content string) {
	m.content = content
}

// SetLegend sets the modal's legend help text.
func (m *Modal) SetLegend(legend string) {
	m.legend = legend
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
	if !m.IsOpen() {
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

	if m.legend != "" {
		legendStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
		content += "\n\n" + legendStyle.Render(m.legend)
	}

	return dialog.Render(content)
}
