package ui

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

const idePathPlaceholder = "$PATH"

func (a *App) openIDEConfig() (tea.Model, tea.Cmd) {
	a.mode = viewIDE
	return a, nil
}

func (a *App) launchIDE() {
	if a.ideConfig.Command == "" {
		a.err = ErrIDENotConfigured
		return
	}

	sel := a.sessionsList.Index()
	if sel < 0 || sel >= len(a.sessions) {
		a.err = ErrNoSessionSelected
		return
	}
	s := a.sessions[sel]

	workingDir := s.WorkingDir
	if workingDir == "" {
		a.err = ErrSessionWorkspaceMissing
		return
	}

	absPath, resolveErr := resolveWorkspacePath(workingDir)
	if resolveErr != nil {
		a.err = resolveErr
		return
	}

	cmd := strings.ReplaceAll(a.ideConfig.Command, idePathPlaceholder, absPath)

	parts := splitCommand(cmd)
	if len(parts) == 0 {
		a.err = ErrInvalidIDECommand
		return
	}

	execCmd := exec.CommandContext(context.Background(), parts[0], parts[1:]...) //nolint:gosec // Command is user-configured IDE path
	execCmd.Dir = absPath
	execCmd.Stdin = nil
	execCmd.Stdout = nil
	execCmd.Stderr = nil

	if startErr := execCmd.Start(); startErr != nil {
		a.err = startErr
		return
	}
	go func() { _ = execCmd.Wait() }()
}

func resolveWorkspacePath(path string) (string, error) {
	if path == "" {
		return "", ErrSessionWorkspaceMissing
	}
	abs, pathErr := toAbsolutePath(path)
	if pathErr != nil {
		return "", pathErr
	}
	if _, statErr := os.Stat(abs); statErr != nil {
		return "", statErr
	}
	return abs, nil
}

func toAbsolutePath(p string) (string, error) {
	return filepath.Abs(p)
}

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
	ErrIDENotConfigured        = &ideError{"IDE command not configured"}
	ErrNoSessionSelected       = &ideError{"no session selected"}
	ErrSessionWorkspaceMissing = &ideError{"session workspace path is missing"}
	ErrInvalidIDECommand       = &ideError{"invalid IDE command"}
)

type ideError struct {
	msg string
}

func (e *ideError) Error() string { return e.msg }

func (a *App) updateIDE(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case keyEsc:
			a.ideInput.Blur()
			a.mode = viewDashboard
			return a, nil
		case keyEnter, "e":
			a.ideInput.SetValue(a.ideConfig.Command)
			a.mode = viewIDEForm
			cmd := a.ideInput.Focus()
			return a, cmd
		}
	}
	return a, nil
}

func (a *App) updateIDEForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case keyEsc:
			a.mode = viewIDE
			return a, nil
		case keyEnter:
			a.ideConfig.Command = a.ideInput.Value()
			a.ideInput.Blur()
			if saveErr := a.repo.SaveIDEConfig(a.ideConfig); saveErr != nil {
				a.err = saveErr
			}
			a.mode = viewIDE
			return a, nil
		}
	}
	var cmd tea.Cmd
	a.ideInput, cmd = a.ideInput.Update(msg)
	return a, cmd
}
