package app

import (
	"errors"
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
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
	editor       cardEditorPort
	sessionEpoch uint64
	sessions     *session.Manager
}

// loadGlobalConfigData reads registries shared by all dashboards.
func (r refresher) loadGlobalConfigData() messages.DataLoadedMsg {
	repo := r.repo
	var errs []error
	skills, err := repo.ListSkills()
	if err != nil {
		errs = append(errs, fmt.Errorf("list skills: %w", err))
		skills = nil
	}
	personas, err := repo.ListPersonas()
	if err != nil {
		errs = append(errs, fmt.Errorf("list personas: %w", err))
		personas = nil
	}
	mcps, err := repo.ListMCPs()
	if err != nil {
		errs = append(errs, fmt.Errorf("list mcps: %w", err))
		mcps = nil
	}
	workspaces, err := repo.ListWorkspaces()
	if err != nil {
		errs = append(errs, fmt.Errorf("list workspaces: %w", err))
		workspaces = nil
	}
	puppets, err := repo.ListPuppets()
	if err != nil {
		errs = append(errs, fmt.Errorf("list puppets: %w", err))
		puppets = nil
	}

	var choices contracts.CardEditorChoices
	if r.editor != nil {
		choices = r.editor.ReadCardEditorChoices()
		errs = append(errs, editorChoicesError(choices))
	}
	return messages.DataLoadedMsg{
		EditorChoices: choices,
		Skills:        skills,
		Personas:      personas,
		MCPs:          mcps,
		Workspaces:    workspaces,
		Puppets:       puppets,
		Err:           errors.Join(errs...),
		RegistryErr:   errors.Join(errs...),
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

// configFingerprintMsg carries focused read results and a certified opaque
// token. Empty tokens never update the baseline; unchanged results carry no data.
type configFingerprintMsg struct {
	token              contracts.RegistryChangeToken
	requestID          uint64
	inventoryRequestID uint64
	dashboardID        string
	generation         uint64
	sessionEpoch       uint64
	data               *messages.DataLoadedMsg
}

func tick() tea.Cmd {
	const tickInterval = time.Second * 2
	return tea.Tick(tickInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type tickMsg time.Time

func editorChoicesError(choices contracts.CardEditorChoices) error {
	return errors.Join(choices.Workspaces.Err, choices.Skills.Err, choices.Personas.Err, choices.MCPs.Err, choices.Puppets.Err)
}
