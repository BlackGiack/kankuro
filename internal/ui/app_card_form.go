package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"kankuro/internal/config"
)

func (a *App) updateCardForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if a.cardFocusIndex < len(a.cardInputs) && !a.cardInputs[a.cardFocusIndex].Focused() {
		a.setCardFocus(a.cardFocusIndex)
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "ctrl+c":
			return a, tea.Quit
		case keyEsc:
			a.mode = viewDashboard
			return a, nil
		case keyTab:
			a.cycleCardFocus()
			return a, nil
		case "up", "k":
			return a.handleCardFormNav(false)
		case "down", "j":
			return a.handleCardFormNav(true)
		case " ":
			return a.handleCardFormToggle()
		case keyEnter:
			return a.handleCardFormSubmit()
		}
	}
	cmds := make([]tea.Cmd, len(a.cardInputs))
	for i := range a.cardInputs {
		a.cardInputs[i], cmds[i] = a.cardInputs[i].Update(msg)
	}
	return a, tea.Batch(cmds...)
}

func (a *App) handleCardFormNav(down bool) (tea.Model, tea.Cmd) {
	if a.isSkillListFocused() {
		if down {
			if a.skillCursor < len(a.skills)-1 {
				a.skillCursor++
			}
		} else if a.skillCursor > 0 {
			a.skillCursor--
		}
		return a, nil
	}
	if a.isAgentsMDFocused() {
		a.agentsMDPreferPath = down
		return a, nil
	}
	return a, nil
}

func (a *App) handleCardFormToggle() (tea.Model, tea.Cmd) {
	if a.isSkillListFocused() {
		a.toggleSelectedSkill()
		return a, nil
	}
	if a.isAgentsMDFocused() {
		if a.agentsMDPreferPath {
			a.agentsPathInput.Focus()
			a.mode = viewAgentsPathInput
		} else {
			a.agentsEditor.SetValue(a.agentsMD)
			a.agentsEditor.Focus()
			a.mode = viewAgentsEditor
		}
		return a, nil
	}
	return a, nil
}

func (a *App) handleCardFormSubmit() (tea.Model, tea.Cmd) {
	card := config.IDCard{
		ID:                 a.editingCardID,
		Name:               strings.TrimSpace(a.cardInputs[0].Value()),
		Driver:             strings.TrimSpace(a.cardInputs[1].Value()),
		Model:              strings.TrimSpace(a.cardInputs[2].Value()),
		WorkingDir:         strings.TrimSpace(a.cardInputs[3].Value()),
		FaceStyle:          strings.TrimSpace(a.cardInputs[4].Value()),
		UseTmpWorkspace:    strings.ToLower(strings.TrimSpace(a.cardInputs[5].Value())) == "y",
		AgentsMD:           a.agentsMD,
		AgentsMDPath:       a.agentsMDPath,
		AgentsMDPreferPath: a.agentsMDPreferPath,
		SkillIDs:           a.selectedSkillIDsSlice(),
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

func (a *App) resetCardForm() {
	for i := range a.cardInputs {
		a.cardInputs[i].SetValue("")
		a.cardInputs[i].Blur()
	}
	a.editingCardID = ""
	a.agentsMD = ""
	a.agentsMDPath = ""
	a.agentsMDPreferPath = false
	a.agentsPathInput.SetValue("")
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

func (a *App) isAgentsMDFocused() bool {
	return a.cardFocusIndex == len(a.cardInputs)
}

func (a *App) isSkillListFocused() bool {
	return a.cardFocusIndex == len(a.cardInputs)+1
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
	total := len(a.cardInputs) + cardInputTotalOffset
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
	start := max(0, a.skillCursor-height/2)
	end := start + height
	if end > total {
		end = total
		start = max(0, end-height)
	}
	return start, end
}

func (a *App) skillListHeight() int {
	if a.height <= 0 {
		return defaultSkillListHeight
	}
	available := a.height - skillListHeightOffset
	return max(minSkillListHeight, available)
}
