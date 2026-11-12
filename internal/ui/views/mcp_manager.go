package views

import (
	"errors"

	tea "charm.land/bubbletea/v2"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/ui/component"
	"github.com/BlackGiack/kankuro/internal/ui/components"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
)

type mcpManagerItem struct{ server config.MCPServer }

func (i mcpManagerItem) ID() string { return i.server.ID }
func (i mcpManagerItem) Display() string {
	if i.server.ID == config.BuiltinMCPKankuro {
		return i.server.Name + " (built-in, read-only)"
	}
	return i.server.Name
}

// MCPManager manages the global MCP registry using repository snapshots.
type MCPManager struct {
	component.Base

	list         *components.SourceList
	modal        *components.Modal
	pendingID    string
	deleting     bool
	loaded       bool
	scope        messages.MCPManagerScope
	pendingVisit uint64
	admitted     bool
}

func NewMCPManager() *MCPManager {
	m := &MCPManager{modal: components.NewModal("Delete MCP?")}
	m.list = components.NewSourceList(components.SourceListConfig{
		Title: "MCP Servers", ListLabel: "Available MCPs:", NewLabel: "[ New MCP ]",
		Help: "Enter=edit/new  Del=delete  Esc=back",
	}, components.SourceListCallbacks{})
	m.modal.SetLegend("Enter/y=confirm  Esc/n=cancel")
	return m
}
func (m *MCPManager) Init() tea.Cmd {
	m.modal.Close()
	return nil
}

// SetScope assigns the identity of the current manager visit.
func (m *MCPManager) SetScope(scope messages.MCPManagerScope) { m.scope = scope }

// Leave cancels a confirmation that has not yet been admitted by the coordinator.
func (m *MCPManager) Leave() {
	m.modal.Close()
	if m.deleting && !m.admitted {
		m.DeleteFinished()
	}
}
func (m *MCPManager) AdmitDelete(visit uint64) bool {
	if !m.deleting || m.pendingVisit != visit || m.admitted {
		return false
	}
	m.admitted = true
	return true
}
func (m *MCPManager) DiscardDelete(visit uint64) {
	if m.deleting && !m.admitted && m.pendingVisit == visit {
		m.DeleteFinished()
	}
}
func (m *MCPManager) SetData(data messages.DataLoadedMsg) {
	if data.MCPs == nil {
		return
	}
	selected := ""
	if m.list.Cursor < len(m.list.Items) {
		selected = m.list.Items[m.list.Cursor].ID()
	}
	wasNew := m.loaded && m.list.Cursor == len(m.list.Items)
	m.loaded = true
	items := make([]components.SourceItem, len(data.MCPs))
	for i, server := range data.MCPs {
		items[i] = mcpManagerItem{server}
	}
	m.list.SetItems(items)
	if wasNew {
		m.list.Cursor = len(items)
	} else if selected != "" {
		for i, item := range items {
			if item.ID() == selected {
				m.list.Cursor = i
				break
			}
		}
	}
}

//nolint:gocognit // Modal and list keys are dispatched together.
func (m *MCPManager) Update(msg tea.Msg) (component.Model, tea.Cmd) {
	key, isKey := msg.(tea.KeyPressMsg)
	if !isKey {
		return m, nil
	}
	if m.modal.IsOpen() {
		switch {
		case key.Code == tea.KeyEsc || key.Text == "n":
			m.modal.Close()
			m.pendingID = ""
		case key.Code == tea.KeyEnter || key.Text == "y":
			id, scope := m.pendingID, m.scope
			m.modal.Close()
			m.deleting = true
			m.pendingVisit = scope.Visit
			m.admitted = false
			return m, func() tea.Msg { return messages.MCPManagerDeleteMsg{MCPID: id, Scope: scope} }
		}
		return m, nil
	}
	switch key.Code {
	case tea.KeyEsc:
		return m, func() tea.Msg { return messages.NavigateBackMsg{} }
	case tea.KeyEnter, tea.KeySpace:
		scope := m.scope
		if m.list.Cursor == len(m.list.Items) {
			return m, func() tea.Msg { return messages.MCPManagerEditMsg{Scope: scope} }
		}
		item, ok := m.list.Items[m.list.Cursor].(mcpManagerItem)
		if !ok {
			return m, nil
		}
		if item.ID() == config.BuiltinMCPKankuro {
			return m, func() tea.Msg { return messages.ErrorMsg{Err: errors.New("the built-in kankuro MCP is read-only")} }
		}
		return m, func() tea.Msg { return messages.MCPManagerEditMsg{MCP: item.server, Scope: scope} }
	case tea.KeyDelete, tea.KeyBackspace:
		if m.deleting || m.list.Cursor >= len(m.list.Items) {
			return m, nil
		}
		item, ok := m.list.Items[m.list.Cursor].(mcpManagerItem)
		if !ok {
			return m, nil
		}
		if item.ID() == config.BuiltinMCPKankuro {
			return m, func() tea.Msg {
				return messages.ErrorMsg{Err: errors.New("the built-in kankuro MCP cannot be deleted")}
			}
		}
		m.pendingID = item.ID()
		const dialogNameMargin = 12
		name := components.TruncateHead(components.DisplayText(item.server.Name), max(1, m.Width-dialogNameMargin))
		m.modal.SetContent("Delete " + name + "?")
		m.modal.Open()
		return m, nil
	}
	if key.Code != tea.KeyUp && key.Code != tea.KeyDown && key.Text != "j" && key.Text != "k" {
		return m, nil
	}
	model, cmd := m.list.Update(msg)
	if list, ok := model.(*components.SourceList); ok {
		m.list = list
	}
	return m, cmd
}
func (m *MCPManager) DeleteFinished() { m.deleting = false; m.pendingID = ""; m.admitted = false }
func (m *MCPManager) View() string {
	if m.modal.IsOpen() {
		return m.modal.View()
	}
	return m.list.View()
}
func (m *MCPManager) SetSize(w, h int) {
	m.Base.SetSize(w, h)
	m.list.SetSize(w, h)
	m.modal.SetSize(max(1, w-4), h) //nolint:mnd // Reserve room for the modal border
}
