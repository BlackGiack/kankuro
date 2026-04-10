package ui

import (
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"kankuro/internal/config"
	"kankuro/internal/session"
)

// New builds the TUI model with persisted config and current session snapshots.
func New(repo config.Repository, sm *session.Manager) (*App, error) {
	cards, err := repo.ListIDCards()
	if err != nil {
		return nil, err
	}
	skills, err := repo.ListSkills()
	if err != nil {
		return nil, err
	}
	mcps, err := repo.ListMCPs()
	if err != nil {
		return nil, err
	}
	personas, err := repo.ListPersonas()
	if err != nil {
		return nil, err
	}

	app := &App{
		repo:           repo,
		session:        sm,
		mode:           viewDashboard,
		cards:          cards,
		skills:         skills,
		personas:       personas,
		mcps:           mcps,
		sessions:       sm.List(),
		cardForm:       NewCardFormState(),
		sessionView:    SessionViewState{FocusedPane: paneCards},
		naming:         NewNamingState(),
		skillSources:   NewSourceFormState(),
		personaSources: NewSourceFormState(),
		workspaceForm:  NewWorkspaceFormState(),
		help:           help.New(),
	}

	app.initLists()
	app.initInputs()
	app.refreshWorkspaces()
	app.refreshLists()

	return app, nil
}

func (a *App) initLists() {
	sessionsDelegate := list.NewDefaultDelegate()
	sessionsDelegate.SetHeight(sessionsDelegateHeight)
	sessionsDelegate.Styles.NormalTitle = sessionsDelegate.Styles.NormalTitle.Foreground(primaryColor)
	sessionsDelegate.Styles.NormalDesc = sessionsDelegate.Styles.NormalDesc.Foreground(gray)
	sessionsDelegate.Styles.SelectedTitle = sessionsDelegate.Styles.SelectedTitle.
		Foreground(highlightColor).
		BorderForeground(primaryColor).
		UnsetBackground()
	sessionsDelegate.Styles.SelectedDesc = sessionsDelegate.Styles.SelectedDesc.
		Foreground(highlightColor).
		BorderForeground(primaryColor).
		UnsetBackground()
	sessionsDelegate.Styles.DimmedTitle = sessionsDelegate.Styles.DimmedTitle.Foreground(gray)
	sessionsDelegate.Styles.DimmedDesc = sessionsDelegate.Styles.DimmedDesc.Foreground(gray)

	cardsDelegate := newCardDelegate(a.skills, a.personas)

	a.cardsList = list.New(toCardItems(a.cards, a.skills, a.personas), cardsDelegate, 0, 0)
	a.cardsList.Title = "ID Cards"
	a.cardsList.SetShowStatusBar(false)
	a.cardsList.SetFilteringEnabled(false)
	a.cardsList.Styles.Title = titleStyle

	a.sessionsList = list.New([]list.Item{}, sessionsDelegate, 0, 0)
	a.sessionsList.Title = "Active Sessions"
	a.sessionsList.SetShowStatusBar(false)
	a.sessionsList.SetFilteringEnabled(false)
	a.sessionsList.Styles.Title = titleStyle

	a.switchList = list.New([]list.Item{}, cardsDelegate, 0, 0)
	a.switchList.Title = "Switch Agent"
	a.switchList.SetShowStatusBar(false)
	a.switchList.SetFilteringEnabled(false)
	a.switchList.Styles.Title = titleStyle
}

func (a *App) initInputs() {
	// Card form inputs
	placeholders := []string{
		"name",
		"driver command (e.g. claude)",
		"model",
		"face_style (neutral/happy/focused/thinking)",
		"use tmp workspace (y/n)",
	}

	for i := range a.cardForm.Inputs {
		a.cardForm.Inputs[i] = textinput.New()
		a.cardForm.Inputs[i].Placeholder = placeholders[i]
		a.cardForm.Inputs[i].Width = inputWidthDefault
	}
	a.cardForm.Inputs[0].Focus()

	// Initialize puppet selector index
	for i, p := range puppetStyles() {
		if p == a.cardForm.Puppet {
			a.cardForm.PuppetSelectorIndex = i
			break
		}
	}

	// MCP inputs
	a.mcpInputs = make([]textinput.Model, mcpInputCount)
	mcpPlaceholders := []string{
		"mcp name",
		"mcp command (local only)",
		"args (comma separated, local only)",
		"cwd (local only)",
		"url (remote only)",
		"env (KEY=VALUE, local only)",
		"headers (KEY=VALUE, remote only)",
	}
	for i := range a.mcpInputs {
		a.mcpInputs[i] = textinput.New()
		a.mcpInputs[i].Placeholder = mcpPlaceholders[i]
		a.mcpInputs[i].Width = inputWidthDefault
	}

	// AgentsMD editor
	a.cardForm.Editor = textarea.New()
	a.cardForm.Editor.Placeholder = "Enter AGENTS.md content here..."
	a.cardForm.Editor.SetWidth(editorWidth)
	a.cardForm.Editor.SetHeight(editorHeight)

	a.cardForm.PathInput = textinput.New()
	a.cardForm.PathInput.Placeholder = "Path to AGENTS.md file..."
	a.cardForm.PathInput.Width = inputWidthDefault

	// skillSourceInput, personaSourceInput, namingInput, workspaceInputs
	// are already initialized in their respective New* constructors.
	// But we might want to override placeholders if needed, though they are set in constructors.
}

// Init starts periodic refresh and textinput blink commands.
func (a *App) Init() tea.Cmd {
	return tea.Batch(tick(), textinput.Blink)
}

func tick() tea.Cmd {
	return tea.Tick(1*time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}
