package ui

import (
	"fmt"
	"io"
	"sort"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"kankuro/internal/config"
	"kankuro/internal/session"
)

type cardDelegate struct {
	height   int
	skills   []config.Skill
	personas []config.Persona
}

func newCardDelegate(skills []config.Skill, personas []config.Persona) cardDelegate {
	return cardDelegate{
		height:   cardDelegateHeight,
		skills:   skills,
		personas: personas,
	}
}

func (d cardDelegate) Height() int                             { return d.height }
func (d cardDelegate) Spacing() int                            { return 0 }
func (d cardDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d cardDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	c, ok := item.(cardItem)
	if !ok {
		return
	}
	selected := index == m.Index()
	content := renderCardItem(c.card, d.skills, d.personas, selected)
	fmt.Fprint(w, content)
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
	items := toCardItems(a.cards, a.skills, a.personas)
	a.cardsList.SetItems(items)
}

func (a *App) refreshSessionList() {
	a.sessions = a.session.List()
	sort.Slice(a.sessions, func(i, j int) bool { return a.sessions[i].StartedAt.After(a.sessions[j].StartedAt) })
	items := make([]list.Item, 0, len(a.sessions))
	for _, s := range a.sessions {
		faceStyle := faceNeutral
		cardName := s.CardID
		for _, c := range a.cards {
			if c.ID == s.CardID {
				faceStyle = c.FaceStyle
				cardName = c.Name
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
			sessionName:   s.Name,
			cardNameFace:  cardName + " " + face,
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

func (a *App) personasForCard(card config.IDCard) []config.Persona {
	if len(card.PersonaIDs) == 0 {
		return nil
	}
	selected := make(map[string]struct{}, len(card.PersonaIDs))
	for _, id := range card.PersonaIDs {
		selected[id] = struct{}{}
	}
	out := make([]config.Persona, 0, len(card.PersonaIDs))
	for _, p := range a.personas {
		if _, ok := selected[p.ID]; ok {
			out = append(out, p)
		}
	}
	return out
}

func (a *App) mcpsForCard(card config.IDCard) []config.MCPServer {
	if len(card.MCPIDs) == 0 {
		return nil
	}
	selected := make(map[string]struct{}, len(card.MCPIDs))
	for _, id := range card.MCPIDs {
		selected[id] = struct{}{}
	}
	out := make([]config.MCPServer, 0, len(card.MCPIDs))
	for _, mcp := range a.mcps {
		if _, ok := selected[mcp.ID]; ok {
			out = append(out, mcp)
		}
	}
	return out
}

func (a *App) injectMCPCliSkill(card config.IDCard, skills []config.Skill) []config.Skill {
	if len(card.MCPIDs) == 0 {
		return skills
	}
	for _, s := range skills {
		if s.Name == "mcp-cli" {
			return skills
		}
	}
	for _, s := range a.skills {
		if s.Name == "mcp-cli" {
			return append(skills, s)
		}
	}
	return skills
}

func toCardItems(cards []config.IDCard, allSkills []config.Skill, allPersonas []config.Persona) []list.Item {
	items := make([]list.Item, 0, len(cards))
	for _, c := range cards {
		skillList := make([]config.Skill, 0, len(c.SkillIDs))
		selectedSkills := make(map[string]struct{})
		for _, id := range c.SkillIDs {
			selectedSkills[id] = struct{}{}
		}
		for _, skill := range allSkills {
			if _, ok := selectedSkills[skill.ID]; ok {
				skillList = append(skillList, skill)
			}
		}

		personaList := make([]config.Persona, 0, len(c.PersonaIDs))
		selectedPersonas := make(map[string]struct{})
		for _, id := range c.PersonaIDs {
			selectedPersonas[id] = struct{}{}
		}
		for _, p := range allPersonas {
			if _, ok := selectedPersonas[p.ID]; ok {
				personaList = append(personaList, p)
			}
		}

		items = append(items, cardItem{
			card:     c,
			skills:   skillList,
			personas: personaList,
			status:   session.StatusRunning,
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
