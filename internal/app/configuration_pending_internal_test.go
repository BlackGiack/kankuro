package app

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

type pendingSaveOverride struct {
	configurationPort

	settingsErr, mcpErr error
}

func (p pendingSaveOverride) SaveSettings(v contracts.Settings) error {
	if p.settingsErr != nil {
		return p.settingsErr
	}
	return p.configurationPort.SaveSettings(v)
}
func (p pendingSaveOverride) SaveMCP(v contracts.MCPEditor) (contracts.MCPEditor, error) {
	if p.mcpErr != nil {
		return contracts.MCPEditor{}, p.mcpErr
	}
	return p.configurationPort.SaveMCP(v)
}

//nolint:gocognit // Each visit and outcome must preserve its independent state.
func TestSettingsSaveCompletionRespectsVisitAndDraft(t *testing.T) {
	for _, fail := range []bool{false, true} {
		for _, destination := range []string{"source", "reopened", "new draft"} {
			t.Run(destination+map[bool]string{true: " failure", false: " success"}[fail], func(t *testing.T) {
				c, _ := newLaunchFocusCoordinator(t)
				failure := errors.New("save failed")
				override := pendingSaveOverride{configurationPort: c.configuration}
				if fail {
					override.settingsErr = failure
				}
				c.configuration = override
				c.navigateTo(ViewSettings)
				c.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
				settings := c.views[ViewSettings].(*views.SettingsView)
				settings.Input().SetValue("first")
				_, held := c.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
				if held == nil {
					t.Fatal("missing save")
				}
				_, openSources := c.Update(tea.KeyPressMsg{Text: "p"})
				if openSources == nil {
					t.Fatal("missing puppet source navigation")
				}
				c.Update(openSources())
				if c.currentView != ViewSourceManagerPuppets {
					t.Fatal("puppet sources did not open")
				}
				if destination != "source" {
					c.navigateTo(ViewSettings)
				}
				if destination == "new draft" {
					c.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
					settings.Input().SetValue("new draft")
					_, blocked := c.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
					if blocked != nil || !strings.Contains(settings.View(), "save is still in progress") {
						t.Fatal("pending save was not explained")
					}
				}
				marker := errors.New("current view error")
				c.err = marker
				c.Update(held())
				wantView := ViewSettings
				if destination == "source" {
					wantView = ViewSourceManagerPuppets
				}
				if c.currentView != wantView || !errors.Is(c.err, marker) {
					t.Fatalf("old result changed view/feedback: view=%v err=%v", c.currentView, c.err)
				}
				if c.settingsSaving {
					t.Fatal("pending save not released")
				}
				if !fail && c.settings.IDECommand != "first" {
					t.Fatal("saved global settings not applied")
				}
				if destination == "new draft" {
					if settings.Mode() != views.SettingsModeEdit || settings.Input().Value() != "new draft" {
						t.Fatal("old result lost newer draft")
					}
					if strings.Contains(settings.View(), "save is still in progress") {
						t.Fatal("stale pending feedback remained")
					}
					_, retry := c.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
					if retry == nil {
						t.Fatal("new draft could not be saved")
					}
				}
			})
		}
	}
}

func TestMCPSavePendingBlocksEditingAndRecoversOnFailure(t *testing.T) {
	c, _ := newLaunchFocusCoordinator(t)
	failure := errors.New("save failed")
	c.configuration = pendingSaveOverride{configurationPort: c.configuration, mcpErr: failure}
	c.navigateTo(ViewMCPForm)
	form := c.views[ViewMCPForm].(*views.MCPForm)
	form.Inputs()[0].SetValue("first")
	form.Inputs()[1].SetValue("echo")
	_, held := c.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if held == nil {
		t.Fatal("missing save")
	}
	c.Update(tea.KeyPressMsg{Text: "x"})
	c.Update(tea.PasteMsg{Content: "pasted"})
	_, again := c.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if again != nil || form.Inputs()[0].Value() != "first" || !strings.Contains(form.View(), "Saving MCP") {
		t.Fatal("pending form accepted input or hid status")
	}
	c.Update(held())
	if !errors.Is(c.err, failure) || c.currentView != ViewMCPForm {
		t.Fatal("failure was not shown")
	}
	c.Update(tea.KeyPressMsg{Text: "x"})
	if form.Inputs()[0].Value() == "first" {
		t.Fatal("failed save left form locked")
	}
	c.configuration = pendingSaveOverride{configurationPort: c.configuration.(pendingSaveOverride).configurationPort}
	_, retry := c.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if retry == nil {
		t.Fatal("retry not admitted")
	}
}

//nolint:gocognit // Exercise both pending keys through failure and retry.
func TestSettingsPendingBlocksEditAndAllowsRetryAfterFailure(t *testing.T) {
	for _, key := range []struct {
		name string
		msg  tea.KeyPressMsg
	}{
		{"enter", tea.KeyPressMsg{Code: tea.KeyEnter}},
		{"edit", tea.KeyPressMsg{Text: "e"}},
	} {
		t.Run(key.name, func(t *testing.T) {
			c, _ := newLaunchFocusCoordinator(t)
			failure := errors.New("settings save failed")
			base := c.configuration
			c.configuration = pendingSaveOverride{configurationPort: base, settingsErr: failure}
			c.navigateTo(ViewSettings)
			c.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
			form := c.views[ViewSettings].(*views.SettingsView)
			form.Input().SetValue("first")
			_, held := c.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
			if held == nil {
				t.Fatal("missing save")
			}
			_, again := c.Update(key.msg)
			if again != nil || form.Mode() != views.SettingsModeDisplay || !strings.Contains(form.View(), "Saving settings") {
				t.Fatal("pending save permitted another edit or hid status")
			}
			c.Update(held())
			if !errors.Is(c.err, failure) {
				t.Fatal("missing save error")
			}
			if strings.Contains(form.View(), "Saving settings") || strings.Contains(form.View(), "save is still in progress") {
				t.Fatal("completed save retained pending feedback")
			}
			if form.Input().Value() != "first" {
				t.Fatal("failed save lost draft")
			}
			c.configuration = base
			c.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
			if form.Mode() != views.SettingsModeEdit {
				t.Fatal("failed save left settings locked")
			}
			_, retry := c.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
			if retry == nil {
				t.Fatal("retry not admitted")
			}
		})
	}
}

func TestMCPSavePendingCommitsOnlyDisplayedDraft(t *testing.T) {
	c, _ := newLaunchFocusCoordinator(t)
	c.navigateTo(ViewMCPForm)
	form := c.views[ViewMCPForm].(*views.MCPForm)
	form.Inputs()[0].SetValue("first")
	form.Inputs()[1].SetValue("echo")
	_, held := c.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if held == nil {
		t.Fatal("missing save")
	}
	c.Update(tea.KeyPressMsg{Text: "x"})
	_, duplicate := c.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if duplicate != nil || form.Inputs()[0].Value() != "first" {
		t.Fatal("pending edit or duplicate submission was accepted")
	}
	c.Update(held())
	rows, err := c.configuration.ReadMCPRows()
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, row := range rows {
		if row.Name == "first" {
			found = true
		}
		if row.Name == "firstx" {
			t.Fatal("blocked input was persisted")
		}
	}
	if !found || c.currentView != ViewDashboard {
		t.Fatalf("successful save: found=%v view=%v", found, c.currentView)
	}
}
