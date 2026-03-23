//nolint:dupl // similar to skill sources
package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (a *App) updatePersonaSources(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return a, nil
	}

	switch keyMsg.String() {
	case keyUp:
		if a.personaSourceCursor > 0 {
			a.personaSourceCursor--
		}
	case keyDown:
		if a.personaSourceCursor < len(a.personaSources) { // +1 for "New"
			a.personaSourceCursor++
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
	a.editingPersonaSourceIndex = -1
	a.personaSourceInput.SetValue("")
	a.personaSourceInput.Focus()
	a.mode = viewPersonaSourceForm
	return a, nil
}

func (a *App) handlePersonaSourceSelect() (tea.Model, tea.Cmd) {
	if a.personaSourceCursor == len(a.personaSources) {
		// "New" selected
		a.editingPersonaSourceIndex = -1
		a.personaSourceInput.SetValue("")
	} else {
		// Edit existing
		a.editingPersonaSourceIndex = a.personaSourceCursor
		a.personaSourceInput.SetValue(a.personaSources[a.personaSourceCursor])
	}
	a.personaSourceInput.Focus()
	a.mode = viewPersonaSourceForm
	return a, nil
}

func (a *App) handlePersonaSourceDelete() (tea.Model, tea.Cmd) {
	if a.personaSourceCursor < len(a.personaSources) {
		// Delete existing
		a.personaSources = append(
			a.personaSources[:a.personaSourceCursor],
			a.personaSources[a.personaSourceCursor+1:]...,
		)
		if err := a.repo.SavePersonaSources(a.personaSources); err != nil {
			a.err = err
		} else {
			a.refreshPersonas()
		}
		if a.personaSourceCursor >= len(a.personaSources) && a.personaSourceCursor > 0 {
			a.personaSourceCursor--
		}
	}
	return a, nil
}

func (a *App) updatePersonaSourceForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	a.personaSourceInput, cmd = a.personaSourceInput.Update(msg)

	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return a, cmd
	}

	switch keyMsg.String() {
	case keyEnter:
		val := strings.TrimSpace(a.personaSourceInput.Value())
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
	if a.editingPersonaSourceIndex == -1 {
		a.personaSources = append(a.personaSources, val)
	} else {
		a.personaSources[a.editingPersonaSourceIndex] = val
	}
	if err := a.repo.SavePersonaSources(a.personaSources); err != nil {
		a.err = err
	} else {
		a.refreshPersonas()
	}
}

func (a *App) renderPersonaSourcesList() string {
	lines := make([]string, 0, len(a.personaSources)+1)
	for i, src := range a.personaSources {
		prefix := "  "
		if i == a.personaSourceCursor {
			prefix = "> "
		}
		line := fmt.Sprintf("%s%s", prefix, src)
		if i == a.personaSourceCursor {
			line = lipgloss.NewStyle().Foreground(highlightColor).Render(line)
		}
		lines = append(lines, line)
	}

	// "New" option
	prefix := "  "
	if a.personaSourceCursor == len(a.personaSources) {
		prefix = "> "
	}
	line := fmt.Sprintf("%s[ New Source ]", prefix)
	if a.personaSourceCursor == len(a.personaSources) {
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
