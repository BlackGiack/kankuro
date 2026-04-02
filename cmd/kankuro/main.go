package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"kankuro/internal/config"
	"kankuro/internal/driver"
	"kankuro/internal/session"
	"kankuro/internal/ui"
)

// main wires repository, session manager, and TUI program startup.
func main() {
	driver.RegisterDriver("claude", func() driver.AgentDriver { return driver.NewClaudeCodeDriver() })
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var cfgDir string
	var debug bool
	flag.StringVar(&cfgDir, "config-dir", os.Getenv("KANKURO_CONFIG_DIR"), "override config directory")
	flag.BoolVar(&debug, "debug", false, "write debug logs")
	flag.Parse()

	if debug {
		if err := os.MkdirAll(logDir(), 0o700); err == nil {
			logFile, logErr := tea.LogToFile(filepath.Join(logDir(), "kankuro.log"), "debug")
			if logErr == nil {
				defer logFile.Close()
			}
		}
	}

	repo, err := config.NewFileRepository(cfgDir)
	if err != nil {
		return fmt.Errorf("config init error: %w", err)
	}

	if err = repo.SyncSkillsFromSources(); err != nil {
		return fmt.Errorf("skill sync error: %w", err)
	}

	sm := session.NewManager(repo.Dir())
	app, err := ui.New(repo, sm)
	if err != nil {
		return fmt.Errorf("ui init error: %w", err)
	}

	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err = p.Run(); err != nil {
		sm.Shutdown()
		return fmt.Errorf("runtime error: %w", err)
	}

	return nil
}

// logDir resolves the directory where debug logs are written.
func logDir() string {
	state, err := os.UserCacheDir()
	if err != nil {
		return os.TempDir()
	}
	return filepath.Join(state, "kankuro")
}
