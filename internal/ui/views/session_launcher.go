package views

import (
	"errors"
	"slices"
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

// SessionLauncherCallbacks defines callbacks for session launcher actions.
type SessionLauncherCallbacks struct {
	OnCancelWorkspace func() tea.Cmd
	OnLaunch          func([]string) tea.Cmd
	OnOpenWorkspaces  func() tea.Cmd
	OnSetNameAsync    func(sessionID, name, processID string) tea.Cmd
}

// SessionLauncher is a simple form for launching a new session.
type SessionLauncher struct {
	component.Base

	card  contracts.CardSummary
	input textinput.Model

	callbacks SessionLauncherCallbacks

	launchDone          bool
	launchErr           error
	processID           string
	effect              contracts.LifecycleEffect
	sessionID           string
	err                 error
	naming              bool
	workspaceForm       bool
	workspaceCancelling bool
	workspaceEditor     *workspaceFields
	workspaceFocus      int
}

// NewSessionLauncher creates a new session launcher.
func NewSessionLauncher(callbacks SessionLauncherCallbacks) *SessionLauncher {
	input := textinput.New()
	input.Placeholder = "Session name (Enter to confirm)"
	input.SetWidth(defaultInputWidth)
	input.Focus()

	return &SessionLauncher{
		input:           input,
		workspaceEditor: newWorkspaceFields(),
		callbacks:       callbacks,
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
	l.workspaceForm = false
	l.workspaceCancelling = false
	l.workspaceEditor.workspacePicker.Close()
	l.workspaceEditor.workspaceID = ""
	l.workspaceEditor.subWorkspaceIDs = nil
	l.workspaceEditor.subWorkspaceEdited = false
}

// SetWorkspaceDraft starts a detached launch selection using the editor's workspace controls.
func (l *SessionLauncher) SetWorkspaceDraft(p contracts.LaunchPreparation) {
	l.workspaceEditor.workspaceID = p.WorkspaceID
	l.workspaceEditor.subWorkspaceIDs = slices.Clone(p.SubWorkspaceIDs)
	l.workspaceFocus = FocusWorkspace
	l.workspaceEditor.setChoices(p.Workspaces)
	l.workspaceForm = true
	l.workspaceCancelling = false
}

// WorkspaceDraftActive identifies a draft that can still accept actions.
func (l *SessionLauncher) WorkspaceDraftActive() bool {
	return l.workspaceForm && !l.workspaceCancelling
}

// WorkspaceDraftCancelling identifies a draft awaiting its scoped cancellation.
func (l *SessionLauncher) WorkspaceDraftCancelling() bool {
	return l.workspaceForm && l.workspaceCancelling
}

func (l *SessionLauncher) SetWorkspaceChoices(choices contracts.CardEditorChoices) {
	l.workspaceEditor.setChoices(choices.Workspaces)
}

func (l *SessionLauncher) SetSize(width, height int) {
	l.Base.SetSize(width, height)
	l.workspaceEditor.workspacePicker.SetSize(width, height)
}

func (l *SessionLauncher) updateWorkspaceForm(msg tea.Msg) tea.Cmd {
	e := l.workspaceEditor
	if e.workspacePicker.IsOpen() {
		return e.updatePicker(msg, l.workspaceFocus == FocusSubWorkspaces)
	}
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}
	switch {
	case key.String() == keyCtrlC:
		return tea.Quit
	case key.Code == tea.KeyTab || key.String() == keyShiftTab:
		if l.workspaceFocus == FocusWorkspace {
			l.workspaceFocus = FocusSubWorkspaces
		} else {
			l.workspaceFocus = FocusWorkspace
		}
	case key.Code == tea.KeySpace:
		e.open(l.workspaceFocus == FocusSubWorkspaces)
	case key.String() == keyCtrlX:
		if l.workspaceFocus == FocusSubWorkspaces {
			e.subWorkspaceIDs = []string{}
			e.subWorkspaceEdited = true
		}
	case key.String() == keyCtrlE:
		if l.callbacks.OnOpenWorkspaces != nil {
			return l.callbacks.OnOpenWorkspaces()
		}
	case key.Code == tea.KeyEsc:
		l.workspaceCancelling = true
		if l.callbacks.OnCancelWorkspace != nil {
			return l.callbacks.OnCancelWorkspace()
		}
		return func() tea.Msg { return messages.NavigateBackMsg{} }
	case key.Code == tea.KeyEnter:
		ids := append([]string{e.workspaceID}, e.subWorkspaceIDs...)
		l.workspaceForm = false
		if l.callbacks.OnLaunch != nil {
			return l.callbacks.OnLaunch(ids)
		}
	}
	return nil
}

// Init initializes the component.
func (l *SessionLauncher) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (l *SessionLauncher) Update(msg tea.Msg) (component.Model, tea.Cmd) {
	if l.workspaceForm {
		if l.workspaceCancelling {
			return l, nil
		}
		return l, l.updateWorkspaceForm(msg)
	}
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
	if l.workspaceForm {
		e := l.workspaceEditor
		if e.workspacePicker.IsOpen() {
			return lipgloss.Place(l.Width, l.Height, lipgloss.Center, lipgloss.Center, e.workspacePicker.View())
		}
		name := components.DisplayText(l.card.Name)
		if l.Width > 0 {
			name = components.TruncateHead(name, max(0, l.Width-len("Card: ")))
		}
		return styles.TitleStyle.Render(
			"Launch Session",
		) + "\n\n" + styles.LabelStyle.Render(
			"Card: ",
		) + name + "\n\n" + e.renderWorkspaceLine(
			l.Width,
			l.workspaceFocus == FocusWorkspace,
		) + "\n" + e.renderSubWorkspacesLine(
			l.Width,
			l.workspaceFocus == FocusSubWorkspaces,
		) + "\n\n" + styles.HelpStyle.Render(
			"Tab: switch | Space: select | Enter: launch | Esc: cancel",
		)
	}

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
