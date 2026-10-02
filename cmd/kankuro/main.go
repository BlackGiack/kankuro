package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

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

	// Bubble Tea catches SIGINT (Run returns ErrInterrupted) and SIGTERM (it
	// becomes a QuitMsg) but not SIGHUP, which is what closing the terminal
	// window sends. Go's default action for SIGHUP would end kankuro on the
	// spot, before any cleanup, so turn it into a normal quit. If the terminal
	// is really gone, Run may return a write error instead of nil; the
	// shutdown below runs either way.
	hup := make(chan os.Signal, 1)
	signal.Notify(hup, syscall.SIGHUP)
	defer signal.Stop(hup)
	go func() {
		<-hup // blocks until SIGHUP; the process exits long before it matters otherwise
		p.Quit()
	}()

	_, runErr := p.Run()

	// Every exit path - q, ctrl+c, SIGINT, SIGTERM, SIGHUP, a runtime error -
	// ends up here, and running agents must not outlive kankuro. Plain
	// sequential code, not a defer, so the order is visible: the TUI is gone,
	// then the agents are stopped, then the error (if any) is returned.
	//
	// Bubble Tea has released the terminal by now, so one line to stderr is
	// safe here and only here; the "never write to stdout/stderr" rule is
	// about the TUI being live. Do not copy this into app code. While this
	// waits, Bubble Tea no longer handles signals, so a second ctrl+c or
	// SIGTERM gets Go's default action and exits at once - an acceptable
	// "force quit".
	stopSessions(sm)

	if runErr != nil {
		return fmt.Errorf("runtime error: %w", runErr)
	}
	return nil
}

// stopSessions kills every running session, announcing it on stderr when
// there is anything to wait for. See the caller for why stderr is allowed.
func stopSessions(sm *session.Manager) {
	if n := sm.RunningCount(); n > 0 {
		fmt.Fprintf(os.Stderr, "kankuro: stopping %d running session(s)…\n", n)
	}
	sm.Shutdown()
}

// logDir resolves the directory where debug logs are written.
func logDir() string {
	state, err := os.UserCacheDir()
	if err != nil {
		return os.TempDir()
	}
	return filepath.Join(state, "kankuro")
}
