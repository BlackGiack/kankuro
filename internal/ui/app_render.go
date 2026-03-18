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
			return lipgloss.NewStyle().Foreground(hotPink).Bold(true).Render(
				prefix + "  " + contentOption + "\n" + prefix + "x " + pathOption,
			)
		}
		return lipgloss.NewStyle().Foreground(hotPink).Bold(true).Render(
			prefix + "x " + contentOption + "\n" + prefix + "  " + pathOption,
		)
	}

	if a.agentsMDPreferPath {
		return lipgloss.NewStyle().Foreground(gray).Render("    " + contentOption + "\n  x " + pathOption)
	}
	return lipgloss.NewStyle().Foreground(gray).Render("  x " + contentOption + "\n    " + pathOption)
}

func (a *App) renderSkillList() string {
	if len(a.skills) == 0 {
		text := "(no skills found)"
		if a.isSkillListFocused() {
			return lipgloss.NewStyle().Foreground(hotPink).Bold(true).Render("> " + text)
		}
		return text
	}
	var header string
	if a.skillSearchQuery != "" {
		header = lipgloss.NewStyle().Foreground(gray).Render("Search: "+a.skillSearchQuery) + "\n"
	}
	filtered := a.filteredSkills()
	if len(filtered) == 0 {
		return header + "(no matching skills)"
	}
	counts := skillNameCounts(a.skills)
	start, end := a.skillWindow(len(filtered))
	lines := make([]string, 0, end-start+skillWindowBuffer)
	if start > 0 {
		lines = append(lines, "  ...")
	}
	for i := start; i < end; i++ {
		skill := filtered[i]
		prefix := "  "
		if a.isSkillListFocused() && i == a.skillCursor {
			prefix = "> "
		}
		mark := " "
		if a.selectedSkillIDs[skill.ID] {
			mark = "x"
		}
		line := fmt.Sprintf("%s[%s] %s", prefix, mark, skillLabel(skill, counts))
		if a.isSkillListFocused() && i == a.skillCursor {
			line = lipgloss.NewStyle().Foreground(hotPink).Render(line)
		}
		lines = append(lines, line)
	}
	if end < len(filtered) {
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

func renderCardItem(card config.IDCard, skills []config.Skill) string {
	skillNames := make([]string, 0, len(card.SkillIDs))
	selected := make(map[string]struct{})
	for _, id := range card.SkillIDs {
		selected[id] = struct{}{}
	}
	for _, skill := range skills {
		if _, ok := selected[skill.ID]; ok {
			skillNames = append(skillNames, skill.Name)
		}
	}

	var skillsLine string
	switch {
	case len(skillNames) == 0:
		skillsLine = "(none)"
	case len(skillNames) > maxSkillsSummary:
		skillsLine = strings.Join(skillNames[:maxSkillsSummary], ", ") + "..."
	default:
		skillsLine = strings.Join(skillNames, ", ")
	}

	style := card.FaceStyle
	if style == "" {
		style = faceNeutral
	}
	face := characters.NewFace(ascii.WithStyle(style))
	w, _ := face.Bounds()
	canvas := ascii.NewCanvas(w, 1)
	face.Render(canvas, 0, 0)
	faceStr := canvas.String()

	infoLines := []string{
		truncate(card.Name, truncateLength),
		truncate(fmt.Sprintf("%s | %s", card.Driver, card.Model), truncateLength),
		truncate(skillsLine, truncateLength),
	}

	infoBox := lipgloss.NewStyle().
		Width(infoBoxWidth).
		Height(infoBoxHeight).
		Render(strings.Join(infoLines, "\n"))

	faceBox := lipgloss.NewStyle().
		Width(faceBoxWidth).
		Height(faceBoxHeight).
		Align(lipgloss.Center, lipgloss.Center).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(purple).
		Render(faceStr)

	content := lipgloss.JoinHorizontal(lipgloss.Top, infoBox, faceBox)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(purple).
		Width(cardItemWidth).
		Render(content)
}

func truncate(s string, w int) string {
	if len(s) > w {
		return s[:w-3] + "..."
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
