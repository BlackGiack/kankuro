package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"kankuro/internal/config"
	"kankuro/internal/session"
)

type viewMode int
type dashboardPane int

const (
	viewDashboard viewMode = iota
	viewCardForm
	viewSkillForm
	viewMCPForm
)

const (
	paneCards dashboardPane = iota
	paneSessions
)

type item struct {
	title  string
	desc   string
	id     string
	status session.Status
}

func (i item) Title() string {
	if i.status == session.StatusWaiting {
		return lipgloss.NewStyle().Foreground(hotPink).Bold(true).Render("󰚩 " + i.title)
	}
	return i.title
}

func (i item) Description() string {
	if i.status == session.StatusWaiting {
		return lipgloss.NewStyle().Foreground(hotPink).Render(i.desc)
	}
	return i.desc
}

func (i item) FilterValue() string { return i.title }

type tickMsg time.Time

type attachFinishedMsg struct {
	sessionID string
	err       error
}

type App struct {
	repo    config.Repository
	session *session.Manager

	mode viewMode

	cards    []config.IDCard
	skills   []config.Skill
	mcps     []config.MCPServer
	sessions []session.Snapshot

	cardsList    list.Model
	sessionsList list.Model

	cardInputs []textinput.Model
	skillInput textinput.Model
	mcpInputs  []textinput.Model

	editingCardID   string
	activeSessionID string
	focusedPane     dashboardPane

	help help.Model

	width  int
	height int
	err    error
}

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

	delegate := list.NewDefaultDelegate()
	delegate.Styles.NormalTitle = delegate.Styles.NormalTitle.Foreground(purple)
	delegate.Styles.NormalDesc = delegate.Styles.NormalDesc.Foreground(gray)
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(hotPink).
		BorderForeground(purple)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(darkPurple).
		BorderForeground(purple)
	delegate.Styles.DimmedTitle = delegate.Styles.DimmedTitle.Foreground(darkPurple)
	delegate.Styles.DimmedDesc = delegate.Styles.DimmedDesc.Foreground(lipgloss.Color("236"))

	cardsList := list.New(toCardItems(cards), delegate, 0, 0)
	cardsList.Title = "ID Cards"
	cardsList.SetShowStatusBar(false)
	cardsList.SetFilteringEnabled(false)
	cardsList.Styles.Title = titleStyle

	sessionsList := list.New([]list.Item{}, delegate, 0, 0)
	sessionsList.Title = "Active Sessions"
	sessionsList.SetShowStatusBar(false)
	sessionsList.SetFilteringEnabled(false)
	sessionsList.Styles.Title = titleStyle

	cardInputs := make([]textinput.Model, 4)
	placeholders := []string{"name", "driver command (e.g. claude)", "model", "working_dir"}
	for i := range cardInputs {
		cardInputs[i] = textinput.New()
		cardInputs[i].Placeholder = placeholders[i]
		cardInputs[i].Width = 60
	}
	cardInputs[0].Focus()

	skillInput := textinput.New()
	skillInput.Placeholder = "skill name"
	skillInput.Width = 60

	mcpInputs := make([]textinput.Model, 2)
	mcpInputs[0] = textinput.New()
	mcpInputs[0].Placeholder = "mcp name"
	mcpInputs[0].Width = 60
	mcpInputs[1] = textinput.New()
	mcpInputs[1].Placeholder = "mcp command"
	mcpInputs[1].Width = 60

	app := &App{
		repo:         repo,
		session:      sm,
		mode:         viewDashboard,
		cards:        cards,
		skills:       skills,
		mcps:         mcps,
		sessions:     sm.List(),
		cardsList:    cardsList,
		sessionsList: sessionsList,
		cardInputs:   cardInputs,
		skillInput:   skillInput,
		mcpInputs:    mcpInputs,
		focusedPane:  paneCards,
		help:         help.New(),
	}
	app.refreshLists()
	return app, nil
}

// Init starts periodic refresh and textinput blink commands.
func (a *App) Init() tea.Cmd {
	return tea.Batch(tick(), textinput.Blink)
}

func tick() tea.Cmd {
	return tea.Tick(1*time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

// Update handles key events, timers, and view-specific state transitions.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.resize()
	case tickMsg:
		a.sessions = a.session.List()
		a.refreshSessionList()
		return a, tick()
	case attachFinishedMsg:
		a.activeSessionID = msg.sessionID
		a.mode = viewDashboard
		a.refreshSessionList()
		if msg.err != nil {
			a.err = msg.err
		}
	}

	switch a.mode {
	case viewDashboard:
		return a.updateDashboard(msg)
	case viewCardForm:
		return a.updateCardForm(msg)
	case viewSkillForm:
		return a.updateSkillForm(msg)
	case viewMCPForm:
		return a.updateMCPForm(msg)
	default:
		return a, nil
	}
}

func (a *App) updateDashboard(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			a.session.Shutdown()
			return a, tea.Quit
		case "tab", "shift+tab", "left", "right":
			a.toggleDashboardPane()
			return a, nil
		case "n":
			a.resetCardForm()
			a.editingCardID = ""
			a.mode = viewCardForm
			return a, nil
		case "e":
			if a.focusedPane != paneCards {
				return a, nil
			}
			selected, ok := a.selectedCard()
			if !ok {
				return a, nil
			}
			a.cardInputs[0].SetValue(selected.Name)
			a.cardInputs[1].SetValue(selected.Driver)
			a.cardInputs[2].SetValue(selected.Model)
			a.cardInputs[3].SetValue(selected.WorkingDir)
			a.editingCardID = selected.ID
			a.mode = viewCardForm
			return a, nil
		case "s":
			a.skillInput.SetValue("")
			a.skillInput.Focus()
			a.mode = viewSkillForm
			return a, nil
		case "m":
			a.mcpInputs[0].SetValue("")
			a.mcpInputs[1].SetValue("")
			a.mcpInputs[0].Focus()
			a.mode = viewMCPForm
			return a, nil
		case "enter":
			if a.focusedPane == paneSessions {
				return a, a.attachSelectedSession()
			}
			selected, ok := a.selectedCard()
			if !ok {
				return a, nil
			}
			id, err := a.session.Launch(selected)
			if err != nil {
				a.err = err
				return a, nil
			}
			a.activeSessionID = id
			a.mode = viewDashboard
			return a, a.attachSession(id)
		case "a":
			return a, a.attachSelectedSession()
		case "x":
			if a.focusedPane != paneSessions {
				return a, nil
			}
			if err := a.killSelectedSession(); err != nil {
				a.err = err
			}
			a.refreshSessionList()
			return a, nil
		case "backspace", "delete":
			if a.focusedPane != paneSessions {
				return a, nil
			}
			if err := a.removeSelectedSession(); err != nil {
				a.err = err
			}
			a.refreshSessionList()
			return a, nil
		}
	}

	if a.focusedPane == paneSessions {
		var cmd tea.Cmd
		a.sessionsList, cmd = a.sessionsList.Update(msg)
		return a, cmd
	}
	var cmd tea.Cmd
	a.cardsList, cmd = a.cardsList.Update(msg)
	return a, cmd
}

func (a *App) updateCardForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			a.mode = viewDashboard
			return a, nil
		case "tab":
			a.cycleFocus(a.cardInputs)
			return a, nil
		case "enter":
			card := config.IDCard{
				ID:         a.editingCardID,
				Name:       strings.TrimSpace(a.cardInputs[0].Value()),
				Driver:     strings.TrimSpace(a.cardInputs[1].Value()),
				Model:      strings.TrimSpace(a.cardInputs[2].Value()),
				WorkingDir: strings.TrimSpace(a.cardInputs[3].Value()),
			}
			var err error
			if a.editingCardID == "" {
				card, err = a.repo.CreateIDCard(card)
			} else {
				card, err = a.repo.UpdateIDCard(card)
			}
			if err != nil {
				a.err = err
				return a, nil
			}
			a.upsertCardLocal(card)
			a.refreshCardList()
			a.editingCardID = ""
			a.mode = viewDashboard
			return a, nil
		}
	}
	cmds := make([]tea.Cmd, len(a.cardInputs))
	for i := range a.cardInputs {
		a.cardInputs[i], cmds[i] = a.cardInputs[i].Update(msg)
	}
	return a, tea.Batch(cmds...)
}

func (a *App) updateSkillForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			a.mode = viewDashboard
			return a, nil
		case "enter":
			skill, err := a.repo.UpsertSkill(config.Skill{Name: strings.TrimSpace(a.skillInput.Value())})
			if err != nil {
				a.err = err
				return a, nil
			}
			a.skills = append(a.skills, skill)
			a.mode = viewDashboard
			return a, nil
		}
	}
	var cmd tea.Cmd
	a.skillInput, cmd = a.skillInput.Update(msg)
	return a, cmd
}

func (a *App) updateMCPForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			a.mode = viewDashboard
			return a, nil
		case "tab":
			a.cycleFocus(a.mcpInputs)
			return a, nil
		case "enter":
			mcp, err := a.repo.UpsertMCP(config.MCPServer{
				Name:    strings.TrimSpace(a.mcpInputs[0].Value()),
				Command: strings.TrimSpace(a.mcpInputs[1].Value()),
			})
			if err != nil {
				a.err = err
				return a, nil
			}
			a.mcps = append(a.mcps, mcp)
			a.mode = viewDashboard
			return a, nil
		}
	}
	cmds := make([]tea.Cmd, len(a.mcpInputs))
	for i := range a.mcpInputs {
		a.mcpInputs[i], cmds[i] = a.mcpInputs[i].Update(msg)
	}
	return a, tea.Batch(cmds...)
}

// View renders the active screen for the Bubble Tea program.
func (a *App) View() string {
	errLine := ""
	if a.err != nil {
		errLine = errorStyle.Render(a.err.Error()) + "\n"
	}

	switch a.mode {
	case viewCardForm:
		title := "New ID Card"
		if a.editingCardID != "" {
			title = "Edit ID Card"
		}
		return errLine + titleStyle.Render(title) + "\n\n" +
			"Name: " + a.cardInputs[0].View() + "\n" +
			"Driver: " + a.cardInputs[1].View() + "\n" +
			"Model: " + a.cardInputs[2].View() + "\n" +
			"Working Dir: " + a.cardInputs[3].View() + "\n\n" +
			"Enter=save Tab=next Esc=cancel"
	case viewSkillForm:
		return errLine + titleStyle.Render("Add Skill") + "\n\n" +
			"Name: " + a.skillInput.View() + "\n\n" +
			"Enter=save Esc=cancel"
	case viewMCPForm:
		return errLine + titleStyle.Render("Add MCP") + "\n\n" +
			"Name: " + a.mcpInputs[0].View() + "\n" +
			"Command: " + a.mcpInputs[1].View() + "\n\n" +
			"Enter=save Tab=next Esc=cancel"
	default:
		leftStyle := lipgloss.NewStyle().Width(max(40, a.width-34)).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(normalBorderColor)
		rightStyle := lipgloss.NewStyle().Width(32).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(normalBorderColor)

		if a.focusedPane == paneCards {
			leftStyle = leftStyle.BorderForeground(focusedBorderColor)
		} else {
			rightStyle = rightStyle.BorderForeground(focusedBorderColor)
		}

		left := a.cardsList.View()
		right := a.sessionsList.View()
		body := lipgloss.JoinHorizontal(lipgloss.Top,
			leftStyle.Render(left),
			rightStyle.Render(right),
		)
		return errLine + titleStyle.Render("KANKURO Dashboard") + "\n" + body + "\n" +
			"Tab=switch Enter=launch/attach x=kill Del=remove terminated n=new e=edit s=skill m=mcp q=quit"
	}
}

func (a *App) resize() {
	a.cardsList.SetSize(max(20, a.width-34), max(10, a.height-4))
	a.sessionsList.SetSize(32, max(10, a.height-4))
}

func (a *App) refreshLists() {
	a.refreshCardList()
	a.refreshSessionList()
}

func (a *App) refreshCardList() {
	items := toCardItems(a.cards)
	a.cardsList.SetItems(items)
}

func (a *App) refreshSessionList() {
	a.sessions = a.session.List()
	sort.Slice(a.sessions, func(i, j int) bool { return a.sessions[i].StartedAt.After(a.sessions[j].StartedAt) })
	items := make([]list.Item, 0, len(a.sessions))
	for _, s := range a.sessions {
		items = append(items, item{
			title:  s.Name,
			desc:   fmt.Sprintf("%s pid=%d", s.Status, s.PID),
			id:     s.ID,
			status: s.Status,
		})
	}
	a.sessionsList.SetItems(items)
	if len(items) == 0 {
		a.focusedPane = paneCards
	}
}

func (a *App) selectedCard() (config.IDCard, bool) {
	it, ok := a.cardsList.SelectedItem().(item)
	if !ok {
		return config.IDCard{}, false
	}
	for _, c := range a.cards {
		if c.ID == it.id {
			return c, true
		}
	}
	return config.IDCard{}, false
}

func (a *App) resetCardForm() {
	for i := range a.cardInputs {
		a.cardInputs[i].SetValue("")
		a.cardInputs[i].Blur()
	}
	a.editingCardID = ""
	a.cardInputs[0].Focus()
}

func (a *App) cycleFocus(inputs []textinput.Model) {
	focused := -1
	for i := range inputs {
		if inputs[i].Focused() {
			focused = i
			inputs[i].Blur()
			break
		}
	}
	next := (focused + 1) % len(inputs)
	inputs[next].Focus()
}

func toCardItems(cards []config.IDCard) []list.Item {
	items := make([]list.Item, 0, len(cards))
	for _, c := range cards {
		items = append(items, item{
			title:  c.Name,
			desc:   fmt.Sprintf("%s | %s", c.Driver, c.Model),
			id:     c.ID,
			status: session.StatusRunning, // default for cards
		})
	}
	return items
}

func (a *App) upsertCardLocal(card config.IDCard) {
	for i := range a.cards {
		if a.cards[i].ID == card.ID {
			a.cards[i] = card
			return
		}
	}
	a.cards = append(a.cards, card)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (a *App) toggleDashboardPane() {
	if a.focusedPane == paneCards && len(a.sessions) > 0 {
		a.focusedPane = paneSessions
		return
	}
	a.focusedPane = paneCards
}

func (a *App) attachSelectedSession() tea.Cmd {
	if len(a.sessions) == 0 {
		return nil
	}
	sel := a.sessionsList.Index()
	if sel < 0 || sel >= len(a.sessions) {
		return nil
	}
	a.activeSessionID = a.sessions[sel].ID
	a.mode = viewDashboard
	return a.attachSession(a.activeSessionID)
}

func (a *App) killSelectedSession() error {
	if len(a.sessions) == 0 {
		return nil
	}
	sel := a.sessionsList.Index()
	if sel < 0 || sel >= len(a.sessions) {
		return nil
	}
	return a.session.Kill(a.sessions[sel].ID)
}

func (a *App) removeSelectedSession() error {
	if len(a.sessions) == 0 {
		return nil
	}
	sel := a.sessionsList.Index()
	if sel < 0 || sel >= len(a.sessions) {
		return nil
	}
	return a.session.Remove(a.sessions[sel].ID)
}
