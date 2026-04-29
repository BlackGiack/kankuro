package components

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"kankuro/internal/config"
	"kankuro/internal/session"
	"kankuro/internal/ui/component"
)

// Session styling .
const (
	SessionDelegateHeight = 4
)

// SessionListItem wraps session.Snapshot for list display.
type SessionListItem struct {
	Session session.Snapshot
	Card    *config.IDCard // Reference to card for face style
}

// SessionListCallbacks defines callbacks for session actions.
type SessionListCallbacks struct {
	OnSelect  func(session.Snapshot) tea.Cmd
	OnAttach  func(sessionID string)
	OnKill    func(sessionID string)
	OnDelete  func(sessionID string)
	OnSwitch  func(sessionID string)
	OnPreview func(sessionID string) tea.Cmd // Called when selection changes
}

// SessionList is a scrollable list of active sessions.
type SessionList struct {
	component.Base

	list        list.Model
	sessions    []session.Snapshot
	cards       []config.IDCard
	searchQuery string
	callbacks   SessionListCallbacks

	// Theme colors (should come from shared styles)
	primaryColor   lipgloss.Color
	highlightColor lipgloss.Color
	gray           lipgloss.Color
}

// NewSessionList creates a new session list component.
func NewSessionList(callbacks SessionListCallbacks) *SessionList {
	// Create default delegate with custom styling
	delegate := list.NewDefaultDelegate()
	delegate.SetHeight(SessionDelegateHeight)

	// Apply custom styles
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

	return &SessionList{
		list:           l,
		callbacks:      callbacks,
		primaryColor:   lipgloss.Color("#8E44AD"),
		highlightColor: lipgloss.Color("#16A085"),
		gray:           lipgloss.Color("240"),
	}
}

func (s *SessionList) List() *list.Model { return &s.list }

func (s *SessionList) Sessions() []session.Snapshot { return s.sessions }

// SetSessions updates the session list data.
func (s *SessionList) SetSessions(sessions []session.Snapshot, cards []config.IDCard) {
	s.sessions = sessions
	s.cards = cards
	s.refreshItems()
}

// SetSearchQuery filters the displayed sessions.
func (s *SessionList) SetSearchQuery(query string) {
	s.searchQuery = query
	s.refreshItems()
}

// Selected returns the currently selected session.
func (s *SessionList) Selected() (session.Snapshot, bool) {
	item := s.list.SelectedItem()
	if item == nil {
		return session.Snapshot{}, false
	}

	sessionItem, ok := item.(SessionListItem)
	if !ok {
		return session.Snapshot{}, false
	}
	return sessionItem.Session, true
}

// SetSelectedByID selects a session by ID.
func (s *SessionList) SetSelectedByID(sessionID string) {
	for i, item := range s.list.Items() {
		if si, ok := item.(SessionListItem); ok {
			if si.Session.ID == sessionID {
				s.list.Select(i)
				return
			}
		}
	}
}

// Init initializes the component.
func (s *SessionList) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (s *SessionList) Update(msg tea.Msg) (component.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Track previous selection for preview sync
	prevIndex := s.list.Index()

	s.list, cmd = s.list.Update(msg)

	// Handle selection change
	if s.list.Index() != prevIndex {
		pCmd := s.handleSelectionChange()
		cmd = tea.Batch(cmd, pCmd)
	}

	// Handle key events
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		s.handleKey(keyMsg)
	}

	return s, cmd
}

func (s *SessionList) handleSelectionChange() tea.Cmd {
	if sel, ok := s.Selected(); ok {
		if s.callbacks.OnPreview != nil {
			return s.callbacks.OnPreview(sel.ID)
		}
	}
	return nil
}

func (s *SessionList) handleKey(keyMsg tea.KeyMsg) {
	//nolint:exhaustive // not all keys need to be handled
	switch keyMsg.Type {
	case tea.KeyEnter:
		s.handleEnter()
	case tea.KeyDelete, tea.KeyBackspace:
		s.handleDelete()
	case tea.KeyRunes:
		switch keyMsg.String() {
		case "x":
			s.handleKill()
		case "s":
			s.handleSwitch()
		}
	}
}

func (s *SessionList) handleEnter() {
	if session, ok := s.Selected(); ok {
		if s.callbacks.OnAttach != nil {
			s.callbacks.OnAttach(session.ID)
		}
	}
}

func (s *SessionList) handleKill() {
	if session, ok := s.Selected(); ok {
		if s.callbacks.OnKill != nil {
			s.callbacks.OnKill(session.ID)
		}
	}
}

func (s *SessionList) handleDelete() {
	if session, ok := s.Selected(); ok {
		if s.callbacks.OnDelete != nil {
			s.callbacks.OnDelete(session.ID)
		}
	}
}

func (s *SessionList) handleSwitch() {
	if session, ok := s.Selected(); ok {
		if s.callbacks.OnSwitch != nil {
			s.callbacks.OnSwitch(session.ID)
		}
	}
}

// View renders the session list.
func (s *SessionList) View() string {
	return s.list.View()
}

const listWidthOffset = 2

// SetSize implements component.Model.
func (s *SessionList) SetSize(width, height int) {
	s.Base.SetSize(width, height)
	s.list.SetSize(width-listWidthOffset, height)
}

// refreshItems rebuilds the list items from sessions.
func (s *SessionList) refreshItems() {
	// Sort sessions by StartedAt (newest first)
	sorted := make([]session.Snapshot, len(s.sessions))
	copy(sorted, s.sessions)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].StartedAt.After(sorted[j].StartedAt)
	})

	// Apply search filter
	filtered := s.filterSessions(sorted)

	// Convert to list items
	items := make([]list.Item, 0, len(filtered))
	for _, sess := range filtered {
		// Find card for face style
		var card *config.IDCard
		for i := range s.cards {
			if s.cards[i].ID == sess.CardID {
				card = &s.cards[i]
				break
			}
		}

		items = append(items, SessionListItem{
			Session: sess,
			Card:    card,
		})
	}

	s.list.SetItems(items)
}

// filterSessions applies the search query filter.
func (s *SessionList) filterSessions(sessions []session.Snapshot) []session.Snapshot {
	if s.searchQuery == "" {
		return sessions
	}

	q := strings.ToLower(s.searchQuery)
	var filtered []session.Snapshot
	for _, sess := range sessions {
		if strings.Contains(strings.ToLower(sess.Name), q) ||
			strings.Contains(strings.ToLower(sess.CardID), q) {
			filtered = append(filtered, sess)
		}
	}
	return filtered
}

// FilterSessions exposes the filter logic for testing.
func (s *SessionList) FilterSessions(sessions []session.Snapshot) []session.Snapshot {
	return s.filterSessions(sessions)
}

// ===== SessionListItem implements list.Item =====

func (i SessionListItem) Title() string {
	if i.Session.Status == session.StatusWaiting {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF69B4")).
			Bold(true).
			Render(i.Session.Name)
	}
	return i.Session.Name
}

func (i SessionListItem) Description() string {
	// Find card name for display
	cardName := i.Session.CardID
	if i.Card != nil {
		cardName = i.Card.Name
	}

	// Render face
	faceStyle := "neutral"
	if i.Session.Status != session.StatusWaiting && i.Card != nil && i.Card.FaceStyle != "" {
		faceStyle = i.Card.FaceStyle
	}
	face := RenderAgentFace(faceStyle)

	cardLine := cardName + " " + face
	descLine := fmt.Sprintf("%s pid=%d", i.Session.Status, i.Session.PID)

	const (
		maxWorkspaceLen = 25
		ellipsisLen     = 3
	)

	// Session detail (ID or working dir)
	var detailLine string
	if i.Session.UseTmpWorkspace {
		detailLine = "ID: " + i.Session.ID
	} else {
		wsPath := i.Session.WorkingDir
		if len(wsPath) > maxWorkspaceLen {
			// Trim workspace path to fit, accounting for ellipsis length (3 chars)
			wsPath = "..." + wsPath[len(wsPath)-(maxWorkspaceLen-ellipsisLen):]
		}
		detailLine = "WS: " + wsPath
	}

	// Apply waiting status styling
	if i.Session.Status == session.StatusWaiting {
		cardLine = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF69B4")).
			Bold(true).
			Render(cardLine)
		descLine = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#16A085")).
			Render(descLine)
	}

	return cardLine + "\n" + descLine + "\n" + detailLine
}

func (i SessionListItem) FilterValue() string {
	return i.Session.Name
}
