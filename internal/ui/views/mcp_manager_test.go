//nolint:testpackage // The tests inspect modal and selection state.
package views

import (
	"strconv"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/uitest"
)

func TestMCPManagerConfirmationKeepsTargetAcrossRefresh(t *testing.T) {
	m := NewMCPManager()
	m.SetSize(80, 22)
	m.SetData(messages.DataLoadedMsg{MCPs: []config.MCPServer{{ID: "a", Name: "Alpha"}, {ID: "b", Name: "Beta"}}})
	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyDelete})
	if cmd != nil || !m.modal.IsOpen() {
		t.Fatal("delete must open confirmation without mutation")
	}
	m.SetData(messages.DataLoadedMsg{MCPs: []config.MCPServer{{ID: "b", Name: "Beta"}, {ID: "a", Name: "Alpha"}}})
	_, cmd = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	got, ok := cmd().(messages.MCPManagerDeleteMsg)
	if !ok || got.MCPID != "a" {
		t.Fatalf("wrong deletion: %#v", got)
	}
	m.DeleteFinished()
	m.Update(tea.KeyPressMsg{Code: tea.KeyDelete})
	m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if m.modal.IsOpen() {
		t.Fatal("cancel did not close modal")
	}
	m.SetData(messages.DataLoadedMsg{})
	if len(m.list.Items) != 2 {
		t.Fatal("nil snapshot discarded list")
	}
}

func TestMCPManagerProtectedAndNew(t *testing.T) {
	m := NewMCPManager()
	m.SetSize(80, 22)
	m.SetData(messages.DataLoadedMsg{MCPs: []config.MCPServer{{ID: config.BuiltinMCPKankuro, Name: "Kankuro"}}})
	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if _, ok := cmd().(messages.ErrorMsg); !ok {
		t.Fatal("built-in must be read-only")
	}
	_, cmd = m.Update(tea.KeyPressMsg{Code: tea.KeyDelete})
	if _, ok := cmd().(messages.ErrorMsg); !ok {
		t.Fatal("built-in must be protected")
	}
	m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	_, cmd = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if msg, ok := cmd().(messages.MCPManagerEditMsg); !ok || msg.MCP.ID != "" {
		t.Fatal("new must open blank form")
	}
	_, cmd = m.Update(tea.KeyPressMsg{Code: tea.KeyDelete})
	if cmd != nil {
		t.Fatal("new cannot be deleted")
	}
	if !strings.Contains(m.View(), "[ New MCP ]") {
		t.Fatal("new row missing")
	}
}

func TestMCPManagerLongListFits(t *testing.T) {
	m := NewMCPManager()
	m.SetSize(80, 22)
	servers := make([]config.MCPServer, 40)
	for i := range servers {
		servers[i] = config.MCPServer{ID: strconv.Itoa(i), Name: strings.Repeat("漢", 100) + "\x1b[31m"}
	}
	m.SetData(messages.DataLoadedMsg{MCPs: servers})
	for range servers {
		uitest.AssertWithin(t, m.View(), 80, 22, "MCP list")
		m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	}
	if !strings.Contains(m.View(), "[ New MCP ]") {
		t.Fatal("new row unreachable")
	}
	m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	m.Update(tea.KeyPressMsg{Code: tea.KeyDelete})
	uitest.AssertWithin(t, m.View(), 80, 22, "MCP confirmation")
}
