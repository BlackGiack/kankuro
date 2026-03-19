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
	if len(a.skills) == 0 {
		text := "(no skills found)"
		if a.isSkillListFocused() {
			return lipgloss.NewStyle().Foreground(highlightColor).Bold(true).Render("> " + text)
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
			line = lipgloss.NewStyle().Foreground(highlightColor).Render(line)
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

func renderCardItem(card config.IDCard, skills []config.Skill, selected bool) string {
	skillNames := make([]string, 0, len(card.SkillIDs))
	selectedSkills := make(map[string]struct{})
	for _, id := range card.SkillIDs {
		selectedSkills[id] = struct{}{}
	}
	for _, skill := range skills {
		if _, ok := selectedSkills[skill.ID]; ok {
			skillNames = append(skillNames, skill.Name)
		}
	}

	pillBg := highlightColor
	if selected {
		pillBg = white
	}
	pillFg := white
	if selected {
		pillFg = primaryColor
	}

	var pills []string
	if len(skillNames) > 0 {
		for _, s := range skillNames {
			if len(s) > pillTextLimit {
				s = s[:pillTextLimit-1] + "."
			}
			pills = append(pills, lipgloss.NewStyle().
				Foreground(pillFg).
				Background(pillBg).
				Padding(0, 1).
				MarginRight(1).
				Render(s))
		}
	} else {
		pills = append(pills, lipgloss.NewStyle().Foreground(gray).Render("(no skills)"))
	}
	skillsLine := lipgloss.JoinHorizontal(lipgloss.Top, pills...)

	style := card.FaceStyle
	if style == "" {
		style = faceNeutral
	}
	face := characters.NewFace(ascii.WithStyle(style))
	w, h := face.Bounds()
	canvas := ascii.NewCanvas(faceBoxWidth, faceBoxHeight)
	face.Render(canvas, (faceBoxWidth-w)/2, (faceBoxHeight-h)/2) //nolint:mnd // Centering logic.
	faceStr := canvas.String()

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
		Background(faceBg).
		Foreground(white).
		Render(faceStr)

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
