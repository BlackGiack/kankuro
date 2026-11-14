package views

import (
	"errors"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/ui/component"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/styles"
)

// SessionLauncherCallbacks defines callbacks for session launcher actions.
type SessionLauncherCallbacks struct {
	OnSetNameAsync func(sessionID, name, processID string) tea.Cmd
}

// SessionLauncher is a simple form for launching a new session.
type SessionLauncher struct {
	component.Base

	card  contracts.CardSummary
	input textinput.Model

	callbacks SessionLauncherCallbacks

	launchDone bool
	launchErr  error
	processID  string
	effect     contracts.LifecycleEffect
	sessionID  string
	err        error
	naming     bool
}

// NewSessionLauncher creates a new session launcher.
func NewSessionLauncher(callbacks SessionLauncherCallbacks) *SessionLauncher {
	input := textinput.New()
	input.Placeholder = "Session name (Enter to confirm)"
	input.SetWidth(defaultInputWidth)
	input.Focus()

	return &SessionLauncher{
		input:     input,
		callbacks: callbacks,
	}
}

// SetCard sets the ID card to launch.
func (l *SessionLauncher) SetCard(card contracts.CardSummary) {
	l.card = card
}

// SetName sets the session name.
func (l *SessionLauncher) SetName(name string) {
	l.input.SetValue(name)
}

// Reset clears the launcher.
func (l *SessionLauncher) Reset() {
	l.card = contracts.CardSummary{}
	l.input.SetValue("")
	l.input.Focus()
	l.launchDone = false
	l.launchErr = nil
	l.sessionID = ""
	l.processID = ""
	l.effect = ""
	l.err = nil
	l.naming = false
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
		if l.launchFatal() {
			// No process ever started; the only sensible action is to leave.
			if keyMsg.Code == tea.KeyEsc {
				return func() tea.Msg { return messages.NavigateBackMsg{} }, true
			}
			return nil, false
		}
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

func (l *SessionLauncher) launchFatal() bool { return l.effect != contracts.UsableActivation }

// View renders the launcher.
//
//nolint:nestif // The launcher renders each activation outcome and its naming state.
func (l *SessionLauncher) View() string {
	var content strings.Builder

	if l.err != nil {
		content.WriteString(styles.ErrorStyle.Render(l.err.Error()) + "\n\n")
	}

	if l.launchDone {
		if l.launchFatal() {
			label := "Launch failed: "
			if l.effect == contracts.UncertainEffect {
				label = "Launch outcome uncertain: "
			}
			content.WriteString(styles.ErrorStyle.Render(label+l.launchErr.Error()) + "\n\n")
			content.WriteString(styles.HelpStyle.Render("Press Esc to return"))
			return content.String()
		}
		if l.launchErr != nil {
			// The process is running; only its state failed to save. Keep
			// the normal name/attach flow below, just flag the warning.
			content.WriteString(styles.ErrorStyle.Render("Warning: "+l.launchErr.Error()) + "\n\n")
		} else {
			content.WriteString(styles.SuccessStyle.Render("Session launched successfully!") + "\n\n")
		}
		content.WriteString(styles.LabelStyle.Render("Session ID: ") + l.sessionID + "\n\n")
		content.WriteString(styles.LabelStyle.Render("Session Name") + "\n")
		content.WriteString(l.input.View() + "\n\n")
		content.WriteString(styles.HelpStyle.Render("Enter: Confirm name and attach | Esc: Cancel"))
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
func (l *SessionLauncher) SetLaunchResult(result contracts.LifecycleResult) {
	l.launchDone = true
	if err := result.Validate(); err != nil {
		result = contracts.LifecycleResult{
			Effect:         contracts.UncertainEffect,
			Diagnostic:     err.Error(),
			DiagnosticCode: contracts.DiagnosticInvalid,
		}
	}
	l.effect, l.sessionID, l.processID = result.Effect, result.SessionID, result.ProcessActivationID
	l.launchErr = nil
	if result.Diagnostic != "" {
		l.launchErr = errors.New(result.Diagnostic)
	}
}

func (l *SessionLauncher) ProcessActivationID() string { return l.processID }

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
func (l *SessionLauncher) Card() contracts.CardSummary {
	return l.card
}

// Input returns the underlying text input model.
func (l *SessionLauncher) Input() *textinput.Model {
	return &l.input
}

func (l *SessionLauncher) handleNameSubmit() tea.Cmd {
	if l.naming {
		return nil
	}
	name := strings.TrimSpace(l.input.Value())
	if name == "" {
		name = l.sessionID
	}

	if l.callbacks.OnSetNameAsync == nil {
		return nil
	}
	l.naming = true
	return l.callbacks.OnSetNameAsync(l.sessionID, name, l.processID)
}

// SetNameResult releases the pending form and displays a rename failure.
func (l *SessionLauncher) SetNameResult(err error) { l.naming = false; l.err = err }
