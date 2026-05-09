package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"kankuro/internal/config"
	"kankuro/internal/ui/commands"
	"kankuro/internal/ui/component"
	"kankuro/internal/ui/messages"
)

const (
	settingsBoxPadding      = 2
	settingsBoxMinWidth     = 30
	settingsExplanationWide = 40
)

type SettingsMode int

const (
	SettingsModeDisplay SettingsMode = iota
	SettingsModeEdit
)

const (
	settingsFocusIDECommand = iota
	settingsFocusPersonas
	settingsFocusMCPs
	settingsFocusCount
)

type SettingsCallbacks struct {
	OnSave func(settings config.AppSettings) (tea.Cmd, error)
}

type SettingsView struct {
	component.Base

	mode      SettingsMode
	focus     int
	input     textinput.Model
	settings  config.AppSettings
	callbacks SettingsCallbacks
	cmds      *commands.Commands
}

func NewSettingsView(callbacks SettingsCallbacks, cmds *commands.Commands) *SettingsView {
	input := textinput.New()
	input.Placeholder = "e.g. code $PATH"
	input.Width = 60

	return &SettingsView{
		mode:      SettingsModeDisplay,
		input:     input,
		settings:  config.DefaultAppSettings(),
		callbacks: callbacks,
		cmds:      cmds,
	}
}

func (s *SettingsView) SetSettings(cfg config.AppSettings) {
	s.settings = cfg
	s.input.SetValue(cfg.IDECommand)
}

func (s *SettingsView) Settings() config.AppSettings {
	return s.settings
}

func (s *SettingsView) Mode() SettingsMode {
	return s.mode
}

func (s *SettingsView) Input() *textinput.Model {
	return &s.input
}

func (s *SettingsView) FocusIndex() int {
	return s.focus
}

func (s *SettingsView) Init() tea.Cmd {
	return nil
}

func (s *SettingsView) SetData(_ messages.DataLoadedMsg) {}

func (s *SettingsView) Update(msg tea.Msg) (component.Model, tea.Cmd) {
	if s.mode == SettingsModeEdit {
		return s.updateEdit(msg)
	}
	return s.updateDisplay(msg)
}

func (s *SettingsView) updateDisplay(msg tea.Msg) (component.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return s, nil
	}

	switch keyMsg.Type { //nolint:exhaustive // display mode only handles enter, quit, and back actions
	case tea.KeyEsc:
		return s, func() tea.Msg { return messages.NavigateBackMsg{} }
	case tea.KeyEnter:
		s.startEdit()
		return s, nil
	case tea.KeyCtrlC:
		return s, tea.Quit
	case tea.KeyRunes:
		if keyMsg.String() == "e" {
			s.startEdit()
			return s, nil
		}
	}

	return s, nil
}

func (s *SettingsView) updateEdit(msg tea.Msg) (component.Model, tea.Cmd) {
	var cmd tea.Cmd
	if s.focus == settingsFocusIDECommand {
		s.input, cmd = s.input.Update(msg)
	}

	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return s, cmd
	}

	switch keyMsg.Type { //nolint:exhaustive // edit mode only handles the settings form interaction keys
	case tea.KeyEsc:
		s.mode = SettingsModeDisplay
		s.input.Blur()
		s.input.SetValue(s.settings.IDECommand)
		return s, nil
	case tea.KeyCtrlC:
		return s, tea.Quit
	case tea.KeyTab, tea.KeyDown:
		s.focus = (s.focus + 1) % settingsFocusCount
		s.syncFocus()
		return s, nil
	case tea.KeyShiftTab, tea.KeyUp:
		s.focus = (s.focus - 1 + settingsFocusCount) % settingsFocusCount
		s.syncFocus()
		return s, nil
	case tea.KeySpace:
		if s.focus == settingsFocusPersonas {
			s.settings.ShowPersonasInCardEditor = !s.settings.ShowPersonasInCardEditor
			return s, nil
		}
		if s.focus == settingsFocusMCPs {
			s.settings.ShowMCPsInCardEditor = !s.settings.ShowMCPsInCardEditor
			return s, nil
		}
	case tea.KeyEnter:
		s.settings.IDECommand = strings.TrimSpace(s.input.Value())
		s.input.Blur()
		s.mode = SettingsModeDisplay
		if s.callbacks.OnSave != nil {
			saveCmd, saveErr := s.callbacks.OnSave(s.settings)
			if saveErr != nil {
				return s, func() tea.Msg { return messages.ErrorMsg{Err: saveErr} }
			}
			return s, saveCmd
		}
		return s, nil
	}

	return s, cmd
}

func (s *SettingsView) startEdit() {
	s.mode = SettingsModeEdit
	s.focus = settingsFocusIDECommand
	s.input.SetValue(s.settings.IDECommand)
	s.syncFocus()
}

func (s *SettingsView) syncFocus() {
	if s.focus == settingsFocusIDECommand {
		s.input.Focus()
		return
	}
	s.input.Blur()
}

func (s *SettingsView) View() string {
	if s.mode == SettingsModeEdit {
		return s.renderEdit()
	}
	return s.renderDisplay()
}

func (s *SettingsView) renderDisplay() string {
	title := titleStyle.Padding(0, 1).Render("Settings")
	currentCmd := s.settings.IDECommand
	if currentCmd == "" {
		currentCmd = "(not configured)"
	}

	maxWidth := max(settingsBoxMinWidth, s.input.Width+settingsBoxPadding*2)
	currentBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(gray).
		Padding(1, settingsBoxPadding).
		Width(maxWidth).
		Render(strings.Join([]string{
			"IDE Command: " + currentCmd,
			fmt.Sprintf("Show Personas In Card Editor: %s", enabledLabel(s.settings.ShowPersonasInCardEditor)),
			fmt.Sprintf("Show MCPs In Card Editor: %s", enabledLabel(s.settings.ShowMCPsInCardEditor)),
		}, "\n"))

	helpText := lipgloss.NewStyle().
		Foreground(gray).
		Padding(1, 0).
		Render("Press e or Enter to edit • Esc to go back")

	pathHelp := lipgloss.NewStyle().
		Foreground(secondaryColor).
		Render("Use $PATH in your IDE command to reference the session workspace path")

	return lipgloss.JoinVertical(lipgloss.Left, title, "", currentBox, "", pathHelp, "", helpText)
}

func (s *SettingsView) renderEdit() string {
	title := titleStyle.Padding(0, 1).Render("Edit Settings")
	explanationWidth := max(settingsExplanationWide, s.input.Width)

	explanation := lipgloss.NewStyle().
		Foreground(secondaryColor).
		Width(explanationWidth).
		Render("Configure the IDE command and the optional card editor sections.\nUse $PATH as the session workspace path placeholder.")

	inputBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(s.focusBorderColor(settingsFocusIDECommand)).
		Padding(1, settingsBoxPadding).
		Render(s.input.View())

	personasLine := s.renderToggleLine(settingsFocusPersonas, "Show Personas In Card Editor", s.settings.ShowPersonasInCardEditor)
	mcpsLine := s.renderToggleLine(settingsFocusMCPs, "Show MCPs In Card Editor", s.settings.ShowMCPsInCardEditor)

	footer := lipgloss.NewStyle().
		Foreground(gray).
		Render("Tab=next Shift+Tab=prev Space=toggle Enter=save Esc=cancel")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		explanation,
		"",
		lipgloss.NewStyle().Bold(true).Render("IDE Command"),
		inputBox,
		"",
		personasLine,
		mcpsLine,
		"",
		footer,
	)
}

func (s *SettingsView) renderToggleLine(index int, label string, enabled bool) string {
	prefix := "  "
	style := lipgloss.NewStyle()
	if s.focus == index {
		prefix = "> "
		style = style.Foreground(highlightColor).Bold(true)
	}
	return style.Render(fmt.Sprintf("%s[%s] %s", prefix, enabledLabel(enabled), label))
}

func (s *SettingsView) focusBorderColor(index int) lipgloss.Color {
	if s.focus == index {
		return focusedBorderColor
	}
	return gray
}

func enabledLabel(enabled bool) string {
	if enabled {
		return "on"
	}
	return "off"
}

func (s *SettingsView) SetSize(width, height int) {
	s.Base.SetSize(width, height)
	s.input.Width = min(width*9/10, 80) //nolint:mnd // responsive width keeps the field usable without overflowing narrow terminals
}
