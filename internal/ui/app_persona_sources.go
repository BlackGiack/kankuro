//nolint:dupl // similar to skill sources
package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"kankuro/internal/ui/components"
)

func (a *App) updatePersonaSources(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return a, nil
	}

	switch keyMsg.String() {
	case keyUp:
		if a.personaSources.Cursor > 0 {
			a.personaSources.Cursor--
		}
	case keyDown:
		if a.personaSources.Cursor < len(a.personaSources.Sources) { // +1 for "New"
			a.personaSources.Cursor++
		}
	case "n":
		return a.handlePersonaSourceNew()
	case " ", keyEnter:
		return a.handlePersonaSourceSelect()
	case "d", keyBackspace:
		return a.handlePersonaSourceDelete()
	case keyEsc:
		a.mode = viewCardForm
	case keyCtrlC:
		return a, tea.Quit
	}
	return a, nil
}

func (a *App) handlePersonaSourceNew() (tea.Model, tea.Cmd) {
	a.personaSources.EditingIdx = -1
	a.personaSources.Input.SetValue("")
	a.personaSources.Input.Focus()
	a.mode = viewPersonaSourceForm
	return a, nil
}

func (a *App) handlePersonaSourceSelect() (tea.Model, tea.Cmd) {
	if a.personaSources.Cursor == len(a.personaSources.Sources) {
		// "New" selected
		a.personaSources.EditingIdx = -1
		a.personaSources.Input.SetValue("")
	} else {
		// Edit existing
		a.personaSources.EditingIdx = a.personaSources.Cursor
		a.personaSources.Input.SetValue(a.personaSources.Sources[a.personaSources.Cursor])
	}
	a.personaSources.Input.Focus()
	a.mode = viewPersonaSourceForm
	return a, nil
}

func (a *App) handlePersonaSourceDelete() (tea.Model, tea.Cmd) {
	if a.personaSources.Cursor < len(a.personaSources.Sources) {
		// Delete existing
		a.personaSources.Sources = append(
			a.personaSources.Sources[:a.personaSources.Cursor],
			a.personaSources.Sources[a.personaSources.Cursor+1:]...,
		)
		if err := a.repo.SavePersonaSources(a.personaSources.Sources); err != nil {
			a.err = err
		} else {
			a.refreshPersonas()
		}
		if a.personaSources.Cursor >= len(a.personaSources.Sources) && a.personaSources.Cursor > 0 {
			a.personaSources.Cursor--
		}
	}
	return a, nil
}

func (a *App) updatePersonaSourceForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	a.personaSources.Input, cmd = a.personaSources.Input.Update(msg)

	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return a, cmd
	}

	switch keyMsg.String() {
	case keyEnter:
		val := strings.TrimSpace(a.personaSources.Input.Value())
		if val != "" {
			a.savePersonaSource(val)
		}
		a.mode = viewPersonaSources
	case keyEsc:
		a.mode = viewPersonaSources
	case keyCtrlC:
		return a, tea.Quit
	}
	return a, cmd
}

func (a *App) savePersonaSource(val string) {
	if a.personaSources.EditingIdx == -1 {
		a.personaSources.Sources = append(a.personaSources.Sources, val)
	} else {
		a.personaSources.Sources[a.personaSources.EditingIdx] = val
	}
	if err := a.repo.SavePersonaSources(a.personaSources.Sources); err != nil {
		a.err = err
	} else {
		a.refreshPersonas()
	}
}

func (a *App) renderPersonaSourcesList() string {
	lines := make([]string, 0, len(a.personaSources.Sources)+1)
	for i, src := range a.personaSources.Sources {
		prefix := "  "
		if i == a.personaSources.Cursor {
			prefix = "> "
		}
		line := fmt.Sprintf("%s%s", prefix, src)
		if i == a.personaSources.Cursor {
			line = lipgloss.NewStyle().Foreground(highlightColor).Render(line)
		}
		lines = append(lines, line)
	}

	// "New" option
	prefix := "  "
	if a.personaSources.Cursor == len(a.personaSources.Sources) {
		prefix = "> "
	}
	line := fmt.Sprintf("%s[ New Source ]", prefix)
	if a.personaSources.Cursor == len(a.personaSources.Sources) {
		line = lipgloss.NewStyle().Foreground(highlightColor).Render(line)
	}
	lines = append(lines, line)

	return strings.Join(lines, "\n")
}

func (a *App) refreshPersonas() {
	if err := a.repo.SyncPersonasFromSources(); err != nil {
		a.err = err
		return
	}
	personas, err := a.repo.ListPersonas()
	if err != nil {
		a.err = err
		return
	}
	a.personas = personas
}

// ===== Integration with new SourceList component =====

func (a *App) newPersonaSourceListComponent() *components.SourceList {
	config := components.SourceListConfig{
		Title:      "Manage Persona Sources",
		FormPrompt: "Path to persona source...",
		FormFields: 1,
	}

	callbacks := components.SourceListCallbacks{
		OnSave: func(item components.SourceItem, idx int) error {
			path := strings.TrimSpace(item.Display())
			if idx == -1 {
				a.personaSources.Sources = append(a.personaSources.Sources, path)
			} else {
				if idx >= 0 && idx < len(a.personaSources.Sources) {
					a.personaSources.Sources[idx] = path
				}
			}
			if err := a.repo.SavePersonaSources(a.personaSources.Sources); err != nil {
				a.err = err
				return err
			}
			a.refreshPersonas()
			return nil
		},

		OnDelete: func(idx int) error {
			if idx >= 0 && idx < len(a.personaSources.Sources) {
				a.personaSources.Sources = append(
					a.personaSources.Sources[:idx],
					a.personaSources.Sources[idx+1:]...,
				)
			}
			if err := a.repo.SavePersonaSources(a.personaSources.Sources); err != nil {
				a.err = err
				return err
			}
			a.refreshPersonas()
			return nil
		},

		OnRefresh: func() {
			a.refreshPersonas()
		},

		OnClose: func() {
			a.mode = viewCardForm
		},
	}

	sl := components.NewSourceList(config, callbacks)
	sl.SetItems(components.ToSimpleItems(a.personaSources.Sources))
	return sl
}
