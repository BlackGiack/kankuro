package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"kankuro/internal/config"
	"kankuro/internal/session"
	"kankuro/internal/ui"
)

// main wires repository, session manager, and TUI program startup.
func main() {
	var cfgDir string
	var debug bool
	flag.StringVar(&cfgDir, "config-dir", "", "override config directory")
	flag.BoolVar(&debug, "debug", false, "write debug logs")
	flag.Parse()

	if debug {
		if err := os.MkdirAll(logDir(), 0o755); err == nil {
			f, err := tea.LogToFile(filepath.Join(logDir(), "kankuro.log"), "debug")
			if err == nil {
				defer f.Close()
			}
		}
	}

	repo, err := config.NewFileRepository(cfgDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config init error: %v\n", err)
		os.Exit(1)
	}
	if err := repo.SyncSkillsFromSources(); err != nil {
		fmt.Fprintf(os.Stderr, "skill sync error: %v\n", err)
		os.Exit(1)
	}

	sm := session.NewManager(repo.Dir())
	app, err := ui.New(repo, sm)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ui init error: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "runtime error: %v\n", err)
		sm.Shutdown()
		os.Exit(1)
	}
}

// logDir resolves the directory where debug logs are written.
func logDir() string {
	state, err := os.UserCacheDir()
	if err != nil {
		return os.TempDir()
	}
	return filepath.Join(state, "kankuro")
}
