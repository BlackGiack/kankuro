package ui

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/charmbracelet/lipgloss"

	"kankuro/internal/ascii"
	"kankuro/internal/ascii/characters"
	"kankuro/internal/config"
	"kankuro/internal/session"
)

const (
	previewIDLabelPrefixLen   = 4
	previewCardLabelPrefixLen = 6
	previewEllipsisLen        = 3
	previewBorderLines        = 2
	previewMaxScrollbackBytes = 8 * 1024

	headerHeight = 2 // title + margin
	footerHeight = 1

	widthBorders = 6 // nb of borders in display
)

var (
	ansiEscapeRE   = regexp.MustCompile(`\x1b\[[0-9;]*[A-Za-z]`)
	controlCharRE  = regexp.MustCompile(`[\x00-\x08\x0b\x0c\x0e-\x1f\x7f]`)
	kittyProtcolRE = regexp.MustCompile(`(?:\x1b_|\x1b\[|\x1b\]|\x1bP).*?(\x1b\\|[\x07\x40-\x7e])`)
)

func StripAllControlCodes(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) || r == '\n' || r == '\t' {
			return r
		}
		return -1
	}, s)
}

func renderSessionPreview(a *App, previewWidth int) string {
	if a.sessionView.PreviewSessionID == "" {
		placeholder := `
   No session
   selected

   Select a session
   from the list to
   see a preview here.
`
		return lipgloss.NewStyle().
			Width(previewWidth-widthBorders).
			Height(a.sessionsList.Height()+previewBorderLines).
			Align(lipgloss.Center, lipgloss.Center).
			Foreground(gray).
			Render(placeholder)
	}

	snap, ok := findSessionSnapshot(a)
	if !ok {
		return lipgloss.NewStyle().
			Width(previewWidth-widthBorders).
			Height(a.sessionsList.Height()+previewBorderLines).
			Align(lipgloss.Center, lipgloss.Center).
			Foreground(gray).
			Render("Session not found")
	}

	card := findCardForSession(a, snap.CardID)

	statusBadge := renderStatusBadge(snap.Status)

	sessionName := snap.Name
	if sessionName == "" {
		sessionName = snap.ID
	}

	wdStr := truncatePath(snap.WorkingDir, pathTruncateLimit)
	startedStr := snap.StartedAt.Format("15:04:05")

	puppetStyle := defaultPuppet
	if card != nil && card.Puppet != "" {
		puppetStyle = card.Puppet
	}
	puppet := characters.NewPuppet(func(cfg *ascii.Config) { cfg.Set("style", puppetStyle) })
	w, h := puppet.Bounds()
	canvas := ascii.NewCanvas(w+previewWidth/6, h)
	puppet.Render(canvas, previewWidth/6, 0) //nolint:mnd // a sixth of the preview size
	puppetStr := canvas.String()

	headerLines := []string{
		titleStyle.Render(truncate(sessionName, previewWidth-widthBorders)),
		lipgloss.JoinHorizontal(lipgloss.Center, statusBadge),
		lipgloss.NewStyle().Foreground(gray).Render("ID: " + truncate(snap.ID, previewWidth-previewIDLabelPrefixLen-widthBorders)),
		lipgloss.NewStyle().Foreground(gray).Render("WS: " + wdStr),
		lipgloss.NewStyle().Foreground(gray).Render("Started: " + startedStr),
	}
	if card != nil {
		headerLines = append(headerLines,
			lipgloss.NewStyle().Foreground(gray).Render("Card: "+truncate(card.Name, previewWidth-previewCardLabelPrefixLen-widthBorders)),
		)
	}
	headerLines = append(headerLines, "")

	header := strings.Join(headerLines, "\n")

	scrollbackContentHeigth := a.sessionsList.Height() - previewBorderLines - headerHeight - footerHeight
	scrollbackContent := renderScrollback(a, previewWidth, scrollbackContentHeigth)

	header = lipgloss.JoinHorizontal(lipgloss.Center, header, puppetStr)

	body := lipgloss.JoinVertical(lipgloss.Left,
		header,
		scrollbackContent,
	)

	return lipgloss.NewStyle().Render(body)
}

func renderStatusBadge(status session.Status) string {
	switch status {
	case session.StatusRunning:
		return statusRunningStyle.Render("Running")
	case session.StatusWaiting:
		return statusWaitingStyle.Render("Waiting")
	case session.StatusTerminated:
		return statusTerminatedStyle.Render("Terminated")
	default:
		return statusTerminatedStyle.Render("Unknown")
	}
}

func renderScrollback(a *App, width, height int) string {
	scrollback := a.sessionView.PreviewScrollback
	if len(scrollback) == 0 {
		return lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(gray).
			Width(width).
			Height(height-1).
			Align(lipgloss.Center, lipgloss.Center).
			Foreground(gray).
			Render("(no output yet)")
	}

	// Cap scrollback size
	if len(scrollback) > previewMaxScrollbackBytes {
		scrollback = scrollback[len(scrollback)-previewMaxScrollbackBytes:]
	}

	// Strip ANSI sequences and control characters
	stripped := ansiEscapeRE.ReplaceAll(scrollback, []byte{})
	stripped = controlCharRE.ReplaceAll(stripped, []byte{})
	stripped = kittyProtcolRE.ReplaceAll(stripped, []byte{})
	text := strings.ReplaceAll(string(stripped), "\r", "")
	text = StripAllControlCodes(text)

	lines := strings.Split(text, "\n")
	maxLines := max(1, height-previewBorderLines)
	if len(lines) > maxLines {
		lines = lines[len(lines)-maxLines:]
	}

	var truncated []string
	for _, line := range lines {
		clean := strings.TrimSpace(line)
		if clean == "" {
			continue
		}
		if len(clean) > width-2 {
			truncated = append(truncated, clean[:width-5]+"...")
		} else {
			truncated = append(truncated, clean)
		}
	}

	if len(truncated) == 0 {
		return lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(gray).
			Width(width).
			Height(height-1).
			Align(lipgloss.Center, lipgloss.Center).
			Foreground(gray).
			Render("(no output yet)")
	}

	content := strings.Join(truncated, "\n")

	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(gray).
		Width(width).
		Height(height - 1).
		Render(content)
}

func truncatePath(path string, limit int) string {
	if len(path) <= limit {
		return path
	}
	return "..." + path[len(path)-(limit-previewEllipsisLen):]
}

func findSessionSnapshot(a *App) (session.Snapshot, bool) {
	for _, s := range a.sessions {
		if s.ID == a.sessionView.PreviewSessionID {
			return s, true
		}
	}
	// Also check manager directly in case session isn't in a.sessions yet
	snapshots := a.session.List()
	for _, s := range snapshots {
		if s.ID == a.sessionView.PreviewSessionID {
			return s, true
		}
	}
	return session.Snapshot{}, false
}

func findCardForSession(a *App, cardID string) *config.IDCard {
	for i := range a.cards {
		if a.cards[i].ID == cardID {
			return &a.cards[i]
		}
	}
	return nil
}
