package session_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/BlackGiack/kankuro/internal/session"
)

func TestSplitCommand_Simple(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{
			input:    "code /path/to/project",
			expected: []string{"code", "/path/to/project"},
		},
		{
			input:    "vim",
			expected: []string{"vim"},
		},
		{
			input:    "code --new-window /path",
			expected: []string{"code", "--new-window", "/path"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := session.SplitCommand(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d parts, got %d: %v", len(tt.expected), len(result), result)
				return
			}

			for i, part := range result {
				if part != tt.expected[i] {
					t.Errorf("Part %d: expected '%s', got '%s'", i, tt.expected[i], part)
				}
			}
		})
	}
}

func TestWorkspaceLauncherHelperProcess(t *testing.T) {
	if os.Getenv("KANKURO_LAUNCHER_HELPER") != "1" {
		return
	}
	data, err := json.Marshal(map[string]any{"args": os.Args, "dir": mustGetwd()})
	if err != nil {
		t.Fatal(err)
	}
	writeErr := os.WriteFile(os.Getenv("KANKURO_LAUNCHER_RECORD"), data, 0o600)
	if writeErr != nil {
		t.Fatal(writeErr)
	}
}

func mustGetwd() string { value, _ := os.Getwd(); return value }

func TestLaunchWorkspaceCommandExecutesInWorkspace(t *testing.T) {
	for _, tc := range []struct {
		name         string
		placeholders int
	}{{name: "quoted path and repeated placeholders", placeholders: 2}, {name: "working directory without placeholder", placeholders: 0}} {
		t.Run(tc.name, func(t *testing.T) {
			runWorkspaceLauncherCase(t, tc.placeholders)
		})
	}
}

func runWorkspaceLauncherCase(t *testing.T, placeholders int) {
	t.Helper()
	workspace := filepath.Join(t.TempDir(), "workspace with spaces")
	if err := os.Mkdir(workspace, 0o700); err != nil {
		t.Fatal(err)
	}
	record := filepath.Join(t.TempDir(), "record.json")
	t.Setenv("KANKURO_LAUNCHER_HELPER", "1")
	t.Setenv("KANKURO_LAUNCHER_RECORD", record)
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	var command strings.Builder
	fmt.Fprintf(&command, "\"%s\" -test.run=TestWorkspaceLauncherHelperProcess --", binary)
	for range placeholders {
		command.WriteString(` "$PATH"`)
	}
	if placeholders == 0 {
		command.WriteString(` ordinary-argument`)
	}
	launchErr := session.LaunchWorkspaceCommand(command.String(), workspace)
	if launchErr != nil {
		t.Fatalf("launch: %v", launchErr)
	}
	got := readWorkspaceLauncherRecord(t, record)
	abs, err := filepath.Abs(workspace)
	if err != nil {
		t.Fatal(err)
	}
	if got.Dir != abs {
		t.Errorf("child dir %q, want %q", got.Dir, abs)
	}
	wantCount := placeholders
	for _, arg := range got.Args {
		if arg == abs {
			wantCount--
		}
	}
	if wantCount != 0 {
		t.Errorf("child args %q do not contain workspace %d times", got.Args, placeholders)
	}
	if placeholders == 0 && !strings.Contains(strings.Join(got.Args, " "), "ordinary-argument") {
		t.Errorf("no-placeholder command arguments %q lost their configured argument", got.Args)
	}
}

func readWorkspaceLauncherRecord(t *testing.T, path string) struct {
	Args []string `json:"args"`
	Dir  string   `json:"dir"`
} {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	var data []byte
	var readErr error
	for time.Now().Before(deadline) {
		data, readErr = os.ReadFile(path)
		if readErr == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if readErr != nil {
		t.Fatalf("helper did not write record: %v", readErr)
	}
	var got struct {
		Args []string `json:"args"`
		Dir  string   `json:"dir"`
	}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	return got
}

func TestLaunchWorkspaceCommandErrors(t *testing.T) {
	for _, tc := range []struct {
		command, path string
		want          error
	}{
		{"", t.TempDir(), session.ErrCommandNotConfigured}, {"   ", t.TempDir(), session.ErrInvalidCommand}, {"true", "", session.ErrWorkspacePathEmpty},
		{"true", filepath.Join(t.TempDir(), "missing"), os.ErrNotExist}, {"kankuro-command-that-does-not-exist", t.TempDir(), nil},
	} {
		err := session.LaunchWorkspaceCommand(tc.command, tc.path)
		if tc.want != nil && !errors.Is(err, tc.want) {
			t.Errorf("LaunchWorkspaceCommand(%q,%q) error %v, want %v", tc.command, tc.path, err, tc.want)
		}
		if tc.want == nil && err == nil {
			t.Errorf("invalid executable unexpectedly launched")
		}
	}
}

func TestSplitCommand_Quotes(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{
			input:    `code "/path/with spaces/project"`,
			expected: []string{"code", "/path/with spaces/project"},
		},
		{
			input:    `'/path/with spaces/project'`,
			expected: []string{"/path/with spaces/project"},
		},
		{
			input:    `cmd "arg1" 'arg2' arg3`,
			expected: []string{"cmd", "arg1", "arg2", "arg3"},
		},
		{
			input:    `code "$PATH"`,
			expected: []string{"code", "$PATH"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := session.SplitCommand(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d parts, got %d: %v", len(tt.expected), len(result), result)
				return
			}

			for i, part := range result {
				if part != tt.expected[i] {
					t.Errorf("Part %d: expected '%s', got '%s'", i, tt.expected[i], part)
				}
			}
		})
	}
}

func TestSplitCommand_Empty(t *testing.T) {
	result := session.SplitCommand("")

	if len(result) != 0 {
		t.Errorf("Empty input should return empty slice, got %v", result)
	}
}

func TestSplitCommand_Whitespace(t *testing.T) {
	result := session.SplitCommand("   ")

	if len(result) != 0 {
		t.Errorf("Whitespace-only input should return empty slice, got %v", result)
	}
}

func TestSplitCommand_MultipleSpaces(t *testing.T) {
	result := session.SplitCommand("cmd   arg1    arg2")
	expected := []string{"cmd", "arg1", "arg2"}

	if len(result) != len(expected) {
		t.Errorf("Expected %d parts, got %d", len(expected), len(result))
	}
}

func TestIDELauncher_New(t *testing.T) {
	launcher := session.NewIDELauncher("code $PATH")

	if launcher == nil {
		t.Fatal("session.NewIDELauncher returned nil")
	}

	if launcher.Command() != "code $PATH" {
		t.Errorf("Expected command 'code $PATH', got '%s'", launcher.Command())
	}
}

func TestIDELauncher_Errors(t *testing.T) {
	tests := []struct {
		name    string
		command string
		path    string
		wantErr error
	}{
		{
			name:    "no command configured",
			command: "",
			path:    "/tmp",
			wantErr: session.ErrIDENotConfigured,
		},
		{
			name:    "empty path",
			command: "code $PATH",
			path:    "",
			wantErr: session.ErrWorkspacePathEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			launcher := session.NewIDELauncher(tt.command)
			err := launcher.Launch(tt.path)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestIDELauncher_PathReplacement(t *testing.T) {
	launcher := session.NewIDELauncher("code $PATH")

	expected := "code /home/user/project"
	result := strings.ReplaceAll(launcher.Command(), session.PathPlaceholder, "/home/user/project")

	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestIDELauncher_PathReplacement_Multiple(t *testing.T) {
	launcher := session.NewIDELauncher("code $PATH --args $PATH")

	expected := "code /home/user/project --args /home/user/project"
	result := strings.ReplaceAll(launcher.Command(), session.PathPlaceholder, "/home/user/project")

	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestIDEErrors(t *testing.T) {
	tests := []struct {
		err      error
		expected string
	}{
		{session.ErrIDENotConfigured, "IDE command not configured"},
		{session.ErrWorkspacePathEmpty, "workspace path is empty"},
		{session.ErrInvalidCommand, "invalid IDE command"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.err.Error() != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, tt.err.Error())
			}
		})
	}
}

func TestIDEErrors_Is(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		target error
		want   bool
	}{
		{
			name:   "same error",
			err:    session.ErrIDENotConfigured,
			target: session.ErrIDENotConfigured,
			want:   true,
		},
		{
			name:   "different error",
			err:    session.ErrIDENotConfigured,
			target: session.ErrWorkspacePathEmpty,
			want:   false,
		},
		{
			name:   "non-ide error",
			err:    session.ErrIDENotConfigured,
			target: errors.New("other error"),
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := errors.Is(tt.err, tt.target)
			if result != tt.want {
				t.Errorf("errors.Is(%v, %v) = %v, want %v", tt.err, tt.target, result, tt.want)
			}
		})
	}
}

func TestIDEErrors_Format(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		verb     rune
		expected string
	}{
		{
			name:     "default format",
			err:      session.ErrIDENotConfigured,
			verb:     's',
			expected: "IDE command not configured",
		},
		{
			name:     "quoted format",
			err:      session.ErrIDENotConfigured,
			verb:     'q',
			expected: `"IDE command not configured"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fmt.Sprintf("%"+string(tt.verb), tt.err)
			if result != tt.expected {
				t.Errorf("fmt.Sprintf(\"%%%c\", %v) = %q, want %q", tt.verb, tt.err, result, tt.expected)
			}
		})
	}
}

func TestIDELauncher_CommandWithFlags(t *testing.T) {
	launcher := session.NewIDELauncher("code --new-window --wait $PATH")

	if launcher.Command() != "code --new-window --wait $PATH" {
		t.Errorf("Expected command 'code --new-window --wait $PATH', got '%s'", launcher.Command())
	}
}

func TestSplitCommand_Tabs(t *testing.T) {
	result := session.SplitCommand("cmd\targ1\targ2")
	expected := []string{"cmd", "arg1", "arg2"}

	if len(result) != len(expected) {
		t.Errorf("Expected %d parts, got %d", len(expected), len(result))
	}

	for i, part := range result {
		if part != expected[i] {
			t.Errorf("Part %d: expected '%s', got '%s'", i, expected[i], part)
		}
	}
}

func TestSplitCommand_MixedWhitespace(t *testing.T) {
	result := session.SplitCommand("cmd arg1\targ2  arg3")
	expected := []string{"cmd", "arg1", "arg2", "arg3"}

	if len(result) != len(expected) {
		t.Errorf("Expected %d parts, got %d", len(expected), len(result))
	}

	for i, part := range result {
		if part != expected[i] {
			t.Errorf("Part %d: expected '%s', got '%s'", i, expected[i], part)
		}
	}
}
