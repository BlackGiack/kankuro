package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"kankuro/internal/config"
)

const printableCharMin = 33
const printableCharMax = 126

func (a *App) filteredSkills() []config.Skill {
	if a.skillSearchQuery == "" {
		return a.skills
	}
	query := strings.ToLower(a.skillSearchQuery)
	var filtered []config.Skill
	for _, skill := range a.skills {
		if strings.Contains(strings.ToLower(skill.Name), query) ||
			strings.Contains(strings.ToLower(skill.Description), query) {
			filtered = append(filtered, skill)
		}
	}
	return filtered
}

func (a *App) handleCardFormEscape() {
	if a.isSkillListFocused() && a.skillSearchQuery != "" {
		a.skillSearchQuery = ""
		a.skillCursor = 0
		return
	}
	a.mode = viewDashboard
}

func (a *App) handleSkillListSearchInput(keyMsg tea.KeyMsg) bool {
	if !a.isSkillListFocused() {
		return false
	}
	if keyMsg.String() == keyBackspace && len(a.skillSearchQuery) > 0 {
		a.skillSearchQuery = a.skillSearchQuery[:len(a.skillSearchQuery)-1]
		a.skillCursor = 0
		return true
	}
	if len(keyMsg.Runes) > 0 {
		ch := keyMsg.Runes[0]
		if ch >= printableCharMin && ch <= printableCharMax {
			a.skillSearchQuery += string(ch)
			a.skillCursor = 0
			return true
		}
	}
	return false
}

func (a *App) updateCardForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if a.cardFocusIndex < len(a.cardInputs) && !a.cardInputs[a.cardFocusIndex].Focused() {
		a.setCardFocus(a.cardFocusIndex)
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if cmd, handled := a.handleCardFormKeyMsg(keyMsg); handled {
			return a, cmd
		}
	}
	cmds := make([]tea.Cmd, len(a.cardInputs))
	for i := range a.cardInputs {
		a.cardInputs[i], cmds[i] = a.cardInputs[i].Update(msg)
	}
	return a, tea.Batch(cmds...)
}

func (a *App) handleCardFormKeyMsg(keyMsg tea.KeyMsg) (tea.Cmd, bool) {
	switch keyMsg.String() {
	case keyCtrlC:
		return tea.Quit, true
	case keyEsc:
		a.handleCardFormEscape()
		return nil, true
	case keyTab:
		a.skillSearchQuery = ""
		a.cycleCardFocus()
		return nil, true
	case "up", "k":
		a.handleCardFormNav(false)
		return nil, true
	case "down", "j":
		a.handleCardFormNav(true)
		return nil, true
	case " ":
		if a.isSkillListFocused() {
			a.handleCardFormToggle()
			return nil, true
		}
	case keyEnter:
		a.handleCardFormSubmit()
		return nil, true
	case "ctrl+e":
		if a.isSkillListFocused() {
			a.openSkillSources()
			return nil, true
		}
	}
	return nil, a.handleSkillListSearchInput(keyMsg)
}

func (a *App) openSkillSources() {
	sources, err := a.repo.ListSkillSources()
	if err != nil {
		a.err = err
		return
	}
	a.skillSources = sources
	a.skillSourceCursor = 0
	a.mode = viewSkillSources
}

func (a *App) handleCardFormNav(down bool) {
	if a.isSkillListFocused() {
		filtered := a.filteredSkills()
		if down {
			if a.skillCursor < len(filtered)-1 {
				a.skillCursor++
			}
		} else if a.skillCursor > 0 {
			a.skillCursor--
		}
	}
	if a.isAgentsMDFocused() {
		a.agentsMDPreferPath = down
	}
}

func (a *App) handleCardFormToggle() {
	if a.isSkillListFocused() {
		a.toggleSelectedSkill()
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
	}
}

func (a *App) handleCardFormSubmit() {
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
	}
	a.upsertCardLocal(card)
	a.refreshCardList()
	a.editingCardID = ""
	a.mode = viewDashboard
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
	a.skillSearchQuery = ""
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
	filtered := a.filteredSkills()
	if len(filtered) == 0 || a.skillCursor < 0 || a.skillCursor >= len(filtered) {
		return
	}
	id := filtered[a.skillCursor].ID
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
