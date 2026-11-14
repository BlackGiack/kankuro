package views

import (
	"testing"

	"github.com/BlackGiack/kankuro/internal/ui/models"
)

func TestAgentSwitchSetRowsDetachesNestedCardData(t *testing.T) {
	rows := []models.CardRow{{
		ID: "card", SkillIDs: []string{"skill"}, SkillNames: []string{"Skill"},
		PuppetLines: []string{"custom art"},
	}}
	view := NewAgentSwitchView(AgentSwitchCallbacks{})
	view.SetRows(rows)
	rows[0].SkillIDs[0] = "changed"
	rows[0].SkillNames[0] = "Changed"
	rows[0].PuppetLines[0] = "changed art"
	if view.cards[0].SkillIDs[0] != "skill" || view.cards[0].SkillNames[0] != "Skill" || view.cards[0].PuppetLines[0] != "custom art" {
		t.Fatalf("Agent Switch retained caller-owned nested data: %+v", view.cards[0])
	}
}
