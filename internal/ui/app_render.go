package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"kankuro/internal/ascii"
	"kankuro/internal/ascii/characters"
	"kankuro/internal/config"
)

func (a *App) renderAgentsMDRow() string {
	contentSummary := "empty"
	if a.agentsMD != "" {
		lines := strings.Split(a.agentsMD, `\n`)
		contentSummary = lines[0]
		if len(lines) > 1 || len(contentSummary) > contentSummaryLimit {
			if len(contentSummary) > contentSummaryLimit {
				contentSummary = contentSummary[:contentSummaryKeep] + "..."
			} else {
				contentSummary += "..."
			}
		}
	}

	pathSummary := "empty"
	if a.agentsMDPath != "" {
		pathSummary = a.agentsMDPath
		if len(pathSummary) > contentSummaryLimit {
			pathSummary = "..." + pathSummary[len(pathSummary)-contentSummaryKeep:]
		}
	}

	contentOption := "[ Content: " + contentSummary + " ]"
	pathOption := "[ Path: " + pathSummary + " ]"

	if a.isAgentsMDFocused() {
		prefix := "> "
		if a.agentsMDPreferPath {
			return lipgloss.NewStyle().Foreground(highlightColor).Bold(true).Render(
				prefix + "  " + contentOption + "\n" + prefix + "x " + pathOption,
			)
		}
		return lipgloss.NewStyle().Foreground(highlightColor).Bold(true).Render(
			prefix + "x " + contentOption + "\n" + prefix + "  " + pathOption,
		)
	}

	if a.agentsMDPreferPath {
		return lipgloss.NewStyle().Foreground(gray).Render("    " + contentOption + "\n  x " + pathOption)
	}
	return lipgloss.NewStyle().Foreground(gray).Render("  x " + contentOption + "\n    " + pathOption)
}

func (a *App) renderSkillList() string {
	return renderRegistryList("Skills (ctrl+e=sources)", a.isSkillListFocused(), a.skillSearchQuery,
		func() []config.Skill { return a.filteredSkills() },
		func(total int) (int, int) { return a.skillWindow(total) },
		a.skillCursor,
		func(id string) bool { return a.selectedSkillIDs[id] },
		func(item config.Skill) string { return skillLabel(item, skillNameCounts(a.skills)) },
		func(item config.Skill) string { return item.ID },
	)
}

func (a *App) renderPersonaList() string {
	return renderRegistryList("Personas (ctrl+e=sources)", a.isPersonaListFocused(), a.personaSearchQuery,
		func() []config.Persona { return a.filteredPersonas() },
		func(total int) (int, int) { return a.personaWindow(total) },
		a.personaCursor,
		func(id string) bool { return a.selectedPersonaIDs[id] },
		func(item config.Persona) string { return item.Name },
		func(item config.Persona) string { return item.ID },
	)
}

func (a *App) renderMCPList() string {
	return renderRegistryList("MCPs (ctrl+e=add)", a.isMCPListFocused(), a.mcpSearchQuery,
		func() []config.MCPServer { return a.filteredMCPs() },
		func(total int) (int, int) { return a.mcpWindow(total) },
		a.mcpCursor,
		func(id string) bool { return a.selectedMCPIDs[id] },
		func(item config.MCPServer) string { return item.Name },
		func(item config.MCPServer) string { return item.ID },
	)
}

func renderRegistryList[T any](
	title string,
	focused bool,
	query string,
	filter func() []T,
	window func(int) (int, int),
	cursor int,
	selected func(string) bool,
	label func(T) string,
	getID func(T) string,
) string {
	items := filter()
	header := lipgloss.NewStyle().Foreground(gray).Render(title)
	if query != "" {
		header += lipgloss.NewStyle().Foreground(gray).Render(" | Search: " + query)
	}
	header += "\n"

	if len(items) == 0 {
		text := "(no matching items)"
		if focused {
			return header + lipgloss.NewStyle().Foreground(highlightColor).Bold(true).Render("> "+text)
		}
		return header + text
	}

	start, end := window(len(items))
	lines := make([]string, 0, end-start+skillWindowBuffer)
	if start > 0 {
		lines = append(lines, "  ...")
	}
	for i := start; i < end; i++ {
		item := items[i]
		prefix := "  "
		if focused && i == cursor {
			prefix = "> "
		}
		mark := " "
		if selected(getID(item)) {
			mark = "x"
		}
		line := fmt.Sprintf("%s[%s] %s", prefix, mark, label(item))
		if focused && i == cursor {
			line = lipgloss.NewStyle().Foreground(highlightColor).Render(line)
		}
		lines = append(lines, line)
	}
	if end < len(items) {
		lines = append(lines, "  ...")
	}
	return header + strings.Join(lines, "\n")
}

func skillNameCounts(skills []config.Skill) map[string]int {
	counts := make(map[string]int)
	for _, skill := range skills {
		counts[skill.Name]++
	}
	return counts
}

func skillLabel(skill config.Skill, counts map[string]int) string {
	if counts[skill.Name] > 1 && skill.Source != "" {
		return fmt.Sprintf("%s (%s)", skill.Name, skill.Source)
	}
	return skill.Name
}

func renderCardItem(card config.IDCard, skills []config.Skill, personas []config.Persona, selected bool) string {
	skillNames := getSelectedNames(card.SkillIDs, func(id string) (string, bool) {
		for _, s := range skills {
			if s.ID == id {
				return s.Name, true
			}
		}
		return "", false
	})

	personaNames := getSelectedNames(card.PersonaIDs, func(id string) (string, bool) {
		for _, p := range personas {
			if p.ID == id {
				return p.Name, true
			}
		}
		return "", false
	})

	pills := renderPills(skillNames, personaNames)
	skillsLine := lipgloss.JoinHorizontal(lipgloss.Top, pills...)

	style := card.Puppet
	if style == "" {
		style = defaultPuppet
	}
	puppet := characters.NewPuppet(func(c *ascii.Config) { c.Set("style", style) })
	w, h := puppet.Bounds()
	canvas := ascii.NewCanvas(w, h)
	puppet.Render(canvas, 0, 0)
	puppetStr := canvas.String()

	driverLine := fmt.Sprintf("%s | %s", card.Driver, card.Model)

	titleStyleLocal := titleStyle
	descStyleLocal := lipgloss.NewStyle().Foreground(gray)

	if selected {
		titleStyleLocal = titleStyleLocal.Foreground(highlightColor)
	}

	infoLines := []string{
		titleStyleLocal.Render(truncate(card.Name, truncateLength)),
		descStyleLocal.Render(truncate(driverLine, truncateLength)),
		"",
		skillsLine,
	}

	infoBox := lipgloss.NewStyle().
		Width(infoBoxWidth).
		Height(infoBoxHeight).
		Render(strings.Join(infoLines, "\n"))

	faceBg := deepCharcoal
	if selected {
		faceBg = primaryColor
	}

	faceBox := faceBoxStyle.
		Width(faceBoxWidth).
		Height(faceBoxHeight).
		Align(lipgloss.Center, lipgloss.Center).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		Background(faceBg).
		Foreground(white).
		Render(puppetStr)

	content := lipgloss.JoinHorizontal(lipgloss.Top, infoBox, faceBox)

	borderColor := normalBorderColor
	if selected {
		borderColor = primaryColor
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(cardItemWidth).
		Render(content)
}

func truncate(s string, w int) string {
	if len(s) > w {
		return s[:w-3] + "..."
	}
	return s
}

func getSelectedNames(ids []string, resolver func(string) (string, bool)) []string {
	var names []string
	for _, id := range ids {
		if name, ok := resolver(id); ok {
			names = append(names, name)
		}
	}
	return names
}

func renderPills(skillNames, personaNames []string) []string {
	var pills []string
	if len(skillNames) == 0 && len(personaNames) == 0 {
		return []string{lipgloss.NewStyle().Foreground(gray).Render("(no capabilities)")}
	}

	for _, s := range skillNames {
		if len(pills) >= maxPillsShown {
			break
		}
		pills = append(pills, pillStyle(highlightColor).Render(truncatePill(s)))
	}
	for _, s := range personaNames {
		if len(pills) >= maxPillsShown {
			break
		}
		pills = append(pills, pillStyle(personaPillColor).Render(truncatePill(s)))
	}
	return pills
}

func pillStyle(bg lipgloss.Color) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(white).
		Background(bg).
		Padding(0, 1).
		MarginRight(1)
}

func truncatePill(s string) string {
	if len(s) > pillTextLimit {
		return s[:pillTextLimit-1] + "."
	}
	return s
}

func renderAgentCard(card config.IDCard) string {
	style := card.FaceStyle
	if style == "" {
		style = faceNeutral
	}
	face := characters.NewFace(ascii.WithStyle(style))
	w, _ := face.Bounds()
	canvas := ascii.NewCanvas(w, 1)
	face.Render(canvas, 0, 0)
	return canvas.String()
}
