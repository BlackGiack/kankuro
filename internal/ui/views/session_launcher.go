package views

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"kankuro/internal/config"
	"kankuro/internal/ui/commands"
	"kankuro/internal/ui/component"
	"kankuro/internal/ui/messages"
	"kankuro/internal/ui/styles"
)

// SessionLauncherCallbacks defines callbacks for session launcher actions.
type SessionLauncherCallbacks struct {
	OnSetName func(sessionID, name string) error
	OnAttach  func(sessionID string) tea.Cmd
}

// SessionLauncher is a simple form for launching a new session.
type SessionLauncher struct {
	component.Base

	card  config.IDCard
	input textinput.Model

	callbacks SessionLauncherCallbacks
	cmds      *commands.Commands

	launchDone bool
	launchErr  error
	sessionID  string
	err        error
}

// NewSessionLauncher creates a new session launcher.
func NewSessionLauncher(callbacks SessionLauncherCallbacks, cmds *commands.Commands) *SessionLauncher {
	input := textinput.New()
	input.Placeholder = "Session name (Enter to confirm)"
	input.SetWidth(defaultInputWidth)
	input.Focus()

	return &SessionLauncher{
		input:     input,
		callbacks: callbacks,
		cmds:      cmds,
	}
}

// SetCard sets the ID card to launch.
func (l *SessionLauncher) SetCard(card config.IDCard) {
	l.card = card
}

// SetName sets the session name.
func (l *SessionLauncher) SetName(name string) {
	l.input.SetValue(name)
}

// Reset clears the launcher.
func (l *SessionLauncher) Reset() {
	l.card = config.IDCard{}
	l.input.SetValue("")
	l.input.Focus()
	l.launchDone = false
	l.launchErr = nil
	l.sessionID = ""
	l.err = nil
}

// Init initializes the component.
func (l *SessionLauncher) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (l *SessionLauncher) Update(msg tea.Msg) (component.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		if cmd, handled := l.handleKey(keyMsg); handled {
			return l, cmd
		}
	}

	var cmd tea.Cmd
	l.input, cmd = l.input.Update(msg)
	return l, cmd
}

func (l *SessionLauncher) handleKey(keyMsg tea.KeyPressMsg) (tea.Cmd, bool) {
	if keyMsg.String() == keyCtrlC {
		return tea.Quit, true
	}

	if l.launchDone {
		switch keyMsg.Code {
		case tea.KeyEnter:
			return l.handleNameSubmit(), true
		case tea.KeyEsc:
			return func() tea.Msg { return messages.NavigateBackMsg{} }, true
		}
		return nil, false
	}

	if keyMsg.Code == tea.KeyEsc {
		return func() tea.Msg { return messages.NavigateBackMsg{} }, true
	}

	return nil, false
}

// View renders the launcher.
func (l *SessionLauncher) View() string {
	var content strings.Builder

	if l.err != nil {
		content.WriteString(styles.ErrorStyle.Render(l.err.Error()) + "\n\n")
	}

	if l.launchDone {
		if l.launchErr != nil {
			content.WriteString(styles.ErrorStyle.Render("Launch failed: "+l.launchErr.Error()) + "\n\n")
			content.WriteString(styles.HelpStyle.Render("Press Esc to return"))
		} else {
			content.WriteString(styles.SuccessStyle.Render("Session launched successfully!") + "\n\n")
			content.WriteString(styles.LabelStyle.Render("Session ID: ") + l.sessionID + "\n\n")
			content.WriteString(styles.LabelStyle.Render("Session Name") + "\n")
			content.WriteString(l.input.View() + "\n\n")
			content.WriteString(styles.HelpStyle.Render("Enter: Confirm name and attach | Esc: Cancel"))
		}
		return content.String()
	}

	content.WriteString(styles.TitleStyle.Render("Launch Session") + "\n\n")

	if l.card.ID != "" {
		content.WriteString(styles.LabelStyle.Render("Card: ") + l.card.Name + "\n\n")
	}

	content.WriteString(styles.LabelStyle.Render("Session Name") + "\n")
	content.WriteString(l.input.View() + "\n\n")

	if !l.launchDone {
		content.WriteString(styles.LabelStyle.Render("Launching session...") + "\n\n")
		content.WriteString(styles.HelpStyle.Render("Enter: Confirm name and attach | Esc: Cancel"))
	} else {
		content.WriteString(styles.LabelStyle.Render("Session ID: ") + l.sessionID + "\n\n")
		content.WriteString(styles.HelpStyle.Render("Enter: Confirm name and attach | Esc: Cancel"))
	}

	return content.String()
}

// SetLaunchResult sets the result of an asynchronous launch operation.
func (l *SessionLauncher) SetLaunchResult(sessionID string, err error) {
	l.launchDone = true
	l.launchErr = err
	l.sessionID = sessionID
	if err != nil {
		l.err = err
	}
}

// LaunchDone returns true if the launch has completed.
func (l *SessionLauncher) LaunchDone() bool {
	return l.launchDone
}

// LaunchErr returns the launch error, if any.
func (l *SessionLauncher) LaunchErr() error {
	return l.launchErr
}

// SessionID returns the session ID.
func (l *SessionLauncher) SessionID() string {
	return l.sessionID
}

// Card returns the current ID card.
func (l *SessionLauncher) Card() config.IDCard {
	return l.card
}

// Input returns the underlying text input model.
func (l *SessionLauncher) Input() *textinput.Model {
	return &l.input
}

func (l *SessionLauncher) handleNameSubmit() tea.Cmd {
	name := strings.TrimSpace(l.input.Value())
	if name == "" {
		name = l.sessionID
	}

	var cmds []tea.Cmd

	if l.callbacks.OnSetName != nil {
		if err := l.callbacks.OnSetName(l.sessionID, name); err != nil {
			l.err = err
			return nil
		}
	}

	if l.callbacks.OnAttach != nil {
		cmds = append(cmds, l.callbacks.OnAttach(l.sessionID))
	}

	cmds = append(cmds, func() tea.Msg { return messages.NavigateBackMsg{} })

	return tea.Batch(cmds...)
}
