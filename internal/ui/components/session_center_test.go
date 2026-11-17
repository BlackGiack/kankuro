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
	if strings.Contains(got, "\x1b") || !strings.Contains(got, "hello") {
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
