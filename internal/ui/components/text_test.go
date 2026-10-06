package components_test

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"

	"kankuro/internal/ui/components"
)

func truncateInputs() []string {
	return []string{
		"",
		"short",
		"/home/user/very/long/path/to/some/project",
		"/家/项目/很长的路径/名字/文件夹/子目录",
		"/home/usér/proje\u0301ct/nested/deep/path",
		"/home/👍🎉🚀/project/deep/nested/path",
		"👨‍👩‍👧‍👦 family ❤️ heart 🇮🇹 flag",
		"名前名前名前名前名前名前",
	}
}

func TestTruncateFitsAtEveryWidth(t *testing.T) {
	fns := map[string]func(string, int) string{
		"TruncateHead": components.TruncateHead,
		"TruncateTail": components.TruncateTail,
	}
	for name, fn := range fns {
		for _, in := range truncateInputs() {
			for w := -1; w <= ansi.StringWidth(in)+2; w++ {
				got := fn(in, w)
				if !utf8.ValidString(got) {
					t.Fatalf("%s(%q, %d) = %q: invalid UTF-8", name, in, w, got)
				}
				if gw := ansi.StringWidth(got); gw > max(w, 0) {
					t.Fatalf("%s(%q, %d) = %q: width %d exceeds %d", name, in, w, got, gw, w)
				}
				if ansi.StringWidth(in) <= w && got != in {
					t.Fatalf("%s(%q, %d) = %q: text that fits was changed", name, in, w, got)
				}
			}
		}
	}
}

func TestTruncateKeepsTheRightEnd(t *testing.T) {
	in := "/home/user/very/long/path/to/some/project"
	if got, want := components.TruncateTail(in, 20), "...h/to/some/project"; got != want {
		t.Errorf("TruncateTail = %q, want %q", got, want)
	}
	if got, want := components.TruncateHead(in, 20), "/home/user/very/l..."; got != want {
		t.Errorf("TruncateHead = %q, want %q", got, want)
	}
	if got := components.TruncatePath(in, 20); got != components.TruncateTail(in, 20) {
		t.Errorf("TruncatePath = %q, want TruncateTail's result", got)
	}
}

func TestTruncateBelowEllipsisWidth(t *testing.T) {
	for _, w := range []int{1, 2, 3} {
		if got := components.TruncateHead("abcdefgh", w); got != "abcdefgh"[:w] {
			t.Errorf("TruncateHead w=%d = %q", w, got)
		}
		if got := components.TruncateTail("abcdefgh", w); got != "abcdefgh"[8-w:] {
			t.Errorf("TruncateTail w=%d = %q", w, got)
		}
	}
	if got := components.TruncateHead("abc", 0); got != "" {
		t.Errorf("w=0 should give an empty string, got %q", got)
	}
}

func TestDisplayText(t *testing.T) {
	tests := []struct{ name, in, want string }{
		{"plain", "my session", "my session"},
		{"tab", "a\tb", "a b"},
		{"newlines", "a\nb\r\nc", "a b  c"},
		{"control bytes dropped", "a\x00b\x07c\x1fd\x7fe", "abcde"},
		{"ansi stripped", "\x1b[31mred\x1b[0m text", "red text"},
		{"osc stripped", "\x1b]0;title\x07name", "name"},
		{"unicode kept", "名前 👍", "名前 👍"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := components.DisplayText(tt.in)
			if got != tt.want {
				t.Errorf("DisplayText(%q) = %q, want %q", tt.in, got, tt.want)
			}
			if strings.ContainsAny(got, "\t\n\r\x1b") {
				t.Errorf("DisplayText(%q) left a control character in %q", tt.in, got)
			}
		})
	}
}
