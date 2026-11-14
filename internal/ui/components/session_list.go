package components

import (
	"fmt"
	"image/color"
	"sort"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/BlackGiack/kankuro/internal/ui/component"
	"github.com/BlackGiack/kankuro/internal/ui/models"
)

const (
	SessionDelegateHeight  = 4
	registrationErrorWidth = 18
	conversationLabelWidth = 12
	conversationIDWidth    = 12
	waitingStatus          = "waiting"
)

type SessionListItem struct {
	Session models.SessionRow
	Width   int
}

type SessionListCallbacks struct {
	OnSelect  func(models.SessionRow) tea.Cmd
	OnAttach  func(sessionID string)
	OnKill    func(sessionID string)
	OnDelete  func(sessionID string)
	OnSwitch  func(sessionID string)
	OnPreview func(sessionID string) tea.Cmd
}

type SessionList struct {
	component.Base

	list           list.Model
	sessions       []models.SessionRow
	searchQuery    string
	callbacks      SessionListCallbacks
	primaryColor   color.Color
	highlightColor color.Color
	gray           color.Color
}

func NewSessionList(callbacks SessionListCallbacks) *SessionList {
	delegate := list.NewDefaultDelegate()
	delegate.SetHeight(SessionDelegateHeight)
	delegate.Styles.NormalTitle = delegate.Styles.NormalTitle.Foreground(lipgloss.Color("#8E44AD"))
	delegate.Styles.NormalDesc = delegate.Styles.NormalDesc.Foreground(lipgloss.Color("240"))
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#16A085")).
		BorderForeground(lipgloss.Color("#8E44AD")).
		UnsetBackground()
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("#16A085")).
		BorderForeground(lipgloss.Color("#8E44AD")).
		UnsetBackground()
	delegate.Styles.DimmedTitle = delegate.Styles.DimmedTitle.Foreground(lipgloss.Color("240"))
	delegate.Styles.DimmedDesc = delegate.Styles.DimmedDesc.Foreground(lipgloss.Color("240"))
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	return &SessionList{
		list: l, callbacks: callbacks, primaryColor: lipgloss.Color("#8E44AD"),
		highlightColor: lipgloss.Color("#16A085"), gray: lipgloss.Color("240"),
	}
}

func (s *SessionList) Sessions() []models.SessionRow { return models.CloneSessionRows(s.sessions) }
func (s *SessionList) Select(index int)              { s.list.Select(index) }
func (s *SessionList) ItemCount() int                { return len(s.list.Items()) }
func (s *SessionList) ItemsOnPage() int {
	return s.list.Paginator.ItemsOnPage(len(s.list.Items()))
}

// SetRows replaces the visible session inventory and preserves selection by ID.
func (s *SessionList) SetRows(rows []models.SessionRow) {
	selectedID := ""
	if selected, ok := s.Selected(); ok {
		selectedID = selected.ID
	}
	s.sessions = models.CloneSessionRows(rows)
	s.refreshItems()
	if selectedID != "" {
		s.SetSelectedByID(selectedID)
	}
}

func (s *SessionList) SetSearchQuery(query string) { s.searchQuery = query; s.refreshItems() }
func (s *SessionList) Selected() (models.SessionRow, bool) {
	item := s.list.SelectedItem()
	if item == nil {
		return models.SessionRow{}, false
	}
	session, ok := item.(SessionListItem)
	if !ok {
		return models.SessionRow{}, false
	}
	return models.CloneSessionRow(session.Session), true
}
func (s *SessionList) SetSelectedByID(sessionID string) {
	for i, item := range s.list.Items() {
		if row, ok := item.(SessionListItem); ok && row.Session.ID == sessionID {
			s.list.Select(i)
			return
		}
	}
}
func (s *SessionList) Init() tea.Cmd { return nil }
func (s *SessionList) Update(msg tea.Msg) (component.Model, tea.Cmd) {
	prevIndex := s.list.Index()
	var cmd tea.Cmd
	s.list, cmd = s.list.Update(msg)
	if s.list.Index() != prevIndex {
		if previewCmd := s.handleSelectionChange(); previewCmd != nil {
			cmd = tea.Batch(cmd, previewCmd)
		}
	}
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		s.handleKey(keyMsg)
	}
	return s, cmd
}
func (s *SessionList) handleSelectionChange() tea.Cmd {
	if selected, ok := s.Selected(); ok && s.callbacks.OnPreview != nil {
		return s.callbacks.OnPreview(selected.ID)
	}
	return nil
}
func (s *SessionList) handleKey(key tea.KeyPressMsg) {
	switch key.Code {
	case tea.KeyEnter:
		s.handleEnter()
	case tea.KeyDelete, tea.KeyBackspace:
		s.handleDelete()
	default:
		switch key.Text {
		case "x":
			s.handleKill()
		case "s":
			s.handleSwitch()
		}
	}
}
func (s *SessionList) handleEnter() {
	if row, ok := s.Selected(); ok && s.callbacks.OnAttach != nil {
		s.callbacks.OnAttach(row.ID)
	}
}
func (s *SessionList) handleKill() {
	if row, ok := s.Selected(); ok && s.callbacks.OnKill != nil {
		s.callbacks.OnKill(row.ID)
	}
}
func (s *SessionList) handleDelete() {
	if row, ok := s.Selected(); ok && s.callbacks.OnDelete != nil {
		s.callbacks.OnDelete(row.ID)
	}
}
func (s *SessionList) handleSwitch() {
	if row, ok := s.Selected(); ok && s.callbacks.OnSwitch != nil {
		s.callbacks.OnSwitch(row.ID)
	}
}
func (s *SessionList) View() string { return s.list.View() }
func (s *SessionList) SetSize(width, height int) {
	s.Base.SetSize(width, height)
	s.list.SetSize(width, height)
}

func (s *SessionList) refreshItems() {
	previousID := ""
	if selected, ok := s.Selected(); ok {
		previousID = selected.ID
	}
	sorted := append([]models.SessionRow(nil), s.sessions...)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].StartedAt.After(sorted[j].StartedAt) })
	filtered := s.filterSessions(sorted)
	items := make([]list.Item, 0, len(filtered))
	for _, row := range filtered {
		items = append(items, SessionListItem{Session: models.CloneSessionRow(row), Width: s.Width})
	}
	s.list.SetItems(items)
	if previousID != "" {
		s.SetSelectedByID(previousID)
	}
}
func (s *SessionList) filterSessions(rows []models.SessionRow) []models.SessionRow {
	if s.searchQuery == "" {
		return rows
	}
	q := strings.ToLower(s.searchQuery)
	filtered := make([]models.SessionRow, 0, len(rows))
	for _, row := range rows {
		if strings.Contains(strings.ToLower(row.Name), q) || strings.Contains(strings.ToLower(row.CardID), q) {
			filtered = append(filtered, row)
		}
	}
	return filtered
}
func (s *SessionList) FilterSessions(rows []models.SessionRow) []models.SessionRow {
	return s.filterSessions(rows)
}

func (i SessionListItem) Title() string {
	if i.Session.Status == waitingStatus {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#FF69B4")).Bold(true).Render(DisplayText(i.Session.Name))
	}
	return DisplayText(i.Session.Name)
}
func (i SessionListItem) Description() string {
	cardName := DisplayText(i.Session.CardName)
	faceStyle := i.Session.FaceStyle
	if i.Session.Status == "waiting" || faceStyle == "" {
		faceStyle = "neutral"
	}
	cardLine := cardName + " " + RenderAgentFace(faceStyle)
	descLine := fmt.Sprintf("%s pid=%d", i.Session.Status, i.Session.PID)
	if i.Session.Registration != "" && i.Session.Registration != "unavailable" {
		descLine += " reg=" + i.Session.Registration
		if i.Session.RegistrationPersistence == "unsaved" {
			descLine += "/not-saved"
		}
		if i.Session.RegistrationError != "" {
			descLine += " !" + TruncateTail(DisplayText(i.Session.RegistrationError), registrationErrorWidth)
		}
	}
	if i.Session.ActiveConversationID != "" {
		name := TruncateTail(DisplayText(i.Session.ActiveConversationName), conversationLabelWidth)
		id := TruncateTail(DisplayText(i.Session.ActiveConversationID), conversationIDWidth)
		identity := name + ":" + id
		descLine += " " + identity
	}
	if i.Width > 0 {
		descLine = TruncateHead(DisplayText(descLine), i.Width)
	}
	const maxWorkspaceLen = 25
	detailLine := "WS: " + TruncateTail(DisplayText(i.Session.WorkingDir), maxWorkspaceLen)
	if i.Session.TemporaryWorkspace {
		detailLine = "ID: " + i.Session.ID
	}
	if i.Session.Status == "waiting" {
		cardLine = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF69B4")).Bold(true).Render(cardLine)
		descLine = lipgloss.NewStyle().Foreground(lipgloss.Color("#16A085")).Render(descLine)
	}
	return cardLine + "\n" + descLine + "\n" + detailLine
}
func (i SessionListItem) FilterValue() string { return i.Session.FilterKey }
