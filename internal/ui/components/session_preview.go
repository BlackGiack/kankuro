package components

import (
	"bytes"
	"strings"
	"unicode"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/mattn/go-runewidth"

	"kankuro/internal/config"
	"kankuro/internal/session"
	"kankuro/internal/ui/component"
	"kankuro/internal/ui/styles"
)

// Preview layout .
const (
	PreviewIDLabelPrefixLen   = 4
	PreviewCardLabelPrefixLen = 6
	PreviewEllipsisLen        = 3
	PreviewBorderLines        = 2
	PreviewMaxScrollbackBytes = 8 * 1024 // 8KB cap
	PreviewContentHPadding    = 2

	WidthBorders = 6
)

// Status badge styles.
var (
	StatusRunningStyle = lipgloss.NewStyle(). //nolint:gochecknoglobals // UI style definitions
				Foreground(styles.White).
				Background(styles.HighlightColor).
				Padding(0, 1)

	StatusWaitingStyle = lipgloss.NewStyle(). //nolint:gochecknoglobals // UI style definitions
				Foreground(styles.White).
				Background(styles.HotPink).
				Padding(0, 1)

	StatusTerminatedStyle = lipgloss.NewStyle(). //nolint:gochecknoglobals // UI style definitions
				Foreground(styles.White).
				Background(styles.Gray).
				Padding(0, 1)

	TitleStyle = lipgloss.NewStyle(). //nolint:gochecknoglobals // UI style definitions
			Bold(true).
			Foreground(styles.PrimaryColor)
)

// SessionPreview displays detailed information about a selected session.
type SessionPreview struct {
	component.Base

	sessionID           string
	snapshot            *session.Snapshot
	card                *config.IDCard
	scrollback          []byte
	puppets             []config.Puppet
	truncateLimit       int
	lastMeaningfulLines []string
	lastSessionID       string

	// Stateful parser
	parserState parserState
	escBuf      []byte
	sanitized   []byte
	lineBuf     strings.Builder
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

// SetPuppets sets the puppet catalog for rendering.
func (p *SessionPreview) SetPuppets(puppets []config.Puppet) {
	p.puppets = puppets
}

// SetSession updates the displayed session.
func (p *SessionPreview) SetSession(snap *session.Snapshot, card *config.IDCard) {
	// If the session changed, reset the internal state to avoid stale content
	if snap == nil || snap.ID != p.sessionID {
		p.lastMeaningfulLines = nil
		p.lastSessionID = p.sessionID
		p.parserState = stateNormal
		p.escBuf = nil
		p.sanitized = nil
		p.scrollback = nil
		p.lineBuf.Reset()
	}

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
	// Reset parser state for full set
	p.parserState = stateNormal
	p.escBuf = nil
	p.sanitized = nil
	p.lineBuf.Reset()

	// If the buffer is at the session manager's limit, it was likely truncated.
	// We probe the first partial line for residue to avoid misinterpreting
	// don't misinterpret mid-sequence residue as normal text.
	if len(data) >= session.ScrollbackLimit {
		if i := bytes.IndexByte(data, '\n'); i != -1 {
			firstLine := data[:i]
			if isLikelyResidue(string(firstLine)) {
				data = data[i+1:]
			}
		} else if isLikelyResidue(string(data)) {
			// Giant line/payload with no newlines that looks like residue; drop it.
			data = nil
		}
	}

	p.processData(data)
}

func isLikelyResidue(s string) bool {
	// If it's short, we don't treat it as problematic residue.
	// Legitimate hashes (up to 128 chars for SHA512) and tokens are preserved.
	const minResidueLen = 129
	if len(s) < minResidueLen {
		return false
	}

	// If it contains spaces, it's likely real human-readable text or structured code.
	if strings.Contains(s, " ") {
		return false
	}

	// If it contains "safe" symbols common in paths, URLs, and code but absent
	// in standard Base64 data (used by Kitty), it's likely legitimate content.
	// Base64 uses [A-Za-z0-9+/=]. We check for common delimiters like '.', ':', '_', '-'.
	if strings.ContainsAny(s, ".:_-") {
		return false
	}

	// If it's very long, has no spaces, and no common human/path delimiters,
	// it's very likely a truncated protocol payload.
	return true
}

// AppendScrollback appends data to the scrollback buffer.
func (p *SessionPreview) AppendScrollback(data []byte) {
	p.scrollback = append(p.scrollback, data...)
	if len(p.scrollback) > PreviewMaxScrollbackBytes {
		p.scrollback = p.scrollback[len(p.scrollback)-PreviewMaxScrollbackBytes:]
	}
	p.processData(data)
}

func (p *SessionPreview) processData(data []byte) {
	for _, b := range data {
		switch p.parserState {
		case stateNormal:
			p.parserState, p.escBuf = p.handleStateNormal(b)
		case stateESC:
			p.parserState, p.escBuf = handleStateESC(b)
		case stateCSI:
			p.parserState, p.escBuf = p.handleStateCSI(b, p.escBuf)
		case stateOSC:
			p.parserState, p.escBuf = handleStateOSC(b, p.escBuf)
		case stateDCS:
			p.parserState, p.escBuf = handleStateDCS(b, p.escBuf)
		case stateAPC:
			p.parserState, p.escBuf = handleStateAPC(b, p.escBuf)
		case stateCR:
			p.parserState, p.escBuf = p.handleStateCR(b)
		}
	}

	// Cap sanitized buffer to avoid memory leaks.
	const capLimit = PreviewMaxScrollbackBytes * 2
	if len(p.sanitized) > capLimit {
		p.sanitized = p.sanitized[len(p.sanitized)-capLimit:]
	}
}

// Clear resets the preview to empty state.
func (p *SessionPreview) Clear() {
	p.sessionID = ""
	p.snapshot = nil
	p.card = nil
	p.scrollback = nil
	p.lastMeaningfulLines = nil
	p.lastSessionID = ""
	p.parserState = stateNormal
	p.escBuf = nil
	p.sanitized = nil
	p.lineBuf.Reset()
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

	return clampToSize(
		lipgloss.NewStyle().
			Width(p.Width-WidthBorders).
			Height(p.Height).
			Align(lipgloss.Center, lipgloss.Center).
			Foreground(styles.Gray).
			Render(placeholder),
		p.Width, p.Height,
	)
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
	puppetStr := RenderAgentPuppet(puppetStyle, p.puppets)

	// Build header
	headerLines := []string{
		TitleStyle.Render(truncateString(sessionName, p.Width-WidthBorders)),
		lipgloss.JoinHorizontal(lipgloss.Center, statusBadge),
		lipgloss.NewStyle().Foreground(styles.Gray).Render(
			"ID: " + truncateString(p.snapshot.ID, p.Width-PreviewIDLabelPrefixLen-WidthBorders),
		),
		lipgloss.NewStyle().Foreground(styles.Gray).Render("WS: " + wdStr),
		lipgloss.NewStyle().Foreground(styles.Gray).Render("Started: " + startedStr),
	}

	if p.card != nil {
		headerLines = append(headerLines,
			lipgloss.NewStyle().Foreground(styles.Gray).Render(
				"Card: "+truncateString(p.card.Name, p.Width-PreviewCardLabelPrefixLen-WidthBorders),
			),
		)
	}
	headerLines = append(headerLines, "")

	header := strings.Join(headerLines, "\n")

	// Combine header with puppet
	headerWithPuppet := lipgloss.JoinHorizontal(lipgloss.Top, header, "   ", puppetStr)
	topHeight := lipgloss.Height(headerWithPuppet)
	scrollbackHeight := max(1, p.Height-topHeight)

	scrollbackContent := p.renderScrollback(scrollbackHeight)

	// Combine header and scrollback
	body := lipgloss.JoinVertical(lipgloss.Left,
		headerWithPuppet,
		scrollbackContent,
	)

	// Defensive assertion: if the body is still too tall, trim meaningful lines
	// and re-render. This handles unknown unknowns or weird lipgloss interactions.
	if lipgloss.Height(body) > p.Height && len(p.lastMeaningfulLines) > 0 {
		lines := append([]string(nil), p.lastMeaningfulLines...)
		// We stop at 1 line so that the preview is never completely empty if there
		// is at least one meaningful line. The outer MaxHeight clamp handles any
		// remaining overflow.
		for lipgloss.Height(body) > p.Height && len(lines) > 1 {
			lines = lines[1:]
			content := strings.Join(lines, "\n")
			scrollbackContent = p.renderScrollbackStyle(content, scrollbackHeight)
			body = lipgloss.JoinVertical(lipgloss.Left,
				headerWithPuppet,
				scrollbackContent,
			)
		}
	}

	// Final hard clamp: split into visual lines, cap to p.Height rows, and
	// ANSI-aware truncate each row to p.Width visible columns. This is the
	// post-sanitization guarantee that the View() output never exceeds the
	// allotted size, even if lipgloss soft-wraps or pads content unexpectedly.
	return clampToSize(body, p.Width, p.Height)
}

// ClampToSize hard-truncates a rendered string to at most maxHeight rows and
// at most maxWidth visual columns per row, preserving ANSI escape sequences
// when possible. As a last-resort safety net, if any line is still wider than
// maxWidth after ANSI-aware truncation (e.g. malformed escapes, exotic
// graphemes), it is replaced by a plain-text truncation that strips styling
// rather than risk layout overflow.
func ClampToSize(s string, maxWidth, maxHeight int) string {
	return clampToSize(s, maxWidth, maxHeight)
}

func clampToSize(s string, maxWidth, maxHeight int) string {
	if maxWidth <= 0 || maxHeight <= 0 {
		return ""
	}
	lines := strings.Split(s, "\n")
	if len(lines) > maxHeight {
		lines = lines[:maxHeight]
	}
	for i, line := range lines {
		if ansi.StringWidth(line) <= maxWidth {
			continue
		}
		clamped := ansi.Truncate(line, maxWidth, "")
		if ansi.StringWidth(clamped) > maxWidth {
			// Defensive fallback: strip ANSI and hard-truncate by visual width.
			clamped = truncateToVisibleWidth(ansi.Strip(line), maxWidth)
		}
		lines[i] = clamped
	}
	return strings.Join(lines, "\n")
}

type parserState int

const (
	stateNormal parserState = iota
	stateESC
	stateCSI
	stateOSC
	stateDCS
	stateAPC
	stateCR
)

func (p *SessionPreview) renderScrollbackStyle(content string, height int) string {
	// Under lipgloss v2, height is border-inclusive (border-box model).
	// We want the total outer box (with top border included) to be exactly `height` rows.
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(styles.Gray).
		Width(p.Width).
		Height(max(1, height)).
		MaxWidth(p.Width).
		MaxHeight(max(1, height)).
		Render(content)
}

// renderScrollback renders the scrollback buffer area.
func (p *SessionPreview) renderScrollback(height int) string {
	// Reset sticky cache if session changed
	if p.sessionID != p.lastSessionID {
		p.lastMeaningfulLines = nil
		p.lastSessionID = p.sessionID
	}

	if len(p.sanitized) == 0 && p.lineBuf.Len() == 0 && len(p.lastMeaningfulLines) == 0 {
		return p.renderNoOutput(height)
	}

	// 1. Use already sanitized text from stateful parser, combining committed lines
	// and the currently active line.
	text := string(p.sanitized) + p.lineBuf.String()

	// 2. Extract meaningful lines from sanitized text
	lines := ExtractMeaningfulLines(text, p.Width, height)

	// 3. Update sticky cache ONLY if new meaningful lines found
	if len(lines) > 0 {
		p.lastMeaningfulLines = lines
	}

	// 4. If we have NO meaningful content ever, show "no output"
	if len(p.lastMeaningfulLines) == 0 {
		return p.renderNoOutput(height)
	}

	// 5. Use cached lines (either from this pass or previous ones)
	content := strings.Join(p.lastMeaningfulLines, "\n")

	return p.renderScrollbackStyle(content, height)
}

func (p *SessionPreview) renderNoOutput(height int) string {
	// Width/Height are required here so Align(Center) works correctly.
	// Under lipgloss v2, height is border-inclusive (border-box model).
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(styles.Gray).
		Width(p.Width).
		Height(max(1, height)).
		MaxWidth(p.Width).
		MaxHeight(max(1, height)).
		Align(lipgloss.Center, lipgloss.Center).
		Foreground(styles.Gray).
		Render("(no output yet)")
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

const (
	charBEL = 0x07
	charESC = 0x1b
)

func (p *SessionPreview) handleStateNormal(b byte) (parserState, []byte) {
	switch {
	case b == charESC:
		return stateESC, []byte{b}
	case b == '\r':
		// Defer decision on \r until next byte (CRLF vs Redraw)
		return stateCR, nil
	case b == '\n':
		// Commit current line to sanitized buffer.
		p.lineBuf.WriteByte('\n')
		p.sanitized = append(p.sanitized, []byte(p.lineBuf.String())...)
		p.lineBuf.Reset()
		return stateNormal, nil
	case b == '\t' || (b >= 0x20 && b <= 0x7e) || b >= 0x80:
		p.lineBuf.WriteByte(b)
		return stateNormal, nil
	default:
		return stateNormal, nil
	}
}

func (p *SessionPreview) handleStateCR(b byte) (parserState, []byte) {
	switch {
	case b == '\n':
		// CRLF sequence: commit current line as newline.
		p.lineBuf.WriteByte('\n')
		p.sanitized = append(p.sanitized, []byte(p.lineBuf.String())...)
		p.lineBuf.Reset()
		return stateNormal, nil
	case b == '\r':
		// Repeated \r: still in CR state, treat as one overwrite.
		return stateCR, nil
	case b == charESC:
		// Redraw overwrite followed by ESC sequence
		p.lineBuf.Reset()
		return stateESC, []byte{charESC, b}
	case b == '\t' || (b >= 0x20 && b <= 0x7e) || b >= 0x80:
		// Redraw overwrite followed by printable text
		p.lineBuf.Reset()
		p.lineBuf.WriteByte(b)
		return stateNormal, nil
	default:
		// Other control chars: ignore but reset lineBuf as it's an overwrite attempt
		p.lineBuf.Reset()
		return stateNormal, nil
	}
}

func handleStateESC(b byte) (parserState, []byte) {
	escBuf := []byte{charESC, b}
	switch b {
	case '[':
		return stateCSI, escBuf
	case ']':
		return stateOSC, escBuf
	case 'P':
		return stateDCS, escBuf
	case '_':
		return stateAPC, escBuf
	default:
		return stateNormal, nil
	}
}

func (p *SessionPreview) handleStateCSI(b byte, escBuf []byte) (parserState, []byte) {
	escBuf = append(escBuf, b)
	if b >= 0x40 && b <= 0x7e {
		// CSI sequence terminated; discard. Previously the 'u' terminator was
		// translated to printable text (Kitty keyboard protocol), but that
		// produced noisy strings like "^K" / "<backspace>" in the preview, so
		// we now drop these silently.
		return stateNormal, nil
	}
	return stateCSI, escBuf
}

func handleStateOSC(b byte, escBuf []byte) (parserState, []byte) {
	escBuf = append(escBuf, b)
	if b == charBEL {
		return stateNormal, nil
	}
	if b == '\\' && len(escBuf) >= 2 && escBuf[len(escBuf)-2] == charESC {
		return stateNormal, nil
	}
	return stateOSC, escBuf
}

func handleStateDCS(b byte, escBuf []byte) (parserState, []byte) {
	escBuf = append(escBuf, b)
	if b == '\\' && len(escBuf) >= 2 && escBuf[len(escBuf)-2] == charESC {
		return stateNormal, nil
	}
	return stateDCS, escBuf
}

func handleStateAPC(b byte, escBuf []byte) (parserState, []byte) {
	escBuf = append(escBuf, b)
	// APC (including Kitty graphics G payloads) are swallowed wholesale until
	// terminator. We intentionally do not surface "[image]" or any other
	// placeholder text — those previously caused noisy preview output and
	// risk widening lines past the pane.
	if b == charBEL {
		return stateNormal, nil
	}
	if b == '\\' && len(escBuf) >= 2 && escBuf[len(escBuf)-2] == charESC {
		return stateNormal, nil
	}
	return stateAPC, escBuf
}

// ExtractMeaningfulLines processes sanitized text into a window of meaningful preview lines.
func ExtractMeaningfulLines(text string, width, height int) []string {
	rawLines := strings.Split(text, "\n")
	var meaningful []string

	maxLines := max(1, height-PreviewBorderLines)

	for _, line := range rawLines {
		// Drop any non-printable bytes/runes that the streaming parser may have
		// passed through (e.g. raw >=0x80 bytes that aren't valid UTF-8). This
		// is the same conservative filter the pre-c1df769 sanitizer applied
		// and prevents the terminal from rendering replacement-char rectangles
		// or other unpredictable glyphs.
		trimmed := strings.TrimSpace(StripAllControlCodes(line))
		if trimmed == "" {
			continue
		}

		if !isMeaningful(trimmed) {
			continue
		}

		// Width truncation: show the TAIL of long lines in the preview
		// so that incremental updates to a single long line are visible.
		if width > 5 && lipgloss.Width(trimmed) > width-PreviewContentHPadding {
			meaningful = append(meaningful, truncateToVisibleWidth(trimmed, width-PreviewContentHPadding))
		} else {
			meaningful = append(meaningful, trimmed)
		}
	}

	if len(meaningful) > maxLines {
		meaningful = meaningful[len(meaningful)-maxLines:]
	}

	return meaningful
}

func isMeaningful(s string) bool {
	// Heuristic: requires at least 2 letters to be considered meaningful.
	// If digits are present, it requires more letters than digits to avoid
	// terminal residue like "5;75HI".
	const minLetters = 2
	letterCount := 0
	digitCount := 0

	for _, r := range s {
		if unicode.IsLetter(r) {
			letterCount++
		} else if unicode.IsDigit(r) {
			digitCount++
		}
	}

	return letterCount >= minLetters && letterCount > digitCount
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

// TruncatePath truncates a path with ellipsis, keeping the tail (the most
// specific path segments) visible.
func TruncatePath(path string, limit int) string {
	return truncateToVisibleWidth(path, limit)
}

// truncateString truncates a string with ellipsis, keeping the head visible.
func truncateString(s string, w int) string {
	return truncateToVisibleWidthLeft(s, w)
}

// truncateToVisibleWidth trims s so that lipgloss.Width(result) <= maxWidth,
// keeping the tail (rightmost runes) and prefixing "..." when truncated.
// Measures and slices by rune, not byte, so multibyte glyphs (CJK, emoji,
// accented characters) are never cut mid-rune.
func truncateToVisibleWidth(s string, maxWidth int) string {
	if runewidth.StringWidth(s) <= maxWidth {
		return s
	}
	if maxWidth <= PreviewEllipsisLen {
		return s
	}

	runes := []rune(s)
	budget := maxWidth - PreviewEllipsisLen
	w := 0
	i := len(runes)

	// Walk from the tail computing width
	for i > 0 {
		rw := runewidth.RuneWidth(runes[i-1])
		if w+rw > budget {
			break
		}
		w += rw
		i--
	}

	return "..." + string(runes[i:])
}

// truncateToVisibleWidthLeft trims s so that lipgloss.Width(result) <=
// maxWidth, keeping the head (leftmost runes) and suffixing "..." when
// truncated. Measures and slices by rune, not byte, so multibyte glyphs are
// never cut mid-rune.
func truncateToVisibleWidthLeft(s string, maxWidth int) string {
	if runewidth.StringWidth(s) <= maxWidth {
		return s
	}
	if maxWidth <= PreviewEllipsisLen {
		return s
	}

	runes := []rune(s)
	budget := maxWidth - PreviewEllipsisLen
	w := 0
	i := 0

	for i < len(runes) {
		rw := runewidth.RuneWidth(runes[i])
		if w+rw > budget {
			break
		}
		w += rw
		i++
	}

	return string(runes[:i]) + "..."
}
