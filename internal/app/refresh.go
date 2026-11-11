package app

import (
	"errors"
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
)

// refresher builds the commands that read config and session state off the
// UI goroutine. It deliberately holds no reference to the Coordinator or its
// views: every closure it returns can only reach repo, sessions and the
// values passed in as arguments, so invariant 1 (no tea.Cmd closure touches
// coordinator state) holds here by construction.
type refresher struct {
	repo         config.Repository
	sessionEpoch uint64
	sessions     *session.Manager
}

// loadGlobalConfigData reads registries shared by all dashboards.
func (r refresher) loadGlobalConfigData() messages.DataLoadedMsg {
	var errs []error
	skills, err := r.repo.ListSkills()
	if err != nil {
		errs = append(errs, fmt.Errorf("list skills: %w", err))
		skills = nil
	}
	personas, err := r.repo.ListPersonas()
	if err != nil {
		errs = append(errs, fmt.Errorf("list personas: %w", err))
		personas = nil
	}
	mcps, err := r.repo.ListMCPs()
	if err != nil {
		errs = append(errs, fmt.Errorf("list mcps: %w", err))
		mcps = nil
	}
	workspaces, err := r.repo.ListWorkspaces()
	if err != nil {
		errs = append(errs, fmt.Errorf("list workspaces: %w", err))
		workspaces = nil
	}
	puppets, err := r.repo.ListPuppets()
	if err != nil {
		errs = append(errs, fmt.Errorf("list puppets: %w", err))
		puppets = nil
	}

	return messages.DataLoadedMsg{
		Skills:     skills,
		Personas:   personas,
		MCPs:       mcps,
		Workspaces: workspaces,
		Puppets:    puppets,
		Err:        errors.Join(errs...),
	}
}

// sessionLoadErrors surfaces any error hit while hydrating persisted
// sessions at startup (corrupt/unreadable session files are skipped, not
// fatal, but the user should know they lost a session record).
func (r refresher) sessionLoadErrors() tea.Cmd {
	return func() tea.Msg {
		if err := r.sessions.LoadErrors(); err != nil {
			return messages.ErrorMsg{Err: err}
		}
		return nil
	}
}

// syncPuppets rescans the puppet sources off the UI goroutine.
func (r refresher) syncPuppets() tea.Cmd {
	return func() tea.Msg {
		_ = r.repo.SyncPuppetsFromSources()
		return messages.PuppetsSyncedMsg{}
	}
}

// configFingerprintMsg reports the result of a fingerprint check. stamps is
// only non-nil when the baseline should actually change - the "unchanged"
// result from configIfChanged deliberately leaves it nil so it never
// clobbers a newer baseline a concurrent fullSeeded may have just set
// (see Coordinator.handleConfigFingerprint). data is only non-nil when a
// registry actually changed and was re-read.
type configFingerprintMsg struct {
	stamps map[string]config.FileStamp
	data   *messages.DataLoadedMsg
}

func fingerprintsEqual(a, b map[string]config.FileStamp) bool {
	if a == nil || b == nil {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	for name, stampA := range a {
		stampB, ok := b[name]
		if !ok || !stampA.ModTime.Equal(stampB.ModTime) || stampA.Size != stampB.Size {
			return false
		}
	}
	return true
}

func tick() tea.Cmd {
	const tickInterval = time.Second * 2
	return tea.Tick(tickInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type tickMsg time.Time
