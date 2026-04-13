package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"kankuro/internal/config"
)

const printableCharMin = 33
const printableCharMax = 126

func (a *App) filteredSkills() []config.Skill {
	return filterByQuery(a.skills, a.cardForm.SkillSearchQuery, func(s config.Skill, q string) bool {
		return strings.Contains(strings.ToLower(s.Name), q) ||
			strings.Contains(strings.ToLower(s.Description), q)
	})
}

func (a *App) filteredPersonas() []config.Persona {
	return filterByQuery(a.personas, a.cardForm.PersonaSearchQuery, func(p config.Persona, q string) bool {
		return strings.Contains(strings.ToLower(p.Name), q)
	})
}

func (a *App) filteredMCPs() []config.MCPServer {
	return filterByQuery(a.mcps, a.cardForm.MCPSearchQuery, func(m config.MCPServer, q string) bool {
		return strings.Contains(strings.ToLower(m.Name), q) ||
			strings.Contains(strings.ToLower(m.Description), q)
	})
}

func (a *App) handleCardFormEscape() {
	if a.isSkillListFocused() && a.cardForm.SkillSearchQuery != "" {
		a.cardForm.SkillSearchQuery = ""
		a.cardForm.SkillCursor = 0
		return
	}
	if a.isPersonaListFocused() && a.cardForm.PersonaSearchQuery != "" {
		a.cardForm.PersonaSearchQuery = ""
		a.cardForm.PersonaCursor = 0
		return
	}
	if a.isMCPListFocused() && a.cardForm.MCPSearchQuery != "" {
		a.cardForm.MCPSearchQuery = ""
		a.cardForm.MCPCursor = 0
		return
	}
	a.mode = viewDashboard
}

func (a *App) handleSkillListSearchInput(keyMsg tea.KeyMsg) bool {
	if !a.isSkillListFocused() {
		return false
	}
	if keyMsg.String() == keyBackspace && len(a.cardForm.SkillSearchQuery) > 0 {
		a.cardForm.SkillSearchQuery = a.cardForm.SkillSearchQuery[:len(a.cardForm.SkillSearchQuery)-1]
		a.cardForm.SkillCursor = 0
		return true
	}
	if len(keyMsg.Runes) > 0 {
		ch := keyMsg.Runes[0]
		if ch >= printableCharMin && ch <= printableCharMax {
			a.cardForm.SkillSearchQuery += string(ch)
			a.cardForm.SkillCursor = 0
			return true
		}
	}
	return false
}

func (a *App) handlePersonaListSearchInput(keyMsg tea.KeyMsg) bool {
	if !a.isPersonaListFocused() {
		return false
	}
	if keyMsg.String() == keyBackspace && len(a.cardForm.PersonaSearchQuery) > 0 {
		a.cardForm.PersonaSearchQuery = a.cardForm.PersonaSearchQuery[:len(a.cardForm.PersonaSearchQuery)-1]
		a.cardForm.PersonaCursor = 0
		return true
	}
	if len(keyMsg.Runes) > 0 {
		ch := keyMsg.Runes[0]
		if ch >= printableCharMin && ch <= printableCharMax {
			a.cardForm.PersonaSearchQuery += string(ch)
			a.cardForm.PersonaCursor = 0
			return true
		}
	}
	return false
}

func (a *App) handleMCPListSearchInput(keyMsg tea.KeyMsg) bool {
	if !a.isMCPListFocused() {
		return false
	}
	if keyMsg.String() == keyBackspace && len(a.cardForm.MCPSearchQuery) > 0 {
		a.cardForm.MCPSearchQuery = a.cardForm.MCPSearchQuery[:len(a.cardForm.MCPSearchQuery)-1]
		a.cardForm.MCPCursor = 0
		return true
	}
	if len(keyMsg.Runes) > 0 {
		ch := keyMsg.Runes[0]
		if ch >= printableCharMin && ch <= printableCharMax {
			a.cardForm.MCPSearchQuery += string(ch)
			a.cardForm.MCPCursor = 0
			return true
		}
	}
	return false
}

func (a *App) updateCardForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if a.cardForm.FocusIndex < len(a.cardForm.Inputs) && !a.cardForm.Inputs[a.cardForm.FocusIndex].Focused() {
		a.setCardFocus(a.cardForm.FocusIndex)
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if cmd, handled := a.handleCardFormKeyMsg(keyMsg); handled {
			return a, cmd
		}
	}
	cmds := make([]tea.Cmd, len(a.cardForm.Inputs))
	for i := range a.cardForm.Inputs {
		a.cardForm.Inputs[i], cmds[i] = a.cardForm.Inputs[i].Update(msg)
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
		a.cardForm.SkillSearchQuery = ""
		a.cardForm.PersonaSearchQuery = ""
		a.cardForm.MCPSearchQuery = ""
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
	if a.cardForm.FocusIndex == cardWorkspaceIndex {
		a.openWorkspacePicker()
		return true
	}
	if a.isPuppetFocused() {
		a.cardForm.PuppetSelectorPrev = a.cardForm.Puppet
		a.cardForm.PuppetSelectorIndex = 0
		for i, p := range puppetStyles() {
			if p == a.cardForm.Puppet {
				a.cardForm.PuppetSelectorIndex = i
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
		a.cardForm.MCPFormPrevMode = viewCardForm
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
	a.skillSources.Sources = sources
	a.skillSources.Cursor = 0
	a.mode = viewSkillSources
}

func (a *App) openPersonaSources() {
	sources, err := a.repo.ListPersonaSources()
	if err != nil {
		a.err = err
		return
	}
	a.personaSources.Sources = sources
	a.personaSources.Cursor = 0
	a.mode = viewPersonaSources
}

func (a *App) handleCardFormNav(down bool) {
	switch {
	case a.isSkillListFocused():
		a.cardForm.SkillCursor = a.calculateNavCursor(a.cardForm.SkillCursor, len(a.filteredSkills()), down)
	case a.isPersonaListFocused():
		a.cardForm.PersonaCursor = a.calculateNavCursor(a.cardForm.PersonaCursor, len(a.filteredPersonas()), down)
	case a.isMCPListFocused():
		a.cardForm.MCPCursor = a.calculateNavCursor(a.cardForm.MCPCursor, len(a.filteredMCPs()), down)
	case a.isAgentsMDFocused():
		a.cardForm.PreferPath = down
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
		if a.cardForm.PreferPath {
			a.cardForm.PathInput.Focus()
			a.mode = viewAgentsPathInput
		} else {
			a.cardForm.Editor.SetValue(a.cardForm.Content)
			a.cardForm.Editor.Focus()
			a.mode = viewAgentsEditor
		}
	}
}

func (a *App) handleCardFormSubmit() {
	card := config.IDCard{
		ID:              a.cardForm.EditingID,
		Name:            strings.TrimSpace(a.cardForm.Inputs[0].Value()),
		Driver:          strings.TrimSpace(a.cardForm.Inputs[1].Value()),
		Model:           strings.TrimSpace(a.cardForm.Inputs[2].Value()),
		WorkspaceID:     a.cardForm.SelectedWorkspaceID,
		FaceStyle:       strings.TrimSpace(a.cardForm.Inputs[3].Value()),
		Puppet:          a.cardForm.Puppet,
		UseTmpWorkspace: strings.ToLower(strings.TrimSpace(a.cardForm.Inputs[4].Value())) == "y",
		AgentsMD:        a.cardForm.Content,

		AgentsMDPath:       a.cardForm.Path,
		AgentsMDPreferPath: a.cardForm.PreferPath,
		SkillIDs:           a.selectedSkillIDsSlice(),
		PersonaIDs:         a.selectedPersonaIDsSlice(),
		MCPIDs:             a.selectedMCPIDsSlice(),
	}
	var err error
	if a.cardForm.EditingID == "" {
		card, err = a.repo.CreateIDCard(card)
	} else {
		card, err = a.repo.UpdateIDCard(card)
	}
	if err != nil {
		a.err = err
	}
	a.upsertCardLocal(card)
	a.refreshCardList()
	a.cardForm.EditingID = ""
	a.mode = viewDashboard
}

func (a *App) resetCardForm() {
	for i := range a.cardForm.Inputs {
		a.cardForm.Inputs[i].SetValue("")
		a.cardForm.Inputs[i].Blur()
	}
	a.cardForm.EditingID = ""
	a.cardForm.SelectedWorkspaceID = ""
	a.cardForm.WorkspaceCursor = 0
	a.cardForm.WorkspaceSearchQuery = ""
	a.cardForm.Content = ""
	a.cardForm.Path = ""
	a.cardForm.PreferPath = false
	a.cardForm.PathInput.SetValue("")
	a.cardForm.SelectedSkillIDs = make(map[string]bool)
	a.cardForm.SkillCursor = 0
	a.cardForm.SkillSearchQuery = ""
	// default puppet for new cards
	a.cardForm.Puppet = defaultPuppet
	a.cardForm.PuppetSelectorIndex = 1
	a.cardForm.SelectedPersonaIDs = make(map[string]bool)
	a.cardForm.PersonaCursor = 0
	a.cardForm.PersonaSearchQuery = ""
	a.cardForm.SelectedMCPIDs = make(map[string]bool)
	a.cardForm.MCPCursor = 0
	a.cardForm.MCPSearchQuery = ""
	a.cardForm.FocusIndex = 0
	a.setCardFocus(a.cardForm.FocusIndex)
}

func (a *App) setSelectedSkills(ids []string) {
	a.cardForm.SelectedSkillIDs = make(map[string]bool)
	for _, id := range ids {
		a.cardForm.SelectedSkillIDs[id] = true
	}
	a.cardForm.SkillCursor = 0
}

func (a *App) setSelectedPersonas(ids []string) {
	a.cardForm.SelectedPersonaIDs = make(map[string]bool)
	for _, id := range ids {
		a.cardForm.SelectedPersonaIDs[id] = true
	}
	a.cardForm.PersonaCursor = 0
}

func (a *App) setSelectedMCPs(ids []string) {
	a.cardForm.SelectedMCPIDs = make(map[string]bool)
	for _, id := range ids {
		a.cardForm.SelectedMCPIDs[id] = true
	}
	a.cardForm.MCPCursor = 0
}

func (a *App) selectedSkillIDsSlice() []string {
	ids := make([]string, 0, len(a.cardForm.SelectedSkillIDs))
	for _, skill := range a.skills {
		if a.cardForm.SelectedSkillIDs[skill.ID] {
			ids = append(ids, skill.ID)
		}
	}
	return ids
}

// Focus index mapping:
// 0..4 -> cardForm.Inputs[0..4]
// 5 -> puppet (non-input)
// 6 -> cardForm.Inputs[5] (use tmp workspace)
// 7 -> agents MD
// 8 -> skill list.
func (a *App) isPuppetFocused() bool {
	return a.cardForm.FocusIndex == cardPuppetIndex
}

func (a *App) selectedPersonaIDsSlice() []string {
	ids := make([]string, 0, len(a.cardForm.SelectedPersonaIDs))
	for _, persona := range a.personas {
		if a.cardForm.SelectedPersonaIDs[persona.ID] {
			ids = append(ids, persona.ID)
		}
	}
	return ids
}

func (a *App) selectedMCPIDsSlice() []string {
	ids := make([]string, 0, len(a.cardForm.SelectedMCPIDs))
	for _, mcp := range a.mcps {
		if a.cardForm.SelectedMCPIDs[mcp.ID] {
			ids = append(ids, mcp.ID)
		}
	}
	return ids
}

func (a *App) isAgentsMDFocused() bool {
	return a.cardForm.FocusIndex == cardAgentsMDIndex
}

func (a *App) isSkillListFocused() bool {
	return a.cardForm.FocusIndex == cardSkillListIndex
}

func (a *App) isPersonaListFocused() bool {
	return a.cardForm.FocusIndex == cardPersonaListIndex
}

func (a *App) isMCPListFocused() bool {
	return a.cardForm.FocusIndex == cardMCPListIndex
}

// setCardFocus sets focus for text inputs mapping logical focus index to input indices.
func (a *App) setCardFocus(index int) {
	a.cardForm.FocusIndex = index
	// determine which textinput should be focused (if any)
	var mapped int
	switch {
	case index < cardWorkspaceIndex: // 0, 1, 2
		mapped = index
	case index == cardWorkspaceIndex: // 3 (Picker)
		mapped = -1
	case index == cardFaceStyleIndex: // 4 (Face Style)
		mapped = 3
	case index == cardPuppetIndex: // 5 (Puppet)
		mapped = -1
	case index == cardUseTmpIndex: // 6
		mapped = 4
	default:
		mapped = -1
	}
	for i := range a.cardForm.Inputs {
		if i == mapped {
			a.cardForm.Inputs[i].Focus()
		} else {
			a.cardForm.Inputs[i].Blur()
		}
	}
}

func (a *App) cycleCardFocus() {
	total := cardMCPListIndex + 1
	next := (a.cardForm.FocusIndex + 1) % total
	a.setCardFocus(next)
}

func (a *App) toggleSelectedSkill() {
	filtered := a.filteredSkills()
	if len(filtered) == 0 || a.cardForm.SkillCursor < 0 || a.cardForm.SkillCursor >= len(filtered) {
		return
	}
	id := filtered[a.cardForm.SkillCursor].ID
	if a.cardForm.SelectedSkillIDs[id] {
		delete(a.cardForm.SelectedSkillIDs, id)
		return
	}
	a.cardForm.SelectedSkillIDs[id] = true
}

// updatePuppetSelector handles key events while in the puppet-selection subview.
func (a *App) updatePuppetSelector(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "left":
			if a.cardForm.PuppetSelectorIndex > 0 {
				a.cardForm.PuppetSelectorIndex--
			} else {
				a.cardForm.PuppetSelectorIndex = len(puppetStyles()) - 1
			}
			return a, nil
		case "right":
			if a.cardForm.PuppetSelectorIndex < len(puppetStyles())-1 {
				a.cardForm.PuppetSelectorIndex++
			} else {
				a.cardForm.PuppetSelectorIndex = 0
			}
			return a, nil
		case keyEnter:
			// commit selection
			a.cardForm.Puppet = puppetStyles()[a.cardForm.PuppetSelectorIndex]
			a.mode = viewCardForm
			return a, nil
		case keyEsc:
			// cancel
			a.cardForm.Puppet = a.cardForm.PuppetSelectorPrev
			a.mode = viewCardForm
			return a, nil
		}
	}
	return a, nil
}

func (a *App) toggleSelectedPersona() {
	filtered := a.filteredPersonas()
	if len(filtered) == 0 || a.cardForm.PersonaCursor < 0 || a.cardForm.PersonaCursor >= len(filtered) {
		return
	}
	id := filtered[a.cardForm.PersonaCursor].ID
	if a.cardForm.SelectedPersonaIDs[id] {
		delete(a.cardForm.SelectedPersonaIDs, id)
		return
	}
	a.cardForm.SelectedPersonaIDs[id] = true
}

func (a *App) toggleSelectedMCP() {
	filtered := a.filteredMCPs()
	if len(filtered) == 0 || a.cardForm.MCPCursor < 0 || a.cardForm.MCPCursor >= len(filtered) {
		return
	}
	id := filtered[a.cardForm.MCPCursor].ID
	if a.cardForm.SelectedMCPIDs[id] {
		delete(a.cardForm.SelectedMCPIDs, id)
		return
	}
	a.cardForm.SelectedMCPIDs[id] = true
	// Auto-select mcp-cli skill if not already selected
	a.autoSelectMCPCLISkill()
}

// autoSelectMCPCLISkill ensures the mcp-cli built-in skill is selected when an MCP is added.
func (a *App) autoSelectMCPCLISkill() {
	for _, s := range a.skills {
		if s.Name == "mcp-cli" {
			if !a.cardForm.SelectedSkillIDs[s.ID] {
				a.cardForm.SelectedSkillIDs[s.ID] = true
			}
			return
		}
	}
}

func (a *App) deleteSelectedMCP() {
	filtered := a.filteredMCPs()
	if len(filtered) == 0 || a.cardForm.MCPCursor < 0 || a.cardForm.MCPCursor >= len(filtered) {
		return
	}
	id := filtered[a.cardForm.MCPCursor].ID
	if err := a.repo.DeleteMCP(id); err != nil {
		a.err = err
		return
	}
	// Refresh MCP list from repo
	a.mcps, _ = a.repo.ListMCPs()
	if a.cardForm.MCPCursor >= len(a.filteredMCPs()) {
		a.cardForm.MCPCursor = max(0, len(a.filteredMCPs())-1)
	}
}

func (a *App) skillWindow(total int) (int, int) {
	return calculateWindow(a.cardForm.SkillCursor, total, a.skillListHeight())
}

func (a *App) personaWindow(total int) (int, int) {
	return calculateWindow(a.cardForm.PersonaCursor, total, a.personaListHeight())
}

func (a *App) mcpWindow(total int) (int, int) {
	return calculateWindow(a.cardForm.MCPCursor, total, a.mcpListHeight())
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
