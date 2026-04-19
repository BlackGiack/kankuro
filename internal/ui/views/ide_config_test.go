package views

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"kankuro/internal/config"
	"kankuro/internal/ui/commands"
	"kankuro/internal/ui/messages"
)

func TestIDEConfig_New(t *testing.T) {
	callbacks := IDEConfigCallbacks{}
	c := NewIDEConfig(callbacks, &commands.Commands{})

	if c == nil {
		t.Fatal("NewIDEConfig returned nil")
	}

	if c.mode != IDEModeDisplay {
		t.Error("Should start in display mode")
	}
}

func TestIDEConfig_SetConfig(t *testing.T) {
	c := NewIDEConfig(IDEConfigCallbacks{}, &commands.Commands{})

	cfg := config.IDEConfig{Command: "code $PATH"}
	c.SetConfig(cfg)

	if c.config.Command != "code $PATH" {
		t.Errorf("Expected 'code $PATH', got '%s'", c.config.Command)
	}

	if c.input.Value() != "code $PATH" {
		t.Error("Input should be populated with config command")
	}
}

func TestIDEConfig_SwitchToEditMode(t *testing.T) {
	c := NewIDEConfig(IDEConfigCallbacks{}, &commands.Commands{})
	c.SetConfig(config.IDEConfig{Command: "existing"})

	model, _ := c.Update(tea.KeyMsg{Type: tea.KeyEnter})
	c = model.(*IDEConfig)

	if c.mode != IDEModeEdit {
		t.Error("Should be in edit mode after Enter")
	}

	if c.input.Value() != "existing" {
		t.Error("Input should have existing command")
	}
}

func TestIDEConfig_SwitchToEditMode_EKey(t *testing.T) {
	c := NewIDEConfig(IDEConfigCallbacks{}, &commands.Commands{})

	model, _ := c.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	c = model.(*IDEConfig)

	if c.mode != IDEModeEdit {
		t.Error("Should be in edit mode after 'e'")
	}
}

func TestIDEConfig_SaveInEditMode(t *testing.T) {
	var savedCommand string

	callbacks := IDEConfigCallbacks{
		OnSave: func(cmd string) error {
			savedCommand = cmd
			return nil
		},
	}

	c := NewIDEConfig(callbacks, &commands.Commands{})
	c.mode = IDEModeEdit
	c.input.SetValue("vim $PATH")
	c.input.Focus()

	model, _ := c.Update(tea.KeyMsg{Type: tea.KeyEnter})
	c = model.(*IDEConfig)

	if c.mode != IDEModeDisplay {
		t.Error("Should return to display mode after save")
	}

	if savedCommand != "vim $PATH" {
		t.Errorf("Expected saved command 'vim $PATH', got '%s'", savedCommand)
	}

	if c.config.Command != "vim $PATH" {
		t.Error("Config should be updated")
	}
}

func TestIDEConfig_CancelEdit(t *testing.T) {
	c := NewIDEConfig(IDEConfigCallbacks{}, &commands.Commands{})
	c.SetConfig(config.IDEConfig{Command: "original"})
	c.mode = IDEModeEdit
	c.input.SetValue("changed")
	c.input.Focus()

	model, _ := c.Update(tea.KeyMsg{Type: tea.KeyEsc})
	c = model.(*IDEConfig)

	if c.mode != IDEModeDisplay {
		t.Error("Should return to display mode on cancel")
	}

	if c.config.Command != "original" {
		t.Error("Config should remain unchanged on cancel")
	}
}

func TestIDEConfig_NavigateBack(t *testing.T) {
	c := NewIDEConfig(IDEConfigCallbacks{}, &commands.Commands{})
	c.mode = IDEModeDisplay

	_, cmd := c.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if cmd == nil {
		t.Error("Should return a command on Esc in display mode")
	}

	msg := cmd()
	if navMsg, ok := msg.(messages.NavigateMsg); ok {
		if navMsg.View != messages.ViewDashboard {
			t.Errorf("Should navigate to ViewDashboard, got %d", navMsg.View)
		}
	} else {
		t.Error("Should emit NavigateMsg")
	}
}

func TestIDEConfig_View_DisplayMode(t *testing.T) {
	c := NewIDEConfig(IDEConfigCallbacks{}, &commands.Commands{})
	c.SetConfig(config.IDEConfig{Command: "code $PATH"})
	c.SetSize(80, 24)

	view := c.View()

	if view == "" {
		t.Error("View should not be empty")
	}

	if !containsString(view, "IDE Configuration") {
		t.Error("Should contain title")
	}

	if !containsString(view, "code $PATH") {
		t.Error("Should show current command")
	}

	if !containsString(view, "$PATH") {
		t.Error("Should show placeholder help")
	}
}

func TestIDEConfig_View_EditMode(t *testing.T) {
	c := NewIDEConfig(IDEConfigCallbacks{}, &commands.Commands{})
	c.mode = IDEModeEdit
	c.input.SetValue("vim $PATH")
	c.SetSize(80, 24)

	view := c.View()

	if view == "" {
		t.Error("View should not be empty")
	}

	if !containsString(view, "Edit IDE Command") {
		t.Error("Should contain edit title")
	}

	if !containsString(view, "vim $PATH") {
		t.Error("Should show input value")
	}
}

func TestIDEConfig_View_NotConfigured(t *testing.T) {
	c := NewIDEConfig(IDEConfigCallbacks{}, &commands.Commands{})

	view := c.View()

	if !containsString(view, "not configured") {
		t.Error("Should show 'not configured' when empty")
	}
}

func TestIDEConfig_SetSize(t *testing.T) {
	c := NewIDEConfig(IDEConfigCallbacks{}, &commands.Commands{})
	c.SetSize(100, 30)

	if c.Width != 100 {
		t.Errorf("Expected width 100, got %d", c.Width)
	}

	if c.input.Width > inputMaxWidth {
		t.Errorf("Input width should be capped at %d, got %d", inputMaxWidth, c.input.Width)
	}
}

func TestIDEConfig_OnSaveError(t *testing.T) {
	callbacks := IDEConfigCallbacks{
		OnSave: func(_ string) error {
			return errors.New("save failed")
		},
	}

	c := NewIDEConfig(callbacks, &commands.Commands{})
	c.mode = IDEModeEdit
	c.input.SetValue("test")

	model, _ := c.Update(tea.KeyMsg{Type: tea.KeyEnter})
	c = model.(*IDEConfig)

	if c.mode != IDEModeDisplay {
		t.Error("Should return to display mode even on error")
	}
}

func TestIDEConfig_OnClose(t *testing.T) {
	closed := false

	callbacks := IDEConfigCallbacks{
		OnClose: func() {
			closed = true
		},
	}

	c := NewIDEConfig(callbacks, &commands.Commands{})
	c.mode = IDEModeDisplay

	_, _ = c.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if !closed {
		t.Error("OnClose should be called")
	}
}

func TestIDEConfig_Quit(t *testing.T) {
	c := NewIDEConfig(IDEConfigCallbacks{}, &commands.Commands{})

	_, cmd := c.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	if cmd == nil {
		t.Error("Should return quit command")
	}
}

func TestIDEConfig_QuitInEditMode(t *testing.T) {
	c := NewIDEConfig(IDEConfigCallbacks{}, &commands.Commands{})
	c.mode = IDEModeEdit

	_, cmd := c.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	if cmd == nil {
		t.Error("Should return quit command")
	}
}
