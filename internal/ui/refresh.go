package ui

import (
	"context"
	"errors"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
)

// refresher builds the commands that read config and session state off the
// UI goroutine. It deliberately holds no reference to the Coordinator or its
// views: every closure it returns can only reach repo, sessions and the
// values passed in as arguments, so invariant 1 (no tea.Cmd closure touches
// coordinator state) holds here by construction.
type refresher struct {
	editor        cardEditorPort
	configuration configurationPort
	sessionEpoch  uint64
}

// loadGlobalConfigData reads registries shared by all dashboards.
func (r refresher) loadGlobalConfigData() messages.DataLoadedMsg {
	var errs []error

	var rows []contracts.MCPRow
	if r.configuration != nil {
		var rowErr error
		rows, rowErr = r.configuration.ReadMCPRows()
		if rowErr != nil {
			errs = append(errs, rowErr)
		}
	}
	var choices contracts.CardEditorChoices
	if r.editor != nil {
		choices = r.editor.ReadCardEditorChoices()
		errs = append(errs, editorChoicesError(choices))
	}
	return messages.DataLoadedMsg{
		EditorChoices: choices,
		MCPRows:       rows,
		Err:           errors.Join(errs...),
		RegistryErr:   errors.Join(errs...),
	}
}

func startupErrorCommand(err error) tea.Cmd {
	return func() tea.Msg {
		if err != nil {
			return messages.ErrorMsg{Err: err}
		}
		return nil
	}
}

// syncPuppets rescans the puppet sources off the UI goroutine.
func (r refresher) syncPuppets() tea.Cmd {
	return func() tea.Msg {
		if r.configuration == nil {
			return nil
		}
		result := r.configuration.SyncSources(context.Background(), contracts.SourcePuppets)
		if result.Err != nil {
			return tea.Batch(
				func() tea.Msg { return messages.ErrorMsg{Err: result.Err} },
				func() tea.Msg { return messages.PuppetsSyncedMsg{} },
			)()
		}
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
