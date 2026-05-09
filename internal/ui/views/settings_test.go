package views_test

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"kankuro/internal/config"
	"kankuro/internal/ui/commands"
	"kankuro/internal/ui/messages"
	"kankuro/internal/ui/views"
)

func TestSettingsView_New(t *testing.T) {
	v := views.NewSettingsView(views.SettingsCallbacks{}, &commands.Commands{})
	if v == nil {
		t.Fatal("NewSettingsView returned nil")
	}
	if v.Mode() != views.SettingsModeDisplay {
		t.Error("should start in display mode")
	}
}

func TestSettingsView_SetSettings(t *testing.T) {
	v := views.NewSettingsView(views.SettingsCallbacks{}, &commands.Commands{})
	cfg := config.AppSettings{
		IDECommand:               "code $PATH",
		ShowPersonasInCardEditor: false,
		ShowMCPsInCardEditor:     true,
	}
	v.SetSettings(cfg)

	if v.Settings() != cfg {
		t.Errorf("expected settings %+v, got %+v", cfg, v.Settings())
	}
	if v.Input().Value() != "code $PATH" {
		t.Error("input should be populated with the IDE command")
	}
}

func TestSettingsView_SwitchToEditMode(t *testing.T) {
	v := views.NewSettingsView(views.SettingsCallbacks{}, &commands.Commands{})
	v.SetSettings(config.AppSettings{IDECommand: "existing"})

	model, _ := v.Update(tea.KeyMsg{Type: tea.KeyEnter})
	v = model.(*views.SettingsView)

	if v.Mode() != views.SettingsModeEdit {
		t.Error("should be in edit mode after Enter")
	}
	if v.Input().Value() != "existing" {
		t.Error("input should have existing command")
	}
}

func TestSettingsView_SaveInEditMode(t *testing.T) {
	var saved config.AppSettings

	v := views.NewSettingsView(views.SettingsCallbacks{
		OnSave: func(settings config.AppSettings) (tea.Cmd, error) {
			saved = settings
			return func() tea.Msg { return nil }, nil
		},
	}, &commands.Commands{})

	v.Update(tea.KeyMsg{Type: tea.KeyEnter})
	v.Input().SetValue("vim $PATH")

	model, _ := v.Update(tea.KeyMsg{Type: tea.KeyEnter})
	v = model.(*views.SettingsView)

	if v.Mode() != views.SettingsModeDisplay {
		t.Error("should return to display mode after save")
	}
	if saved.IDECommand != "vim $PATH" {
		t.Errorf("expected saved command %q, got %q", "vim $PATH", saved.IDECommand)
	}
}

func TestSettingsView_ToggleRows(t *testing.T) {
	v := views.NewSettingsView(views.SettingsCallbacks{}, &commands.Commands{})
	v.Update(tea.KeyMsg{Type: tea.KeyEnter})

	v.Update(tea.KeyMsg{Type: tea.KeyTab})
	if v.FocusIndex() != 1 {
		t.Fatalf("expected focus on personas toggle, got %d", v.FocusIndex())
	}

	v.Update(tea.KeyMsg{Type: tea.KeySpace})
	if v.Settings().ShowPersonasInCardEditor {
		t.Error("expected personas toggle to turn off")
	}

	v.Update(tea.KeyMsg{Type: tea.KeyTab})
	v.Update(tea.KeyMsg{Type: tea.KeySpace})
	if v.Settings().ShowMCPsInCardEditor {
		t.Error("expected MCP toggle to turn off")
	}
}

func TestSettingsView_EnterSavesFromToggleRow(t *testing.T) {
	var saved config.AppSettings

	v := views.NewSettingsView(views.SettingsCallbacks{
		OnSave: func(settings config.AppSettings) (tea.Cmd, error) {
			saved = settings
			return func() tea.Msg { return messages.ClearErrorMsg{} }, nil
		},
	}, &commands.Commands{})
	v.SetSettings(config.AppSettings{
		IDECommand:               "code $PATH",
		ShowPersonasInCardEditor: true,
		ShowMCPsInCardEditor:     false,
	})

	v.Update(tea.KeyMsg{Type: tea.KeyEnter})
	v.Update(tea.KeyMsg{Type: tea.KeyTab})

	model, _ := v.Update(tea.KeyMsg{Type: tea.KeyEnter})
	v = model.(*views.SettingsView)

	if v.Mode() != views.SettingsModeDisplay {
		t.Error("expected Enter on toggle row to save and exit edit mode")
	}
	if !saved.ShowPersonasInCardEditor {
		t.Error("expected Enter to save current toggle state without flipping it")
	}
	if saved.IDECommand != "code $PATH" {
		t.Errorf("expected IDE command to be preserved, got %q", saved.IDECommand)
	}
}

func TestSettingsView_CancelEdit(t *testing.T) {
	v := views.NewSettingsView(views.SettingsCallbacks{}, &commands.Commands{})
	v.SetSettings(config.AppSettings{IDECommand: "original"})
	v.Update(tea.KeyMsg{Type: tea.KeyEnter})
	v.Input().SetValue("changed")

	model, _ := v.Update(tea.KeyMsg{Type: tea.KeyEsc})
	v = model.(*views.SettingsView)

	if v.Mode() != views.SettingsModeDisplay {
		t.Error("should return to display mode on cancel")
	}
	if v.Settings().IDECommand != "original" {
		t.Error("settings should remain unchanged on cancel")
	}
}

func TestSettingsView_NavigateBack(t *testing.T) {
	v := views.NewSettingsView(views.SettingsCallbacks{}, &commands.Commands{})
	_, cmd := v.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if cmd == nil {
		t.Fatal("expected NavigateBackMsg command")
	}
	if _, ok := cmd().(messages.NavigateBackMsg); !ok {
		t.Fatalf("expected NavigateBackMsg, got %T", cmd())
	}
}

func TestSettingsView_View_DisplayMode(t *testing.T) {
	v := views.NewSettingsView(views.SettingsCallbacks{}, &commands.Commands{})
	v.SetSettings(config.AppSettings{
		IDECommand:               "code $PATH",
		ShowPersonasInCardEditor: true,
		ShowMCPsInCardEditor:     false,
	})
	v.SetSize(80, 24)

	view := v.View()
	if !strings.Contains(view, "Settings") {
		t.Error("should contain title")
	}
	if !strings.Contains(view, "code $PATH") {
		t.Error("should show current command")
	}
	if !strings.Contains(view, "Show Personas In Card Editor: on") {
		t.Error("should show personas status")
	}
	if !strings.Contains(view, "Show MCPs In Card Editor: off") {
		t.Error("should show MCP status")
	}
}

func TestSettingsView_View_EditMode(t *testing.T) {
	v := views.NewSettingsView(views.SettingsCallbacks{}, &commands.Commands{})
	v.Update(tea.KeyMsg{Type: tea.KeyEnter})
	v.Input().SetValue("vim $PATH")
	v.SetSize(80, 24)

	view := v.View()
	if !strings.Contains(view, "Edit Settings") {
		t.Error("should contain edit title")
	}
	if !strings.Contains(view, "vim $PATH") {
		t.Error("should show input value")
	}
	if !strings.Contains(view, "Show Personas In Card Editor") {
		t.Error("should show personas toggle")
	}
}

func TestSettingsView_OnSaveError(t *testing.T) {
	v := views.NewSettingsView(views.SettingsCallbacks{
		OnSave: func(_ config.AppSettings) (tea.Cmd, error) {
			return nil, errors.New("save failed")
		},
	}, &commands.Commands{})
	v.Update(tea.KeyMsg{Type: tea.KeyEnter})
	v.Input().SetValue("test")

	model, _ := v.Update(tea.KeyMsg{Type: tea.KeyEnter})
	v = model.(*views.SettingsView)

	if v.Mode() != views.SettingsModeDisplay {
		t.Error("should return to display mode even on error")
	}
}

func TestSettingsView_SetSize(t *testing.T) {
	v := views.NewSettingsView(views.SettingsCallbacks{}, &commands.Commands{})
	v.SetSize(100, 30)
	if v.Width != 100 {
		t.Errorf("expected width 100, got %d", v.Width)
	}
}
