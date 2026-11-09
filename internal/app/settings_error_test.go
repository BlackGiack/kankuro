package app_test

import (
	"errors"
	"reflect"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/BlackGiack/kankuro/internal/app"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

type settingsFailingRepository struct {
	config.Repository

	err       error
	attempted config.AppSettings
}

func (r *settingsFailingRepository) SaveAppSettings(settings config.AppSettings) error {
	r.attempted = settings
	return r.err
}

func TestSettingsSaveFailureRemainsVisible(t *testing.T) {
	repo, err := newCurrentRepository(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	before, err := repo.GetAppSettings()
	if err != nil {
		t.Fatal(err)
	}
	failure := errors.New("settings disk unavailable")
	failing := &settingsFailingRepository{Repository: repo, err: failure}
	coord := newOpenedCoordinator(failing, newTestManager(t, repo.Dir(), repo))
	coord.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	drive(t, coord, coord.NavigateTo(app.ViewSettings))
	coord.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	view := coord.Views()[app.ViewSettings].(*views.SettingsView)
	view.Input().SetValue("changed-editor")
	_, cmd := coord.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("save produced no command")
	}
	var results int
	for _, msg := range collectMsgs(t, cmd) {
		if saved, ok := msg.(messages.SettingsSavedMsg); ok {
			results++
			if !errors.Is(saved.Err, failure) {
				t.Fatalf("save error: %v", saved.Err)
			}
		}
		coord.Update(msg)
	}
	if results != 1 || failing.attempted.IDECommand != "changed-editor" {
		t.Fatalf("save was not executed: results=%d settings=%+v", results, failing.attempted)
	}
	assertBannerShown(t, coord, failure, "asynchronous settings failure")
	if coord.CurrentView() != app.ViewSettings || view.Mode() != views.SettingsModeDisplay || view.Input().Value() != "changed-editor" {
		t.Fatal("failed save lost settings view or edited value")
	}
	after, err := repo.GetAppSettings()
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatalf("failed save changed persisted settings: %+v %v", after, err)
	}
	coord.Update(messages.ClearErrorMsg{})
	if coord.Err() != nil {
		t.Fatal("banner did not clear")
	}
}
