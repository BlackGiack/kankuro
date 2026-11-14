package views_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/ui/models"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

func TestCardEditorRetainsMissingSelectionsAcrossChoiceReads(t *testing.T) {
	e := views.NewCardEditor(views.CardEditorCallbacks{})
	e.LoadEditor(
		contracts.CardEditorData{
			CanEdit: true,
			Values: contracts.CardValues{
				SkillIDs:    []string{"known", "missing"},
				WorkspaceID: "missing-workspace",
				Puppet:      "missing-puppet",
			},
		},
	)
	e.SetChoices(
		contracts.CardEditorChoices{
			Skills: contracts.ReadResult[[]contracts.EditorOption]{Attempted: true, Value: []contracts.EditorOption{{ID: "known"}}},
		},
	)
	e.SetChoices(
		contracts.CardEditorChoices{Skills: contracts.ReadResult[[]contracts.EditorOption]{Attempted: true, Err: errors.New("failed")}},
	)
	e.SetChoices(
		contracts.CardEditorChoices{
			Skills: contracts.ReadResult[[]contracts.EditorOption]{Attempted: true, Value: []contracts.EditorOption{}},
		},
	)
	d := e.Draft()
	if !reflect.DeepEqual(d.Values.SkillIDs, []string{"known", "missing"}) || d.Changes.SkillIDs.Present ||
		d.Values.WorkspaceID != "missing-workspace" ||
		d.Values.Puppet != "missing-puppet" {
		t.Fatalf("unresolved selections changed: %+v", d)
	}
	e.SetChoices(
		contracts.CardEditorChoices{
			Skills: contracts.ReadResult[[]contracts.EditorOption]{
				Attempted: true,
				Value:     []contracts.EditorOption{{ID: "missing"}, {ID: "known"}},
			},
		},
	)
	e.SkillList().Deselect("missing")
	d = e.Draft()
	if !d.Changes.SkillIDs.Present || !reflect.DeepEqual(d.Changes.SkillIDs.Value, []string{"known"}) {
		t.Fatalf("explicit removal lost: %+v", d.Changes.SkillIDs)
	}
}

func TestCardEditorHiddenSelectionsAndReset(t *testing.T) {
	e := views.NewCardEditor(views.CardEditorCallbacks{})
	e.SetChoices(
		contracts.CardEditorChoices{
			Personas: contracts.ReadResult[[]contracts.EditorOption]{Attempted: true, Value: []contracts.EditorOption{{ID: "persona"}}},
		},
	)
	e.LoadEditor(
		contracts.CardEditorData{
			CardID:   "card",
			Revision: "v1:test",
			CanEdit:  true,
			Values:   contracts.CardValues{PersonaIDs: []string{"persona"}},
		},
	)
	e.SetSettings(models.EditorSettings{ShowPersonas: false, ShowMCPs: false})
	if e.Draft().Changes.PersonaIDs.Present {
		t.Fatal("hiding an untouched section created an edit")
	}
	e.SetSettings(models.EditorSettings{ShowPersonas: true})
	for e.FocusIndex() != views.FocusPersonas {
		e.CycleFocus(1)
	}
	e.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
	e.SetSettings(models.EditorSettings{})
	change := e.Draft().Changes.PersonaIDs
	if !change.Present || change.Value == nil || len(change.Value) != 0 {
		t.Fatalf("hidden explicit clear lost: %+v", change)
	}
	e.Reset()
	if e.EditingID() != "" || e.Draft().Changes.PersonaIDs.Present || e.Draft().Values.Puppet != "kankuro" {
		t.Fatalf("reset retained prior draft: %+v", e.Draft())
	}
}

func TestCardEditorShowsActualOwnerForReadOnlyCard(t *testing.T) {
	e := views.NewCardEditor(views.CardEditorCallbacks{})
	e.LoadEditor(contracts.CardEditorData{DashboardID: "consumer", OwnerDashboardID: "owner", CardID: "card", CanEdit: false})
	e.SetSize(80, 24)
	output := e.View()
	if !strings.Contains(output, "Owner: owner") || !strings.Contains(output, "Read-only") || strings.Contains(output, "Enter=save") {
		t.Fatalf("read-only owner or help incorrect: %s", output)
	}
}

func TestCardEditorOpenPickersKeepMissingReferences(t *testing.T) {
	e := views.NewCardEditor(views.CardEditorCallbacks{})
	e.SetChoices(contracts.CardEditorChoices{
		Workspaces: contracts.ReadResult[[]contracts.WorkspaceOption]{
			Attempted: true,
			Value:     []contracts.WorkspaceOption{{ID: "a"}, {ID: "b"}},
		},
		Puppets: contracts.ReadResult[[]contracts.PuppetChoice]{
			Attempted: true,
			Value:     []contracts.PuppetChoice{{Name: "a"}, {Name: "b"}},
		},
	})
	e.LoadEditor(contracts.CardEditorData{CanEdit: true, Values: contracts.CardValues{WorkspaceID: "a", Puppet: "a"}})
	for e.FocusIndex() != views.FocusWorkspace {
		e.CycleFocus(1)
	}
	e.Update(tea.KeyPressMsg{Code: tea.KeySpace})
	e.SetChoices(
		contracts.CardEditorChoices{
			Workspaces: contracts.ReadResult[[]contracts.WorkspaceOption]{Attempted: true, Value: []contracts.WorkspaceOption{{ID: "b"}}},
		},
	)
	e.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if got := e.Draft(); got.Values.WorkspaceID != "a" || got.Changes.WorkspaceID.Present {
		t.Fatalf("workspace changed: %+v", got)
	}
	for e.FocusIndex() != views.FocusPuppet {
		e.CycleFocus(1)
	}
	e.Update(tea.KeyPressMsg{Code: tea.KeySpace})
	e.SetChoices(
		contracts.CardEditorChoices{
			Puppets: contracts.ReadResult[[]contracts.PuppetChoice]{Attempted: true, Value: []contracts.PuppetChoice{{Name: "b"}}},
		},
	)
	e.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if got := e.Draft(); got.Values.Puppet != "a" || got.Changes.Puppet.Present {
		t.Fatalf("puppet changed: %+v", got)
	}
}

func TestCardEditorWorkspaceSearchRefreshKeepsVisibleChoice(t *testing.T) {
	e := views.NewCardEditor(views.CardEditorCallbacks{})
	choices := contracts.CardEditorChoices{
		Workspaces: contracts.ReadResult[[]contracts.WorkspaceOption]{
			Attempted: true,
			Value:     []contracts.WorkspaceOption{{ID: "a", Name: "Alpha"}, {ID: "b", Name: "Beta"}},
		},
	}
	e.SetChoices(choices)
	e.LoadEditor(contracts.CardEditorData{CanEdit: true, Values: contracts.CardValues{WorkspaceID: "a"}})
	for e.FocusIndex() != views.FocusWorkspace {
		e.CycleFocus(1)
	}
	e.Update(tea.KeyPressMsg{Code: tea.KeySpace})
	e.Update(tea.KeyPressMsg{Text: "Beta"})
	e.SetChoices(choices)
	e.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	draft := e.Draft()
	if draft.Values.WorkspaceID != "b" || !draft.Changes.WorkspaceID.Present || draft.Changes.WorkspaceID.Value != "b" {
		t.Fatalf("visible choice not submitted: %+v", draft.Changes.WorkspaceID)
	}
}
