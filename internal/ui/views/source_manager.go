package views

import (
	"context"
	"errors"

	tea "charm.land/bubbletea/v2"
	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/ui/component"
	"github.com/BlackGiack/kankuro/internal/ui/components"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
)

// SourceManagerCallbacks are application capabilities. The manager invokes them only from commands.
type SourceManagerCallbacks struct {
	Load   func(contracts.SourceKind) ([]contracts.SourceRow, error)
	Mutate func(context.Context, contracts.SourceMutation) contracts.SourceMutationResult
	Sync   func(context.Context, contracts.SourceKind) contracts.SourceSyncResult
}

type sourceLoaded struct {
	kind           SourceType
	visit, request uint64
	rows           []contracts.SourceRow
	err            error
}
type sourceMutated struct {
	kind           SourceType
	visit, request uint64
	result         contracts.SourceMutationResult
}
type sourceSynced struct {
	kind           SourceType
	visit, request uint64
	result         contracts.SourceSyncResult
}

type SourceManager struct {
	component.Base

	sourceList     *components.SourceList
	sourceType     SourceType
	typeConfig     SourceTypeConfig
	callbacks      SourceManagerCallbacks
	visit, request uint64
	pending        bool
}

func NewSourceManager(sourceType SourceType, callbacks SourceManagerCallbacks) *SourceManager {
	cfg := GetSourceTypeConfig(sourceType)
	m := &SourceManager{sourceType: sourceType, typeConfig: cfg, callbacks: callbacks}
	m.sourceList = components.NewSourceList(
		components.SourceListConfig{
			Title:       cfg.Title,
			Legend:      cfg.Legend,
			FormPrompt:  cfg.FormPrompt,
			FormFields:  cfg.FormFields,
			Field2Label: cfg.Field2Label,
		},
		components.SourceListCallbacks{
			OnSave: func(item components.SourceItem, _ int) (tea.Cmd, error) {
				if m.pending {
					return nil, errors.New("source operation in progress")
				}
				row := contracts.SourceMutation{Kind: contracts.SourceKind(m.sourceType), Identity: item.ID()}
				if ws, ok := item.(components.WorkspaceSourceItem); ok {
					row.Path, row.IsRoot = ws.Source.Path, ws.Source.IsRoot
				} else {
					row.Path = item.Display()
				}
				return m.mutate(row), nil
			},
			OnDelete: func(index int) (tea.Cmd, error) {
				if m.pending || index < 0 || index >= len(m.sourceList.Items) {
					return nil, errors.New("source operation unavailable")
				}
				return m.mutate(
					contracts.SourceMutation{
						Kind:     contracts.SourceKind(m.sourceType),
						Identity: m.sourceList.Items[index].ID(),
						Delete:   true,
					},
				), nil
			},
			OnRefresh: func() tea.Cmd { return m.sync() },
		},
	)
	return m
}

func (m *SourceManager) load() tea.Cmd {
	if m.callbacks.Load == nil {
		return nil
	}
	m.request++
	load, kind, visit, request := m.callbacks.Load, m.sourceType, m.visit, m.request
	return func() tea.Msg {
		rows, err := load(contracts.SourceKind(kind))
		return sourceLoaded{kind: kind, visit: visit, request: request, rows: rows, err: err}
	}
}

func (m *SourceManager) mutate(input contracts.SourceMutation) tea.Cmd {
	if m.callbacks.Mutate == nil {
		return nil
	}
	m.pending = true
	m.request++
	mutate, kind, visit, request := m.callbacks.Mutate, m.sourceType, m.visit, m.request
	return func() tea.Msg {
		result := mutate(context.Background(), input)
		msg := sourceMutated{kind: kind, visit: visit, request: request, result: result}
		if result.Refresh {
			return tea.Batch(func() tea.Msg { return msg }, func() tea.Msg { return messages.DataRefreshRequestedMsg{} })()
		}
		return msg
	}
}

func (m *SourceManager) sync() tea.Cmd {
	if m.pending || m.callbacks.Sync == nil {
		return nil
	}
	m.pending = true
	m.request++
	sync, kind, visit, request := m.callbacks.Sync, m.sourceType, m.visit, m.request
	return func() tea.Msg {
		result := sync(context.Background(), contracts.SourceKind(kind))
		msg := sourceSynced{kind: kind, visit: visit, request: request, result: result}
		if result.Refresh {
			return tea.Batch(func() tea.Msg { return msg }, func() tea.Msg { return messages.DataRefreshRequestedMsg{} })()
		}
		return msg
	}
}

func (m *SourceManager) Init() tea.Cmd                    { m.visit++; m.pending = false; return m.load() }
func (m *SourceManager) SetData(_ messages.DataLoadedMsg) {}

// Reload schedules a fresh read without replacing an in-flight mutation.
func (m *SourceManager) Reload() tea.Cmd {
	if m.pending {
		return nil
	}
	return m.load()
}
func (m *SourceManager) Update(msg tea.Msg) (component.Model, tea.Cmd) {
	switch value := msg.(type) {
	case sourceLoaded:
		if value.kind != m.sourceType || value.visit != m.visit || value.request != m.request {
			return m, nil
		}
		if value.err != nil {
			return m, func() tea.Msg { return messages.ErrorMsg{Err: value.err} }
		}
		items := make([]components.SourceItem, 0, len(value.rows))
		for _, row := range value.rows {
			if m.sourceType == SourceTypeWorkspaces {
				items = append(items, components.WorkspaceSourceItem{Source: row})
			} else {
				items = append(items, components.PathSourceItem{Path: row.Path, Identity: row.Identity})
			}
		}
		m.sourceList.SetItems(items)
		return m, nil
	case sourceMutated:
		if value.kind != m.sourceType || value.visit != m.visit || value.request != m.request {
			return m, nil
		}
		m.pending = false
		result := value.result
		if result.Saved {
			m.sourceList.Mode = components.SourceListModeList
		}
		var load tea.Cmd
		if result.Saved || result.Reload {
			load = m.load()
		}
		err := result.Err
		if err == nil {
			err = result.Warning
		}
		if err != nil {
			return m, tea.Batch(load, func() tea.Msg { return messages.ErrorMsg{Err: err} })
		}
		return m, load
	case sourceSynced:
		if value.kind != m.sourceType || value.visit != m.visit || value.request != m.request {
			return m, nil
		}
		m.pending = false
		if value.result.Err != nil {
			err := value.result.Err
			return m, tea.Batch(m.load(), func() tea.Msg { return messages.ErrorMsg{Err: err} })
		}
		return m, m.load()
	case tea.KeyPressMsg:
		if value.Code == tea.KeyEsc && m.sourceList.Mode == components.SourceListModeList {
			return m, func() tea.Msg { return messages.NavigateBackMsg{} }
		}
		if m.pending {
			return m, nil
		}
	}
	model, cmd := m.sourceList.Update(msg)
	if list, ok := model.(*components.SourceList); ok {
		m.sourceList = list
	}
	return m, cmd
}
func (m *SourceManager) View() string { return m.sourceList.View() }
func (m *SourceManager) SetSize(width, height int) {
	m.Base.SetSize(width, height)
	m.sourceList.SetSize(width, height)
}
func (m *SourceManager) SourceType() SourceType             { return m.sourceType }
func (m *SourceManager) TypeConfig() SourceTypeConfig       { return m.typeConfig }
func (m *SourceManager) SourceList() *components.SourceList { return m.sourceList }
