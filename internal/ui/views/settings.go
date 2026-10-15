package views

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/ui/component"
	"github.com/BlackGiack/kankuro/internal/ui/components"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/styles"
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
	settingsFocusTerminalCommand
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

	mode          SettingsMode
	focus         int
	input         textinput.Model
	terminalInput textinput.Model
	settings      config.AppSettings
	callbacks     SettingsCallbacks
}

func NewSettingsView(callbacks SettingsCallbacks) *SettingsView {
	input := textinput.New()
	input.Placeholder = "e.g. code $PATH"
	input.SetWidth(defaultInputWidth)
	terminalInput := textinput.New()
	terminalInput.Placeholder = "e.g. xterm -e sh -lc 'cd \"$PATH\" && exec $SHELL'"
	terminalInput.SetWidth(defaultInputWidth)

	return &SettingsView{
		mode:          SettingsModeDisplay,
		input:         input,
		terminalInput: terminalInput,
		settings:      config.DefaultAppSettings(),
		callbacks:     callbacks,
	}
}

func (s *SettingsView) SetSettings(cfg config.AppSettings) {
	s.settings = cfg
	s.input.SetValue(cfg.IDECommand)
	s.terminalInput.SetValue(cfg.TerminalCommand)
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

func (s *SettingsView) TerminalInput() *textinput.Model { return &s.terminalInput }

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
	switch s.focus {
	case settingsFocusIDECommand:
		s.input, cmd = s.input.Update(msg)
	case settingsFocusTerminalCommand:
		s.terminalInput, cmd = s.terminalInput.Update(msg)
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
		s.terminalInput.Blur()
		s.input.SetValue(s.settings.IDECommand)
		s.terminalInput.SetValue(s.settings.TerminalCommand)
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
	s.settings.TerminalCommand = strings.TrimSpace(s.terminalInput.Value())
	s.input.Blur()
	s.terminalInput.Blur()
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
	s.terminalInput.SetValue(s.settings.TerminalCommand)
	s.syncFocus()
}

func (s *SettingsView) syncFocus() {
	if s.focus == settingsFocusIDECommand {
		s.input.Focus()
		s.terminalInput.Blur()
		return
	}
	if s.focus == settingsFocusTerminalCommand {
		s.input.Blur()
		s.terminalInput.Focus()
		return
	}
	s.input.Blur()
	s.terminalInput.Blur()
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
	terminalCmd := s.settings.TerminalCommand
	if terminalCmd == "" {
		terminalCmd = "(not configured)"
	}

	maxWidth := max(settingsBoxMinWidth, s.input.Width()+settingsBoxPadding*2)
	commandWidth := max(0, maxWidth-4-lipgloss.Width("Terminal Command: "))
	ideCommandWidth := max(0, maxWidth-4-lipgloss.Width("IDE Command: "))
	currentCmd = components.TruncateHead(components.DisplayText(currentCmd), ideCommandWidth)
	terminalCmd = components.TruncateHead(components.DisplayText(terminalCmd), commandWidth)
	currentBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Gray).
		Padding(1, settingsBoxPadding).
		Width(maxWidth).
		Render(strings.Join([]string{
			"IDE Command: " + currentCmd,
			"Terminal Command: " + terminalCmd,
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
		Render("Use $PATH in either command for the session workspace path")

	return lipgloss.JoinVertical(lipgloss.Left, title, "", currentBox, "", puppetAction, "", pathHelp, "", helpText)
}

func (s *SettingsView) renderEdit() string {
	title := styles.TitleStyle.Padding(0, 1).Render("Edit Settings")
	explanationWidth := max(settingsExplanationWide, s.input.Width())

	explanation := lipgloss.NewStyle().
		Foreground(styles.SecondaryColor).
		Width(explanationWidth).
		Render("Configure launch commands and optional card editor sections. Use $PATH as the session workspace path.")

	inputBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(s.focusBorderColor(settingsFocusIDECommand)).
		Padding(0, settingsBoxPadding).
		Render(s.input.View())
	terminalBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(s.focusBorderColor(settingsFocusTerminalCommand)).
		Padding(0, settingsBoxPadding).
		Render(s.terminalInput.View())

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
		lipgloss.NewStyle().Bold(true).Render("Terminal Command (opens a separate window)"),
		terminalBox,
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
	inputWidth := max(1, min(width*9/10, 80)-2*settingsBoxPadding-2) //nolint:mnd // include border and horizontal padding
	s.input.SetWidth(inputWidth)
	s.terminalInput.SetWidth(inputWidth)
}
