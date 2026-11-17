package ui

import (
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/ui/models"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

type fakeCardEditorPort struct {
	create  contracts.CreateCardInput
	fail    bool
	saveErr bool
}

func (f *fakeCardEditorPort) LoadCardEditor(dashboard, card string) (contracts.CardEditorData, error) {
	if f.fail {
		return contracts.CardEditorData{}, errors.New("load failed")
	}
	return contracts.CardEditorData{
		DashboardID: dashboard,
		CardID:      card,
		Revision:    "revision-" + card,
		CanEdit:     true,
		Values:      contracts.CardValues{Name: card},
	}, nil
}
func (f *fakeCardEditorPort) NewCardEditor(dashboard string) (contracts.CardEditorData, error) {
	return contracts.CardEditorData{DashboardID: dashboard, CanEdit: true}, nil
}
func (f *fakeCardEditorPort) ReadCardEditorChoices() contracts.CardEditorChoices {
	return contracts.CardEditorChoices{}
}
func (f *fakeCardEditorPort) CreateEditorCard(in contracts.CreateCardInput) (contracts.CardSaveResult, error) {
	f.create = in
	return contracts.CardSaveResult{CardID: "created", Revision: "saved-revision", Status: contracts.CardSaveSucceeded}, nil
}
func (f *fakeCardEditorPort) UpdateEditorCard(in contracts.UpdateCardInput) (contracts.CardSaveResult, error) {
	if f.saveErr {
		return contracts.CardSaveResult{}, errors.New("save failed")
	}
	return contracts.CardSaveResult{CardID: in.CardID, Revision: "saved-revision", Status: contracts.CardSaveSucceeded}, nil
}

func TestCardEditorAsyncRejectsOldLoadsAndFailedReloadRetainsDraft(t *testing.T) {
	c, _, _ := newCountingCoordinator(t)
	fake := &fakeCardEditorPort{}
	c.editor = fake
	first, second := c.requestCardEditor("first"), c.requestCardEditor("second")
	firstResult, secondResult := first(), second()
	c.Update(secondResult)
	c.Update(firstResult)
	e := c.views[ViewCardEditor].(*views.CardEditor)
	if e.EditingID() != "second" || e.Draft().Values.Name != "second" {
		t.Fatalf("old load replaced editor: %+v", e.Draft())
	}
	e.Inputs()[0].SetValue("unsaved")
	fake.fail = true
	reload := c.cardEditorCallbacks().OnReload()
	c.Update(reload())
	if e.Draft().Values.Name != "unsaved" || c.editorRevision != "revision-second" || c.err == nil {
		t.Fatalf("failed reload lost draft or revision: %+v", e.Draft())
	}
}

func TestCardEditorSaveCapturesDraftAndPreventsDuplicateCreation(t *testing.T) {
	c, _, _ := newCountingCoordinator(t)
	fake := &fakeCardEditorPort{}
	c.editor = fake
	c.Update(c.requestCardEditor("")())
	draft := models.CardEditorDraft{Values: contracts.CardValues{SkillIDs: []string{"original"}}}
	callbacks := c.cardEditorCallbacks()
	command := callbacks.OnSave(draft)
	if command == nil || callbacks.OnSave(draft) != nil {
		t.Fatal("duplicate save accepted")
	}
	draft.Values.SkillIDs[0] = "mutated"
	c.Update(command())
	if len(fake.create.Values.SkillIDs) != 1 || fake.create.Values.SkillIDs[0] != "original" {
		t.Fatalf("save captured mutable draft: %+v", fake.create)
	}
}

func TestCardEditorLateSaveCannotCloseNewEditor(t *testing.T) {
	c, _, _ := newCountingCoordinator(t)
	fake := &fakeCardEditorPort{}
	c.editor = fake
	c.Update(c.requestCardEditor("first")())
	save := c.cardEditorCallbacks().OnSave(models.CardEditorDraft{})
	result := save()
	c.editorEpoch++ // Simulate a dashboard change while the result is in flight.
	c.currentView = ViewDashboard
	c.Update(c.requestCardEditor("second")())
	c.Update(result)
	editor := c.views[ViewCardEditor].(*views.CardEditor)
	if c.currentView != ViewCardEditor || editor.EditingID() != "second" {
		t.Fatalf("late save closed newer editor: view=%v id=%q", c.currentView, editor.EditingID())
	}
}

//nolint:gocognit // Exercises both create and update across the complete offscreen flow.
func TestCardEditorOffscreenSaveRequiresReloadBeforeAnotherSave(t *testing.T) {
	for _, id := range []string{"", "existing"} {
		t.Run(id, func(t *testing.T) {
			c, _, _ := newCountingCoordinator(t)
			fake := &fakeCardEditorPort{}
			c.editor = fake
			c.Update(c.requestCardEditor(id)())
			editor := c.views[ViewCardEditor].(*views.CardEditor)
			editor.Inputs()[0].SetValue("submitted")
			callbacks := c.cardEditorCallbacks()
			save := callbacks.OnSave(editor.Draft())
			// A user may continue interacting with the form before leaving it.
			editor.Inputs()[0].SetValue("unsaved after submission")
			c.navigateTo(ViewSourceManagerSkills)
			c.Update(save())
			c.navigateTo(ViewCardEditor)
			if !c.editorCompleted || c.editorSaving || editor.Draft().Values.Name != "unsaved after submission" {
				t.Fatalf(
					"offscreen completion lost state: completed=%t saving=%t draft=%+v",
					c.editorCompleted,
					c.editorSaving,
					editor.Draft(),
				)
			}
			if callbacks.OnSave(editor.Draft()) != nil {
				t.Fatal("completed offscreen save permitted another save without reload")
			}
			expectedID := id
			if id == "" {
				expectedID = "created"
			}
			if c.editorCardID != expectedID || c.editorRevision != "saved-revision" {
				t.Fatalf("saved identity/revision = %q / %q", c.editorCardID, c.editorRevision)
			}
			// Reload is deliberate: the pending draft is retained until the result arrives.
			reload := callbacks.OnReload()
			if editor.Draft().Values.Name != "unsaved after submission" {
				t.Fatal("reload discarded draft before completion")
			}
			c.Update(reload())
			if c.editorCompleted || editor.EditingID() != expectedID {
				t.Fatalf("reload did not reopen saved card: id=%q completed=%t", editor.EditingID(), c.editorCompleted)
			}
			if callbacks.OnSave(editor.Draft()) == nil {
				t.Fatal("reloaded card could not be saved")
			}
		})
	}
}

func TestCardEditorFreezesDraftAndNavigationWhileSavePending(t *testing.T) {
	c, _, _ := newCountingCoordinator(t)
	c.editor = &fakeCardEditorPort{}
	c.Update(c.requestCardEditor("existing")())
	editor := c.views[ViewCardEditor].(*views.CardEditor)
	queuedNavigation := c.cardEditorCallbacks().OnOpenSources("skills")
	save := c.cardEditorCallbacks().OnSave(editor.Draft())
	editor.Update(tea.KeyPressMsg{Code: 'x', Text: "x"})
	if _, cmd := editor.Update(tea.KeyPressMsg{Code: tea.KeyEsc}); cmd != nil {
		t.Fatal("escape left an uncertain save in flight")
	}
	if c.cardEditorCallbacks().OnCancel() != nil {
		t.Fatal("cancel accepted while save is in flight")
	}
	c.Update(queuedNavigation())
	if c.currentView != ViewCardEditor || editor.Draft().Values.Name != "existing" {
		t.Fatalf("in-flight editor changed: view=%v name=%q", c.currentView, editor.Draft().Values.Name)
	}
	c.Update(save())
	if c.currentView != ViewDashboard {
		t.Fatalf("successful save did not close editor: %v", c.currentView)
	}
	c.Update(queuedNavigation())
	if c.currentView != ViewDashboard {
		t.Fatalf("stale navigation reopened a source screen: %v", c.currentView)
	}
}

func TestCardEditorCancelIsCorrelatedWithEditorEpoch(t *testing.T) {
	t.Run("late duplicate cannot close new editor", func(t *testing.T) {
		c, _, _ := newCountingCoordinator(t)
		c.editor = &fakeCardEditorPort{}
		c.Update(c.requestCardEditor("first")())
		first := c.cardEditorCallbacks().OnCancel()
		delayed := c.cardEditorCallbacks().OnCancel()
		c.Update(first())
		if c.currentView != ViewDashboard {
			t.Fatalf("normal cancel did not close editor: %v", c.currentView)
		}
		c.Update(c.requestCardEditor("second")())
		c.Update(delayed())
		editor := c.views[ViewCardEditor].(*views.CardEditor)
		if c.currentView != ViewCardEditor || editor.EditingID() != "second" {
			t.Fatalf("stale cancel closed new editor: view=%v id=%s", c.currentView, editor.EditingID())
		}
	})
	t.Run("cancel before reload becomes stale", func(t *testing.T) {
		c, _, _ := newCountingCoordinator(t)
		c.editor = &fakeCardEditorPort{}
		c.Update(c.requestCardEditor("card")())
		delayed := c.cardEditorCallbacks().OnCancel()
		reload := c.cardEditorCallbacks().OnReload()
		c.Update(reload())
		c.Update(delayed())
		if c.currentView != ViewCardEditor {
			t.Fatalf("stale cancel closed reloaded editor: %v", c.currentView)
		}
	})
	t.Run("cancel before failed save becomes stale", func(t *testing.T) {
		c, _, _ := newCountingCoordinator(t)
		c.editor = &fakeCardEditorPort{saveErr: true}
		c.Update(c.requestCardEditor("card")())
		delayed := c.cardEditorCallbacks().OnCancel()
		save := c.cardEditorCallbacks().OnSave(models.CardEditorDraft{})
		c.Update(save())
		if c.err == nil || c.currentView != ViewCardEditor {
			t.Fatal("failed save did not retain editor")
		}
		c.Update(delayed())
		if c.currentView != ViewCardEditor {
			t.Fatal("old cancel closed editor after failed save")
		}
	})
	t.Run("completed editor can be cancelled", func(t *testing.T) {
		c, _, _ := newCountingCoordinator(t)
		c.editor = &fakeCardEditorPort{}
		c.Update(c.requestCardEditor("card")())
		save := c.cardEditorCallbacks().OnSave(models.CardEditorDraft{})
		c.navigateTo(ViewSourceManagerSkills)
		c.Update(save())
		c.navigateTo(ViewCardEditor)
		editor := c.views[ViewCardEditor].(*views.CardEditor)
		_, cancel := editor.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
		if cancel == nil {
			t.Fatal("completed editor did not offer cancel")
		}
		c.Update(cancel())
		if c.currentView != ViewDashboard {
			t.Fatalf("completed editor cancel left view %v", c.currentView)
		}
	})
}
