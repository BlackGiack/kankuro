//nolint:dupl // similar to persona sources
package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"kankuro/internal/ui/components"
)

func (a *App) updateSkillSources(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return a, nil
	}

	switch keyMsg.String() {
	case keyUp:
		if a.skillSources.Cursor > 0 {
			a.skillSources.Cursor--
		}
	case keyDown:
		if a.skillSources.Cursor < len(a.skillSources.Sources) { // +1 for "New"
			a.skillSources.Cursor++
		}
	case "n":
		return a.handleSkillSourceNew()
	case " ", keyEnter:
		return a.handleSkillSourceSelect()
	case "d", keyBackspace:
		return a.handleSkillSourceDelete()
	case keyEsc:
		a.mode = viewCardForm
	case keyCtrlC:
		return a, tea.Quit
	}
	return a, nil
}

func (a *App) handleSkillSourceNew() (tea.Model, tea.Cmd) {
	a.skillSources.EditingIdx = -1
	a.skillSources.Input.SetValue("")
	a.skillSources.Input.Focus()
	a.mode = viewSkillSourceForm
	return a, nil
}

func (a *App) handleSkillSourceSelect() (tea.Model, tea.Cmd) {
	if a.skillSources.Cursor == len(a.skillSources.Sources) {
		// "New" selected
		a.skillSources.EditingIdx = -1
		a.skillSources.Input.SetValue("")
	} else {
		// Edit existing
		a.skillSources.EditingIdx = a.skillSources.Cursor
		a.skillSources.Input.SetValue(a.skillSources.Sources[a.skillSources.Cursor])
	}
	a.skillSources.Input.Focus()
	a.mode = viewSkillSourceForm
	return a, nil
}

func (a *App) handleSkillSourceDelete() (tea.Model, tea.Cmd) {
	if a.skillSources.Cursor < len(a.skillSources.Sources) {
		// Delete existing
		a.skillSources.Sources = append(a.skillSources.Sources[:a.skillSources.Cursor], a.skillSources.Sources[a.skillSources.Cursor+1:]...)
		if err := a.repo.SaveSkillSources(a.skillSources.Sources); err != nil {
			a.err = err
		} else {
			a.refreshSkills()
		}
		if a.skillSources.Cursor >= len(a.skillSources.Sources) && a.skillSources.Cursor > 0 {
			a.skillSources.Cursor--
		}
	}
	return a, nil
}

func (a *App) updateSkillSourceForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	a.skillSources.Input, cmd = a.skillSources.Input.Update(msg)

	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return a, cmd
	}

	switch keyMsg.String() {
	case keyEnter:
		val := strings.TrimSpace(a.skillSources.Input.Value())
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
	if a.skillSources.EditingIdx == -1 {
		a.skillSources.Sources = append(a.skillSources.Sources, val)
	} else {
		a.skillSources.Sources[a.skillSources.EditingIdx] = val
	}
	if err := a.repo.SaveSkillSources(a.skillSources.Sources); err != nil {
		a.err = err
	} else {
		a.refreshSkills()
	}
}

func (a *App) renderSkillSourcesList() string {
	lines := make([]string, 0, len(a.skillSources.Sources)+1)
	for i, src := range a.skillSources.Sources {
		prefix := "  "
		if i == a.skillSources.Cursor {
			prefix = "> "
		}
		line := fmt.Sprintf("%s%s", prefix, src)
		if i == a.skillSources.Cursor {
			line = lipgloss.NewStyle().Foreground(highlightColor).Render(line)
		}
		lines = append(lines, line)
	}

	// "New" option
	prefix := "  "
	if a.skillSources.Cursor == len(a.skillSources.Sources) {
		prefix = "> "
	}
	line := fmt.Sprintf("%s[ New Source ]", prefix)
	if a.skillSources.Cursor == len(a.skillSources.Sources) {
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

// ===== Integration with new SourceList component =====

func (a *App) newSkillSourceListComponent() *components.SourceList {
	config := components.SourceListConfig{
		Title:      "Manage Skill Sources",
		FormPrompt: "Path to skill source...",
		FormFields: 1,
	}

	callbacks := components.SourceListCallbacks{
		OnSave: func(item components.SourceItem, idx int) error {
			path := strings.TrimSpace(item.Display())
			if idx == -1 {
				// New
				a.skillSources.Sources = append(a.skillSources.Sources, path)
			} else {
				// Edit
				if idx >= 0 && idx < len(a.skillSources.Sources) {
					a.skillSources.Sources[idx] = path
				}
			}
			if err := a.repo.SaveSkillSources(a.skillSources.Sources); err != nil {
				a.err = err
				return err
			}
			a.refreshSkills()
			return nil
		},

		OnDelete: func(idx int) error {
			if idx >= 0 && idx < len(a.skillSources.Sources) {
				a.skillSources.Sources = append(
					a.skillSources.Sources[:idx],
					a.skillSources.Sources[idx+1:]...,
				)
			}
			if err := a.repo.SaveSkillSources(a.skillSources.Sources); err != nil {
				a.err = err
				return err
			}
			a.refreshSkills()
			return nil
		},

		OnRefresh: func() {
			a.refreshSkills()
		},

		OnClose: func() {
			a.mode = viewCardForm
		},
	}

	sl := components.NewSourceList(config, callbacks)
	sl.SetItems(components.ToSimpleItems(a.skillSources.Sources))
	return sl
}
