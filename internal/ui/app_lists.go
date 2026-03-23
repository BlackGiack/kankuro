package ui

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/bubbles/list"

	"kankuro/internal/config"
	"kankuro/internal/session"
)

func newCardDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	d.SetHeight(cardDelegateHeight)
	d.Styles.NormalTitle = d.Styles.NormalTitle.Foreground(primaryColor)
	d.Styles.NormalDesc = d.Styles.NormalDesc.Foreground(gray)

	// Keep same base styles for selected entries but adjust border/foreground to avoid double borders/backgrounds
	d.Styles.SelectedTitle = d.Styles.NormalTitle.
		Foreground(highlightColor).
		BorderForeground(primaryColor).
		UnsetBackground()
	d.Styles.SelectedDesc = d.Styles.NormalDesc.
		Foreground(highlightColor).
		BorderForeground(primaryColor).
		UnsetBackground()

	d.Styles.DimmedTitle = d.Styles.DimmedTitle.Foreground(gray)
	d.Styles.DimmedDesc = d.Styles.DimmedDesc.Foreground(gray)
	return d
}

func (a *App) updateCardSelection() {
	items := a.cardsList.Items()
	cursor := a.cardsList.Cursor()
	for i := range items {
		if c, ok := items[i].(cardItem); ok {
			c.selected = (i == cursor)
			items[i] = c
		}
	}
	a.cardsList.SetItems(items)
}

func (a *App) resize() {
	a.cardsList.SetSize(
		max(cardListMinWidth, a.width-cardsListPadding),
		max(cardListMinHeight, a.height-cardsListHeight),
	)
	a.sessionsList.SetSize(sessionListWidth, max(cardListMinHeight, a.height-sessionsListHeight))
	a.switchList.SetSize(
		max(switchListMinWidth, a.width-switchListPadding),
		max(cardListMinHeight, a.height-switchListHeightDiff),
	)
}

func (a *App) refreshLists() {
	a.refreshCardList()
	a.refreshSessionList()
}

func (a *App) refreshCardList() {
	items := toCardItems(a.cards, a.skills)
	a.cardsList.SetItems(items)
	a.updateCardSelection()
}

func (a *App) updateSwitchSelection() {
	items := a.switchList.Items()
	cursor := a.switchList.Cursor()
	for i := range items {
		if c, ok := items[i].(cardItem); ok {
			c.selected = (i == cursor)
			items[i] = c
		}
	}
	a.switchList.SetItems(items)
}

func (a *App) refreshSessionList() {
	a.sessions = a.session.List()
	sort.Slice(a.sessions, func(i, j int) bool { return a.sessions[i].StartedAt.After(a.sessions[j].StartedAt) })
	items := make([]list.Item, 0, len(a.sessions))
	for _, s := range a.sessions {
		faceStyle := faceNeutral
		for _, c := range a.cards {
			if c.ID == s.CardID {
				faceStyle = c.FaceStyle
				break
			}
		}

		if s.Status == session.StatusWaiting {
			faceStyle = faceNeutral
		}

		face := renderAgentCard(config.IDCard{FaceStyle: faceStyle})
		var sessionDetail string
		if s.UseTmpWorkspace {
			sessionDetail = "ID: " + s.ID
		} else {
			wsPath := s.WorkingDir
			if len(wsPath) > pathTruncateLimit {
				wsPath = "..." + wsPath[len(wsPath)-pathTruncateKeep:]
			}
			sessionDetail = "WS: " + wsPath
		}
		items = append(items, sessionListItem{
			title:         s.Name + " " + face,
			desc:          fmt.Sprintf("%s pid=%d", s.Status, s.PID),
			sessionDetail: sessionDetail,
			id:            s.ID,
			status:        s.Status,
		})
	}
	a.sessionsList.SetItems(items)
	if len(items) == 0 {
		a.focusedPane = paneCards
	}
}

func (a *App) selectedCard() (config.IDCard, bool) {
	it, ok := a.cardsList.SelectedItem().(cardItem)
	if !ok {
		return config.IDCard{}, false
	}
	return it.card, true
}

func (a *App) selectedSwitchCard() (config.IDCard, bool) {
	it, ok := a.switchList.SelectedItem().(cardItem)
	if !ok {
		return config.IDCard{}, false
	}
	return it.card, true
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

func toCardItems(cards []config.IDCard, allSkills []config.Skill) []list.Item {
	items := make([]list.Item, 0, len(cards))
	for _, c := range cards {
		skillList := make([]config.Skill, 0, len(c.SkillIDs))
		selected := make(map[string]struct{})
		for _, id := range c.SkillIDs {
			selected[id] = struct{}{}
		}
		for _, skill := range allSkills {
			if _, ok := selected[skill.ID]; ok {
				skillList = append(skillList, skill)
			}
		}
		items = append(items, cardItem{
			card:   c,
			skills: skillList,
			status: session.StatusRunning,
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
