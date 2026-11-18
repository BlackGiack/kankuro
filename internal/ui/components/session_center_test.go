package components_test

import (
	"strings"
	"testing"
	"time"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/ui/components"
	"github.com/BlackGiack/kankuro/internal/ui/models"
	"github.com/charmbracelet/x/ansi"
)

func TestSessionCenterReportsFitAndSanitize(t *testing.T) {
	p := components.NewSessionCenter()
	p.SetSize(20, 4)
	p.SetSession(&models.SessionRow{ID: "one"})
	p.SetData(
		models.SessionCenterData{
			SessionID: "one",
			Reports: []contracts.SessionReport{
				{CreatedAt: time.Unix(0, 0), Kind: "progress", Text: "\x1b[31mhello\x1b[0m 世界\nsecond"},
				{Text: "older"},
			},
		},
	)
	got := p.View()
	if strings.Contains(got, "\x1b[31m") || !strings.Contains(got, "hello") {
		t.Fatalf("unsafe or missing report: %q", got)
	}
	lines := strings.Split(got, "\n")
	if len(lines) > 4 {
		t.Fatalf("height %d", len(lines))
	}
	for _, line := range lines {
		if ansi.StringWidth(line) > 20 {
			t.Fatalf("wide line: %q", line)
		}
	}
	p.SetSession(&models.SessionRow{ID: "two"})
	if strings.Contains(p.View(), "hello") {
		t.Fatal("old session content survived selection")
	}
	p.SetSize(0, 0)
	if p.View() != "" || p.ReportLimit() != 0 {
		t.Fatal("zero size not safe")
	}
}

func TestSessionCenterKeepsSessionHeaderAboveBody(t *testing.T) {
	p := components.NewSessionCenter()
	p.SetSize(50, 20)
	row := models.SessionRow{
		ID: "session-id", Name: "Selected session", Status: "running",
		WorkingDir: "/work/project", StartedAt: time.Date(2026, 1, 1, 13, 4, 5, 0, time.UTC),
		CardName: "Active card", PuppetName: "custom", PuppetLines: []string{"puppet-artwork"},
	}
	p.SetSession(&row)
	row.PuppetLines[0] = "other-artwork"
	p.SetData(models.SessionCenterData{
		SessionID: row.ID, Workspace: &contracts.SessionWorkspaceResult{Kind: contracts.WorkspaceDirectory, WorkingDir: row.WorkingDir},
	})
	view := ansi.Strip(p.View())
	lines := strings.Split(view, "\n")
	wantLines := []string{"Selected session", "Running", "ID: session-id", "WS: /work/project", "Started: 13:04:05", "Card: Active card", "", strings.Repeat("─", 50), "Directory: /work/project"}
	if len(lines) != len(wantLines) {
		t.Fatalf("unexpected header and body rows:\n%s", view)
	}
	for i, want := range wantLines {
		line := strings.TrimSpace(lines[i])
		if !strings.HasPrefix(line, want) || (want == "" && line != "") {
			t.Fatalf("row %d should start with %q in:\n%s", i, want, view)
		}
	}
	if !strings.Contains(lines[0], "puppet-artwork") || strings.Contains(view, "other-artwork") {
		t.Fatalf("session artwork missing or aliased:\n%s", view)
	}

	row.Name, row.Status = "Renamed session", "waiting"
	p.SetSession(&row)
	view = ansi.Strip(p.View())
	if !strings.Contains(view, "Renamed session") || !strings.Contains(view, "Waiting") ||
		!strings.Contains(view, "other-artwork") || !strings.Contains(view, "Directory: /work/project") {
		t.Fatalf("same-session refresh lost metadata or body:\n%s", view)
	}
	p.SetData(models.SessionCenterData{SessionID: row.ID, Reports: []contracts.SessionReport{{Text: "First authored report", Kind: "progress"}}})
	view = ansi.Strip(p.View())
	if !strings.Contains(view, "Renamed session") || !strings.Contains(view, "First authored report") || strings.Contains(view, "Directory:") {
		t.Fatalf("report did not take over the body beneath the header:\n%s", view)
	}
	p.SetSession(&models.SessionRow{ID: "next-session", Status: "terminated"})
	p.SetData(models.SessionCenterData{SessionID: row.ID, Reports: []contracts.SessionReport{{Text: "Late report"}}})
	view = ansi.Strip(p.View())
	if !strings.Contains(view, "next-session") || !strings.Contains(view, "Terminated") || !strings.Contains(view, "Loading…") ||
		strings.Contains(view, "Renamed session") || strings.Contains(view, "report") {
		t.Fatalf("new selection retained the previous session:\n%s", view)
	}
}

func TestSessionCenterReservesSpaceForBody(t *testing.T) {
	p := components.NewSessionCenter()
	row := models.SessionRow{
		ID: "session-id", Name: strings.Repeat("界", 24), Status: "running",
		WorkingDir: "/long/workspace/path/project", CardName: "card",
		PuppetName: "custom", PuppetLines: []string{"puppet-artwork"},
	}
	p.SetSession(&row)
	p.SetData(models.SessionCenterData{SessionID: row.ID, Reports: []contracts.SessionReport{{Text: "report text"}}})
	for _, tc := range []struct {
		width, height, limit int
		art                  bool
	}{{50, 20, 10, true}, {20, 16, 6, false}, {50, 4, 1, false}, {20, 2, 1, false}, {50, 1, 1, false}} {
		p.SetSize(tc.width, tc.height)
		view := ansi.Strip(p.View())
		if strings.Contains(view, "puppet-artwork") != tc.art || !strings.Contains(view, "report text") {
			t.Fatalf("%dx%d rendered unexpected header or body:\n%s", tc.width, tc.height, view)
		}
		if got := p.ReportLimit(); got != tc.limit {
			t.Fatalf("%dx%d report limit = %d, want %d", tc.width, tc.height, got, tc.limit)
		}
		lines := strings.Split(view, "\n")
		if len(lines) > tc.height {
			t.Fatalf("%dx%d rendered %d rows", tc.width, tc.height, len(lines))
		}
		for _, line := range lines {
			if ansi.StringWidth(line) > tc.width {
				t.Fatalf("%dx%d rendered wide line %q", tc.width, tc.height, line)
			}
		}
		if tc.width == 20 && tc.height == 16 && !strings.Contains(view, "project") {
			t.Fatalf("workspace tail missing from:\n%s", view)
		}
	}
}

//nolint:gocognit // The table checks several distinct rendering boundaries.
func TestSessionCenterWorkspaceNamesCannotForgeRows(t *testing.T) {
	target := "target\nWarning: forged\t\x01end"
	cases := []struct {
		name      string
		workspace contracts.SessionWorkspaceResult
		want      []string
	}{
		{
			name: "git paths",
			workspace: contracts.SessionWorkspaceResult{
				Kind:           contracts.WorkspaceGit,
				RepositoryRoot: "/repo",
				GitEntries: []contracts.GitEntry{
					{IndexStatus: '?', WorktreeStatus: '?', Path: "actual\n?? forged\t\x01end", OriginalPath: "old\nWarning: forged"},
				},
			},
			want: []string{"?? actual ?? forged", "old Warning: forged"},
		},
		{
			name: "directory name and symlink",
			workspace: contracts.SessionWorkspaceResult{
				Kind:       contracts.WorkspaceDirectory,
				WorkingDir: "/workspace",
				Entries:    []contracts.DirectoryEntry{{Name: "link\n?? forged\t\x01end", SymlinkTarget: &target}},
			},
			want: []string{"link ?? forged", "target Warning: forged"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := components.NewSessionCenter()
			p.SetSize(200, 10)
			p.SetSession(&models.SessionRow{ID: "s"})
			p.SetData(models.SessionCenterData{SessionID: "s", Workspace: &tc.workspace})
			got := p.View()
			if strings.ContainsAny(got, "\t\x01") {
				t.Fatalf("control character leaked: %q", got)
			}
			for _, want := range tc.want {
				if !strings.Contains(got, want) {
					t.Errorf("missing sanitized %q in %q", want, got)
				}
			}
			for line := range strings.SplitSeq(got, "\n") {
				if strings.HasPrefix(line, "?? forged") || strings.HasPrefix(line, "Warning: forged") {
					t.Errorf("filename forged a row: %q", line)
				}
				if ansi.StringWidth(line) > 200 {
					t.Errorf("line overflow: %q", line)
				}
			}
		})
	}
}

func TestSessionCenterPreservesMultilineReport(t *testing.T) {
	p := components.NewSessionCenter()
	p.SetSize(80, 8)
	p.SetSession(&models.SessionRow{ID: "s"})
	p.SetData(
		models.SessionCenterData{SessionID: "s", Reports: []contracts.SessionReport{{Text: "first line\nsecond line", Kind: "progress"}}},
	)
	if got := p.View(); !strings.Contains(got, "first line\nsecond line") {
		t.Fatalf("report line breaks lost: %q", got)
	}
}

func TestSessionCenterReportHeaderKeepsKindWhenItFits(t *testing.T) {
	for _, tc := range []struct {
		width    int
		wantTime string
	}{{80, "2026-11-16 12:34"}, {20, "12:34"}} {
		p := components.NewSessionCenter()
		p.SetSize(tc.width, 2)
		p.SetSession(&models.SessionRow{ID: "s"})
		p.SetData(
			models.SessionCenterData{
				SessionID: "s",
				Reports: []contracts.SessionReport{
					{Kind: "blocked", Text: "Need input", CreatedAt: time.Date(2026, 11, 16, 12, 34, 0, 0, time.Local)},
				},
			},
		)
		got := p.View()
		if !strings.Contains(got, tc.wantTime) || !strings.Contains(got, "blocked") || !strings.Contains(got, "Need input") {
			t.Fatalf("width %d lost metadata or content: %q", tc.width, got)
		}
	}
}
