package components

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	tea "charm.land/bubbletea/v2"
	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/ui/component"
	"github.com/BlackGiack/kankuro/internal/ui/models"
	"github.com/charmbracelet/x/ansi"
)

type SessionCenter struct {
	component.Base

	sessionID string
	data      models.SessionCenterData
}

func NewSessionCenter() *SessionCenter                               { return &SessionCenter{} }
func (p *SessionCenter) Init() tea.Cmd                               { return nil }
func (p *SessionCenter) Update(_ tea.Msg) (component.Model, tea.Cmd) { return p, nil }
func (p *SessionCenter) SessionID() string                           { return p.sessionID }
func (p *SessionCenter) SetSession(row *models.SessionRow) {
	id := ""
	if row != nil {
		id = row.ID
	}
	if id != p.sessionID {
		p.sessionID = id
		p.data = models.SessionCenterData{SessionID: id, Loading: id != ""}
	}
}
func (p *SessionCenter) Clear() { p.SetSession(nil) }
func (p *SessionCenter) SetData(data models.SessionCenterData) {
	if data.SessionID == p.sessionID {
		p.data = models.CloneSessionCenterData(data)
	}
}
func (p *SessionCenter) ReportLimit() int {
	if p.Width <= 0 || p.Height <= 0 {
		return 0
	}
	return max(1, p.Height-2-len(p.data.Warnings))
}
func centerSafeText(text string) string {
	text = ansi.Strip(text)
	return strings.Map(func(r rune) rune {
		if r == '\n' {
			return r
		}
		if r == '\t' || r == '\r' {
			return ' '
		}
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, text)
}

// centerSafePath keeps each filesystem name within its own rendered row.
func centerSafePath(text string) string {
	text = ansi.Strip(text)
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) || unicode.IsSpace(r) {
			return ' '
		}
		return r
	}, text)
}

//nolint:gocognit,mnd // Rendering preserves bounded, warning-aware panel layout.
func (p *SessionCenter) View() string {
	if p.Width <= 0 || p.Height <= 0 {
		return ""
	}
	var lines []string
	add := func(s string) {
		if len(lines) >= p.Height {
			return
		}
		wrapped := ansi.Hardwrap(centerSafeText(s), p.Width, true)
		lines = append(lines, strings.Split(wrapped, "\n")...)
	}
	if p.sessionID == "" {
		add("Select a session")
		return ClampToSize(strings.Join(lines, "\n"), p.Width, p.Height)
	}
	if p.data.Stale {
		add("[stale]")
	}
	for _, warning := range p.data.Warnings {
		add("Warning: " + centerSafePath(warning))
	}
	// Keep at least one row for the newest report when warnings are long.
	if len(p.data.Reports) > 0 && len(lines) >= p.Height {
		lines = lines[:max(0, p.Height-1)]
	} else if p.data.Workspace != nil && len(lines) > max(0, p.Height-2) {
		lines = lines[:max(0, p.Height-2)]
	}
	switch {
	case len(p.data.Reports) > 0:
		for _, report := range p.data.Reports {
			if p.Height-len(lines) >= 2 {
				stamp := report.CreatedAt.Local().Format("15:04")
				if p.Width >= 20 {
					stamp = report.CreatedAt.Local().Format("2006-01-02 15:04")
				}
				header := stamp + " [" + report.Kind + "]"
				if p.Height-len(lines) == 2 && ansi.StringWidth(centerSafeText(header)) > p.Width {
					header = report.CreatedAt.Local().Format("15:04") + " [" + report.Kind + "]"
					header = TruncateHead(centerSafeText(header), p.Width)
				}
				add(header)
				if len(lines) >= p.Height {
					lines = lines[:p.Height-1]
				}
			}
			add(report.Text)
			if len(lines) >= p.Height {
				break
			}
		}
		if p.data.Omitted > 0 && len(lines) < p.Height {
			add(fmt.Sprintf("… %d older reports omitted", p.data.Omitted))
		}
	case p.data.Workspace != nil:
		p.renderWorkspace(p.data.Workspace, add)
	case p.data.Loading:
		add("Loading…")
	default:
		add("No workspace information")
	}
	return ClampToSize(strings.Join(lines, "\n"), p.Width, p.Height)
}

//nolint:gocognit,nestif // Workspace rows preserve detailed metadata at small widths.
func (p *SessionCenter) renderWorkspace(w *contracts.SessionWorkspaceResult, add func(string)) {
	if w.Kind == contracts.WorkspaceGit {
		add(TruncateHead("Git: "+centerSafePath(w.RepositoryRoot), p.Width))
		if w.Truncated {
			add("… Git status truncated")
		}
		if len(w.GitEntries) == 0 {
			add("Working tree clean")
		}
		for _, e := range w.GitEntries {
			s := fmt.Sprintf("%c%c %s", e.IndexStatus, e.WorktreeStatus, centerSafePath(e.Path))
			if e.OriginalPath != "" {
				s += " ← " + centerSafePath(e.OriginalPath)
			}
			add(s)
		}
	} else {
		add(TruncateHead("Directory: "+centerSafePath(w.WorkingDir), p.Width))
		if w.Truncated {
			add("… directory sample truncated")
		}
		for _, e := range w.Entries {
			owner, group, links := "-", "-", "-"
			if e.Owner != nil {
				owner = strconv.FormatUint(uint64(*e.Owner), 10)
			}
			if e.Group != nil {
				group = strconv.FormatUint(uint64(*e.Group), 10)
			}
			if e.LinkCount != nil {
				links = strconv.FormatUint(*e.LinkCount, 10)
			}
			s := fmt.Sprintf(
				"%s  %s %s %s %s %d %s",
				centerSafePath(e.Name),
				centerMode(e.Mode),
				links,
				owner,
				group,
				e.Size,
				e.ModifiedAt.Local().Format("2006-01-02 15:04"),
			)
			if e.SymlinkTarget != nil {
				s += " -> " + centerSafePath(*e.SymlinkTarget)
			}
			add(s)
		}
	}
}

// centerMode formats the stable mode bits supplied by the workspace contract.
func centerMode(mode uint32) string {
	const (
		directory            = uint32(1 << 31)
		symlink              = uint32(1 << 27)
		device               = uint32(1 << 26)
		namedPipe            = uint32(1 << 25)
		socket               = uint32(1 << 24)
		setUID               = uint32(1 << 23)
		setGID               = uint32(1 << 22)
		charDevice           = uint32(1 << 21)
		sticky               = uint32(1 << 20)
		highestPermissionBit = 8
	)
	kind := byte('-')
	switch {
	case mode&directory != 0:
		kind = 'd'
	case mode&symlink != 0:
		kind = 'l'
	case mode&namedPipe != 0:
		kind = 'p'
	case mode&socket != 0:
		kind = 's'
	case mode&charDevice != 0:
		kind = 'c'
	case mode&device != 0:
		kind = 'b'
	}
	bits := []byte("---------")
	for i := range 9 {
		if mode&(1<<(highestPermissionBit-i)) != 0 {
			bits[i] = "rwx"[i%3]
		}
	}
	for _, item := range []struct {
		mask         uint32
		index        int
		lower, upper byte
	}{{setUID, 2, 's', 'S'}, {setGID, 5, 's', 'S'}, {sticky, 8, 't', 'T'}} {
		if mode&item.mask != 0 {
			if bits[item.index] == 'x' {
				bits[item.index] = item.lower
			} else {
				bits[item.index] = item.upper
			}
		}
	}
	return string(kind) + string(bits)
}
