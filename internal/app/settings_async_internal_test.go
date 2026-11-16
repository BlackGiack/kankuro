package app

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

type settingsReadOverride struct {
	configurationPort

	value contracts.Settings
	err   error
}

func (s settingsReadOverride) ReadSettings() (contracts.Settings, error) { return s.value, s.err }

func TestSettingsLateLoadPreservesDraft(t *testing.T) {
	c, _ := newLaunchFocusCoordinator(t)
	c.configuration = settingsReadOverride{configurationPort: c.configuration, value: contracts.Settings{IDECommand: "loaded"}}
	held := c.loadSettings()
	c.navigateTo(ViewSettings)
	settings := c.views[ViewSettings].(*views.SettingsView)
	c.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	settings.Input().SetValue("draft")
	c.Update(held())
	if settings.Input().Value() != "draft" || settings.Mode() != views.SettingsModeEdit {
		t.Fatal("late read replaced edit")
	}
}

func TestSettingsReadFailurePreservesAppliedValue(t *testing.T) {
	c, _ := newLaunchFocusCoordinator(t)
	c.settings.IDECommand = "prior"
	expected := errors.New("read failed")
	c.configuration = settingsReadOverride{configurationPort: c.configuration, err: expected}
	c.Update(c.loadSettings()())
	if c.settings.IDECommand != "prior" || !errors.Is(c.err, expected) {
		t.Fatalf("settings=%+v err=%v", c.settings, c.err)
	}
}

//nolint:gocognit // Verify both save and cancel against the loaded baseline.
func TestSettingsDelayedInitialReadPreservesUntouchedValues(t *testing.T) {
	for _, cancel := range []bool{false, true} {
		name := "save"
		if cancel {
			name = "cancel"
		}
		t.Run(name, func(t *testing.T) {
			c, _, repo := newLaunchFocusCoordinatorWithRepo(t)
			stored := config.AppSettings{
				IDECommand:               "old ide",
				TerminalCommand:          "old terminal",
				ShowPersonasInCardEditor: false,
				ShowMCPsInCardEditor:     false,
			}
			if err := repo.SaveAppSettings(stored); err != nil {
				t.Fatal(err)
			}
			delayed := c.loadSettings()()
			c.navigateTo(ViewSettings)
			c.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
			form := c.views[ViewSettings].(*views.SettingsView)
			form.Input().SetValue("new ide")
			c.Update(delayed)
			if form.Input().Value() != "new ide" || form.TerminalInput().Value() != stored.TerminalCommand {
				t.Fatal("late load lost edit or untouched terminal")
			}
			if cancel {
				c.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
				got := form.Settings()
				if got.IDECommand != stored.IDECommand || got.TerminalCommand != stored.TerminalCommand || got.ShowPersonasInCardEditor ||
					got.ShowMCPsInCardEditor {
					t.Fatalf("cancel baseline: %+v", got)
				}
				if form.Input().Value() != stored.IDECommand || form.TerminalInput().Value() != stored.TerminalCommand {
					t.Fatal("cancel did not restore loaded inputs")
				}
				return
			}
			_, save := c.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
			if save == nil {
				t.Fatal("missing save")
			}
			c.Update(save())
			got, err := repo.GetAppSettings()
			if err != nil {
				t.Fatal(err)
			}
			if got.IDECommand != "new ide" || got.TerminalCommand != stored.TerminalCommand || got.ShowPersonasInCardEditor ||
				got.ShowMCPsInCardEditor {
				t.Fatalf("saved settings: %+v", got)
			}
		})
	}
}

func TestSettingsDelayedReadPreservesExplicitlyEditedFields(t *testing.T) {
	c, _ := newLaunchFocusCoordinator(t)
	loaded := contracts.Settings{
		IDECommand:               "loaded",
		TerminalCommand:          "terminal",
		ShowPersonasInCardEditor: false,
		ShowMCPsInCardEditor:     false,
	}
	c.configuration = settingsReadOverride{configurationPort: c.configuration, value: loaded}
	delayed := c.loadSettings()()
	c.navigateTo(ViewSettings)
	c.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	form := c.views[ViewSettings].(*views.SettingsView)
	c.Update(tea.KeyPressMsg{Text: "x"})
	c.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	c.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	c.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	c.Update(tea.KeyPressMsg{Code: tea.KeySpace})
	c.Update(tea.KeyPressMsg{Code: tea.KeySpace})
	c.Update(delayed)
	if form.Input().Value() != "" || !form.Settings().ShowPersonasInCardEditor || form.Settings().ShowMCPsInCardEditor {
		t.Fatalf("edits not preserved: %+v ide=%q", form.Settings(), form.Input().Value())
	}
	if form.TerminalInput().Value() != "terminal" {
		t.Fatal("untouched field not updated")
	}
}

//nolint:gocognit // Exercise pending and completed reads for success and failure.
func TestSettingsSaveWaitsForInitialRead(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(fmt.Sprintf("read failure %v", fail), func(t *testing.T) {
			c, _, repo := newLaunchFocusCoordinatorWithRepo(t)
			stored := config.AppSettings{IDECommand: "stored", TerminalCommand: "retained"}
			if err := repo.SaveAppSettings(stored); err != nil {
				t.Fatal(err)
			}
			if fail {
				c.configuration = settingsReadOverride{configurationPort: c.configuration, err: errors.New("read failed")}
			}
			held := c.loadSettings()
			c.navigateTo(ViewSettings)
			c.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
			form := c.views[ViewSettings].(*views.SettingsView)
			form.Input().SetValue("draft")
			_, save := c.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
			if save != nil || form.Mode() != views.SettingsModeEdit || !strings.Contains(form.View(), "Loading settings") {
				t.Fatal("save was not blocked while reading")
			}
			before, err := repo.GetAppSettings()
			if err != nil || before != stored {
				t.Fatalf("premature save: %+v %v", before, err)
			}
			c.Update(held())
			if strings.Contains(form.View(), "Loading settings") || form.Input().Value() != "draft" {
				t.Fatal("read completion lost draft or retained feedback")
			}
			_, save = c.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
			if save == nil {
				t.Fatal("save remains blocked after read")
			}
		})
	}
}
