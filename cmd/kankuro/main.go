package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"

	"kankuro/internal/app"
	"kankuro/internal/config"
	"kankuro/internal/driver"
	"kankuro/internal/session"
)

// main wires repository, session manager, and TUI program startup.
func main() {
	// Create and configure driver registry
	registry := driver.NewBuiltinRegistry()

	if err := run(registry); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(registry driver.Registry) error {
	var cfgDir string
	var debug bool
	flag.StringVar(&cfgDir, "config-dir", os.Getenv("KANKURO_CONFIG_DIR"), "override config directory")
	flag.BoolVar(&debug, "debug", false, "write debug logs")
	flag.Parse()

	// internal/session logs background-only failures (no caller to report
	// them to, e.g. a persist failure from the watch ticker) via an injected
	// *slog.Logger (session.WithLogger) rather than a global one, so it never
	// risks writing to stderr and corrupting the rendered TUI frame. Without
	// --debug that logger stays a no-op (its default); with --debug, point it
	// at the same debug log file used for bubbletea's own logging.
	var sessionLogger *slog.Logger
	if debug {
		if err := os.MkdirAll(logDir(), 0o700); err == nil {
			logFile, logErr := tea.LogToFile(filepath.Join(logDir(), "kankuro.log"), "debug")
			if logErr == nil {
				defer logFile.Close()
				sessionLogger = slog.New(slog.NewTextHandler(logFile, nil))
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

	sm := session.NewManager(
		repo.Dir(),
		repo,
		session.WithDriverRegistry(registry),
		session.WithLogger(sessionLogger),
	)
	application, err := app.New(repo, sm)
	if err != nil {
		return fmt.Errorf("ui init error: %w", err)
	}

	p := tea.NewProgram(application)
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
