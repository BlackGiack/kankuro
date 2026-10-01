package app

import (
	"errors"
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"

	"kankuro/internal/config"
	"kankuro/internal/session"
	"kankuro/internal/ui/messages"
)

// refresher builds the commands that read config and session state off the
// UI goroutine. It deliberately holds no reference to the Coordinator or its
// views: every closure it returns can only reach repo, sessions and the
// values passed in as arguments, so invariant 1 (no tea.Cmd closure touches
// coordinator state) holds here by construction.
type refresher struct {
	repo     config.Repository
	sessions *session.Manager
}

// full performs a full, unconditional refresh of every registry plus the
// session list and preview scrollback in one DataLoadedMsg, without
// seeding the fingerprint baseline. It exists solely for
// Coordinator.RefreshData(), the exported wrapper integration tests drive to
// get one deterministic DataLoadedMsg back - see fullSeeded for the variant
// every real call site uses.
func (r refresher) full(previewID string) tea.Cmd {
	return func() tea.Msg {
		return r.loadFullData(previewID)
	}
}

// fullSeeded is the combined refresh every mutation-triggered path (Init,
// DataRefreshRequestedMsg, CardSavedMsg, ...) actually uses. Unlike full, it
// also stats the registries (same as configIfChanged) and wraps the result
// in configFingerprintMsg so the coordinator adopts the new baseline on its
// own goroutine. Without this, a full refresh left the baseline stale, so the
// very next idle tick always saw a "changed" stamp and reloaded everything a
// second time - every mutation was parsed twice - see debt_impl_plan_5.md
// review finding 2.
func (r refresher) fullSeeded(previewID string) tea.Cmd {
	return func() tea.Msg {
		stamps, err := r.repo.Fingerprint()
		data := r.loadFullData(previewID)
		if err != nil {
			// Fingerprint itself failed - don't block the refresh on it,
			// just skip seeding the baseline; the next tick falls back to
			// one extra reload, same as before this fix.
			return data
		}
		return configFingerprintMsg{stamps: stamps, data: &data}
	}
}

// sessionsOnly reads only the (in-memory) session list and the current
// preview's scrollback - no disk I/O - and returns it as a
// SessionsRefreshedMsg, not a DataLoadedMsg: the latter would broadcast to
// every DataReceiver, including config-registry views like SourceManager,
// whose SetData reloads source files from disk regardless of what changed
// (see debt_impl_plan_5.md review finding 1). This is what the idle tick
// calls every 2s.
func (r refresher) sessionsOnly(previewID string) tea.Cmd {
	return func() tea.Msg {
		sessions := r.sessions.List()
		var scrollback []byte
		if previewID != "" {
			scrollback = r.sessions.GetScrollback(previewID)
		}
		return messages.SessionsRefreshedMsg{
			Sessions:          sessions,
			PreviewSessionID:  previewID,
			PreviewScrollback: scrollback,
		}
	}
}

// configIfChanged is the idle tick's config half: it stats every registry
// file (cheap - no parsing) and only re-reads/re-parses/re-sorts them if at
// least one changed since prev, the baseline the caller snapshotted. See
// debt_impl_plan_5.md step 1 and config.FileRepository.Fingerprint for the
// mtime-precision caveat this accepts.
func (r refresher) configIfChanged(prev map[string]config.FileStamp) tea.Cmd {
	return func() tea.Msg {
		stamps, err := r.repo.Fingerprint()
		if err == nil && fingerprintsEqual(prev, stamps) {
			// Nothing changed - leave stamps nil so this result cannot
			// stomp a newer baseline a concurrent mutation-triggered
			// fullSeeded may have just set (review_dp_5_2.md finding 1).
			return configFingerprintMsg{}
		}
		data := r.loadConfigData()
		return configFingerprintMsg{stamps: stamps, data: &data}
	}
}

// loadFullData reads every config registry plus the session list and
// preview scrollback into one DataLoadedMsg. It performs blocking disk I/O
// and must only be called from inside a tea.Cmd closure.
func (r refresher) loadFullData(previewID string) messages.DataLoadedMsg {
	configMsg := r.loadConfigData()

	sessions := r.sessions.List()
	var scrollback []byte
	if previewID != "" {
		scrollback = r.sessions.GetScrollback(previewID)
	}

	configMsg.Sessions = sessions
	configMsg.PreviewSessionID = previewID
	configMsg.PreviewScrollback = scrollback
	return configMsg
}

// loadConfigData reads every config registry once and returns a
// DataLoadedMsg carrying just those fields (Sessions/PreviewSessionID/
// PreviewScrollback left zero) - the shared core of loadFullData (used by
// full and fullSeeded) and configIfChanged. It performs blocking disk I/O
// and must only be called from inside a tea.Cmd closure, never on the main
// coordinator goroutine.
func (r refresher) loadConfigData() messages.DataLoadedMsg {
	var errs []error

	cards, err := r.repo.ListIDCards()
	if err != nil {
		errs = append(errs, fmt.Errorf("list cards: %w", err))
		cards = nil
	}
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
		Cards:      cards,
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
