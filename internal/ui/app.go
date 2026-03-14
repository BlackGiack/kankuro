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

	"kankuro/internal/ascii"
	"kankuro/internal/ascii/characters"
	"kankuro/internal/config"
	"kankuro/internal/session"
)

type viewMode int
type dashboardPane int

const (
	viewDashboard viewMode = iota
	viewCardForm
	viewMCPForm
	viewDeleteConfirm
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
		return lipgloss.NewStyle().Foreground(hotPink).Bold(true).Render(i.title)
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
	mcpInputs  []textinput.Model

	cardFocusIndex   int
	skillCursor      int
	selectedSkillIDs map[string]bool

	editingCardID   string
	activeSessionID string
	pendingDeleteID string
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

	cardInputs := make([]textinput.Model, 6)
	placeholders := []string{"name", "driver command (e.g. claude)", "model", "working_dir", "face_style (neutral/happy/focused/thinking)", "use tmp workspace (y/n)"}
	for i := range cardInputs {
		cardInputs[i] = textinput.New()
		cardInputs[i].Placeholder = placeholders[i]
		cardInputs[i].Width = 60
	}
	cardInputs[0].Focus()

	mcpInputs := make([]textinput.Model, 2)
	mcpInputs[0] = textinput.New()
	mcpInputs[0].Placeholder = "mcp name"
	mcpInputs[0].Width = 60
	mcpInputs[1] = textinput.New()
	mcpInputs[1].Placeholder = "mcp command"
	mcpInputs[1].Width = 60

	app := &App{
		repo:             repo,
		session:          sm,
		mode:             viewDashboard,
		cards:            cards,
		skills:           skills,
		mcps:             mcps,
		sessions:         sm.List(),
		cardsList:        cardsList,
		sessionsList:     sessionsList,
		cardInputs:       cardInputs,
		mcpInputs:        mcpInputs,
		selectedSkillIDs: make(map[string]bool),
		focusedPane:      paneCards,
		help:             help.New(),
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
	case viewMCPForm:
		return a.updateMCPForm(msg)
	case viewDeleteConfirm:
		return a.updateDeleteConfirm(msg)
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
			a.cardInputs[4].SetValue(selected.FaceStyle)
			tmpVal := "n"
			if selected.UseTmpWorkspace {
				tmpVal = "y"
			}
			a.cardInputs[5].SetValue(tmpVal)
			a.setSelectedSkills(selected.SkillIDs)
			a.editingCardID = selected.ID
			a.cardFocusIndex = 0
			a.setCardFocus(a.cardFocusIndex)
			a.mode = viewCardForm
			return a, nil
		case "m":
			a.mcpInputs[0].SetValue("")
			a.mcpInputs[1].SetValue("")
			a.mcpInputs[0].Focus()
			a.mode = viewMCPForm
			return a, nil
		case "enter":
			if a.focusedPane == paneSessions {
				return a, a.activateSelectedSession()
			}
			selected, ok := a.selectedCard()
			if !ok {
				return a, nil
			}
			skills := a.skillsForCard(selected)
			id, err := a.session.Launch(selected, skills)
			if err != nil {
				a.err = err
				return a, nil
			}
			a.activeSessionID = id
			a.mode = viewDashboard
			return a, a.attachSession(id)
		case "a":
			return a, a.activateSelectedSession()
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
			sel := a.sessionsList.Index()
			if sel < 0 || sel >= len(a.sessions) {
				return a, nil
			}
			s := a.sessions[sel]
			if s.UseTmpWorkspace {
				a.pendingDeleteID = s.ID
				a.mode = viewDeleteConfirm
				return a, nil
			}
			if err := a.session.Remove(s.ID); err != nil {
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
			a.cycleCardFocus()
			return a, nil
		case "up", "k":
			if a.isSkillListFocused() {
				if a.skillCursor > 0 {
					a.skillCursor--
				}
				return a, nil
			}
		case "down", "j":
			if a.isSkillListFocused() {
				if a.skillCursor < len(a.skills)-1 {
					a.skillCursor++
				}
				return a, nil
			}
		case " ":
			if a.isSkillListFocused() {
				a.toggleSelectedSkill()
				return a, nil
			}
		case "enter":
			card := config.IDCard{
				ID:              a.editingCardID,
				Name:            strings.TrimSpace(a.cardInputs[0].Value()),
				Driver:          strings.TrimSpace(a.cardInputs[1].Value()),
				Model:           strings.TrimSpace(a.cardInputs[2].Value()),
				WorkingDir:      strings.TrimSpace(a.cardInputs[3].Value()),
				FaceStyle:       strings.TrimSpace(a.cardInputs[4].Value()),
				UseTmpWorkspace: strings.ToLower(strings.TrimSpace(a.cardInputs[5].Value())) == "y",
				SkillIDs:        a.selectedSkillIDsSlice(),
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

func (a *App) updateDeleteConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			if err := a.session.Remove(a.pendingDeleteID); err != nil {
				a.err = err
			}
			a.pendingDeleteID = ""
			a.mode = viewDashboard
			a.refreshSessionList()
			return a, nil
		case "n", "N", "esc":
			a.pendingDeleteID = ""
			a.mode = viewDashboard
			return a, nil
		}
	}
	return a, nil
}

// View renders the active screen for the Bubble Tea program.
func (a *App) View() string {
	errLine := ""
	if a.err != nil {
		errLine = errorStyle.Render(a.err.Error()) + "\n"
	}

	switch a.mode {
	case viewDeleteConfirm:
		return errLine + titleStyle.Render("Delete Session?") + "\n\n" +
			fmt.Sprintf("Are you sure you want to delete session %s?\n", a.pendingDeleteID) +
			"This will also delete its temporary workspace if it exists.\n\n" +
			"y=Confirm  n=Cancel"
	case viewCardForm:
		title := "New ID Card"
		if a.editingCardID != "" {
			title = "Edit ID Card"
		}
		return errLine + titleStyle.Render(title) + "\n\n" +
			"Name: " + a.cardInputs[0].View() + "\n" +
			"Driver: " + a.cardInputs[1].View() + "\n" +
			"Model: " + a.cardInputs[2].View() + "\n" +
			"Working Dir: " + a.cardInputs[3].View() + "\n" +
			"Face Style: " + a.cardInputs[4].View() + "\n" +
			"Use Tmp Workspace: " + a.cardInputs[5].View() + "\n\n" +
			"Skills (space=toggle, arrows=scroll):\n" + a.renderSkillList() + "\n\n" +
			"Enter=save Tab=next Esc=cancel"
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
			"Tab=switch Enter=launch/attach x=kill Del=remove terminated n=new e=edit m=mcp q=quit"
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
		faceStyle := "neutral"
		for _, c := range a.cards {
			if c.ID == s.CardID {
				faceStyle = c.FaceStyle
				break
			}
		}

		if s.Status == session.StatusWaiting {
			faceStyle = "neutral"
		}

		face := renderAgentCard(config.IDCard{FaceStyle: faceStyle})
		items = append(items, item{
			title:  s.Name + " " + face,
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
	a.selectedSkillIDs = make(map[string]bool)
	a.skillCursor = 0
	a.cardFocusIndex = 0
	a.setCardFocus(a.cardFocusIndex)
}

func (a *App) setSelectedSkills(ids []string) {
	a.selectedSkillIDs = make(map[string]bool)
	for _, id := range ids {
		a.selectedSkillIDs[id] = true
	}
	a.skillCursor = 0
}

func (a *App) selectedSkillIDsSlice() []string {
	ids := make([]string, 0, len(a.selectedSkillIDs))
	for _, skill := range a.skills {
		if a.selectedSkillIDs[skill.ID] {
			ids = append(ids, skill.ID)
		}
	}
	return ids
}

func (a *App) isSkillListFocused() bool {
	return a.cardFocusIndex >= len(a.cardInputs)
}

func (a *App) setCardFocus(index int) {
	a.cardFocusIndex = index
	for i := range a.cardInputs {
		if i == index {
			a.cardInputs[i].Focus()
		} else {
			a.cardInputs[i].Blur()
		}
	}
}

func (a *App) cycleCardFocus() {
	total := len(a.cardInputs) + 1
	next := (a.cardFocusIndex + 1) % total
	a.setCardFocus(next)
}

func (a *App) toggleSelectedSkill() {
	if len(a.skills) == 0 || a.skillCursor < 0 || a.skillCursor >= len(a.skills) {
		return
	}
	id := a.skills[a.skillCursor].ID
	if a.selectedSkillIDs[id] {
		delete(a.selectedSkillIDs, id)
		return
	}
	a.selectedSkillIDs[id] = true
}

func (a *App) skillWindow(total int) (int, int) {
	if total == 0 {
		return 0, 0
	}
	height := a.skillListHeight()
	if height >= total {
		return 0, total
	}
	start := a.skillCursor - height/2
	if start < 0 {
		start = 0
	}
	end := start + height
	if end > total {
		end = total
		start = end - height
		if start < 0 {
			start = 0
		}
	}
	return start, end
}

func (a *App) skillListHeight() int {
	if a.height <= 0 {
		return 10
	}
	available := a.height - 12
	if available < 3 {
		return 3
	}
	return available
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

func renderAgentCard(card config.IDCard) string {
	style := card.FaceStyle
	if style == "" {
		style = "neutral"
	}
	face := characters.NewFace(ascii.WithStyle(style))
	w, _ := face.Bounds()
	canvas := ascii.NewCanvas(w, 1)
	face.Render(canvas, 0, 0)
	return canvas.String()
}

func (a *App) skillsForCard(card config.IDCard) []config.Skill {
	if len(card.SkillIDs) == 0 {
		return nil
	}
	selected := make(map[string]struct{}, len(card.SkillIDs))
	for _, id := range card.SkillIDs {
		selected[id] = struct{}{}
	}
	out := make([]config.Skill, 0, len(card.SkillIDs))
	for _, skill := range a.skills {
		if _, ok := selected[skill.ID]; ok {
			out = append(out, skill)
		}
	}
	return out
}

func skillNameCounts(skills []config.Skill) map[string]int {
	counts := make(map[string]int)
	for _, skill := range skills {
		counts[skill.Name]++
	}
	return counts
}

func skillLabel(skill config.Skill, counts map[string]int) string {
	if counts[skill.Name] > 1 && skill.Source != "" {
		return fmt.Sprintf("%s (%s)", skill.Name, skill.Source)
	}
	return skill.Name
}

func (a *App) renderSkillList() string {
	if len(a.skills) == 0 {
		return "(no skills found)"
	}
	counts := skillNameCounts(a.skills)
	start, end := a.skillWindow(len(a.skills))
	lines := make([]string, 0, end-start+2)
	if start > 0 {
		lines = append(lines, "  ...")
	}
	for i := start; i < end; i++ {
		skill := a.skills[i]
		prefix := "  "
		if a.isSkillListFocused() && i == a.skillCursor {
			prefix = "> "
		}
		mark := " "
		if a.selectedSkillIDs[skill.ID] {
			mark = "x"
		}
		line := fmt.Sprintf("%s[%s] %s", prefix, mark, skillLabel(skill, counts))
		if a.isSkillListFocused() && i == a.skillCursor {
			line = lipgloss.NewStyle().Foreground(hotPink).Render(line)
		}
		lines = append(lines, line)
	}
	if end < len(a.skills) {
		lines = append(lines, "  ...")
	}
	return strings.Join(lines, "\n")
}

func toCardItems(cards []config.IDCard) []list.Item {
	items := make([]list.Item, 0, len(cards))
	for _, c := range cards {
		items = append(items, item{
			title:  c.Name + " " + renderAgentCard(c),
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

func (a *App) toggleDashboardPane() {
	if a.focusedPane == paneCards && len(a.sessions) > 0 {
		a.focusedPane = paneSessions
		return
	}
	a.focusedPane = paneCards
}

func (a *App) activateSelectedSession() tea.Cmd {
	if len(a.sessions) == 0 {
		return nil
	}
	sel := a.sessionsList.Index()
	if sel < 0 || sel >= len(a.sessions) {
		return nil
	}
	s := a.sessions[sel]

	if s.Status == session.StatusTerminated {
		var card config.IDCard
		found := false
		for _, c := range a.cards {
			if c.ID == s.CardID {
				card = c
				found = true
				break
			}
		}
		if !found {
			a.err = fmt.Errorf("cannot relaunch: original ID card not found")
			return nil
		}
		skills := a.skillsForCard(card)
		if err := a.session.Relaunch(s.ID, card, skills); err != nil {
			a.err = fmt.Errorf("relaunch failed: %w", err)
			return nil
		}
		// Refresh locally to reflect status change immediately
		a.sessions = a.session.List()
		a.refreshSessionList()
	}

	a.activeSessionID = s.ID
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
