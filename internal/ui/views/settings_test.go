package views_test

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

func TestSettingsView_New(t *testing.T) {
	v := views.NewSettingsView(views.SettingsCallbacks{})
	if v == nil {
		t.Fatal("NewSettingsView returned nil")
	}
	if v.Mode() != views.SettingsModeDisplay {
		t.Error("should start in display mode")
	}
}

func TestSettingsView_SetSettings(t *testing.T) {
	v := views.NewSettingsView(views.SettingsCallbacks{})
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
	v := views.NewSettingsView(views.SettingsCallbacks{})
	v.SetSettings(config.AppSettings{IDECommand: "existing"})

	model, _ := v.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
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
		OnSave: func(settings config.AppSettings) tea.Cmd {
			saved = settings
			return func() tea.Msg { return nil }
		},
	})

	v.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	v.Input().SetValue("vim $PATH")

	model, _ := v.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	v = model.(*views.SettingsView)

	if v.Mode() != views.SettingsModeDisplay {
		t.Error("should return to display mode after save")
	}
	if saved.IDECommand != "vim $PATH" {
		t.Errorf("expected saved command %q, got %q", "vim $PATH", saved.IDECommand)
	}
}

func TestSettingsView_ToggleRows(t *testing.T) {
	v := views.NewSettingsView(views.SettingsCallbacks{})
	v.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	v.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	v.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	if v.FocusIndex() != 2 {
		t.Fatalf("expected focus on personas toggle, got %d", v.FocusIndex())
	}

	v.Update(tea.KeyPressMsg{Code: tea.KeySpace})
	if v.Settings().ShowPersonasInCardEditor {
		t.Error("expected personas toggle to turn off")
	}

	v.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	v.Update(tea.KeyPressMsg{Code: tea.KeySpace})
	if v.Settings().ShowMCPsInCardEditor {
		t.Error("expected MCP toggle to turn off")
	}
}

func TestSettingsView_EnterSavesFromToggleRow(t *testing.T) {
	var saved config.AppSettings

	v := views.NewSettingsView(views.SettingsCallbacks{
		OnSave: func(settings config.AppSettings) tea.Cmd {
			saved = settings
			return func() tea.Msg { return messages.ClearErrorMsg{} }
		},
	})
	v.SetSettings(config.AppSettings{
		IDECommand:               "code $PATH",
		ShowPersonasInCardEditor: true,
		ShowMCPsInCardEditor:     false,
	})

	v.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	v.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	v.Update(tea.KeyPressMsg{Code: tea.KeyTab})

	model, _ := v.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
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

func TestSettingsView_TerminalCommandFocusSaveAndCancel(t *testing.T) {
	var saved config.AppSettings
	v := views.NewSettingsView(views.SettingsCallbacks{OnSave: func(value config.AppSettings) tea.Cmd {
		saved = value
		return nil
	}})
	v.SetSettings(config.AppSettings{IDECommand: "ide", TerminalCommand: "terminal"})
	v.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	v.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	v.TerminalInput().SetValue(" xterm \"$PATH\" ")
	v.Update(tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	if v.Input().Value() != "ide" || v.TerminalInput().Value() != " xterm \"$PATH\" " {
		t.Fatalf("focus change lost command text: IDE %q terminal %q", v.Input().Value(), v.TerminalInput().Value())
	}
	v.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	v.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if saved.IDECommand != "ide" || saved.TerminalCommand != `xterm "$PATH"` {
		t.Fatalf("save did not preserve both commands: %+v", saved)
	}
	v.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	v.Input().SetValue("changed IDE")
	v.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	v.TerminalInput().SetValue("unsaved")
	v.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if v.TerminalInput().Value() != saved.TerminalCommand || v.Input().Value() != saved.IDECommand {
		t.Fatalf("cancel did not restore both commands: IDE %q terminal %q", v.Input().Value(), v.TerminalInput().Value())
	}
}

func TestSettingsView_LongCommandsStayWithinSize(t *testing.T) {
	command := strings.Repeat("é界\tline\n\x1b[31m", 50)
	for _, size := range []struct{ width, height int }{{80, 20}, {120, 36}} {
		v := views.NewSettingsView(views.SettingsCallbacks{})
		v.SetSettings(config.AppSettings{IDECommand: command, TerminalCommand: command})
		v.SetSize(size.width, size.height)
		assertSettingsFits(t, v, size.width, size.height, false)
		v.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		assertSettingsFits(t, v, size.width, size.height, true)
	}
}

func assertSettingsFits(t *testing.T, v *views.SettingsView, width, height int, editing bool) {
	t.Helper()
	output := v.View()
	if strings.Contains(output, "\x1b[31m") || strings.Contains(output, "\nline") {
		t.Fatal("display leaked ANSI or a command newline")
	}
	lines := strings.Split(output, "\n")
	if len(lines) > height {
		t.Errorf("%dx%d output has %d rows", width, height, len(lines))
	}
	for _, line := range lines {
		if lipgloss.Width(line) > width {
			t.Errorf("line width %d exceeds %d", lipgloss.Width(line), width)
		}
	}
	if !strings.Contains(output, "Manage Puppet Sources") || (editing && !strings.Contains(output, "Esc=cancel")) {
		t.Errorf("controls/help missing from %dx%d output", width, height)
	}
}

func TestSettingsView_CancelEdit(t *testing.T) {
	v := views.NewSettingsView(views.SettingsCallbacks{})
	v.SetSettings(config.AppSettings{IDECommand: "original"})
	v.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	v.Input().SetValue("changed")

	model, _ := v.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	v = model.(*views.SettingsView)

	if v.Mode() != views.SettingsModeDisplay {
		t.Error("should return to display mode on cancel")
	}
	if v.Settings().IDECommand != "original" {
		t.Error("settings should remain unchanged on cancel")
	}
}

func TestSettingsView_NavigateDashboard(t *testing.T) {
	v := views.NewSettingsView(views.SettingsCallbacks{})
	_, cmd := v.Update(tea.KeyPressMsg{Code: tea.KeyEsc})

	if cmd == nil {
		t.Fatal("expected NavigateMsg command")
	}
	msg := cmd()
	nav, ok := msg.(messages.NavigateMsg)
	if !ok {
		t.Fatalf("expected NavigateMsg, got %T", msg)
	}
	if nav.View != messages.ViewDashboard {
		t.Fatalf("expected ViewDashboard, got %v", nav.View)
	}
}

func TestSettingsView_View_DisplayMode(t *testing.T) {
	v := views.NewSettingsView(views.SettingsCallbacks{})
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
	v := views.NewSettingsView(views.SettingsCallbacks{})
	v.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
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
		OnSave: func(_ config.AppSettings) tea.Cmd {
			return func() tea.Msg { return messages.ErrorMsg{Err: errors.New("save failed")} }
		},
	})
	v.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	v.Input().SetValue("test")

	model, _ := v.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	v = model.(*views.SettingsView)

	if v.Mode() != views.SettingsModeDisplay {
		t.Error("should return to display mode even on error")
	}
}

func TestSettingsView_SetSize(t *testing.T) {
	v := views.NewSettingsView(views.SettingsCallbacks{})
	v.SetSize(100, 30)
	if v.Width != 100 {
		t.Errorf("expected width 100, got %d", v.Width)
	}
}
