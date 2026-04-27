package views

import (
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
	// boxPadding is the padding for boxes.
	boxPadding = 2
	// boxOffset is the offset for boxes.
	boxOffset = 4
	// inputMaxWidth is the maximum width for text input.
	inputMaxWidth = 60
	// boxMinWidth is the minimum width for boxes.
	boxMinWidth = 30
	// explanationMinWidth is the minimum width for explanation text.
	explanationMinWidth = 40
	// sizeOffset is the offset for size calculations.
	sizeOffset = 10
)

// IDEConfigMode determines the current view mode.
type IDEConfigMode int

const (
	// IDEModeDisplay shows the current configuration.
	IDEModeDisplay IDEConfigMode = iota
	// IDEModeEdit allows editing the configuration.
	IDEModeEdit
)

// IDEConfigCallbacks defines callbacks for IDE config actions.
type IDEConfigCallbacks struct {
	// OnSave is called when the command is saved.
	OnSave func(command string) (tea.Cmd, error)

	// OnLaunch is called when user wants to launch IDE for a session.
	// Optional - may be handled by Dashboard instead.
	OnLaunch func(sessionPath string) error
}

// IDEConfig is a component for managing IDE launch command configuration.
type IDEConfig struct {
	component.Base

	mode   IDEConfigMode
	input  textinput.Model
	config config.IDEConfig

	callbacks IDEConfigCallbacks
	cmds      *commands.Commands
}

// NewIDEConfig creates a new IDE config component.
func NewIDEConfig(callbacks IDEConfigCallbacks, cmds *commands.Commands) *IDEConfig {
	input := textinput.New()
	input.Placeholder = "e.g. code $PATH"
	input.Width = inputMaxWidth

	return &IDEConfig{
		mode:      IDEModeDisplay,
		input:     input,
		callbacks: callbacks,
		cmds:      cmds,
	}
}

// SetConfig updates the current IDE configuration.
func (c *IDEConfig) SetConfig(cfg config.IDEConfig) {
	c.config = cfg
	c.input.SetValue(cfg.Command)
}

// Config returns the current configuration.
func (c *IDEConfig) Config() config.IDEConfig {
	return c.config
}

// Mode returns the current configuration mode.
func (c *IDEConfig) Mode() IDEConfigMode {
	return c.mode
}

// Input returns the underlying text input model.
func (c *IDEConfig) Input() *textinput.Model {
	return &c.input
}

// SetData implements messages.DataReceiver to load IDE config.
func (c *IDEConfig) SetData(_ messages.DataLoadedMsg) {
	// IDE config is not in DataLoadedMsg, so we need to load it separately
	// This is handled by the coordinator when opening the view
}

// Init initializes the component.
func (c *IDEConfig) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (c *IDEConfig) Update(msg tea.Msg) (component.Model, tea.Cmd) {
	if c.mode == IDEModeEdit {
		return c.updateEdit(msg)
	}
	return c.updateDisplay(msg)
}

// updateDisplay handles display mode input.
func (c *IDEConfig) updateDisplay(msg tea.Msg) (component.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return c, nil
	}

	switch keyMsg.String() {
	case keyEsc:
		return c, func() tea.Msg { return messages.NavigateBackMsg{} }

	case keyEnter, "e":
		c.mode = IDEModeEdit
		c.input.SetValue(c.config.Command)
		c.input.Focus()
		return c, nil

	case keyCtrlC:
		return c, tea.Quit
	}

	return c, nil
}

// updateEdit handles edit mode input.
func (c *IDEConfig) updateEdit(msg tea.Msg) (component.Model, tea.Cmd) {
	var cmd tea.Cmd
	c.input, cmd = c.input.Update(msg)

	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return c, cmd
	}

	switch keyMsg.String() {
	case keyEsc:
		c.mode = IDEModeDisplay
		c.input.Blur()
		return c, nil

	case keyEnter:
		c.config.Command = strings.TrimSpace(c.input.Value())
		c.input.Blur()
		c.mode = IDEModeDisplay

		if c.callbacks.OnSave != nil {
			saveCmd, saveErr := c.callbacks.OnSave(c.config.Command)
			if saveErr != nil {
				return c, func() tea.Msg {
					return messages.ErrorMsg{Err: saveErr}
				}
			}
			backCmd := func() tea.Msg { return messages.NavigateBackMsg{} }
			return c, tea.Batch(saveCmd, backCmd)
		}

	case keyCtrlC:
		return c, tea.Quit
	}

	return c, cmd
}

// View renders the component.
func (c *IDEConfig) View() string {
	if c.mode == IDEModeEdit {
		return c.renderEdit()
	}
	return c.renderDisplay()
}

// renderDisplay renders the display mode.
func (c *IDEConfig) renderDisplay() string {
	title := titleStyle.Padding(0, 1).Render("IDE Configuration")

	currentCmd := c.config.Command
	if currentCmd == "" {
		currentCmd = "(not configured)"
	}

	maxWidth := max(boxMinWidth, min(c.Width-boxOffset, inputMaxWidth+boxPadding*2))

	currentBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(gray).
		Padding(1, boxPadding).
		Width(maxWidth).
		Render("Command: " + currentCmd)

	helpText := lipgloss.NewStyle().
		Foreground(gray).
		Padding(1, 0).
		Render("Press e or Enter to edit • Esc to go back")

	pathHelp := lipgloss.NewStyle().
		Foreground(secondaryColor).
		Render("Use $PATH in your command to reference the session workspace path")

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		currentBox,
		"",
		pathHelp,
		"",
		helpText,
	)
}

// renderEdit renders the edit mode.
func (c *IDEConfig) renderEdit() string {
	title := titleStyle.Padding(0, 1).Render("Edit IDE Command")

	explanationWidth := max(explanationMinWidth, min(c.Width-boxOffset, inputMaxWidth))

	explanation := lipgloss.NewStyle().
		Foreground(secondaryColor).
		Width(explanationWidth).
		Render("Enter the command to launch your IDE.\n" +
			"Use $PATH as a placeholder for the session workspace path.\n" +
			"Example: code $PATH")

	inputBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(focusedBorderColor).
		Padding(1, boxPadding).
		Render(c.input.View())

	footer := lipgloss.NewStyle().
		Foreground(gray).
		Render("Enter to save • Esc to cancel")

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		explanation,
		"",
		inputBox,
		"",
		footer,
	)
}

// SetSize implements component.Model.
func (c *IDEConfig) SetSize(width, height int) {
	c.Base.SetSize(width, height)
	c.input.Width = min(inputMaxWidth, width-sizeOffset)
}
