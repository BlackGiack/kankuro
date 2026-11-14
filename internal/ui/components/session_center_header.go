package components

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/BlackGiack/kankuro/internal/ui/models"
	"github.com/BlackGiack/kankuro/internal/ui/styles"
	"github.com/charmbracelet/x/ansi"
)

const (
	centerMinBodyRows  = 2
	centerPuppetGap    = 3
	centerMinTextWidth = 24
)

var (
	centerSessionTitle = lipgloss.NewStyle().Bold(true).Foreground(styles.PrimaryColor) //nolint:gochecknoglobals // UI style definition.
	centerSessionField = lipgloss.NewStyle().Foreground(styles.Gray)                    //nolint:gochecknoglobals // UI style definition.
	centerStatusStyle  = lipgloss.NewStyle().Foreground(styles.White).Padding(0, 1)     //nolint:gochecknoglobals // UI style definition.
)

func (p *SessionCenter) headerRows() []string {
	maxHeight := p.Height - centerMinBodyRows
	if p.session.ID == "" || maxHeight <= 0 {
		return nil
	}

	art := RenderPuppetArtwork(p.session.PuppetName, p.session.PuppetBuiltin, p.session.PuppetLines)
	artRows := strings.Split(art, "\n")
	for i, line := range artRows {
		artRows[i] = DisplayText(line)
	}
	art = strings.Join(artRows, "\n")
	artWidth := lipgloss.Width(art)
	showArt := artWidth > 0 && p.Width >= artWidth+centerPuppetGap+centerMinTextWidth

	fields := sessionHeaderFields(p.session, p.Width)
	// Keep the original blank row and divider when the full header fits.
	if len(fields)+2 > maxHeight || (showArt && lipgloss.Height(art)+1 > maxHeight) {
		return fields[:min(len(fields), maxHeight)]
	}
	if showArt {
		textWidth := p.Width - artWidth - centerPuppetGap
		artFields := append(sessionHeaderFields(p.session, textWidth), "")
		text := lipgloss.NewStyle().Width(textWidth).Render(strings.Join(artFields, "\n"))
		top := lipgloss.JoinHorizontal(lipgloss.Top, text, strings.Repeat(" ", centerPuppetGap), art)
		fields = strings.Split(top, "\n")
	} else {
		fields = append(fields, "")
	}
	return append(fields, centerSessionField.Render(strings.Repeat(lipgloss.NormalBorder().Top, p.Width)))
}

func sessionHeaderFields(row models.SessionRow, width int) []string {
	name := row.Name
	if name == "" {
		name = row.ID
	}
	started := "N/A"
	if !row.StartedAt.IsZero() {
		started = row.StartedAt.Format("15:04:05")
	}
	lines := []string{
		centerSessionTitle.Render(TruncateHead(DisplayText(name), width)),
		ClampToSize(sessionStatusBadge(row.Status), width, 1),
		sessionHeaderField("ID: ", row.ID, width, false),
		sessionHeaderField("WS: ", row.WorkingDir, width, true),
		sessionHeaderField("Started: ", started, width, false),
	}
	if row.CardName != "" {
		lines = append(lines, sessionHeaderField("Card: ", row.CardName, width, false))
	}
	return lines
}

func sessionHeaderField(label, value string, width int, keepTail bool) string {
	valueWidth := max(0, width-ansi.StringWidth(label))
	value = DisplayText(value)
	if keepTail {
		value = TruncateTail(value, valueWidth)
	} else {
		value = TruncateHead(value, valueWidth)
	}
	return centerSessionField.Render(TruncateHead(label+value, width))
}

func sessionStatusBadge(status string) string {
	switch status {
	case "running":
		return centerStatusStyle.Background(styles.HighlightColor).Render("Running")
	case waitingStatus:
		return centerStatusStyle.Background(styles.HotPink).Render("Waiting")
	case "terminated":
		return centerStatusStyle.Background(styles.Gray).Render("Terminated")
	default:
		return centerStatusStyle.Background(styles.Gray).Render("Unknown")
	}
}
