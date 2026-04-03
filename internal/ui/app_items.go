package ui

import (
	"github.com/charmbracelet/lipgloss"

	"kankuro/internal/config"
	"kankuro/internal/session"
)

type sessionListItem struct {
	sessionName   string
	cardNameFace  string
	desc          string
	sessionDetail string
	id            string
	status        session.Status
}

func (i sessionListItem) Title() string {
	if i.status == session.StatusWaiting {
		return lipgloss.NewStyle().Foreground(highlightColor).Bold(true).Render(i.sessionName)
	}
	return i.sessionName
}

func (i sessionListItem) Description() string {
	cardLine := i.cardNameFace
	if i.status == session.StatusWaiting {
		cardLine = lipgloss.NewStyle().Foreground(highlightColor).Bold(true).Render(i.cardNameFace)
	}
	descLine := i.desc
	if i.status == session.StatusWaiting {
		descLine = lipgloss.NewStyle().Foreground(highlightColor).Render(i.desc)
	}
	detailLine := i.sessionDetail
	if detailLine != "" {
		return cardLine + "\n" + descLine + "\n" + detailLine
	}
	return cardLine + "\n" + descLine
}

func (i sessionListItem) FilterValue() string { return i.sessionName }

type cardItem struct {
	card     config.IDCard
	skills   []config.Skill
	personas []config.Persona
	status   session.Status
}

func (i cardItem) Title() string       { return "" }
func (i cardItem) Description() string { return "" }
func (i cardItem) FilterValue() string { return i.card.Name }
