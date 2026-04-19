package views

import (
	"errors"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"kankuro/internal/config"
	"kankuro/internal/ui/commands"
	"kankuro/internal/ui/component"
	"kankuro/internal/ui/components"
	"kankuro/internal/ui/messages"
)

// SourceManagerCallbacks defines repository callbacks.
type SourceManagerCallbacks struct {
	LoadSources func() []components.SourceItem

	SaveSimpleSource    func(path string, editingIndex int) error
	SaveWorkspaceSource func(source config.WorkspaceSource, editingIndex int) error

	DeleteSource func(index int) error

	Refresh func()
}

// SourceManager is a composite view for managing source lists.
// It wraps SourceList and configures it for skills, personas, or workspaces.
type SourceManager struct {
	component.Base

	sourceList *components.SourceList
	sourceType SourceType
	typeConfig SourceTypeConfig
	callbacks  SourceManagerCallbacks

	prevView messages.ViewType
	cmds     *commands.Commands

	err error
}

// NewSourceManager creates a new source manager for the given type.
func NewSourceManager(
	sourceType SourceType,
	callbacks SourceManagerCallbacks,
	cmds *commands.Commands,
) *SourceManager {
	typeConfig := GetSourceTypeConfig(sourceType)

	listConfig := components.SourceListConfig{
		Title:       typeConfig.Title,
		Legend:      typeConfig.Legend,
		FormPrompt:  typeConfig.FormPrompt,
		FormFields:  typeConfig.FormFields,
		Field2Label: typeConfig.Field2Label,
	}

	listCallbacks := components.SourceListCallbacks{
		OnSave: func(item components.SourceItem, editingIndex int) error {
			return handleSave(sourceType, callbacks, item, editingIndex)
		},
		OnDelete: func(index int) error {
			if callbacks.DeleteSource != nil {
				return callbacks.DeleteSource(index)
			}
			return nil
		},
		OnRefresh: func() {
			if callbacks.Refresh != nil {
				callbacks.Refresh()
			}
		},
		OnClose: func() {},
	}

	sourceList := components.NewSourceList(listConfig, listCallbacks)

	return &SourceManager{
		sourceList: sourceList,
		sourceType: sourceType,
		typeConfig: typeConfig,
		callbacks:  callbacks,
		cmds:       cmds,
	}
}

// handleSave routes save to the appropriate callback based on source type.
func handleSave(
	sourceType SourceType,
	callbacks SourceManagerCallbacks,
	item components.SourceItem,
	editingIndex int,
) error {
	switch sourceType {
	case SourceTypeSkills, SourceTypePersonas:
		if callbacks.SaveSimpleSource != nil {
			return callbacks.SaveSimpleSource(item.Display(), editingIndex)
		}

	case SourceTypeWorkspaces:
		if wsItem, ok := item.(components.WorkspaceSourceItem); ok {
			if callbacks.SaveWorkspaceSource != nil {
				return callbacks.SaveWorkspaceSource(wsItem.Source, editingIndex)
			}
		}
		return errors.New("invalid workspace source item")
	}

	return nil
}

// SetPrevView sets where to return when closed.
func (m *SourceManager) SetPrevView(view messages.ViewType) {
	m.prevView = view
}

// SetData loads sources into the manager.
func (m *SourceManager) SetData() {
	if m.callbacks.LoadSources != nil {
		items := m.callbacks.LoadSources()
		m.sourceList.SetItems(items)
	}
}

// Init initializes the manager.
func (m *SourceManager) Init() tea.Cmd {
	m.SetData()
	return m.sourceList.Init()
}

// Update handles messages.
func (m *SourceManager) Update(msg tea.Msg) (component.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.String() == "esc" && m.sourceList.Mode == components.SourceListModeList {
			return m, func() tea.Msg {
				return messages.NavigateMsg{View: m.prevView}
			}
		}
	}

	model, cmd := m.sourceList.Update(msg)
	sl, ok := model.(*components.SourceList)
	if !ok {
		return m, cmd
	}
	m.sourceList = sl

	return m, cmd
}

// View renders the manager.
func (m *SourceManager) View() string {
	if m.err != nil {
		return m.sourceList.View() + "\n\nError: " + m.err.Error()
	}
	return m.sourceList.View()
}

// SetSize implements component.Model.
func (m *SourceManager) SetSize(width, height int) {
	m.Base.SetSize(width, height)
	m.sourceList.SetSize(width, height)
}

// SourceType returns the managed source type.
func (m *SourceManager) SourceType() SourceType {
	return m.sourceType
}

// GenerateID creates a new unique ID for workspace sources.
func GenerateID() string {
	return fmt.Sprintf("ws-%d", time.Now().UnixNano())
}
