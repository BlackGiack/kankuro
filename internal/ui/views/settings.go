package views

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
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
	OnSave              func(settings contracts.Settings) tea.Cmd
	SavePending         func() bool
	ReadPending         func() bool
	OnOpenPuppetSources func() tea.Cmd
	OnBack              func() tea.Cmd
}

type SettingsView struct {
	component.Base

	mode                                                  SettingsMode
	focus                                                 int
	input                                                 textinput.Model
	terminalInput                                         textinput.Model
	settings                                              contracts.Settings
	baseline                                              contracts.Settings
	ideEdited, terminalEdited, personasEdited, mcpsEdited bool
	callbacks                                             SettingsCallbacks
	draftID                                               uint64
	saving                                                bool
	readBlocked                                           bool
	saveBlocked                                           bool
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
		settings:      contracts.Settings{},
		callbacks:     callbacks,
	}
}

func (s *SettingsView) SetSettings(cfg contracts.Settings) {
	s.settings = cfg
	s.baseline = cfg
	s.input.SetValue(cfg.IDECommand)
	s.terminalInput.SetValue(cfg.TerminalCommand)
}

// ApplyLoadedSettings updates the baseline while preserving fields edited in the current draft.
func (s *SettingsView) ApplyLoadedSettings(cfg contracts.Settings) {
	if s.mode != SettingsModeEdit {
		s.SetSettings(cfg)
		return
	}
	if !s.ideEdited && s.input.Value() == s.baseline.IDECommand {
		s.input.SetValue(cfg.IDECommand)
	}
	if !s.terminalEdited && s.terminalInput.Value() == s.baseline.TerminalCommand {
		s.terminalInput.SetValue(cfg.TerminalCommand)
	}
	if !s.personasEdited {
		s.settings.ShowPersonasInCardEditor = cfg.ShowPersonasInCardEditor
	}
	if !s.mcpsEdited {
		s.settings.ShowMCPsInCardEditor = cfg.ShowMCPsInCardEditor
	}
	s.baseline = cfg
}

func (s *SettingsView) Settings() contracts.Settings {
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
	s.saving = false
	s.saveBlocked = false
	s.readBlocked = false
	return nil
}

func (s *SettingsView) DraftID() uint64 { return s.draftID }
func (s *SettingsView) SaveFinished() {
	s.saving = false
	s.saveBlocked = false
}

func (s *SettingsView) ReadFinished() { s.readBlocked = false }

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
		if s.callbacks.OnBack != nil {
			return s, s.callbacks.OnBack()
		}
		return s, nil
	case tea.KeyEnter:
		if s.saving {
			return s, nil
		}
		s.startEdit()
		return s, nil
	default:
		if len(keyMsg.Text) > 0 {
			switch keyMsg.String() {
			case "e":
				if s.saving {
					return s, nil
				}
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
	oldIDE, oldTerminal := s.input.Value(), s.terminalInput.Value()
	switch s.focus {
	case settingsFocusIDECommand:
		s.input, cmd = s.input.Update(msg)
	case settingsFocusTerminalCommand:
		s.terminalInput, cmd = s.terminalInput.Update(msg)
	}

	s.ideEdited = s.ideEdited || s.input.Value() != oldIDE
	s.terminalEdited = s.terminalEdited || s.terminalInput.Value() != oldTerminal
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
		s.readBlocked = false
		s.settings = s.baseline
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
		s.personasEdited = true
		s.settings.ShowPersonasInCardEditor = !s.settings.ShowPersonasInCardEditor
		return s, nil
	case settingsFocusMCPs:
		s.mcpsEdited = true
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
	if s.callbacks.ReadPending != nil && s.callbacks.ReadPending() {
		s.readBlocked = true
		return s, nil
	}
	if s.callbacks.SavePending != nil && s.callbacks.SavePending() {
		s.saveBlocked = true
		return s, nil
	}
	s.saveBlocked = false
	s.readBlocked = false
	s.settings.IDECommand = strings.TrimSpace(s.input.Value())
	s.settings.TerminalCommand = strings.TrimSpace(s.terminalInput.Value())
	s.input.Blur()
	s.terminalInput.Blur()
	s.mode = SettingsModeDisplay
	if s.callbacks.OnSave != nil {
		cmd := s.callbacks.OnSave(s.settings)
		if cmd == nil {
			s.baseline = s.settings
		}
		s.saving = cmd != nil
		return s, cmd
	}
	return s, nil
}

func (s *SettingsView) startEdit() {
	s.draftID++
	s.readBlocked = false
	s.ideEdited, s.terminalEdited, s.personasEdited, s.mcpsEdited = false, false, false, false
	s.saveBlocked = false
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
	var content string
	if s.mode == SettingsModeEdit {
		content = s.renderEdit()
	} else {
		content = s.renderDisplay()
	}
	if s.saving {
		return content + "\nSaving settings…"
	}
	if s.readBlocked {
		return content + "\nLoading settings…"
	}
	if s.saveBlocked {
		return content + "\nA settings save is still in progress"
	}
	return content
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
