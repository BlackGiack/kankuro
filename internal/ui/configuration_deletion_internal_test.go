package ui

import (
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

type deletingConfiguration struct {
	configurationPort

	ids []string
	err error
}

func (d *deletingConfiguration) DeleteConfiguredMCP(id string) error {
	d.ids = append(d.ids, id)
	return d.err
}

//nolint:gocognit // Both entry points must propagate success and failure.
func TestMCPDeletionUsesConfigurationCapability(t *testing.T) {
	for _, manager := range []bool{false, true} {
		for _, fail := range []bool{false, true} {
			name := "editor"
			if manager {
				name = "manager"
			}
			if fail {
				name += " failure"
			}
			t.Run(name, func(t *testing.T) {
				c, _ := newLaunchFocusCoordinator(t)
				fake := &deletingConfiguration{configurationPort: c.configuration}
				if fail {
					fake.err = errors.New("delete denied")
				}
				c.configuration = fake
				var cmd tea.Cmd
				if !manager {
					_, cmd = c.Update(messages.MCPDeleteRequestedMsg{MCPID: "custom"})
				} else {
					c.navigateTo(ViewMCPManager)
					view := c.views[ViewMCPManager].(*views.MCPManager)
					view.SetData(
						messages.DataLoadedMsg{
							MCPRows: []contracts.MCPRow{{ID: "custom", Name: "Custom", Delete: contracts.Availability{Enabled: true}}},
						},
					)
					c.Update(tea.KeyPressMsg{Code: tea.KeyDelete})
					_, confirm := c.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
					if confirm == nil {
						t.Fatal("missing confirmation")
					}
					_, cmd = c.Update(confirm())
				}
				if cmd == nil {
					t.Fatal("missing deletion command")
				}
				if len(fake.ids) != 0 {
					t.Fatal("deletion ran synchronously")
				}
				replacement := &deletingConfiguration{configurationPort: fake.configurationPort}
				c.configuration = replacement
				result := cmd()
				if len(replacement.ids) != 0 {
					t.Fatal("command used replacement capability")
				}
				if len(fake.ids) != 1 || fake.ids[0] != "custom" {
					t.Fatalf("calls: %v", fake.ids)
				}
				if manager {
					m, ok := result.(messages.MCPManagerDeletedMsg)
					if !ok || !errors.Is(m.Err, fake.err) {
						t.Fatalf("result: %#v", result)
					}
				} else if fail {
					m, ok := result.(messages.ErrorMsg)
					if !ok || !errors.Is(m.Err, fake.err) {
						t.Fatalf("result: %#v", result)
					}
				} else if _, ok := result.(messages.MCPDeletedMsg); !ok {
					t.Fatalf("result: %#v", result)
				}
			})
		}
	}
}
