package views_test

import (
	"fmt"
	"testing"

	"charm.land/lipgloss/v2"

	"kankuro/internal/config"
	"kankuro/internal/ui/components"
	"kankuro/internal/ui/views"
)

type testCapItem struct {
	id    string
	label string
}

func (i testCapItem) GetID() string          { return i.id }
func (i testCapItem) GetLabel() string       { return i.label }
func (i testCapItem) GetDescription() string { return "description" }

func TestCardEditor_SetSize_Idempotency(t *testing.T) {
	// Setup
	editor := views.NewCardEditor(views.CardEditorCallbacks{})

	settings := config.DefaultAppSettings()
	settings.ShowPersonasInCardEditor = true
	settings.ShowMCPsInCardEditor = true
	editor.SetSettings(settings)

	// Add many items to lists to trigger overflow if not bounded
	items := make([]components.CapabilityItem, 20)
	for i := range 20 {
		items[i] = testCapItem{id: fmt.Sprintf("item-%d", i), label: fmt.Sprintf("Item %d", i)}
	}
	editor.SkillList().SetItems(items)
	editor.PersonaList().SetItems(items)
	editor.MCPList().SetItems(items)

	// First SetSize
	width, height := 100, 40
	editor.SetSize(width, height)

	skillHeight1 := lipgloss.Height(editor.SkillList().View())
	personaHeight1 := lipgloss.Height(editor.PersonaList().View())
	mcpHeight1 := lipgloss.Height(editor.MCPList().View())

	// Assert they are reasonably bounded.
	// Expected height per list is roughly (height - header - help - separators) / 3
	// Header is ~15 lines. help=1, separators=2*3=6. (40-15-1-6)/3 = 6.
	// We allow a small margin for list header and ellipsis.
	expectedMax := height / 3
	if skillHeight1 > expectedMax {
		t.Errorf("Skill list height should be bounded by %d, got %d", expectedMax, skillHeight1)
	}
	if skillHeight1 <= 0 {
		t.Errorf("Skill list height should be positive, got %d", skillHeight1)
	}

	// Second SetSize with same dimensions
	editor.SetSize(width, height)

	skillHeight2 := lipgloss.Height(editor.SkillList().View())
	personaHeight2 := lipgloss.Height(editor.PersonaList().View())
	mcpHeight2 := lipgloss.Height(editor.MCPList().View())

	// Assert idempotency
	if skillHeight1 != skillHeight2 {
		t.Errorf("Skill list height changed on second SetSize: %d -> %d", skillHeight1, skillHeight2)
	}
	if personaHeight1 != personaHeight2 {
		t.Errorf("Persona list height changed on second SetSize: %d -> %d", personaHeight1, personaHeight2)
	}
	if mcpHeight1 != mcpHeight2 {
		t.Errorf("MCP list height changed on second SetSize: %d -> %d", mcpHeight1, mcpHeight2)
	}

	// Verify total form height doesn't overflow terminal height
	formHeight := lipgloss.Height(editor.View())
	if formHeight > height {
		t.Errorf("Editor View overflows terminal height: %d > %d", formHeight, height)
	}
}

func TestCardEditor_SetSize_SmallTerminal(t *testing.T) {
	// Setup
	editor := views.NewCardEditor(views.CardEditorCallbacks{})

	settings := config.DefaultAppSettings()
	settings.ShowPersonasInCardEditor = true
	settings.ShowMCPsInCardEditor = true
	editor.SetSettings(settings)

	// Add many items to all 3 lists to trigger minCapListHeight floor
	items := make([]components.CapabilityItem, 10)
	for i := range 10 {
		items[i] = testCapItem{id: fmt.Sprintf("item-%d", i), label: "Item"}
	}
	editor.SkillList().SetItems(items)
	editor.PersonaList().SetItems(items)
	editor.MCPList().SetItems(items)

	// Set size to something very small (smaller than header alone)
	width, height := 100, 10
	editor.SetSize(width, height)

	// Should not panic and should maintain min heights
	formHeight := lipgloss.Height(editor.View())
	if formHeight <= 0 {
		t.Error("Editor View height should be positive even on small terminals")
	}

	// Each capability list should be at least minCapListHeight = 2
	// Which renders as 3 lines (Header + 1 Item OR Header + ellipsis)
	// due to \n at the end of each rendered line.
	skillHeight := lipgloss.Height(editor.SkillList().View())
	personaHeight := lipgloss.Height(editor.PersonaList().View())
	mcpHeight := lipgloss.Height(editor.MCPList().View())

	const minExpectedHeight = 2
	if skillHeight < minExpectedHeight {
		t.Errorf("Skill list height should be at least %d, got %d", minExpectedHeight, skillHeight)
	}
	if personaHeight < minExpectedHeight {
		t.Errorf("Persona list height should be at least %d, got %d", minExpectedHeight, personaHeight)
	}
	if mcpHeight < minExpectedHeight {
		t.Errorf("MCP list height should be at least %d, got %d", minExpectedHeight, mcpHeight)
	}
}
