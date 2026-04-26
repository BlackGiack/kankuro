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

func TestIDEConfig_New(t *testing.T) {
	callbacks := views.IDEConfigCallbacks{}
	c := views.NewIDEConfig(callbacks, &commands.Commands{})

	if c == nil {
		t.Fatal("NewIDEConfig returned nil")
	}

	if c.Mode() != views.IDEModeDisplay {
		t.Error("Should start in display mode")
	}
}

func TestIDEConfig_SetConfig(t *testing.T) {
	c := views.NewIDEConfig(views.IDEConfigCallbacks{}, &commands.Commands{})

	cfg := config.IDEConfig{Command: "code $PATH"}
	c.SetConfig(cfg)

	if c.Config().Command != "code $PATH" {
		t.Errorf("Expected 'code $PATH', got '%s'", c.Config().Command)
	}

	if c.Input().Value() != "code $PATH" {
		t.Error("Input should be populated with config command")
	}
}

func TestIDEConfig_SwitchToEditMode(t *testing.T) {
	c := views.NewIDEConfig(views.IDEConfigCallbacks{}, &commands.Commands{})
	c.SetConfig(config.IDEConfig{Command: "existing"})

	model, _ := c.Update(tea.KeyMsg{Type: tea.KeyEnter})
	c = model.(*views.IDEConfig)

	if c.Mode() != views.IDEModeEdit {
		t.Error("Should be in edit mode after Enter")
	}

	if c.Input().Value() != "existing" {
		t.Error("Input should have existing command")
	}
}

func TestIDEConfig_SwitchToEditMode_EKey(t *testing.T) {
	c := views.NewIDEConfig(views.IDEConfigCallbacks{}, &commands.Commands{})

	model, _ := c.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	c = model.(*views.IDEConfig)

	if c.Mode() != views.IDEModeEdit {
		t.Error("Should be in edit mode after 'e'")
	}
}

func TestIDEConfig_SaveInEditMode(t *testing.T) {
	var savedCommand string

	callbacks := views.IDEConfigCallbacks{
		OnSave: func(cmd string) (tea.Cmd, error) {
			savedCommand = cmd
			return func() tea.Msg { return nil }, errTest
		},
	}

	c := views.NewIDEConfig(callbacks, &commands.Commands{})
	// We need a way to set mode to edit. I'll use Update.
	c.Update(tea.KeyMsg{Type: tea.KeyEnter})
	c.Input().SetValue("vim $PATH")
	c.Input().Focus()

	model, _ := c.Update(tea.KeyMsg{Type: tea.KeyEnter})
	c = model.(*views.IDEConfig)

	if c.Mode() != views.IDEModeDisplay {
		t.Error("Should return to display mode after save")
	}

	if savedCommand != "vim $PATH" {
		t.Errorf("Expected saved command 'vim $PATH', got '%s'", savedCommand)
	}

	if c.Config().Command != "vim $PATH" {
		t.Error("Config should be updated")
	}
}

func TestIDEConfig_CancelEdit(t *testing.T) {
	c := views.NewIDEConfig(views.IDEConfigCallbacks{}, &commands.Commands{})
	c.SetConfig(config.IDEConfig{Command: "original"})
	c.Update(tea.KeyMsg{Type: tea.KeyEnter})
	c.Input().SetValue("changed")
	c.Input().Focus()

	model, _ := c.Update(tea.KeyMsg{Type: tea.KeyEsc})
	c = model.(*views.IDEConfig)

	if c.Mode() != views.IDEModeDisplay {
		t.Error("Should return to display mode on cancel")
	}

	if c.Config().Command != "original" {
		t.Error("Config should remain unchanged on cancel")
	}
}

func TestIDEConfig_NavigateBack(t *testing.T) {
	c := views.NewIDEConfig(views.IDEConfigCallbacks{}, &commands.Commands{})
	// Mode is display by default

	_, cmd := c.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if cmd == nil {
		t.Error("Should return a command on Esc in display mode")
	}

	msg := cmd()
	if _, ok := msg.(messages.NavigateBackMsg); !ok {
		t.Errorf("Should emit NavigateBackMsg, got %T", msg)
	}
}

func TestIDEConfig_View_DisplayMode(t *testing.T) {
	c := views.NewIDEConfig(views.IDEConfigCallbacks{}, &commands.Commands{})
	c.SetConfig(config.IDEConfig{Command: "code $PATH"})
	c.SetSize(80, 24)

	view := c.View()

	if view == "" {
		t.Error("View should not be empty")
	}

	if !strings.Contains(view, "IDE Configuration") {
		t.Error("Should contain title")
	}

	if !strings.Contains(view, "code $PATH") {
		t.Error("Should show current command")
	}

	if !strings.Contains(view, "$PATH") {
		t.Error("Should show placeholder help")
	}
}

func TestIDEConfig_View_EditMode(t *testing.T) {
	c := views.NewIDEConfig(views.IDEConfigCallbacks{}, &commands.Commands{})
	c.Update(tea.KeyMsg{Type: tea.KeyEnter})
	c.Input().SetValue("vim $PATH")
	c.SetSize(80, 24)

	view := c.View()

	if view == "" {
		t.Error("View should not be empty")
	}

	if !strings.Contains(view, "Edit IDE Command") {
		t.Error("Should contain edit title")
	}

	if !strings.Contains(view, "vim $PATH") {
		t.Error("Should show input value")
	}
}

func TestIDEConfig_View_NotConfigured(t *testing.T) {
	c := views.NewIDEConfig(views.IDEConfigCallbacks{}, &commands.Commands{})

	view := c.View()

	if !strings.Contains(view, "not configured") {
		t.Error("Should show 'not configured' when empty")
	}
}

func TestIDEConfig_SetSize(t *testing.T) {
	c := views.NewIDEConfig(views.IDEConfigCallbacks{}, &commands.Commands{})
	c.SetSize(100, 30)

	// Since width is private in Base, I can't check c.Width easily if it's not exported.
	// But component.Base has Width exported!
	// Wait, internal/ui/component/component.go has Base.Width exported.

	if c.Width != 100 {
		t.Errorf("Expected width 100, got %d", c.Width)
	}

	// inputMaxWidth is private, I can't check it directly.
	// I'll just skip the check or use the hardcoded value if I know it.
}

func TestIDEConfig_OnSaveError(t *testing.T) {
	callbacks := views.IDEConfigCallbacks{
		OnSave: func(_ string) (tea.Cmd, error) {
			return nil, errors.New("save failed")
		},
	}

	c := views.NewIDEConfig(callbacks, &commands.Commands{})
	c.Update(tea.KeyMsg{Type: tea.KeyEnter})
	c.Input().SetValue("test")

	model, _ := c.Update(tea.KeyMsg{Type: tea.KeyEnter})
	c = model.(*views.IDEConfig)

	if c.Mode() != views.IDEModeDisplay {
		t.Error("Should return to display mode even on error")
	}
}

func TestIDEConfig_EscNavigatesBack(t *testing.T) {
	callbacks := views.IDEConfigCallbacks{}

	c := views.NewIDEConfig(callbacks, &commands.Commands{})
	// Mode is display by default

	model, cmd := c.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if cmd == nil {
		t.Error("Expected NavigateBackMsg command")
		return
	}

	// Execute the command to get the message
	msg := cmd()
	if _, ok := msg.(messages.NavigateBackMsg); !ok {
		t.Errorf("Expected NavigateBackMsg, got %T", msg)
	}

	_ = model
}

func TestIDEConfig_Quit(t *testing.T) {
	c := views.NewIDEConfig(views.IDEConfigCallbacks{}, &commands.Commands{})

	_, cmd := c.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	if cmd == nil {
		t.Error("Should return quit command")
	}
}

func TestIDEConfig_QuitInEditMode(t *testing.T) {
	c := views.NewIDEConfig(views.IDEConfigCallbacks{}, &commands.Commands{})
	c.Update(tea.KeyMsg{Type: tea.KeyEnter})

	_, cmd := c.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	if cmd == nil {
		t.Error("Should return quit command")
	}
}
