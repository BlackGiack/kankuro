package views

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"kankuro/internal/config"
	"kankuro/internal/ui/commands"
	"kankuro/internal/ui/component"
	"kankuro/internal/ui/messages"
	"kankuro/internal/ui/styles"
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
	settingsFocusPuppetSources
	settingsFocusCount
)

type SettingsCallbacks struct {
	OnSave              func(settings config.AppSettings) tea.Cmd
	OnOpenPuppetSources func() tea.Cmd
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
	input.SetWidth(defaultInputWidth)

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
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return s, nil
	}

	if keyMsg.String() == keyCtrlC {
		return s, tea.Quit
	}

	switch keyMsg.Code {
	case tea.KeyEsc:
		return s, func() tea.Msg { return messages.NavigateMsg{View: messages.ViewDashboard} }
	case tea.KeyEnter:
		s.startEdit()
		return s, nil
	default:
		if len(keyMsg.Text) > 0 {
			switch keyMsg.String() {
			case "e":
				s.startEdit()
				return s, nil
			case "p":
				if s.callbacks.OnOpenPuppetSources != nil {
					return s, s.callbacks.OnOpenPuppetSources()
				}
			}
		}
	}

	return s, nil
}

func (s *SettingsView) updateEdit(msg tea.Msg) (component.Model, tea.Cmd) {
	var cmd tea.Cmd
	if s.focus == settingsFocusIDECommand {
		s.input, cmd = s.input.Update(msg)
	}

	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return s, cmd
	}

	if keyMsg.String() == keyCtrlC {
		return s, tea.Quit
	}

	if keyMsg.String() == keyShiftTab {
		s.focus = (s.focus - 1 + settingsFocusCount) % settingsFocusCount
		s.syncFocus()
		return s, nil
	}

	switch keyMsg.Code {
	case tea.KeyEsc:
		s.mode = SettingsModeDisplay
		s.input.Blur()
		s.input.SetValue(s.settings.IDECommand)
		return s, nil
	case tea.KeyTab, tea.KeyDown:
		s.focus = (s.focus + 1) % settingsFocusCount
		s.syncFocus()
		return s, nil
	case tea.KeyUp:
		s.focus = (s.focus - 1 + settingsFocusCount) % settingsFocusCount
		s.syncFocus()
		return s, nil
	case tea.KeySpace:
		return s.handleEditSpace()
	case tea.KeyEnter:
		return s.handleEditEnter()
	}

	return s, cmd
}

func (s *SettingsView) handleEditSpace() (component.Model, tea.Cmd) {
	switch s.focus {
	case settingsFocusPersonas:
		s.settings.ShowPersonasInCardEditor = !s.settings.ShowPersonasInCardEditor
		return s, nil
	case settingsFocusMCPs:
		s.settings.ShowMCPsInCardEditor = !s.settings.ShowMCPsInCardEditor
		return s, nil
	case settingsFocusPuppetSources:
		if s.callbacks.OnOpenPuppetSources != nil {
			return s, s.callbacks.OnOpenPuppetSources()
		}
		return s, nil
	}
	return s, nil
}

func (s *SettingsView) handleEditEnter() (component.Model, tea.Cmd) {
	if s.focus == settingsFocusPuppetSources {
		if s.callbacks.OnOpenPuppetSources != nil {
			return s, s.callbacks.OnOpenPuppetSources()
		}
		return s, nil
	}
	s.settings.IDECommand = strings.TrimSpace(s.input.Value())
	s.input.Blur()
	s.mode = SettingsModeDisplay
	if s.callbacks.OnSave != nil {
		return s, s.callbacks.OnSave(s.settings)
	}
	return s, nil
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
	title := styles.TitleStyle.Padding(0, 1).Render("Settings")
	currentCmd := s.settings.IDECommand
	if currentCmd == "" {
		currentCmd = "(not configured)"
	}

	maxWidth := max(settingsBoxMinWidth, s.input.Width()+settingsBoxPadding*2)
	currentBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Gray).
		Padding(1, settingsBoxPadding).
		Width(maxWidth).
		Render(strings.Join([]string{
			"IDE Command: " + currentCmd,
			fmt.Sprintf("Show Personas In Card Editor: %s", enabledLabel(s.settings.ShowPersonasInCardEditor)),
			fmt.Sprintf("Show MCPs In Card Editor: %s", enabledLabel(s.settings.ShowMCPsInCardEditor)),
		}, "\n"))

	actionStyle := lipgloss.NewStyle().Foreground(styles.HighlightColor).Bold(true)
	puppetAction := actionStyle.Render("→ Manage Puppet Sources")

	helpText := lipgloss.NewStyle().
		Foreground(styles.Gray).
		Padding(1, 0).
		Render("e=edit • p=puppet sources • Esc=go back")

	pathHelp := lipgloss.NewStyle().
		Foreground(styles.SecondaryColor).
		Render("Use $PATH in your IDE command to reference the session workspace path")

	return lipgloss.JoinVertical(lipgloss.Left, title, "", currentBox, "", puppetAction, "", pathHelp, "", helpText)
}

func (s *SettingsView) renderEdit() string {
	title := styles.TitleStyle.Padding(0, 1).Render("Edit Settings")
	explanationWidth := max(settingsExplanationWide, s.input.Width())

	explanation := lipgloss.NewStyle().
		Foreground(styles.SecondaryColor).
		Width(explanationWidth).
		Render("Configure the IDE command and the optional card editor sections.\nUse $PATH as the session workspace path placeholder.")

	inputBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(s.focusBorderColor(settingsFocusIDECommand)).
		Padding(1, settingsBoxPadding).
		Render(s.input.View())

	personasLine := s.renderToggleLine(settingsFocusPersonas, "Show Personas In Card Editor", s.settings.ShowPersonasInCardEditor)
	mcpsLine := s.renderToggleLine(settingsFocusMCPs, "Show MCPs In Card Editor", s.settings.ShowMCPsInCardEditor)
	puppetSourcesLine := s.renderActionLine(settingsFocusPuppetSources, "Manage Puppet Sources")

	footer := lipgloss.NewStyle().
		Foreground(styles.Gray).
		Render("Tab=next Shift+Tab=prev Space=toggle Enter=select Esc=cancel")

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
		puppetSourcesLine,
		"",
		footer,
	)
}

func (s *SettingsView) renderToggleLine(index int, label string, enabled bool) string {
	prefix := "  "
	style := lipgloss.NewStyle()
	if s.focus == index {
		prefix = "> "
		style = style.Foreground(styles.HighlightColor).Bold(true)
	}
	return style.Render(fmt.Sprintf("%s[%s] %s", prefix, enabledLabel(enabled), label))
}

func (s *SettingsView) renderActionLine(index int, label string) string {
	prefix := "  "
	style := lipgloss.NewStyle()
	if s.focus == index {
		prefix = "> "
		style = style.Foreground(styles.HighlightColor).Bold(true)
	}
	return style.Render(fmt.Sprintf("%s→ %s", prefix, label))
}

func (s *SettingsView) focusBorderColor(index int) color.Color {
	if s.focus == index {
		return focusedBorderColor
	}
	return styles.Gray
}

func enabledLabel(enabled bool) string {
	if enabled {
		return "on"
	}
	return "off"
}

func (s *SettingsView) SetSize(width, height int) {
	s.Base.SetSize(width, height)
	s.input.SetWidth(min(width*9/10, 80)) //nolint:mnd // responsive width keeps the field usable without overflowing narrow terminals
}
