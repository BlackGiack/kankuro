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
		repo:               repo,
		session:            sm,
		mode:               viewDashboard,
		cards:              cards,
		skills:             skills,
		personas:           personas,
		mcps:               mcps,
		sessions:           sm.List(),
		selectedSkillIDs:   make(map[string]bool),
		selectedPersonaIDs: make(map[string]bool),
		focusedPane:        paneCards,
		help:               help.New(),
	}

	app.initLists()
	app.initInputs()
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
	a.cardInputs = make([]textinput.Model, cardInputCount)
	placeholders := []string{
		"name",
		"driver command (e.g. claude)",
		"model",
		"working_dir",
		"face_style (neutral/happy/focused/thinking)",
		"use tmp workspace (y/n)",
	}
	for i := range a.cardInputs {
		a.cardInputs[i] = textinput.New()
		a.cardInputs[i].Placeholder = placeholders[i]
		a.cardInputs[i].Width = inputWidthDefault
	}
	a.cardInputs[0].Focus()
	// default puppet value
	a.cardPuppet = defaultPuppet
	for i, p := range puppetStyles() {
		if p == a.cardPuppet {
			a.puppetSelectorIndex = i
			break
		}
	}

	a.mcpInputs = make([]textinput.Model, mcpInputCount)
	a.mcpInputs[0] = textinput.New()
	a.mcpInputs[0].Placeholder = "mcp name"
	a.mcpInputs[0].Width = inputWidthDefault
	a.mcpInputs[1] = textinput.New()
	a.mcpInputs[1].Placeholder = "mcp command"
	a.mcpInputs[1].Width = inputWidthDefault

	a.agentsEditor = textarea.New()
	a.agentsEditor.Placeholder = "Enter AGENTS.md content here..."
	a.agentsEditor.SetWidth(editorWidth)
	a.agentsEditor.SetHeight(editorHeight)

	a.agentsPathInput = textinput.New()
	a.agentsPathInput.Placeholder = "Path to AGENTS.md file..."
	a.agentsPathInput.Width = inputWidthDefault

	a.skillSourceInput = textinput.New()
	a.skillSourceInput.Placeholder = "Path to skill source directory..."
	a.skillSourceInput.Width = inputWidthDefault

	a.personaSourceInput = textinput.New()
	a.personaSourceInput.Placeholder = "Path to persona source directory..."
	a.personaSourceInput.Width = inputWidthDefault
}

// Init starts periodic refresh and textinput blink commands.
func (a *App) Init() tea.Cmd {
	return tea.Batch(tick(), textinput.Blink)
}

func tick() tea.Cmd {
	return tea.Tick(1*time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}
