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

func (a *App) filteredPersonas() []config.Persona {
	if a.personaSearchQuery == "" {
		return a.personas
	}
	query := strings.ToLower(a.personaSearchQuery)
	var filtered []config.Persona
	for _, p := range a.personas {
		if strings.Contains(strings.ToLower(p.Name), query) {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

func (a *App) filteredMCPs() []config.MCPServer {
	if a.mcpSearchQuery == "" {
		return a.mcps
	}
	query := strings.ToLower(a.mcpSearchQuery)
	var filtered []config.MCPServer
	for _, mcp := range a.mcps {
		if strings.Contains(strings.ToLower(mcp.Name), query) ||
			strings.Contains(strings.ToLower(mcp.Description), query) {
			filtered = append(filtered, mcp)
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
	if a.isPersonaListFocused() && a.personaSearchQuery != "" {
		a.personaSearchQuery = ""
		a.personaCursor = 0
		return
	}
	if a.isMCPListFocused() && a.mcpSearchQuery != "" {
		a.mcpSearchQuery = ""
		a.mcpCursor = 0
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

func (a *App) handlePersonaListSearchInput(keyMsg tea.KeyMsg) bool {
	if !a.isPersonaListFocused() {
		return false
	}
	if keyMsg.String() == keyBackspace && len(a.personaSearchQuery) > 0 {
		a.personaSearchQuery = a.personaSearchQuery[:len(a.personaSearchQuery)-1]
		a.personaCursor = 0
		return true
	}
	if len(keyMsg.Runes) > 0 {
		ch := keyMsg.Runes[0]
		if ch >= printableCharMin && ch <= printableCharMax {
			a.personaSearchQuery += string(ch)
			a.personaCursor = 0
			return true
		}
	}
	return false
}

func (a *App) handleMCPListSearchInput(keyMsg tea.KeyMsg) bool {
	if !a.isMCPListFocused() {
		return false
	}
	if keyMsg.String() == keyBackspace && len(a.mcpSearchQuery) > 0 {
		a.mcpSearchQuery = a.mcpSearchQuery[:len(a.mcpSearchQuery)-1]
		a.mcpCursor = 0
		return true
	}
	if len(keyMsg.Runes) > 0 {
		ch := keyMsg.Runes[0]
		if ch >= printableCharMin && ch <= printableCharMax {
			a.mcpSearchQuery += string(ch)
			a.mcpCursor = 0
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
		a.personaSearchQuery = ""
		a.mcpSearchQuery = ""
		a.cycleCardFocus()
		return nil, true
	case "up":
		a.handleCardFormNav(false)
		return nil, true
	case "down":
		a.handleCardFormNav(true)
		return nil, true
	case " ":
		return nil, a.handleCardFormSpace()
	case keyEnter:
		a.handleCardFormSubmit()
		return nil, true
	case "delete", "del":
		if a.isMCPListFocused() {
			a.deleteSelectedMCP()
			return nil, true
		}
	case "ctrl+e":
		return nil, a.handleCardFormCtrlE()
	}
	return nil, a.handleCardFormSearch(keyMsg)
}

func (a *App) handleCardFormSpace() bool {
	if a.isPuppetFocused() {
		a.puppetSelectorPrev = a.cardPuppet
		a.puppetSelectorIndex = 0
		for i, p := range puppetStyles() {
			if p == a.cardPuppet {
				a.puppetSelectorIndex = i
				break
			}
		}
		a.mode = viewPuppetSelector
		return true
	}
	if a.isSkillListFocused() || a.isPersonaListFocused() || a.isAgentsMDFocused() || a.isMCPListFocused() {
		a.handleCardFormToggle()
		return true
	}
	return false
}

func (a *App) handleCardFormCtrlE() bool {
	if a.isSkillListFocused() {
		a.openSkillSources()
		return true
	}
	if a.isPersonaListFocused() {
		a.openPersonaSources()
		return true
	}
	if a.isMCPListFocused() {
		a.mcpFormPrevMode = viewCardForm
		a.mode = viewMCPForm
		a.mcpInputs[0].Focus()
		return true
	}
	return false
}

func (a *App) handleCardFormSearch(keyMsg tea.KeyMsg) bool {
	if a.isSkillListFocused() {
		return a.handleSkillListSearchInput(keyMsg)
	}
	if a.isPersonaListFocused() {
		return a.handlePersonaListSearchInput(keyMsg)
	}
	if a.isMCPListFocused() {
		return a.handleMCPListSearchInput(keyMsg)
	}
	return false
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

func (a *App) openPersonaSources() {
	sources, err := a.repo.ListPersonaSources()
	if err != nil {
		a.err = err
		return
	}
	a.personaSources = sources
	a.personaSourceCursor = 0
	a.mode = viewPersonaSources
}

func (a *App) handleCardFormNav(down bool) {
	switch {
	case a.isSkillListFocused():
		a.skillCursor = a.calculateNavCursor(a.skillCursor, len(a.filteredSkills()), down)
	case a.isPersonaListFocused():
		a.personaCursor = a.calculateNavCursor(a.personaCursor, len(a.filteredPersonas()), down)
	case a.isMCPListFocused():
		a.mcpCursor = a.calculateNavCursor(a.mcpCursor, len(a.filteredMCPs()), down)
	case a.isAgentsMDFocused():
		a.agentsMDPreferPath = down
	}
}

func (a *App) calculateNavCursor(current, total int, down bool) int {
	if total == 0 {
		return 0
	}
	if down {
		if current < total-1 {
			return current + 1
		}
	} else {
		if current > 0 {
			return current - 1
		}
	}
	return current
}

func (a *App) handleCardFormToggle() {
	if a.isSkillListFocused() {
		a.toggleSelectedSkill()
	}
	if a.isPersonaListFocused() {
		a.toggleSelectedPersona()
	}
	if a.isMCPListFocused() {
		a.toggleSelectedMCP()
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
		Puppet:             a.cardPuppet,
		UseTmpWorkspace:    strings.ToLower(strings.TrimSpace(a.cardInputs[5].Value())) == "y",
		AgentsMD:           a.agentsMD,
		AgentsMDPath:       a.agentsMDPath,
		AgentsMDPreferPath: a.agentsMDPreferPath,
		SkillIDs:           a.selectedSkillIDsSlice(),
		PersonaIDs:         a.selectedPersonaIDsSlice(),
		MCPIDs:             a.selectedMCPIDsSlice(),
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
	// default puppet for new cards
	a.cardPuppet = defaultPuppet
	a.puppetSelectorIndex = 1
	a.selectedPersonaIDs = make(map[string]bool)
	a.personaCursor = 0
	a.personaSearchQuery = ""
	a.selectedMCPIDs = make(map[string]bool)
	a.mcpCursor = 0
	a.mcpSearchQuery = ""
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

func (a *App) setSelectedPersonas(ids []string) {
	a.selectedPersonaIDs = make(map[string]bool)
	for _, id := range ids {
		a.selectedPersonaIDs[id] = true
	}
	a.personaCursor = 0
}

func (a *App) setSelectedMCPs(ids []string) {
	a.selectedMCPIDs = make(map[string]bool)
	for _, id := range ids {
		a.selectedMCPIDs[id] = true
	}
	a.mcpCursor = 0
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

// Focus index mapping:
// 0..4 -> cardInputs[0..4]
// 5 -> puppet (non-input)
// 6 -> cardInputs[5] (use tmp workspace)
// 7 -> agents MD
// 8 -> skill list.
func (a *App) isPuppetFocused() bool {
	return a.cardFocusIndex == cardPuppetIndex
}

func (a *App) selectedPersonaIDsSlice() []string {
	ids := make([]string, 0, len(a.selectedPersonaIDs))
	for _, persona := range a.personas {
		if a.selectedPersonaIDs[persona.ID] {
			ids = append(ids, persona.ID)
		}
	}
	return ids
}

func (a *App) selectedMCPIDsSlice() []string {
	ids := make([]string, 0, len(a.selectedMCPIDs))
	for _, mcp := range a.mcps {
		if a.selectedMCPIDs[mcp.ID] {
			ids = append(ids, mcp.ID)
		}
	}
	return ids
}

func (a *App) isAgentsMDFocused() bool {
	return a.cardFocusIndex == cardAgentsMDIndex
}

func (a *App) isSkillListFocused() bool {
	return a.cardFocusIndex == cardSkillListIndex
}

func (a *App) isPersonaListFocused() bool {
	return a.cardFocusIndex == cardPersonaListIndex
}

func (a *App) isMCPListFocused() bool {
	return a.cardFocusIndex == cardMCPListIndex
}

// setCardFocus sets focus for text inputs mapping logical focus index to input indices.
func (a *App) setCardFocus(index int) {
	a.cardFocusIndex = index
	// determine which textinput should be focused (if any)
	var mapped int
	switch {
	case index < cardPuppetIndex:
		mapped = index
	case index == cardUseTmpIndex:
		mapped = 5
	default:
		mapped = -1
	}
	for i := range a.cardInputs {
		if i == mapped {
			a.cardInputs[i].Focus()
		} else {
			a.cardInputs[i].Blur()
		}
	}
}

func (a *App) cycleCardFocus() {
	total := len(a.cardInputs) + cardInputTotalOffset + 1
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

// updatePuppetSelector handles key events while in the puppet-selection subview.
func (a *App) updatePuppetSelector(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "left":
			if a.puppetSelectorIndex > 0 {
				a.puppetSelectorIndex--
			} else {
				a.puppetSelectorIndex = len(puppetStyles()) - 1
			}
			return a, nil
		case "right":
			if a.puppetSelectorIndex < len(puppetStyles())-1 {
				a.puppetSelectorIndex++
			} else {
				a.puppetSelectorIndex = 0
			}
			return a, nil
		case keyEnter:
			// commit selection
			a.cardPuppet = puppetStyles()[a.puppetSelectorIndex]
			a.mode = viewCardForm
			return a, nil
		case keyEsc:
			// cancel
			a.cardPuppet = a.puppetSelectorPrev
			a.mode = viewCardForm
			return a, nil
		}
	}
	return a, nil
}

func (a *App) toggleSelectedPersona() {
	filtered := a.filteredPersonas()
	if len(filtered) == 0 || a.personaCursor < 0 || a.personaCursor >= len(filtered) {
		return
	}
	id := filtered[a.personaCursor].ID
	if a.selectedPersonaIDs[id] {
		delete(a.selectedPersonaIDs, id)
		return
	}
	a.selectedPersonaIDs[id] = true
}

func (a *App) toggleSelectedMCP() {
	filtered := a.filteredMCPs()
	if len(filtered) == 0 || a.mcpCursor < 0 || a.mcpCursor >= len(filtered) {
		return
	}
	id := filtered[a.mcpCursor].ID
	if a.selectedMCPIDs[id] {
		delete(a.selectedMCPIDs, id)
		return
	}
	a.selectedMCPIDs[id] = true
}

func (a *App) deleteSelectedMCP() {
	filtered := a.filteredMCPs()
	if len(filtered) == 0 || a.mcpCursor < 0 || a.mcpCursor >= len(filtered) {
		return
	}
	id := filtered[a.mcpCursor].ID
	if err := a.repo.DeleteMCP(id); err != nil {
		a.err = err
		return
	}
	// Refresh MCP list from repo
	a.mcps, _ = a.repo.ListMCPs()
	if a.mcpCursor >= len(a.filteredMCPs()) {
		a.mcpCursor = max(0, len(a.filteredMCPs())-1)
	}
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

func (a *App) personaWindow(total int) (int, int) {
	if total == 0 {
		return 0, 0
	}
	height := a.personaListHeight()
	if height >= total {
		return 0, total
	}
	start := max(0, a.personaCursor-height/2)
	end := start + height
	if end > total {
		end = total
		start = max(0, end-height)
	}
	return start, end
}

func (a *App) mcpWindow(total int) (int, int) {
	if total == 0 {
		return 0, 0
	}
	height := a.mcpListHeight()
	if height >= total {
		return 0, total
	}
	start := max(0, a.mcpCursor-height/2)
	end := start + height
	if end > total {
		end = total
		start = max(0, end-height)
	}
	return start, end
}

func (a *App) skillListHeight() int {
	if a.height <= 0 {
		return defaultSkillListHeight / skillListHeightDivisor
	}
	available := a.height - skillListHeightOffset
	return max(minSkillListHeight, available/skillListHeightDivisor)
}

func (a *App) personaListHeight() int {
	return a.skillListHeight()
}

func (a *App) mcpListHeight() int {
	return a.skillListHeight()
}
