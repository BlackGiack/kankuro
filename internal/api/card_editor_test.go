//nolint:testpackage // Tests exercise package-private revision and service fixtures.
package api

import (
	"reflect"
	"sync"
	"testing"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/config"
)

func TestEditorSavePreservesUneditableFieldsAndRejectsStaleRevision(t *testing.T) {
	service, repo, _, _, workspaceID := newServiceFixture(t)
	card := saveCard(
		t,
		repo.FileRepository,
		config.IDCard{
			DashboardID:       "default",
			Name:              "before",
			Driver:            "cat",
			WorkspaceID:       workspaceID,
			SystemInstruction: "secret",
			ReasoningEffort:   "high",
			Env:               map[string]string{"a": "b"},
		},
	)
	before, err := repo.ListIDCards()
	if err != nil {
		t.Fatal(err)
	}
	data, err := service.LoadCardEditor("default", card.ID)
	if err != nil {
		t.Fatal(err)
	}
	input := contracts.UpdateCardInput{
		DashboardID:      "default",
		CardID:           card.ID,
		ExpectedRevision: data.Revision,
		Changes:          contracts.CardChanges{Name: contracts.Change[string]{Present: true, Value: "after"}},
	}
	saved, err := service.UpdateEditorCard(input)
	if err != nil {
		t.Fatal(err)
	}
	if saved.Revision == data.Revision {
		t.Fatal("revision did not change")
	}
	cards, err := repo.ListIDCards()
	if err != nil {
		t.Fatal(err)
	}
	got := cards[0]
	if got.Name != "after" || got.SystemInstruction != card.SystemInstruction || got.Env["a"] != "b" || got.ID != card.ID ||
		got.DashboardID != card.DashboardID ||
		!got.CreatedAt.Equal(card.CreatedAt) {
		t.Fatalf("unrelated fields changed: %+v", got)
	}
	beforeValue, afterValue := reflect.ValueOf(before[0]), reflect.ValueOf(got)
	for i := range beforeValue.NumField() {
		field := beforeValue.Type().Field(i)
		if field.Name == "Name" || field.Name == "UpdatedAt" {
			continue
		}
		if !reflect.DeepEqual(beforeValue.Field(i).Interface(), afterValue.Field(i).Interface()) {
			t.Errorf("unedited field %s changed", field.Name)
		}
	}
	if _, updateErr := service.UpdateEditorCard(input); updateErr == nil {
		t.Fatal("stale revision saved")
	}
}

func TestEditorConcurrentSavesOnlyOneSucceeds(t *testing.T) {
	service, repo, _, _, workspaceID := newServiceFixture(t)
	card := saveCard(t, repo.FileRepository, config.IDCard{DashboardID: "default", Name: "before", Driver: "cat", WorkspaceID: workspaceID})
	data, err := service.LoadCardEditor("default", card.ID)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for _, name := range []string{"one", "two"} {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			_, updateErr := service.UpdateEditorCard(
				contracts.UpdateCardInput{
					DashboardID:      "default",
					CardID:           card.ID,
					ExpectedRevision: data.Revision,
					Changes:          contracts.CardChanges{Name: contracts.Change[string]{Present: true, Value: name}},
				},
			)
			errs <- updateErr
		}(name)
	}
	wg.Wait()
	close(errs)
	success := 0
	for err := range errs {
		if err == nil {
			success++
		}
	}
	if success != 1 {
		t.Fatalf("successful saves = %d, want 1", success)
	}
}

func TestEditorExplicitClearAndHiddenFieldConflict(t *testing.T) {
	service, repo, _, _, workspaceID := newServiceFixture(t)
	card := saveCard(
		t,
		repo.FileRepository,
		config.IDCard{
			DashboardID:       "default",
			Name:              "before",
			Driver:            "cat",
			Model:             "model",
			ReasoningEffort:   "high",
			WorkspaceID:       workspaceID,
			UseTmpWorkspace:   true,
			DashboardControl:  true,
			AgentsMD:          "instructions",
			SkillIDs:          []string{config.BuiltinSkillMCPCLI},
			SystemInstruction: "hidden",
		},
	)
	data, err := service.LoadCardEditor("default", card.ID)
	if err != nil {
		t.Fatal(err)
	}
	changes := contracts.CardChanges{
		Model:           contracts.Change[string]{Present: true, Value: ""},
		ReasoningEffort: contracts.Change[string]{Present: true, Value: ""},
		UseTmpWorkspace: contracts.Change[bool]{Present: true, Value: false},
		SkillIDs:        contracts.Change[[]string]{Present: true, Value: []string{}},
	}
	result, err := service.UpdateEditorCard(
		contracts.UpdateCardInput{DashboardID: "default", CardID: card.ID, ExpectedRevision: data.Revision, Changes: changes},
	)
	if err != nil || result.Status != contracts.CardSaveSucceeded {
		t.Fatalf("save: %+v %v", result, err)
	}
	cards, err := repo.ListIDCards()
	if err != nil {
		t.Fatal(err)
	}
	got := cards[0]
	if got.Model != "" || got.ReasoningEffort != "" || got.UseTmpWorkspace || len(got.SkillIDs) != 0 ||
		!got.DashboardControl || got.AgentsMD != "instructions" {
		t.Fatalf("presence behavior: %+v", got)
	}
	loaded, err := service.LoadCardEditor("default", card.ID)
	if err != nil {
		t.Fatal(err)
	}
	got.ReasoningEffort = "changed elsewhere"
	if _, updateErr := repo.UpdateIDCard(got); updateErr != nil {
		t.Fatal(updateErr)
	}
	conflict, err := service.UpdateEditorCard(
		contracts.UpdateCardInput{
			DashboardID:      "default",
			CardID:           card.ID,
			ExpectedRevision: loaded.Revision,
			Changes:          contracts.CardChanges{Name: contracts.Change[string]{Present: true, Value: "lost"}},
		},
	)
	if err == nil || conflict.Status != contracts.CardSaveConflict {
		t.Fatalf("hidden-field conflict: %+v %v", conflict, err)
	}
}

func TestEditorSaveResultRevisionMatchesPersistedCard(t *testing.T) {
	service, _, _, _, workspaceID := newServiceFixture(t)
	created, err := service.CreateEditorCard(
		contracts.CreateCardInput{
			DashboardID: "default",
			Values: contracts.CardValues{
				Name: "card", Driver: "cat", ReasoningEffort: "high",
				WorkspaceID: workspaceID, SubWorkspaceIDs: []string{},
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := service.LoadCardEditor("default", created.CardID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Values.ReasoningEffort != "high" {
		t.Fatalf("created effort = %q", loaded.Values.ReasoningEffort)
	}
	if loaded.Revision != created.Revision {
		t.Fatalf("create revision differs from persisted value: %s != %s", created.Revision, loaded.Revision)
	}
	updated, err := service.UpdateEditorCard(
		contracts.UpdateCardInput{
			DashboardID:      "default",
			CardID:           created.CardID,
			ExpectedRevision: created.Revision,
			Changes:          contracts.CardChanges{SubWorkspaceIDs: contracts.Change[[]string]{Present: true, Value: []string{}}},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err = service.LoadCardEditor("default", created.CardID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Revision != updated.Revision {
		t.Fatalf("update revision differs from persisted value: %s != %s", updated.Revision, loaded.Revision)
	}
	_, err = service.UpdateEditorCard(
		contracts.UpdateCardInput{
			DashboardID:      "default",
			CardID:           created.CardID,
			ExpectedRevision: updated.Revision,
			Changes:          contracts.CardChanges{Name: contracts.Change[string]{Present: true, Value: "next"}},
		},
	)
	if err != nil {
		t.Fatalf("second update using returned revision: %v", err)
	}
}
