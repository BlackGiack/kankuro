package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (a *App) updateSkillSources(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return a, nil
	}

	switch keyMsg.String() {
	case "up":
		if a.skillSourceCursor > 0 {
			a.skillSourceCursor--
		}
	case "down":
		if a.skillSourceCursor < len(a.skillSources) { // +1 for "New"
			a.skillSourceCursor++
		}
	case "n":
		return a.handleSkillSourceNew()
	case " ", keyEnter:
		return a.handleSkillSourceSelect()
	case "d", "backspace":
		return a.handleSkillSourceDelete()
	case keyEsc:
		a.mode = viewCardForm
	case keyCtrlC:
		return a, tea.Quit
	}
	return a, nil
}

func (a *App) handleSkillSourceNew() (tea.Model, tea.Cmd) {
	a.editingSkillSourceIndex = -1
	a.skillSourceInput.SetValue("")
	a.skillSourceInput.Focus()
	a.mode = viewSkillSourceForm
	return a, nil
}

func (a *App) handleSkillSourceSelect() (tea.Model, tea.Cmd) {
	if a.skillSourceCursor == len(a.skillSources) {
		// "New" selected
		a.editingSkillSourceIndex = -1
		a.skillSourceInput.SetValue("")
	} else {
		// Edit existing
		a.editingSkillSourceIndex = a.skillSourceCursor
		a.skillSourceInput.SetValue(a.skillSources[a.skillSourceCursor])
	}
	a.skillSourceInput.Focus()
	a.mode = viewSkillSourceForm
	return a, nil
}

func (a *App) handleSkillSourceDelete() (tea.Model, tea.Cmd) {
	if a.skillSourceCursor < len(a.skillSources) {
		// Delete existing
		a.skillSources = append(a.skillSources[:a.skillSourceCursor], a.skillSources[a.skillSourceCursor+1:]...)
		if err := a.repo.SaveSkillSources(a.skillSources); err != nil {
			a.err = err
		} else {
			a.refreshSkills()
		}
		if a.skillSourceCursor >= len(a.skillSources) && a.skillSourceCursor > 0 {
			a.skillSourceCursor--
		}
	}
	return a, nil
}

func (a *App) updateSkillSourceForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	a.skillSourceInput, cmd = a.skillSourceInput.Update(msg)

	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return a, cmd
	}

	switch keyMsg.String() {
	case keyEnter:
		val := strings.TrimSpace(a.skillSourceInput.Value())
		if val != "" {
			a.saveSkillSource(val)
		}
		a.mode = viewSkillSources
	case keyEsc:
		a.mode = viewSkillSources
	case keyCtrlC:
		return a, tea.Quit
	}
	return a, cmd
}

func (a *App) saveSkillSource(val string) {
	if a.editingSkillSourceIndex == -1 {
		a.skillSources = append(a.skillSources, val)
	} else {
		a.skillSources[a.editingSkillSourceIndex] = val
	}
	if err := a.repo.SaveSkillSources(a.skillSources); err != nil {
		a.err = err
	} else {
		a.refreshSkills()
	}
}

func (a *App) renderSkillSourcesList() string {
	lines := make([]string, 0, len(a.skillSources)+1)
	for i, src := range a.skillSources {
		prefix := "  "
		if i == a.skillSourceCursor {
			prefix = "> "
		}
		line := fmt.Sprintf("%s%s", prefix, src)
		if i == a.skillSourceCursor {
			line = lipgloss.NewStyle().Foreground(highlightColor).Render(line)
		}
		lines = append(lines, line)
	}

	// "New" option
	prefix := "  "
	if a.skillSourceCursor == len(a.skillSources) {
		prefix = "> "
	}
	line := fmt.Sprintf("%s[ New Source ]", prefix)
	if a.skillSourceCursor == len(a.skillSources) {
		line = lipgloss.NewStyle().Foreground(highlightColor).Render(line)
	}
	lines = append(lines, line)

	return strings.Join(lines, "\n")
}

func (a *App) refreshSkills() {
	if err := a.repo.SyncSkillsFromSources(); err != nil {
		a.err = err
		return
	}
	skills, err := a.repo.ListSkills()
	if err != nil {
		a.err = err
		return
	}
	a.skills = skills
}
