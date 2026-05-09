package components

import (
	"regexp"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"kankuro/internal/config"
	"kankuro/internal/session"
	"kankuro/internal/ui/component"
)

// Preview layout .
const (
	PreviewIDLabelPrefixLen   = 4
	PreviewCardLabelPrefixLen = 6
	PreviewEllipsisLen        = 3
	PreviewBorderLines        = 2
	PreviewMaxScrollbackBytes = 8 * 1024 // 8KB cap

	WidthBorders = 6

	scrollbackHeightAdj = 2
	lineTruncateAdj     = 5
)

// Regex patterns for text sanitization.
var (
	ansiEscapeRE    = regexp.MustCompile(`\x1b\[[0-9;]*[A-Za-z]`)
	controlCharRE   = regexp.MustCompile(`[\x00-\x08\x0b\x0c\x0e-\x1f\x7f]`)
	kittyProtocolRE = regexp.MustCompile(`(?:\x1b_|\x1b\[|\x1b\]|\x1bP).*?(\x1b\\|[\x07\x40-\x7e])`)
)

// Theme colors (local copies to keep component self-contained for now, or use styles.go).
var (
	ColorPrimary   = lipgloss.Color("#8E44AD") //nolint:gochecknoglobals // UI theme colors
	ColorHighlight = lipgloss.Color("#16A085") //nolint:gochecknoglobals // UI theme colors
	ColorHotPink   = lipgloss.Color("#FF69B4") //nolint:gochecknoglobals // UI theme colors
	ColorGray      = lipgloss.Color("240")     //nolint:gochecknoglobals // UI theme colors
	ColorWhite     = lipgloss.Color("#FFFFFF") //nolint:gochecknoglobals // UI theme colors
)

// Status badge styles.
var (
	StatusRunningStyle = lipgloss.NewStyle(). //nolint:gochecknoglobals // UI style definitions
				Foreground(ColorWhite).
				Background(ColorHighlight).
				Padding(0, 1)

	StatusWaitingStyle = lipgloss.NewStyle(). //nolint:gochecknoglobals // UI style definitions
				Foreground(ColorWhite).
				Background(ColorHotPink).
				Padding(0, 1)

	StatusTerminatedStyle = lipgloss.NewStyle(). //nolint:gochecknoglobals // UI style definitions
				Foreground(ColorWhite).
				Background(ColorGray).
				Padding(0, 1)

	TitleStyle = lipgloss.NewStyle(). //nolint:gochecknoglobals // UI style definitions
			Bold(true).
			Foreground(ColorPrimary)
)

// SessionPreview displays detailed information about a selected session.
type SessionPreview struct {
	component.Base

	sessionID     string
	snapshot      *session.Snapshot
	card          *config.IDCard
	scrollback    []byte
	truncateLimit int
}

// NewSessionPreview creates a new session preview component.
func NewSessionPreview() *SessionPreview {
	return &SessionPreview{
		truncateLimit: 25, //nolint:mnd // responsive sizing
	}
}

// SetSize implements component.Model.
func (p *SessionPreview) SetSize(width, height int) {
	p.Base.SetSize(width, height)
	p.truncateLimit = max(10, width/3) //nolint:mnd // responsive sizing
}

// SetSession updates the displayed session.
func (p *SessionPreview) SetSession(snap *session.Snapshot, card *config.IDCard) {
	p.snapshot = snap
	p.card = card
	if snap != nil {
		p.sessionID = snap.ID
	} else {
		p.sessionID = ""
	}
}

// SetScrollback updates the scrollback buffer.
func (p *SessionPreview) SetScrollback(data []byte) {
	p.scrollback = data
}

// AppendScrollback appends data to the scrollback buffer.
func (p *SessionPreview) AppendScrollback(data []byte) {
	p.scrollback = append(p.scrollback, data...)
	if len(p.scrollback) > PreviewMaxScrollbackBytes {
		p.scrollback = p.scrollback[len(p.scrollback)-PreviewMaxScrollbackBytes:]
	}
}

// SetSessionByID sets the session by ID (requires separate SetSession call).
func (p *SessionPreview) SetSessionByID(sessionID string) {
	p.sessionID = sessionID
}

// Clear resets the preview to empty state.
func (p *SessionPreview) Clear() {
	p.sessionID = ""
	p.snapshot = nil
	p.card = nil
	p.scrollback = nil
}

// SessionID returns the current session ID.
func (p *SessionPreview) SessionID() string {
	return p.sessionID
}

// Snapshot returns the current session snapshot.
func (p *SessionPreview) Snapshot() *session.Snapshot {
	return p.snapshot
}

// Card returns the current ID card.
func (p *SessionPreview) Card() *config.IDCard {
	return p.card
}

// Scrollback returns the current scrollback data.
func (p *SessionPreview) Scrollback() []byte {
	return p.scrollback
}

// Init initializes the component.
func (p *SessionPreview) Init() tea.Cmd {
	return nil
}

// Update handles messages (passive component, no interaction).
func (p *SessionPreview) Update(_ tea.Msg) (component.Model, tea.Cmd) {
	// SessionPreview is display-only, no interaction handling
	return p, nil
}

// View renders the session preview.
func (p *SessionPreview) View() string {
	if p.sessionID == "" || p.snapshot == nil {
		return p.renderPlaceholder()
	}

	return p.renderPreview()
}

// renderPlaceholder shows the empty state.
func (p *SessionPreview) renderPlaceholder() string {
	placeholder := `
   No session
   selected

   Select a session
   from the list to
   see a preview here.
`

	const titleOffset = 2

	return lipgloss.NewStyle().
		Width(p.Width-WidthBorders).
		Height(p.Height+titleOffset).
		Align(lipgloss.Center, lipgloss.Center).
		Foreground(ColorGray).
		Render(placeholder)
}

// renderPreview renders the full session preview.
func (p *SessionPreview) renderPreview() string {
	statusBadge := RenderStatusBadge(p.snapshot.Status)

	sessionName := p.snapshot.Name
	if sessionName == "" {
		sessionName = p.snapshot.ID
	}

	wdStr := TruncatePath(p.snapshot.WorkingDir, p.truncateLimit)
	var startedStr string
	if !p.snapshot.StartedAt.IsZero() {
		startedStr = p.snapshot.StartedAt.Format("15:04:05")
	} else {
		startedStr = "N/A"
	}

	// Render puppet
	puppetStyle := "kankuro"
	if p.card != nil && p.card.Puppet != "" {
		puppetStyle = p.card.Puppet
	}
	puppetStr := RenderAgentPuppet(puppetStyle)

	// Build header
	headerLines := []string{
		TitleStyle.Render(truncateString(sessionName, p.Width-WidthBorders)),
		lipgloss.JoinHorizontal(lipgloss.Center, statusBadge),
		lipgloss.NewStyle().Foreground(ColorGray).Render(
			"ID: " + truncateString(p.snapshot.ID, p.Width-PreviewIDLabelPrefixLen-WidthBorders),
		),
		lipgloss.NewStyle().Foreground(ColorGray).Render("WS: " + wdStr),
		lipgloss.NewStyle().Foreground(ColorGray).Render("Started: " + startedStr),
	}

	if p.card != nil {
		headerLines = append(headerLines,
			lipgloss.NewStyle().Foreground(ColorGray).Render(
				"Card: "+truncateString(p.card.Name, p.Width-PreviewCardLabelPrefixLen-WidthBorders),
			),
		)
	}
	headerLines = append(headerLines, "")

	header := strings.Join(headerLines, "\n")

	// Calculate scrollback height
	// Approximate header height: len(headerLines) + some padding
	// But let's use a fixed offset for now or calculate it
	actualHeaderHeight := lipgloss.Height(header)
	scrollbackHeight := p.Height - actualHeaderHeight + scrollbackHeightAdj

	// Combine header with puppet
	headerWithPuppet := lipgloss.JoinHorizontal(lipgloss.Top, header, "   ", puppetStr)

	scrollbackContent := p.renderScrollback(scrollbackHeight)

	// Combine header and scrollback
	body := lipgloss.JoinVertical(lipgloss.Left,
		headerWithPuppet,
		scrollbackContent,
	)

	return lipgloss.NewStyle().Render(body)
}

// renderScrollback renders the scrollback buffer area.
func (p *SessionPreview) renderScrollback(height int) string {
	if len(p.scrollback) == 0 {
		return lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(ColorGray).
			Width(p.Width).
			Height(max(1, height-1)).
			Align(lipgloss.Center, lipgloss.Center).
			Foreground(ColorGray).
			Render("(no output yet)")
	}

	// Cap scrollback size
	scrollback := p.scrollback
	if len(scrollback) > PreviewMaxScrollbackBytes {
		scrollback = scrollback[len(scrollback)-PreviewMaxScrollbackBytes:]
	}

	// Sanitize text
	text := SanitizeScrollback(scrollback)

	// Split and limit lines
	lines := strings.Split(text, "\n")
	maxLines := max(1, height-PreviewBorderLines)
	if len(lines) > maxLines {
		lines = lines[len(lines)-maxLines:]
	}

	// Process lines
	var truncated []string
	for _, line := range lines {
		clean := strings.TrimSpace(line)
		if clean == "" {
			continue
		}
		if len(clean) > p.Width-2 {
			truncated = append(truncated, clean[:p.Width-lineTruncateAdj]+"...")
		} else {
			truncated = append(truncated, clean)
		}
	}

	if len(truncated) == 0 {
		return lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(ColorGray).
			Width(p.Width).
			Height(max(1, height-1)).
			Align(lipgloss.Center, lipgloss.Center).
			Foreground(ColorGray).
			Render("(no output yet)")
	}

	content := strings.Join(truncated, "\n")

	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(ColorGray).
		Width(p.Width).
		Height(max(1, height-1)).
		Render(content)
}

// RenderStatusBadge renders the status badge.
func RenderStatusBadge(status session.Status) string {
	switch status {
	case session.StatusRunning:
		return StatusRunningStyle.Render("Running")
	case session.StatusWaiting:
		return StatusWaitingStyle.Render("Waiting")
	case session.StatusTerminated:
		return StatusTerminatedStyle.Render("Terminated")
	default:
		return StatusTerminatedStyle.Render("Unknown")
	}
}

// SanitizeScrollback strips ANSI codes and control characters.
func SanitizeScrollback(data []byte) string {
	// Strip ANSI sequences
	stripped := ansiEscapeRE.ReplaceAll(data, []byte{})

	// Strip control characters
	stripped = controlCharRE.ReplaceAll(stripped, []byte{})

	// Strip Kitty protocol sequences
	stripped = kittyProtocolRE.ReplaceAll(stripped, []byte{})

	// Remove carriage returns
	text := strings.ReplaceAll(string(stripped), "\r", "")

	// Strip remaining non-printable characters
	text = StripAllControlCodes(text)

	return text
}

// StripAllControlCodes removes all non-printable characters except newlines and tabs.
func StripAllControlCodes(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) || r == '\n' || r == '\t' {
			return r
		}
		return -1
	}, s)
}

// TruncatePath truncates a path with ellipsis.
func TruncatePath(path string, limit int) string {
	if len(path) <= limit {
		return path
	}
	return "..." + path[len(path)-(limit-PreviewEllipsisLen):]
}

// truncateString truncates a string with ellipsis.
func truncateString(s string, w int) string {
	if len(s) > w && w > 3 {
		return s[:w-3] + "..."
	}
	return s
}
