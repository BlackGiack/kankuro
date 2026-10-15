package session

import (
	"errors"
	"fmt"
	"strings"
)

const (
	// PathPlaceholder is the placeholder for workspace path in IDE commands.
	PathPlaceholder = "$PATH"
)

// IDELauncher provides IDE launching functionality.
type IDELauncher struct {
	command string
}

// NewIDELauncher creates a launcher with the given command.
func NewIDELauncher(command string) *IDELauncher {
	return &IDELauncher{command: command}
}

// Command returns the underlying command.
func (l *IDELauncher) Command() string {
	return l.command
}

// Launch starts the IDE for the given workspace path.
func (l *IDELauncher) Launch(workspacePath string) error {
	if l.command == "" {
		return ErrIDENotConfigured
	}
	return LaunchWorkspaceCommand(l.command, workspacePath)
}

// SplitCommand parses a command string with quote support.
func SplitCommand(cmd string) []string {
	return splitCommand(cmd)
}

// splitCommand parses a command string with quote support.
func splitCommand(cmd string) []string {
	var parts []string
	var current strings.Builder
	inQuote := false
	quoteChar := rune(0)

	for _, c := range cmd {
		if inQuote {
			if c == quoteChar {
				inQuote = false
				quoteChar = 0
			} else {
				current.WriteRune(c)
			}
			continue
		}

		if c == '"' || c == '\'' {
			inQuote = true
			quoteChar = c
			continue
		}

		if c == ' ' || c == '\t' {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
			continue
		}

		current.WriteRune(c)
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// IDE errors.
var (
	// ErrIDENotConfigured is returned when IDE command is not configured.
	ErrIDENotConfigured = &ideError{"IDE command not configured"}
	// ErrWorkspacePathEmpty is returned when workspace path is empty.
	ErrWorkspacePathEmpty = &ideError{"workspace path is empty"}
	// ErrInvalidCommand is returned when IDE command is invalid.
	ErrInvalidCommand = &ideError{"invalid IDE command"}
)

type ideError struct {
	msg string
}

func (e *ideError) Error() string {
	return e.msg
}

// Is checks if an error matches a specific IDE error.
func (e *ideError) Is(target error) bool {
	t, ok := target.(*ideError)
	if !ok {
		return false
	}
	return e.msg == t.msg
}

// Unwrap returns the underlying error message.
func (e *ideError) Unwrap() error {
	return errors.New(e.msg)
}

// Format implements [fmt.Formatter] for better error messages.
func (e *ideError) Format(f fmt.State, verb rune) {
	switch verb {
	case 'v':
		if f.Flag('+') {
			fmt.Fprintf(f, "%s: %s", "ideError", e.msg)
			return
		}
		fallthrough
	case 's':
		fmt.Fprint(f, e.msg)
	case 'q':
		fmt.Fprintf(f, "%q", e.msg)
	}
}
