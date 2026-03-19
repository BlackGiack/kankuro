package ui

import (
	"github.com/charmbracelet/lipgloss"

	"kankuro/internal/config"
	"kankuro/internal/session"
)

type sessionListItem struct {
	title         string
	desc          string
	sessionDetail string
	id            string
	status        session.Status
}

func (i sessionListItem) Title() string {
	if i.status == session.StatusWaiting {
		return lipgloss.NewStyle().Foreground(highlightColor).Bold(true).Render(i.title)
	}
	return i.title
}

func (i sessionListItem) Description() string {
	baseDesc := i.desc
	if i.status == session.StatusWaiting {
		baseDesc = lipgloss.NewStyle().Foreground(highlightColor).Render(i.desc)
	}
	if i.sessionDetail != "" {
		return baseDesc + "\n" + i.sessionDetail
	}
	return baseDesc
}

func (i sessionListItem) FilterValue() string { return i.title }

type cardItem struct {
	card     config.IDCard
	skills   []config.Skill
	status   session.Status
	selected bool
}

func (i cardItem) Title() string       { return "" }
func (i cardItem) Description() string { return renderCardItem(i.card, i.skills, i.selected) }
func (i cardItem) FilterValue() string { return i.card.Name }
